package service

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WsProxy OpenClaw WebSocket 代理（支持启停和端口配置）
type WsProxy struct {
	mu        sync.Mutex
	port      int
	authToken string
	upgrader  websocket.Upgrader
	listener  net.Listener
	server    *http.Server
	running   bool
	useSSL    bool

	// 活跃连接追踪（Stop 时强制关闭）
	connsMu sync.Mutex
	conns   []*websocket.Conn
}

// 全局单例
var globalProxy *WsProxy

func init() {
	globalProxy = &WsProxy{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// GetGlobalProxy 获取全局 WS 代理实例
func GetGlobalProxy() *WsProxy {
	return globalProxy
}

// GetPort 获取代理端口
func (p *WsProxy) GetPort() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.port
}

// IsRunning 是否正在运行
func (p *WsProxy) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// GetAuthToken 获取认证令牌（懒加载）
func (p *WsProxy) GetAuthToken() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.authToken == "" {
		b := make([]byte, 24)
		rand.Read(b)
		p.authToken = fmt.Sprintf("%x", b)
	}
	return p.authToken
}

// CheckPort 检查端口是否可用
func CheckPort(port int) (bool, string) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false, fmt.Sprintf("端口 %d 已被占用: %v", port, err)
	}
	ln.Close()
	return true, ""
}

// trackConn 追踪活跃 WS 连接
func (p *WsProxy) trackConn(conn *websocket.Conn) {
	p.connsMu.Lock()
	p.conns = append(p.conns, conn)
	p.connsMu.Unlock()
}

// untrackConn 移除已关闭的连接
func (p *WsProxy) untrackConn(conn *websocket.Conn) {
	p.connsMu.Lock()
	defer p.connsMu.Unlock()
	for i, c := range p.conns {
		if c == conn {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			return
		}
	}
}

// closeAllConns 关闭所有活跃连接
func (p *WsProxy) closeAllConns() {
	p.connsMu.Lock()
	conns := make([]*websocket.Conn, len(p.conns))
	copy(conns, p.conns)
	p.conns = nil
	p.connsMu.Unlock()

	for _, c := range conns {
		c.Close()
	}
	if len(conns) > 0 {
		log.Printf("[WS-Proxy] 已关闭 %d 个活跃连接", len(conns))
	}
}

// Start 启动 WS 代理（指定端口，0 = 随机）
func (p *WsProxy) Start(port int, useSSL bool) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return p.port, nil // 已在运行
	}

	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("端口 %d 监听失败（可能已被占用）: %v", port, err)
	}
	actualPort := ln.Addr().(*net.TCPAddr).Port

	// 确保 authToken 已生成（在同一锁内完成，避免死锁）
	if p.authToken == "" {
		b := make([]byte, 24)
		rand.Read(b)
		p.authToken = fmt.Sprintf("%x", b)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/chat", p.handleWs)

	p.listener = ln
	p.server = &http.Server{Handler: mux}
	p.port = actualPort
	p.running = true
	p.useSSL = useSSL

	if useSSL {
		// 生成自签名 TLS 证书（内存中，不落盘）
		tlsCert, err := generateSelfSignedCert()
		if err != nil {
			ln.Close()
			p.running = false
			return 0, fmt.Errorf("生成 TLS 证书失败: %v", err)
		}
		p.server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		}
		tlsListener := tls.NewListener(ln, p.server.TLSConfig)
		p.listener = tlsListener
		log.Printf("[WS-Proxy] 🔒 WSS 模式启动，端口 %d", actualPort)
		go func() {
			if err := p.server.Serve(tlsListener); err != nil && err != http.ErrServerClosed {
				log.Printf("[WS-Proxy] 服务退出: %v", err)
			}
			p.mu.Lock()
			p.running = false
			p.mu.Unlock()
		}()
	} else {
		log.Printf("[WS-Proxy] WS 模式启动，端口 %d", actualPort)
		go func() {
			if err := p.server.Serve(ln); err != nil && err != http.ErrServerClosed {
				log.Printf("[WS-Proxy] 服务退出: %v", err)
			}
			p.mu.Lock()
			p.running = false
			p.mu.Unlock()
		}()
	}

	// 更新全局信息供 GetClawWsInfo 使用
	SetWsProxyInfo(actualPort, p.authToken)

	return actualPort, nil
}

// generateSelfSignedCert 生成自签名 TLS 证书（内存中，有效期 1 年）
func generateSelfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"GMSSH WS Proxy"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost", "*.gmssh.com"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("0.0.0.0")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}

// Stop 停止 WS 代理
func (p *WsProxy) Stop() {
	// 先关闭所有活跃 WS 连接（不持 mu 锁，避免和 handleWs 死锁）
	p.closeAllConns()

	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}
	// 关闭 listener 释放端口
	if p.listener != nil {
		p.listener.Close()
	}
	if p.server != nil {
		p.server.Close()
	}
	p.running = false
	p.port = 0
	p.listener = nil
	p.server = nil
	// 清空全局信息
	SetWsProxyInfo(0, "")
	// 等待端口完全释放
	time.Sleep(200 * time.Millisecond)
}

// getClawConnInfo 从 openclaw.json 获取 gateway 端口和 token
func getClawConnInfo() (int, string, error) {
	config, err := readOpenClawConfig()
	if err != nil {
		return 0, "", fmt.Errorf("读取配置失败: %v", err)
	}
	var port int
	var token string

	if gw, ok := config["gateway"].(map[string]any); ok {
		if p, ok := gw["port"].(float64); ok {
			port = int(p)
		}
		// token 在 gateway.auth.token
		if auth, ok := gw["auth"].(map[string]any); ok {
			if t, ok := auth["token"].(string); ok {
				token = t
			}
		}
	}
	return port, token, nil
}

// handleWs 处理前端 WS 连接
func (p *WsProxy) handleWs(w http.ResponseWriter, r *http.Request) {

	// 0. 验证认证令牌
	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		log.Printf("[WS-Proxy] ❌ 认证失败: 缺少令牌 from %s", r.RemoteAddr)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	decoded, err := base64.RawURLEncoding.DecodeString(rawToken)
	if err != nil || string(decoded) != p.authToken {
		log.Printf("[WS-Proxy] ❌ 认证失败: 无效令牌 from %s", r.RemoteAddr)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 1. 读取 OpenClaw 配置
	clawPort, clawToken, err := getClawConnInfo()
	if err != nil {
		log.Printf("[WS-Proxy] ⚠️ OpenClaw 未部署或配置不存在")
		errConn, upgradeErr := p.upgrader.Upgrade(w, r, nil)
		if upgradeErr == nil {
			errConn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "OpenClaw 未部署"))
			errConn.Close()
		}
		return
	}


	// 2. 升级前端连接为 WS
	frontConn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS-Proxy] ❌ 前端 WS 升级失败: %v", err)
		return
	}
	// 追踪连接，Stop 时可强制关闭
	p.trackConn(frontConn)
	defer func() {
		p.untrackConn(frontConn)
		frontConn.Close()
	}()

	// 3. 连接到 OpenClaw gateway
	clawUrl := fmt.Sprintf("ws://127.0.0.1:%d", clawPort)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	requestHeader := http.Header{}
	requestHeader.Set("Origin", fmt.Sprintf("http://127.0.0.1:%d", clawPort))
	clawConn, _, err := dialer.Dial(clawUrl, requestHeader)
	if err != nil {
		log.Printf("[WS-Proxy] ❌ 连接 OpenClaw 失败: %v", err)
		frontConn.WriteJSON(map[string]any{
			"type":  "event",
			"event": "proxy.error",
			"payload": map[string]any{
				"message": fmt.Sprintf("连接 OpenClaw 失败: %v", err),
			},
		})
		return
	}
	// 追踪 OpenClaw 连接
	p.trackConn(clawConn)
	defer func() {
		p.untrackConn(clawConn)
		clawConn.Close()
	}()

	// 4. 进行 challenge 认证
	authOk := false

	_, rawMsg, err := clawConn.ReadMessage()
	if err != nil {
		log.Printf("[WS-Proxy] ❌ 读取 challenge 失败: %v", err)
		return
	}

	var challengeMsg map[string]any
	if err := json.Unmarshal(rawMsg, &challengeMsg); err != nil {
		log.Printf("[WS-Proxy] ❌ 解析 challenge 失败: %v", err)
		return
	}

	if challengeMsg["type"] == "event" && challengeMsg["event"] == "connect.challenge" {

		connectReq := map[string]any{
			"type":   "req",
			"id":     genUUID(),
			"method": "connect",
			"params": map[string]any{
				"minProtocol": 3,
				"maxProtocol": 3,
				"client": map[string]any{
					"id":       "openclaw-control-ui",
					"version":  "dev",
					"platform": "linux",
					"mode":     "webchat",
				},
				"role":   "operator",
				"scopes": []string{"operator.admin", "operator.approvals", "operator.pairing"},
				"caps":   []any{},
				"auth": map[string]any{
					"token": clawToken,
				},
				"userAgent": "Mozilla/5.0 (Linux) GMClaw-WsProxy/1.0",
				"locale":    "zh-CN",
			},
		}
		reqJSON, _ := json.MarshalIndent(connectReq, "", "  ")
		log.Printf("[WS-Proxy] 📤 challenge: %s", string(rawMsg))
		log.Printf("[WS-Proxy] 📤 发送认证请求: %s", string(reqJSON))
		if err := clawConn.WriteJSON(connectReq); err != nil {
			log.Printf("[WS-Proxy] ❌ 发送认证失败: %v", err)
			return
		}

		_, rawRes, err := clawConn.ReadMessage()
		if err != nil {
			log.Printf("[WS-Proxy] ❌ 读取认证响应失败: %v", err)
			return
		}

		var authRes map[string]any
		if err := json.Unmarshal(rawRes, &authRes); err == nil {
			if payload, ok := authRes["payload"].(map[string]any); ok {
				if payload["type"] == "hello-ok" {
					authOk = true
				}
			}
		}

		if !authOk {
			log.Printf("[WS-Proxy] ❌ 认证失败, 响应内容: %s", string(rawRes))
		} else {
			log.Printf("[WS-Proxy] ✅ 认证成功")
		}
	} else {
		log.Printf("[WS-Proxy] ❌ 第一条消息不是 connect.challenge: type=%v event=%v", challengeMsg["type"], challengeMsg["event"])
	}

	if !authOk {
		frontConn.WriteJSON(map[string]any{
			"type":  "event",
			"event": "proxy.error",
			"payload": map[string]any{
				"message": "OpenClaw 认证失败",
			},
		})
		return
	}

	// 5. 通知前端连接就绪
	frontConn.WriteJSON(map[string]any{
		"type":  "event",
		"event": "proxy.connected",
		"payload": map[string]any{
			"message": "已连接到 OpenClaw",
		},
	})

	// 6. 双向代理消息
	var wg sync.WaitGroup
	wg.Add(2)

	allowedEvents := map[string]bool{
		"agent": true,
		"chat":  true,
	}

	// OpenClaw → 前端
	go func() {
		defer wg.Done()
		for {
			msgType, msg, err := clawConn.ReadMessage()
			if err != nil {
				frontConn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "OpenClaw disconnected"))
				return
			}
			if msgType == websocket.TextMessage {
				var peek struct {
					Type  string `json:"type"`
					Event string `json:"event"`
				}
				if json.Unmarshal(msg, &peek) == nil {
					if peek.Type == "event" && !allowedEvents[peek.Event] {
						continue
					}
				}
			}
			if err := frontConn.WriteMessage(msgType, msg); err != nil {
				log.Printf("[WS-Proxy] 发送到前端失败: %v", err)
				return
			}
		}
	}()

	// 前端 → OpenClaw
	go func() {
		defer wg.Done()
		for {
			msgType, msg, err := frontConn.ReadMessage()
			if err != nil {
				clawConn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client disconnected"))
				return
			}
			if err := clawConn.WriteMessage(msgType, msg); err != nil {
				log.Printf("[WS-Proxy] 发送到 OpenClaw 失败: %v", err)
				return
			}
		}
	}()

	wg.Wait()
	log.Printf("[WS-Proxy] 会话结束")
}

// genUUID 使用 crypto/rand 生成 UUID v4
func genUUID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
