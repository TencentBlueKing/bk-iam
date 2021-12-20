# 本地开发环境搭建


## 1. 安装 go1.17 或更高的版本

[Go: Download and install](https://golang.org/doc/install)

```shell
$ go version
go version go1.17.5 darwin/amd64
```

## 2. 初始化表结构

目前所有的db migration以sql文件的形式组织, 放在 `build/support-files/sql` 下, 文件名中带了序号

先按顺序初始化后台数据库表结构(`bkiam`)

```bash
MYSQL="mysql -h127.0.0.1 -P3306 -uroot -p123456 --default_character_set utf8"

files=`ls build/support-files/sql/*.sql`
for f in $files
do
    echo "source the sql file: $f"
    $MYSQL bkiam < $f
done
```

如果本地没有蓝鲸社区版/企业版 PaaS, 需要新建一个`open_paas`库及表结构

```sql
CREATE DATABASE IF NOT EXISTS open_paas DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;

CREATE TABLE `esb_app_account` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `app_code` varchar(30) NOT NULL,
  `app_token` varchar(128) NOT NULL,
  `introduction` longtext NOT NULL,
  `created_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `app_code` (`app_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

此时, 应用`app_code`/`app_secret`通过`open_paas.esb_app_account`控制, 新增应用, 需要手动插入一条记录, 才能在本地开发环境中调用


```sql
insert into esb_app_account(app_code, app_token, introduction, created_time) values('demo', 'c2cfbc91-28a2-420c-b567-cf7dc33cf39f', '', '2019-08-28 16:52:09');
```

## 3. clone代码并修改配置文件

```shell
$ git clone https://github.com/TencentBlueKing/bk-iam.git
$ cd bk-iam

$ cp config.yaml.tpl config.yaml
$ vim config.yaml # 编辑配置文件中db/redis等配置
```

## 4. 编译启动

所有相关命令都放在 [Makefile](./Makefile)中

```shell
$ make dep
$ make serve  # 编译并拉起服务
```

确认服务正常

```shell
$ curl -vv http://127.0.0.1:9000/ping
pong
```

## 5. 其他命令

```shell
$ make init # 初始化本地开发环境
$ make dep  # 执行 go mod tidy && go mod vendor
$ make doc  # 生成swagger文档
$ make mock # 生成mock文件
$ make lint # lint检查
$ make test # 执行单元测试
$ make build # 编译
$ make build-linux # 交叉编译GOOS=linux GOARCH=amd64
$ make serve # 编译并启动
```
