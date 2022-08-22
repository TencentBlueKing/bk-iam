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

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

type OpenAbacPolicyManager interface {
	Get(pk int64) (Policy, error)
	GetCountByActionBeforeExpiredAt(actionPK int64, expiredAt int64) (int64, error)
	ListPagingByActionPKBeforeExpiredAt(actionPK int64, expiredAt int64, offset int64, limit int64) ([]Policy, error)
	ListByPKs(pks []int64) ([]Policy, error)
}

type openAbacPolicyManager struct {
	DB *sqlx.DB
}

// NewPolicyManager ...
func NewOpenAbacPolicyManager() OpenAbacPolicyManager {
	return &openAbacPolicyManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

func (m *openAbacPolicyManager) Get(pk int64) (policy Policy, err error) {
	query := `SELECT
			pk,
			subject_pk,
			action_pk,
			expression_pk,
			expired_at,
			template_id
			FROM policy
			WHERE pk = ?
			LIMIT 1`
	err = database.SqlxGet(m.DB, &policy, query, pk)
	return
}

// GetCountByActionBeforeExpiredAt ...
func (m *openAbacPolicyManager) GetCountByActionBeforeExpiredAt(
	actionPK int64,
	expiredAt int64,
) (count int64, err error) {
	query := `SELECT
	count(*)
	FROM policy
	WHERE action_pk = ?
	AND expired_at > ?`
	err = database.SqlxGet(m.DB, count, query, actionPK, expiredAt)
	return
}

// ListPagingByActionPKBeforeExpiredAt ...
func (m *openAbacPolicyManager) ListPagingByActionPKBeforeExpiredAt(
	actionPK int64,
	expiredAt int64,
	offset int64,
	limit int64,
) (policies []Policy, err error) {
	query := `SELECT
	pk,
	subject_pk,
	action_pk,
	expression_pk,
	expired_at,
	template_id
	FROM policy
	WHERE action_pk = ?
	AND expired_at > ?
	ORDER BY pk asc
	LIMIT ?, ?`

	// TODO: check when the performance is affect if the offset is greater than?
	if offset > 10000 {
		query = `SELECT
		t.pk,
		t.subject_pk,
		t.action_pk,
		t.expression_pk,
		t.expired_at,
		t.template_id
		FROM policy t
		INNER JOIN
		(
			SELECT pk
			FROM policy
			WHERE action_pk = ?
			AND expired_at > ?
			ORDER BY pk asc
			LIMIT ?, ?
		) p ON t.pk = p.pk`
	}

	err = database.SqlxSelect(m.DB, &policies, query, actionPK, expiredAt, offset, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return policies, err
}

// ListByPKs ...
func (m *openAbacPolicyManager) ListByPKs(pks []int64) (policies []Policy, err error) {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id
		FROM policy
		WHERE pk in (?)`
	err = database.SqlxSelect(m.DB, &policies, query, pks)

	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}
