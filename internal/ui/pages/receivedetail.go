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

func (page *ReceiveDetailPage) SetFileName(name string) {
	page.fileName = name
}

func (page *ReceiveDetailPage) SetSenderInfo(info string) {
	page.senderInfo = info
}

func (page *ReceiveDetailPage) SetState(state ReceiveDetailState) {
	page.state = state
}

func (page *ReceiveDetailPage) SetProgress(progress float64) {
	page.progress = progress
}

func (page *ReceiveDetailPage) SetStatusMessage(msg string) {
	page.statusMsg = msg
}

func (page *ReceiveDetailPage) SetSavePath(path string) {
	page.savePath = path
}

func (page *ReceiveDetailPage) Build() fyne.CanvasObject {
	// 信息卡片
	infoCard := widget.NewCard("传输信息", "", container.NewVBox(
		page.createInfoRow("文件:", page.fileName, "等待信息..."),
		page.createInfoRow("发送者:", page.senderInfo, "获取中..."),
		page.createInfoRow("保存到:", page.savePath, "默认下载目录"),
		page.createInfoRow("状态:", page.getStateText(), ""),
	))

	// 进度卡片
	var progressCard fyne.CanvasObject
	if page.state == ReceiveDetailStateReceiving || page.state == ReceiveDetailStateConnecting {
		progressBar := widget.NewProgressBar()
		progressBar.SetValue(page.progress)
		progressCard = widget.NewCard("传输进度", "", container.NewVBox(
			progressBar,
			widget.NewLabel(fmt.Sprintf("%.1f%%", page.progress*100)),
		))
	} else {
		progressCard = widget.NewLabel("")
	}

	// 状态消息
	if page.statusMsg == "" {
		page.statusMsg = "正在连接发送方..."
	}
	statusCard := widget.NewCard("状态信息", "", widget.NewLabel(page.statusMsg))

	// 操作按钮
	var actionButton *widget.Button
	switch page.state {
	case ReceiveDetailStateConnecting, ReceiveDetailStateReceiving:
		actionButton = widget.NewButtonWithIcon("取消接收", theme.CancelIcon(), page.onCancel)
	case ReceiveDetailStateCompleted:
		actionButton = widget.NewButtonWithIcon("完成", theme.ConfirmIcon(), page.onBack)
		actionButton.Importance = widget.HighImportance
	case ReceiveDetailStateFailed, ReceiveDetailStateCancelled:
		actionButton = widget.NewButtonWithIcon("重新接收", theme.ViewRefreshIcon(), page.onBack)
		actionButton.Importance = widget.MediumImportance
	default:
		actionButton = widget.NewButtonWithIcon("返回", theme.NavigateBackIcon(), page.onBack)
	}

	actionCard := widget.NewCard("操作", "", actionButton)

	// 主内容
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
func (page *ReceiveDetailPage) createInfoRow(label, value, placeholder string) fyne.CanvasObject {
	labelWidget := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	valueWidget := widget.NewLabel(value)
	if value == "" {
		valueWidget.SetText(placeholder)
	}

	return container.NewHBox(labelWidget, valueWidget)
}

func (page *ReceiveDetailPage) getStateText() string {
	switch page.state {
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
