package evalctx

import (
	"time"

	gocache "github.com/wklken/go-cache"
)

var localTimeEnvsCache *gocache.Cache

func init() {
	localTimeEnvsCache = gocache.New(10*time.Second, 30*time.Second)
}
