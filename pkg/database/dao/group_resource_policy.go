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

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

type GroupResourcePolicy struct {
	PK int64 `db:"pk"`
	// signature字段主要是替换唯一索引
	// resource_id + resource_type_pk + system_id + action_related_resource_type_pk + group_pk + template_id
	//  避免过多字段的索引，导致插入变更性能较差
	//  signature = md5(group_pk:template_id:system_id:action_related_resource_type_pk:resource_type_pk:resource_id)
	Signature string `db:"signature"`

	GroupPK    int64  `db:"group_pk"`    // 用户组对应subject的自增列ID
	TemplateID int64  `db:"template_id"` // 模板ID，自定义权限则为0
	SystemID   string `db:"system_id"`

	ActionPKs                   string `db:"action_pks"`                      // json存储了action_pk列表
	ActionRelatedResourceTypePK int64  `db:"action_related_resource_type_pk"` // 操作关联的资源类型

	// 授权的资源实例
	ResourceTypePK int64  `db:"resource_type_pk"`
	ResourceID     string `db:"resource_id"`
}

type ThinGroupResourcePolicy struct {
	GroupPK   int64  `db:"group_pk"`
	ActionPKs string `db:"action_pks"`
}

type GroupResourcePolicyManager interface {
	ListBySignatures(signatures []string) (policies []GroupResourcePolicy, err error)
	ListByGroupSystemActionRelatedResourceType(
		groupPK int64,
		systemID string,
		actionRelatedResourceTypePK int64,
	) (policies []GroupResourcePolicy, err error)
	ListActionPKsByGroup(groupPK int64) ([]string, error)
	BulkCreateWithTx(tx *sqlx.Tx, policies []GroupResourcePolicy) error
	BulkUpdateActionPKsWithTx(tx *sqlx.Tx, policies []GroupResourcePolicy) error
	BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) error

	// auth
	ListThinByResource(
		systemID string,
		actionResourceTypePK, resourceTypePK int64,
		resourceID string,
	) (policies []ThinGroupResourcePolicy, err error)
}

type groupResourcePolicyManager struct {
	DB *sqlx.DB
}

func NewGroupResourcePolicyManager() GroupResourcePolicyManager {
	return &groupResourcePolicyManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

func (m *groupResourcePolicyManager) ListBySignatures(signatures []string) (policies []GroupResourcePolicy, err error) {
	if len(signatures) == 0 {
		return
	}

	query := `SELECT 
		pk,
		signature,
		group_pk,
		template_id,
		system_id,
		action_pks,
		action_related_resource_type_pk,
		resource_type_pk,
		resource_id
		FROM rbac_group_resource_policy
		WHERE signature IN (?)`
	err = database.SqlxSelect(m.DB, &policies, query, signatures)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

func (m *groupResourcePolicyManager) ListByGroupSystemActionRelatedResourceType(
	groupPK int64,
	systemID string,
	actionRelatedResourceTypePK int64,
) (policies []GroupResourcePolicy, err error) {
	query := `SELECT 
		pk,
		signature,
		group_pk,
		template_id,
		system_id,
		action_pks,
		action_related_resource_type_pk,
		resource_type_pk,
		resource_id
		FROM rbac_group_resource_policy
		WHERE group_pk = ?
		AND action_related_resource_type_pk = ?
		AND system_id = ?`
	err = database.SqlxSelect(m.DB, &policies, query, groupPK, actionRelatedResourceTypePK, systemID)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

func (m *groupResourcePolicyManager) BulkCreateWithTx(tx *sqlx.Tx, policies []GroupResourcePolicy) error {
	// 防御，避免空数据时SQL执行报错
	if len(policies) == 0 {
		return nil
	}

	sql := `INSERT INTO rbac_group_resource_policy (
		signature,
		group_pk,
		template_id,
		system_id,
		action_pks,
		action_related_resource_type_pk,
		resource_type_pk,
		resource_id
	) VALUES (
		:signature,
		:group_pk,
		:template_id,
		:system_id,
		:action_pks,
		:action_related_resource_type_pk,
		:resource_type_pk,
		:resource_id
	)`

	return database.SqlxBulkInsertWithTx(tx, sql, policies)
}

func (m *groupResourcePolicyManager) BulkUpdateActionPKsWithTx(tx *sqlx.Tx, policies []GroupResourcePolicy) error {
	// 防御，避免空数据时SQL执行报错
	if len(policies) == 0 {
		return nil
	}

	sql := `UPDATE rbac_group_resource_policy SET action_pks = :action_pks WHERE pk = :pk`
	return database.SqlxBulkUpdateWithTx(tx, sql, policies)
}

func (m *groupResourcePolicyManager) BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) error {
	if len(pks) == 0 {
		return nil
	}

	sql := `DELETE FROM rbac_group_resource_policy WHERE pk IN (?)`
	return database.SqlxDeleteWithTx(tx, sql, pks)
}

func (m *groupResourcePolicyManager) ListThinByResource(
	systemID string,
	actionResourceTypePK, resourceTypePK int64,
	resourceID string,
) (policies []ThinGroupResourcePolicy, err error) {
	query := `SELECT 
		group_pk,
		action_pks
		FROM rbac_group_resource_policy
		WHERE system_id = ?
		AND action_related_resource_type_pk = ?
		AND resource_type_pk = ?
		AND resource_id = ?`
	err = database.SqlxSelect(m.DB, &policies, query, systemID, actionResourceTypePK, resourceTypePK, resourceID)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}

	return policies, err
}

func (m *groupResourcePolicyManager) ListActionPKsByGroup(groupPK int64) (actionPKsList []string, err error) {
	query := `SELECT 
		action_pks
		FROM rbac_group_resource_policy
		WHERE group_pk = ?`
	err = database.SqlxSelect(m.DB, &actionPKsList, query, groupPK)
	if errors.Is(err, sql.ErrNoRows) {
		return actionPKsList, nil
	}

	return actionPKsList, err
}
