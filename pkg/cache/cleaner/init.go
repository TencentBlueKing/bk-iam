/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cleaner

import (
	"context"

	"github.com/TencentBlueKing/gopkg/cache"
	log "github.com/sirupsen/logrus"

	"iam/pkg/util"
)

// it's a goroutine

// go CacheCleaner.Delete(a)
// go CacheCleaner.Delete([a1, a2, a3])

// then => put into channel

// the consumer:
// a => type => will case some other cache delete
// 例如 delete subject => will delete subject-group / subject-department / subjectpk ....

const defaultCacheCleanerBufferSize = 2000

// CacheDeleter ...
type CacheDeleter interface {
	Execute(key cache.Key) error
}

// CacheCleaner ...
type CacheCleaner struct {
	name   string
	ctx    context.Context
	buffer chan cache.Key

	deleter CacheDeleter
}

// NewCacheCleaner ...
func NewCacheCleaner(name string, deleter CacheDeleter) *CacheCleaner {
	ctx := context.Background()
	return &CacheCleaner{
		name:    name,
		ctx:     ctx,
		buffer:  make(chan cache.Key, defaultCacheCleanerBufferSize),
		deleter: deleter,
	}
}

// Run ...
func (r *CacheCleaner) Run() {
	log.Infof("running a cache cleaner: %s", r.name)
	var err error
	for {
		select {
		case <-r.ctx.Done():
			return
		case d := <-r.buffer:
			err = r.deleter.Execute(d)
			if err != nil {
				log.Errorf("delete cache key=%s fail: %s", d.Key(), err)

				// report to sentry
				util.ReportToSentry(
					"cache error: delete key fail",
					map[string]interface{}{
						"key":   d.Key(),
						"error": err.Error(),
					},
				)
			}
		}
	}
}

// Delete ...
func (r *CacheCleaner) Delete(key cache.Key) {
	r.buffer <- key
}

// BatchDelete ...
func (r *CacheCleaner) BatchDelete(keys []cache.Key) {
	// TODO: support batch delete in pipeline or tx?
	for _, key := range keys {
		r.buffer <- key
	}
}
