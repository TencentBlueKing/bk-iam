/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package stats

import (
	"time"

	"iam/pkg/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Stats ...
type Stats struct {
	Label string

	TotalCount          int64
	SuccessCount        int64
	FailCount           int64
	StartTime           time.Time
	LastShowProcessTime time.Time
}

func NewStats(label string) *Stats {
	return &Stats{
		Label:               label,
		StartTime:           time.Now(),
		LastShowProcessTime: time.Now(),
	}
}

func (s *Stats) Log(logger logrus.FieldLogger) {
	if s.TotalCount%1000 == 0 || time.Since(s.LastShowProcessTime) > 30*time.Second {
		s.LastShowProcessTime = time.Now()
		logger.Infof("%s processed total count: %d, success count: %d, fail count: %d, elapsed: %s",
			s.Label, s.TotalCount, s.SuccessCount, s.FailCount, time.Since(s.StartTime))
	}

	// set metrics
	metric.TaskTotalCount.With(prometheus.Labels{
		"process": s.Label,
	}).Set(float64(s.TotalCount))

	metric.TaskSuccessCount.With(prometheus.Labels{
		"process": s.Label,
	}).Set(float64(s.SuccessCount))

	metric.TaskFailCount.With(prometheus.Labels{
		"process": s.Label,
	}).Set(float64(s.FailCount))
}
