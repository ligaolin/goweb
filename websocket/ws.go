package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
)

type WebSocket struct {
	sync.RWMutex
	Groups        map[int32][]*Conn
	Error         error
	Stats         *WebSocketStats
	HeartbeatTime time.Duration
}

type WebSocketStats struct {
	sync.RWMutex
	TotalConnections  int64
	ActiveConnections int64
	MessagesSent      int64
	MessagesReceived  int64
}

var (
	wss  *WebSocket
	once sync.Once
)

func GetSocket() *WebSocket {
	once.Do(func() {
		wss = &WebSocket{
			Groups:        make(map[int32][]*Conn),
			Stats:         &WebSocketStats{},
			HeartbeatTime: 30 * time.Second,
		}
	})
	return wss
}

// Add 添加新的WebSocket连接
func (ws *WebSocket) Add(w http.ResponseWriter, r *http.Request, uid int32, heartbeat bool, handler func(*Conn, Request) (Response, error)) (*Conn, error) {
	ws.Lock()
	defer ws.Unlock()

	conn, err := NewConn(w, r, uid, heartbeat)
	if err != nil {
		return nil, err
	}

	// 设置连接关闭回调
	conn.OnClose = func(c *Conn) {
		ws.Remove(uid, c)
	}

	// 设置消息处理回调
	conn.OnMessage = func(c *Conn, req Request) {
		ws.handleMessage(c, req, handler)
	}

	// 添加到用户组
	ws.Groups[uid] = append(ws.Groups[uid], conn)

	// 更新统计信息
	ws.Stats.Lock()
	ws.Stats.TotalConnections++
	ws.Stats.ActiveConnections++
	ws.Stats.Unlock()

	logc.Info(context.Background(), "WebSocket连接已添加: UID=%d, 总连接数=%d", uid, len(ws.Groups[uid]))
	return conn, nil
}

// SetMessage 向指定用户发送消息
func (ws *WebSocket) SetMessage(uid int32, resp Response) error {
	ws.RLock()
	connections, exists := ws.Groups[uid]
	ws.RUnlock()

	if !exists || len(connections) == 0 {
		logc.Info(context.Background(), "用户%d没有活跃的连接", uid)
		return nil
	}

	successCount := 0

	for _, conn := range connections {
		if conn.SafeSend(resp) {
			successCount++
			ws.Stats.Lock()
			ws.Stats.MessagesSent++
			ws.Stats.Unlock()
		}
	}

	logc.Info(context.Background(), "消息已发送给用户%d: 成功%d个连接", uid, successCount)
	return nil
}

// Broadcast 广播消息给所有用户
func (ws *WebSocket) Broadcast(resp Response) error {
	ws.RLock()
	groups := make(map[int32][]*Conn)
	for uid, connections := range ws.Groups {
		groups[uid] = make([]*Conn, len(connections))
		copy(groups[uid], connections)
	}
	ws.RUnlock()

	totalSent := 0
	totalConnections := 0

	for _, connections := range groups {
		totalConnections += len(connections)
		for _, conn := range connections {
			if conn.SafeSend(resp) {
				totalSent++
				ws.Stats.Lock()
				ws.Stats.MessagesSent++
				ws.Stats.Unlock()
			}
		}
	}

	logc.Info(context.Background(), "广播消息已发送: 总数%d/%d", totalSent, totalConnections)
	return nil
}

// Remove 移除关闭的连接
func (ws *WebSocket) Remove(uid int32, conn *Conn) error {
	ws.Lock()
	defer ws.Unlock()

	connections, exists := ws.Groups[uid]
	if !exists {
		return nil
	}

	for i, c := range connections {
		if c == conn {
			// 安全删除连接
			ws.Groups[uid] = append(connections[:i], connections[i+1:]...)

			// 清理空组
			if len(ws.Groups[uid]) == 0 {
				delete(ws.Groups, uid)
			}

			// 更新统计信息
			ws.Stats.Lock()
			ws.Stats.ActiveConnections--
			ws.Stats.Unlock()

			logc.Info(context.Background(), "WebSocket连接已移除: UID=%d, 剩余连接数=%d", uid, len(ws.Groups[uid]))
			return nil
		}
	}

	return nil
}

// GetStats 获取WebSocket统计信息
func (ws *WebSocket) GetStats() WebSocketStats {
	ws.Stats.RLock()
	defer ws.Stats.RUnlock()

	// 返回统计信息的副本，避免返回包含锁的结构体
	return WebSocketStats{
		TotalConnections:  ws.Stats.TotalConnections,
		ActiveConnections: ws.Stats.ActiveConnections,
		MessagesSent:      ws.Stats.MessagesSent,
		MessagesReceived:  ws.Stats.MessagesReceived,
	}
}

// GetUserConnections 获取指定用户的连接数
func (ws *WebSocket) GetUserConnections(uid int32) int {
	ws.RLock()
	defer ws.RUnlock()

	if connections, exists := ws.Groups[uid]; exists {
		return len(connections)
	}
	return 0
}

// GetAllConnections 获取所有活跃连接数
func (ws *WebSocket) GetAllConnections() int {
	ws.RLock()
	defer ws.RUnlock()

	total := 0
	for _, connections := range ws.Groups {
		total += len(connections)
	}
	return total
}

// GetActiveUsers 获取活跃用户数
func (ws *WebSocket) GetActiveUsers() int {
	ws.RLock()
	defer ws.RUnlock()

	return len(ws.Groups)
}

// handleMessage 处理接收到的消息
func (ws *WebSocket) handleMessage(c *Conn, req Request, handler func(*Conn, Request) (Response, error)) {
	ws.Stats.Lock()
	ws.Stats.MessagesReceived++
	ws.Stats.Unlock()

	res, err := handler(c, req)
	if err != nil {
		res.Type = "error"
		res.Data = err.Error()
	}
	// 发送响应消息，使用安全的写入方法
	if err := c.safeWriteJSON(res); err != nil {
		logc.Errorf(context.Background(), "发送websocket消息错误: %v\n", err)
	}
}

// Cleanup 清理所有连接
func (ws *WebSocket) Cleanup() {
	ws.Lock()
	defer ws.Unlock()

	for uid, connections := range ws.Groups {
		for _, conn := range connections {
			conn.Close()
		}
		delete(ws.Groups, uid)
	}

	ws.Stats.Lock()
	ws.Stats.ActiveConnections = 0
	ws.Stats.Unlock()

	logc.Info(context.Background(), "WebSocket连接已全部清理")
}

type SendError struct {
	UID     int32
	Conn    *Conn
	Message string
}

func (e *SendError) Error() string {
	return fmt.Sprintf("SendError[UID=%d]: %s", e.UID, e.Message)
}
