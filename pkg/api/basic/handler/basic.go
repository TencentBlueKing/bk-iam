/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"iam/pkg/version"
)

// Ping godoc
// @Summary ping-pong for alive test
// @Description /ping to get response from iam, make sure the server is alive
// @ID ping
// @Tags basic
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Header 200 {string} X-Request-Id "the request id"
// @Router /ping [get]
func Pong(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// Version godoc
// @Summary version for identify
// @Description /version to get the version of iam
// @ID version
// @Tags basic
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Header 200 {string} X-Request-Id "the request id"
// @Router /version [get]
func Version(c *gin.Context) {
	runEnv := os.Getenv("RUN_ENV")
	now := time.Now()
	c.JSON(200, gin.H{
		"version":   version.Version,
		"commit":    version.Commit,
		"buildTime": version.BuildTime,
		"goVersion": version.GoVersion,
		"env":       runEnv,
		// return the date and timestamp
		"timestamp": now.Unix(),
		"date":      now,
	})
}
