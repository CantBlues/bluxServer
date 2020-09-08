package db

import (
	"blux/global"
	"blux/video"
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"os"
	"regexp"
)

// PATH is video file directory
const PATH string = global.PATH

// Video Model
type Video struct {
	gorm.Model
	ID           uint `gorm:"primary_key"`
	Md5          string
	Name         string
	Path         string
	Images       bool
	ProcessImage bool
}

// DB ...
func DB() (db *gorm.DB, err error) {
	dns := "root:34652402@tcp(127.0.0.1:3306)/blux?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dns), &gorm.Config{})
	fmt.Println("DBlink", db)
	return
}

// Fetch videos by paginate
func Fetch(r *http.Request, body []byte) string {
	var (
		videos []Video
		count  int64
	)
	type result struct {
		Data   []Video
		Count  int64
		Status bool
	}
	type query struct {
		Page int
	}
	request := new(query)
	json.Unmarshal(body, &request)
	db, _ := DB()
	db.Model(&Video{}).Count(&count)
	db.Scopes(paginate(request.Page)).Find(&videos)
	results := result{Data: videos, Count: count, Status: true}
	resultByte, _ := json.Marshal(results)
	return string(resultByte)
}

// ShutDownResponse return true
func ShutDownResponse() string {
	type result struct {
		Status bool
	}
	ret := result{Status: true}
	content, _ := json.Marshal(ret)
	return string(content)
}

// UpdateVideos is accept a string list to delete in database
func UpdateVideos(del, add map[string]string) {
	db, _ := DB()
	videos := []Video{}
	var delList, thumbList, processList []string
	for i := range del {
		delList = append(delList, i)
	}
	db.Where("Md5 IN ?", delList).Delete(Video{})
	for i, v := range add {
		name := regexp.MustCompile(`[^/\\\\]+$`).FindStringSubmatch(v)[0]
		videos = append(videos, Video{Md5: i, Name: name, Path: v[28:]})
	}
	if len(videos) > 0 {
		db.Create(&videos)
	}
	video.HandleVideosChange(del, add)
	for i := range add {
		if _, err := os.Stat(PATH + "imgs/" + i + "thumb.jpg"); err == nil {
			thumbList = append(thumbList, i)
		}
		if _, err := os.Stat(PATH + "imgs/" + i + "process.jpg"); err == nil {
			processList = append(processList, i)
		}
	}
	db.Model(&Video{}).Where("Md5 IN ?", thumbList).Select("images").Update("images", true)
	db.Model(&Video{}).Where("Md5 IN ?", processList).Select("process_image").Update("process_image", true)
}

func paginate(page int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}
		pageSize := 10
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
