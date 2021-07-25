package main

import (
	"context"
	"testing"

	"github.com/golang-module/carbon"
	"go.uber.org/zap"
)

func Test_PushMsg(t *testing.T) {
	InitCfg()
	InitRobots(globalCfg)
	api := NewAPI(context.Background(), NewHTTPClient())
	if err := api.Login(getAc().Username, getAc().Password); err != nil {
		t.Fatal(err)
	}
	err := pushMsg(context.Background(), getAc(), api, carbon.Now(), false, DishTypeDinner)
	if err != nil {
		log.Error("晚餐消息推送失败", zap.Any("用户", getAc()), zap.Error(err))
		return
	}
	return
}

func Test_PushTomorrowMsg(t *testing.T) {
	InitCfg()
	InitRobots(globalCfg)
	api := NewAPI(context.Background(), NewHTTPClient())
	if err := api.Login(getAc().Username, getAc().Password); err != nil {
		t.Fatal(err)
	}
	err := pushMsg(context.Background(), getAc(), api, carbon.Now().AddDays(2), true, DishTypeUndefined)
	if err != nil {
		log.Error("预定提醒消息推送失败", zap.Any("用户", getAc()), zap.Error(err))
		return
	}
	return
}
