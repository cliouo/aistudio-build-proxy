package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// --- Constants ---
const (
	wsPath              = "/v1/ws"
	proxyListenAddr     = ":5345"
	wsReadTimeout       = 60 * time.Second
	proxyRequestTimeout = 600 * time.Second
)

// --- 1. 连接管理与负载均衡 ---

// UserConnection 存储单个WebSocket连接及其元数据
type UserConnection struct {
	Conn       *websocket.Conn
	UserID     string
	LastActive time.Time
	writeMutex sync.Mutex // 保护对此单个连接的并发写入
}

// safeWriteJSON 线程安全地向单个WebSocket连接写入JSON
func (uc *UserConnection) safeWriteJSON(v interface{}) error {
	uc.writeMutex.Lock()
	defer uc.writeMutex.Unlock()
	return uc.Conn.WriteJSON(v)
}

// UserConnections 维护单个用户的所有连接和负载均衡状态
type UserConnections struct {
	sync.Mutex
	Connections []*UserConnection
	NextIndex   int // 用于轮询 (round-robin)
}

// ConnectionPool 全局连接池，并发安全
type ConnectionPool struct {
	sync.RWMutex
	Users map[string]*UserConnections
}

var globalPool = &ConnectionPool{
	Users: make(map[string]*UserConnections),
}

// AddConnection 将新连接添加到池中
func (p *ConnectionPool) AddConnection(userID string, conn *websocket.Conn) *UserConnection {
	userConn := &UserConnection{
		Conn:       conn,
		UserID:     userID,
		LastActive: time.Now(),
	}

	p.Lock()
	defer p.Unlock()

	userConns, exists := p.Users[userID]
	if !exists {
		userConns = &UserConnections{
			Connections: make([]*UserConnection, 0),
			NextIndex:   0,
		}
		p.Users[userID] = userConns
	}

	userConns.Lock()
	userConns.Connections = append(userConns.Connections, userConn)
	userConns.Unlock()

	log.Printf("WebSocket connected: UserID=%s, Total connections for user: %d", userID, len(userConns.Connections))
	return userConn
}

// RemoveConnection 从池中移除连接
func (p *ConnectionPool) RemoveConnection(userID string, conn *websocket.Conn) {
	p.Lock()
	defer p.Unlock()

	userConns, exists := p.Users[userID]
	if !exists {
		return
	}

	userConns.Lock()
	defer userConns.Unlock()

	// 查找并移除连接
	for i, uc := range userConns.Connections {
		if uc.Conn == conn {
			// 高效删除：将最后一个元素移到当前位置，然后截断切片
			userConns.Connections[i] = userConns.Connections[len(userConns.Connections)-1]
			userConns.Connections = userConns.Connections[:len(userConns.Connections)-1]
			log.Printf("WebSocket disconnected: UserID=%s, Remaining connections for user: %d", userID, len(userConns.Connections))
			break
		}
	}

	// 如果该用户没有连接了，可以从主map中删除用户条目（可选）
	if len(userConns.Connections) == 0 {
		delete(p.Users, userID)
	}
}

// GetConnection 使用轮询策略为用户选择一个连接
func (p *ConnectionPool) GetConnection(userID string) (*UserConnection, error) {
	p.RLock()
	userConns, exists := p.Users[userID]
	p.RUnlock()

	if !exists {
		return nil, errors.New("no available client for this user")
	}

	userConns.Lock()
	defer userConns.Unlock()

	numConns := len(userConns.Connections)
	if numConns == 0 {
		// 理论上如果存在于p.Users中，这里不应该为0，但为了健壮性还是检查
		return nil, errors.New("no available client for this user")
	}

	// 轮询负载均衡
	idx := userConns.NextIndex % numConns
	selectedConn := userConns.Connections[idx]
	userConns.NextIndex = (userConns.NextIndex + 1) % numConns // 更新索引

	return selectedConn, nil
}

// --- 2. WebSocket 消息结构 & 待处理请求 ---

// WSMessage 是前后端之间通信的基本结构
type WSMessage struct {
	ID      string                 `json:"id"`      // 请求/响应的唯一ID
	Type    string                 `json:"type"`    // ping, pong, http_request, http_response, stream_start, stream_chunk, stream_end, error
	Payload map[string]interface{} `json:"payload"` // 具体数据
}

// pendingRequests 存储待处理的HTTP请求，等待WS响应
// key: reqID (string), value: chan *WSMessage
var pendingRequests sync.Map

// --- 3. WebSocket 处理器和心跳 ---

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 生产环境中应设置严格的CheckOrigin
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 认证
	authToken := r.URL.Query().Get("auth_token")
	userID, err := validateJWT(authToken)
	if err != nil {
		log.Printf("WebSocket authentication failed: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 升级连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	// 添加到连接池
	userConn := globalPool.AddConnection(userID, conn)

	// 启动读取循环
	go readPump(userConn)
}

// readPump 处理来自单个WebSocket连接的所有传入消息
func readPump(uc *UserConnection) {
	defer func() {
		globalPool.RemoveConnection(uc.UserID, uc.Conn)
		_ = uc.Conn.Close()
		log.Printf("readPump closed for user %s", uc.UserID)
	}()

	// 设置读取超时 (心跳机制)
	_ = uc.Conn.SetReadDeadline(time.Now().Add(wsReadTimeout))

	for {
		_, message, err := uc.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error for user %s: %v", uc.UserID, err)
			} else {
				log.Printf("WebSocket closed for user %s: %v", uc.UserID, err)
			}
			// 如果读取失败（包括超时），退出循环并清理连接
			break
		}

		// 收到任何消息，重置读取超时
		_ = uc.Conn.SetReadDeadline(time.Now().Add(wsReadTimeout))
		uc.LastActive = time.Now()

		// 解析消息
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshalling WebSocket message: %v", err)
			continue
		}

		switch msg.Type {
		case "ping":
			// 心跳响应
			err := uc.safeWriteJSON(map[string]string{"type": "pong", "id": msg.ID})
			if err != nil {
				log.Printf("Error sending pong: %v", err)
				return // 发送失败，认为连接已断
			}
		case "http_response", "stream_start", "stream_chunk", "stream_end", "error":
			// 路由响应到等待的HTTP Handler
			if ch, ok := pendingRequests.Load(msg.ID); ok {
				respChan := ch.(chan *WSMessage)
				// 尝试发送，如果通道已满（不太可能，但为了安全），则记录日志
				select {
				case respChan <- &msg:
				default:
					log.Printf("Warning: Response channel full for request ID %s, dropping message type %s", msg.ID, msg.Type)
				}
			} else {
				log.Printf("Received response for unknown or timed-out request ID: %s", msg.ID)
			}
		default:
			log.Printf("Received unknown message type from client: %s", msg.Type)
		}
	}
}

// --- 4. HTTP 反向代理与 WS 隧道 ---

func HandleProxyRequest(w http.ResponseWriter, r *http.Request) {
	// 1. 认证并获取UserID (这里模拟)
	userID, err := authenticateHTTPRequest(r)
	if err != nil {
		http.Error(w, "Proxy authentication failed", http.StatusUnauthorized)
		return
	}

	// 2. 生成唯一请求ID
	reqID := uuid.NewString()

	// 3. 创建响应通道并注册
	// 使用带缓冲的通道以适应流式响应块
	respChan := make(chan *WSMessage, 10)
	pendingRequests.Store(reqID, respChan)
	defer pendingRequests.Delete(reqID) // 确保请求结束后清理

	// 4. 选择一个WebSocket连接
	selectedConn, err := globalPool.GetConnection(userID)
	if err != nil {
		log.Printf("Error getting connection for user %s: %v", userID, err)
		http.Error(w, "Service Unavailable: No active client connected", http.StatusServiceUnavailable)
		return
	}

	// 5. 封装HTTP请求为WS消息
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer func() { _ = r.Body.Close() }()

	// 注意：将Header直接序列化为JSON可能需要一些处理，这里简化处理
	// 对于生产环境，可能需要更精细的Header转换
	headers := make(map[string][]string)
	for k, v := range r.Header {
		// 过滤掉一些HTTP/1.1特有的或代理不应转发的头
		if k != "Connection" && k != "Keep-Alive" && k != "Proxy-Authenticate" && k != "Proxy-Authorization" && k != "Te" && k != "Trailers" && k != "Transfer-Encoding" && k != "Upgrade" {
			headers[k] = v
		}
	}

	requestPayload := WSMessage{
		ID:   reqID,
		Type: "http_request",
		Payload: map[string]interface{}{
			"method": r.Method,
			// 假设前端知道如何处理这个相对URL，或者您在这里构建完整的外部URL
			"url":     "https://generativelanguage.googleapis.com" + r.URL.String(),
			"headers": headers,
			"body":    string(bodyBytes), // 对于二进制数据，应使用base64编码
		},
	}

	// 6. 发送请求到WebSocket客户端
	if err := selectedConn.safeWriteJSON(requestPayload); err != nil {
		log.Printf("Failed to send request over WebSocket: %v", err)
		http.Error(w, "Bad Gateway: Failed to send request to client", http.StatusBadGateway)
		return
	}

	// 7. 异步等待并处理响应
	processWebSocketResponse(w, r, respChan)
}

// processWebSocketResponse 处理来自WS通道的响应，构建HTTP响应
func processWebSocketResponse(w http.ResponseWriter, r *http.Request, respChan chan *WSMessage) {
	// 设置超时
	ctx, cancel := context.WithTimeout(r.Context(), proxyRequestTimeout)
	defer cancel()

	// 获取Flusher以支持流式响应
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Println("Warning: ResponseWriter does not support flushing, streaming will be buffered.")
	}

	headersSet := false

	for {
		select {
		case msg, ok := <-respChan:
			if !ok {
				// 通道被关闭，理论上不应该发生，除非有panic
				if !headersSet {
					http.Error(w, "Internal Server Error: Response channel closed unexpectedly", http.StatusInternalServerError)
				}
				return
			}

			switch msg.Type {
			case "http_response":
				// 标准单个响应
				if headersSet {
					log.Println("Received http_response after headers were already set. Ignoring.")
					return
				}
				setResponseHeaders(w, msg.Payload)
				writeStatusCode(w, msg.Payload)
				writeBody(w, msg.Payload)
				return // 请求结束

			case "stream_start":
				// 流开始
				if headersSet {
					log.Println("Received stream_start after headers were already set. Ignoring.")
					continue
				}
				setResponseHeaders(w, msg.Payload)
				writeStatusCode(w, msg.Payload)
				headersSet = true
				if flusher != nil {
					flusher.Flush()
				}

			case "stream_chunk":
				// 流数据块
				if !headersSet {
					// 如果还没收到stream_start，先设置默认头
					log.Println("Warning: Received stream_chunk before stream_start. Using default 200 OK.")
					w.WriteHeader(http.StatusOK)
					headersSet = true
				}
				writeBody(w, msg.Payload)
				if flusher != nil {
					flusher.Flush() // 立即将数据块发送给客户端
				}

			case "stream_end":
				// 流结束
				if !headersSet {
					// 如果流结束了但还没设置头，设置一个默认的
					w.WriteHeader(http.StatusOK)
				}
				return // 请求结束

			case "error":
				// 前端返回错误
				if !headersSet {
					errMsg := "Bad Gateway: Client reported an error"
					if payloadErr, ok := msg.Payload["error"].(string); ok {
						errMsg = payloadErr
					}
					statusCode := http.StatusBadGateway
					if code, ok := msg.Payload["status"].(float64); ok {
						statusCode = int(code)
					}
					http.Error(w, errMsg, statusCode)
				} else {
					// 如果已经开始发送流，我们只能记录错误并关闭连接
					log.Printf("Error received from client after stream started: %v", msg.Payload)
				}
				return // 请求结束

			default:
				log.Printf("Received unexpected message type %s while waiting for response", msg.Type)
			}

		case <-ctx.Done():
			// 超时
			if !headersSet {
				log.Printf("Gateway Timeout: No response from client for request %s", r.URL.Path)
				http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
			} else {
				// 如果流已经开始，我们只能记录日志并断开连接
				log.Printf("Gateway Timeout: Stream incomplete for request %s", r.URL.Path)
			}
			return
		}
	}
}

// --- 辅助函数 ---

// setResponseHeaders 从payload中解析并设置HTTP响应头
func setResponseHeaders(w http.ResponseWriter, payload map[string]interface{}) {
	headers, ok := payload["headers"].(map[string]interface{})
	if !ok {
		return
	}
	for key, value := range headers {
		// 假设值是 []interface{} 或 string
		if values, ok := value.([]interface{}); ok {
			for _, v := range values {
				if strV, ok := v.(string); ok {
					w.Header().Add(key, strV)
				}
			}
		} else if strV, ok := value.(string); ok {
			w.Header().Set(key, strV)
		}
	}
}

// writeStatusCode 从payload中解析并设置HTTP状态码
func writeStatusCode(w http.ResponseWriter, payload map[string]interface{}) {
	status, ok := payload["status"].(float64) // JSON数字默认为float64
	if !ok {
		w.WriteHeader(http.StatusOK) // 默认200
		return
	}
	w.WriteHeader(int(status))
}

// writeBody 从payload中解析并写入HTTP响应体
func writeBody(w http.ResponseWriter, payload map[string]interface{}) {
	var bodyData []byte
	// 对于 http_response，body 键通常包含数据
	if body, ok := payload["body"].(string); ok {
		bodyData = []byte(body)
	}
	// 对于 stream_chunk，data 键通常包含数据
	if data, ok := payload["data"].(string); ok {
		bodyData = []byte(data)
	}
	// 注意：如果前端发送的是二进制数据，这里应该假设它是base64编码的字符串并进行解码

	if len(bodyData) > 0 {
		_, _ = w.Write(bodyData)
	}
}

// validateJWT 模拟JWT验证并返回userID
func validateJWT(token string) (string, error) {
	if token == "" {
		return "", errors.New("missing auth_token")
	}
	// 实际应用中，这里需要使用JWT库（如golang-jwt/jwt）来验证签名和过期时间
	// 这里我们简单地将token当作userID
	if token == "valid-token-user-1" {
		return "user-1", nil
	}
	//if token == "valid-token-user-2" {
	//	return "user-2", nil
	//}
	return "", errors.New("invalid token")
}

// authenticateHTTPRequest 模拟HTTP代理请求的认证
func authenticateHTTPRequest(r *http.Request) (string, error) {
	// 实际应用中，可能检查Authorization头或其他API Key
	//apiKey := r.Header.Get("x-goog-api-key")
	//if apiKey == "secret-key-for-user-1" {
	//	return "user-1", nil
	//}
	//if apiKey == "secret-key-for-user-2" {
	//	return "user-2", nil
	//}
	//// 也可以从请求路径中获取 /proxy/user-1/...
	//return "", errors.New("invalid API key")
	return "user-1", nil
}

// Start starts the HTTP server with default routes.
func Start(addr string) error {
	http.HandleFunc(wsPath, HandleWebSocket)
	http.HandleFunc("/", HandleProxyRequest)
	log.Printf("Starting server on %s", addr)
	log.Printf("WebSocket endpoint available at ws://%s%s", addr, wsPath)
	log.Printf("HTTP proxy available at http://%s/", addr)
	return http.ListenAndServe(addr, nil)
}
