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

  - id: "open_paas"
    host: "127.0.0.1"
    port: 3306
    user: "root"
    password: "123456"
    name: "open_paas"

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
