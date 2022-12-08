/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package database

import (
	"fmt"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/config"
)

// ! set the default https://making.pusher.com/production-ready-connection-pooling-in-go/
// https://www.alexedwards.net/blog/configuring-sqldb
// SetMaxOpenConns
// SetMaxIdleConns
// SetConnMaxLifetime
const (
	// maxOpenConns >= maxIdleConns

	defaultMaxOpenConns    = 100
	defaultMaxIdleConns    = 25
	defaultConnMaxLifetime = 10 * time.Minute
)

// DBClient MySQL DB Instance
type DBClient struct {
	name string

	DB *sqlx.DB

	dataSource string

	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
}

// TestConnection ...
func (db *DBClient) TestConnection() (err error) {
	conn, err := sqlx.Connect("mysql", db.dataSource)
	if err != nil {
		return
	}

	conn.Close()
	return nil
}

// Connect connect to db, and update some settings
func (db *DBClient) Connect() error {
	var err error
	db.DB, err = sqlx.Connect("mysql", db.dataSource)
	if err != nil {
		return err
	}

	db.DB.SetMaxOpenConns(db.maxOpenConns)
	db.DB.SetMaxIdleConns(db.maxIdleConns)
	db.DB.SetConnMaxLifetime(db.connMaxLifetime)

	log.Infof("connect to database: %s[maxOpenConns=%d, maxIdleConns=%d, connMaxLifetime=%s]",
		db.name, db.maxOpenConns, db.maxIdleConns, db.connMaxLifetime)

	return nil
}

// Close close db connection
func (db *DBClient) Close() {
	if db.DB != nil {
		db.DB.Close()
	}
}

// NewDBClient :
func NewDBClient(cfg *config.Database) *DBClient {
	dataSource := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=%s&parseTime=True&interpolateParams=true&loc=%s&time_zone=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		"utf8",
		"UTC",
		url.QueryEscape("+00:00"),
	)

	maxOpenConns := defaultMaxOpenConns
	if cfg.MaxOpenConns > 0 {
		maxOpenConns = cfg.MaxOpenConns
	}

	maxIdleConns := defaultMaxIdleConns
	if cfg.MaxIdleConns > 0 {
		maxIdleConns = cfg.MaxIdleConns
	}

	if maxOpenConns < maxIdleConns {
		log.Errorf("error config for database %s, maxOpenConns should greater or equals to maxIdleConns, will"+
			"use the default [defaultMaxOpenConns=%d, defaultMaxIdleConns=%d]",
			cfg.Name, defaultMaxOpenConns, defaultMaxIdleConns)
		maxOpenConns = defaultMaxOpenConns
		maxIdleConns = defaultMaxIdleConns
	}

	connMaxLifetime := defaultConnMaxLifetime
	if cfg.ConnMaxLifetimeSecond > 0 {
		if cfg.ConnMaxLifetimeSecond >= 60 {
			connMaxLifetime = time.Duration(cfg.ConnMaxLifetimeSecond) * time.Second
		} else {
			log.Errorf("error config for database %s, connMaxLifetimeSeconds should be greater than 60 seconds"+
				"use the default [defaultConnMaxLifetime=%s]",
				cfg.Name, defaultConnMaxLifetime)
		}
	}

	return &DBClient{
		name:            cfg.Name,
		dataSource:      dataSource,
		maxOpenConns:    maxOpenConns,
		maxIdleConns:    maxIdleConns,
		connMaxLifetime: connMaxLifetime,
	}
}
