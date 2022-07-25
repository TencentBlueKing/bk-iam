/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/bsm/redislock"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/cacheimpls"
	"iam/pkg/database"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

var ErrNeedRetry = errors.New("need retry")

// GroupAlterMessageHandler ...
type GroupAlterMessageHandler interface {
	Handle(message GroupAlterMessage) error
}

type groupAlterMessageHandler struct {
	groupService                      service.GroupService
	subjectActionGroupResourceService service.SubjectActionGroupResourceService
	subjectActionExpressionService    service.SubjectActionExpressionService

	locker *subjectActionLocker
}

// NewGroupAlterMessageHandler ...
func NewGroupAlterMessageHandler() GroupAlterMessageHandler {
	return &groupAlterMessageHandler{
		groupService:                      service.NewGroupService(),
		subjectActionGroupResourceService: service.NewSubjectActionGroupResourceService(),
		subjectActionExpressionService:    service.NewSubjectActionExpressionService(),
		locker:                            newSubjectActionLocker(),
	}
}

// Handle ...
func (h *groupAlterMessageHandler) Handle(message GroupAlterMessage) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ConnTypeConsumer, "Handle")

	subjectPK := message.SubjectPK
	actionPK := message.ActionPK
	groupPK := message.GroupPK

	// 分布式锁, subject_pk, action_pk
	// 请求锁最多3分钟
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	lock, err := h.locker.acquire(ctx, subjectPK, actionPK)
	if errors.Is(err, redislock.ErrNotObtained) {
		// 锁没请求到, 消息需要重试
		return ErrNeedRetry
	} else if err != nil {
		return errorWrapf(err, "acquire lock fail, subjectPK=`%d`, actionPK`%d`", subjectPK, actionPK)
	}
	defer lock.Release(ctx)

	tx, err := database.GenerateDefaultDBTx()
	if err != nil {
		return err
	}
	defer database.RollBackWithLog(tx)

	var obj types.SubjectActionGroupResource
	if groupPK != 0 {
		// 查询subject group关系过期时间
		expiredAt, found, err := h.groupService.GetExpiredAtBySubjectGroup(subjectPK, groupPK)
		if err != nil {
			return errorWrapf(err,
				"groupService.GetExpiredAtBySubjectGroup fail, subjectPK=`%d`, groupPK=`%d`",
				subjectPK, groupPK,
			)
		}

		// 查询group action授权资源实例
		resourceMap, err := cacheimpls.GetGroupActionAuthorizedResource(
			groupPK,
			actionPK,
		)
		if err != nil {
			return errorWrapf(err,
				"cacheimpls.GetGroupActionAuthorizedResource fail, groupPK=`%d`, actionPK=`%d`",
				groupPK, actionPK,
			)
		}

		if found && len(resourceMap) != 0 {
			obj, err = h.subjectActionGroupResourceService.CreateOrUpdateWithTx(
				tx,
				subjectPK,
				actionPK,
				groupPK,
				expiredAt,
				resourceMap,
			)
			if err != nil {
				return errorWrapf(err,
					"subjectActionGroupResourceService.CreateOrUpdateWithTx fail, subjectPK=`%d`, actionPK=`%d`, "+
						"groupPK=`%d`, expiredAt=`%d`, resourceMap=`%+v`",
					subjectPK, actionPK, groupPK, expiredAt, resourceMap,
				)
			}
		} else {
			// 关系不存在, 或者group授权的资源实例为空, 从subject action group resource中删除对应的groupPK
			obj, err = h.subjectActionGroupResourceService.DeleteGroupWithTx(tx, subjectPK, actionPK, groupPK)
			if err != nil {
				return errorWrapf(err,
					"subjectActionGroupResourceService.DeleteGroupWithTx fail, subjectPK=`%d`, actionPK=`%d`, "+
						"groupPK=`%d`",
					subjectPK, actionPK, groupPK,
				)
			}
		}
	} else {
		// groupPK == 0, 只更新表达式
		obj, err = h.subjectActionGroupResourceService.Get(subjectPK, groupPK)
		if err != nil {
			return errorWrapf(err,
				"subjectActionGroupResourceService.Get fail, subjectPK=`%d`, groupPK=`%d`",
				subjectPK, groupPK,
			)
		}
	}

	// subject action resource group -> subject action expression
	expression, err := convertToSubjectActionExpression(obj)
	if err != nil {
		return errorWrapf(err,
			"convertToSubjectActionExpression fail, subjectActionResourceGroup=`%+v`",
			obj,
		)
	}

	// 更新subject action expression
	err = h.subjectActionExpressionService.CreateOrUpdateWithTx(tx, expression)
	if err != nil {
		return errorWrapf(err,
			"subjectActionExpressionService.CreateOrUpdateWithTx fail, subjectActionExpression=`%+v`",
			expression,
		)
	}

	// TODO 清理 subject action expression缓存

	return tx.Commit()
}

func convertToSubjectActionExpression(
	obj types.SubjectActionGroupResource,
) (expression types.SubjectActionExpression, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ConnTypeConsumer, "convertToSubjectActionExpression")

	// 组合 subject 所有 group 授权的资源实例
	now := time.Now().Unix()
	minExpiredAt := int64(util.NeverExpiresUnixTime)                  // 所有用户组中, 最小的过期时间
	resourceMap := make(map[int64][]string, len(obj.GroupResource)*2) // resource_type_pk -> resource_ids
	for _, groupResource := range obj.GroupResource {
		// 忽略已过期的用户组
		if groupResource.ExpiredAt < now {
			continue
		}

		if groupResource.ExpiredAt < minExpiredAt {
			minExpiredAt = groupResource.ExpiredAt
		}

		// 组合所有有效的资源实例
		for resourceTypePK, resourceIDs := range groupResource.Resources {
			resourceMap[resourceTypePK] = append(resourceMap[resourceTypePK], resourceIDs...)
		}
	}

	if len(resourceMap) == 0 {
		// 有权限的资源实例为空, 表达式为空, 过期时间为0
		expression = types.SubjectActionExpression{
			SubjectPK:  obj.SubjectPK,
			ActionPK:   obj.ActionPK,
			Expression: `{}`,
			ExpiredAt:  0,
		}
		return
	}

	// 查询操作的信息
	action, err := cacheimpls.GetAction(obj.ActionPK)
	if err != nil {
		err = errorWrapf(err, "cacheimpls.GetAction fail, actionPK=`%d`", obj.ActionPK)
		return expression, err
	}

	// 查询操作关联的资源类型
	system := action.System
	actionDetail, err := cacheimpls.GetActionDetail(system, action.ID)
	if err != nil {
		err = errorWrapf(err, "cacheimpls.GetActionDetail fail, system=`%s`, actionID=`%s`", system, action.ID)
		return expression, err
	}

	if len(actionDetail.ResourceTypes) != 1 {
		err = errorWrapf(fmt.Errorf(
			"rbac action must related one resource type, but got %d, actionPK=`%d`",
			len(actionDetail.ResourceTypes),
			obj.ActionPK,
		), "")
		return expression, err
	}

	actionResourceType := actionDetail.ResourceTypes[0]
	actionResourceTypeID := actionResourceType.ID

	// 查询资源类型pk
	actionResourceTypePK, err := cacheimpls.GetLocalResourceTypePK(actionResourceType.System, actionResourceTypeID)
	if err != nil {
		err = errorWrapf(
			err,
			"cacheimpls.GetLocalResourceTypePK fail, system=`%s`, resourceTypeID=`%s`",
			actionResourceType.System,
			actionResourceTypeID,
		)
		return expression, err
	}

	// 生成表达式
	content := make([]interface{}, 0, len(resourceMap))
	for resourceTypePK, resourceIDs := range resourceMap {
		if resourceTypePK == actionResourceTypePK {
			// 授权的资源类型与操作的资源类型相同, 生成StringEquals表达式
			// {"StringEquals": {"system_id.resource_type_id.id": ["resource_id"]}}
			content = append(content, map[string]interface{}{
				"StringEquals": map[string]interface{}{
					fmt.Sprintf("%s.%s.id", system, actionResourceTypeID): resourceIDs,
				},
			})

			continue
		}

		// 查询资源类型
		resourceType, err := cacheimpls.GetThinResourceType(resourceTypePK)
		if err != nil {
			return expression, err
		}
		resourceTypeID := resourceType.ID

		resourceNodes := make([]string, 0, len(resourceIDs))
		for _, resourceID := range resourceIDs {
			resourceNodes = append(resourceNodes, fmt.Sprintf("/%s,%s/", resourceTypeID, resourceID))
		}

		// 资源类型与操作的资源类型不同, 生成StringContains表达式
		// {"StringContains": {"system_id.resource_type_id._bk_iam_path_": ["/resource_type_id,resource_id/"]}}
		content = append(content, map[string]interface{}{
			"StringContains": map[string]interface{}{
				fmt.Sprintf("%s.%s._bk_iam_path_", system, actionResourceTypeID): resourceNodes,
			},
		})
	}

	// 组合表达式
	var exp interface{}
	if len(content) == 1 {
		exp = content[0]
	} else {
		// {"OR": {"content": []}}
		exp = map[string]interface{}{
			"OR": map[string]interface{}{
				"content": content,
			},
		}
	}

	expStr, err := jsoniter.MarshalToString(exp)
	if err != nil {
		err = errorWrapf(err, "jsoniter.MarshalToString fail, exp=`%s`", exp)
		return expression, err
	}

	expression = types.SubjectActionExpression{
		SubjectPK:  obj.SubjectPK,
		ActionPK:   obj.ActionPK,
		Expression: expStr,
		ExpiredAt:  minExpiredAt,
	}
	return expression, nil
}
