package main

import (
	"context"
	"os"
	"testing"

	"github.com/golang-module/carbon"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

const (
	orderTestFilename = "order_test.txt"
)

func Test_Api(t *testing.T) {
	// 读取配置
	InitCfg()
	if len(globalCfg.Accounts) == 0 {
		t.Fatal("no accounts")
	}
	account := globalCfg.Accounts[0]

	api := NewAPI(context.Background(), nil)
	if err := api.Login(account.Username, account.Password); err != nil {
		t.Fatal(err)
	}

	orders, err := api.GetOrders(carbon.Yesterday().Time, carbon.Tomorrow().Time)
	if err != nil {
		t.Fatal(err)
	}
	ordersBytes, _ := jsoniter.Marshal(orders)
	if err = os.WriteFile(orderTestFilename, ordersBytes, 0600); err != nil {
		t.Fatal(err)
	}
}

func Test_FormatOrders(t *testing.T) {
	InitCfg()
	if _, err := os.Stat(orderTestFilename); err != nil {
		t.Fatal(err)
	}
	ordersBytes, err := os.ReadFile(orderTestFilename)
	if err != nil {
		t.Fatal(err)
	}
	orders := new(Orders)
	if err = jsoniter.Unmarshal(ordersBytes, orders); err != nil {
		t.Fatal(err)
	}
	log.Info("时间", zap.String("开始时间", orders.StartDate), zap.String("结束时间", orders.EndDate))

	for _, v := range orders.DateList {
		if len(v.CalendarItemList) == 0 {
			log.Error("缺乏午餐或晚餐信息", zap.Any("raw", v))
			continue
		}
		lunch, dinner := v.CalendarItemList[0], v.CalendarItemList[1]
		log.Sugar().Infof("午餐 %v", lunch)
		switch lunch.Status {
		case OrderStatusClosed:
		case OrderStatusAvailable:
		case OrderStatusOrder:
			// 已经下达订单
			log.Sugar().Infof("菜品信息 %v", GetDish(lunch.CorpOrderUser))
		}
		log.Sugar().Infof("饭餐 %v", dinner)
		switch lunch.Status {
		case OrderStatusClosed:
		case OrderStatusAvailable:
		case OrderStatusOrder:
			// 已经下达订单
			log.Sugar().Infof("菜品信息 %v", GetDish(dinner.CorpOrderUser))
		}
	}
}
