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
	"fmt"
	"iam/pkg/cache/impls"
	"iam/pkg/util"

	"github.com/gin-gonic/gin"
)

// CredentialsVerify godoc
// @Summary Credentials Verify/认证信息验证
// @Description check app code secret exists
// @ID api-engine-check-appcode-secret-exists
// @Tags engine
// @Accept json
// @Produce json
// @Param body body credentialsVerifySerializer true "the credential request"
// @Success 200 {object} util.Response{data=credentialsVerifyResponseSerializer}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/engine/credentials/verify [post]
func CredentialsVerify(c *gin.Context) {
	// 当前只支持 app code secret 认证
	var req credentialsVerifySerializer

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if req.Type == "app" {
		valid := impls.VerifyAppCodeAppSecret(req.Data.AppCode, req.Data.AppSecret)
		util.SuccessJSONResponse(c, "ok", credentialsVerifyResponseSerializer{Valid: valid})
		return
	}

	util.BadRequestErrorJSONResponse(c, fmt.Sprintf("Unsupported credential type: %s", req.Type))
}
