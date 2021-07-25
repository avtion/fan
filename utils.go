package main

import (
	"math/rand"
	"time"
	"unsafe"
)

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

type StringsPicker struct {
	ownRand *rand.Rand
}

func (p *StringsPicker) Pick(strArr ...string) string {
	return strArr[p.ownRand.Intn(len(strArr))]
}

var Picker = &StringsPicker{ownRand: rand.New(rand.NewSource(time.Now().Unix()))}

var (
	foodEmojis = "🍏🍎🍐🍊🍋🍌🍉🍇🍓🍈🍒🍑🍍🥥🥝🍅🥑🍆🌶🥒🥦🌽🥕🥗🥔🍠🥜🍯🍞🥐🥖🥨🥞🧀🍗🍖🥩🍤🥚🥚🍳🥓🍔🍟" +
		"🌭🍕🍝🥪🌮🌯🥙🍜🍲🥘🍥🍱🍣🍙🍛🍘🍚🥟🍢🍡🍧🍨🍦🍰🎂🥧🍮🍭🍬🍫🍿🍩🍪🥠☕🍵🥣🍼🥤🥛🍺🍻🍷🥂🥃🍸🍹🍾🍶🥄🍴🍽🥢"
	loveEmojis = []string{"💝", "💞", "💟", "💘", "❤"}
)

func GetWeekStr(dayOfWeek int) string {
	switch dayOfWeek {
	case 1:
		return "元气满满の星期一"
	case 2:
		return "咪咪の星期二"
	case 3:
		return "一览纵山小の星期三"
	case 4:
		return "抚剑独怆神の星期四"
	case 5:
		return "人间遍种自由花の星期五"
	case 6:
		return "起飞の星期六"
	case 7:
		return "陪你看星星の星期日"
	}
	return ""
}
