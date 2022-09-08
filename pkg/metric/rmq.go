/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package metric

import (
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/prometheus/client_golang/prometheus"
)

type queueStatsCounters struct {
	readyCount      prometheus.Gauge
	rejectedCount   prometheus.Gauge
	connectionCount prometheus.Gauge
	consumerCount   prometheus.Gauge
	unackedCount    prometheus.Gauge
}

func RecordRmqMetrics(connection rmq.Connection) {
	counters := registerCounters(connection)

	go func() {
		for {
			queues, _ := connection.GetOpenQueues()
			stats, _ := connection.CollectStats(queues)
			for queue, queueStats := range stats.QueueStats {
				if counter, ok := counters[queue]; ok {
					counter.readyCount.Set(float64(queueStats.ReadyCount))
					counter.rejectedCount.Set(float64(queueStats.RejectedCount))
					counter.connectionCount.Set(float64(queueStats.ConnectionCount()))
					counter.consumerCount.Set(float64(queueStats.ConsumerCount()))
					counter.unackedCount.Set(float64(queueStats.UnackedCount()))
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

func registerCounters(connection rmq.Connection) map[string]queueStatsCounters {
	counters := map[string]queueStatsCounters{}

	queues, _ := connection.GetOpenQueues()
	for _, queue := range queues {
		counters[queue] = queueStatsCounters{
			readyCount: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   "rmq",
				Name:        "bkiam_rmq_ready",
				Help:        "Number of ready messages on queue",
				ConstLabels: prometheus.Labels{"service": serviceName, "queue": queue},
			}),
			rejectedCount: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   "rmq",
				Name:        "bkiam_rmq_rejected",
				Help:        "Number of rejected messages on queue",
				ConstLabels: prometheus.Labels{"service": serviceName, "queue": queue},
			}),
			connectionCount: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   "rmq",
				Name:        "bkiam_rmq_connection",
				Help:        "Number of connections consuming a queue",
				ConstLabels: prometheus.Labels{"service": serviceName, "queue": queue},
			}),
			consumerCount: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   "rmq",
				Name:        "bkiam_rmq_consumer",
				Help:        "Number of consumers consuming messages for a queue",
				ConstLabels: prometheus.Labels{"service": serviceName, "queue": queue},
			}),
			unackedCount: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   "rmq",
				Name:        "bkiam_rmq_unacked",
				Help:        "Number of unacked messages on a consumer",
				ConstLabels: prometheus.Labels{"service": serviceName, "queue": queue},
			}),
		}

		prometheus.MustRegister(counters[queue].readyCount)
		prometheus.MustRegister(counters[queue].rejectedCount)
		prometheus.MustRegister(counters[queue].connectionCount)
		prometheus.MustRegister(counters[queue].consumerCount)
		prometheus.MustRegister(counters[queue].unackedCount)
	}

	return counters
}
