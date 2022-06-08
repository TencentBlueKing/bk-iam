package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"iam/pkg/config"
	"iam/pkg/util"
)

const (
	rateLimitKey = "rate_limit"
	// 2000 for per app_code per iam instance
	defaultRateLimitPerSecondPerClient = 2000
)

var rateLimiters = sync.Map{}

// NOTE: this middleware used for api rate limit of calling directly
//       all api will be maintained by APIGateway
//       REMOVE THIS MIDDLEWARE WHEN WE NOT SUPPORT CALLING DIRECTLY

func NewRateLimitMiddleware(c *config.Config) gin.HandlerFunc {
	limitCount, ok := c.Quota.API[rateLimitKey]
	if !ok {
		limitCount = defaultRateLimitPerSecondPerClient
	}

	return func(c *gin.Context) {
		appCode := util.GetClientID(c)

		value, _ := rateLimiters.LoadOrStore(appCode, rate.NewLimiter(rate.Limit(limitCount), 2*limitCount))
		limiter := value.(*rate.Limiter)

		if !limiter.Allow() {
			util.TooManyRequestsJSONResponse(c, "hit the rate limit")
			c.Abort()
			return
		}
		c.Next()
	}
}
