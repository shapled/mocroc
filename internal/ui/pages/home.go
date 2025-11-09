package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type HomePage struct {
	window    fyne.Window
	onSend    func()
	onReceive func()
	onHistory func()
}

func NewHomePage(window fyne.Window, onSend, onReceive, onHistory func()) *HomePage {
	return &HomePage{
		window:    window,
		onSend:    onSend,
		onReceive: onReceive,
		onHistory: onHistory,
	}
}

func (h *HomePage) Build() fyne.CanvasObject {
	// 标题
	title := widget.NewLabelWithStyle("Croc Mobile", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.TextStyle = fyne.TextStyle{Bold: true}

	// 副标题
	subtitle := widget.NewLabelWithStyle("点对点文件传输工具", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// 功能按钮
	sendBtn := widget.NewButtonWithIcon("发送文件", theme.ContentPasteIcon(), h.onSend)
	sendBtn.Importance = widget.HighImportance

	receiveBtn := widget.NewButtonWithIcon("接收文件", theme.DownloadIcon(), h.onReceive)
	receiveBtn.Importance = widget.HighImportance

	historyBtn := widget.NewButtonWithIcon("传输历史", theme.HistoryIcon(), h.onHistory)

	// 按钮样式设置
	for _, btn := range []*widget.Button{sendBtn, receiveBtn, historyBtn} {
		btn.Resize(fyne.NewSize(200, 60))
	}

	// 创建按钮容器
	buttonContainer := container.NewVBox(
		// 添加间距
		widget.NewLabel(""),
		sendBtn,
		receiveBtn,
		historyBtn,
	)

	// 主内容布局
	content := container.NewVBox(
		title,
		widget.NewLabel(""),
		subtitle,
		widget.NewLabel(""),
		widget.NewLabel(""),
		buttonContainer,
	)

	// 居中布局
	return container.NewCenter(content)
}