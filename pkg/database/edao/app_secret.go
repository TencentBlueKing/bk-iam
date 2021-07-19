/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package edao

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

// BKPaaSApp ...
type BKPaaSApp struct {
	Code      string `db:"code"`
	AuthToken string `db:"auth_token"`
}

// ESBAppAccount ...
type ESBAppAccount struct {
	AppCode  string `db:"app_code"`
	AppToken string `db:"app_token"`
}

// AppSecretManager ...
type AppSecretManager interface {
	Exists(appCode string, appSecret string) (bool, error)
}

type appSecretManager struct {
	DB *sqlx.DB
}

// NewAppSecretManager ...
func NewAppSecretManager() AppSecretManager {
	return appSecretManager{
		DB: database.GetBKPaaSDBClient().DB,
	}
}

// Exists ...
func (m appSecretManager) Exists(appCode, appSecret string) (bool, error) {
	app := BKPaaSApp{}
	err := m.selectFromBKPaaS(&app, appCode, appSecret)
	if err == nil {
		return true, nil
	}

	esbAccount := ESBAppAccount{}
	err = m.selectFromESBAppAccount(&esbAccount, appCode, appSecret)
	if err == nil {
		return true, nil
	}

	// not exists in both two tables
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return false, err
}

func (m appSecretManager) selectFromBKPaaS(app *BKPaaSApp, appCode, appSecret string) error {
	query := `SELECT
        code,
        auth_token
		FROM paas_app
		WHERE code = ?
		AND auth_token = ?`
	return database.SqlxSensitiveGet(m.DB, app, query, appCode, appSecret)
}

func (m appSecretManager) selectFromESBAppAccount(esbAccount *ESBAppAccount, appCode, appSecret string) error {
	query := `SELECT
        app_code,
        app_token
		FROM esb_app_account
		WHERE app_code = ?
		AND app_token = ?`
	return database.SqlxSensitiveGet(m.DB, esbAccount, query, appCode, appSecret)
}
