package main

import (
	"blux/video"
	"blux/db"
	"blux/usage"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"
	_ "net/http/pprof"

)

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
	http.HandleFunc("/usage/getLastDate", usage.GetLastDate)
	http.HandleFunc("/usage/sendData", usage.ReceiveData)
	http.HandleFunc("/usage/getData", usage.GetData)
	http.ListenAndServe(":9999", nil)
}


func checkOnline(w http.ResponseWriter, r *http.Request)  {
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

