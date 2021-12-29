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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cacheimpls"
	"iam/pkg/errorx"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// UserGroups godoc
// @Summary user groups
// @Description get a user's groups, include the inherit groups from department
// @ID api-open-user-groups-get
// @Tags open
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param inherit query bool true "get subject's inherit groups from it's departments"
// @Success 200 {object} util.Response{data=subjectGroupsResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/open/users/{user_id}/groups [get]
func UserGroups(c *gin.Context) {
	subjectID := c.Param("user_id")

	inheritValue, ok := c.GetQuery("inherit")
	inherit := ok && strings.ToLower(inheritValue) == "true"

	handleSubjectGroups(c, types.UserType, subjectID, inherit)
}

// DepartmentGroups godoc
// @Summary department groups
// @Description get a department's groups
// @ID api-open-department-groups-get
// @Tags open
// @Accept json
// @Produce json
// @Param department_id path string true "Department ID"
// @Success 200 {object} util.Response{data=subjectGroupsResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/open/departments/{user_id}/groups [get]
func DepartmentGroups(c *gin.Context) {
	subjectID := c.Param("department_id")

	handleSubjectGroups(c, types.DepartmentType, subjectID, false)
}

func handleSubjectGroups(c *gin.Context, subjectType, subjectID string, inherit bool) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "subject_groups")

	subjectPK, err := cacheimpls.GetLocalSubjectPK(subjectType, subjectID)
	if err != nil {
		// 不存在的情况, 404
		if errors.Is(err, sql.ErrNoRows) {
			util.NotFoundJSONResponse(c, "subject not exist")
			return
		}

		util.SystemErrorJSONResponse(c, err)
		return
	}

	// NOTE: group被删除的时候, subjectDetail引用的并不会清理
	subjectDetail, err := cacheimpls.GetSubjectDetail(subjectPK)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	nowUnix := time.Now().Unix()

	// 1. get the subject's groups
	groups := subjectDetail.SubjectGroups
	groupPKs := util.NewFixedLengthInt64Set(len(groups))
	for _, group := range groups {
		// 仅仅在有效期内才需要
		if group.PolicyExpiredAt > nowUnix {
			groupPKs.Add(group.PK)
		}
	}

	// 2. get the subject-department's groups
	deptPKs := subjectDetail.DepartmentPKs
	if inherit && len(deptPKs) > 0 {
		subjectGroups, newErr := cacheimpls.ListSubjectEffectGroups(deptPKs)
		if newErr != nil {
			newErr = errorWrapf(newErr, "ListSubjectEffectGroups deptPKs=`%+v` fail", deptPKs)
			util.SystemErrorJSONResponse(c, newErr)
			return
		}
		for _, sg := range subjectGroups {
			if sg.PolicyExpiredAt > nowUnix {
				groupPKs.Add(sg.PK)
			}
		}
	}

	// 3. build the response
	data := subjectGroupsResponse{}
	for _, pk := range groupPKs.ToSlice() {
		// NOTE: 一个group 被删除, 可能 1min 之内, 还会出现在列表中
		subj, err := cacheimpls.GetSubjectByPK(pk)
		if err != nil {
			// no log
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			// get subject fail, continue
			log.Info(errorWrapf(err, "subject_groups GetSubjectByPK subject_pk=`%d` fail", pk))
			continue
		}

		data = append(data, responseSubject{
			Type: subj.Type,
			ID:   subj.ID,
			Name: subj.Name,
		})
	}

	util.SuccessJSONResponse(c, "ok", data)
}
