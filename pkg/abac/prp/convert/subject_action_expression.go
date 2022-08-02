/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package convert

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

const convertLayer = "convert"

func SubjectActionGroupResourceToExpression(
	obj types.SubjectActionGroupResource,
) (expression types.SubjectActionExpression, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(convertLayer, "convertToSubjectActionExpression")

	// 组合 subject 所有 group 授权的资源实例
	minExpiredAt, resourceMap := mergeGroupResource(obj)

	// 授权的资源实例为空, 返回空表达式, 过期时间为0
	if len(resourceMap) == 0 {
		return types.SubjectActionExpression{
			SubjectPK:  obj.SubjectPK,
			ActionPK:   obj.ActionPK,
			Expression: `{}`,
			ExpiredAt:  0,
		}, util.ErrNilRequestBody
	}

	// 生成表达式内容
	content, err := genExpressionContent(obj.ActionPK, resourceMap)
	if err != nil {
		err = errorWrapf(err, "genExpressionContent fail, actionPK=`%d`, resourceMap=`%+v`", obj.ActionPK, resourceMap)
		return expression, err
	}

	// 转换表达式字符串
	exp, err := convertToString(content)
	if err != nil {
		err = errorWrapf(err, "convertToString fail, content=`%+v`", content)
		return expression, err
	}

	expression = types.SubjectActionExpression{
		SubjectPK:  obj.SubjectPK,
		ActionPK:   obj.ActionPK,
		Expression: exp,
		ExpiredAt:  minExpiredAt,
	}
	return expression, nil
}

func convertToString(content []interface{}) (string, error) {
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
		return "", err
	}
	return expStr, nil
}

// genExpressionContent 生成表达式内容
func genExpressionContent(actionPK int64, resourceMap map[int64][]string) ([]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(convertLayer, "genExpressionContent")
	// 查询操作的信息
	action, err := cacheimpls.GetAction(actionPK)
	if err != nil {
		err = errorWrapf(err, "cacheimpls.GetAction fail, actionPK=`%d`", actionPK)
		return nil, err
	}

	// 查询操作关联的资源类型
	system := action.System
	actionDetail, err := cacheimpls.GetLocalActionDetail(system, action.ID)
	if err != nil {
		err = errorWrapf(err, "cacheimpls.GetActionDetail fail, system=`%s`, actionID=`%s`", system, action.ID)
		return nil, err
	}

	if len(actionDetail.ResourceTypes) != 1 {
		err = errorWrapf(fmt.Errorf(
			"rbac action must related one resource type, but got %d, actionPK=`%d`",
			len(actionDetail.ResourceTypes),
			actionPK,
		), "")
		return nil, err
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
		return nil, err
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
			err = errorWrapf(
				err,
				"cacheimpls.GetThinResourceType fail, resourceTypePK=`%d`",
				resourceTypePK,
			)
			return nil, err
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
	return content, nil
}

// mergeGroupResource 合并用户组授权的资源实例
func mergeGroupResource(obj types.SubjectActionGroupResource) (int64, map[int64][]string) {
	// 组合 subject 所有 group 授权的资源实例
	now := time.Now().Unix()
	minExpiredAt := int64(util.NeverExpiresUnixTime)                  // 所有用户组中, 最小的过期时间
	resourceMap := make(map[int64][]string, len(obj.GroupResource)*2) // resource_type_pk -> resource_ids

	for _, groupResource := range obj.GroupResource {
		// 忽略过期的用户组
		if groupResource.ExpiredAt < now {
			continue
		}

		if groupResource.ExpiredAt < minExpiredAt {
			minExpiredAt = groupResource.ExpiredAt
		}

		for resourceTypePK, resourceIDs := range groupResource.Resources {
			resourceMap[resourceTypePK] = append(resourceMap[resourceTypePK], resourceIDs...)
		}
	}

	return minExpiredAt, resourceMap
}
