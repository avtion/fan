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
	foodEmojis = []string{"๐", "๐", "๐", "๐", "๐", "๐", "๐", "๐", "๐", "๐", "๐", "๐", "๐", "๐ฅฅ", "๐ฅ",
		"๐", "๐ฅ", "๐", "๐ถ", "๐ฅ", "๐ฅฆ", "๐ฝ", "๐ฅ", "๐ฅ", "๐ฅ", "๐ ", "๐ฅ", "๐ฏ", "๐", "๐ฅ", "๐ฅ", "๐ฅจ", "๐ฅ",
		"๐ง", "๐", "๐", "๐ฅฉ", "๐ค", "๐ฅ", "๐ฅ", "๐ณ", "๐ฅ", "๐", "๐", "๐ญ", "๐", "๐", "๐ฅช", "๐ฎ", "๐ฏ", "๐ฅ",
		"๐", "๐ฒ", "๐ฅ", "๐ฅ", "๐ฑ", "๐ฃ", "๐", "๐", "๐", "๐", "๐ฅ", "๐ข", "๐ก", "๐ง", "๐จ", "๐ฆ", "๐ฐ", "๐",
		"๐ฅง", "๐ฎ", "๐ญ", "๐ฌ", "๐ซ", "๐ฟ", "๐ฉ", "๐ช", "๐ฅ ", "โ", "๐ต", "๐ฅฃ", "๐ผ", "๐ฅค", "๐ฅ", "๐บ", "๐ป", "๐ท",
		"๐ฅ", "๐ฅ", "๐ธ", "๐น", "๐พ", "๐ถ", "๐ฅ", "๐ด", "๐ฝ", "๐ฅข"}
	loveEmojis = []string{"๐", "๐", "๐", "๐", "โค"}
)

func GetWeekStr(dayOfWeek int) string {
	switch dayOfWeek {
	case 1:
		return "ๅๆฐๆปกๆปกใฎๆๆไธ"
	case 2:
		return "ๅชๅชใฎๆๆไบ"
	case 3:
		return "ไธ่ง็บตๅฑฑๅฐใฎๆๆไธ"
	case 4:
		return "ๆๅ็ฌๆ็ฅใฎๆๆๅ"
	case 5:
		return "ไบบ้ด้็ง่ช็ฑ่ฑใฎๆๆไบ"
	case 6:
		return "่ตท้ฃใฎๆๆๅญ"
	case 7:
		return "้ชไฝ ็ๆๆใฎๆๆๆฅ"
	}
	return ""
}

// GetCalendarItemByTitle ่ฟๆปคๆไธ้่ฆ็Titleๅฐๅ
func GetCalendarItemByTitle(list []*CalendarItem, keyWord KeyWord, filters ...string) *CalendarItem {
	if keyWord == "" {
		return nil
	}
	for _, v := range list {
		if item := func() *CalendarItem {
			// ๅๆ้ค่ฟๆปคๆกไปถ
			for _, filter := range filters {
				if strings.Contains(v.Title, filter) {
					return nil
				}
			}
			// ๅๆพไธไธๆๆฒกๆๅฏนๅบๅณ้ฎๅญ็
			if strings.Contains(v.Title, keyWord) {
				return v
			}
			return nil
		}(); item != nil {
			// ไธไธบ็ฉบๅฐฑ่ฎคไธบๆฏๆพๅฐไบ
			return item
		}
	}
	return nil
}

type KeyWord = string

const (
	KeyWordLunch  KeyWord = "ๅ"
	KeyWordDinner         = "ๆ"
)

var _ = []KeyWord{KeyWordLunch, KeyWordDinner}
