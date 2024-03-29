package main

import (
	"blux/db"
	"blux/video"
	"blux/ws"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"time"

	"github.com/CantBlues/v2sub/ping"
	"github.com/CantBlues/v2sub/types"
	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	LogPath    string `yaml:"log_path"`
	RouterAddr string `yaml:"router"`
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

	http.HandleFunc("/v2ray/detect", detectV2ray)
	http.HandleFunc("/v2ray/detect/nodes", detectV2rayNodes)

	http.HandleFunc("/ws", ws.WsEndpoint)
	http.ListenAndServe(":9999", nil)
}

func checkOnline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "online")
}

func getVideos(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	ret := db.Fetch(r, body)
	fmt.Fprint(w, ret)
}

func getAudios(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	ret := db.FetchAudio(r, body)
	fmt.Fprint(w, ret)
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
	fmt.Fprint(w, db.ShutDownResponse())
}

func detectV2ray(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get(Config.RouterAddr + "fetch")
	if err != nil {
		log.Println("Failed to fetch config from router", err)
		return
	}
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)
	var data types.Config
	err = json.Unmarshal(b, &data)

	if err != nil {
		log.Println("Failed to Unmarshall json data", err)
		return
	}
	go func() {
		ping.TestAll(data.Nodes)
		buff := bytes.NewBuffer(nil)
		encoder := json.NewEncoder(buff)

		err := encoder.Encode(data.Nodes)
		if err != nil {
			log.Println("Failed to encode json data", err)
			return
		}

		go http.Post("http://blux.lanbin.com/api/v2ray/nodes/save", "application/json", buff)
		http.Post(Config.RouterAddr+"nodes/receive", "application/json", buff)
	}()

	w.Write([]byte{'1'})
}

func detectV2rayNodes(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Source string      `json:"source"`
		Nodes  types.Nodes `json:"nodes"`
	}

	var request req
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		log.Println("Failed to Unmarshall json data", err)
		return
	}
	defer r.Body.Close()

	source := request.Source
	nodes := request.Nodes

	go func(s string, n types.Nodes) {
		ping.TestAll(n)
		buff := bytes.NewBuffer(nil)
		encoder := json.NewEncoder(buff)

		err := encoder.Encode(n)
		if err != nil {
			log.Println("Failed to encode json data", err)
			return
		}
		b, _ := ioutil.ReadAll(buff)
		data1 := bytes.NewBuffer(b)
		data2 := bytes.NewBuffer(b)

		switch s {
		case "instant":
			http.Post(Config.RouterAddr+"nodes/receive", "application/json", data1)
		case "mark":
			http.Post(Config.RouterAddr+"nodes/receiveMark", "application/json", data1)
		case "history":
			break
		default:
			break
		}
		http.Post("http://blux.lanbin.com/api/v2ray/nodes/save", "application/json", data2)
	}(source, nodes)

	w.Write([]byte{'1'})
}
