package server

import (
	"lockstep-core/src/config"
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

// WebTransportServer 封装 WebTransport 服务器
type WebTransportServer struct {
	config   *config.ServerConfig
	wtServer *webtransport.Server
	mux      *http.ServeMux
}

// NewWebTransportServer 创建一个新的 WebTransportServer
func NewWebTransportServer(cfg *config.ServerConfig) *WebTransportServer {
	mux := http.NewServeMux()

	wtServer := &webtransport.Server{
		H3: http3.Server{
			Addr:      cfg.Addr,
			Handler:   mux,
			TLSConfig: cfg.TLSConfig,
		},
		CheckOrigin: func(r *http.Request) bool {
			if !cfg.CheckOriginEnabled {
				return true
			}
			// 在生产环境中应进行严格的来源检查
			// 这里可以根据需要添加自定义的 origin 检查逻辑
			return true
		},
	}

	return &WebTransportServer{
		config:   cfg,
		wtServer: wtServer,
		mux:      mux,
	}
}

// GetMux 获取 HTTP 多路复用器
func (s *WebTransportServer) GetMux() *http.ServeMux {
	return s.mux
}

// GetWTServer 获取 WebTransport 服务器实例
func (s *WebTransportServer) GetWTServer() *webtransport.Server {
	return s.wtServer
}

// RegisterHandler 注册 HTTP 处理器
func (s *WebTransportServer) RegisterHandler(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

// Start 启动服务器
func (s *WebTransportServer) Start() error {
	log.Printf("Starting WebTransport server on %s", s.config.Addr)
	return s.wtServer.ListenAndServe()
}

// UpgradeToWebTransport 升级 HTTP 连接到 WebTransport
func (s *WebTransportServer) UpgradeToWebTransport(w http.ResponseWriter, r *http.Request) (*webtransport.Session, error) {
	session, err := s.wtServer.Upgrade(w, r)
	if err != nil {
		log.Printf("Failed to upgrade to WebTransport: %v", err)
		return nil, err
	}
	return session, nil
}
