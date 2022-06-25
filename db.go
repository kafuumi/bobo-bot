package main

import (
	"database/sql"
	"os"

	"github.com/Hami-Lemon/bobo-bot/logger"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn   *sql.DB
	logger *logger.Logger
}

// NewDB 连接数据库
func NewDB(dbname string) *DB {
	_, err := os.Open(dbname)
	isNotExist := err != nil && os.IsNotExist(err)
	sqliteDB, err := sql.Open("sqlite3", dbname)
	if err != nil {
		mainLogger.Error("连接数据库失败！%v", err)
		return nil
	}
	err = sqliteDB.Ping()
	if err != nil {
		mainLogger.Error("连接数据库失败！name=%s, err=%v", dbname, err)
		return nil
	}
	mainLogger.Debug("连接 sqlite3 数据库 %s 成功", dbname)
	if isNotExist {
		mainLogger.Debug("创建 comment 表")
		//建表
		_, err = sqliteDB.Exec(`create table comment
(
    id integer primary key autoincrement ,
    oid       integer, -- 评论区oid
    type_code integer, -- 评论区type
    rpid      integer, -- 评论rpid
    ctime     integer, -- 评论发布时间
    msg       text,    -- 评论内容
    like_time integer, -- 点赞时间
    uid       integer, -- 评论发送者uid
    uname     text     -- 评论发送者用户名
);`)
		if err != nil {
			mainLogger.Error("建立 comment 表失败，%v", err)
			return nil
		}
		mainLogger.Debug("创建 follower 表")
		_, err = sqliteDB.Exec(`create table follower
(
    id    integer primary key autoincrement,
    uid   integer, -- 账号对应的uid
    ctime integer, -- 对应的时间点,时间戳形式单位秒
    fans  integer  -- 粉丝数
);`)
		if err != nil {
			mainLogger.Error("创建 follower 表失败，%v", err)
			return nil
		}
	}
	return &DB{
		conn:   sqliteDB,
		logger: logger.New("db", logLevel, logDst),
	}
}

// InsertComment 向数据库中插入评论数据
func (d *DB) InsertComment(comment Comment, likeTime int64) {
	stmt, err := d.conn.Prepare(`insert into comment
(oid, type_code, rpid, ctime, msg, like_time, uid, uname)
values (?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		d.logger.Error("InsertComment: prepare, %v", err)
		return
	}
	_, err = stmt.Exec(comment.oid, comment.typeCode, comment.replyId,
		comment.ctime, comment.msg, likeTime, comment.uid, comment.uname)
	if err != nil {
		d.logger.Error("InsertComment: exec, %v", err)
		return
	}
	d.logger.Debug("InsertComment 成功，oid=%d, rpid=%d, msg=%s",
		comment.oid, comment.replyId, comment.msg)
}

// InsertFollower 插入粉丝数
func (d *DB) InsertFollower(uid uint64, ctime int64, fans int) {
	stmt, err := d.conn.Prepare(`insert into follower(uid, ctime, fans)
values (?, ?, ?)`)
	if err != nil {
		d.logger.Error("InsertFollower: prepare, %v", err)
		return
	}
	_, err = stmt.Exec(uid, ctime, fans)
	if err != nil {
		d.logger.Error("InsertFollower: exec, %v", err)
		return
	}
	d.logger.Debug("InsertFollower 成功， uid=%d, ctime=%d, fans=%d", uid, ctime, fans)
}

func (d *DB) Close() {
	d.logger.Debug("断开连接")
	_ = d.conn.Close()
}
