package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logc"
)

type Request struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Response struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Conn struct {
	sync.RWMutex
	UID             int32
	Conn            *websocket.Conn
	Send            chan Response
	OnClose         func(*Conn)          // 关闭回调函数
	OnMessage       func(*Conn, Request) // 消息回调函数
	IsAlive         bool                 // 连接是否活跃
	LastActive      time.Time            // 最后活跃时间
	heartbeatTicker *time.Ticker         // 心跳定时器
	closeChan       chan struct{}        // 关闭信号通道
	writeMutex      sync.Mutex           // 写操作互斥锁
}

func NewConn(w http.ResponseWriter, r *http.Request, uid int32, heartbeat bool) (*Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	c := &Conn{
		UID:        uid,
		Conn:       conn,
		Send:       make(chan Response, 1024),
		IsAlive:    true,
		LastActive: time.Now(),
		closeChan:  make(chan struct{}),
	}

	go c.readPump()  // 处理读
	go c.writePump() // 处理写
	if heartbeat {   // 设置心跳检测
		c.heartbeatTicker = time.NewTicker(30 * time.Second)
		go c.heartbeat()
	}

	logc.Info(context.Background(), "WebSocket连接已创建: UID=%d", uid)
	return c, nil
}

// 处理读循环
func (c *Conn) readPump() {
	defer func() {
		if r := recover(); r != nil {
			logc.Errorf(context.Background(), "WebSocket读循环异常: UID=%d, 错误=%v", c.UID, r)
		}
		c.close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logc.Errorf(context.Background(), "WebSocket连接异常关闭: UID=%d, 错误=%v", c.UID, err)
			}
			break
		}

		// 更新最后活跃时间
		c.Lock()
		c.LastActive = time.Now()
		c.Unlock()

		// 重置读超时
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// 处理心跳消息
		if messageType == websocket.PingMessage {
			c.safeWriteMessage(websocket.PongMessage, nil)
			continue
		}

		// 处理文本消息
		if messageType == websocket.TextMessage {
			req := Request{}
			if err := json.Unmarshal(p, &req); err == nil {
				if c.OnMessage != nil {
					c.OnMessage(c, req)
				}
			} else {
				logc.Errorf(context.Background(), "消息解析失败: UID=%d, 错误=%v", c.UID, err)
			}
		}
	}
}

// 处理写循环
func (c *Conn) writePump() {
	defer func() {
		if r := recover(); r != nil {
			logc.Errorf(context.Background(), "WebSocket写循环异常: UID=%d, 错误=%v", c.UID, r)
		}
		c.close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				// 通道已关闭
				c.safeWriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := c.safeWriteJSON(msg)
			if err != nil {
				logc.Errorf(context.Background(), "消息发送失败: UID=%d, 错误=%v", c.UID, err)
				return
			}

		case <-c.closeChan:
			return
		}
	}
}

// 心跳检测
func (c *Conn) heartbeat() {
	defer c.heartbeatTicker.Stop()

	for {
		select {
		case <-c.heartbeatTicker.C:
			// 检查连接是否超时
			c.RLock()
			lastActive := c.LastActive
			c.RUnlock()

			if time.Since(lastActive) > 60*time.Second {
				logc.Errorf(context.Background(), "WebSocket连接超时: UID=%d", c.UID)
				c.close()
				return
			}

			// 发送心跳ping
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.safeWriteMessage(websocket.PingMessage, nil); err != nil {
				logc.Errorf(context.Background(), "心跳发送失败: UID=%d, 错误=%v", c.UID, err)
				c.close()
				return
			}

		case <-c.closeChan:
			return
		}
	}
}

// 安全写入消息（带锁保护）
func (c *Conn) safeWriteMessage(messageType int, data []byte) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if !c.IsConnected() {
		return websocket.ErrCloseSent
	}

	return c.Conn.WriteMessage(messageType, data)
}

// 安全写入JSON（带锁保护）
func (c *Conn) safeWriteJSON(v interface{}) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if !c.IsConnected() {
		return websocket.ErrCloseSent
	}

	return c.Conn.WriteJSON(v)
}

// 安全关闭连接
func (c *Conn) close() {
	c.Lock()
	defer c.Unlock()

	if !c.IsAlive {
		return
	}

	c.IsAlive = false
	close(c.closeChan)

	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
	}

	c.Conn.Close()
	close(c.Send)

	if c.OnClose != nil {
		c.OnClose(c)
	}

	logc.Errorf(context.Background(), "WebSocket连接已关闭: UID=%d", c.UID)
}

// Close 主动关闭连接
func (c *Conn) Close() {
	c.close()
}

// IsConnected 检查连接是否活跃
func (c *Conn) IsConnected() bool {
	c.RLock()
	defer c.RUnlock()
	return c.IsAlive
}

// GetLastActive 获取最后活跃时间
func (c *Conn) GetLastActive() time.Time {
	c.RLock()
	defer c.RUnlock()
	return c.LastActive
}

// SafeSend 安全发送消息，避免通道阻塞
func (c *Conn) SafeSend(msg Response) bool {
	if !c.IsConnected() {
		return false
	}

	select {
	case c.Send <- msg:
		return true
	case <-time.After(5 * time.Second):
		logc.Errorf(context.Background(), "WebSocket消息发送超时: UID=%d", c.UID)
		return false
	default:
		logc.Errorf(context.Background(), "WebSocket消息通道已满: UID=%d", c.UID)
		return false
	}
}

// DirectSend 直接发送消息（绕过通道，用于紧急消息）
func (c *Conn) DirectSend(msg Response) error {
	if !c.IsConnected() {
		return websocket.ErrCloseSent
	}

	c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.safeWriteJSON(msg)
}
