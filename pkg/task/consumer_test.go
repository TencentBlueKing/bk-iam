/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package task

import (
	"github.com/adjust/rmq/v4"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

type h struct{}

func (h *h) Handle(msg GroupAlterMessage) error {
	return nil
}

var _ = Describe("Consumer", func() {
	Describe("Consume", func() {
		var mockHandler GroupAlterMessageHandler
		BeforeEach(func() {
			mockHandler = &h{}
		})

		It("Consume", func() {
			delivery := rmq.NewTestDeliveryString(`{"subject_pk":1,"group_pk":2,"action_pk":3}`)
			consumer := &Consumer{GroupAlterMessageHandler: mockHandler}
			consumer.Consume(delivery)
			assert.Equal(GinkgoT(), rmq.Acked, delivery.State)
		})
	})
})
