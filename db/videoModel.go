package db

import (
	"blux/global"
	"blux/video/dll"
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sync"
	"syscall"
	"time"
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
	CreatedAt    time.Time
}

// Audio Model
type Audio struct {
	gorm.Model
	ID        uint `gorm:"primary_key"`
	Md5       string
	Name      string
	Path      string
	CreatedAt    time.Time
}

// Info Model stroge md5 of audios and videos
type Info struct {
	ID    uint `gorm:"primary_key"`
	Audio string
	Video string
}

// DB ...
var DB *gorm.DB

func init() {
	dns := "root:34652402@tcp(127.0.0.1:3306)/blux?charset=utf8mb4&parseTime=True&loc=Local"
	DB, _ = gorm.Open(mysql.Open(dns), &gorm.Config{})
	fmt.Println("DBlink", DB)
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

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
	DB.Model(&Video{}).Count(&count)
	DB.Scopes(paginate(request.Page)).Find(&videos)
	results := result{Data: videos, Count: count, Status: true}
	resultByte, _ := json.Marshal(results)
	return string(resultByte)
}

func FetchAudio(r *http.Request, body []byte) string {
	var (
		audios []Audio
		count  int64
	)
	type result struct {
		Data   []Audio
		Count  int64
		Status bool
	}
	type query struct {
		Page int
	}
	request := new(query)
	json.Unmarshal(body, &request)
	DB.Model(&Audio{}).Count(&count)
	DB.Scopes(paginate(request.Page)).Find(&audios)
	results := result{Data: audios, Count: count, Status: true}
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
	videos := []Video{}
	var delList, thumbList, processList []string
	for i := range del {
		delList = append(delList, i)
	}
	DB.Where("Md5 IN ?", delList).Delete(Video{})
	for i, v := range add {
		name := regexp.MustCompile(`[^/\\\\]+$`).FindStringSubmatch(v)[0]
		videos = append(videos, Video{Md5: i, Name: name, Path: v[28:], CreatedAt: time.Unix(getFileCreateTime(v), 0)})
	}
	if len(videos) > 0 {
		DB.Create(&videos)
	}
	handleVideosChange(del, add)
	for i := range add {
		if _, err := os.Stat(PATH + "imgs/" + i + "thumb.jpg"); err == nil {
			thumbList = append(thumbList, i)
		}
		if _, err := os.Stat(PATH + "imgs/" + i + "process.jpg"); err == nil {
			processList = append(processList, i)
		}
	}
	DB.Model(&Video{}).Where("Md5 IN ?", thumbList).Select("images").Update("images", true)
	DB.Model(&Video{}).Where("Md5 IN ?", processList).Select("process_image").Update("process_image", true)
}

func UpdateAudios(del, add map[string]string) {
	audios := []Audio{}
	var delList []string
	for i := range del {
		delList = append(delList, i)
	}
	DB.Where("Md5 IN ?", delList).Delete(Audio{})
	for i, v := range add {
		name := regexp.MustCompile(`[^/\\\\]+$`).FindStringSubmatch(v)[0]
		audios = append(audios, Audio{Md5: i, Name: name, Path: v[28:], CreatedAt: time.Unix(getFileCreateTime(v), 0)})
	}
	if len(audios) > 0 {
		DB.Create(&audios)
	}
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

func getFileCreateTime(path string) int64 {
	osType := runtime.GOOS
	fileInfo, _ := os.Stat(path)
	if osType == "windows" {
		wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
		tNanSeconds := wFileSys.CreationTime.Nanoseconds() /// 返回的是纳秒
		tSec := tNanSeconds / 1e9                          ///秒
		return tSec
	}
	return time.Now().Unix()
}

func handleVideosChange(del, add map[string]string) {
	var wg sync.WaitGroup
	var maxGoroutines int = 4
	ch := make(chan int, maxGoroutines)
	for i, v := range add {
		wg.Add(1)
		ch <- 1
		go func(v, i string) {
			dll.Deal(v, PATH+`imgs\`+i)
			//dealVideo(v,i)
			<-ch
			wg.Done()
		}(v, i)
	}
	wg.Wait()
	for i := range del {
		os.Remove(PATH + `imgs\` + i + "thumb.jpg")
		os.Remove(PATH + `imgs\` + i + "process.jpg")
	}
}
