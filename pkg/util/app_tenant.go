package util

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

// APIGateway JWT 中 app.tenant_mode 的取值
const (
	AppTenantModeGlobal = "global" // 全租户应用
	AppTenantModeSingle = "single" // 单租户应用
)

// InferSystemTenantID 根据当前请求的网关 JWT 信息，推断创建/更新 system 时应使用的 tenant_id
//
//   - 全租户应用（tenant_mode=global）：返回 ""，该 system 对所有租户可见
//   - 单租户应用（tenant_mode=single）：返回 app.tenant_id
//   - 其他情况：返回 error，调用方应拒绝本次请求
//
// 注意：本函数依赖 ClientAuthMiddleware 已将 JWT 解析结果写入 Context
func InferSystemTenantID(c *gin.Context) (string, error) {
	mode := GetAppTenantMode(c)
	appTenantID := GetAppTenantID(c)

	switch mode {
	case AppTenantModeGlobal:
		return "", nil
	case AppTenantModeSingle:
		if appTenantID == "" {
			return "", errors.New("single tenant app must have tenant_id in jwt")
		}
		return appTenantID, nil
	default:
		return "", fmt.Errorf(
			"unknown or missing tenant_mode: %q, request must come from apigateway with a valid jwt",
			mode,
		)
	}
}

// CanAccessSystem 判断当前请求方（根据网关 JWT 中的 app 信息）是否有权访问指定 system
//
// 访问规则：
//   - 全租户应用（tenant_mode=global）：可以访问所有 system（含全租户系统与任意单租户系统）
//   - 单租户应用（tenant_mode=single）：
//   - 可以访问全租户系统（systemTenantID == ""）
//   - 只能访问属于自己租户的单租户系统（systemTenantID == app.tenant_id）
//   - 其它情况（JWT 缺失或非法）：一律拒绝
//
// 注意：本函数依赖 ClientAuthMiddleware 已将 JWT 解析结果写入 Context
func CanAccessSystem(c *gin.Context, systemTenantID string) bool {
	mode := GetAppTenantMode(c)
	appTenantID := GetAppTenantID(c)

	switch mode {
	case AppTenantModeGlobal:
		// 全租户应用：放行全部
		return true
	case AppTenantModeSingle:
		// 单租户应用：必须有 tenant_id
		if appTenantID == "" {
			return false
		}
		// 全租户系统所有单租户应用均可访问；单租户系统仅限同租户
		return systemTenantID == "" || systemTenantID == appTenantID
	default:
		return false
	}
}
