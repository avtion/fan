package main

import (
	"context"
	"testing"
)

func init() {
	InitMsgTest()
}

func InitMsgTest() {
	InitCfg()
	if len(globalCfg.Accounts) == 0 {
		panic("no account")
	}
	ac := getAc()
	if ac.FeiShuRobot == "" || ac.FeiShuWebHook == "" {
		panic("no robot")
	}
}

func getAc() *Account {
	return globalCfg.Accounts[0]
}

func Test_TextMsg(t *testing.T) {
	if err := SendMsg(context.Background(), getAc(), NewTextMsg("恰饭")); err != nil {
		t.Fatal(err)
	}
}

func Test_MarkdownMsg(t *testing.T) {
	msg := NewMarkdownMsg("午饭提醒",
		[]*Div{{Tag: "text", Text: "今天的午饭是"}},
		[]*Div{{Tag: "a", Text: "铁板黑椒鸡扒饭套餐", Href: "https://www.example.com/"}},
		[]*Div{{Tag: "text", Text: "这……这种事情，我当然知道，我……我可不是要说给你听的，我只是觉得你不知道的话太可怜了……对，就是这样……所以给我认认真真的记住！"}},
	)
	if err := SendMsg(context.Background(), getAc(), msg); err != nil {
		t.Fatal(err)
	}
}

func Test_CardMsg(t *testing.T) {
	msg := NewCardMsg(NewCardHeader("星期一午饭提醒", HeaderColorSuccess)).
		AddContents("主人&sim;今天是 **星期一** 啦！", "这是今天的午饭哦&sim;要按时吃饭！不然……不然就吃我！").
		AddAction(NewCardAction(actionTypePrimary, "🍗 铁板黑椒鸡扒饭套餐", "https://www.baidu.com/")).
		AddNotes("这……这种事情，我当然知道，我……我可不是要说给你听的，我只是觉得你不知道的话太可怜了……对，就是这样……所以给我认认真真的记住！")
	if err := SendMsg(context.Background(), getAc(), msg); err != nil {
		t.Fatal(err)
	}
}
