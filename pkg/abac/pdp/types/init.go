package types

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

var localTimeEnvsCache *gocache.Cache

func init() {
	localTimeEnvsCache = gocache.New(10*time.Second, 30*time.Second)
}
