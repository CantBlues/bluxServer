package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type MsgConfig struct {
	Type string `json:"type,omitempty"`
	Uid  string `json:"uid,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

var connMap = make(map[string]*websocket.Conn)

func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	reader(ws)
}

func reader(conn *websocket.Conn) {
	replyMsg := MsgConfig{}
	for {
		// read in a message
		_, p, err := conn.ReadMessage()

		if err != nil {
			// if _, k:= err.(*websocket.CloseError);k {
			delete(connMap, replyMsg.Uid)
			// }
			log.Println(err)
			return
		}

		err = json.Unmarshal(p, &replyMsg)
		if err != nil {
			log.Println("json decode error", err)
		}

		if replyMsg.Type == "login" && replyMsg.Uid != "" {
			connMap[replyMsg.Uid] = conn
			log.Println(connMap)
		}
		for k, v := range connMap {
			go sendMessage(replyMsg, v, k)
		}
		// print out that message for clarity
		// fmt.Println(string(p))

		// if err := conn.WriteMessage(messageType, p); err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

	}
}

func sendMessage(replyMsg MsgConfig, conn *websocket.Conn, connUid string) {

	if connUid == replyMsg.Uid {
		if replyMsg.Type == "login" {
			replyMsg.Msg = "连接成功"
			replyMsg.Type = "msg"
		} else {
			return
			// todo here can be verify client message
		}
	}
	msg, err := json.Marshal(replyMsg)
	if err != nil {
		log.Println(err)
	}

	if err := conn.WriteMessage(1, []byte(msg)); err != nil {
		log.Println("Can't send", err)
	}

}
