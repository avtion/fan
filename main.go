package main

import (
	"context"
	"errors"
	"time"
	_ "time/tzdata"

	"github.com/golang-module/carbon"
	"github.com/google/gops/agent"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cast"
	"github.com/valyala/fasttemplate"
	"go.uber.org/zap"
)

func init() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Error("gops客户端初始化失败", zap.Error(err))
		return
	}
}

func initTimezone() error {
	if globalCfg.Timezone == "" {
		return nil
	}
	var err error
	time.Local, err = time.LoadLocation(globalCfg.Timezone)
	if err != nil {
		log.Error("加载时区失败", zap.Error(err), zap.String("时区设置", globalCfg.Timezone))
		return err
	}
	log.Info("加载时区成功", zap.String("时区设置", globalCfg.Timezone))
	return nil
}

func main() {
	InitCfg()
	InitRobots(globalCfg)
	_ = initTimezone()
	if len(globalCfg.Accounts) == 0 {
		log.Error("警告: 没有加载账号", zap.Any("配置", globalCfg))
		// return 不退出程序
	}
	ctx := context.Background()

	schedule := cron.New(
		cron.WithLocation(time.Local),
		cron.WithLogger(getCronLogger()),
		cron.WithSeconds(),
	)

	for index := range globalCfg.Accounts {
		api := NewAPI(ctx, NewHTTPClient())
		if err := api.Login(globalCfg.Accounts[index].Username, globalCfg.Accounts[index].Password); err != nil {
			log.Error("用户登录失败",
				zap.Error(err), zap.String("username", globalCfg.Accounts[index].Username))
			continue
		}

		// 推送午餐消息
		for _, v := range globalCfg.Push.Lunch {
			if _, err := schedule.AddFunc(v, func() {
				// 尝试登录一下
				if err := api.Login(globalCfg.Accounts[index].Username, globalCfg.Accounts[index].Password); err != nil {
					log.Error("用户登录失败",
						zap.Error(err), zap.String("username", globalCfg.Accounts[index].Username))
					return
				}
				err := pushMsg(ctx, globalCfg.Accounts[index], api, carbon.Now(), false, DishTypeLunch)
				if err != nil {
					log.Error("午餐消息推送失败", zap.Any("用户", globalCfg.Accounts[index]), zap.Error(err))
					return
				}
				return
			}); err != nil {
				log.Error("添加调度任务失败", zap.Error(err), zap.String("调度设置", v))
				continue
			}
		}

		// 晚餐消息
		for _, v := range globalCfg.Push.Dinner {
			if _, err := schedule.AddFunc(v, func() {
				// 尝试登录一下
				if err := api.Login(globalCfg.Accounts[index].Username, globalCfg.Accounts[index].Password); err != nil {
					log.Error("用户登录失败",
						zap.Error(err), zap.String("username", globalCfg.Accounts[index].Username))
					return
				}
				err := pushMsg(ctx, globalCfg.Accounts[index], api, carbon.Now(), false, DishTypeDinner)
				if err != nil {
					log.Error("晚餐消息推送失败", zap.Any("用户", globalCfg.Accounts[index]), zap.Error(err))
					return
				}
				return
			}); err != nil {
				log.Error("添加调度任务失败", zap.Error(err), zap.String("调度设置", v))
				continue
			}
		}

		// 预定提醒消息
		for _, v := range globalCfg.Push.PreOrder {
			if _, err := schedule.AddFunc(v, func() {
				// 尝试登录一下
				if err := api.Login(globalCfg.Accounts[index].Username, globalCfg.Accounts[index].Password); err != nil {
					log.Error("用户登录失败",
						zap.Error(err), zap.String("username", globalCfg.Accounts[index].Username))
					return
				}
				err := pushMsg(ctx, globalCfg.Accounts[index], api, carbon.Tomorrow(), true, DishTypeUndefined)
				if err != nil {
					log.Error("预定提醒消息推送失败", zap.Any("用户", globalCfg.Accounts[index]), zap.Error(err))
					return
				}
				return
			}); err != nil {
				log.Error("添加调度任务失败", zap.Error(err), zap.String("调度设置", v))
				continue
			}
		}
	}

	log.Info("调度器任务加载完毕", zap.Int("任务数量", len(schedule.Entries())))
	schedule.Run()
}

// 推送消息
func pushMsg(ctx context.Context, account *account, api *Api, day carbon.Carbon,
	isPreOrder bool, lunchOrDinner DishType) error {

	// 如果今天是周末就终止流程
	if day.IsWeekend() && account.EnableWeekendPass {
		// 检查用户是否启用周日问候
		if account.EnableWeekendGreeting && !isPreOrder {
			log.Info("用户启用了周末问候", zap.String("username", account.Username))

			weekdayStr := GetWeekStr(day.DayOfWeek())
			msg := NewCardMsg(NewCardHeader(titleTemplate.ExecuteString(map[string]interface{}{
				"day": weekdayStr,
			}), HeaderColorDefault)).AddContents(greetingTemplate.ExecuteString(map[string]interface{}{
				"emoji": Picker.Pick(loveEmojis...),
				"day":   weekdayStr,
			}), "啊咧，今天居然是 **休息日** ~主人要好好休息哦!").
				AddNotes(Picker.Pick(globalCfg.Msg.OrderSuccess...))
			_ = SendMsg(ctx, account, msg)
		}
		return nil
	}

	// 获取订单
	orders, err := api.GetOrders(day.Time, day.Time)
	if err != nil {
		log.Error("获取订单失败", zap.Error(err))
		return err
	}
	if len(orders.DateList) == 0 || len(orders.DateList[0].CalendarItemList) < 2 {
		log.Error("订单解析失败", zap.Any("订单信息", orders))
		return errors.New("订单解析失败")
	}
	dateItem := orders.DateList[0]

	// 构建消息
	var msg Msg
	switch isPreOrder {
	case true:
		// 预定订单
		msg, err = buildPreOrderMsg(dateItem)
	case false:
		// 非预定订单
		switch lunchOrDinner {
		case DishTypeLunch:
			// 午餐
			msg, err = buildTodayMsg(dateItem.Date, globalCfg.Msg.Lunch, dateItem.CalendarItemList[0])
		case DishTypeDinner:
			// 晚餐
			msg, err = buildTodayMsg(dateItem.Date, globalCfg.Msg.Dinner, dateItem.CalendarItemList[1])
		}
	}
	if err != nil {
		log.Error("构建推送消息失败", zap.Error(err), zap.String("username", account.Username))
		return err
	}
	if msg == nil {
		log.Info("没有需要推送的消息", zap.String("username", account.Username))
		return nil
	}
	// 发送消息
	if err = SendMsg(ctx, account, msg); err != nil {
		log.Error("发送推送失败", zap.Error(err), zap.String("username", account.Username))
		return err
	}
	log.Info("消息推送成功",
		zap.String("username", account.Username),
		zap.String("time", day.ToDateTimeString()),
		zap.Bool("是否预定提醒", isPreOrder),
		zap.Uint("餐类", lunchOrDinner),
	)
	return nil
}

// 第二天的推送消息
func buildPreOrderMsg(item *DateItem) (Msg, error) {
	lunch, dinner := item.CalendarItemList[0], item.CalendarItemList[1]
	// 如果不需要订餐就没有提醒了吧
	if lunch.Status != OrderStatusAvailable && dinner.Status != OrderStatusAvailable {
		return nil, nil
	}

	dayT, _ := time.Parse("2006-01-02", item.Date)
	weekdayStr := GetWeekStr(carbon.CreateFromTimestamp(dayT.Unix()).DayOfWeek())
	msg := NewCardMsg(NewCardHeader(titleTemplate.ExecuteString(map[string]interface{}{
		"day":      weekdayStr,
		"dishType": "预定提醒",
	}), HeaderColorDefault)).AddContents(
		fasttemplate.New("{{emoji}} 主人&sim;明天是 **{{day}}** 啦！", "{{", "}}").ExecuteString(
			map[string]interface{}{"emoji": Picker.Pick(loveEmojis...), "day": weekdayStr}))

	var isNeedToOrder bool // 用来加备注模块信息的
	// 追加午餐信息
	switch lunch.Status {
	case OrderStatusOrder:
		msg.AddContents(fasttemplate.New(
			"明天的午餐是 {{emoji}} **{{food}}**", "{{", "}}").
			ExecuteString(map[string]interface{}{
				"emoji": Picker.Pick(foodEmojis...),
				"food":  GetDish(lunch.CorpOrderUser).Name,
			}))
	case OrderStatusAvailable:
		msg.AddContents("主人是不是忘记点明天的**午餐**……了？").
			AddAction(NewCardAction(actionTypePrimary, "别拉我，我要点午餐！",
				jumpTemplate.ExecuteString(map[string]interface{}{
					"DateUnix":   cast.ToString(dayT.Unix()),
					"UniqueId":   lunch.UserTab.UniqueId,
					"TargetTime": cast.ToString(lunch.TargetTime),
				})))
		isNeedToOrder = true
	}

	// 追加晚餐信息
	switch dinner.Status {
	case OrderStatusOrder:
		msg.AddContents(fasttemplate.New(
			"明天的晚餐是 {{emoji}} **{{food}}**", "{{", "}}").
			ExecuteString(map[string]interface{}{
				"emoji": Picker.Pick(foodEmojis...),
				"food":  GetDish(dinner.CorpOrderUser).Name,
			}))
	case OrderStatusAvailable:
		msg.AddContents("主人是不是忘记点明天的**晚餐**……了？").
			AddAction(NewCardAction(actionTypePrimary, "那个……很抱歉，我马上点晚餐",
				jumpTemplate.ExecuteString(map[string]interface{}{
					"DateUnix":   cast.ToString(dayT.Unix()),
					"UniqueId":   dinner.UserTab.UniqueId,
					"TargetTime": cast.ToString(dinner.TargetTime),
				})))
		isNeedToOrder = true
	}

	if isNeedToOrder {
		msg.AddNotes(Picker.Pick(globalCfg.Msg.OrderAvailable...))
	} else {
		msg.AddNotes(Picker.Pick(globalCfg.Msg.OrderSuccess...))
	}
	return msg, nil
}

// 当天的推送消息
func buildTodayMsg(date string, dishType string, item *CalendarItem) (Msg, error) {
	var (
		msg Msg
		err error
	)
	dayT, _ := time.Parse("2006-01-02", date)
	weekdayStr := GetWeekStr(carbon.CreateFromTimestamp(dayT.Unix()).DayOfWeek())
	greeting := greetingTemplate.ExecuteString(map[string]interface{}{
		"emoji": Picker.Pick(loveEmojis...),
		"day":   weekdayStr,
	})

	switch item.Status {
	case OrderStatusClosed:
		msg = NewCardMsg(
			NewCardHeader(titleTemplate.ExecuteString(map[string]interface{}{
				"day":      weekdayStr,
				"dishType": dishType,
			}), HeaderColorFailed)).
			AddContents(greeting, "这一顿饭……可能饭酱没办法陪你了，主人忘记点饭了！").
			AddNotes(Picker.Pick(globalCfg.Msg.OrderClosed...))
	case OrderStatusAvailable:
		msg = NewCardMsg(
			NewCardHeader(titleTemplate.ExecuteString(map[string]interface{}{
				"day":      weekdayStr,
				"dishType": dishType,
			}), HeaderColorDefault)).
			AddContents(greeting, "球球了，可以抽空点一下饭嘛……？不吃饭的话，我会心疼你的~").
			AddAction(NewCardAction(actionTypePrimary, "我要点餐",
				jumpTemplate.ExecuteString(map[string]interface{}{
					"DateUnix":   cast.ToString(dayT.Unix()),
					"UniqueId":   item.UserTab.UniqueId,
					"TargetTime": cast.ToString(item.TargetTime),
				}))).
			AddNotes(Picker.Pick(globalCfg.Msg.OrderAvailable...))
	case OrderStatusOrder:
		dish := GetDish(item.CorpOrderUser)
		msg = NewCardMsg(
			NewCardHeader(titleTemplate.ExecuteString(map[string]interface{}{
				"day":      weekdayStr,
				"dishType": dishType,
			}), HeaderColorSuccess)).
			AddContents(greeting, eatTipTemplate.ExecuteString(map[string]interface{}{
				"dishType": dishType,
			})).
			AddAction(NewCardAction(actionTypePrimary, fooTemplate.ExecuteString(map[string]interface{}{
				"emoji": Picker.Pick(foodEmojis...),
				"food":  dish.Name,
			}), "https://www.meican.com/")).
			AddNotes(Picker.Pick(globalCfg.Msg.OrderSuccess...))
	default:
		log.Error("unknown order status", zap.String("status", item.Status))
		err = errors.New("unknown order status")
	}
	return msg, err
}
