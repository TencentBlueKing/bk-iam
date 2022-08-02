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
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"

	"iam/pkg/database/dao"
)

// SubjectBlackListSVC ...
const SubjectBlackListSVC = "SubjectBlackListSVC"

// SubjectBlackListService subject加载器
type SubjectBlackListService interface {
	ListSubjectPK() ([]int64, error)

	BulkCreate(subjectPKs []int64) error
	BulkDelete(subjectPKs []int64) error
	BulkDeleteWithTx(tx *sqlx.Tx, subjectPKs []int64) error
}
type subjectBlackListService struct {
	manager dao.SubjectBlackListManager
}

// NewSubjectBlackListService SubjectBlackListService
func NewSubjectBlackListService() SubjectBlackListService {
	return &subjectBlackListService{
		manager: dao.NewSubjectBlackListManager(),
	}
}

func (l *subjectBlackListService) ListSubjectPK() ([]int64, error) {
	return l.manager.ListSubjectPK()
}

// BulkCreate ...
func (l *subjectBlackListService) BulkCreate(subjectPKs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectBlackListSVC, "BulkCreate")

	existSubjectPKs, err := l.manager.ListSubjectPK()
	if err != nil {
		return errorWrapf(err, "manager.ListSubjectPK fail", "")
	}
	existSubjectPKSet := set.NewInt64SetWithValues(existSubjectPKs)

	subjectBlackList := make([]dao.SubjectBlackList, 0, len(subjectPKs))
	for _, subjectPK := range subjectPKs {
		if existSubjectPKSet.Has(subjectPK) {
			continue
		}
		subjectBlackList = append(subjectBlackList, dao.SubjectBlackList{SubjectPK: subjectPK})
	}

	err = l.manager.BulkCreate(subjectBlackList)
	if err != nil {
		return errorWrapf(err, "manager.BulkCreate subjectPKs=`%+v` fail", subjectPKs)
	}
	return nil
}

// BulkDelete ...
func (l *subjectBlackListService) BulkDelete(subjectPKs []int64) error {
	return l.manager.BulkDelete(subjectPKs)
}

// BulkDeleteWithTx ...
func (l *subjectBlackListService) BulkDeleteWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	return l.manager.BulkDeleteWithTx(tx, subjectPKs)
}
