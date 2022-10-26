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

// EngineAbacPolicy ...
type EngineAbacPolicy struct {
	Policy
	UpdatedAt time.Time `db:"updated_at"`
}

// EngineAbacPolicyManager provide the database query for iam-engine
type EngineAbacPolicyManager interface {
	ListBetweenPK(expiredAt, minPK, maxPK int64) (policies []EngineAbacPolicy, err error)
	ListByPKs(pks []int64) (policies []EngineAbacPolicy, err error)
	GetMaxPKBeforeUpdatedAt(updatedAt int64) (pk int64, err error)
	ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt int64) (pks []int64, err error)
}

type engineAbacPolicyManager struct {
	DB *sqlx.DB
}

// NewAbacEnginePolicyManager create EnginePolicyManager
func NewAbacEnginePolicyManager() EngineAbacPolicyManager {
	return &engineAbacPolicyManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListBetweenPK 查询 range pk 之间的所有策略
func (m *engineAbacPolicyManager) ListBetweenPK(
	expiredAt,
	minPK,
	maxPK int64,
) (policies []EngineAbacPolicy, err error) {
	err = m.selectBetweenPK(&policies, expiredAt, minPK, maxPK)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// ListByPKs 查询指定pk的策略
func (m *engineAbacPolicyManager) ListByPKs(pks []int64) (policies []EngineAbacPolicy, err error) {
	err = m.selectByPKs(&policies, pks)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// GetMaxPKBeforeUpdatedAt 查询更新时间之前的最大pk
func (m *engineAbacPolicyManager) GetMaxPKBeforeUpdatedAt(updatedAt int64) (pk int64, err error) {
	var maxPK sql.NullInt64
	err = m.selectMaxPKBeforeUpdatedAt(&maxPK, updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return
	}

	// not valid
	if !maxPK.Valid {
		return 0, nil
	}
	// valid
	pk = maxPK.Int64
	return pk, nil
}

// ListPKBetweenUpdatedAt 查询更新时间之间的所有pk
func (m *engineAbacPolicyManager) ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt int64) (pks []int64, err error) {
	err = m.selectPKBetweenUpdatedAt(&pks, beginUpdatedAt, endUpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return pks, nil
	}
	return
}

func (m *engineAbacPolicyManager) selectBetweenPK(
	policies *[]EngineAbacPolicy,
	expiredAt int64,
	minPK int64,
	maxPK int64,
) error {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id,
		updated_at
		FROM policy
		WHERE expired_at > ?
		AND pk BETWEEN ? AND ?`
	return database.SqlxSelect(m.DB, policies, query, expiredAt, minPK, maxPK)
}

func (m *engineAbacPolicyManager) selectByPKs(policies *[]EngineAbacPolicy, pks []int64) error {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id,
		updated_at
		FROM policy
		WHERE pk IN (?)`
	return database.SqlxSelect(m.DB, policies, query, pks)
}

func (m *engineAbacPolicyManager) selectPKBetweenUpdatedAt(pks *[]int64, beginUpdatedAt, endUpdatedAt int64) error {
	query := `SELECT pk FROM policy WHERE updated_at BETWEEN FROM_UNIXTIME(?) AND FROM_UNIXTIME(?)`
	return database.SqlxSelect(m.DB, pks, query, beginUpdatedAt, endUpdatedAt)
}

func (m *engineAbacPolicyManager) selectMaxPKBeforeUpdatedAt(maxPK *sql.NullInt64, updatedAt int64) error {
	query := `SELECT MAX(pk) FROM policy WHERE updated_at <= FROM_UNIXTIME(?)`
	return database.SqlxGet(m.DB, maxPK, query, updatedAt)
}
