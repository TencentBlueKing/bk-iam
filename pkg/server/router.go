/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package server

import (
	"github.com/gin-gonic/gin"

	"iam/pkg/api/basic"
	"iam/pkg/api/debug"
	"iam/pkg/api/engine"
	"iam/pkg/api/model"
	"iam/pkg/api/open"
	"iam/pkg/api/policy"
	"iam/pkg/api/web"
	"iam/pkg/config"
	"iam/pkg/middleware"
)

// NewRouter ...
func NewRouter(cfg *config.Config) *gin.Engine {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	// disable console log color
	gin.DisableConsoleColor()

	// router := gin.Default()
	router := gin.New()
	// MW: gin default logger
	router.Use(gin.Logger())
	// MW: recovery with sentry
	router.Use(middleware.Recovery(cfg.Sentry.Enable))
	// MW: request_id
	router.Use(middleware.RequestID())

	// basic apis
	basic.Register(cfg, router)

	// web apis for SaaS
	webRouter := router.Group("/api/v1/web")
	webRouter.Use(middleware.Metrics())
	webRouter.Use(middleware.WebLogger())
	webRouter.Use(middleware.NewClientAuthMiddleware(cfg))
	webRouter.Use(middleware.SuperClientMiddleware())
	web.Register(webRouter)

	// policy apis for auth/query
	policyRouter := router.Group("/api/v1/policy")
	policyRouter.Use(middleware.Metrics())
	policyRouter.Use(middleware.APILogger())
	policyRouter.Use(middleware.NewClientAuthMiddleware(cfg))
	policyRouter.Use(middleware.NewRateLimitMiddleware(cfg))
	policy.Register(policyRouter)

	// restful apis for open api
	openAPIRouter := router.Group("/api/v1/systems")
	openAPIRouter.Use(middleware.Metrics())
	openAPIRouter.Use(middleware.APILogger())
	openAPIRouter.Use(middleware.NewClientAuthMiddleware(cfg))
	policyRouter.Use(middleware.NewRateLimitMiddleware(cfg))
	open.Register(openAPIRouter)

	// perm-model for register
	permModelRouter := router.Group("/api/v1/model")
	permModelRouter.Use(middleware.Metrics())
	permModelRouter.Use(middleware.Audit())
	permModelRouter.Use(middleware.NewClientAuthMiddleware(cfg))
	policyRouter.Use(middleware.NewRateLimitMiddleware(cfg))
	model.Register(permModelRouter)

	// debug api
	debugRouter := router.Group("/api/v1/debug")
	debugRouter.Use(middleware.NewClientAuthMiddleware(cfg))
	debugRouter.Use(middleware.SuperClientMiddleware())
	debug.Register(debugRouter)

	// apis for iam engine
	engineRouter := router.Group("/api/v1/engine")
	engineRouter.Use(middleware.Metrics())
	// NOTE: disable the log
	//engineRouter.Use(middleware.WebLogger())
	engineRouter.Use(middleware.NewClientAuthMiddleware(cfg))
	engineRouter.Use(middleware.SuperClientMiddleware())
	engine.Register(engineRouter)

	return router
}
