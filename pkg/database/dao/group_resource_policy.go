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
	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

type GroupResourcePolicy struct {
	PK int64 `db:"pk"`

	GroupPK    int64  `db:"group_pk"`    // 用户组对应subject的自增列ID
	TemplateID int64  `db:"template_id"` // 模板ID，自定义权限则为0
	SystemID   string `db:"system_id"`

	ActionPKs                   string `db:"action_pks"`                      // json存储了action_pk列表
	ActionRelatedResourceTypePK int64  `db:"action_related_resource_type_pk"` // 操作关联的资源类型

	// 授权的资源实例
	ResourceTypePK int64  `db:"resource_type_pk"`
	ResourceID     string `db:"resource_id"`
}

// GroupResourcePolicyPKActionPKs keep the PrimaryKey and action_pks
type GroupResourcePolicyPKActionPKs struct {
	PK        int64  `db:"pk"`
	ActionPKs string `db:"action_pks"` // json存储了action_pk列表
}

type GroupResourcePolicyManager interface {
	GetPKAndActionPKs(
		groupPK, templateID int64,
		systemID string,
		actionRelatedResourceTypePK, resourceTypePK int64,
		resourceID string,
	) (pkActionPKs GroupResourcePolicyPKActionPKs, err error)
	BulkCreateWithTx(tx *sqlx.Tx, groupResourcePolicies []GroupResourcePolicy) error
	BulkUpdateActionPKsWithTx(tx *sqlx.Tx, pkActionPKss []GroupResourcePolicyPKActionPKs) error
	BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) error
}

type groupResourcePolicyManager struct {
	DB *sqlx.DB
}

func NewGroupResourcePolicyManager() GroupResourcePolicyManager {
	return &groupResourcePolicyManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

func (m *groupResourcePolicyManager) GetPKAndActionPKs(
	groupPK, templateID int64,
	systemID string,
	actionRelatedResourceTypePK, resourceTypePK int64,
	resourceID string,
) (pkActionPKs GroupResourcePolicyPKActionPKs, err error) {
	query := `SELECT 
		pk, 
		action_pks
		FROM group_resource_policy
		WHERE group_pk = ?
		AND template_id = ? 
		AND system_id = ? 
		AND action_related_resource_type_pk = ? 
		AND resource_type_pk = ?
		AND resource_id = ?
		LIMIT 1`
	err = database.SqlxGet(
		m.DB, &pkActionPKs, query,
		groupPK, templateID, systemID, actionRelatedResourceTypePK, resourceTypePK, resourceID,
	)
	return
}

func (m *groupResourcePolicyManager) BulkCreateWithTx(tx *sqlx.Tx, groupResourcePolicies []GroupResourcePolicy) error {
	// 防御，避免空数据时SQL执行报错
	if len(groupResourcePolicies) == 0 {
		return nil
	}

	sql := `INSERT INTO group_resource_policy (
		group_pk,
		template_id,
		system_id,
		action_pks,
		action_related_resource_type_pk,
		resource_type_pk,
		resource_id
	) VALUES (
		:group_pk,
		:template_id,
		:system_id,
		:action_pks,
		:action_related_resource_type_pk,
		:resource_type_pk,
		:resource_id
	)`

	return database.SqlxBulkInsertWithTx(tx, sql, groupResourcePolicies)
}

func (m *groupResourcePolicyManager) BulkUpdateActionPKsWithTx(
	tx *sqlx.Tx,
	pkActionPKss []GroupResourcePolicyPKActionPKs,
) error {
	// 防御，避免空数据时SQL执行报错
	if len(pkActionPKss) == 0 {
		return nil
	}

	sql := `UPDATE group_resource_policy SET action_pks = :action_pks WHERE pk = :pk`
	return database.SqlxBulkUpdateWithTx(tx, sql, pkActionPKss)
}

func (m *groupResourcePolicyManager) BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) error {
	if len(pks) == 0 {
		return nil
	}

	sql := `DELETE FROM group_resource_policy WHERE pk IN (?)`
	return database.SqlxDeleteWithTx(tx, sql, pks)
}
