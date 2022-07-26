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
	DeleteGroupWithTx(
		tx *sqlx.Tx,
		subjectPK, actionPK, groupPK int64,
	) (obj types.SubjectActionGroupResource, err error)
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
		obj, err = s.updateWithTx(tx, daoObj, groupPK, expiredAt, resources)
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

// DeleteGroupWithTx ...
func (s *subjectActionGroupResourceService) DeleteGroupWithTx(
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

	delete(obj.GroupResource, groupPK)

	daoObj, err = convertToDaoSubjectActionGroupResource(daoObj.PK, obj)
	if err != nil {
		err = errorWrapf(err, "convertToDaoSubjectActionGroupResource fail, obj=`%+v`", obj)
		return obj, err
	}

	err = s.manager.UpdateGroupResourceWithTx(tx, daoObj)
	if err != nil {
		err = errorWrapf(err, "manager.UpdateWithTx fail, daoObj=`%+v`", daoObj)
	}

	return obj, err
}

func (s *subjectActionGroupResourceService) updateWithTx(
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

	obj.GroupResource[groupPK] = types.ExpiredAtResource{
		ExpiredAt: expiredAt,
		Resources: resources,
	}

	daoObj, err = convertToDaoSubjectActionGroupResource(daoObj.PK, obj)
	if err != nil {
		return obj, err
	}

	err = s.manager.UpdateGroupResourceWithTx(tx, daoObj)
	return obj, err
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
		GroupResource: map[int64]types.ExpiredAtResource{
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
