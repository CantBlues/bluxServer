package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"net/http"
	"testing"

	"github.com/CantBlues/v2sub/types"
)

func TestPost(t *testing.T) {
	println("test post")

	var data = &types.Nodes{{Port: 10, Name: "test"}}
	buff := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buff)
	err := encoder.Encode(data)
	if err != nil {
		println("Failed to encode json data", err)
		return
	}

	b, _ := ioutil.ReadAll(buff)
	data1 := bytes.NewBuffer(b)
	data2 := bytes.NewBuffer(b)

	http.Post("http://blux.lanbin.com/api/v2ray/nodes/save", "application/json", data1)

	http.Post("http://blux.lanbin.com/api/v2ray/nodes/save", "application/json", data2)

}
