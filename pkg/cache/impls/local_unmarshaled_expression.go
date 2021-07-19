/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package impls

import (
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"

	"iam/pkg/abac/pdp/types"
	"iam/pkg/cache"
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

	expressions := []types.ResourceExpression{}
	err := jsoniter.UnmarshalFromString(k.expression, &expressions)
	// 无效的policy条件表达式, 容错
	if err != nil {
		err = fmt.Errorf("cache UnmarshalExpression unmarshal %s error: %w",
			k.expression, err)
		return nil, err
	}
	return expressions, nil
}

// GetUnmarshalledResourceExpression ...
func GetUnmarshalledResourceExpression(
	expression string,
	signature string,
) (expressions []types.ResourceExpression, err error) {
	key := ResourceExpressionCacheKey{
		expression: expression,
		signature:  signature,
	}

	var value interface{}
	value, err = LocalUnmarshaledExpressionCache.Get(key)
	if err != nil {
		return
	}

	var ok bool
	expressions, ok = value.([]types.ResourceExpression)
	if !ok {
		err = errors.New("not []types.ResourceExpression in cache")
		return
	}

	return
}
