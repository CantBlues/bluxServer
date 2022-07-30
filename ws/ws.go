package ws

import (
	"encoding/json"
	"log"
	"github.com/gorilla/websocket"
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

	err = ws.WriteMessage(1, []byte("Hi Client!"))
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
	// msg := replyMsg.Uid + "说:" + replyMsg.Msg
	msg := replyMsg.Msg
	if connUid == replyMsg.Uid {

		if replyMsg.Type == "login" {
			msg = "连接成功"
		} else {
			return
			// todo here can be verify client message
			// msg = "你说：" + replyMsg.Msg
		}

	}

	if err := conn.WriteMessage(1, []byte(msg)); err != nil {
		log.Println("Can't send", err)
	}

}
