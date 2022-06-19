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
	commentDB, err := sql.Open("sqlite3", dbname)
	if err != nil {
		mainLogger.Error("连接数据库失败！%v", err)
		return nil
	}
	err = commentDB.Ping()
	if err != nil {
		mainLogger.Error("连接数据库失败！name=%s, err=%v", dbname, err)
	}
	mainLogger.Debug("连接 sqlite3 数据库 %s 成功", dbname)
	if isNotExist {
		mainLogger.Debug("comment 表不存在，建表")
		//建表
		_, err = commentDB.Exec(`create table comment
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
		}
	}
	return &DB{
		conn:   commentDB,
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
	}
	_, err = stmt.Exec(comment.oid, comment.typeCode, comment.replyId,
		comment.ctime, comment.msg, likeTime, comment.uid, comment.uname)
	if err != nil {
		d.logger.Error("InsertComment: exec, %v", err)
	}
	d.logger.Debug("InsertComment 成功，oid=%d, rpid=%d, msg=%s",
		comment.oid, comment.replyId, comment.msg)
}

func (d *DB) Close() {
	d.logger.Debug("断开连接")
	_ = d.conn.Close()
}
