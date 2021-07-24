package main

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// 目标URL
//_______________________________________________________________________

const (
	BaseHost       = "https://meican.com"
	LoginURL       = BaseHost + "/preference/preorder/api/v2.0/oauth/token"
	OrderURL       = BaseHost + "/preorder/api/v2.1/calendaritems/list"
	RestaurantsAPI = BaseHost + "/preorder/api/v2.1/restaurants/list"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// 响应结构体
//_______________________________________________________________________

type LoginResp struct {
	AccessToken       string `json:"access_token"`
	TokenType         string `json:"token_type"`
	RefreshToken      string `json:"refresh_token"`
	ExpiresIn         int    `json:"expires_in"`
	NeedResetPassword bool   `json:"need_reset_password"`
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// 订单查询
//_______________________________________________________________________

type (
	// Orders 订单
	Orders struct {
		StartDate string `json:"startDate"` // 开始时间
		EndDate   string `json:"endDate"`   // 结束时间
		DateList  []struct {
			Date             string `json:"date"` // 日期
			CalendarItemList []struct {
				TargetTime    int64          `json:"targetTime"` // 目标时间 eg.1625445000000
				Title         string         `json:"title"`      // 描述
				Status        string         `json:"status"`
				Reason        string         `json:"reason"`
				UserTab       *UserTab       `json:"userTab"`
				OpeningTime   *OpeningTime   `json:"openingTime"`
				CorpOrderUser *CorpOrderUser `json:"corpOrderUser"`
			} `json:"calendarItemList"`
		} `json:"dateList"`
	}

	// OpeningTime 预定时间
	OpeningTime struct {
		UniqueId         string `json:"uniqueId"`
		Name             string `json:"name"`
		OpenTime         string `json:"openTime"`  // 开放预定时间
		CloseTime        string `json:"closeTime"` // 预定结束时间
		DefaultAlarmTime string `json:"defaultAlarmTime"`
		PostboxOpenTime  string `json:"postboxOpenTime"`
	}

	// UserTab 用户下单选择项
	UserTab struct {
		Corp struct {
			UniqueId                  string  `json:"uniqueId"`
			UseCloset                 bool    `json:"useCloset"`
			Name                      string  `json:"name"` // 订餐地点
			Namespace                 string  `json:"namespace"`
			PriceVisible              bool    `json:"priceVisible"`
			ShowPrice                 bool    `json:"showPrice"`
			PriceLimit                float64 `json:"priceLimit"`
			PriceLimitInCent          float64 `json:"priceLimitInCent"`
			AcceptCashPaymentToMeican bool    `json:"acceptCashPaymentToMeican"`
			AlwaysOpen                bool    `json:"alwaysOpen"`
			AddressList               []struct {
				UniqueId        string `json:"uniqueId"`        // 地点唯一ID
				Address         string `json:"address"`         // 详细地址
				CorpAddressCode string `json:"corpAddressCode"` // 地点代码 eg.A\B\C
				PickUpLocation  string `json:"pickUpLocation"`  // 大致地址 eg.地点+楼层
			} `json:"addressList"` // 订餐地点列表
		} `json:"corp"`
		Latitude     string `json:"latitude"`
		Longitude    string `json:"longitude"`
		Name         string `json:"name"`
		LastUsedTime int64  `json:"lastUsedTime"`
		UniqueId     string `json:"uniqueId"`
	}

	// CorpOrderUser 用户已经选择的餐品信息
	CorpOrderUser struct {
		IsLegacyPay        bool   `json:"isLegacyPay"`
		PayStatus          string `json:"payStatus"`
		RestaurantItemList []struct {
			UniqueId     string `json:"uniqueId"`
			DishItemList []struct {
				Dish  *Dish `json:"dish"`
				Count int   `json:"count"`
			} `json:"dishItemList"`
		} `json:"restaurantItemList"`
		Corp struct {
			UniqueId                  string      `json:"uniqueId"`
			UseCloset                 bool        `json:"useCloset"`
			Name                      string      `json:"name"`
			Namespace                 string      `json:"namespace"`
			PriceVisible              bool        `json:"priceVisible"`
			ShowPrice                 bool        `json:"showPrice"`
			PriceLimit                int         `json:"priceLimit"`
			PriceLimitInCent          int         `json:"priceLimitInCent"`
			AcceptCashPaymentToMeican bool        `json:"acceptCashPaymentToMeican"`
			AlwaysOpen                bool        `json:"alwaysOpen"`
			AddressList               interface{} `json:"addressList"`
		} `json:"corp"`
		ReadyToDelete                 bool   `json:"readyToDelete"`
		ActionRequiredLevel           string `json:"actionRequiredLevel"`
		CorpOrderStatus               string `json:"corpOrderStatus"`
		ShowPrice                     bool   `json:"showPrice"`
		UnpaidUserToMeicanPrice       string `json:"unpaidUserToMeicanPrice"`
		UnpaidUserToMeicanPriceInCent int    `json:"unpaidUserToMeicanPriceInCent"`
		PaidUserToMeicanPrice         string `json:"paidUserToMeicanPrice"`
		PaidUserToMeicanPriceInCent   int    `json:"paidUserToMeicanPriceInCent"`
		Timestamp                     int64  `json:"timestamp"`
		UniqueId                      string `json:"uniqueId"`
	}

	// Dish 餐品
	Dish struct {
		Name                 string `json:"name"`                // 餐名
		PriceInCent          int    `json:"priceInCent"`         // 费用
		OriginalPriceInCent  int    `json:"originalPriceInCent"` // 原始价格
		IsSection            bool   `json:"isSection"`
		ActionRequiredLevel  string `json:"actionRequiredLevel"`
		ActionRequiredReason string `json:"actionRequiredReason"`
		Id                   int    `json:"id"`
	}
)

type OrderStatus = string

const (
	OrderStatusClosed    OrderStatus = "CLOSED"    // 不能点餐
	OrderStatusOrder     OrderStatus = "ORDER"     // 已经点餐
	OrderStatusAvailable OrderStatus = "AVAILABLE" // 开放点餐
)

var _ = []OrderStatus{OrderStatusClosed, OrderStatusOrder, OrderStatusAvailable}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// 下单
//_______________________________________________________________________

// Restaurants 餐厅
type Restaurants struct {
	NoMore         bool   `json:"noMore"`
	TargetTime     string `json:"targetTime"`
	RestaurantList []struct {
		UniqueId           string      `json:"uniqueId"` // 餐厅ID
		Name               string      `json:"name"`     // 餐厅名
		Tel                string      `json:"tel"`
		Rating             int         `json:"rating"`
		DeliveryRangeMeter interface{} `json:"deliveryRangeMeter"`
		MinimumOrder       int         `json:"minimumOrder"`
		Latitude           float64     `json:"latitude"`
		Longitude          float64     `json:"longitude"`
		Warning            string      `json:"warning"`
		OpeningTime        string      `json:"openingTime"`
		OnlinePayment      bool        `json:"onlinePayment"`
		Open               bool        `json:"open"`
		AvailableDishCount int         `json:"availableDishCount"` // 可用餐品数量
		DishLimit          int         `json:"dishLimit"`
		RestaurantStatus   int         `json:"restaurantStatus"`
		RemarkEnabled      bool        `json:"remarkEnabled"`
	} `json:"restaurantList"`
}

type Dishes struct {
	AdditionalInfo struct {
		Address        string `json:"address"`
		AssessDate     string `json:"assessDate"`
		AssessEndDate  string `json:"assessEndDate"`
		BusinessType   string `json:"businessType"`
		CityName       string `json:"cityName"`
		CityUrl        string `json:"cityUrl"`
		CompanyName    string `json:"companyName"`
		District       string `json:"district"`
		Level          string `json:"level"`
		LicenseNumber  string `json:"licenseNumber"`
		Representative string `json:"representative"`
	} `json:"additionalInfo"`
	Assessment struct {
		AssessmentIconUrl string `json:"assessmentIconUrl"`
		Fields            []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"fields"`
	} `json:"assessment"`
	AvailableDishCount int         `json:"availableDishCount"`
	BusinessLicenseUrl string      `json:"businessLicenseUrl"`
	CorpRestaurantId   int         `json:"corpRestaurantId"`
	DeliveryFeeInCent  interface{} `json:"deliveryFeeInCent"`
	DeliveryRange      interface{} `json:"deliveryRange"`
	DeliveryRangeMeter interface{} `json:"deliveryRangeMeter"`
	DishLimit          int         `json:"dishLimit"`
	DishList           []struct {
		DishSectionId       int    `json:"dishSectionId"` // 餐品选择ID
		Id                  int    `json:"id"`            // 餐品ID
		IsSection           bool   `json:"isSection"`
		Name                string `json:"name"` // 餐名
		OriginalPriceInCent int    `json:"originalPriceInCent"`
		PriceInCent         int    `json:"priceInCent"` // 价格
		PriceString         string `json:"priceString"`
	} `json:"dishList"`
	Latitude                      float64     `json:"latitude"`
	Longitude                     float64     `json:"longitude"`
	MinimumOrder                  interface{} `json:"minimumOrder"`
	MyRegularDishIdList           []int       `json:"myRegularDishIdList"`
	Name                          string      `json:"name"`
	OnlinePayment                 bool        `json:"onlinePayment"`
	Open                          bool        `json:"open"`
	OpeningTime                   string      `json:"openingTime"`
	OthersRegularDishIdList       []int       `json:"othersRegularDishIdList"`
	OthersRegularDishIdListSource string      `json:"othersRegularDishIdListSource"`
	Rating                        int         `json:"rating"`
	RemarkEnabled                 bool        `json:"remarkEnabled"`
	RestaurantId                  int         `json:"restaurantId"`
	RestaurantStatus              int         `json:"restaurantStatus"`
	SanitationCertificateUrl      string      `json:"sanitationCertificateUrl"`
	SectionList                   []struct {
		Id         int    `json:"id"`
		DishIdList []int  `json:"dishIdList"`
		Name       string `json:"name"`
	} `json:"sectionList"`
	ShowPrice  bool   `json:"showPrice"`
	TargetTime string `json:"targetTime"`
	Tel        string `json:"tel"`
	UniqueId   string `json:"uniqueId"`
	Warning    string `json:"warning"`
}
