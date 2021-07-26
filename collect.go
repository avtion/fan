package main

import (
	"context"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

//——————————————————————————
// 数据上报
//——————————————————————————

// Collector 负责数据上报的对象
type (
	Collector struct {
		ctx   context.Context
		mysql *sqlx.DB
	}
	CollectorOpt func(c *Collector)
)

var globalCollector *Collector

func NewCollector(ctx context.Context, opts ...CollectorOpt) *Collector {
	c := &Collector{
		ctx: ctx,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// CollectorOptMySQL 新增MySQL数据库初始化
func CollectorOptMySQL(dsn string) CollectorOpt {
	return func(c *Collector) {
		if dsn == "" {
			return
		}

		db, err := sqlx.Open("mysql", dsn)
		if err != nil {
			log.Error("数据收集初始化失败", zap.Error(err))
			return
		}
		if err = db.Ping(); err != nil {
			log.Error("数据收集无法连接数据库", zap.Error(err))
			return
		}
		db.SetMaxOpenConns(100)
		db.SetMaxIdleConns(25)
		c.mysql = db
		log.Info("数据上报初始化MySQL数据库成功")
	}
}
