/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

import (
	"database/sql"
	"errors"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// SubjectActionGroupResourceSVC ...
const SubjectActionGroupResourceSVC = "SubjectActionGroupResourceSVC"

// SubjectActionGroupResourceService ...
type SubjectActionGroupResourceService interface {
	Get(subjectPK, actionPK int64) (obj types.SubjectActionGroupResource, err error)
	CreateOrUpdateWithTx(
		tx *sqlx.Tx,
		subjectPK, actionPK, groupPK, expiredAt int64,
		resources map[int64][]string,
	) (obj types.SubjectActionGroupResource, err error)
	DeleteGroupResourceWithTx(
		tx *sqlx.Tx,
		subjectPK, actionPK, groupPK int64,
	) (obj types.SubjectActionGroupResource, err error)
	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error

	HasAnyByActionPK(actionPK int64) (bool, error)
	DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK int64) error
}

type subjectActionGroupResourceService struct {
	manager dao.SubjectActionGroupResourceManager
}

func NewSubjectActionGroupResourceService() SubjectActionGroupResourceService {
	return &subjectActionGroupResourceService{
		manager: dao.NewSubjectActionGroupResourceManager(),
	}
}

// Get ...
func (s *subjectActionGroupResourceService) Get(
	subjectPK, actionPK int64,
) (obj types.SubjectActionGroupResource, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionGroupResourceSVC, "Get")

	daoObj, err := s.manager.GetBySubjectAction(subjectPK, actionPK)
	if err != nil {
		err = errorWrapf(err, "manager.GetBySubjectAction fail, subjectPK=`%d`, actionPK=`%d`", subjectPK, actionPK)
		return
	}

	obj, err = convertToSvcSubjectActionGroupResource(daoObj)
	if err != nil {
		err = errorWrapf(err, "convertToSvcSubjectActionGroupResource fail, daoObj=`%+v`", daoObj)
		return obj, err
	}

	return
}

// CreateOrUpdateWithTx ...
func (s *subjectActionGroupResourceService) CreateOrUpdateWithTx(
	tx *sqlx.Tx,
	subjectPK, actionPK, groupPK, expiredAt int64,
	resources map[int64][]string,
) (obj types.SubjectActionGroupResource, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionGroupResourceSVC, "CreateOrUpdate")

	daoObj, err := s.manager.GetBySubjectAction(subjectPK, actionPK)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		err = errorWrapf(err, "manager.GetBySubjectAction fail, subjectPK=`%d`, actionPK=`%d`", subjectPK, actionPK)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		// create
		obj, err = s.createWithTx(tx, subjectPK, actionPK, groupPK, expiredAt, resources)
		if err != nil {
			err = errorWrapf(err,
				"createWithTx fail, subjectPK=`%d`, actionPK=`%d`, groupPK=`%d`, expiredAt=`%d`, resources=`%+v`",
				subjectPK, actionPK, groupPK, expiredAt, resources,
			)
			return
		}
	} else {
		// update
		obj, err = s.updateGroupResourceWithTx(tx, daoObj, groupPK, expiredAt, resources)
		if err != nil {
			err = errorWrapf(err,
				"updateWithTx fail, daoObj=`%+v`, groupPK=`%d`, expiredAt=`%d`, resources=`%+v`",
				daoObj, groupPK, expiredAt, resources,
			)
			return
		}
	}

	return obj, err
}

// DeleteGroupResourceWithTx ...
func (s *subjectActionGroupResourceService) DeleteGroupResourceWithTx(
	tx *sqlx.Tx,
	subjectPK, actionPK, groupPK int64,
) (obj types.SubjectActionGroupResource, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionGroupResourceSVC, "DeleteGroup")

	daoObj, err := s.manager.GetBySubjectAction(subjectPK, actionPK)
	if err != nil {
		err = errorWrapf(err, "manager.GetBySubjectAction fail, subjectPK=`%d`, actionPK=`%d`", subjectPK, actionPK)
		return
	}

	obj, err = convertToSvcSubjectActionGroupResource(daoObj)
	if err != nil {
		err = errorWrapf(err, "convertToSvcSubjectActionGroupResource fail, daoObj=`%+v`", daoObj)
		return obj, err
	}

	// 删除group resource
	obj.DeleteGroupResource(groupPK)

	daoObj, err = convertToDaoSubjectActionGroupResource(daoObj.PK, obj)
	if err != nil {
		err = errorWrapf(err, "convertToDaoSubjectActionGroupResource fail, obj=`%+v`", obj)
		return obj, err
	}

	err = s.manager.UpdateGroupResourceWithTx(tx, daoObj.PK, daoObj.GroupResource)
	if err != nil {
		err = errorWrapf(err, "manager.UpdateWithTx fail, daoObj=`%+v`", daoObj)
	}

	return obj, err
}

func (s *subjectActionGroupResourceService) updateGroupResourceWithTx(
	tx *sqlx.Tx,
	daoObj dao.SubjectActionGroupResource,
	groupPK int64,
	expiredAt int64,
	resources map[int64][]string,
) (types.SubjectActionGroupResource, error) {
	obj, err := convertToSvcSubjectActionGroupResource(daoObj)
	if err != nil {
		return obj, err
	}

	obj.UpdateGroupResource(groupPK, resources, expiredAt)

	daoObj, err = convertToDaoSubjectActionGroupResource(daoObj.PK, obj)
	if err != nil {
		return obj, err
	}

	err = s.manager.UpdateGroupResourceWithTx(tx, daoObj.PK, daoObj.GroupResource)
	return obj, err
}

// BulkDeleteBySubjectPKsWithTx ...
func (s *subjectActionGroupResourceService) BulkDeleteBySubjectPKsWithTx(
	tx *sqlx.Tx,
	subjectPKs []int64,
) error {
	return s.manager.BulkDeleteBySubjectPKsWithTx(tx, subjectPKs)
}

func (s *subjectActionGroupResourceService) createWithTx(
	tx *sqlx.Tx,
	subjectPK int64,
	actionPK int64,
	groupPK int64,
	expiredAt int64,
	resources map[int64][]string,
) (types.SubjectActionGroupResource, error) {
	obj := types.SubjectActionGroupResource{
		SubjectPK: subjectPK,
		ActionPK:  actionPK,
		GroupResource: map[int64]types.ResourceExpiredAt{
			groupPK: {
				ExpiredAt: expiredAt,
				Resources: resources,
			},
		},
	}

	daoObj, err := convertToDaoSubjectActionGroupResource(0, obj)
	if err != nil {
		return obj, err
	}

	err = s.manager.CreateWithTx(tx, daoObj)
	return obj, err
}

// HasAnyByActionPK ...
func (s *subjectActionGroupResourceService) HasAnyByActionPK(actionPK int64) (bool, error) {
	return s.manager.HasAnyByActionPK(actionPK)
}

// DeleteByActionPK ...
func (s *subjectActionGroupResourceService) DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "DeleteByActionPK")
	// 由于删除时可能数量较大，耗时长，锁行数据较多，影响鉴权，所以需要循环删除，限制每次删除的记录数，以及最多执行删除多少次
	rowLimit := int64(10000)
	maxAttempts := 100 // 相当于最多删除100万数据

	for i := 0; i < maxAttempts; i++ {
		rowsAffected, err := s.manager.DeleteByActionPKWithTx(tx, actionPK, rowLimit)
		if err != nil {
			return errorWrapf(err, "manager.DeleteByActionPKWithTx actionPK=`%d`", actionPK)
		}
		// 如果已经没有需要删除的了，就停止
		if rowsAffected == 0 {
			break
		}
	}

	return nil
}

func convertToSvcSubjectActionGroupResource(
	daoObj dao.SubjectActionGroupResource,
) (obj types.SubjectActionGroupResource, err error) {
	obj = types.SubjectActionGroupResource{
		SubjectPK: daoObj.SubjectPK,
		ActionPK:  daoObj.ActionPK,
	}

	err = jsoniter.UnmarshalFromString(daoObj.GroupResource, &obj.GroupResource)
	return obj, err
}

func convertToDaoSubjectActionGroupResource(
	pk int64,
	obj types.SubjectActionGroupResource,
) (daoObj dao.SubjectActionGroupResource, err error) {
	daoObj = dao.SubjectActionGroupResource{
		PK:        pk,
		SubjectPK: obj.SubjectPK,
		ActionPK:  obj.ActionPK,
	}
	daoObj.GroupResource, err = jsoniter.MarshalToString(obj.GroupResource)
	return daoObj, err
}
