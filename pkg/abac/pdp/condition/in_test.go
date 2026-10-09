/*
 * TencentBlueKing is pleased to support the open source community by making BlueKing-IAM available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 */

package condition

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("IN", func() {
	var c *InCondition

	BeforeEach(func() {
		c = &InCondition{baseCondition{Key: "group_ids", Value: []interface{}{"group1", "group2"}}}
	})

	It("registers and creates the native condition", func() {
		condition, err := newConditionFromInterface(map[string]interface{}{
			"IN": map[string]interface{}{"group_ids": []interface{}{"group1", "group2"}},
		})

		assert.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), "IN", condition.GetName())
	})

	It("evaluates scalar and array attributes", func() {
		assert.True(GinkgoT(), c.Eval(strCtx("group1")))
		assert.True(GinkgoT(), c.Eval(listCtx{"other", "group2"}))
		assert.False(GinkgoT(), c.Eval(listCtx{"other"}))
	})

	It("translates to the distributed expression", func() {
		expression, err := c.Translate(true)

		assert.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), map[string]interface{}{
			"op": "in", "field": "group_ids", "value": []interface{}{"group1", "group2"},
		}, expression)
	})

	It("rejects empty policy values", func() {
		condition, err := newInCondition("group_ids", nil)
		assert.NoError(GinkgoT(), err)

		_, err = condition.Translate(true)
		assert.ErrorIs(GinkgoT(), err, errMustNotEmpty)
	})
})
