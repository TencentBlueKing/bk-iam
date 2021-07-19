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
	"iam/pkg/errorx"
	"iam/pkg/service"
	"iam/pkg/util"

	"github.com/gin-gonic/gin"
)

// ListResourceType 查询系统所有资源类型
func ListResourceType(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListResourceType")

	var query resourceTypeSerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if valid, message := query.validate(); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	systems := query.systems()
	fieldsSet := query.fieldsSet()
	data := make(map[string][]map[string]interface{}, len(systems))

	svc := service.NewResourceTypeService()

	for _, system := range systems {
		resourceTypes, err := svc.ListBySystem(system)
		if err != nil {
			err = errorWrapf(err, "svc.ListBySystem systemID=`%s`", system)
			util.SystemErrorJSONResponse(c, err)
			return
		}

		types := make([]map[string]interface{}, 0, len(resourceTypes))
		for _, r := range resourceTypes {
			//types = append(types, filterFields(fieldsSet, r))
			rf, err := filterFields(fieldsSet, r)
			if err != nil {
				err = errorWrapf(err, "filterFields set=`%+v`, system=`%s`", fieldsSet, system)
				util.SystemErrorJSONResponse(c, err)
				return
			}

			types = append(types, rf)
		}

		data[system] = types
	}

	util.SuccessJSONResponse(c, "ok", data)
}
