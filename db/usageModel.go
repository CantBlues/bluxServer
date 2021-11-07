package db

import (
// "fmt"
)

type PhoneUsage struct {
	ID    uint `gorm:"primary_key"`
	Usage int  `gorm:"column:duration"`
	Node  int
	Appid int
}

type PhoneApp struct {
	ID      uint `gorm:"primary_key"`
	Name    string
	Package string `gorm:"column:package_name"`
}

func UsageLastDate() int {
	var last PhoneUsage
	DB.Last(&last) // whether need order by node
	return last.Node
}

func UsageReceiveData(last int, usages []PhoneUsage, apps []PhoneApp) {
	if len(usages) > 0 && len(apps) > 0 {
		DB.Delete(PhoneUsage{}, "node = ?", last)
		DB.Delete(PhoneApp{}, "id > ?", 0)
		DB.Omit("ID").CreateInBatches(usages, 100)
		DB.Create(&apps)
	}
}
