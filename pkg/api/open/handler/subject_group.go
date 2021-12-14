package handler

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cacheimpls"
	"iam/pkg/errorx"
	"iam/pkg/util"
)

// TODO:
//    1. url怎么设置?
//    2. 直接插 subject-relation表好, 还是全部走缓存?
//    3. group被删除的时候, subjectDetail引用的并不会清理 => 一个group 被删除, 可能 1min 之内, 还会出现在列表中

// SubjectGroups godoc
// @Summary subject group
// @Description get a subject's groups, include the inherit groups from department
// @ID api-open-subject-groups-get
// @Tags open
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param subject_type path string true "Subject Type"
// @Param subject_id path string true "Subject ID"
// @Param inherit query bool true "get subject's inherit groups from it's departments"
// @Success 200 {object} util.Response{data=subjectGroupsResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/subjects/{subject_type}/{subject_id}/groups [get]
func SubjectGroups(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "subject_groups")

	var pathParams subjectGroupsSerializer
	if err := c.ShouldBindUri(&pathParams); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	_, inherit := c.GetQuery("inherit")

	subjectPK, err := cacheimpls.GetLocalSubjectPK(pathParams.SubjectType, pathParams.SubjectID)
	if err != nil {
		// 不存在的情况, 404
		if errors.Is(err, sql.ErrNoRows) {
			util.NotFoundJSONResponse(c, "subject not exist")
			return
		}

		util.SystemErrorJSONResponse(c, err)
		return
	}

	// NOTE: group被删除的时候, subjectDetail引用的并不会清理=> how to?
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
