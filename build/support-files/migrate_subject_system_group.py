#!/usr/bin/python3

import argparse
import json
import logging
import time
from collections import defaultdict, namedtuple
from textwrap import dedent
from typing import Dict, Generator, List, Set, Tuple

import pymysql

logging.basicConfig(level=logging.INFO)


"""
!!!重要: 脚本需要在权限中心后台版本 v1.11.x 升级到 v1.12.x升级 前 执行

本脚本用于权限中心后台版本 v1.11.x 升级到 v1.12.x

升级之前需要执行该脚本用于迁移 subject_system_group 表与 group_system_auth_type 表数据

使用说明:

1. 安装依赖

python3 -m pip install PyMySQL

2. 执行迁移

python3 migrate_subject_system_group.py migrate -H localhost -P 3306 -u root -p '' -D bkiam

3. 数据检查

python3 migrate_subject_system_group.py check -H localhost -P 3306 -u root -p '' -D bkiam

使用帮助:

python3 migrate_subject_system_group.py -h
"""


def db_create_table(cursor: pymysql.cursors.Cursor) -> None:
    logging.info("try to create table: subject_system_group")

    # 创建表SQL语句
    sql = dedent(
        """\
        CREATE TABLE IF NOT EXISTS `subject_system_group` (
          `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
          `system_id` varchar(32) NOT NULL,
          `subject_pk` int(10) unsigned NOT NULL,
          `groups` mediumtext NOT NULL, -- JSON
          `reversion` int(10) unsigned NOT NULL DEFAULT 0,
          `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
          PRIMARY KEY (`pk`),
          UNIQUE KEY `idx_uk_subject_system` (`subject_pk`,`system_id`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8;"""
    )

    # 执行SQL语句
    cursor.execute(sql)

    if not db_check_table_exists(cursor, "subject_system_group"):
        raise Exception("subject_system_group table create fail")

    logging.info("try to create table: subject_system_group success")

    logging.info("create table: group_system_auth_type")

    # 创建表SQL语句
    sql = dedent(
        """\
        CREATE TABLE IF NOT EXISTS `group_system_auth_type` (
          `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
          `system_id` varchar(32) NOT NULL,
          `group_pk` int(10) unsigned NOT NULL,
          `auth_type` tinyint(3) unsigned NOT NULL, -- 0 none 1 abac 2 rbac 3 rabc and abac, 支持位运算
          `reversion` int(10) unsigned NOT NULL DEFAULT 0,
          `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
          PRIMARY KEY (`pk`),
          UNIQUE KEY `idx_uk_group_system` (`group_pk`,`system_id`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8;"""
    )

    # 执行SQL语句
    cursor.execute(sql)

    if not db_check_table_exists(cursor, "group_system_auth_type"):
        raise Exception("group_system_auth_type table create fail")

    logging.info("try to create table: group_system_auth_type success")


def db_list_system_id(cursor: pymysql.cursors.Cursor) -> List[str]:
    # 查询数据库中的所有表名
    sql = "SELECT id FROM system_info;"
    cursor.execute(sql)
    return [row[0] for row in cursor.fetchall()]


def db_list_action_pk_by_system(
    cursor: pymysql.cursors.Cursor, system_id: str
) -> List[int]:
    sql = "SELECT pk FROM action WHERE system_id = %s;"
    cursor.execute(sql, [system_id])
    return [row[0] for row in cursor.fetchall()]


def db_list_subject_pk(
    cursor: pymysql.cursors.Cursor, types: List[str]
) -> Generator[int, None, None]:
    # 查询数据库中的所有表名
    sql = "SELECT MAX(pk) FROM subject WHERE type IN %s;"
    cursor.execute(sql, [tuple(types)])
    max_pk = cursor.fetchone()[0]  # type: ignore
    if not max_pk:
        return

    # 每次查询1000条
    sql = "SELECT pk FROM subject WHERE type IN %s AND pk >= %s AND pk < %s;"
    for i in range(1, max_pk + 1000, 1000):
        cursor.execute(sql, [tuple(types), i, i + 1000])
        for row in cursor.fetchall():
            yield row[0]


def db_get_subject_count(cursor: pymysql.cursors.Cursor, types: List[str]) -> int:
    sql = "SELECT COUNT(*) FROM subject WHERE type IN %s;"
    cursor.execute(sql, [tuple(types)])
    return cursor.fetchone()[0]  # type: ignore


def db_list_subject_authorized_action(
    cursor: pymysql.cursors.Cursor, subject_pk: int
) -> List[int]:
    sql = "SELECT DISTINCT(action_pk) FROM policy WHERE subject_pk=%s;"
    cursor.execute(sql, [subject_pk])
    return [row[0] for row in cursor.fetchall()]


GroupExpiredAt = namedtuple("GroupExpiredAt", ["pk", "expired_at"])


def db_list_subject_authorized_group(
    cursor: pymysql.cursors.Cursor, system_id: str, subject_pk: int
) -> List[GroupExpiredAt]:
    sql = (
        "SELECT groups FROM subject_system_group WHERE system_id=%s AND subject_pk=%s;"
    )
    cursor.execute(sql, [system_id, subject_pk])
    row = cursor.fetchone()
    if not row:
        return []
    data = json.loads(row[0])
    return [GroupExpiredAt(pk=int(k), expired_at=v) for k, v in data.items()]


def db_list_user_group_before_expired_at(
    cursor: pymysql.cursors.Cursor, subject_pk: int, expired_at: int
) -> List[GroupExpiredAt]:
    sql = "SELECT parent_pk, policy_expired_at FROM subject_relation WHERE subject_pk=%s AND policy_expired_at>%s;"
    cursor.execute(sql, [subject_pk, expired_at])
    return [GroupExpiredAt(pk=row[0], expired_at=row[1]) for row in cursor.fetchall()]


def db_create_subject_system_group(
    cursor: pymysql.cursors.Cursor,
    system_id: str,
    subject_pk: int,
    groups: List[GroupExpiredAt],
) -> None:
    sql = dedent(
        """\
        INSERT INTO subject_system_group (
            system_id,
            subject_pk,
            groups
        ) VALUES (
            %s,
            %s,
            %s
        ) ON DUPLICATE KEY UPDATE groups=%s;"""
    )
    groups_str = json.dumps({g.pk: g.expired_at for g in groups}, separators=(",", ":"))
    cursor.execute(sql, [system_id, subject_pk, groups_str, groups_str])


def db_batch_create_group_system_auth_type(
    cursor: pymysql.cursors.Cursor, system_id: str, group_pks: List[int]
) -> None:
    sql = dedent(
        """\
        INSERT IGNORE INTO group_system_auth_type (
            system_id,
            group_pk,
            auth_type
        ) VALUES (
            %s,
            %s,
            1
        )"""
    )
    cursor.executemany(sql, [[system_id, group_pk] for group_pk in group_pks])


def db_check_table_exists(cursor: pymysql.cursors.Cursor, table_name: str) -> bool:
    sql = "SHOW TABLES LIKE %s"
    cursor.execute(sql, [table_name])
    row = cursor.fetchone()
    return bool(row)


def db_check_group_system_auth_type_exists(
    cursor: pymysql.cursors.Cursor, group_pk: int, system_id: str
) -> bool:
    sql = "SELECT 1 FROM group_system_auth_type WHERE group_pk = %s AND system_id = %s;"
    cursor.execute(sql, [group_pk, system_id])
    row = cursor.fetchone()
    return bool(row)


def list_group_authorized_system(
    cursor: pymysql.cursors.Cursor,
) -> Generator[Tuple[int, str], None, None]:
    # 查询所有system id
    system_ids = db_list_system_id(cursor)
    system_action_map = {}

    # 查询所有系统的action pk, 生成map
    for system_id in system_ids:
        system_action_map[system_id] = set(db_list_action_pk_by_system(cursor, system_id))

    count = db_get_subject_count(cursor, ["group"])
    logging.info("all group count: %s", count)

    # 检查group_system_auth_type表数据
    for i, group_pk in enumerate(db_list_subject_pk(cursor, ["group"])):
        if i != 0 and i % 1000 == 0:
            logging.info("process count: %s", i)

        logging.debug("check group %s system auth type", group_pk)
        # 查询group授权的所有 action pk
        action_pks = db_list_subject_authorized_action(cursor, group_pk)
        logging.debug("group %s authorized action %s", group_pk, action_pks)

        # 通过action pk转换成group授权的system id
        group_authorized_system_set = set()
        for action_pk in action_pks:
            for system_id, action_set in system_action_map.items():
                if action_pk in action_set:
                    group_authorized_system_set.add(system_id)
                    break  # 一个action_pk 只可能在一个system中, 跳出system检查循环

        logging.debug(
            "group %s authorized system %s", group_pk, group_authorized_system_set
        )
        # 检查group system group type是否存在
        for system_id in group_authorized_system_set:
            yield (group_pk, system_id)

    logging.info("all group process completed, count: %s", count)


def list_user_system_authorized_groups(
    cursor: pymysql.cursors.Cursor, system_group_map: Dict[str, Set[int]], now: int
) -> Generator[Tuple[int, str, List[GroupExpiredAt]], None, None]:
    # 查询所有system id
    system_ids = db_list_system_id(cursor)

    count = db_get_subject_count(cursor, ["user", "department"])
    logging.info("all subject count: %s", count)

    # 2. 创建subject system group

    # 遍历查询所有的用户
    for i, user_pk in enumerate(db_list_subject_pk(cursor, ["user", "department"])):
        if i != 0 and i % 1000 == 0:
            logging.info("process count: %s", i)

        logging.debug("check subject %s subject system group", user_pk)

        # 查询用户关联的用户组
        groups = db_list_user_group_before_expired_at(cursor, user_pk, now)
        logging.debug("user %s got groups %s", user_pk, groups)

        if len(groups) == 0:
            continue

        # 遍历所有的系统
        for system_id in system_ids:
            group_set = system_group_map[system_id]

            # 从用户加入的所有用户组中, 筛选出有system_id权限的用户组
            authorized_groups = [group for group in groups if group.pk in group_set]

            yield (user_pk, system_id, authorized_groups)  # 同时兼容创建与检查逻辑

    logging.info("all subject process completed, count: %s", count)


def gen_system_group_map(cursor) -> Dict[str, Set[int]]:
    logging.info("start create system group map")
    system_group_map = defaultdict(set)
    for group_pk, system_id in list_group_authorized_system(cursor):
        system_group_map[system_id].add(group_pk)
    logging.info("create system group map completed")
    return system_group_map


def create_all_subject_system_group(
    cursor: pymysql.cursors.Cursor, system_group_map: Dict[str, Set[int]]
):
    logging.info("start create all subject system group")

    for user_pk, system_id, authorized_groups in list_user_system_authorized_groups(
        cursor, system_group_map, int(time.time())
    ):
        if authorized_groups:
            db_create_subject_system_group(cursor, system_id, user_pk, authorized_groups)

    logging.info("create all subject system group completed")


def create_all_group_system_auth_type(
    cursor: pymysql.cursors.Cursor, system_group_map: Dict[str, Set[int]]
):
    logging.info("start create all group system auth type")

    for system_id, group_set in system_group_map.items():
        group_pks = list(group_set)
        group_pks.sort()

        # 分批次创建
        for i in range(0, len(group_pks), 200):
            db_batch_create_group_system_auth_type(
                cursor, system_id, group_pks[i : i + 200]
            )

    logging.info("create all group system auth type completed")


def migrate(args):
    # 建议新版鉴权逻辑的数据
    db = pymysql.connect(
        host=args.host,
        port=int(args.port),
        user=args.user,
        password=args.password,
        database=args.database,
        use_unicode=True,
        charset="utf8",
    )
    cursor = db.cursor()
    cursor.execute('SET NAMES utf8;') 
    cursor.execute('SET CHARACTER SET utf8;')
    cursor.execute('SET character_set_connection=utf8;')

    logging.info("start migrate")

    # 1. 建表
    db_create_table(cursor)

    system_group_map = gen_system_group_map(cursor)

    # 2. 创建所有用户的 subject system group
    create_all_subject_system_group(cursor, system_group_map)

    # 3. 创建 group system auth type
    create_all_group_system_auth_type(cursor, system_group_map)

    db.commit()
    db.close()


def check_group_system_auth_type(
    cursor: pymysql.cursors.Cursor, system_group_map: Dict[str, Set[int]]
):
    logging.info("start check group system auth type")

    for system_id, group_pks in system_group_map.items():
        for group_pk in group_pks:
            if not db_check_group_system_auth_type_exists(cursor, group_pk, system_id):
                raise Exception(
                    f"group {group_pk} system_id {system_id} auth type not exists"
                )

    logging.info("check group system auth type completed")


def check_subject_system_group(
    cursor: pymysql.cursors.Cursor, system_group_map: Dict[str, Set[int]]
):
    logging.info("start check subject system group")

    now = int(time.time())
    for user_pk, system_id, authorized_groups in list_user_system_authorized_groups(
        cursor, system_group_map, now
    ):
        query_groups = db_list_subject_authorized_group(cursor, system_id, user_pk)
        # 都为空
        if len(authorized_groups) == 0 and len(query_groups) == 0:
            continue

        # 校验的group pk一致
        authorized_group_set = {group.pk for group in authorized_groups}
        query_group_set = {
            group.pk for group in query_groups if group.expired_at > now
        }   # NOTE: 从subject system group 中取出来的可能已经过期, 需要筛选掉
        if authorized_group_set == query_group_set:
            continue

        raise Exception(
            f"user {user_pk} system {system_id} group not match, expected: {authorized_groups}, actual: {query_groups}"
        )

    logging.info("check subject system group completed")


def check(args):
    # 检查迁移结果
    db = pymysql.connect(
        host=args.host,
        port=int(args.port),
        user=args.user,
        password=args.password,
        database=args.database,
        use_unicode=True,
        charset="utf8",
    )
    cursor = db.cursor()
    cursor.execute('SET NAMES utf8;') 
    cursor.execute('SET CHARACTER SET utf8;')
    cursor.execute('SET character_set_connection=utf8;')

    logging.info("start check")

    # 1. 检查表是否创建成功
    logging.info("start check table exists")
    for table_name in ("subject_system_group", "group_system_auth_type"):
        if not db_check_table_exists(cursor, table_name):
            raise Exception(f"{table_name} table not exists")
    logging.info("check table exists completed")

    system_group_map = gen_system_group_map(cursor)

    # 2. 检查group auth type是否完整
    check_group_system_auth_type(cursor, system_group_map)

    # 3. 检查subject system group数据
    check_subject_system_group(cursor, system_group_map)

    logging.info("check completed, all data migrate completed")

    db.close()


def read_args():
    parser = argparse.ArgumentParser(description="执行或检查iam backend数据迁移")
    parser.add_argument(
        "action",
        help="执行的操作, 支持的操作有: migrate, check",
    )
    parser.add_argument(
        "-H",
        "--host",
        help="mysql host",
    )
    parser.add_argument(
        "-P",
        "--port",
        default=3306,
        help="mysql port",
    )
    parser.add_argument(
        "-u",
        "--user",
        help="mysql user",
    )
    parser.add_argument(
        "-p",
        "--password",
        help="mysql password",
    )
    parser.add_argument(
        "-D",
        "--database",
        default="bkiam",
        help="mysql database",
    )
    return parser.parse_args()


if __name__ == "__main__":
    args = read_args()
    if args.action == "migrate":
        migrate(args)
    elif args.action == "check":
        check(args)
    else:
        raise Exception("only support migrate/check action")
