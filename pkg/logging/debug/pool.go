/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package debug

import (
	"sync"
	"time"
)

type entryPool struct {
	pool sync.Pool
}

func newEntryPool() *entryPool {
	return &entryPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Entry{
					// Default is three fields, plus one optional.  Give a little extra room.
					Context:   make(Fields, 6),
					Steps:     make([]Step, 0, 5),
					SubDebugs: make([]*Entry, 0, 5),
					Evals:     make(map[int64]string, 3),
				}
			},
		},
	}
}

// Get ...
func (p *entryPool) Get() *Entry {
	entry := p.pool.Get().(*Entry)
	entry.Time = time.Now()

	return entry
}

// Put ...
func (p *entryPool) Put(e *Entry) {
	if e == nil {
		return
	}

	if len(e.SubDebugs) > 0 {
		for _, se := range e.SubDebugs {
			p.Put(se)
		}
	}

	// reference: https://github.com/sirupsen/logrus/pull/796/files
	e.Context = map[string]interface{}{}
	e.Steps = []Step{}
	e.SubDebugs = []*Entry{}
	e.Evals = map[int64]string{}
	e.Error = ""

	p.pool.Put(e)
}
