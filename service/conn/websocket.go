package conn

import (
	"github.com/gorilla/websocket"
	"github.com/gridsx/micro-conf/service/app"
	"github.com/kataras/iris/v12"

	"log"
	"net/http"
	"sync"
	"time"
)

var (
	upgrade = websocket.Upgrader{
		HandshakeTimeout:  time.Second * 30,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: false,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

//	handles websocket requests from the peer.
//
// 如果重复发送，那么得主动断开以前的连接，否则连接会过多
func serveWs(ctx iris.Context) {
	key := ctx.URLParam("key")
	existClient := GetClient(key)
	if existClient != nil {
		existClient.TryClose()
	}
	conn, err := upgrade.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256), key: key, lock: sync.Mutex{}}
	registerClient(client)
	go client.Read()
	go client.Write()
}

func RouteWs(a *iris.Application) {
	a.Get("/api/ws", app.RequireToken, serveWs) // 需要token
}
