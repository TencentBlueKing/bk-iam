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

	"iam/pkg/cacheimpls"
	"iam/pkg/errorx"
	"iam/pkg/service"
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

	// NOTE: 防止敏感信息泄漏, 只能查询自己系统 + 自己action的
	// 1. query policy
	svc := service.NewPolicyService()
	policies, err := svc.ListQueryByPKs(pks)
	if err != nil {
		err = fmt.Errorf("svc.ListQueryByPKs system=`%s`, policy_ids=`%+v` fail. err=%w",
			systemID, pks, err)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	data := policySubjectsResponse{}
	for _, policy := range policies {
		sa, err := cacheimpls.GetAction(policy.ActionPK)
		if err != nil {
			log.Info(errorWrapf(err,
				"policy_list.GetSystemAction action_pk=`%d` fail",
				policy.ActionPK))
			continue
		}
		// 不是本系统的策略, 过滤掉. not my system policy, continue
		if systemID != sa.System {
			continue
		}

		subj, err1 := cacheimpls.GetSubjectByPK(policy.SubjectPK)
		// if get subject fail, continue
		if err1 != nil {
			log.Info(errorWrapf(err1,
				"policy_list.PoliciesSubjects GetSubjectByPK subject_pk=`%d` fail",
				policy.SubjectPK))
			continue
		}

		data = append(data, policyIDSubject{
			PolicyID: policy.PK,
			Subject: policyResponseSubject{
				Type: subj.Type,
				ID:   subj.ID,
				Name: subj.Name,
			},
		})
	}

	util.SuccessJSONResponse(c, "ok", data)
}
