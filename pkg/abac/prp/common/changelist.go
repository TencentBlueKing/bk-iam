/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common

import (
	"time"

	rds "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/impls"
	"iam/pkg/cache/redis"
	"iam/pkg/util"
)

const (
	changeListLayer = "ChangeList"
)

// ChangeList keep the recent changed keys with the timestamp in redis.
// use redis sorted-set.  key -> { member1: timestamp1, member2: timestamp2}
type ChangeList struct {
	Type     string
	TTL      int64
	MaxCount int64

	KeyPrefix string
}

// NewChangeList create a change list(redis sorted-set),
// only fetch the members changed in ttl, maximum maxCount to prevent the performance
func NewChangeList(_type string, ttl int64, maxCount int64) *ChangeList {
	return &ChangeList{
		Type:     _type,
		TTL:      ttl,
		MaxCount: maxCount,

		KeyPrefix: _type + ":",
	}
}

// FetchList will fetch the recent changed members, score between [nowTimestamp-TTL, nowTimestamp]
func (r *ChangeList) FetchList(key string) (data map[string]int64, err error) {
	max := time.Now().Unix()
	min := max - r.TTL

	// wrap the key
	changeListKey := r.KeyPrefix + key

	zs, err := impls.ChangeListCache.ZRevRangeByScore(changeListKey, min, max, 0, r.MaxCount)
	if err != nil {
		log.WithError(err).Errorf(
			"[%s:%s] zrange by scores fail changeListKey=`%s`, min=`%d`, max=`%d`, offset=`0`, count=`%d`",
			changeListLayer, r.Type, changeListKey, min, max, r.MaxCount)
		return
	}

	// just log
	if int64(len(zs)) == r.MaxCount {
		log.Errorf("[%s:%s] zrange by scores list almost full changeListKey=`%s`, min=`%d`, max=`%d`, offset=`0`, count=`%d`",
			changeListLayer, r.Type, changeListKey, min, max, r.MaxCount)
	}

	data = make(map[string]int64, len(zs))
	for _, z := range zs {
		data[z.Member.(string)] = int64(z.Score)
	}
	return data, nil
}

// AddToChangeList will add changed members to the change list(sorted-set)
func (r *ChangeList) AddToChangeList(keyMembers map[string][]string) error {
	nowUnix := time.Now().Unix()
	score := float64(nowUnix)

	zDataList := make([]redis.ZData, 0, len(keyMembers))
	for key, members := range keyMembers {
		zs := make([]*rds.Z, 0, len(members))
		for _, member := range members {
			zs = append(zs, &rds.Z{
				Score:  score,
				Member: member,
			})
		}

		zDataList = append(zDataList, redis.ZData{
			// wrap the keys
			Key: r.KeyPrefix + key,
			Zs:  zs,
		})
	}

	err := impls.ChangeListCache.BatchZAdd(zDataList)
	if err != nil {
		log.WithError(err).Errorf("[%s:%s]  add items to change list fail zDataList=`%v`",
			changeListLayer, r.Type, zDataList)

		// report to sentry
		util.ReportToSentry("cache error: add items to change list fail",
			map[string]interface{}{
				"layer": changeListLayer,
				"type":  r.Type,
				"data":  zDataList,
				"error": err.Error(),
			})

		return err
	}

	return nil
}

// Truncate will truncate the change list, remove the items expired, who's score is less than (now - TTL)
func (r *ChangeList) Truncate(keys []string) error {
	nowUnix := time.Now().Unix()

	// wrap the keys
	changeListKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		changeListKeys = append(changeListKeys, r.KeyPrefix+key)
	}

	expiredTimestamp := nowUnix - r.TTL
	err := impls.ChangeListCache.BatchZRemove(changeListKeys, 0, expiredTimestamp)
	if err != nil {
		log.WithError(err).Errorf("[%s:%s]  truncated changelist fail keys=`%v`, expiredTimestamp=`%d`",
			changeListLayer, r.Type, changeListKeys, expiredTimestamp)

		// report to sentry
		util.ReportToSentry("cache error: truncate changelist fail",
			map[string]interface{}{
				"layer": changeListLayer,
				"type":  r.Type,
				"keys":  changeListKeys,
				"min":   0,
				"max":   expiredTimestamp,
				"error": err.Error(),
			})

		return err
	}
	return nil
}
