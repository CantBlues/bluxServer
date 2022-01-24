package usage

import (
	"blux/db"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	json.Unmarshal(body, &req)
	db.UsageReceiveData(req.Last, req.Usage, req.Apps)
	fmt.Fprintf(w, "saved")

}

func GetData(w http.ResponseWriter, r *http.Request) {
	type ret struct {
		Usage []db.PhoneUsage
		Apps  []db.PhoneApp
	}
	apps, usages := db.UsageGetAll()
	data, _ := json.Marshal(ret{Apps: apps, Usage: usages})
	w.Header().Set("Access-Control-Allow-Origin","*")
	fmt.Fprint(w, string(data))
}
