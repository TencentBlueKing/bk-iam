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
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/prp"
	"iam/pkg/cacheimpls"
	"iam/pkg/util"
)

// PolicyList godoc
// @Summary policy list
// @Description list policies of a action
// @ID api-open-system-policies-list
// @Tags open
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param params query listQuerySerializer true "the list request"
// @Success 200 {object} util.Response{data=policyListResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/policies/ [get]
func PolicyList(c *gin.Context) {
	// TODO: 翻页接口是否有性能问题 / 限制调用并发, 用drl
	var query listQuerySerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	ok, message := query.validate()
	if !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}
	// init the default query
	query.initDefault()

	// action 必须为路径下system_id的一个合法注册的操作
	systemID := c.Param("system_id")

	// 2. action exists
	actionID := query.ActionID
	actionPK, err := cacheimpls.GetActionPK(systemID, actionID)
	if err != nil {
		// 在本系统内找不到这个action, 返回404
		if errors.Is(err, sql.ErrNoRows) {
			util.NotFoundJSONResponse(c, "action_id not exist in this system")
			return
		}

		err = fmt.Errorf("cacheimpls.GetActionPK system=`%s`, action=`%s` fail. err=%w", systemID, actionID, err)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	offset := (query.Page - 1) * query.PageSize
	limit := query.PageSize

	manager := prp.NewOpenPolicyManager()
	count, policies, err := manager.List(query.Type, actionPK, query.Timestamp, offset, limit)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	if count == 0 {
		util.SuccessJSONResponse(c, "ok", policyListResponse{
			Metadata: policyListResponseMetadata{
				System:    systemID,
				Action:    policyResponseAction{ID: actionID},
				Timestamp: query.Timestamp,
			},
			Count:   count,
			Results: nil,
		})
		return
	}

	var results []thinPolicyResponse
	results, err = convertOpenPoliciesToThinPolicies(policies)
	if err != nil {
		err = fmt.Errorf(
			"convertQueryPoliciesToThinPolicies system=`%s`, action=`%s` fail. err=%w",
			systemID, actionID, err)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// 返回每条策略, 包含的过期时间, 接入方得二次校验
	util.SuccessJSONResponse(c, "ok", policyListResponse{
		Metadata: policyListResponseMetadata{
			System:    systemID,
			Action:    policyResponseAction{ID: actionID},
			Timestamp: query.Timestamp,
		},
		Count:   count,
		Results: results,
	})
}

func convertOpenPoliciesToThinPolicies(
	policies []prp.OpenPolicy,
) (responses []thinPolicyResponse, err error) {
	if len(policies) == 0 {
		return
	}

	// loop policies to build thinPolicies
	for _, p := range policies {
		subj, err1 := cacheimpls.GetSubjectByPK(p.SubjectPK)
		// if get subject fail, continue
		if err1 != nil {
			log.WithError(err1).
				Errorf("policy_list.convertQueryPoliciesToThinPolicies get subject subject_pk=`%d` fail", p.SubjectPK)

			continue
		}

		responses = append(responses, thinPolicyResponse{
			Version: p.Version,
			ID:      p.ID,
			Subject: policyResponseSubject{
				Type: subj.Type,
				ID:   subj.ID,
				Name: subj.Name,
			},
			Expression: p.Expression,
			ExpiredAt:  p.ExpiredAt,
		})
	}
	return responses, nil
}
