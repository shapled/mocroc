package pages

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ReceiveDetailState int

const (
	ReceiveDetailStateConnecting ReceiveDetailState = iota
	ReceiveDetailStateReceiving
	ReceiveDetailStateCompleted
	ReceiveDetailStateFailed
	ReceiveDetailStateCancelled
)

type ReceiveDetailPage struct {
	window     fyne.Window
	onBack     func()
	onCancel   func()
	state      ReceiveDetailState
	fileName   string
	senderInfo string
	progress   float64
	statusMsg  string
	savePath   string
}

func NewReceiveDetailPage(window fyne.Window, onBack, onCancel func()) *ReceiveDetailPage {
	return &ReceiveDetailPage{
		window:   window,
		onBack:   onBack,
		onCancel: onCancel,
		state:    ReceiveDetailStateConnecting,
		progress: 0.0,
	}
}

func (p *ReceiveDetailPage) SetFileName(name string) {
	p.fileName = name
}

func (p *ReceiveDetailPage) SetSenderInfo(info string) {
	p.senderInfo = info
}

func (p *ReceiveDetailPage) SetState(state ReceiveDetailState) {
	p.state = state
}

func (p *ReceiveDetailPage) SetProgress(progress float64) {
	p.progress = progress
}

func (p *ReceiveDetailPage) SetStatusMessage(msg string) {
	p.statusMsg = msg
}

func (p *ReceiveDetailPage) SetSavePath(path string) {
	p.savePath = path
}

func (p *ReceiveDetailPage) Build() fyne.CanvasObject {
	// 信息卡片
	infoCard := widget.NewCard("传输信息", "", container.NewVBox(
		p.createInfoRow("文件:", p.fileName, "等待信息..."),
		p.createInfoRow("发送者:", p.senderInfo, "获取中..."),
		p.createInfoRow("保存到:", p.savePath, "默认下载目录"),
		p.createInfoRow("状态:", p.getStateText(), ""),
	))

	// 进度卡片
	var progressCard fyne.CanvasObject
	if p.state == ReceiveDetailStateReceiving || p.state == ReceiveDetailStateConnecting {
		progressBar := widget.NewProgressBar()
		progressBar.SetValue(p.progress)
		progressCard = widget.NewCard("传输进度", "", container.NewVBox(
			progressBar,
			widget.NewLabel(fmt.Sprintf("%.1f%%", p.progress*100)),
		))
	} else {
		progressCard = widget.NewLabel("")
	}

	// 状态消息
	if p.statusMsg == "" {
		p.statusMsg = "正在连接发送方..."
	}
	statusCard := widget.NewCard("状态信息", "", widget.NewLabel(p.statusMsg))

	// 操作按钮
	var actionButton *widget.Button
	switch p.state {
	case ReceiveDetailStateConnecting, ReceiveDetailStateReceiving:
		actionButton = widget.NewButtonWithIcon("取消接收", theme.CancelIcon(), p.onCancel)
	case ReceiveDetailStateCompleted:
		actionButton = widget.NewButtonWithIcon("完成", theme.ConfirmIcon(), p.onBack)
		actionButton.Importance = widget.HighImportance
	case ReceiveDetailStateFailed, ReceiveDetailStateCancelled:
		actionButton = widget.NewButtonWithIcon("重新接收", theme.ViewRefreshIcon(), p.onBack)
		actionButton.Importance = widget.MediumImportance
	default:
		actionButton = widget.NewButtonWithIcon("返回", theme.NavigateBackIcon(), p.onBack)
	}

	actionCard := widget.NewCard("操作", "", actionButton)

	// 主内容
	mainContent := container.NewVBox(
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
func (p *ReceiveDetailPage) createInfoRow(label, value, placeholder string) fyne.CanvasObject {
	labelWidget := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	valueWidget := widget.NewLabel(value)
	if value == "" {
		valueWidget.SetText(placeholder)
	}

	return container.NewHBox(labelWidget, valueWidget)
}

func (p *ReceiveDetailPage) getStateText() string {
	switch p.state {
	case ReceiveDetailStateConnecting:
		return "连接中"
	case ReceiveDetailStateReceiving:
		return "接收中"
	case ReceiveDetailStateCompleted:
		return "接收完成"
	case ReceiveDetailStateFailed:
		return "接收失败"
	case ReceiveDetailStateCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}
