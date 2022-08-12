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

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/prp"
	"iam/pkg/cacheimpls"
	"iam/pkg/util"
)

// PoliciesSubjects godoc
// @Summary query subjects/获取有权限的用户列表
// @Description system+action+policy_ids => subjects
// @ID api-open-system-policies-subjects
// @Tags open
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param params query subjectsSerializer true "the subject request"
// @Success 200 {object} util.Response{data=policySubjectsResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/policies/-/subjects [get]
func PoliciesSubjects(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "policy_list.PoliciesSubjects")
	// TODO: maybe add cache here;
	// key is subjectsSerializer md5 => value is []subjects; search from git history LocalPolicySubjectsCache
	systemID := c.Param("system_id")
	var query subjectsSerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	pks, err := util.StringToInt64Slice(query.IDs, ",")
	if err != nil {
		util.BadRequestErrorJSONResponse(c, fmt.Sprintf("ids(%s) should be string like 1,2,3", query.IDs))
		return
	}

	_type := "abac"

	manager := prp.NewOpenPolicyManager()
	policySubjects, err := manager.ListSubjects(_type, systemID, pks)
	if err != nil {
		err = fmt.Errorf(
			"manager.ListSubjects _type=`%s`, systemID=`%+s`, pks=`%+v` fail. err=%w",
			_type,
			systemID,
			pks,
			err,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	data := policySubjectsResponse{}
	for policyPK, subjectPK := range policySubjects {
		subj, err1 := cacheimpls.GetSubjectByPK(subjectPK)
		// if get subject fail, continue
		if err1 != nil {
			log.Info(errorWrapf(err1, "cacheimpls.GetSubjectByPK subject_pk=`%d` fail", subjectPK))

			continue
		}

		data = append(data, policyIDSubject{
			PolicyID: policyPK,
			Subject: policyResponseSubject{
				Type: subj.Type,
				ID:   subj.ID,
				Name: subj.Name,
			},
		})
	}

	util.SuccessJSONResponse(c, "ok", data)
}
