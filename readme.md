![](docs/resource/img/bk_iam_zh.png)
---

[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://github.com/TencentBlueKing/bk-iam/blob/master/LICENSE.txt) [![Release Version](https://img.shields.io/badge/release-1.8.1-brightgreen.svg)](https://github.com/TencentBlueKing/bk-iam/releases) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/TencentBlueKing/bk-iam/pulls)

[(English Documents Available)](readme_en.md)

## Overview

蓝鲸权限中心（BK-IAM）是蓝鲸智云提供的集中权限管理服务，支持基于蓝鲸开发框架的SaaS和企业第三方系统的权限控制接入，以及支持细粒度的权限管理。

- [架构设计](./docs/overview/architecture.md)
- [代码目录](./docs/overview/project_codes.md)

## Features

蓝鲸权限中心是基于 ABAC 强大权限模型，结合蓝鲸体系内各种业务场景而研发的通用的权限管控产品，可以满足各种业务场景的权限管控场景。

- 强大的权限模型引擎: 基于强大的 ABAC 权限模型, 能够支持尽可能丰富的业务权限场景。
- 细粒度的权限控制: 支持实例级别的权限控制粒度
- 灵活的权限获取方式: 用户可以通过多种途径获取：自定义申请、申请加入用户组、接入系统侧无权限跳转、管理员授权等
- 权限分级管理: 支持超级管理员、系统管理员、分级管理员三种级别的管理模式。
- 组织架构权限管理: 支持通过组织架构来管理权限，包括个人、组织的权限管理。


## Getting started

- [本地开发环境搭建](./docs/quick_start/develop.md)
- [部署及运维](https://bk.tencent.com/docs/document/6.0/160/8394) / [日志说明](https://bk.tencent.com/docs/document/6.0/160/8398?r=1)
- [官方文档: 接入指引](https://bk.tencent.com/docs/document/6.0/160/8391)

## Roadmap

- [版本日志](release.md)

## SDK

- [TencentBlueKing/iam-python-sdk](https://github.com/TencentBlueKing/iam-python-sdk)
- [TencentBlueKing/iam-go-sdk](https://github.com/TencentBlueKing/iam-go-sdk)

## Support

- [蓝鲸论坛](https://bk.tencent.com/s-mart/community)
- [蓝鲸 DevOps 在线视频教程](https://cloud.tencent.com/developer/edu/major-100008)
- 联系我们，技术交流QQ群：

<img src="https://github.com/Tencent/bk-PaaS/raw/master/docs/resource/img/bk_qq_group.png" width="250" hegiht="250" align=center />


## BlueKing Community

- [BK-CI](https://github.com/Tencent/bk-ci)：蓝鲸持续集成平台是一个开源的持续集成和持续交付系统，可以轻松将你的研发流程呈现到你面前。
- [BK-BCS](https://github.com/Tencent/bk-bcs)：蓝鲸容器管理平台是以容器技术为基础，为微服务业务提供编排管理的基础服务平台。
- [BK-BCS-SaaS](https://github.com/Tencent/bk-bcs-saas)：蓝鲸容器管理平台SaaS基于原生Kubernetes和Mesos自研的两种模式，提供给用户高度可扩展、灵活易用的容器产品服务。
- [BK-PaaS](https://github.com/Tencent/bk-PaaS)：蓝鲸PaaS平台是一个开放式的开发平台，让开发者可以方便快捷地创建、开发、部署和管理SaaS应用。
- [BK-SOPS](https://github.com/Tencent/bk-sops)：标准运维（SOPS）是通过可视化的图形界面进行任务流程编排和执行的系统，是蓝鲸体系中一款轻量级的调度编排类SaaS产品。
- [BK-CMDB](https://github.com/Tencent/bk-cmdb)：蓝鲸配置平台是一个面向资产及应用的企业级配置管理平台。

## Contributing

如果你有好的意见或建议，欢迎给我们提 Issues 或 Pull Requests，为蓝鲸开源社区贡献力量。

## License

基于 MIT 协议， 详细请参考[LICENSE](LICENSE.txt)
