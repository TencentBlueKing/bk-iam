/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package dao

import (
	"database/sql"
	"errors"

	"iam/pkg/database"

	"github.com/jmoiron/sqlx"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// SubjectBlackList 黑名单, 用于权限冻结/解冻
// 注意, 当前只支持全局冻结/解冻, 如果需要细粒度冻结/解冻, 需要加扩展字段支持
type SubjectBlackList struct {
	PK        int64 `db:"pk"`
	SubjectPK int64 `db:"subject_pk"`
}

// SubjectBlackListManager ...
type SubjectBlackListManager interface {
	ListSubjectPK() ([]int64, error)

	BulkCreate(subjectBlackList []SubjectBlackList) error
	BulkDelete(subjectPKs []int64) error
	BulkDeleteWithTx(tx *sqlx.Tx, subjectPKs []int64) error
}

type subjectBlackListManager struct {
	DB *sqlx.DB
}

// NewSubjectBlackListManager ...
func NewSubjectBlackListManager() SubjectBlackListManager {
	return &subjectBlackListManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListSubjectPK get all the subject pks in black list
func (m *subjectBlackListManager) ListSubjectPK() ([]int64, error) {
	subjectPKs := []int64{}
	query := `SELECT
		subject_pk
		FROM subject_black_list`
	err := database.SqlxSelect(m.DB, &subjectPKs, query)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return subjectPKs, nil
	}
	return subjectPKs, err
}

// BulkCreate ...
func (m *subjectBlackListManager) BulkCreate(subjectBlackList []SubjectBlackList) error {
	if len(subjectBlackList) == 0 {
		return nil
	}
	sql := `INSERT INTO subject_black_list (
		subject_pk
	) VALUES (
		:subject_pk)`
	return database.SqlxBulkInsert(m.DB, sql, subjectBlackList)
}

// BulkDelete ...
func (m *subjectBlackListManager) BulkDelete(subjectPKs []int64) error {
	sql := `DELETE FROM subject_black_list WHERE subject_pk in (?)`
	_, err := database.SqlxDelete(m.DB, sql, subjectPKs)
	return err
}

// BulkDeleteWithTx ...
func (m *subjectBlackListManager) BulkDeleteWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}
	sql := `DELETE FROM subject_black_list WHERE subject_pk in (?)`
	err := database.SqlxDeleteWithTx(tx, sql, subjectPKs)
	return err
}
