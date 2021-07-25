package main

// GetDish 获取菜品信息
func GetDish(corp *CorpOrderUser) *Dish {
	if corp == nil {
		return nil
	}
	if len(corp.RestaurantItemList) == 0 ||
		len(corp.RestaurantItemList[0].DishItemList) == 0 {
		return nil
	}
	return corp.RestaurantItemList[0].DishItemList[0].Dish
}
