package ui

import (
	"context"

	"github.com/schollz/croc/v10/src/croc"
)

// CrocManager 定义 Croc 操作接口
type CrocManager interface {
	GetContext() context.Context
	Cancel()
	CreateCrocClient(options croc.Options) (*croc.Client, error)
	GetCrocClient() *croc.Client
	Close()
	Log(msg string)
}
