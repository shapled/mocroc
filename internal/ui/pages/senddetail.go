package pages

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SendDetailState int

const (
	SendDetailStatePreparing SendDetailState = iota
	SendDetailStateWaiting                   // 等待接收端连接
	SendDetailStateSending                   // 正在发送数据
	SendDetailStateCompleted
	SendDetailStateFailed
	SendDetailStateCancelled
)

type SendDetailPage struct {
	window    fyne.Window
	onBack    func()
	onCancel  func()
	state     SendDetailState
	fileName  string
	code      string
	progress  float64
	statusMsg string
}

func NewSendDetailPage(window fyne.Window, onBack, onCancel func()) *SendDetailPage {
	return &SendDetailPage{
		window:   window,
		onBack:   onBack,
		onCancel: onCancel,
		state:    SendDetailStatePreparing,
		progress: 0.0,
	}
}

func (page *SendDetailPage) SetFileName(name string) {
	page.fileName = name
}

func (page *SendDetailPage) SetCode(code string) {
	page.code = code
}

func (page *SendDetailPage) SetState(state SendDetailState) {
	page.state = state
}

func (page *SendDetailPage) SetStatusMessage(msg string) {
	page.statusMsg = msg
}

// SetStateAndMessage 同时设置状态和消息
func (page *SendDetailPage) SetStateAndMessage(state SendDetailState, message string) {
	page.state = state
	page.statusMsg = message
}

// SetProgress 设置进度
func (page *SendDetailPage) SetProgress(progress float64) {
	page.progress = progress
}

func (page *SendDetailPage) Build() fyne.CanvasObject {
	// 信息卡片
	infoCard := widget.NewCard("传输信息", "", container.NewVBox(
		page.createInfoRow("文件:", page.fileName, "准备中..."),
		page.createInfoRow("接收码:", page.code, "生成中..."),
		page.createInfoRow("状态:", page.getStateText(), ""),
	))

	// 进度卡片
	var progressCard fyne.CanvasObject
	switch page.state {
	case SendDetailStateSending:
		progressBar := widget.NewProgressBar()
		progressBar.SetValue(page.progress)
		progressCard = widget.NewCard("传输进度", "", container.NewVBox(
			progressBar,
			widget.NewLabel(fmt.Sprintf("%.1f%%", page.progress*100)),
		))
	case SendDetailStateWaiting:
		// 等待状态显示无限进度条
		progressBar := widget.NewProgressBarInfinite()
		progressCard = widget.NewCard("等待连接", "", container.NewVBox(
			progressBar,
			widget.NewLabel("等待接收端输入接收码..."),
		))
	default:
		progressCard = widget.NewLabel("")
	}

	// 状态消息
	if page.statusMsg == "" {
		page.statusMsg = "准备发送..."
	}
	statusCard := widget.NewCard("状态信息", "", widget.NewLabel(page.statusMsg))

	// 操作按钮
	var actionButton *widget.Button
	switch page.state {
	case SendDetailStatePreparing, SendDetailStateWaiting, SendDetailStateSending:
		actionButton = widget.NewButtonWithIcon("取消发送", theme.CancelIcon(), page.onCancel)
	case SendDetailStateCompleted:
		actionButton = widget.NewButtonWithIcon("完成", theme.ConfirmIcon(), page.onBack)
		actionButton.Importance = widget.HighImportance
	case SendDetailStateFailed, SendDetailStateCancelled:
		actionButton = widget.NewButtonWithIcon("重新发送", theme.ViewRefreshIcon(), page.onBack)
		actionButton.Importance = widget.MediumImportance
	default:
		actionButton = widget.NewButtonWithIcon("返回", theme.NavigateBackIcon(), page.onBack)
	}

	actionCard := widget.NewCard("操作", "", actionButton)

	// 主内容 - 使用边框布局让内容更好地填充空间
	mainContent := container.NewVBox(
		infoCard,
		progressCard,
		statusCard,
		actionCard,
	)

	// 使用滚动容器
	return container.NewScroll(container.NewPadded(mainContent))
}

// createInfoRow 创建信息行
func (page *SendDetailPage) createInfoRow(label, value, placeholder string) fyne.CanvasObject {
	labelWidget := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	valueWidget := widget.NewLabel(value)
	if value == "" {
		valueWidget.SetText(placeholder)
	}

	return container.NewHBox(labelWidget, valueWidget)
}

func (page *SendDetailPage) getStateText() string {
	switch page.state {
	case SendDetailStatePreparing:
		return "准备中"
	case SendDetailStateWaiting:
		return "等待接收端连接"
	case SendDetailStateSending:
		return "发送中"
	case SendDetailStateCompleted:
		return "发送完成"
	case SendDetailStateFailed:
		return "发送失败"
	case SendDetailStateCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}
