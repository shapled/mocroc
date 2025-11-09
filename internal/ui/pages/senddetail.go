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
	SendDetailStateWaiting   // 等待接收端连接
	SendDetailStateSending   // 正在发送数据
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

func (p *SendDetailPage) SetFileName(name string) {
	p.fileName = name
}

func (p *SendDetailPage) SetCode(code string) {
	p.code = code
}

func (p *SendDetailPage) SetState(state SendDetailState) {
	p.state = state
}

func (p *SendDetailPage) SetStatusMessage(msg string) {
	p.statusMsg = msg
}

// SetStateAndMessage 同时设置状态和消息
func (p *SendDetailPage) SetStateAndMessage(state SendDetailState, message string) {
	p.state = state
	p.statusMsg = message
}

// SetProgress 设置进度
func (p *SendDetailPage) SetProgress(progress float64) {
	p.progress = progress
}

func (p *SendDetailPage) Build() fyne.CanvasObject {
	// 标题
	title := widget.NewLabelWithStyle("发送详情", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// 信息卡片
	infoCard := widget.NewCard("传输信息", "", container.NewVBox(
		p.createInfoRow("文件:", p.fileName, "准备中..."),
		p.createInfoRow("接收码:", p.code, "生成中..."),
		p.createInfoRow("状态:", p.getStateText(), ""),
	))

	// 进度卡片
	var progressCard fyne.CanvasObject
	if p.state == SendDetailStateSending {
		progressBar := widget.NewProgressBar()
		progressBar.SetValue(p.progress)
		progressCard = widget.NewCard("传输进度", "", container.NewVBox(
			progressBar,
			widget.NewLabel(fmt.Sprintf("%.1f%%", p.progress*100)),
		))
	} else if p.state == SendDetailStateWaiting {
		// 等待状态显示无限进度条
		progressBar := widget.NewProgressBarInfinite()
		progressCard = widget.NewCard("等待连接", "", container.NewVBox(
			progressBar,
			widget.NewLabel("等待接收端输入接收码..."),
		))
	} else {
		progressCard = widget.NewLabel("")
	}

	// 状态消息
	if p.statusMsg == "" {
		p.statusMsg = "准备发送..."
	}
	statusCard := widget.NewCard("状态信息", "", widget.NewLabel(p.statusMsg))

	// 操作按钮
	var actionButton *widget.Button
	switch p.state {
	case SendDetailStatePreparing, SendDetailStateWaiting, SendDetailStateSending:
		actionButton = widget.NewButtonWithIcon("取消发送", theme.CancelIcon(), p.onCancel)
	case SendDetailStateCompleted:
		actionButton = widget.NewButtonWithIcon("完成", theme.ConfirmIcon(), p.onBack)
		actionButton.Importance = widget.HighImportance
	case SendDetailStateFailed, SendDetailStateCancelled:
		actionButton = widget.NewButtonWithIcon("重新发送", theme.ViewRefreshIcon(), p.onBack)
		actionButton.Importance = widget.MediumImportance
	default:
		actionButton = widget.NewButtonWithIcon("返回", theme.NavigateBackIcon(), p.onBack)
	}

	actionCard := widget.NewCard("操作", "", actionButton)

	// 主内容 - 使用边框布局让内容更好地填充空间
	mainContent := container.NewVBox(
		widget.NewLabel(""), // 顶部间距
		title,
		widget.NewLabel(""),
		infoCard,
		widget.NewLabel(""),
		progressCard,
		widget.NewLabel(""),
		statusCard,
		widget.NewLabel(""),
		actionCard,
		widget.NewLabel(""), // 底部间距
	)

	// 使用滚动容器
	return container.NewScroll(container.NewPadded(mainContent))
}

// createInfoRow 创建信息行
func (p *SendDetailPage) createInfoRow(label, value, placeholder string) fyne.CanvasObject {
	labelWidget := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	valueWidget := widget.NewLabel(value)
	if value == "" {
		valueWidget.SetText(placeholder)
	}

	return container.NewHBox(labelWidget, valueWidget)
}

func (p *SendDetailPage) getStateText() string {
	switch p.state {
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