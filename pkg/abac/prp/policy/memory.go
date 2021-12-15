/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

import (
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"go.uber.org/multierr"

	"iam/pkg/abac/prp/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/types"
)

const (
	MemoryLayer = "PolicyMemoryLayer"

	changeListTypePolicy = "policy"
	// local cache ttl should equals to changeList get by score (max=now, min=max-localCacheTTL)
	policyLocalCacheTTL = 60

	// fetch the top 1000 in change list, protect the auth/query api performance,
	// if hit 1000, some local-cached(system-action -> subjectPK) updated event will not be notified,
	// will be expired in policyLocalCacheTTL, accepted by now
	maxChangeListCount = 1000
)

var changeList = common.NewChangeList(changeListTypePolicy, policyLocalCacheTTL, maxChangeListCount)

type memoryRetriever struct {
	system              string
	actionPK            int64
	missingRetrieveFunc MissingRetrieveFunc

	changeListKey string
	keyPrefix     string
}

func newMemoryRetriever(system string, actionPK int64, retrieveFunc MissingRetrieveFunc) *memoryRetriever {
	return &memoryRetriever{
		system:              system,
		actionPK:            actionPK,
		missingRetrieveFunc: retrieveFunc,

		changeListKey: fmt.Sprintf("%s:%d", system, actionPK),
		keyPrefix:     fmt.Sprintf("%s:%d:", system, actionPK),
	}
}

type cachedPolicy struct {
	timestamp int64
	policies  []types.AuthPolicy
}

func (r *memoryRetriever) genKey(subjectPKStr string) string {
	return r.keyPrefix + subjectPKStr
}

func (r *memoryRetriever) retrieve(subjectPKs []int64) ([]types.AuthPolicy, []int64, error) {
	nowUnix := time.Now().Unix()

	missSubjectPKs := make([]int64, 0, len(subjectPKs))
	policies := make([]types.AuthPolicy, 0, len(subjectPKs))

	changedTimestamps, err := changeList.FetchList(r.changeListKey)
	if err != nil {
		log.WithError(err).Errorf("[%s] batchFetchSubjectPolicyChangedList fail, will re-fetch all subjectPKs=`%v`",
			MemoryLayer, subjectPKs)
		// 全部重查, 不重查可能有脏数据
		missSubjectPKs = subjectPKs
	} else {
		for _, subjectPK := range subjectPKs {
			subjectPKStr := strconv.FormatInt(subjectPK, 10)

			key := r.genKey(subjectPKStr)
			value, found := cacheimpls.LocalPolicyCache.Get(key)
			if !found {
				missSubjectPKs = append(missSubjectPKs, subjectPK)
				continue
			}

			cached, ok := value.(*cachedPolicy)
			if !ok {
				log.Errorf("[%s] parse cachedPolicy in memory cache fail, will do retrieve!", MemoryLayer)
				missSubjectPKs = append(missSubjectPKs, subjectPK)
				continue
			}

			// 如果 5min内有更新, 那么这个算missing
			if changedTS, ok := changedTimestamps[subjectPKStr]; ok {
				if cached.timestamp < changedTS {
					// not the newest
					// 1. append to missing
					missSubjectPKs = append(missSubjectPKs, subjectPK)
					// 2. delete from local. NOTE: you don't need to do here, setMissing will overwrite the cache key-value
					continue
				}
			}

			// skip empty
			if len(cached.policies) == 0 {
				continue
			}

			// NOTE: check expired first!!!!!!!
			for _, p := range cached.policies {
				if p.ExpiredAt > nowUnix {
					policies = append(policies, p)
				}
			}
		}
	}

	if len(missSubjectPKs) > 0 {
		retrievedPolicies, missingPKs, err := r.missingRetrieveFunc(missSubjectPKs)
		if err != nil {
			return nil, nil, err
		}
		// set missing into cache
		r.setMissing(retrievedPolicies, missingPKs)
		// append the retrieved
		policies = append(policies, retrievedPolicies...)

		return policies, missingPKs, nil
	}

	return policies, nil, nil
}

// nolint:unparam
// setMissing will set the retrieved policies and missingPKs into local cache.
// the missingPKs will cached with empty policy list, make sure will not retrieve again next time
func (r *memoryRetriever) setMissing(policies []types.AuthPolicy, missingSubjectPKs []int64) error {
	nowTimestamp := time.Now().Unix()

	// group policies by subjectPK
	groupedPolicies := map[int64][]types.AuthPolicy{}

	// some subject_pk has no policies, so init all as empty list first
	for _, subjectPK := range missingSubjectPKs {
		groupedPolicies[subjectPK] = []types.AuthPolicy{}
	}

	// append the exists policies into groupedPolicies
	for _, p := range policies {
		groupedPolicies[p.SubjectPK] = append(groupedPolicies[p.SubjectPK], p)
	}

	// set into cache
	for subjectPK, ps := range groupedPolicies {
		subjectPKStr := strconv.FormatInt(subjectPK, 10)
		key := r.genKey(subjectPKStr)

		cacheimpls.LocalPolicyCache.Set(
			key,
			&cachedPolicy{
				timestamp: nowTimestamp,
				policies:  ps,
			},
			policyLocalCacheTTL*time.Second)
	}
	return nil
}

func deleteSystemSubjectPKsFromMemory(system string, subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}

	return batchDeleteSystemSubjectPKsFromMemory([]string{system}, subjectPKs)
}

func batchDeleteSystemSubjectPKsFromMemory(systems []string, subjectPKs []int64) error {
	// NOTE :Can't delete the key from the local cache, while no actionPK
	//       Local cache will be update after next retrieve

	if len(systems) == 0 || len(subjectPKs) == 0 {
		return nil
	}

	// NOTE: 当前通过 扩散写的方式, 将这个 `系统-操作列表-用户` => 更新到所有 actionPK 维度的 changelist => 隔离影响
	//       - 普通逻辑:  单个系统下action数量(max 100) * subjectPKs
	//       - 删除group:  系统数量 * 单个系统下action数量(max 100) * group subjectPKs

	// NOTE: 风险 => 当前pipeline 10000次速度非常快, 但是这个数量不可控的(系统数量增多), 此时数量可能非常大
	//       现在, 实现方式比较粗放, 牺牲了变更的性能, 提升每一次鉴权的性能, 代价是值得的
	//       未来, 这里如果出现瓶颈, 那么需要将上层delete subject 细化, 必须先查出subject-action, 再delete => 精细化

	// the max actions per system is 100, make the average 50
	keyMembers := make(map[string][]string, len(systems)*50)
	changeListKeys := make([]string, 0, len(systems)*50)

	actionSVC := service.NewActionService()
	for _, system := range systems {
		keyPrefix := system + ":"

		actions, err := actionSVC.ListThinActionBySystem(system)
		if err != nil {
			log.WithError(err).Errorf(
				"[%s] list system actions fail system=`%s`, subjectPKs=`%v` the changelist will not add these subjectPKs",
				MemoryLayer, system, subjectPKs,
			)

			continue
		}

		for _, action := range actions {
			members := make([]string, 0, len(subjectPKs))
			for _, subjectPK := range subjectPKs {
				members = append(members, strconv.FormatInt(subjectPK, 10))
			}

			key := keyPrefix + strconv.FormatInt(action.PK, 10)
			keyMembers[key] = members
			changeListKeys = append(changeListKeys, key)
		}
	}

	err := multierr.Combine(
		changeList.AddToChangeList(keyMembers),
		changeList.Truncate(changeListKeys),
	)
	return err
}
