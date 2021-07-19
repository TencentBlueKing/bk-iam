# 蓝鲸权限中心代码结构

bk-iam 后台源码目录:

```
bk-iam
├── build            社区版/企业版打包依赖文件
├── cmd              主程序入口
├── pkg              主体代码
│   ├── abac         abac权限模型
│   ├── api          api view层函数
│   │    ├── common
│   │    ├── basic       /ping, /metrics等基础api
│   │    ├── debug       /api/v1/debug 提供给debug-cli的 API
│   │    ├── engine      /api/v1/engine 提供给iam-engine的 API
│   │    ├── model       /api/v1/model 模型注册及变更
│   │    ├── open        /api/v1/open  开放 API
│   │    ├── policy      /api/v1/policy 策略查询及鉴权
│   │    └── web         /api/v1/web 提供给SaaS的 API
│   ├── cache        缓存层
│   ├── service      service层
│   ├── database     数据库
│   ├── component    组件
│   ├── config       配置
│   ├── errorx       错误
│   ├── logging      日志
│   ├── metric       监控
│   ├── middleware   中间件
│   ├── server       server端启动
│   ├── util         工具
│   └── version      版本
├── docs             文档
├── config.yaml.tpl  配置文件模板
├── main.go
├── Makefile
├── LICENSE.txt
├── release.md       版本日志
├── VERSION          版本号
├── go.mod           go相关依赖
├── go.sum
└── vendor
```

代码分层:

```
api -> [abac] -> cache|service -> database/*dao -> database
```

