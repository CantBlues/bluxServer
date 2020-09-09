package main

import (
	"blux/db"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"
	_ "net/http/pprof"
)

// PATH is video file directory
const PATH string = db.PATH

func main() {

	startServe()

	//debug(test)
}

func test() {
	if isFilesChanged(PATH) {
		delTargets, addTargets := compareFilesMd5()
		db.UpdateVideos(delTargets, addTargets)
	}

}

func debug(f func()) {
	startTime := time.Now().UnixNano() / 1e6
	// performance record start
	f()
	// performance record end
	endTime := time.Now().UnixNano() / 1e6
	fmt.Printf("start time: %v;\n", startTime)
	fmt.Printf("end time: %v;\n", endTime)
	fmt.Printf("spend time: %v;\n", endTime-startTime)
}

func startServe() {
	http.HandleFunc("/", GetVideoFile)
	http.HandleFunc("/update", UpdateFiles)
	http.HandleFunc("/shutdown", ShutDown)
	http.ListenAndServe("192.168.0.174:9999", nil)
}

type result struct {
	data   string
	status bool
}

// GetVideoFile is get video rows from flutter
func GetVideoFile(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	ret := db.Fetch(r, body)
	fmt.Fprintf(w, ret)
}

// UpdateFiles is refresh files info from flutter app
func UpdateFiles(w http.ResponseWriter, r *http.Request) {
	if isFilesChanged(PATH) {
		delTargets, addTargets := compareFilesMd5()
		db.UpdateVideos(delTargets, addTargets)
		fmt.Fprintf(w, "1")
	} else {
		fmt.Fprintf(w, "0")
	}
}

// ShutDown means
func ShutDown(w http.ResponseWriter, r *http.Request) {
	exec.Command("cmd", "/C", "shutdown -s -t 0").Run()
	fmt.Fprintf(w, db.ShutDownResponse())
}

// MD5File is a function to divide the file and generate its md5 value
func MD5File(fileName string) (string, error) {
	var result []byte
	file, err := os.Open(fileName)
	if err != nil {
		return "error", err
	}
	defer file.Close()
	buffer := make([]byte, 256)
	for i := 0; i < 30; i++ {
		if i%3 == 0 {
			ret, _ := file.Seek(0, 1)
			file.Seek(ret*5, 0)
		}
		bytesread, err := file.Read(buffer)
		result = append(result, buffer[:bytesread]...)
		if err == io.EOF || bytesread == 0 {
			//fmt.Println("break!!!")
			break
		}
	}
	hash := md5.New()
	hash.Write(result)
	hash.Write([]byte(fileName))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getVideoFiles(vpath string) (int, []string) {
	rd, _ := ioutil.ReadDir(vpath)
	videoFiles := []string{}
	suffixList := []interface{}{".mp4", ".mkv", ".avi", ".MPG", ".rm", ".rmvb"}
	count := 0
	for _, file := range rd {
		if file.IsDir() {
			fileCount, filesList := getVideoFiles(vpath + file.Name() + "/")
			videoFiles = append(videoFiles, filesList...)
			count += fileCount
		} else if file.Size() > 5*1024*1024 {
			suffix := path.Ext(file.Name())
			if ok := inArray(suffix, suffixList); ok {
				videoFiles = append(videoFiles, vpath+file.Name())
				count++
			}
		}
	}
	return count, videoFiles
}

func isFilesChanged(path string) bool {
	oldMd5, _ := ioutil.ReadFile("./md5")
	_, files := getVideoFiles(path)
	filesMd5 := md5.New()
	for _, file := range files {
		filesMd5.Write([]byte(file))
	}
	newMd5 := filesMd5.Sum(nil)
	if string(oldMd5) == string(newMd5) {
		return false
	}
	ioutil.WriteFile("./md5", newMd5, 0666)
	return true
}

func compareFilesMd5() (del, add map[string]string) {
	var videos []db.Video
	_, files := getVideoFiles(PATH)
	filesMap, rowsMap := map[string]string{}, map[string]string{}
	for _, file := range files {
		md5, _ := MD5File(file)
		filesMap[md5] = file
	}
	database, _ := db.DB()
	database.Select("Md5").Find(&videos)
	for _, video := range videos {
		rowsMap[video.Md5] = ""
	}
	for i := range rowsMap {
		_, exist := filesMap[i]
		if exist {
			delete(filesMap, i)
			delete(rowsMap, i)
		}
	}
	return rowsMap, filesMap
	// sort.Strings(md5List)
	// sort.Strings(md5FromDb)
	// for i, v := range md5FromDb {
	// 	for index, value := range md5List {
	// 		if v == value {
	// 			md5List = append(md5List[:index], md5List[index+1:]...)
	// 			md5FromDb = append(md5FromDb[:i], md5FromDb[i+1:]...)
	// 		}
	// 	}
	// }
}

func inArray(need interface{}, needArr []interface{}) bool {
	for _, v := range needArr {
		if need == v {
			return true
		}
	}
	return false
}
