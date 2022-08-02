/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/prp/convert"
	"iam/pkg/cacheimpls"
	"iam/pkg/config"
	"iam/pkg/database"
	"iam/pkg/locker"
	"iam/pkg/logging"
	"iam/pkg/service"
	"iam/pkg/service/types"
)

const handlerLayer = "handler"

type groupAlterMessageHandler struct {
	groupService                      service.GroupService
	groupAlterEventService            service.GroupAlterEventService
	subjectActionGroupResourceService service.SubjectActionGroupResourceService
	subjectActionExpressionService    service.SubjectActionExpressionService

	locker *locker.SubjectDistributedActionLocker
}

// NewGroupAlterMessageHandler ...
func NewGroupAlterMessageHandler() MessageHandler {
	return &groupAlterMessageHandler{
		groupService:                      service.NewGroupService(),
		groupAlterEventService:            service.NewGroupAlterEventService(),
		subjectActionGroupResourceService: service.NewSubjectActionGroupResourceService(),
		subjectActionExpressionService:    service.NewSubjectActionExpressionService(),
		locker:                            locker.NewDistributedSubjectActionLocker(),
	}
}

// Handle ...
func (h *groupAlterMessageHandler) Handle(message string) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(handlerLayer, "handleEvent")

	pk, err := strconv.ParseInt(message, 10, 64)
	if err != nil {
		err = errorWrapf(err, "parse message to pk fail, message=`%s`", message)
		return err
	}

	event, err := h.groupAlterEventService.Get(pk)
	if err != nil {
		err = errorWrapf(err, "groupAlterEventService.Get event fail, pk=`%d`", pk)
		return err
	}

	// 判断event check times超限，不再处理
	maxCheckCount := int64(config.MaxGroupAlterEventCheckCount)
	if event.CheckCount > maxCheckCount {
		logger := logging.GetWorkerLogger()
		logger.Errorf("group event pk=`%d` check times exceed limit, check times=`%d`", pk, event.CheckCount)
		return nil
	}

	// 循环处理所有事件
	groupPK := event.GroupPK
	for _, actionPK := range event.ActionPKs {
		for _, subjectPK := range event.SubjectPKs {
			err = h.alterSubjectActionGroupResource(subjectPK, actionPK, groupPK)
			if err != nil {
				return err
			}
		}
	}

	err = h.groupAlterEventService.Delete(pk)
	if err != nil {
		err = errorWrapf(err, "groupAlterEventService.Delete event fail, pk=`%d`", pk)
		return err
	}

	return nil
}

// alterSubjectActionGroupResource 处理独立的事件
func (h *groupAlterMessageHandler) alterSubjectActionGroupResource(subjectPK, actionPK, groupPK int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(handlerLayer, "handleEvent")

	logger := logging.GetWorkerLogger()

	// 分布式锁, subject_pk, action_pk
	// 请求锁最多3分钟
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	lock, err := h.locker.Acquire(ctx, subjectPK, actionPK)
	if err != nil {
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
		expiredAt, err := h.groupService.GetExpiredAtBySubjectGroup(subjectPK, groupPK)
		if err != nil && !errors.Is(err, service.ErrGroupMemberNotFound) {
			return errorWrapf(err,
				"groupService.GetExpiredAtBySubjectGroup fail, subjectPK=`%d`, groupPK=`%d`",
				subjectPK, groupPK,
			)
		}
		found := !errors.Is(err, service.ErrGroupMemberNotFound)

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
			obj, err = h.subjectActionGroupResourceService.DeleteGroupResourceWithTx(tx, subjectPK, actionPK, groupPK)
			if errors.Is(err, sql.ErrNoRows) {
				logger.Warnf("subject action group resource not found, subjectPK=`%d`, actionPK=`%d`", subjectPK, actionPK)
				return nil
			}

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
		obj, err = h.subjectActionGroupResourceService.Get(subjectPK, actionPK)
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warnf("subject action group resource not found, subjectPK=`%d`, actionPK=`%d`", subjectPK, actionPK)
			return nil
		}

		if err != nil {
			return errorWrapf(err,
				"subjectActionGroupResourceService.Get fail, subjectPK=`%d`, groupPK=`%d`",
				subjectPK, groupPK,
			)
		}
	}

	// subject action resource group -> subject action expression
	expression, err := convert.ConvertSubjectActionGroupResourceToExpression(obj)
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

	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx.Commit fail")
	}

	// 清理 subject action expression缓存
	cacheimpls.DeleteSubjectActionExpressionCache(subjectPK, actionPK)

	return nil
}
