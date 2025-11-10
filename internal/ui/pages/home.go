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

func (page *HomePage) Build() fyne.CanvasObject {
	// 标题
	title := widget.NewLabelWithStyle("MoCroc", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.TextStyle = fyne.TextStyle{Bold: true}

	// 副标题
	subtitle := widget.NewLabelWithStyle("点对点文件传输工具", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// 功能按钮
	sendBtn := widget.NewButtonWithIcon("发送文件", theme.ContentPasteIcon(), page.onSend)
	sendBtn.Importance = widget.HighImportance

	receiveBtn := widget.NewButtonWithIcon("接收文件", theme.DownloadIcon(), page.onReceive)
	receiveBtn.Importance = widget.HighImportance

	historyBtn := widget.NewButtonWithIcon("传输历史", theme.HistoryIcon(), page.onHistory)

	// 按钮样式设置
	for _, btn := range []*widget.Button{sendBtn, receiveBtn, historyBtn} {
		btn.Resize(fyne.NewSize(200, 60))
	}

	// 创建按钮容器
	buttonContainer := container.NewVBox(
		sendBtn,
		receiveBtn,
		historyBtn,
	)

	// 主内容布局
	content := container.NewVBox(
		title,
		subtitle,
		widget.NewLabel(""),
		buttonContainer,
		widget.NewLabel(""),
		widget.NewLabel(""),
		widget.NewLabel(""),
		widget.NewLabel(""),
	)

	// 居中布局
	return container.NewCenter(content)
}
