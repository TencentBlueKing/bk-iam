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
	"fmt"

	"github.com/TencentBlueKing/gopkg/cache"

	"iam/pkg/service"
	"iam/pkg/service/types"
)

func retrieveSubject(key cache.Key) (interface{}, error) {
	k := key.(SubjectPKCacheKey)
	svc := service.NewSubjectService()
	return svc.Get(k.PK)
}

// GetSubjectByPK ...
func GetSubjectByPK(pk int64) (subject types.Subject, err error) {
	key := SubjectPKCacheKey{
		PK: pk,
	}

	var value interface{}
	value, err = LocalSubjectCache.Get(key)
	if err != nil {
		return
	}

	var ok bool
	subject, ok = value.(types.Subject)
	if !ok {
		err = errors.New("not types.Subject in cache")
		return
	}

	return
}

// BatchGet ...
func BatchGet(pks []int64) (subjects []types.Subject, err error) {
	subjects = make([]types.Subject, 0, len(pks))
	missPKs := make([]int64, 0, len(pks))
	for _, pk := range pks {
		key := SubjectPKCacheKey{
			PK: pk,
		}
		value, ok := LocalSubjectCache.DirectGet(key)
		if !ok {
			missPKs = append(missPKs, pk)
			continue
		}

		subject, ok := value.(types.Subject)
		if !ok {
			err = fmt.Errorf("not types.Subject in cache for pk=%d", pk)
			return
		}
		subjects = append(subjects, subject)
	}

	if len(missPKs) > 0 {
		svc := service.NewSubjectService()
		missSubjects, err := svc.ListByPKs(missPKs)
		if err != nil {
			return nil, err
		}

		for _, subject := range missSubjects {
			key := SubjectPKCacheKey{
				PK: subject.PK,
			}
			LocalSubjectCache.Set(key, subject)
		}

		subjects = append(subjects, missSubjects...)
	}

	return subjects, nil
}
