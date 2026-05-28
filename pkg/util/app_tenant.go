package util

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

// 根据当前请求的网关 JWT 信息，推断创建/更新 system 时应使用的 tenant_id
//   - 全租户应用（tenant_mode=global）：返回 ""，该 system 对所有租户可见
//   - 单租户应用（tenant_mode=single）：返回 app.tenant_id
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
