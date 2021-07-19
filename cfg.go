package main

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type (
	cfg struct {
		Accounts []*Account
		FeiShu   map[string]*FeiShu
	}

	// Account 干饭账号配置
	Account struct {
		Username, Password                 string
		FeiShuWebHook, FeiShuRobot, OpenID string
	}

	// FeiShu 飞书机器人配置
	FeiShu struct {
		AppID, AppSecret string
	}
)

var globalCfg = new(cfg)

func InitCfg() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("load config failed", zap.Error(err))
		return
	}
	if err := viper.Unmarshal(globalCfg); err != nil {
		log.Panic("load config failed", zap.Error(err))
		return
	}

	// debug print
	printConfig(globalCfg)
}

// 打印配置信息
func printConfig(c *cfg) {
	if c == nil {
		return
	}

	// account info
	for _, v := range c.Accounts {
		log.Info(
			"account info",
			zap.String("username", v.Username),
			zap.String("password", v.Password),
			zap.String("webhook", v.FeiShuWebHook),
			zap.String("robot", v.FeiShuRobot),
			zap.String("openID", v.OpenID),
		)
	}

	// FeiShu
	for k, v := range c.FeiShu {
		log.Info(
			"FeiShu Robot info",
			zap.String("key", k),
			zap.String("id", v.AppID),
			zap.String("secret", v.AppSecret),
		)
	}
}
