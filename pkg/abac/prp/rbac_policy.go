/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package prp

import (
	"strconv"
	"sync"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"

	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/task"
	"iam/pkg/task/handler"
	"iam/pkg/task/producer"
)

/*
subject RBAC expression 查询缓存:

1. 批量查询action_pk, subject_pks缓存

2. 如果其中有数据已经过期
	重新计算对应subject_pk, action_pk的表达式，并更新缓存
	发送变更通知
*/

var singletonRbacPolicyRedisRetriever RbacPolicyRetriever

var rbacPolicyRedisRetrieverOnce sync.Once

type rbacPolicyRedisRetriever struct {
	subjectActionExpressionService    service.SubjectActionExpressionService
	subjectActionGroupResourceService service.SubjectActionGroupResourceService
	groupAlterEventService            service.GroupAlterEventService

	alterEventProducer producer.Producer

	G singleflight.Group
}

// NOTE: 为保证singleflight.Group的使用，这里返回单例
func newRbacPolicyRedisRetriever() RbacPolicyRetriever {
	if singletonRbacPolicyRedisRetriever == nil {
		rbacPolicyRedisRetrieverOnce.Do(func() {
			singletonRbacPolicyRedisRetriever = &rbacPolicyRedisRetriever{
				subjectActionExpressionService:    service.NewSubjectActionExpressionService(),
				subjectActionGroupResourceService: service.NewSubjectActionGroupResourceService(),
				groupAlterEventService:            service.NewGroupAlterEventService(),
				alterEventProducer:                producer.NewRedisProducer(task.GetRbacEventQueue()),
			}
		})
	}
	return singletonRbacPolicyRedisRetriever
}

// ListBySubjectAction ...
func (r *rbacPolicyRedisRetriever) ListBySubjectAction(
	subjectPKs []int64,
	actionPK int64,
) ([]types.SubjectActionExpression, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "ListBySubjectAction")
	expressions, err := r.listBySubjectAction(subjectPKs, actionPK)
	if err != nil {
		return nil, err
	}

	nowUnix := time.Now().Unix()
	validExpression := make([]types.SubjectActionExpression, 0, len(expressions))
	for _, expression := range expressions {
		if expression.ExpiredAt < nowUnix {
			// 已过期的数据，从原始数据中刷新缓存，并发送变更事件
			key := strconv.FormatInt(expression.SubjectPK, 10) + ":" + strconv.FormatInt(actionPK, 10)
			expI, err, _ := r.G.Do(key, func() (interface{}, error) {
				return r.refreshSubjectActionExpression(expression.SubjectPK, actionPK)
			})
			if err != nil {
				err = errorWrapf(
					err,
					"refreshSubjectActionExpression fail, subjectPK=`%d`, actionPK=`%d`",
					expression.SubjectPK,
					actionPK,
				)
				return nil, err
			}

			exp := expI.(types.SubjectActionExpression)
			if exp.ExpiredAt != 0 { // NOTE: 如果过期时间为0，说明所有的group都以过期，无效数据
				validExpression = append(validExpression, exp)
			}
		}

		validExpression = append(validExpression, expression)
	}

	return validExpression, nil
}

func (r *rbacPolicyRedisRetriever) listBySubjectAction(
	subjectPKs []int64,
	actionPK int64,
) ([]types.SubjectActionExpression, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "listBySubjectAction")

	// query from cache
	expressions, missSubjectPKs, err := r.batchGet(subjectPKs, actionPK)
	if err != nil {
		err = errorWrapf(
			err,
			"batchGet subject action expression fail, subjectPKs=`%+v`, actionPK=`%d`",
			subjectPKs,
			actionPK,
		)
		return nil, err
	}

	if len(missSubjectPKs) == 0 {
		return expressions, nil
	}

	// query from db
	svcExpressions, err := r.subjectActionExpressionService.ListBySubjectAction(missSubjectPKs, actionPK)
	if err != nil {
		err = errorWrapf(
			err,
			"subjectActionExpressionService.ListBySubjectAction fail, subjectPKs=`%+v`, actionPK=`%d`",
			missSubjectPKs,
			actionPK,
		)
		return nil, err
	}

	// set missing cache
	err = r.setMissing(svcExpressions, missSubjectPKs, actionPK)
	if err != nil {
		err = errorWrapf(
			err,
			"setMissing subject action expression fail, subjectPKs=`%+v`, actionPK=`%d`",
			missSubjectPKs,
			actionPK,
		)
		return nil, err
	}

	expressions = append(expressions, svcExpressions...)

	return expressions, nil
}

func (r *rbacPolicyRedisRetriever) setMissing(
	expressions []types.SubjectActionExpression,
	missingSubjectPKs []int64,
	actionPK int64,
) error {
	hitSubjectPKs := set.NewInt64Set()
	kvs := make([]redis.KV, 0, len(missingSubjectPKs))
	for _, expression := range expressions {
		key := cacheimpls.SubjectActionCacheKey{SubjectPK: expression.SubjectPK, ActionPK: actionPK}

		hitSubjectPKs.Add(expression.SubjectPK)
		value, err := cacheimpls.SubjectActionExpressionCache.Marshal(expression)
		if err != nil {
			log.WithError(err).
				Errorf("message pack marshall subject action expression fail, expression=`%+v`",
					expression)

			continue
		}

		kvs = append(kvs, redis.KV{
			Key:   key.Key(),
			Value: conv.BytesToString(value),
		})
	}

	// 填充缓存空值
	for _, subjectPK := range missingSubjectPKs {
		if !hitSubjectPKs.Has(subjectPK) {
			key := cacheimpls.SubjectActionCacheKey{SubjectPK: subjectPK, ActionPK: actionPK}
			kvs = append(kvs, redis.KV{
				Key:   key.Key(),
				Value: "",
			})
		}
	}

	// set cache
	return cacheimpls.SubjectActionExpressionCache.BatchSetWithTx(kvs, 0)
}

func (r *rbacPolicyRedisRetriever) batchGet(
	subjectPKs []int64,
	actionPK int64,
) ([]types.SubjectActionExpression, []int64, error) {
	keys := make([]cache.Key, 0, len(subjectPKs))
	for _, subjectPK := range subjectPKs {
		keys = append(keys, cacheimpls.SubjectActionCacheKey{SubjectPK: subjectPK, ActionPK: actionPK})
	}

	hitValues, err := cacheimpls.SubjectActionExpressionCache.BatchGet(keys)
	if err != nil {
		return nil, nil, err
	}

	expressions := make([]types.SubjectActionExpression, 0, len(hitValues))
	hitSubjectPKs := set.NewInt64Set()
	for kf, value := range hitValues {
		key := kf.(cacheimpls.SubjectActionCacheKey)

		// NOTE： 缓存为空值，跳过
		if value == "" {
			hitSubjectPKs.Add(key.SubjectPK)
			continue
		}

		var expression types.SubjectActionExpression
		err = cacheimpls.SubjectActionExpressionCache.Unmarshal(conv.StringToBytes(value), &expression)
		if err != nil {
			log.WithError(err).
				Errorf("parse string to subject action expression fail, actionPK=`%d`, subjectPK=`%d`",
					key.ActionPK, key.SubjectPK)

			continue
		}

		expressions = append(expressions, expression)
		hitSubjectPKs.Add(key.SubjectPK)
	}

	missSubjectPKs := make([]int64, 0, len(subjectPKs)-hitSubjectPKs.Size())
	for _, subjectPK := range subjectPKs {
		if !hitSubjectPKs.Has(subjectPK) {
			missSubjectPKs = append(missSubjectPKs, subjectPK)
		}
	}
	return expressions, missSubjectPKs, nil
}

func (r *rbacPolicyRedisRetriever) refreshSubjectActionExpression(
	subjectPK, actionPK int64,
) (expression types.SubjectActionExpression, err error) {
	// query subject action group resource from db
	obj, err := r.subjectActionGroupResourceService.Get(subjectPK, actionPK)
	if err != nil {
		return
	}

	// to subject action expression
	expression, err = handler.ConvertSubjectActionGroupResourceToExpression(obj)
	if err != nil {
		return
	}

	var value interface{} = expression
	// NOTE: if expiredAt == 0, means all group resource is expired
	if expression.ExpiredAt == 0 {
		value = ""
	}

	// set cache
	key := cacheimpls.SubjectActionCacheKey{SubjectPK: subjectPK, ActionPK: actionPK}
	err = cacheimpls.SubjectActionExpressionCache.Set(key, value, 0)
	if err != nil {
		return
	}

	// send message to update subject action expression
	go r.sendSubjectActionRefreshMessage(subjectPK, actionPK)
	return expression, nil
}

func (r *rbacPolicyRedisRetriever) sendSubjectActionRefreshMessage(subjectPK, actionPK int64) {
	pk, err := r.groupAlterEventService.CreateBySubjectActionGroup(subjectPK, actionPK, 0)
	if err != nil {
		log.WithError(err).Errorf("create group alter event fail, subjectPK=`%d`, actionPK=`%d`",
			subjectPK, actionPK)
		return
	}

	// send rmq message
	err = r.alterEventProducer.Publish(strconv.FormatInt(pk, 10))
	if err != nil {
		log.WithError(err).Errorf("publish alter event message fail, pk=`%d`", pk)
	}
}

type rbacPolicyDatabaseRetriever struct {
	subjectActionExpressionService    service.SubjectActionExpressionService
	subjectActionGroupResourceService service.SubjectActionGroupResourceService
}

func newRbacPolicyDatabaseRetriever() RbacPolicyRetriever {
	return &rbacPolicyDatabaseRetriever{
		subjectActionExpressionService:    service.NewSubjectActionExpressionService(),
		subjectActionGroupResourceService: service.NewSubjectActionGroupResourceService(),
	}
}

func (r *rbacPolicyDatabaseRetriever) ListBySubjectAction(
	subjectPKs []int64,
	actionPK int64,
) ([]types.SubjectActionExpression, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "ListBySubjectAction")
	expressions, err := r.subjectActionExpressionService.ListBySubjectAction(subjectPKs, actionPK)
	if err != nil {
		err = errorWrapf(
			err,
			"subjectActionExpressionService.ListBySubjectAction fail, subjectPKs=`%+v`, actionPK=`%d`",
			subjectPKs,
			actionPK,
		)
		return nil, err
	}

	nowUnix := time.Now().Unix()
	validExpression := make([]types.SubjectActionExpression, 0, len(expressions))
	for _, expression := range expressions {
		if expression.ExpiredAt < nowUnix {
			// 已过期的数据，从原始数据中转换获取
			exp, err := r.refreshSubjectActionExpression(expression.SubjectPK, actionPK)
			if err != nil {
				err = errorWrapf(
					err,
					"refreshSubjectActionExpression fail, subjectPK=`%d`, actionPK=`%d`",
					expression.SubjectPK,
					actionPK,
				)
				return nil, err
			}

			if exp.ExpiredAt != 0 { // NOTE: 如果过期时间为0，说明所有的group都以过期，无效数据
				validExpression = append(validExpression, exp)
			}
		}

		validExpression = append(validExpression, expression)
	}

	return validExpression, nil
}

func (r *rbacPolicyDatabaseRetriever) refreshSubjectActionExpression(
	subjectPK, actionPK int64,
) (expression types.SubjectActionExpression, err error) {
	// query subject action group resource from db
	obj, err := r.subjectActionGroupResourceService.Get(subjectPK, actionPK)
	if err != nil {
		return
	}

	// to subject action expression
	return handler.ConvertSubjectActionGroupResourceToExpression(obj)
}
