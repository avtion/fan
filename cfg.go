package main

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type (
	cfg struct {
		Accounts []*Account
		FeiShu   map[string]*FeiShu
		Msg      *MsgTemplate
	}

	// Account 干饭账号配置
	Account struct {
		Username, Password                 string
		FeiShuWebHook, FeiShuRobot, OpenID string
		EnableAllRobot                     bool // 启用所有机器人发送消息
		DisableWeekendPass                 bool // 关闭周末提醒跳过
		EnableWeekendGreeting              bool // 启用周末问候消息
	}

	// FeiShu 飞书机器人配置
	FeiShu struct {
		AppID, AppSecret string
	}

	MsgTemplate struct {
		AppName        string   // 应用自称
		Lunch          string   // 午饭
		Dinner         string   // 晚饭
		OrderClosed    []string // 错过下单时间
		OrderAvailable []string // 需要下单的提示
		OrderSuccess   []string // 下单成功
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

	// 语录
	log.Sugar().Infof("名字: %s | 午饭: %s | 晚饭: %s | 错过下单时间: %d | 等待下单: %d | 已经下单: %d",
		c.Msg.AppName, c.Msg.Lunch, c.Msg.Dinner,
		len(c.Msg.OrderClosed), len(c.Msg.OrderAvailable), len(c.Msg.OrderAvailable))
}
