/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cacheimpls

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/stringx"
	"github.com/sirupsen/logrus"

	"iam/pkg/abac/pdp/condition"
	"iam/pkg/abac/pdp/translate"
	"iam/pkg/abac/types"
)

// ResourceExpressionCacheKey is the key for a policy expression, signature is unique
type ResourceExpressionCacheKey struct {
	expression string
	signature  string
}

// Key return the real key of ResourceExpressionCache
func (k ResourceExpressionCacheKey) Key() string {
	return k.signature
}

// UnmarshalExpression will unmarshal the raw data from expression string
func UnmarshalExpression(key cache.Key) (interface{}, error) {
	k := key.(ResourceExpressionCacheKey)

	return translate.PolicyExpressionToCondition(k.expression)
}

// GetUnmarshalledResourceExpression ...
func GetUnmarshalledResourceExpression(
	expression string,
	signature string,
	timestampNano int64,
) (c condition.Condition, err error) {
	// 预防signature为空导致缓存数据冲突
	if signature == "" {
		signature = stringx.MD5Hash(expression)
	}

	key := ResourceExpressionCacheKey{
		expression: expression,
		signature:  signature,
	}

	var value interface{}
	value, exists := LocalUnmarshaledExpressionCache.GetAfterExpirationAnchor(key.Key(), timestampNano)
	if !exists {
		value, err = UnmarshalExpression(key)
		if err != nil {
			return nil, err
		}

		LocalUnmarshaledExpressionCache.Set(key.Key(), value, 0)
	}

	var ok bool
	c, ok = value.(condition.Condition)
	if !ok {
		err = errors.New("not condition.Condition in cache")
		return
	}

	return c, nil
}

func PoliciesTranslate(policies []types.AuthPolicy) (map[string]interface{}, error) {
	timestampNano := time.Now().UnixNano()

	conditions := make([]condition.Condition, 0, len(policies))
	for _, policy := range policies {
		cond, err := GetUnmarshalledResourceExpression(policy.Expression, policy.ExpressionSignature, timestampNano)
		if err != nil {
			logrus.Debugf("pdp EvalPolicy policy id: %d expression: %s format error: %v",
				policy.ID, policy.Expression, err)
			return nil, err
		}
		conditions = append(conditions, cond)
	}

	return translate.ConditionsTranslate(conditions)
}
