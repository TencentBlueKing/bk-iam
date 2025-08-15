/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云 - 权限中心 (BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"

	"iam/pkg/abac/pap"
	"iam/pkg/api/common"
	"iam/pkg/util"
)

// BatchCreateSubjectTemplateGroup 批量创建 subject-template-group
func BatchCreateSubjectTemplateGroup(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchCreateSubjectTemplateGroup")

	var body []subjectTemplateGroupSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := common.ValidateArray(body); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	papSubjectTemplateGroups := convertToPapSubjectTemplateGroup(body)

	ctl := pap.NewGroupController()
	var err error
	// Note: 由于底层 subject_system_group 的变更可能会导致死锁，所以这里进行了重试
	for i := 0; i < util.DBDeadLockRetryCount; i++ {
		err = ctl.BulkCreateSubjectTemplateGroup(papSubjectTemplateGroups)
		// 非死锁，无需重试
		if !util.IsDeadLockError(err) {
			break
		}
		time.Sleep(util.DBDeadLockRetryInterval)
	}

	if err != nil {
		err = errorWrapf(
			err,
			"ctl.BulkCreateSubjectTemplateGroup",
			"subjectTemplateGroups=`%+v`",
			papSubjectTemplateGroups,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

func BatchDeleteSubjectTemplateGroup(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchDeleteSubjectTemplateGroup")

	var body []subjectTemplateGroupSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := common.ValidateArray(body); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	papSubjectTemplateGroups := convertToPapSubjectTemplateGroup(body)

	ctl := pap.NewGroupController()
	var err error
	for i := 0; i < util.DBDeadLockRetryCount; i++ {
		err = ctl.BulkDeleteSubjectTemplateGroup(papSubjectTemplateGroups)
		// 非死锁，无需重试
		if util.IsDeadLockError(err) {
			break
		}
		time.Sleep(util.DBDeadLockRetryInterval)
	}

	if err != nil {
		err = errorWrapf(
			err,
			"ctl.BulkDeleteSubjectTemplateGroup",
			"subjectTemplateGroups=`%+v`",
			papSubjectTemplateGroups,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

// BatchUpdateSubjectTemplateGroupExpiredAt 批量更新 subject-template-group 过期时间
func BatchUpdateSubjectTemplateGroupExpiredAt(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchUpdateSubjectTemplateGroupExpiredAt")

	var body []subjectTemplateGroupSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := common.ValidateArray(body); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	papSubjectTemplateGroups := convertToPapSubjectTemplateGroup(body)

	ctl := pap.NewGroupController()
	err := ctl.UpdateSubjectTemplateGroupExpiredAt(papSubjectTemplateGroups)
	if err != nil {
		err = errorWrapf(
			err,
			"ctl.UpdateSubjectTemplateGroupExpiredAt",
			"subjectTemplateGroups=`%+v`",
			papSubjectTemplateGroups,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

func convertToPapSubjectTemplateGroup(body []subjectTemplateGroupSerializer) []pap.SubjectTemplateGroup {
	papSubjectTemplateGroups := make([]pap.SubjectTemplateGroup, 0, len(body))
	for _, m := range body {
		papSubjectTemplateGroups = append(papSubjectTemplateGroups, pap.SubjectTemplateGroup{
			Type:       m.Type,
			ID:         m.ID,
			TemplateID: m.TemplateID,
			GroupID:    m.GroupID,
			ExpiredAt:  m.ExpiredAt,
		})
	}
	return papSubjectTemplateGroups
}
