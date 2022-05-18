# 1.11.4

- bugfix: temporary policy fill expression signature

# 1.11.3

- add: model change event delete api

# 1.11.2

- add: model share api support

# 1.11.1

- add: temporary policy
- upgrade: go 1.18

# 1.10.4

- bugfix: /api/v1/web/unreferenced-expressions timeout
- upgrade: component log request if latency is greater than 200ms

# 1.10.3

- upgrade: release environment attributes

# 1.10.2

- bugfix: API /api/v1/engine/credentials/verify use auth

# 1.10.1

- add: API /api/v1/web//unreferenced-expressions
- upgrade: go version 1.17

# 1.10.0

- upgrade: permission model supports environment attributes

# 1.9.5

- bugfix: healthz check fail if enable bkauth
- bugfix: ModelChangeEvent when action(no policy related) be deleted


# 1.9.4

- add: bkauth support

# 1.9.3

- update: replace some lib with https://github.com/TencentBlueKing/gopkg
- move: API /api/v1/systems/ to /api/v1/open/systems/
- add: API /api/v1/open/users/:user_id/groups

# 1.9.2

- hotfix: condition StringPrefix eval wrong when key is _bk_iam_path_
- bugfix: remove redundant policy index

# 1.9.1

- bugfix: partialEval out of index while the any condition key is empty

# 1.9.0

- refactor: pdp 模块重构, 完备的表达式, 支持两阶段计算
- add: 支持新老版本表达式unmarshal
- add: 支持gt/gte/lt/lte操作符

# 1.8.11

- upgrade: metrics name add prefix bkiam


# 1.8.10

- upgrade: config template update cmdb resource limit

# 1.8.9

- bugfix: AlterCustomPolicies create batch policies with action without resource return 500

# 1.8.8

- upgrade: change local subject pk get from redis first, instead of db first
- bugfix: modify sqlxBulkInsertReturnIDWithTx return id in batches

# 1.8.7

- bugfix: role group member renew cache clean issue
- bugfix: model api asynchronous delete action empty issue

# 1.8.6

- add: local cache expire with a random duration

# 1.8.5

- upgrade: update the expression table structure, delete useless columns

# 1.8.4

- bugfix: ratelimit middleware use wrong first param Limit, should be float number, not 1 every second

# 1.8.3

- add: zap buffered logger
- add: rate limit for api

# 1.8.2

- bugfix: wrong config reference by web logger

# 1.8.1

- bugfix: policy cache database make slice with wrong cap size

# 1.8.0

- refactor: policy and expression cache layer, refactor to local cache with redis change list; data flow `database->redis->local cache`
- refactor: SubjectDetail use custom msgpack marshal/unmarshal
- refactor: rename subjectRelation to subjectGroup, use `[]ThinSubjectGroup`
- refactor: use zap in api/web logger, for better performance
- add: extra random expired seconds for policy/expression redis cache
- add: unmarshaled expression local cache
- remove: `environment` unused field from all expression struct
- remove: department pks from effective subject pks
- remove: action scope `scope_expression` from all struct
- fix: typo from polices to policies
- bugfix: delete subject cache if update the expiredAt
- add: /version api include the `ts`/`date`
- upgrade: go.mod, the moduel to the newest

# 1.7.7

- add: support asynchronous delete action model and delete action policies

# 1.7.6

- add: policy query auth add expression debug info
- bugfix: engine api sql timestamp between

# 1.7.5

- add: web list instance selections api
- update: engine credentials verify api

# 1.7.4

- add: add api for iam engine

# 1.7.3

- bugfix: s2 compress in go-redis/cache, Fix memcopy writing out-of-bounds.(https://github.com/klauspost/compress/pull/315/commits/587204ab8e90e07ecb90864460f2ecacf5424de2)

# 1.7.2

- bugfix: reset the req.resources in auth_by_resources

# 1.7.1

- update: to go 1.16 and upgrade some dependency


# 1.7.0

- refactor: redis cache, move validClients/subjectRoles/subjectPK from redis cache to local cache
- refactor: policy cache/expression cache
- bugfix: subject groups got expired relations
- bugfix: the permission of deleted group still exists in redis cache

# 1.6.1

- bugfix: msgpack Marshal/Unmarshal error after upgrade go-redis/cache
- add: report system error to sentry

# 1.6.0

- bugfix: component request timeout
- add: get system clients api
- update: go-redis version v8

# 1.5.11

- bugfix: modify action without resource types

# 1.5.10

- add: filter group with expired member api
- add: delete expired expression api
- add: query group expired member api
- update: internal to abac
- add: web handler unittest

# 1.5.9

- bugfix: update judge super system permission

# 1.5.8

- bugfix: judge super system permission not raise error when subject not exists

# 1.5.7

- add: group expired member list api

# 1.5.6

- add: renewal function

# 1.5.5

- add: feature shield rule config
- update: action type support 'use'

# 1.5.4

- add: batch auth api
- update: optimize subject action cache

# 1.5.3

- update: optimize role verification logic

# 1.5.2

- add: default superuser configuration

# 1.5.1

- bugfix: pdp condition get type attribute 
- bugfix: healthz api redis check without pool

# 1.5.0

- add: dynamic selection
- add: grading manager

# 1.4.8

- update: support-files/templates redis config support for render redis mode when deploy 
- add: sentinel redis support multiple sentinel

# 1.4.7

- bugfix: cache for empty subject-groups
- add: debug for /api/policy/query_by_actions

# 1.4.6

- update: query subject groups support return created_at fields

# 1.4.5

- add: /version to get identify info
- add: switch support in config

# 1.4.3

- add: protect action from delete or update related_resource_types if the action has related-policies
- add: unittest via ginkgo
- update: component log support latency and response body if error
- update: remove sensitive in error message of iam/pkg/component remoet_resource
- update: merge OR conditions of the same filed with op=eq/in
- change: mysql expression.expression to type MEDIUMTEXT
- change: truncate the sql log if the args too long
- bugfix: healthz error log when mysql ConnMaxLifetimeSecond=60s

# 1.4.2

- add: list remote resource support local cache for 30 seconds

# 1.4.1

- add: quota for system action/resourceType/instanceSelection
- update: remove sensitive info from component log

# 1.4.0

- add: resource creator action support

# 1.3.9

- update: policy api return clear error message when vaild error

# 1.3.8

- add: action_groups web api

# 1.3.7

- update: query_by_ext_resource ext resources can be empty

# 1.3.6

- add: saas_system_configs support
- add: action_groups support
- add: sentinel password for redis

# 1.3.5

- update: resources in policy query request to omitempty

# 1.3.4

- bugfix: web list policy api filter system

# 1.3.3

- update: change filterFields from struts.Map to json.Marshal
- bugfix: delete expression cache fail keys=`[]`

# 1.3.2

- breaking change: /api/v1/policy/query_by_actions response change from action_id to action.id
- add open api: polic get/list/subjects
- bugfix: errors.Is not working

# 1.3.1

- breaking change: `path` to `_iam_path_` for policy
- add: query policy via ext-resources api
- add: api/model action register support `ignore_iam_path` in instance_selection view

# 1.3.0

- bugfix: set wrong expression pk when alter policies

# 1.2.9

- add: policy/query_by_actions support admin any

# 1.2.8

- bugfix: admin any expression wrong

# 1.2.7

- add: admin got all permissions
- add: uinttest for internal
- add: action type support debug/manage/execute

# 1.2.6

- bugfix: unmarshal fail when expression is empty string

# 1.2.5

- bugfix: return instance_selections missing in saas api

# 1.2.4

- add: instance_selection
- modify: action add/update about resource types with related instance_selections
- remove: environment from expression
- change: action without resource types will not save and query expression

# 1.2.3

- bugfix: policy api invalid resource type 500

# 1.2.2

- remove codes of scope
- fix bugs(component init/prometheus metrics)
- refactor pdp
- disable redis cache guard

# 1.2.1

- add: policy cache support `?debug`
- add: error wrap for policy translate
- update: go-mysql-driver interpolateparams to true
- bugfix: action related resource scope={} should not update into database
- bugfix: cache missing with guard can't be clean

# 1.2.0

- break change: change to expression+policy

# 1.1.9

- bugfix: instance_selection to instance_selections

# 1.1.8

- add: support policy cache

# 1.1.7

- update: delte policy api

# 1.1.6

- bugfix: id length validate
- add: action_resource_type add selection_mode
- add: support batch insert
- refactor: cache module
- add: support subject missing no error


# 1.1.5

- bugfix: subject departments empty query id fail

# 1.1.4

- bugfix: build template iam port

# 1.1.3

- add: api/policy/subjects
- refactor: internal + dao + service
- add: api/model/system support provider_config.healthz
- upgrade: go to 1.14.2
- support: batch delete redis cache

# 1.1.2

- update: all mod to newest

# 1.1.1
- refactor: pkg/internal

# 1.1.0

- add: cache support singleflight
- add: api/model support check valid id
- add: api/web del member return count

# 1.0.9

- bugfix: fix healthz db connection leak

# 1.0.8

- ready to release

# 1.0.0

- first version
