package types

import "fyne.io/fyne/v2"

// TabState 标签页状态
type TabState int

const (
	TabStateIdle TabState = iota
	TabStateSending
	TabStateReceiving
)

// TabInterface 定义标签页接口
type TabInterface interface {
	Build() fyne.CanvasObject
	GetState() TabState
	Cancel() error
	IsActive() bool
	SetActive(active bool)
}