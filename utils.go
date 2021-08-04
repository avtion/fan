package main

import (
	"math/rand"
	"strings"
	"time"
)

type StringsPicker struct {
	ownRand *rand.Rand
}

func (p *StringsPicker) Pick(strArr ...string) string {
	return strArr[p.ownRand.Intn(len(strArr))]
}

var Picker = &StringsPicker{ownRand: rand.New(rand.NewSource(time.Now().Unix()))}

var (
	foodEmojis = []string{"🍏", "🍎", "🍐", "🍊", "🍋", "🍌", "🍉", "🍇", "🍓", "🍈", "🍒", "🍑", "🍍", "🥥", "🥝",
		"🍅", "🥑", "🍆", "🌶", "🥒", "🥦", "🌽", "🥕", "🥗", "🥔", "🍠", "🥜", "🍯", "🍞", "🥐", "🥖", "🥨", "🥞",
		"🧀", "🍗", "🍖", "🥩", "🍤", "🥚", "🥚", "🍳", "🥓", "🍔", "🍟", "🌭", "🍕", "🍝", "🥪", "🌮", "🌯", "🥙",
		"🍜", "🍲", "🥘", "🍥", "🍱", "🍣", "🍙", "🍛", "🍘", "🍚", "🥟", "🍢", "🍡", "🍧", "🍨", "🍦", "🍰", "🎂",
		"🥧", "🍮", "🍭", "🍬", "🍫", "🍿", "🍩", "🍪", "🥠", "☕", "🍵", "🥣", "🍼", "🥤", "🥛", "🍺", "🍻", "🍷",
		"🥂", "🥃", "🍸", "🍹", "🍾", "🍶", "🥄", "🍴", "🍽", "🥢"}
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

// GetCalendarItemByTitle 过滤掉不需要的Title地址
func GetCalendarItemByTitle(list []*CalendarItem, keyWord KeyWord, filters ...string) *CalendarItem {
	if keyWord == "" {
		return nil
	}
	for _, v := range list {
		if item := func() *CalendarItem {
			// 先排除过滤条件
			for _, filter := range filters {
				if strings.Contains(v.Title, filter) {
					return nil
				}
			}
			// 再找一下有没有对应关键字的
			if strings.Contains(v.Title, keyWord) {
				return v
			}
			return nil
		}(); item != nil {
			// 不为空就认为是找到了
			return item
		}
	}
	return nil
}

type KeyWord = string

const (
	KeyWordLunch  KeyWord = "午"
	KeyWordDinner         = "晚"
)

var _ = []KeyWord{KeyWordLunch, KeyWordDinner}
