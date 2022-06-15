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
	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"iam/pkg/api/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

func ListFreezedSubjects(c *gin.Context) {
	svc := service.NewSubjectBlackListService()
	subjectPKs, err := svc.ListSubjectPK()
	if err != nil {
		util.SystemErrorJSONResponse(c, fmt.Errorf("subjectBlackListSVC.ListSubjectPK fail, %w", err))
		return
	}
	subjects := make([]types.Subject, 0, len(subjectPKs))
	for _, pk := range subjectPKs {
		s, err := cacheimpls.GetSubjectByPK(pk)
		if err != nil {
			log.WithError(err).Errorf("cacheimpls.GetSubjectByPK(%d) fail", pk)
			continue
		}
		subjects = append(subjects, s)
	}

	util.SuccessJSONResponse(c, "ok", subjects)
}

func batchSubjectBlackListExecute(c *gin.Context, doFunc func(subjectPKs []int64) error) {
	var body []freezedSubjectSerializer

	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if valid, message := common.ValidateArray(body); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	subjectPKs := make([]int64, 0, len(body))
	for _, subject := range body {
		pk, err := cacheimpls.GetLocalSubjectPK(subject.Type, subject.ID)
		if err != nil {
			util.SystemErrorJSONResponse(
				c,
				fmt.Errorf(
					"GetSubjectPK type=`%s`, id=`%s` fail, %w. maybe the subject not exists",
					subject.Type,
					subject.ID,
					err,
				),
			)
			return
		}

		subjectPKs = append(subjectPKs, pk)
	}

	err := doFunc(subjectPKs)
	if err != nil {
		util.SystemErrorJSONResponse(c, fmt.Errorf("subjectBlackListSVC.BulkCreate fail, %w", err))
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

func BatchFreezeSubjects(c *gin.Context) {
	svc := service.NewSubjectBlackListService()
	batchSubjectBlackListExecute(c, svc.BulkCreate)
}

func BatchUnfreezeSubjects(c *gin.Context) {
	svc := service.NewSubjectBlackListService()
	batchSubjectBlackListExecute(c, svc.BulkDelete)
}
