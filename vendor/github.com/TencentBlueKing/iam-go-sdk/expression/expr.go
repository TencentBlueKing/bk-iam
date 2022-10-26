/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-权限中心Go SDK(iam-go-sdk) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package expression

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/TencentBlueKing/iam-go-sdk/expression/eval"
	"github.com/TencentBlueKing/iam-go-sdk/expression/operator"
)

const (
	// KeywordBKIAMPath the field name of BKIAMPath
	KeywordBKIAMPath            = "_bk_iam_path_"
	KeywordBKIAMPathFieldSuffix = "._bk_iam_path_"
)

// ExprCell is the expression cell
type ExprCell struct {
	OP      operator.OP `json:"op"`
	Content []ExprCell  `json:"content"`
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
}

// Eval will evaluate the expression with ObjectSet, return true or false
func (e *ExprCell) Eval(data ObjectSetInterface) bool {
	switch e.OP {

	case operator.AND:
		for _, c := range e.Content {
			if !c.Eval(data) {
				return false
			}
		}
		return true
	case operator.OR:
		for _, c := range e.Content {
			if c.Eval(data) {
				return true
			}
		}
		return false
	default:
		return evalBinaryOperator(e.OP, e.Field, e.Value, data)
	}
}

// String return the text of expression cell
func (e *ExprCell) String() string {
	switch e.OP {
	case operator.AND, operator.OR:
		separator := fmt.Sprintf(" %s ", e.OP)

		subExprs := make([]string, 0, len(e.Content))
		for _, c := range e.Content {
			subExprs = append(subExprs, c.String())
		}
		return fmt.Sprintf("(%s)", strings.Join(subExprs, separator))
	default:
		return fmt.Sprintf("(%v %s %v)", e.Field, e.OP, e.Value)
	}
}

// Render return the rendered text of expression with ObjectSet
func (e *ExprCell) Render(data ObjectSetInterface) string {
	switch e.OP {
	case operator.AND, operator.OR:
		separator := fmt.Sprintf(" %s ", e.OP)

		subExprs := make([]string, 0, len(e.Content))
		for _, c := range e.Content {
			subExprs = append(subExprs, c.Render(data))
		}
		return fmt.Sprintf("(%s)", strings.Join(subExprs, separator))
	default:
		attrValue := data.GetAttribute(e.Field)
		return fmt.Sprintf("(%v %s %v)", attrValue, e.OP, e.Value)
	}
}

func evalBinaryOperator(op operator.OP, field string, policyValue interface{}, data ObjectSetInterface) bool {
	objectValue := data.GetAttribute(field)

	// support _bk_iam_path_, starts with from `/a,1/b,*/` to `/a,1/b,`
	if op == operator.StartsWith && strings.HasSuffix(field, KeywordBKIAMPathFieldSuffix) {
		v, ok := policyValue.(string)
		if ok {
			if strings.HasSuffix(v, ",*/") {
				policyValue = strings.TrimSuffix(v, "*/")
			}
		}
	}

	// NOTE: if you add new operator, read this first: https://github.com/TencentBlueKing/bk-iam-saas/issues/1293
	switch op {
	case operator.Any:
		return true
	case operator.Eq,
		operator.Lt,
		operator.Lte,
		operator.Gt,
		operator.Gte,
		operator.StartsWith,
		operator.EndsWith,
		operator.StringContains:
		// a starts_with b, a not_starts_with, a ends_with b, a not_ends_with b
		// b should be a single value, while a can be a single value or an array
		if isValueTypeArray(policyValue) {
			return false
		}
		return evalPositive(op, objectValue, policyValue)
	case operator.NotEq, operator.NotStartsWith, operator.NotEndsWith:
		// a not_eq b
		// a can be a single value or an array, be should be a single value
		if isValueTypeArray(policyValue) {
			return false
		}
		return evalNegative(op, objectValue, policyValue)
	case operator.In:
		// a in b, a not_in b
		// b should be an array, while a can be a single or an array
		// so we should make the in expression b always be an array
		if !isValueTypeArray(policyValue) {
			return false
		}
		return evalPositive(op, objectValue, policyValue)
	case operator.NotIn:
		if !isValueTypeArray(policyValue) {
			return false
		}
		return evalNegative(op, objectValue, policyValue)
	case operator.Contains:
		// a contains b,  a not_contains b
		// a should be an array, b should be a single value
		// so, we should make the contains expression b always be a single string, while a can be a single value or an array
		if !isValueTypeArray(objectValue) || isValueTypeArray(policyValue) {
			return false
		}
		// NOTE: objectValue is an array, policyValue is single value
		return eval.Contains(objectValue, policyValue)
	case operator.NotContains:
		if !isValueTypeArray(objectValue) || isValueTypeArray(policyValue) {
			return false
		}
		// NOTE: objectValue is an array, policyValue is single value
		return eval.NotContains(objectValue, policyValue)
	default:
		return false
	}
}

func isValueTypeArray(v interface{}) bool {
	if v == nil {
		return false
	}
	kind := reflect.TypeOf(v).Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

// EvalFunc is the func define of eval
type EvalFunc func(e1, e2 interface{}) bool

// evalPositive
//- 1   hit: return True
//- all miss: return False
func evalPositive(op operator.OP, objectValue, policyValue interface{}) bool {
	var evalFunc EvalFunc

	switch op {
	case operator.Eq:
		evalFunc = eval.Equal
	case operator.Lt:
		evalFunc = eval.Less
	case operator.Lte:
		evalFunc = eval.LessOrEqual
	case operator.Gt:
		evalFunc = eval.Greater
	case operator.Gte:
		evalFunc = eval.GreaterOrEqual
	case operator.StartsWith:
		evalFunc = eval.StartsWith
	case operator.EndsWith:
		evalFunc = eval.EndsWith
	case operator.StringContains:
		evalFunc = eval.StringContains
	case operator.In:
		evalFunc = eval.In
	}

	// NOTE: here, the policyValue should not be array! It's single value (except: the In op policyValue is an array)
	// fmt.Println("objectValue isValueTypeArray", objectValue, isValueTypeArray(objectValue))
	if isValueTypeArray(objectValue) {
		// fmt.Println("objectValue is an array", objectValue)
		listValue := reflect.ValueOf(objectValue)
		for i := 0; i < listValue.Len(); i++ {
			if evalFunc(listValue.Index(i).Interface(), policyValue) {
				return true
			}
		}
		return false
	}

	return evalFunc(objectValue, policyValue)
}

// evalNegative:
// - 1   miss: return False
// - all hit: return True
func evalNegative(op operator.OP, objectValue, policyValue interface{}) bool {
	var evalFunc EvalFunc

	switch op {
	case operator.NotEq:
		evalFunc = eval.NotEqual
	case operator.NotStartsWith:
		evalFunc = eval.NotStartsWith
	case operator.NotEndsWith:
		evalFunc = eval.NotEndsWith
	case operator.NotIn:
		evalFunc = eval.NotIn
	}

	// NOTE: here, the policyValue should not be array! It's single value (except: the NotIn op policyValue is an array)
	if isValueTypeArray(objectValue) {
		listValue := reflect.ValueOf(objectValue)
		for i := 0; i < listValue.Len(); i++ {
			if !evalFunc(listValue.Index(i).Interface(), policyValue) {
				return false
			}
		}
		return true
	}

	return evalFunc(objectValue, policyValue)
}
