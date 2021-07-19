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
	svctypes "iam/pkg/service/types"
	"iam/pkg/util"

	"github.com/gin-gonic/gin"
)

// DeleteActionPolicies 并非实时删除，而是发起一个删除事件
func DeleteActionPolicies(c *gin.Context) {
	systemID := c.Param("system_id")
	actionID := c.Param("action_id")

	// 查询ActionPK
	actionSvc := service.NewActionService()
	actionPK, err := actionSvc.GetActionPK(systemID, actionID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "DeleteActionPolicies",
			"actionSvc.GetActionPK systemID=`%s` actionID=`%s` fail", systemID, actionID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	eventSvc := service.NewModelChangeService()
	// 检查是否已经存在，若存在，则直接返回，避免重复添加事件
	// TODO: 由于DB没有相关约束，并发可能有问题
	exist, err := eventSvc.ExistByTypeModel(
		ModelChangeEventTypeActionPolicyDeleted,
		ModelChangeEventStatusPending,
		ModelChangeEventModelTypeAction,
		actionPK,
	)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "DeleteActionPolicies",
			"eventSvc.ExistByTypeModel actionPK=`%d` fail", actionPK)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	// 不存在pending的事件，则添加
	if !exist {
		event := svctypes.ModelChangeEvent{
			Type:      ModelChangeEventTypeActionPolicyDeleted,
			Status:    ModelChangeEventStatusPending,
			SystemID:  systemID,
			ModelType: ModelChangeEventModelTypeAction,
			ModelID:   actionID,
			ModelPK:   actionPK,
		}
		err = eventSvc.BulkCreate([]svctypes.ModelChangeEvent{event})
		if err != nil {
			err = errorx.Wrapf(err, "Handler", "DeleteActionPolicies",
				"eventSvc.BulkCreate event=`%+v` fail", event)
			util.SystemErrorJSONResponse(c, err)
			return
		}
	}

	util.SuccessJSONResponse(c, "ok", nil)
}
