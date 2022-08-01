package main

import (
	"blux/db"
	"blux/video"
	"blux/ws"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"time"
)

type YamlConfig struct {
	LogPath string `yaml:"log_path"`
}

var Config YamlConfig

func init() {
	loadConfig()
	logFile, err := os.OpenFile(Config.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Panic("打开日志文件异常")
	}
	log.SetOutput(logFile)
}

func loadConfig() {
	data, err := ioutil.ReadFile("./conf.yaml")
	if err != nil {
		log.Panic(err)
	}
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Panic(err)
	}
}

// PATH is video file directory
const PATH string = db.PATH

func main() {

	startServe()

	// debug(test)

}

func test() {

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
	http.HandleFunc("/", getVideos)
	http.HandleFunc("/update", updateVideos)
	http.HandleFunc("/shutdown", shutDown)
	http.HandleFunc("/checkonline", checkOnline)
	http.HandleFunc("/getaudios", getAudios)
	http.HandleFunc("/updateaudios", updateAudios)

	http.HandleFunc("/ws", ws.WsEndpoint)
	http.ListenAndServe(":9999", nil)
}

func checkOnline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "online")
}

func getVideos(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	ret := db.Fetch(r, body)
	fmt.Fprintf(w, ret)
}

func getAudios(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	ret := db.FetchAudio(r, body)
	fmt.Fprintf(w, ret)
}

func updateVideos(w http.ResponseWriter, r *http.Request) {
	if video.VideosChanged(PATH) {
		delTargets, addTargets := video.CompareVideosMd5()
		go db.UpdateVideos(delTargets, addTargets)
		fmt.Fprintf(w, "1")
	} else {
		fmt.Fprintf(w, "0")
	}
}

func updateAudios(w http.ResponseWriter, r *http.Request) {
	if video.AudiosChanged(PATH) {
		delTargets, addTargets := video.CompareAudiosMd5()
		db.UpdateAudios(delTargets, addTargets)
		fmt.Fprintf(w, "1")
	} else {
		fmt.Fprintf(w, "0")
	}
}

func shutDown(w http.ResponseWriter, r *http.Request) {
	exec.Command("cmd", "/C", "shutdown -s -hybrid -t 0").Run()
	fmt.Fprintf(w, db.ShutDownResponse())
}
