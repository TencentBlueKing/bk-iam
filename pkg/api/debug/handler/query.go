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
	"errors"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/gin-gonic/gin"

	"iam/pkg/abac/pdp"
	"iam/pkg/abac/pip/group"
	"iam/pkg/abac/types/request"
	modelHandler "iam/pkg/api/model/handler"
	"iam/pkg/cacheimpls"
	"iam/pkg/logging/debug"
	"iam/pkg/service"
	"iam/pkg/util"
)

// QueryModel ...
func QueryModel(c *gin.Context) {
	systemID, ok := c.GetQuery("system")
	if !ok {
		util.BadRequestErrorJSONResponse(c, "system required")
		return
	}

	fields := "base_info,resource_types,actions,action_groups,instance_selections,resource_creator_actions," +
		"common_actions,feature_shield_rules"
	fieldSet := set.SplitStringToSet(fields, ",")
	modelHandler.BuildSystemInfoQueryResponse(c, systemID, fieldSet)
}

// QueryActions ...
func QueryActions(c *gin.Context) {
	systemID, ok := c.GetQuery("system")
	if !ok {
		util.BadRequestErrorJSONResponse(c, "system required")
		return
	}

	// 获取action信息
	svc := service.NewActionService()
	actions, err := svc.ListBySystem(systemID)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// get action pks
	actionPKs := make(map[string]int64, len(actions))
	for _, ac := range actions {
		pk, err := cacheimpls.GetActionPK(systemID, ac.ID)
		if err == nil {
			actionPKs[ac.ID] = pk
		}
	}

	// build and return
	util.SuccessJSONResponse(c, "ok", gin.H{
		"actions": actions,
		"pks":     actionPKs,
	})
}

type querySubjectsSerializer struct {
	Type string `form:"type" binding:"required"`
	ID   string `form:"id" binding:"required"`
}

// QuerySubjects ...
func QuerySubjects(c *gin.Context) {
	var body querySubjectsSerializer
	if err := c.ShouldBindQuery(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	errs := map[string]error{}

	// 1. 查subject本身, 含PK
	subject := gin.H{
		"type": body.Type,
		"id":   body.ID,
	}
	pk, err := cacheimpls.GetSubjectPK(body.Type, body.ID)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}
	subject["pk"] = pk

	// 2. 查subject所属的部门
	// 3. 查subject所属的组
	depts := []gin.H{}

	detail, err := cacheimpls.GetSubjectDetail(pk)
	departments := detail.DepartmentPKs
	groups := detail.SubjectGroups

	if err != nil {
		errs["GetSubjectDepartment"] = err
		errs["GetSubjectGroups"] = err
	} else {
		for _, deptPK := range departments {
			subj, err1 := cacheimpls.GetSubjectByPK(deptPK)
			if err1 != nil {
				depts = append(depts, gin.H{
					"pk":  deptPK,
					"err": err1.Error(),
				})
			} else {
				d := gin.H{
					"pk":   deptPK,
					"type": subj.Type,
					"id":   subj.ID,
					"name": subj.Name,
				}

				// 查询部门所属的组
				subjectGroups, err2 := group.GetSubjectGroupsFromCache(group.SubjectTypeDepartment, []int64{deptPK})
				if err2 != nil {
					d["groups"] = err2.Error()
				} else {
					d["groups"] = subjectGroups[deptPK]
				}

				depts = append(depts, d)
			}
		}
	}

	util.SuccessJSONResponse(c, "ok", gin.H{
		"subject":     subject,
		"departments": depts,
		"groups":      groups,
		"errs":        errs,
	})
}

type queryPoliciesSerializer struct {
	System      string `form:"system" binding:"required"`
	SubjectType string `form:"subject_type" binding:"required"`
	SubjectID   string `form:"subject_id" binding:"required"`
	Action      string `form:"action" binding:"required"`
}

// QueryPolicies ...
func QueryPolicies(c *gin.Context) {
	var body queryPoliciesSerializer
	if err := c.ShouldBindQuery(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	var entry *debug.Entry

	if _, isDebug := c.GetQuery("debug"); isDebug {
		entry = debug.EntryPool.Get()
		defer debug.EntryPool.Put(entry)
	}
	_, isForce := c.GetQuery("force")

	// make a request
	req := request.NewRequest()
	req.System = body.System
	req.Subject.Type = body.SubjectType
	req.Subject.ID = body.SubjectID
	req.Action.ID = body.Action

	// do query
	expr, err := pdp.Query(req, entry, false, isForce)
	debug.WithError(entry, err)
	if err != nil {
		if errors.Is(err, pdp.ErrInvalidAction) {
			util.BadRequestErrorJSONResponse(c, err.Error())
			return
		}

		util.SystemErrorJSONResponseWithDebug(c, err, entry)
		return
	}

	util.SuccessJSONResponseWithDebug(c, "ok", expr, entry)
}
