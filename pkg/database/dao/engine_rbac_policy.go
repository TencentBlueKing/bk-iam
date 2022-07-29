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

// NOTE: this is a proxy of table rbac_group_resource_policy
// each record will translate to a single policy, and be synced to engine
// engine search will hit the policy, use the group as recommendation.
// NOTE: important, the group policy in rbac has no `expired_at`

// EngineRbacPolicy ...
type EngineRbacPolicy struct {
	PK int64 `db:"pk"`

	GroupPK    int64  `db:"group_pk"`    // 用户组对应subject的自增列ID
	TemplateID int64  `db:"template_id"` // 模板ID，自定义权限则为0
	SystemID   string `db:"system_id"`

	ActionPKs                   string `db:"action_pks"`                      // json存储了action_pk列表
	ActionRelatedResourceTypePK int64  `db:"action_related_resource_type_pk"` // 操作关联的资源类型

	// 授权的资源实例
	ResourceTypePK int64  `db:"resource_type_pk"`
	ResourceID     string `db:"resource_id"`

	UpdatedAt time.Time `db:"updated_at"`
}

// EngineRbacPolicyManager provide the database query for iam-engine
type EngineRbacPolicyManager interface {
	ListBetweenPK(minPK, maxPK int64) (policies []EngineRbacPolicy, err error)
	ListByPKs(pks []int64) (policies []EngineRbacPolicy, err error)
	GetMaxPKBeforeUpdatedAt(updatedAt int64) (pk int64, err error)
	ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt int64) (pks []int64, err error)
}

type engineRbacPolicyManager struct {
	DB *sqlx.DB
}

// NewAbacEnginePolicyManager create EnginePolicyManager
func NewRbacEnginePolicyManager() EngineRbacPolicyManager {
	return &engineRbacPolicyManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListBetweenPK 查询 range pk 之间的所有策略
func (m *engineRbacPolicyManager) ListBetweenPK(
	minPK,
	maxPK int64,
) (policies []EngineRbacPolicy, err error) {
	query := `SELECT
		pk,
		group_pk,
		system_id,
		template_id,
		action_pks,
		action_related_resource_type_pk,
		resource_type_pk,
		resource_id,
		updated_at
		FROM rbac_group_resource_policy
		WHERE pk BETWEEN ? AND ?`
	err = database.SqlxSelect(m.DB, &policies, query, minPK, maxPK)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// ListByPKs 查询指定pk的策略
func (m *engineRbacPolicyManager) ListByPKs(pks []int64) (policies []EngineRbacPolicy, err error) {
	query := `SELECT
		pk,
		group_pk,
		system_id,
		template_id,
		action_pks,
		action_related_resource_type_pk,
		resource_type_pk,
		resource_id,
		updated_at
		FROM rbac_group_resource_policy
		WHERE pk IN (?)`
	err = database.SqlxSelect(m.DB, &policies, query, pks)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// GetMaxPKBeforeUpdatedAt 查询更新时间之前的最大pk
func (m *engineRbacPolicyManager) GetMaxPKBeforeUpdatedAt(updatedAt int64) (pk int64, err error) {
	var maxPK sql.NullInt64
	query := `SELECT MAX(pk) FROM rbac_group_resource_policy WHERE updated_at <= FROM_UNIXTIME(?)`
	err = database.SqlxGet(m.DB, &maxPK, query, updatedAt)
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
func (m *engineRbacPolicyManager) ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt int64) (pks []int64, err error) {
	query := `SELECT pk FROM rbac_group_resource_policy WHERE updated_at BETWEEN FROM_UNIXTIME(?) AND FROM_UNIXTIME(?)`
	err = database.SqlxSelect(m.DB, &pks, query, beginUpdatedAt, endUpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return pks, nil
	}
	return
}
