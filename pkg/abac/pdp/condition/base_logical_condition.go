package condition

type baseLogicalCondition struct {
	content []Condition
}

// GetKeys 返回嵌套条件中所有包含的属性key
func (c *baseLogicalCondition) GetKeys() []string {
	keys := make([]string, 0, len(c.content))
	for _, condition := range c.content {
		keys = append(keys, condition.GetKeys()...)
	}
	return keys
}

func (c *baseLogicalCondition) HasKey(f keyMatchFunc) bool {
	for _, condition := range c.content {
		if condition.HasKey(f) {
			return true
		}
	}
	return false
}

func (c *baseLogicalCondition) GetFirstMatchKeyValues(f keyMatchFunc) ([]interface{}, bool) {
	for _, condition := range c.content {
		// got the first one
		if values, ok := condition.GetFirstMatchKeyValues(f); ok {
			return values, ok
		}
	}
	return nil, false
}
