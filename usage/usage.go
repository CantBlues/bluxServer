package usage

import (
	"blux/db"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func GetLastDate(w http.ResponseWriter, r *http.Request) {
	last := db.UsageLastDate()
	fmt.Fprintf(w, "%d", last)
}

func ReceiveData(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Last  int
		Usage []db.PhoneUsage
		Apps  []db.PhoneApp
	}
	body, _ := ioutil.ReadAll(r.Body)
	req := new(request)


	startTime := time.Now().UnixNano() / 1e6

	
	json.Unmarshal(body, &req)
	db.UsageReceiveData(req.Last, req.Usage, req.Apps)
	fmt.Fprintf(w, "saved")


	endTime := time.Now().UnixNano() / 1e6
	fmt.Printf("start time: %v;\n", startTime)
	fmt.Printf("end time: %v;\n", endTime)
	fmt.Printf("spend time: %v;\n", endTime-startTime)
	fmt.Println(len(body))
}
