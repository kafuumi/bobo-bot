# bobo-bot

啵啵点赞bot，点赞b站指定评论区下的评论。

## 实现功能

- [x] 点赞评论
- [x] 保存评论
- [x] 统计评论数据
- [x] 监控粉丝数变化

## 配置

文件名为`setting.json`，需要和可执行文件在同一目录中

```json
{
  "botAccount": {
    "uid": 1086284157,
    "uidMd5": "3a09e9axxxxxxxx cookie中的DedeUserID__ckMd5",
    "sessData": "8fec4f85%2C167xxxxxxxxx cookie中的SESSDATA",
    "csrf": "6bf63e0ee5fb2d4f62xxxxxxxx cookie中的bili_jct",
    "sid": "7hdxxxx cookie中的sid"
  },
  "account": {
    "uid": 33605910,
    "alias": "三三"
  },
  "board": {
    "name": "啵版",
    "oid": 662016827293958168
  },
  "config": {
    "fresh": 2,
    "like": 1,
    "isLike": true,
    "isPost": true,
    "isFans": true,
    "hour": 7,
    "minute": 33,
    "dbname": "database.db"
  },
  "logger": {
    "level": "Info",
    "appender": "file"
  },
  "push": {
    "webhook": "钉钉机器人webhook",
    "secret": ""
  }
}
```

### 字段解释

#### `botAccount`

bot所使用的b站账号，通过cookie方式登录

`uid`：账号的`uid`，也就是cookie中的`DedeUserID`

`uidMd5`：cookie中的`DedeUserID__ckMd5`

`sessData`：cookie中的`SESSDATA`

`csrf`：cookie中的`bili_jct`

`sid`：cookie中的`sid`

#### `account`

对应评论区所属的账号，主要用来统计该账号的粉丝数变化。

`uid`：该账号的`uid`

`alias`：别名

#### `board`

评论区相关信息，只支持动态评论区，视频的评论区不支持。

`name`：别名

`oid`：动态对应的`oid`，例如动态链接为：`https://t.bilibili.com/662016827293958168`，其中的`662016827293958168`则是其对应的`oid`

#### `config`

一些配置参数

`fresh`：刷新时间，单位：秒，每隔`fresh`秒获取一次评论。

`like`：两次点赞间隔时间，可以是小数，单位：秒。

`isLike`：布尔值，代表是否开启评论点赞。

`isPost`：布尔值，代表是否发布数据总结动态。

`isFans`：布尔值，代表是否监控粉丝数。

`hour`，`minute`：生成数据汇总的时间，如果`hour`为`-1`，则是每小时生成一次。

例如：`hour=7,minute=33`，则是在每天的7点33分生成。

`dbname`：sqlite3数据库文件名，用于保存获取到的评论。

#### `logger`

日志配置

`level`：日志级别，可选：`Debug`,`Info`,`Warn`,`Error`，需要注意大小写。

`appender`：日志保存方式。可选：`file`：保存在文件中，会自动按文件大小滚动。`console`：不保存，直接输出到标准输出流中。

#### `push`

信息推送配置，将部分错误信息推送至钉钉机器人。如果留空则不推送，相关配置参见：[钉钉开放文档](https://open.dingtalk.com/document/group/custom-robot-access)



