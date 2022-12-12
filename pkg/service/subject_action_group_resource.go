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
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
	"iam/pkg/util/json"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// SubjectActionGroupResourceSVC ...
const SubjectActionGroupResourceSVC = "SubjectActionGroupResourceSVC"

// SubjectActionGroupResourceService ...
type SubjectActionGroupResourceService interface {
	Get(subjectPK, actionPK int64) (obj types.SubjectActionGroupResource, err error)
	CreateOrUpdateWithTx(tx *sqlx.Tx, obj types.SubjectActionGroupResource) error
	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error

	HasAnyByActionPK(actionPK int64) (bool, error)
	DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK int64) error
	DeleteBySubjectActionWithTx(tx *sqlx.Tx, subjectPK, actionPK int64) error
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
	obj types.SubjectActionGroupResource,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionGroupResourceSVC, "CreateOrUpdate")

	daoObj, err := convertToDaoSubjectActionGroupResource(obj)
	if err != nil {
		err = errorWrapf(err, "convertToDaoSubjectActionGroupResource fail, obj=`%+v`", obj)
		return err
	}

	if obj.PK == 0 {
		// create
		err = s.manager.CreateWithTx(tx, daoObj)
		if err != nil {
			err = errorWrapf(err, "manager.CreateWithTx fail, daoObj=`%+v`", daoObj)
			return err
		}
	} else {
		// update
		err = s.manager.UpdateGroupResourceWithTx(tx, daoObj.PK, daoObj.GroupResource)
		if err != nil {
			err = errorWrapf(err, "manager.UpdateGroupResourceWithTx fail, daoObj=`%+v`", daoObj)
			return err
		}
	}

	return nil
}

// BulkDeleteBySubjectPKsWithTx ...
func (s *subjectActionGroupResourceService) BulkDeleteBySubjectPKsWithTx(
	tx *sqlx.Tx,
	subjectPKs []int64,
) error {
	return s.manager.BulkDeleteBySubjectPKsWithTx(tx, subjectPKs)
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

// DeleteBySubjectActionWithTx ...
func (s *subjectActionGroupResourceService) DeleteBySubjectActionWithTx(tx *sqlx.Tx, subjectPK, actionPK int64) error {
	return s.manager.DeleteBySubjectActionWithTx(tx, subjectPK, actionPK)
}

func convertToSvcSubjectActionGroupResource(
	daoObj dao.SubjectActionGroupResource,
) (obj types.SubjectActionGroupResource, err error) {
	obj = types.SubjectActionGroupResource{
		PK:        daoObj.PK,
		SubjectPK: daoObj.SubjectPK,
		ActionPK:  daoObj.ActionPK,
	}

	err = json.UnmarshalFromString(daoObj.GroupResource, &obj.GroupResource)
	return obj, err
}

func convertToDaoSubjectActionGroupResource(
	obj types.SubjectActionGroupResource,
) (daoObj dao.SubjectActionGroupResource, err error) {
	daoObj = dao.SubjectActionGroupResource{
		PK:        obj.PK,
		SubjectPK: obj.SubjectPK,
		ActionPK:  obj.ActionPK,
	}
	daoObj.GroupResource, err = json.MarshalToString(obj.GroupResource)
	return daoObj, err
}
