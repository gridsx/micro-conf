package conn

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/winjeg/go-commons/log"
)

// 重启的时候client端连接很可能分布不均匀不会重新来新启动的这台，因此websocket是不是也应该设置个时长
// 超过指定时长的连接，自动断开，让客户端重连，这样可以把流量均匀分配一下

var logger = log.GetLogger(nil)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = pongWait / 3

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	key    string
	lock   sync.Mutex
	closed bool
}

func (c *Client) Send(msg string) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Warnf("websocket client send  err: %v\n", err)
		}
	}()
	c.send <- []byte(msg)
}

func (c *Client) Read() {
	defer c.Close()
	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		logger.Warnf("websocket client read set deadline err: %s\n", err.Error())
		return
	}
	c.conn.SetPongHandler(func(s string) error {
		logger.Debugf("websocket read received ping msg: %s\n", s)
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			logger.Warnf("websocket client read ping msg err: %s\n", err.Error())
			return err
		}
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			logger.Warnf("websocket read message error: %s\n", err.Error())
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Debugf("websocket read message unexpected error: %s\n", err.Error())
			}
			break
		}
		info := new(HeartBeat)
		jsonErr := json.Unmarshal(message, info)
		if jsonErr != nil {
			logger.Warningln("websocket failed to unmarshal message: " + string(message))
		} else {
			routeHeartbeat(info)
		}
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() { ticker.Stop(); c.Close() }()
	for {
		select {
		case message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}
			if !ok {
				// The hub closed the channel.
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					logger.Warnf("websocket write close message error: %s\n", err.Error())
					return
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Warnf("websocket write message next writer err: %s\n", err.Error())
				return
			}

			_, writeErr := w.Write(message)
			if writeErr != nil {
				logger.Warnf("websocket write message error: %s\n", err.Error())
				return
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err := w.Write(<-c.send)
				if err != nil {
					logger.Warnf("websocket write message error: %s\n", err.Error())
					return
				}
			}

			if err := w.Close(); err != nil {
				logger.Warnf("websocket write message close error: %s\n", err.Error())
				return
			}
		case <-ticker.C: // 保持心跳
			if deadlineErr := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); deadlineErr != nil {
				logger.Warnf("websocket write ping write deadline error: %s\n", deadlineErr.Error())
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				logger.Warnf("websocket write ping write message error: %s\n", err.Error())
				return
			}
		}
	}

}

func (c *Client) Close() {
	defer func() {
		err := recover()
		if err != nil {
			logger.Warnf("websocket connection close err: %v\n", err)
		}
	}()
	c.lock.Lock()
	if c.closed {
		c.lock.Unlock()
		return
	}
	c.closed = true
	c.lock.Unlock()
	close(c.send)
	unregisterClient(c)
	err := c.conn.Close()
	if err != nil {
		logger.Warnf("websocket connection close err: %s\n", err.Error())
		return
	}
	logger.Infof("websocket client %s closed!\n", c.key)
}

func (c *Client) TryClose() {
	err := c.conn.WriteMessage(websocket.CloseMessage, []byte("tryClose connection closing old..."))
	if err != nil {
		logger.Warnf("TryClose - send close message err: %s\n", err.Error())
		return
	}
}
