package main

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type (
	cfg struct {
		Timezone string // eg.Asia/Shanghai
		Accounts []*account
		FeiShu   map[string]*feiShu
		Msg      *msgTemplate
		Push     *push
	}

	// account 干饭账号配置
	account struct {
		Username, Password                 string
		FeiShuWebHook, FeiShuRobot, OpenID string
		EnableAllRobot                     bool // 启用所有机器人发送消息
		EnableWeekendPass                  bool // 启用周末提醒跳过
		EnableWeekendGreeting              bool // 启用周末问候消息
	}

	// feiShu 飞书机器人配置
	feiShu struct {
		AppID, AppSecret string
	}

	// 消息卡片模板参数
	msgTemplate struct {
		AppName        string   // 应用自称
		Lunch          string   // 午饭
		Dinner         string   // 晚饭
		OrderClosed    []string // 错过下单时间
		OrderAvailable []string // 需要下单的提示
		OrderSuccess   []string // 下单成功
	}

	// 推送设置
	push struct {
		Lunch    []string // 午餐消息
		Dinner   []string // 晚餐消息
		PreOrder []string // 预定消息
	}
)

var (
	pushDefaultLunch    = []string{"0 0 8 * * *", "0 50 11 * * *"}
	pushDefaultDinner   = []string{"0 0 15 * * *", "0 50 18 * * *"}
	pushDefaultPreOrder = []string{"0 0 20,21 * * *"}
)

var globalCfg = new(cfg)

func InitCfg() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("/etc")
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("load config failed", zap.Error(err))
		return
	}
	if err := viper.Unmarshal(globalCfg); err != nil {
		log.Panic("load config failed", zap.Error(err))
		return
	}
	if globalCfg.Push == nil {
		globalCfg.Push = &push{
			Lunch:    pushDefaultLunch,
			Dinner:   pushDefaultDinner,
			PreOrder: pushDefaultPreOrder,
		}
	} else {
		globalCfg.Push.init()
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

	// feiShu
	for k, v := range c.FeiShu {
		log.Info(
			"feiShu Robot info",
			zap.String("key", k),
			zap.String("id", v.AppID),
			zap.String("secret", v.AppSecret),
		)
	}

	// 语录
	log.Sugar().Infof("名字: %s | 午饭: %s | 晚饭: %s | 错过下单时间: %d | 等待下单: %d | 已经下单: %d",
		c.Msg.AppName, c.Msg.Lunch, c.Msg.Dinner,
		len(c.Msg.OrderClosed), len(c.Msg.OrderAvailable), len(c.Msg.OrderAvailable))

	// 推送时间设置
	log.Sugar().Infof("推送时间设置 | 午餐: %s | 晚餐: %s | 预定提醒: %s",
		strings.Join(c.Push.Lunch, ","),
		strings.Join(c.Push.Dinner, ","),
		strings.Join(c.Push.PreOrder, ","),
	)
}

func (p *push) init() {
	if len(p.Lunch) == 0 {
		p.Lunch = append(p.Lunch, pushDefaultLunch...)
	}
	if len(p.Dinner) == 0 {
		p.Dinner = append(p.Dinner, pushDefaultDinner...)
	}
	if len(p.PreOrder) == 0 {
		p.PreOrder = append(p.PreOrder, pushDefaultPreOrder...)
	}
}
