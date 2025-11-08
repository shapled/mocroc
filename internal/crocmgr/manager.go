package crocmgr

import (
	"context"
	"log"

	"github.com/schollz/croc/v10/src/croc"
)

// Manager 封装 Croc 操作
type Manager struct {
	crocClient *croc.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (m *Manager) GetContext() context.Context {
	return m.ctx
}

func (m *Manager) Cancel() {
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *Manager) CreateCrocClient(options croc.Options) (*croc.Client, error) {
	client, err := croc.New(options)
	if err != nil {
		return nil, err
	}
	m.crocClient = client
	return client, nil
}

func (m *Manager) GetCrocClient() *croc.Client {
	return m.crocClient
}

func (m *Manager) Close() {
	// 取消上下文
	if m.cancel != nil {
		m.cancel()
	}
	// 清理客户端
	m.crocClient = nil
}

func (m *Manager) Log(msg string) {
	log.Printf("[CrocMobile] %s", msg)
}