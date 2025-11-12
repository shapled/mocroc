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
	// 标题 - 使用更大字体和更高重要性
	title := widget.NewLabelWithStyle("MoCroc", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance

	// 副标题 - 使用次要重要性
	subtitle := widget.NewLabelWithStyle("点对点文件传输工具", fyne.TextAlignCenter, fyne.TextStyle{})
	subtitle.Importance = widget.MediumImportance

	// 功能按钮 - 使用一致的主题色和图标
	sendBtn := widget.NewButtonWithIcon("发送文件", theme.UploadIcon(), page.onSend)
	sendBtn.Importance = widget.HighImportance

	receiveBtn := widget.NewButtonWithIcon("接收文件", theme.DownloadIcon(), page.onReceive)
	receiveBtn.Importance = widget.HighImportance

	historyBtn := widget.NewButtonWithIcon("传输历史", theme.HistoryIcon(), page.onHistory)
	historyBtn.Importance = widget.MediumImportance

	// 按钮样式设置 - 确保符合移动端标准 (至少48px高度)
	for _, btn := range []*widget.Button{sendBtn, receiveBtn, historyBtn} {
		btn.Resize(fyne.NewSize(280, 56)) // 增加宽度，保持合适的高度
	}

	// 创建按钮容器 - 添加合适的间距
	buttonContainer := container.NewVBox(
		sendBtn,
		widget.NewLabel(""), // 按钮间距
		receiveBtn,
		widget.NewLabel(""), // 按钮间距
		historyBtn,
	)

	// 主内容布局 - 使用更好的间距
	content := container.NewVBox(
		title,
		widget.NewLabel(""), // 标题间距
		subtitle,
		widget.NewLabel(""), // 副标题间距
		widget.NewLabel(""), // 额外间距
		buttonContainer,
	)

	// 居中布局并添加内边距
	centeredContent := container.NewCenter(content)
	return container.NewPadded(centeredContent)
}
