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

// GetAttribute will get the attribute from object, the key is `type.attributeName`,
// will return nil if 1 object not exists 2 object has no that field
func (s *ObjectSet) GetAttribute(key string) interface{} {
	// objField := strings.Split(key, ".")
	// if len(objField) != 2 {
	// 	return nil
	// }
	// _type := objField[0]

	dotIdx := strings.IndexByte(key, '.')
	if dotIdx == -1 {
		return nil
	}

	_type := key[:dotIdx]

	// if !s.Has(_type) {
	// 	return nil
	// }
	obj, exists := s.Get(_type)
	if !exists {
		return nil
	}

	// attributeName := objField[1]
	attributeName := key[dotIdx+1:]

	value, ok := obj[attributeName]
	if !ok {
		return nil
	}

	return value
}
