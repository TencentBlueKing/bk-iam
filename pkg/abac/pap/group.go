/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pap

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/abac"
	"iam/pkg/abac/pip"
	abacTypes "iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
	"iam/pkg/database"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// GroupCTL ...
const GroupCTL = "GroupCTL"

type GroupController interface {
	GetSubjectGroupCountBeforeExpiredAt(_type, id string, beforeExpiredAt int64) (int64, error)
	GetSubjectSystemGroupCountBeforeExpiredAt(_type, id, systemID string, expiredAt int64) (int64, error)
	ListPagingSubjectGroups(_type, id string, beforeExpiredAt, limit, offset int64) ([]SubjectGroup, error)
	ListPagingSubjectSystemGroups(
		_type, id, systemID string, beforeExpiredAt, limit, offset int64,
	) ([]SubjectGroup, error)
	FilterGroupsHasMemberBeforeExpiredAt(subjects []Subject, expiredAt int64) ([]Subject, error)
	CheckSubjectEffectGroups(_type, id string, groupIDs []string) (map[string]map[string]interface{}, error)

	GetGroupMemberCount(_type, id string) (int64, error)
	ListPagingGroupMember(_type, id string, limit, offset int64) ([]GroupMember, error)
	GetGroupMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error)
	ListPagingGroupMemberBeforeExpiredAt(
		_type, id string, expiredAt int64, limit, offset int64,
	) ([]GroupMember, error)
	GetGroupSubjectCountBeforeExpiredAt(expiredAt int64) (count int64, err error)
	ListPagingGroupSubjectBeforeExpiredAt(expiredAt int64, limit, offset int64) ([]GroupSubject, error)
	GetTemplateGroupMemberCount(_type, id string, templateID int64) (int64, error)
	ListPagingTemplateGroupMember(_type, id string, templateID int64, limit, offset int64) ([]GroupMember, error)

	CreateOrUpdateGroupMembers(_type, id string, members []GroupMember) (map[string]int64, error)
	UpdateGroupMembersExpiredAt(_type, id string, members []GroupMember) error
	DeleteGroupMembers(_type, id string, members []Subject) (map[string]int64, error)
	BulkCreateSubjectTemplateGroup(subjectTemplateGroups []SubjectTemplateGroup) error
	BulkDeleteSubjectTemplateGroup(subjectTemplateGroups []SubjectTemplateGroup) error
	UpdateSubjectTemplateGroupExpiredAt(subjectTemplateGroups []SubjectTemplateGroup) error

	ListRbacGroupByResource(systemID string, resource abacTypes.Resource) ([]Subject, error)
	ListRbacGroupByActionResource(systemID, actionID string, resource abacTypes.Resource) ([]Subject, error)
}

type groupController struct {
	service service.GroupService

	subjectService             service.SubjectService
	groupAlterEventService     service.GroupAlterEventService
	groupResourcePolicyService service.GroupResourcePolicyService
}

func NewGroupController() GroupController {
	return &groupController{
		service:                    service.NewGroupService(),
		subjectService:             service.NewSubjectService(),
		groupAlterEventService:     service.NewGroupAlterEventService(),
		groupResourcePolicyService: service.NewGroupResourcePolicyService(),
	}
}

// GetSubjectGroupCountBeforeExpiredAt ...
func (c *groupController) GetSubjectGroupCountBeforeExpiredAt(
	_type, id string,
	expiredAt int64,
) (count int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetSubjectGroupCountBeforeExpiredAt")
	subjectPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err = c.service.GetSubjectGroupCountBeforeExpiredAt(subjectPK, expiredAt)
	if err != nil {
		return 0, errorWrapf(
			err,
			"service.GetSubjectGroupCountBeforeExpiredAt subjectPK=`%s`, expiredAt=`%d`",
			subjectPK,
			expiredAt,
		)
	}

	return count, nil
}

// GetSubjectSystemGroupCountBeforeExpiredAt ...
func (c *groupController) GetSubjectSystemGroupCountBeforeExpiredAt(
	_type, id string,
	systemID string,
	expiredAt int64,
) (count int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetSubjectSystemGroupCountBeforeExpiredAt")
	subjectPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err = c.service.GetSubjectSystemGroupCountBeforeExpiredAt(subjectPK, systemID, expiredAt)
	if err != nil {
		return 0, errorWrapf(
			err,
			"service.GetSubjectSystemGroupCountBeforeExpiredAt subjectPK=`%s`, systemID=`%s`, expiredAt=`%d`",
			subjectPK,
			systemID,
			expiredAt,
		)
	}

	return count, nil
}

// GetGroupSubjectCountBeforeExpiredAt ...
func (c *groupController) GetGroupSubjectCountBeforeExpiredAt(expiredAt int64) (count int64, err error) {
	return c.service.GetGroupSubjectCountBeforeExpiredAt(expiredAt)
}

func (c *groupController) FilterGroupsHasMemberBeforeExpiredAt(subjects []Subject, expiredAt int64) ([]Subject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "FilterGroupsHasMemberBeforeExpiredAt")

	svcSubjects := convertToServiceSubjects(subjects)
	groupPKs, err := c.subjectService.ListPKsBySubjects(svcSubjects)
	if err != nil {
		return nil, errorWrapf(err, "service.ListPKsBySubjects subjects=`%+v` fail", subjects)
	}

	existGroupPKs, err := c.service.FilterGroupPKsHasMemberBeforeExpiredAt(groupPKs, expiredAt)
	if err != nil {
		return nil, errorWrapf(
			err, "service.FilterGroupPKsHasMemberBeforeExpiredAt groupPKs=`%+v`, expiredAt=`%d` fail",
			groupPKs, expiredAt,
		)
	}

	existSubjects, err := cacheimpls.BatchGetSubjectByPKs(existGroupPKs)
	if err != nil {
		return nil, errorWrapf(
			err, "cacheimpls.BatchGetSubjectByPKs groupPKs=`%+v` fail",
			existGroupPKs,
		)
	}

	existGroups := make([]Subject, 0, len(existGroupPKs))
	for _, subject := range existSubjects {
		existGroups = append(existGroups, Subject{
			Type: subject.Type,
			ID:   subject.ID,
			Name: subject.Name,
		})
	}

	return existGroups, nil
}

func (c *groupController) CheckSubjectEffectGroups(
	_type, id string,
	groupIDs []string,
) (map[string]map[string]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "CheckSubjectExistGroups")

	// subject Type+ID to PK
	subjectPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	groupPKToID := make(map[int64]string, len(groupIDs))
	groupPKs := make([]int64, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		// if groupID is empty, skip
		if groupID == "" {
			continue
		}

		// get the groupPK via groupID
		groupPK, err := cacheimpls.GetLocalSubjectPK(types.GroupType, groupID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.WithError(err).Debugf("cacheimpls.GetSubjectPK type=`group`, id=`%s` fail", groupID)
				continue
			}

			return nil, errorWrapf(
				err,
				"cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail",
				types.GroupType,
				groupID,
			)
		}

		groupPKs = append(groupPKs, groupPK)
		groupPKToID[groupPK] = groupID
	}

	// NOTE: if the performance is a problem, change this to a local cache, key: subjectPK, value int64Set
	subjectGroups, err := c.service.ListEffectSubjectGroupsBySubjectPKGroupPKs(subjectPK, groupPKs)
	if err != nil {
		return nil, errorWrapf(
			err,
			"service.ListEffectSubjectGroupsBySubjectPKGroupPKs subjectPKs=`%d`, groupPKs=`%+v` fail",
			subjectPK,
			groupPKs,
		)
	}

	// the result
	groupIDBelong := make(map[string]map[string]interface{}, len(groupIDs))
	for _, group := range subjectGroups {
		groupID, ok := groupPKToID[group.GroupPK]
		if !ok {
			continue
		}

		groupIDBelong[groupID] = map[string]interface{}{
			"belong":     true,
			"expired_at": group.ExpiredAt,
			"created_at": group.CreatedAt,
		}
	}

	for _, groupID := range groupIDs {
		if _, ok := groupIDBelong[groupID]; !ok {
			groupIDBelong[groupID] = map[string]interface{}{
				"belong":     false,
				"expired_at": 0,
				"created_at": time.Time{},
			}
		}
	}

	return groupIDBelong, nil
}

// ListPagingSubjectGroups ...
func (c *groupController) ListPagingSubjectGroups(
	_type, id string,
	beforeExpiredAt, limit, offset int64,
) ([]SubjectGroup, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingSubjectGroups")
	subjectPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcSubjectGroups, err := c.service.ListPagingSubjectGroups(subjectPK, beforeExpiredAt, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingSubjectGroups subjectPK=`%s`, beforeExpiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			subjectPK, beforeExpiredAt, limit, offset,
		)
	}

	groups, err := convertToSubjectGroups(svcSubjectGroups)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectGroups svcSubjectGroups=`%+v` fail", svcSubjectGroups)
	}

	return groups, nil
}

// ListPagingSubjectSystemGroups ...
func (c *groupController) ListPagingSubjectSystemGroups(
	_type, id string,
	systemID string,
	beforeExpiredAt, limit, offset int64,
) ([]SubjectGroup, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingSubjectSystemGroups")
	subjectPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcSubjectGroups, err := c.service.ListPagingSubjectSystemGroups(
		subjectPK,
		systemID,
		beforeExpiredAt,
		limit,
		offset,
	)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingSubjectSystemGroups "+
				"subjectPK=`%s`, systemID=`%s`, beforeExpiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			subjectPK, systemID, beforeExpiredAt, limit, offset,
		)
	}

	groups, err := convertToSubjectGroups(svcSubjectGroups)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectGroups svcSubjectGroups=`%+v` fail", svcSubjectGroups)
	}

	return groups, nil
}

// GetGroupMemberCount ...
func (c *groupController) GetGroupMemberCount(_type, id string) (int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetGroupMemberCount")
	groupPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err := c.service.GetGroupMemberCount(groupPK)
	if err != nil {
		return 0, errorWrapf(err, "service.GetGroupMemberCount groupPK=`%d`", groupPK)
	}

	return count, nil
}

// ListPagingGroupMember ...
func (c *groupController) ListPagingGroupMember(_type, id string, limit, offset int64) ([]GroupMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingGroupMember")
	groupPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcMembers, err := c.service.ListPagingGroupMember(groupPK, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingGroupMember groupPK=`%d`, limit=`%d`, offset=`%d` fail",
			groupPK, limit, offset,
		)
	}

	members, err := convertToGroupMembers(svcMembers)
	if err != nil {
		return nil, errorWrapf(err, "convertToGroupMembers svcMembers=`%+v` fail", svcMembers)
	}

	return members, nil
}

// GetTemplateGroupMemberCount ...
func (c *groupController) GetTemplateGroupMemberCount(_type, id string, templateID int64) (int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetTemplateGroupMemberCount")
	groupPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(
			err,
			"cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s`, templateID=`%d` fail",
			_type,
			id,
			templateID,
		)
	}

	count, err := c.service.GetTemplateGroupMemberCount(groupPK, templateID)
	if err != nil {
		return 0, errorWrapf(
			err,
			"service.GetTemplateGroupMemberCount groupPK=`%d`, templateID=`%d`",
			groupPK,
			templateID,
		)
	}

	return count, nil
}

// ListPagingTemplateGroupMember ...
func (c *groupController) ListPagingTemplateGroupMember(
	_type, id string,
	templateID int64,
	limit, offset int64,
) ([]GroupMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingTemplateGroupMember")
	groupPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(
			err,
			"cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s`, templateID=`%d` fail",
			_type,
			id,
			templateID,
		)
	}

	svcMembers, err := c.service.ListPagingTemplateGroupMember(groupPK, templateID, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingTemplateGroupMember groupPK=`%d`, templateID=`%d`, limit=`%d`, offset=`%d` fail",
			groupPK, templateID, limit, offset,
		)
	}

	members, err := convertToGroupMembers(svcMembers)
	if err != nil {
		return nil, errorWrapf(err, "convertToGroupMembers svcMembers=`%+v` fail", svcMembers)
	}

	return members, nil
}

// ListPagingGroupSubjectBeforeExpiredAt ...
func (c *groupController) ListPagingGroupSubjectBeforeExpiredAt(
	expiredAt int64,
	limit, offset int64,
) ([]GroupSubject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingGroupSubjectBeforeExpiredAt")

	svcRelations, err := c.service.ListPagingGroupSubjectBeforeExpiredAt(expiredAt, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingGroupSubjectBeforeExpiredAt expiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			expiredAt, limit, offset,
		)
	}

	relations, err := convertToGroupSubjects(svcRelations)
	if err != nil {
		return nil, errorWrapf(err, "convertToGroupSubjects svcRelations=`%+v` fail", svcRelations)
	}

	return relations, nil
}

// GetGroupMemberCountBeforeExpiredAt ...
func (c *groupController) GetGroupMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetGroupMemberCountBeforeExpiredAt")
	groupPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err := c.service.GetGroupMemberCountBeforeExpiredAt(groupPK, expiredAt)
	if err != nil {
		return 0, errorWrapf(
			err, "service.GetGroupMemberCountBeforeExpiredAt groupPK=`%d`, expiredAt=`%d`",
			groupPK, expiredAt,
		)
	}

	return count, nil
}

// ListPagingGroupMemberBeforeExpiredAt ...
func (c *groupController) ListPagingGroupMemberBeforeExpiredAt(
	_type, id string, expiredAt int64, limit, offset int64,
) ([]GroupMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingGroupMemberBeforeExpiredAt")
	groupPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcMembers, err := c.service.ListPagingGroupMemberBeforeExpiredAt(groupPK, expiredAt, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err,
			"service.ListPagingGroupMemberBeforeExpiredAt groupPK=`%d`, expiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			groupPK,
			expiredAt,
			limit,
			offset,
		)
	}

	members, err := convertToGroupMembers(svcMembers)
	if err != nil {
		return nil, errorWrapf(err, "convertToGroupMembers svcMembers=`%+v` fail", svcMembers)
	}

	return members, nil
}

type subjectGroupHelper struct {
	service service.GroupService

	groupSystemCache map[int64]string
}

func newSubjectGroupHelper(service service.GroupService) *subjectGroupHelper {
	return &subjectGroupHelper{
		service:          service,
		groupSystemCache: map[int64]string{},
	}
}

func (h *subjectGroupHelper) getSubjectGroup(
	subjectPK, groupPK int64,
) (authorized bool, subjectGroup *types.ThinSubjectGroup, err error) {
	systemID, ok := h.groupSystemCache[groupPK]
	if !ok {
		systemID, err = h.service.GetGroupOneAuthSystem(groupPK)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return false, nil, err
		}
		h.groupSystemCache[groupPK] = systemID
	}

	if systemID == "" {
		return false, nil, nil
	}

	// subject - system的groups列表
	subjectGroups, err := cacheimpls.GetSubjectSystemGroup(systemID, subjectPK)
	if err != nil {
		return false, nil, err
	}

	for _, group := range subjectGroups {
		if group.GroupPK == groupPK {
			return true, &group, nil
		}
	}

	return true, nil, nil
}

func (c *groupController) BulkCreateSubjectTemplateGroup(subjectTemplateGroups []SubjectTemplateGroup) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "BulkCreateSubjectTemplateGroup")

	relations, err := c.convertToSubjectTemplateGroups(subjectTemplateGroups)
	if err != nil {
		return errorWrapf(err, "convertToSubjectTemplateGroups subjectTemplateGroups=`%+v` fail", subjectTemplateGroups)
	}

	subjectGroupHelper := newSubjectGroupHelper(c.service)

	for i := range relations {
		relation := &relations[i]

		authorized, subjectGroup, err := subjectGroupHelper.getSubjectGroup(relation.SubjectPK, relation.GroupPK)
		if err != nil {
			return errorWrapf(
				err,
				"getSubjectGroup subjectPK=`%d`, groupPK=`%d` fail",
				relation.SubjectPK,
				relation.GroupPK,
			)
		}

		// 1. 如果group未授权, 不需要更新
		if !authorized {
			continue
		}

		// 2. 如果已授权并且过期时间大于当前时间, 不需要更新
		if subjectGroup != nil && subjectGroup.ExpiredAt > relation.ExpiredAt {
			relation.ExpiredAt = subjectGroup.ExpiredAt
			continue
		}

		// 3. 其余场景需要更新
		relation.NeedUpdate = true
	}

	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		return errorWrapf(err, "define tx error")
	}

	// 创建subject template group
	err = c.service.BulkCreateSubjectTemplateGroupWithTx(tx, relations)
	if err != nil {
		return errorWrapf(
			err, "service.BulkCreateSubjectTemplateGroupWithTx relations=`%+v` fail", relations,
		)
	}

	// 更新除了subject system group之外的subject group
	err = c.updateSubjectGroupExpiredAtWithTx(tx, relations, true)
	if err != nil {
		return errorWrapf(err, "updateSubjectGroupExpiredAtWithTx relations=`%+v` fail", relations)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx commit error")
	}

	// 清理subject system group 缓存
	c.deleteSubjectTemplateGroupCache(relations)

	return nil
}

func (c *groupController) UpdateSubjectTemplateGroupExpiredAt(subjectTemplateGroups []SubjectTemplateGroup) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "BulkRenewSubjectTemplateGroup")

	relations, err := c.convertToSubjectTemplateGroups(subjectTemplateGroups)
	if err != nil {
		return errorWrapf(err, "convertToSubjectTemplateGroups subjectTemplateGroups=`%+v` fail", subjectTemplateGroups)
	}

	subjectGroupHelper := newSubjectGroupHelper(c.service)

	noAuthorizedRelations := make([]types.SubjectTemplateGroup, 0, len(relations))
	for i := range relations {
		relation := &relations[i]

		authorized, subjectGroup, err := subjectGroupHelper.getSubjectGroup(relation.SubjectPK, relation.GroupPK)
		if err != nil {
			return errorWrapf(
				err,
				"getSubjectGroup subjectPK=`%d`, groupPK=`%d` fail",
				relation.SubjectPK,
				relation.GroupPK,
			)
		}

		// 1. 如果group未授权
		if !authorized {
			noAuthorizedRelations = append(noAuthorizedRelations, *relation)
			continue
		}

		// 2. 如果已授权并且过期时间大于当前时间, 不需要更新
		if subjectGroup != nil && subjectGroup.ExpiredAt > relation.ExpiredAt {
			continue
		}

		// 3. 其余场景需要更新
		relation.NeedUpdate = true
	}

	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		return errorWrapf(err, "define tx error")
	}

	if len(noAuthorizedRelations) > 0 {
		// 只更新subject template group的过期时间
		err = c.service.UpdateSubjectTemplateGroupExpiredAtWithTx(tx, noAuthorizedRelations)
		if err != nil {
			return errorWrapf(
				err, "service.UpdateSubjectTemplateGroupExpiredAtWithTx relations=`%+v` fail", noAuthorizedRelations,
			)
		}
	}

	// 更新subject system group
	err = c.service.BulkUpdateSubjectSystemGroupBySubjectTemplateGroupWithTx(tx, relations)
	if err != nil {
		return errorWrapf(
			err,
			"service.BulkUpdateSubjectSystemGroupBySubjectTemplateGroupWithTx relations=`%+v` fail",
			relations,
		)
	}

	// 更新除了subject system group之外的subject group
	err = c.updateSubjectGroupExpiredAtWithTx(tx, relations, true)
	if err != nil {
		return errorWrapf(err, "updateSubjectGroupExpiredAtWithTx relations=`%+v` fail", relations)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx commit error")
	}

	// 清理subject system group 缓存
	c.deleteSubjectTemplateGroupCache(relations)

	return nil
}

func (*groupController) convertToSubjectTemplateGroups(
	subjectTemplateGroups []SubjectTemplateGroup,
) ([]types.SubjectTemplateGroup, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "convertToSubjectTemplateGroups")

	relations := make([]types.SubjectTemplateGroup, 0, len(subjectTemplateGroups))
	for _, stg := range subjectTemplateGroups {
		subjectPK, err := cacheimpls.GetLocalSubjectPK(stg.Type, stg.ID)
		if err != nil {
			return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", stg.Type, stg.ID)
		}
		groupPK, err := cacheimpls.GetLocalSubjectPK(types.SubjectTypeGroup, strconv.FormatInt(stg.GroupID, 10))
		if err != nil {
			return nil, errorWrapf(
				err,
				"cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail",
				types.SubjectTypeGroup,
				strconv.FormatInt(stg.GroupID, 10),
			)
		}

		relations = append(relations, types.SubjectTemplateGroup{
			SubjectPK:  subjectPK,
			TemplateID: stg.TemplateID,
			GroupPK:    groupPK,
			ExpiredAt:  stg.ExpiredAt,
		})
	}
	return relations, nil
}

func (c *groupController) BulkDeleteSubjectTemplateGroup(subjectTemplateGroups []SubjectTemplateGroup) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "BulkDeleteSubjectTemplateGroup")

	relations, err := c.convertToSubjectTemplateGroups(subjectTemplateGroups)
	if err != nil {
		return errorWrapf(err, "convertToSubjectTemplateGroups subjectTemplateGroups=`%+v` fail", subjectTemplateGroups)
	}

	now := time.Now().Unix()
	for i := range relations {
		relation := &relations[i]

		expiredAt, err := c.service.GetMaxExpiredAtBySubjectGroup(
			relation.SubjectPK,
			relation.GroupPK,
			relation.TemplateID,
		)
		if err != nil && !errors.Is(err, service.ErrGroupMemberNotFound) {
			return errorWrapf(
				err, "GetMaxExpiredAtBySubjectGroup subjectPK=`%d`, groupPK=`%d` fail",
			)
		}

		// 如果有其它的关系, 并且过期时间大于当前时间, 不需要删除
		if err == nil && expiredAt > now {
			continue
		}

		relation.NeedUpdate = true
	}

	// 如果没有其他关系了需要删除subject system group数据
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		return errorWrapf(err, "define tx error")
	}

	err = c.service.BulkDeleteSubjectTemplateGroupWithTx(tx, relations)
	if err != nil {
		return errorWrapf(
			err, "service.BulkDeleteSubjectTemplateGroupWithTx relations=`%+v` fail", relations,
		)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx commit error")
	}

	// 清理subject system group 缓存
	c.deleteSubjectTemplateGroupCache(relations)

	return nil
}

func (c *groupController) deleteSubjectTemplateGroupCache(relations []types.SubjectTemplateGroup) {
	groupSubjects := map[int64][]int64{}
	for _, relation := range relations {
		if !relation.NeedUpdate {
			continue
		}

		if _, ok := groupSubjects[relation.GroupPK]; !ok {
			groupSubjects[relation.GroupPK] = []int64{}
		}
		groupSubjects[relation.GroupPK] = append(groupSubjects[relation.GroupPK], relation.SubjectPK)
	}

	for groupPK, subjectPKs := range groupSubjects {
		c.createGroupAlterEvent(groupPK, subjectPKs)

		cacheimpls.BatchDeleteSubjectAuthSystemGroupCache(subjectPKs, groupPK)
	}
}

// CreateOrUpdateGroupMembers ...
func (c *groupController) CreateOrUpdateGroupMembers(
	_type, id string,
	members []GroupMember,
) (typeCount map[string]int64, err error) {
	return c.alterGroupMembers(_type, id, members, true)
}

func (c *groupController) convertGroupMembersToSubjectTemplateGroups(
	groupPK int64,
	members []GroupMember,
) ([]types.SubjectTemplateGroup, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "convertGroupMembersToSubjectTemplateGroups")

	relations := make([]types.SubjectTemplateGroup, 0, len(members))
	for _, m := range members {
		subjectPK, err := cacheimpls.GetLocalSubjectPK(m.Type, m.ID)
		if err != nil {
			return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", m.Type, m.ID)
		}

		relations = append(relations, types.SubjectTemplateGroup{
			GroupPK:   groupPK,
			SubjectPK: subjectPK,
			ExpiredAt: m.ExpiredAt,
		})
	}

	return relations, nil
}

func (c *groupController) alterGroupMembers(
	_type, id string,
	members []GroupMember,
	createIfNotExists bool,
) (typeCount map[string]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "alterGroupMembers")
	groupPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	relations, err := c.service.ListGroupMember(groupPK)
	if err != nil {
		err = errorWrapf(err, "service.ListGroupMember type=`%s` id=`%s`", _type, id)
		return
	}

	// 重复和已经存在DB里的不需要
	memberMap := make(map[int64]types.GroupMember, len(relations))
	for _, m := range relations {
		memberMap[m.SubjectPK] = m
	}

	// 获取实际需要添加的member
	createMembers := make([]types.SubjectTemplateGroup, 0, len(members))

	// 需要更新过期时间的member
	updateMembers := make([]types.SubjectTemplateGroup, 0, len(members))

	typeCount = map[string]int64{
		types.UserType:       0,
		types.DepartmentType: 0,
	}

	subjectTemplateGroups, err := c.convertGroupMembersToSubjectTemplateGroups(groupPK, members)
	if err != nil {
		return nil, err
	}

	subjectGroupHelper := newSubjectGroupHelper(c.service)
	for i := range subjectTemplateGroups {
		relation := &subjectTemplateGroups[i]

		// 查询 subject group 已有的关系
		authorized, subjectGroup, err := subjectGroupHelper.getSubjectGroup(relation.SubjectPK, groupPK)
		if err != nil {
			return nil, errorWrapf(
				err,
				"getSubjectGroup subjectPK=`%d`, groupPK=`%d` fail",
				relation.SubjectPK,
				groupPK,
			)
		}

		if authorized && subjectGroup != nil && subjectGroup.ExpiredAt > relation.ExpiredAt {
			relation.ExpiredAt = subjectGroup.ExpiredAt
		}

		// member已存在则不再添加
		if oldMember, ok := memberMap[relation.SubjectPK]; ok {
			// 如果过期时间大于已有的时间, 则更新过期时间
			if relation.ExpiredAt > oldMember.ExpiredAt {
				relation.NeedUpdate = true

				updateMembers = append(updateMembers, *relation)
			}
			continue
		}

		if createIfNotExists {
			if authorized && (subjectGroup == nil || subjectGroup.ExpiredAt < relation.ExpiredAt) {
				relation.NeedUpdate = true
			}

			createMembers = append(createMembers, *relation)
			typeCount[members[i].Type]++
		}
	}

	// 按照PK删除Subject所有相关的
	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		return nil, errorWrapf(err, "define tx error")
	}

	if len(updateMembers) != 0 {
		// 更新成员过期时间
		err = c.service.UpdateGroupMembersExpiredAtWithTx(tx, groupPK, updateMembers)
		if err != nil {
			err = errorWrapf(err, "service.UpdateGroupMembersExpiredAtWithTx members=`%+v`", updateMembers)
			return
		}
	}

	// 无成员可添加，直接返回
	if createIfNotExists && len(createMembers) != 0 {
		// 添加成员
		err = c.service.BulkCreateGroupMembersWithTx(tx, groupPK, createMembers)
		if err != nil {
			err = errorWrapf(err, "service.BulkCreateGroupMembersWithTx relations=`%+v`", createMembers)
			return nil, err
		}
	}

	// 更新subject template group过期时间
	err = c.updateSubjectGroupExpiredAtWithTx(tx, subjectTemplateGroups, false)
	if err != nil {
		err = errorWrapf(err, "updateSubjectGroupExpiredAtWithTx subjectTemplateGroups=`%+v`", subjectTemplateGroups)
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, errorWrapf(err, "tx commit error")
	}

	// 清理subject system group 缓存
	c.deleteSubjectTemplateGroupCache(subjectTemplateGroups)

	return typeCount, nil
}

func (c *groupController) updateSubjectGroupExpiredAtWithTx(
	tx *sqlx.Tx,
	subjectTemplateGroups []types.SubjectTemplateGroup,
	updateGroupRelation bool,
) error {
	needUpdateRelations := make([]types.SubjectTemplateGroup, 0, len(subjectTemplateGroups))
	for _, relation := range subjectTemplateGroups {
		if relation.NeedUpdate {
			needUpdateRelations = append(needUpdateRelations, relation)
		}
	}

	if len(needUpdateRelations) != 0 {
		return c.service.UpdateSubjectGroupExpiredAtWithTx(tx, needUpdateRelations, updateGroupRelation)
	}
	return nil
}

// UpdateGroupMembersExpiredAt ...
func (c *groupController) UpdateGroupMembersExpiredAt(_type, id string, members []GroupMember) (err error) {
	_, err = c.alterGroupMembers(_type, id, members, false)
	return
}

// DeleteGroupMembers ...
func (c *groupController) DeleteGroupMembers(
	_type, id string,
	members []Subject,
) (typeCount map[string]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "DeleteGroupMembers")

	userPKs := make([]int64, 0, len(members))
	departmentPKs := make([]int64, 0, len(members))
	for _, m := range members {
		pk, err := cacheimpls.GetLocalSubjectPK(m.Type, m.ID)
		if err != nil {
			return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", m.Type, m.ID)
		}

		if m.Type == types.UserType {
			userPKs = append(userPKs, pk)
		} else if m.Type == types.DepartmentType {
			departmentPKs = append(departmentPKs, pk)
		}
	}

	groupPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	typeCount, err = c.service.BulkDeleteGroupMembers(groupPK, userPKs, departmentPKs)
	if err != nil {
		return nil, errorWrapf(
			err, "service.BulkDeleteGroupMembers groupPK=`%d`, userPKs=`%+v`, departmentPKs=`%+v` failed",
			groupPK, userPKs, departmentPKs,
		)
	}

	// 清理缓存
	subjectPKs := make([]int64, 0, len(members))
	subjectPKs = append(subjectPKs, userPKs...)
	subjectPKs = append(subjectPKs, departmentPKs...)

	// 创建group_alter_event
	c.createGroupAlterEvent(groupPK, subjectPKs)

	// group auth system
	cacheimpls.BatchDeleteSubjectAuthSystemGroupCache(subjectPKs, groupPK)

	return typeCount, nil
}

func (c *groupController) createGroupAlterEvent(groupPK int64, subjectPKs []int64) {
	err := c.groupAlterEventService.CreateByGroupSubject(groupPK, subjectPKs)
	if err != nil {
		log.WithError(err).
			Errorf("groupAlterEventService.CreateByGroupSubject groupPK=%d subjectPKs=%v fail", groupPK, subjectPKs)

		// report to sentry
		util.ReportToSentry("createGroupAlterEvent groupAlterEventService.CreateByGroupSubject fail",
			map[string]interface{}{
				"layer":      GroupCTL,
				"groupPK":    groupPK,
				"subjectPKs": subjectPKs,
				"error":      err.Error(),
			},
		)
	}
}

// ListRbacGroupByResource ...
func (c *groupController) ListRbacGroupByResource(systemID string, resource abacTypes.Resource) ([]Subject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListRbacGroupByResource")

	// 解析资源实例信息
	resourceNodes, err := abac.ParseResourceNode(resource)
	if err != nil {
		err = errorWrapf(
			err, "abac.ParseResourceNode resource=`%+v`",
			resource,
		)
		return nil, err
	}

	// 没有操作筛选的情况下选择最后一个资源类型的类型pk
	actionResourceTypePK := resourceNodes[len(resourceNodes)-1].TypePK

	groupPKset := set.NewInt64Set()
	// 查询有权限的用户组
	for _, resourceNode := range resourceNodes {
		actionGroupPKs, err := c.groupResourcePolicyService.GetAuthorizedActionGroupMap(
			systemID, actionResourceTypePK, resourceNode.TypePK, resourceNode.ID,
		)
		if err != nil {
			err = errorWrapf(
				err,
				"svc.GetAuthorizedActionGroupMap fail, system=`%s`, resource=`%+v`",
				systemID,
				resourceNode,
			)
			return nil, err
		}

		for _, groupPKs := range actionGroupPKs {
			groupPKset.Append(groupPKs...)
		}
	}

	// 查询用户组信息
	groups, err := groupPKsToSubjects(groupPKset.ToSlice())
	if err != nil {
		err = errorWrapf(
			err,
			"groupPKsToSubjects fail",
		)
		return nil, err
	}
	return groups, nil
}

// ListRbacGroupByResource ...
func (c *groupController) ListRbacGroupByActionResource(
	systemID, actionID string,
	resource abacTypes.Resource,
) ([]Subject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListRbacGroupByActionResource")

	// 查询操作相关的信息
	actionPK, authType, actionResourceTypes, err := pip.GetActionDetail(systemID, actionID)
	if err != nil {
		err = errorWrapf(
			err, "pip.GetActionDetail systemID=`%s`, actionID=`%s`",
			systemID, actionID,
		)
		return nil, err
	}

	if authType != types.AuthTypeRBAC {
		return nil, errorWrapf(errors.New("only support rbac"), "authType=`%d`", authType)
	}

	// 查询操作关联的资源类型id
	actionResourceTypePK, err := cacheimpls.GetLocalResourceTypePK(
		actionResourceTypes[0].System, actionResourceTypes[0].Type, // 配置rbac的操作一定只关联了1个资源类型
	)
	if err != nil {
		err = errorWrapf(
			err, "cacheimpls.GetLocalResourceTypePK systemID=`%s`, resourceType=`%s`",
			actionResourceTypes[0].System, actionResourceTypes[0].Type,
		)
		return nil, err
	}

	// 解析资源实例信息
	resourceNodes, err := abac.ParseResourceNode(resource)
	if err != nil {
		err = errorWrapf(
			err, "abac.ParseResourceNode resource=`%+v`",
			resource,
		)
		return nil, err
	}

	groupPKset := set.NewInt64Set()
	// 查询有权限的用户组
	for _, resourceNode := range resourceNodes {
		groupPKs, err := cacheimpls.GetResourceActionAuthorizedGroupPKs(
			systemID,
			actionPK,
			actionResourceTypePK,
			resourceNode.TypePK,
			resourceNode.ID,
		)
		if err != nil {
			err = errorWrapf(
				err,
				"cacheimpls.GetResourceActionAuthorizedGroupPKs fail, system=`%s` action_id=`%s` resource=`%+v`",
				systemID,
				actionID,
				resourceNode,
			)
			return nil, err
		}

		groupPKset.Append(groupPKs...)
	}

	// 查询用户组信息
	groups, err := groupPKsToSubjects(groupPKset.ToSlice())
	if err != nil {
		err = errorWrapf(
			err,
			"groupPKsToSubjects fail",
		)
		return nil, err
	}
	return groups, nil
}

func groupPKsToSubjects(groupPKs []int64) ([]Subject, error) {
	subjects, err := cacheimpls.BatchGetSubjectByPKs(groupPKs)
	if err != nil {
		return nil, fmt.Errorf("cacheimpls.BatchGetSubjectByPKs fail, subjectPKs=`%v`", groupPKs)
	}

	groups := make([]Subject, 0, len(groupPKs))
	for _, subject := range subjects {
		groups = append(groups, Subject{
			Type: subject.Type,
			ID:   subject.ID,
			Name: subject.Name,
		})
	}
	return groups, nil
}

func convertToSubjectGroups(svcSubjectGroups []types.SubjectGroup) ([]SubjectGroup, error) {
	groupPKs := make([]int64, 0, len(svcSubjectGroups))
	for _, m := range svcSubjectGroups {
		groupPKs = append(groupPKs, m.GroupPK)
	}

	subjects, err := cacheimpls.BatchGetSubjectByPKs(groupPKs)
	if err != nil {
		return nil, err
	}

	subjectMap := make(map[int64]types.Subject, len(subjects))
	for _, subject := range subjects {
		subjectMap[subject.PK] = subject
	}

	groups := make([]SubjectGroup, 0, len(svcSubjectGroups))
	for _, m := range svcSubjectGroups {
		subject, ok := subjectMap[m.GroupPK]
		if !ok {
			continue
		}

		groups = append(groups, SubjectGroup{
			PK:        m.PK,
			Type:      subject.Type,
			ID:        subject.ID,
			ExpiredAt: m.ExpiredAt,
			CreatedAt: m.CreatedAt,
		})
	}

	return groups, nil
}

func convertToGroupMembers(svcGroupMembers []types.GroupMember) ([]GroupMember, error) {
	subjectPKs := make([]int64, 0, len(svcGroupMembers))
	for _, m := range svcGroupMembers {
		subjectPKs = append(subjectPKs, m.SubjectPK)
	}
	subjects, err := cacheimpls.BatchGetSubjectByPKs(subjectPKs)
	if err != nil {
		return nil, err
	}

	subjectMap := make(map[int64]types.Subject, len(subjects))
	for _, subject := range subjects {
		subjectMap[subject.PK] = subject
	}

	members := make([]GroupMember, 0, len(svcGroupMembers))
	for _, m := range svcGroupMembers {
		subject, ok := subjectMap[m.SubjectPK]
		if !ok {
			continue
		}

		members = append(members, GroupMember{
			PK:        m.PK,
			Type:      subject.Type,
			ID:        subject.ID,
			Name:      subject.Name,
			ExpiredAt: m.ExpiredAt,
			CreatedAt: m.CreatedAt,
		})
	}

	return members, nil
}

func convertToGroupSubjects(svcGroupSubjects []types.GroupSubject) ([]GroupSubject, error) {
	subjectPKs := set.NewInt64Set()
	for _, m := range svcGroupSubjects {
		subjectPKs.Add(m.SubjectPK)
		subjectPKs.Add(m.GroupPK)
	}

	subjects, err := cacheimpls.BatchGetSubjectByPKs(subjectPKs.ToSlice())
	if err != nil {
		return nil, err
	}

	subjectMap := make(map[int64]types.Subject, len(subjects))
	for _, subject := range subjects {
		subjectMap[subject.PK] = subject
	}

	groupSubjects := make([]GroupSubject, 0, len(svcGroupSubjects))
	for _, m := range svcGroupSubjects {
		subject, ok := subjectMap[m.SubjectPK]
		if !ok {
			continue
		}

		group, ok := subjectMap[m.GroupPK]
		if !ok {
			continue
		}

		groupSubjects = append(groupSubjects, GroupSubject{
			Subject: Subject{
				Type: subject.Type,
				ID:   subject.ID,
				Name: subject.Name,
			},
			Group: Subject{
				Type: group.Type,
				ID:   group.ID,
				Name: group.Name,
			},
			ExpiredAt: m.ExpiredAt,
		})
	}

	return groupSubjects, nil
}
