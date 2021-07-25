package main

import "github.com/valyala/fasttemplate"

var (
	// 消息卡片标题
	titleTemplate = fasttemplate.New("{{day}}{{dishType}}", "{{", "}}")
	// 问候
	greetingTemplate = fasttemplate.New("{{emoji}} 主人&sim;今天是 **{{day}}** 啦！", "{{", "}}")
	// 食物按钮
	fooTemplate = fasttemplate.New("{{emoji}} {{food}}", "{{", "}}")
	// 下单跳转
	jumpTemplate = fasttemplate.New(
		"https://www.meican.com/?date={{DateUnix}}&key={{UniqueId}}X{{TargetTime}}", "{{", "}}")
	// 吃饭提醒
	eatTipTemplate = fasttemplate.New(
		"这是{{dishType}}哦&sim;要按时吃饭！不然……不然就吃我！", "{{", "}}")
)
