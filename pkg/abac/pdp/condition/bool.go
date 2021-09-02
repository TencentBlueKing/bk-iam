package condition

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pdp/types"
)

// BoolCondition bool计算
type BoolCondition struct {
	baseCondition
}

//nolint:unparam
func newBoolCondition(key string, values []interface{}) (Condition, error) {
	return &BoolCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

// GetName 名称
func (c *BoolCondition) GetName() string {
	return "Bool"
}

// Eval 求值
func (c *BoolCondition) Eval(ctx types.AttributeGetter) bool {
	attrValue, err := ctx.GetAttr(c.Key)
	if err != nil {
		log.Debugf("get attr %s from ctx %v error %v", c.Key, ctx, err)
		return false
	}

	exprValues := c.GetValues()

	switch attrValue.(type) {
	case []interface{}:
		// bool 计算不支持多个值
		return false
	default:
		// bool 计算不支持多个值
		if len(exprValues) != 1 {
			return false
		}

		valueBool, ok := attrValue.(bool)
		if !ok {
			return false
		}

		exprBool, ok := exprValues[0].(bool)
		if !ok {
			return false
		}

		return valueBool == exprBool
	}
}

func (c *BoolCondition) Translate() (map[string]interface{}, error) {
	if len(c.Value) != 1 {
		return nil, fmt.Errorf("bool not support multi value %+v", c.Value)
	}

	return map[string]interface{}{
		"op":    "eq",
		"field": c.Key,
		"value": c.Value[0],
	}, nil
}
