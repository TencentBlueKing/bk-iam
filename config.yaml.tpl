debug: true

server:
  host: 127.0.0.1
  port: 9000

  readTimeout: 60
  writeTimeout: 60
  idleTimeout: 180

sentry:
  enable: false
  dsn: ""

# use comma ”,“ separated when multiple app_code
superAppCode: "bk_iam,bk_iam_app"

databases:
  - id: "iam"
    host: "127.0.0.1"
    port: 3306
    user: "root"
    password: "123456"
    name: "bkiam"
    maxOpenConns: 200
    maxIdleConns: 50
    connMaxLifetimeSecond: 600
    tls:                                    # tls配置
      enabled: true                         # 是否开启tls
      certCaFile: "your CA file"            # ca证书路径
      certCertFile: "your cert file"        # 服务器证书路径(可不填)
      certKeyFile: "your key file"          # 服务器私钥路径(可不填)

  - id: "open_paas"
    host: "127.0.0.1"
    port: 3306
    user: "root"
    password: "123456"
    name: "open_paas"
    tls:                                    # tls配置
      enabled: true                         # 是否开启tls
      certCaFile: "your CA file"            # ca证书路径
      certCertFile: "your cert file"        # 证书路径
      certKeyFile: "your key file"          # 私钥路径

redis:
  - id: "cache"
    type: "standalone"
    addr: "localhost:6379"
    password: ""
    db: 0
    # poolSize: 400
    # minIdleConns: 200
    dialTimeout: 5
    readTimeout: 5
    writeTimeout: 5
    masterName: ""
    tls:                                    # tls配置
      enabled: true                         # 是否开启tls
      certCaFile: "your CA file"            # ca证书路径
      certCertFile: "your cert file"        # 证书路径
      certKeyFile: "your key file"          # 私钥路径
  - id: "mq"
    type: "standalone"
    addr: "localhost:6379"
    password: ""
    db: 0
    # poolSize: 400
    # minIdleConns: 200
    dialTimeout: 5
    readTimeout: 5
    writeTimeout: 5
    masterName: ""
    tls:                                    # tls配置
      enabled: true                         # 是否开启tls
      certCaFile: "your CA file"            # ca证书路径
      certCertFile: "your cert file"        # 证书路径
      certKeyFile: "your key file"          # 私钥路径

logger:
  system:
    level: debug
    writer: os
    settings: {name: stdout}
  api:
    level: info
    writer: file
    settings: {name: iam_api.log, size: 100, backups: 10, age: 7, path: ./}
  sql:
    level: debug
    writer: file
    settings: {name: iam_sql.log, size: 100, backups: 10, age: 7, path: ./}
  audit:
    level: info
    writer: file
    settings: {name: iam_audit.log, size: 500, backups: 20, age: 365, path: ./}
  web:
    level: info
    writer: file
    settings: {name: iam_web.log, size: 100, backups: 10, age: 7, path: ./}
  worker:
    level: info
    writer: file
    settings: {name: iam_worker.log, size: 100, backups: 10, age: 7, path: ./}
  component:
    level: info
    writer: file
    settings: {name: iam_component.log, size: 100, backups: 10, age: 7, path: ./}
