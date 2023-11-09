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

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

// SubjectTemplateGroup  用户/部门-人员模版-用户组关系表
type SubjectTemplateGroup struct {
	PK         int64     `db:"pk"`
	SubjectPK  int64     `db:"subject_pk"`
	TemplateID int64     `db:"template_id"`
	GroupPK    int64     `db:"group_pk"`
	ExpiredAt  int64     `db:"expired_at"`
	CreatedAt  time.Time `db:"created_at"`
}

type SubjectTemplateGroupManager interface {
	GetTemplateGroupMemberCount(groupPK, templateID int64) (int64, error)
	ListPagingTemplateGroupMember(
		groupPK, templateID int64,
		limit, offset int64,
	) (members []SubjectTemplateGroup, err error)
	ListRelationBySubjectPKGroupPKs(subjectPK int64, groupPKs []int64) ([]SubjectTemplateGroup, error)
	ListGroupDistinctSubjectPK(groupPK int64) (subjectPKs []int64, err error)
	ListMaxExpiredAtRelation(groupPK int64) ([]SubjectTemplateGroup, error)

	BulkCreateWithTx(tx *sqlx.Tx, relations []SubjectTemplateGroup) error
	BulkUpdateExpiredAtWithTx(tx *sqlx.Tx, relations []SubjectTemplateGroup) error
	BulkUpdateExpiredAtByRelationWithTx(tx *sqlx.Tx, relations []SubjectRelation) error
	BulkDeleteWithTx(tx *sqlx.Tx, relations []SubjectTemplateGroup) error
	GetMaxExpiredAtBySubjectGroup(subjectPK, groupPK int64, excludeTemplateID int64) (int64, error)
}

type subjectTemplateGroupManager struct {
	DB *sqlx.DB
}

// NewSubjectTemplateGroupManager New SubjectTemplateGroupManager
func NewSubjectTemplateGroupManager() SubjectTemplateGroupManager {
	return &subjectTemplateGroupManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// BulkCreateWithTx ...
func (m *subjectTemplateGroupManager) BulkCreateWithTx(tx *sqlx.Tx, relations []SubjectTemplateGroup) error {
	if len(relations) == 0 {
		return nil
	}

	sql := `INSERT INTO subject_template_group (
		subject_pk,
		template_id,
		group_pk,
		expired_at
	) VALUES (:subject_pk,
		:template_id,
		:group_pk,
		:expired_at)`
	return database.SqlxBulkInsertWithTx(tx, sql, relations)
}

// BulkUpdateExpiredAtWithTx ...
func (m *subjectTemplateGroupManager) BulkUpdateExpiredAtByRelationWithTx(
	tx *sqlx.Tx,
	relations []SubjectRelation,
) error {
	sql := `UPDATE subject_template_group
		 SET expired_at = :policy_expired_at
		 WHERE subject_pk = :subject_pk AND group_pk = :parent_pk`
	return database.SqlxBulkUpdateWithTx(tx, sql, relations)
}

// BulkUpdateExpiredAtWithTx ...
func (m *subjectTemplateGroupManager) BulkUpdateExpiredAtWithTx(
	tx *sqlx.Tx,
	relations []SubjectTemplateGroup,
) error {
	sql := `UPDATE subject_template_group
		 SET expired_at = :expired_at
		 WHERE subject_pk = :subject_pk
		 AND group_pk = :group_pk
		 AND template_id = :template_id`
	return database.SqlxBulkUpdateWithTx(tx, sql, relations)
}

// BulkDeleteWithTx ...
func (m *subjectTemplateGroupManager) BulkDeleteWithTx(tx *sqlx.Tx, relations []SubjectTemplateGroup) error {
	if len(relations) == 0 {
		return nil
	}

	sql := `DELETE FROM subject_template_group
		 WHERE subject_pk = :subject_pk
		 AND group_pk = :group_pk
		 AND template_id = :template_id`
	return database.SqlxBulkUpdateWithTx(tx, sql, relations)
}

func (m *subjectTemplateGroupManager) GetMaxExpiredAtBySubjectGroup(
	subjectPK, groupPK int64,
	excludeTemplateID int64,
) (int64, error) {
	var expiredAt int64
	query := `SELECT
		 COALESCE(MAX(expired_at), 0)
		 FROM subject_template_group
		 WHERE subject_pk = ?
		 AND group_pk = ?
		 AND template_id != ?`
	err := database.SqlxGet(m.DB, &expiredAt, query, subjectPK, groupPK, excludeTemplateID)
	return expiredAt, err
}

func (m *subjectTemplateGroupManager) ListPagingTemplateGroupMember(
	groupPK, templateID int64,
	limit, offset int64,
) (members []SubjectTemplateGroup, err error) {
	query := `SELECT
		 pk,
		 subject_pk,
		 template_id,
		 group_pk,
		 expired_at,
		 created_at
		 FROM subject_template_group
		 WHERE group_pk = ?
		 AND template_id = ?
		 ORDER BY pk DESC
		 LIMIT ? OFFSET ?`
	err = database.SqlxSelect(m.DB, &members, query, groupPK, templateID, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// GetTemplateGroupMemberCount ...
func (m *subjectTemplateGroupManager) GetTemplateGroupMemberCount(groupPK, templateID int64) (int64, error) {
	var count int64
	query := `SELECT
		 COUNT(*)
		 FROM subject_template_group
		 WHERE group_pk = ?
		 AND template_id = ?`
	err := database.SqlxGet(m.DB, &count, query, groupPK, templateID)
	return count, err
}

func (m *subjectTemplateGroupManager) ListRelationBySubjectPKGroupPKs(
	subjectPK int64,
	groupPKs []int64,
) ([]SubjectTemplateGroup, error) {
	relations := []SubjectTemplateGroup{}

	query := `SELECT
		 pk,
		 subject_pk,
		 template_id,
		 group_pk,
		 expired_at,
		 created_at
		 FROM subject_template_group
		 WHERE subject_pk = ?
		 AND group_pk in (?)`

	err := database.SqlxSelect(m.DB, &relations, query, subjectPK, groupPKs)
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}

	return relations, err
}

// ListGroupDistinctSubjectPK ...
func (m *subjectTemplateGroupManager) ListGroupDistinctSubjectPK(groupPK int64) (subjectPKs []int64, err error) {
	query := `SELECT
		 DISTINCT(subject_pk)
		 FROM subject_template_group
		 WHERE group_pk = ?`
	err = database.SqlxSelect(m.DB, &subjectPKs, query, groupPK)
	if errors.Is(err, sql.ErrNoRows) {
		return subjectPKs, nil
	}
	return
}

func (m *subjectTemplateGroupManager) ListMaxExpiredAtRelation(groupPK int64) ([]SubjectTemplateGroup, error) {
	relations := []SubjectTemplateGroup{}
	query := `SELECT
		 pk,
		 subject_pk,
		 template_id,
		 group_pk,
		 MAX(expired_at) AS expired_at,
		 created_at
		 FROM subject_template_group
		 WHERE group_pk = ?
		 GROUP BY subject_pk`

	err := database.SqlxSelect(m.DB, &relations, query, groupPK)
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}

	return relations, err
}
