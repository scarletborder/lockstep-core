package session

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/quic-go/webtransport-go"
)

// webtransport session

type WtSession struct {
	ctx    context.Context
	cancel context.CancelFunc
	// webtransport 会话
	session *webtransport.Session
}

func NewWtSession(session *webtransport.Session) *WtSession {
	ctx, cancel := context.WithCancel(context.Background())
	return &WtSession{
		session: session,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (ws *WtSession) Close() error {
	if ws.cancel != nil {
		ws.cancel()
		ws.cancel = nil
	}
	return ws.session.CloseWithError(0, "player disconnected")
}

func (ws *WtSession) IsConnected() bool {
	return ws.session != nil && ws.session.Context().Err() == nil
}
func (ws *WtSession) SendDatagram(data []byte) error {
	if !ws.IsConnected() {
		return fmt.Errorf("session is nil")
	}
	err := ws.session.SendDatagram(data)
	if err != nil {
		log.Printf("🔴 SendDatagram error for player %v: %v", ws.session.RemoteAddr(), err)
	}
	return err
}

func (ws *WtSession) ReceiveDatagram() ([]byte, error) {
	if !ws.IsConnected() {
		return nil, fmt.Errorf("session is nil")
	}
	data, err := ws.session.ReceiveDatagram(ws.ctx)
	if err != nil {
		log.Printf("🔴 ReceiveDatagram error: %v", err)
	}
	return data, err
}

func (ws *WtSession) GetRemoteAddr() net.Addr {
	return ws.session.RemoteAddr()
}
