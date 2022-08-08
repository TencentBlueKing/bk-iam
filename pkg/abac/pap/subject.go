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
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pap/event"
	pl "iam/pkg/abac/prp/policy"
	"iam/pkg/cacheimpls"
	"iam/pkg/database"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/task"
	"iam/pkg/task/producer"
	"iam/pkg/util"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// SubjectCTL ...
const SubjectCTL = "SubjectCTL"

type SubjectController interface {
	BulkCreate(subjects []Subject) error
	BulkUpdateName(subjects []Subject) error
	BulkDeleteUserAndDepartment(subjects []Subject) error
	BulkDeleteGroup(subjects []Subject) error
}

type subjectController struct {
	service service.SubjectService

	// 以下manager都是为了BulkDelete, 删除subject时同时删除相关数据
	groupService                      service.GroupService
	departmentService                 service.DepartmentService
	policyService                     service.PolicyService
	subjectBlackListService           service.SubjectBlackListService
	subjectActionExpressionService    service.SubjectActionExpressionService
	subjectActionGroupResourceService service.SubjectActionGroupResourceService
	groupResourcePolicyService        service.GroupResourcePolicyService
	groupAlterEventService            service.GroupAlterEventService
	alterEventProducer                producer.Producer

	subjectEventProducer event.SubjectEventProducer
}

func NewSubjectController() SubjectController {
	return &subjectController{
		service: service.NewSubjectService(),

		groupService:                      service.NewGroupService(),
		departmentService:                 service.NewDepartmentService(),
		policyService:                     service.NewPolicyService(),
		subjectBlackListService:           service.NewSubjectBlackListService(),
		subjectActionExpressionService:    service.NewSubjectActionExpressionService(),
		subjectActionGroupResourceService: service.NewSubjectActionGroupResourceService(),
		groupResourcePolicyService:        service.NewGroupResourcePolicyService(),
		groupAlterEventService:            service.NewGroupAlterEventService(),
		alterEventProducer:                producer.NewRedisProducer(task.GetRbacEventQueue()),
		subjectEventProducer:              event.NewSubjectEventProducer(),
	}
}

// BulkCreate ...
func (c *subjectController) BulkCreate(subjects []Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectCTL, "BulkDelete")

	svcSubjects := convertToServiceSubjects(subjects)

	err := c.service.BulkCreate(svcSubjects)
	if err != nil {
		return errorWrapf(err, "service.BulkCreate subjects=`%+v` failed", svcSubjects)
	}

	return nil
}

// BulkUpdateName ...
func (c *subjectController) BulkUpdateName(subjects []Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectCTL, "BulkUpdateName")

	svcSubjects := convertToServiceSubjects(subjects)

	err := c.service.BulkUpdateName(svcSubjects)
	if err != nil {
		return errorWrapf(err, "service.BulkUpdateName subjects=`%+v` failed", svcSubjects)
	}

	return nil
}

func (c *subjectController) BulkDeleteGroup(subjects []Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectCTL, "BulkDeleteGroup")

	svcSubjects := convertToServiceSubjects(subjects)

	pks, err := c.service.ListPKsBySubjects(svcSubjects)
	if err != nil {
		return errorWrapf(err, "service.ListPKsBySubjects subjects=`%+v` fail", subjects)
	}

	// 按照PK删除Subject所有相关的
	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		return errorWrapf(err, "define tx error")
	}

	// 1. 删除策略
	err = c.policyService.BulkDeleteBySubjectPKsWithTx(tx, pks) // user, group, department
	if err != nil {
		return errorWrapf(err, "policyService.BulkDeleteBySubjectPKsWithTx pks=`%+v` failed", pks)
	}

	// 2. 删除subject relation
	eventMessages := make([]string, 0, len(pks)*2)
	for _, pk := range pks { // only group
		// 2.1 删除subject system group/group system auth type
		// TODO: 同步删除, 数据量大的情况下, 会很慢, 需要优化
		systemIDs, err := c.groupService.ListGroupAuthSystemIDs(pk)
		if err != nil {
			return errorWrapf(err, "groupService.ListGroupAuthSystemIDs pk=`%+v` failed", pk)
		}

		for _, systemID := range systemIDs {
			_, err = c.groupService.AlterGroupAuthType(tx, systemID, pk, types.AuthTypeNone)
			if err != nil {
				return errorWrapf(err, "groupService.AlterGroupAuthType systemID=`%+v` pk=`%+v` failed", systemID, pk)
			}
		}

		// 2.2 生成删除 subject action group resource/subject action expression 事件
		members, err := c.groupService.ListGroupMember(pk)
		if err != nil {
			return errorWrapf(err, "groupService.ListGroupMember pk=`%+v` failed", pk)
		}

		if len(members) == 0 {
			continue
		}

		memberPKs := make([]int64, 0, len(members))
		for _, m := range members {
			memberPKs = append(memberPKs, m.SubjectPK)
		}

		// 生成变更信息
		eventPKs, err := c.groupAlterEventService.CreateByGroupSubject(pk, memberPKs)
		if err != nil {
			return errorWrapf(
				err,
				"groupAlterEventService.CreateByGroupSubject groupPK=`%+v` memberPKs=`%+v` failed",
				pk,
				memberPKs,
			)
		}

		if len(eventPKs) == 0 {
			continue
		}

		eventMessages = append(eventMessages, util.Int64SliceToStringSlice(eventPKs)...)
	}

	// 2.3 删除subject relation
	err = c.groupService.BulkDeleteByGroupPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "groupService.BulkDeleteByGroupPKsWithTx pks=`%+v` failed", pks)
	}

	// 3. 删除subject
	err = c.service.BulkDeleteByPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "service.BulkDeleteByPKsWithTx pks=`%+v` failed", pks)
	}

	// 4. 删除group resource policy
	err = c.groupResourcePolicyService.BulkDeleteByGroupPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "groupResourcePolicyService.BulkDeleteByGroupPKsWithTx groupPKs=`%+v` failed", pks)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx commit error")
	}

	// 发送group 变更事件
	if len(eventMessages) != 0 {
		go c.alterEventProducer.Publish(eventMessages...)
	}

	// 发送事件
	c.subjectEventProducer.PublishDeleteEvent(svcSubjects)

	for _, s := range subjects {
		cacheimpls.DeleteSubjectPK(s.Type, s.ID)
		cacheimpls.DeleteLocalSubjectPK(s.Type, s.ID)
	}

	// Note: 不需要清除subject的成员其对应的SubjectGroup和SubjectDepartment，
	//       =>  保证拿到的group pk 没有对应的policy cache/回源也查不到
	deleteGroupPKPolicyCache(pks)

	// 发送事件
	c.subjectEventProducer.PublishDeleteEvent(svcSubjects)

	return nil
}

// BulkDeleteUserAndDepartment ...
func (c *subjectController) BulkDeleteUserAndDepartment(subjects []Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectCTL, "BulkDeleteUserAndDepartment")

	svcSubjects := convertToServiceSubjects(subjects)

	pks, err := c.service.ListPKsBySubjects(svcSubjects)
	if err != nil {
		return errorWrapf(err, "service.ListPKsBySubjects subjects=`%+v` fail", subjects)
	}

	// 按照PK删除Subject所有相关的
	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		return errorWrapf(err, "define tx error")
	}

	// 1. 删除策略
	err = c.policyService.BulkDeleteBySubjectPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "policyService.BulkDeleteBySubjectPKsWithTx pks=`%+v` failed", pks)
	}

	// 2. 删除subject relation
	err = c.groupService.BulkDeleteBySubjectPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "groupService.BulkDeleteBySubjectPKsWithTx pks=`%+v` failed", pks)
	}

	// 3. 删除subject department
	err = c.departmentService.BulkDeleteBySubjectPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "departmentService.BulkDeleteBySubjectPKsWithTx pks=`%+v` failed", pks)
	}

	// 4. 删除subject
	err = c.service.BulkDeleteByPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "service.BulkDeleteByPKsWithTx pks=`%+v` failed", pks)
	}

	// 5. 删除subject blacklist
	err = c.subjectBlackListService.BulkDeleteWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "subjectBlackListService.BulkDeleteWithTx pks=`%+v` failed", pks)
	}

	// 6. 删除rbac策略
	err = c.subjectActionGroupResourceService.BulkDeleteBySubjectPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(
			err,
			"subjectActionGroupResourceService.BulkDeleteBySubjectPKsWithTx subjectPKs=`%+v` failed",
			pks,
		)
	}

	err = c.subjectActionExpressionService.BulkDeleteBySubjectPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(
			err,
			"subjectActionExpressionService.BulkDeleteBySubjectPKsWithTx subjectPKs=`%+v` failed",
			pks,
		)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx commit error")
	}

	// 5. 清除缓存
	// 清除涉及的所有缓存 [subjectGroup / subjectDetails]
	cacheimpls.BatchDeleteSubjectDepartmentCache(pks)
	cacheimpls.BatchDeleteSubjectGroupCache(pks)

	for _, s := range subjects {
		cacheimpls.DeleteSubjectPK(s.Type, s.ID)
		cacheimpls.DeleteLocalSubjectPK(s.Type, s.ID)
	}

	// 清理subject system group缓存
	cacheimpls.BatchDeleteSubjectAllSystemGroupCache(pks)

	// 发送事件
	c.subjectEventProducer.PublishDeleteEvent(svcSubjects)

	return err
}

func convertToServiceSubjects(subjects []Subject) []types.Subject {
	svcSubjects := make([]types.Subject, 0, len(subjects))
	for _, subject := range subjects {
		svcSubjects = append(svcSubjects, types.Subject{
			ID:   subject.ID,
			Type: subject.Type,
			Name: subject.Name,
		})
	}

	return svcSubjects
}

func deleteGroupPKPolicyCache(groupPKs []int64) {
	// 删除group, 此时group下的所有人subjectDetail 还会有对应的group_pk/dept_pk (这块没有清理, 会导致group虽然被删除,看策略还会被命中)
	// 所以此时需要删除 group 的所有policy cache
	// =>  memory: {system}:{actionPK}:{subjectPK} -> [p1, p2, p3]  | => 这个有change list保证时效
	// =>  redis: {system}:{subjectPK} -> [p1, p2, p3]

	// NOTE: 这里只有group需要delete pks => 其他的呢? 不会有问题, 因为subjectPK被清理了
	// 只delete group policy cache :       groups * system数量 * action数量
	// 不调用这个接口, 删除 group下的所有成员/department下的所有成员的 subjectDetail cache?  groups * 成员列表 * system数量
	var allSystems []types.System
	systemSVC := service.NewSystemService()
	allSystems, err := systemSVC.ListAll()
	if err != nil {
		log.WithError(err).Errorf("deleteGroupPKPolicyCache fail groupPKs=`%v`", groupPKs)
	} else {
		systemIDs := make([]string, 0, len(allSystems))
		for _, s := range allSystems {
			systemIDs = append(systemIDs, s.ID)
		}

		err = pl.BatchDeleteSystemSubjectPKsFromCache(systemIDs, groupPKs)
		if err != nil {
			log.Error(err.Error())
		}
	}
}
