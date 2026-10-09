/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import "strings"

// ObjectSetInterface is the interface for a set of objects
type ObjectSetInterface interface {
	Set(_type string, attributes map[string]interface{})
	Get(_type string) (attrs map[string]interface{}, exists bool)
	Has(_type string) bool
	Del(_type string)
	Size() int
	GetAttribute(key string) interface{}
}

// ObjectAttributeExistenceInterface is an optional extension implemented by
// object sets that can distinguish an absent attribute from a nil value.
// Keeping it separate avoids breaking existing ObjectSetInterface
// implementations.
type ObjectAttributeExistenceInterface interface {
	HasAttribute(key string) bool
}

// ObjectSet is the struct of objects
type ObjectSet struct {
	data map[string]map[string]interface{}
}

// NewObjectSet create an ObjectSet
func NewObjectSet() ObjectSetInterface {
	return &ObjectSet{
		data: make(map[string]map[string]interface{}),
	}
}

// Set will set object, with type and attributes
func (s *ObjectSet) Set(_type string, attributes map[string]interface{}) {
	s.data[_type] = attributes
}

// Get will get attributes of the object by the type
func (s *ObjectSet) Get(_type string) (attrs map[string]interface{}, exists bool) {
	attrs, exists = s.data[_type]
	return
}

// Has will check if the ObjectSet contains the object
func (s *ObjectSet) Has(_type string) bool {
	_, ok := s.data[_type]
	return ok
}

// Del will delete the object from ObjectSet
func (s *ObjectSet) Del(_type string) {
	delete(s.data, _type)
}

// Size will return the size of the set
func (s *ObjectSet) Size() int {
	return len(s.data)
}

type attributeKeyPart struct {
	objectType    string
	attributeName string
}

// splitAttributeKey returns all supported ways to split a condition key.
//
// The legacy format uses the last dot as the separator. Keep it first so
// existing policies retain exactly the same lookup precedence. The canonical
// format is {system}.{resource_type}.{attribute}, where attribute may contain
// dots, so its separator is the second dot.
func splitAttributeKey(key string) []attributeKeyPart {
	parts := make([]attributeKeyPart, 0, 2)

	lastDotIdx := strings.LastIndexByte(key, '.')
	if lastDotIdx <= 0 || lastDotIdx == len(key)-1 {
		return parts
	}
	parts = append(parts, attributeKeyPart{
		objectType:    key[:lastDotIdx],
		attributeName: key[lastDotIdx+1:],
	})

	firstDotIdx := strings.IndexByte(key, '.')
	if firstDotIdx == -1 {
		return parts
	}
	secondDotOffset := strings.IndexByte(key[firstDotIdx+1:], '.')
	if secondDotOffset == -1 {
		return parts
	}
	secondDotIdx := firstDotIdx + 1 + secondDotOffset
	if secondDotIdx == lastDotIdx || secondDotIdx == len(key)-1 {
		return parts
	}

	parts = append(parts, attributeKeyPart{
		objectType:    key[:secondDotIdx],
		attributeName: key[secondDotIdx+1:],
	})
	return parts
}

func (s *ObjectSet) getAttribute(key string) (interface{}, bool) {
	for _, part := range splitAttributeKey(key) {
		obj, exists := s.Get(part.objectType)
		if !exists {
			continue
		}

		value, exists := obj[part.attributeName]
		if exists {
			return value, true
		}
	}
	return nil, false
}

// GetAttribute will get the attribute from object. It supports both the
// legacy `type.attribute` lookup and the canonical
// `{system}.{resource_type}.{attribute}` lookup with dots in attribute IDs.
func (s *ObjectSet) GetAttribute(key string) interface{} {
	value, _ := s.getAttribute(key)
	return value
}

// HasAttribute reports whether the condition key resolves to an existing
// attribute. It deliberately distinguishes an absent attribute from a present
// attribute whose value is nil.
func (s *ObjectSet) HasAttribute(key string) bool {
	_, exists := s.getAttribute(key)
	return exists
}
