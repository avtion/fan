package main

import (
	"bytes"
	_ "embed"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type (
	cfg struct {
		Timezone  string // eg.Asia/Shanghai
		Accounts  []*account
		FeiShu    map[string]*feiShu
		Msg       msgTemplate
		Push      *push
		Collector collectorSetting
	}

	// account 干饭账号配置
	account struct {
		Username, Password                 string
		FeiShuWebHook, FeiShuRobot, OpenID string
		EnableAllRobot                     bool     // 启用所有机器人发送消息
		EnableWeekendPass                  bool     // 启用周末提醒跳过
		EnableWeekendGreeting              bool     // 启用周末问候消息
		TitleFilters                       []string // 地址过滤
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

	// 数据上报
	collectorSetting struct {
		Enable bool
		Mysql  string
	}
)

var (
	pushDefaultLunch    = []string{"0 0 8 * * *", "0 50 11 * * *"}
	pushDefaultDinner   = []string{"0 0 15 * * *", "0 50 18 * * *"}
	pushDefaultPreOrder = []string{"0 0 20,21 * * *"}
)

var globalCfg = new(cfg)

//go:embed config.yaml.example
var defaultConfigBytes []byte // 内嵌默认配置文件

func InitCfg() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("/etc")
	if err := viper.ReadInConfig(); err != nil {
		_ = viper.ReadConfig(bytes.NewReader(defaultConfigBytes))
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
	bindFlags()
	filterAccounts()

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
			//zap.String("password", v.Password),
			zap.String("webhook", v.FeiShuWebHook),
			zap.String("robot", v.FeiShuRobot),
			zap.String("openID", v.OpenID),
			zap.Strings("filters", v.TitleFilters),
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
	log.Sugar().Infof("名字: %s | 午饭: %s | 晚饭: %s | 错过下单时间语录: %d | 等待下单语录: %d | 已经下单语录: %d",
		c.Msg.AppName, c.Msg.Lunch, c.Msg.Dinner,
		len(c.Msg.OrderClosed), len(c.Msg.OrderAvailable), len(c.Msg.OrderAvailable))

	// 推送时间设置
	log.Info("推送时间设置",
		zap.Strings("午餐", c.Push.Lunch),
		zap.Strings("晚餐", c.Push.Dinner),
		zap.Strings("预定提醒", c.Push.PreOrder),
	)

	// 数据上报
	log.Info("数据上报", zap.Bool("是否启用", c.Collector.Enable), zap.String("mysql", c.Collector.Mysql))
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

// 绑定Flags
func bindFlags() {
	const (
		usernameKey = "username"
		passwordKey = "password"
		webhookKey  = "webhook"

		// 地址过滤
		filterNoHangXinKey = "nohx"
		filterNoGaoZhi     = "nogz"
		filterNoXingHui    = "noxh"
	)
	pflag.StringP(usernameKey, "u", "", "美餐用户名")
	pflag.StringP(passwordKey, "p", "", "美餐密码")
	pflag.StringP(webhookKey, "w", "", "飞书群机器人webhook")
	pflag.Bool(filterNoHangXinKey, false, "排除行信点餐点")
	pflag.Bool(filterNoGaoZhi, false, "排除高志点餐点")
	pflag.Bool(filterNoXingHui, false, "排除星辉点餐点")
	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)
	ac := &account{
		Username:      viper.GetString(usernameKey),
		Password:      viper.GetString(passwordKey),
		FeiShuWebHook: viper.GetString(webhookKey),
	}
	if ac.Username != "" && ac.Password != "" && ac.FeiShuWebHook != "" {
		globalCfg.Accounts = append(globalCfg.Accounts, ac)
		log.Info("Flag加载账号成功",
			zap.String("username", ac.Username), zap.String("hook", ac.FeiShuWebHook))
	}

	// 地点过滤
	if viper.GetBool(filterNoHangXinKey) {
		ac.TitleFilters = append(ac.TitleFilters, titleFilterHangXin)
	}
	if viper.GetBool(filterNoGaoZhi) {
		ac.TitleFilters = append(ac.TitleFilters, titleFilterGaoZhi)
	}
	if viper.GetBool(filterNoXingHui) {
		ac.TitleFilters = append(ac.TitleFilters, titleFilterXingHui)
	}
}

// 避免重复用户
func filterAccounts() {
	var (
		accountMapping = make(map[string]struct{}, len(globalCfg.Accounts))
		temp           = make([]*account, 0, len(globalCfg.Accounts))
	)
	for _, ac := range globalCfg.Accounts {
		if _, isExist := accountMapping[ac.Username]; isExist {
			continue
		}
		temp = append(temp, &account{
			Username:              ac.Username,
			Password:              ac.Password,
			FeiShuWebHook:         ac.FeiShuWebHook,
			FeiShuRobot:           ac.FeiShuRobot,
			OpenID:                ac.OpenID,
			EnableAllRobot:        ac.EnableAllRobot,
			EnableWeekendPass:     ac.EnableWeekendPass,
			EnableWeekendGreeting: ac.EnableWeekendGreeting,
			TitleFilters:          ac.TitleFilters,
		})
		accountMapping[ac.Username] = struct{}{}
	}
	globalCfg.Accounts = temp
	return
}

//————————————————
// 地址过滤
//————————————————

const (
	titleFilterHangXin = "行信"
	titleFilterXingHui = "星辉"
	titleFilterGaoZhi  = "高志"
)
