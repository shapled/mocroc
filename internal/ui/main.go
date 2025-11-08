package ui

import (
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/shapled/mocroc/internal/crocmgr"
	"github.com/shapled/mocroc/internal/types"
	"github.com/shapled/mocroc/internal/ui/tabs"
)

type MainUI struct {
	app    fyne.App
	window fyne.Window

	// Croc 管理器
	crocManager *crocmgr.Manager

	// 标签页
	sendTab      types.TabInterface
	receiveTab   types.TabInterface
	historyTab   types.TabInterface
	tabs         *container.AppTabs
}

func NewMainUI(a fyne.App, w fyne.Window) fyne.CanvasObject {
	mainUI := &MainUI{
		app:         a,
		window:      w,
		crocManager: crocmgr.NewManager(),
	}

	// 初始化标签页，传递 Croc 管理器
	mainUI.createTabs()
	mainUI.buildMainWindow()

	return mainUI.tabs
}

func (ui *MainUI) createTabs() {
	// 创建各个标签页
	ui.sendTab = tabs.NewSendTab(ui.crocManager, ui.window)
	ui.receiveTab = tabs.NewReceiveTab(ui.crocManager, ui.window)
	ui.historyTab = tabs.NewHistoryTab()

	// 创建主标签页容器
	ui.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("发送", theme.ContentPasteIcon(), ui.sendTab.Build()),
		container.NewTabItemWithIcon("接收", theme.DownloadIcon(), ui.receiveTab.Build()),
		container.NewTabItemWithIcon("历史", theme.HistoryIcon(), ui.historyTab.Build()),
	)

	// 设置标签页切换监听器
	ui.tabs.OnSelected = func(tab *container.TabItem) {
		ui.onTabChanged(tab)
	}

	// 平台特定的标签位置
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		ui.tabs.SetTabLocation(container.TabLocationTop)
	} else {
		ui.tabs.SetTabLocation(container.TabLocationBottom)
	}

	// 设置初始状态
	ui.sendTab.SetActive(true)
	ui.receiveTab.SetActive(false)
	ui.historyTab.SetActive(false)
}

func (ui *MainUI) buildMainWindow() {
	ui.window.SetContent(ui.tabs)

	// 平台特定的窗口大小
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		ui.window.Resize(fyne.NewSize(900, 700))
		ui.window.SetFixedSize(false)
		ui.window.CenterOnScreen()
	} else {
		ui.window.Resize(fyne.NewSize(360, 640))
		ui.window.SetFixedSize(true)
	}
}

// onTabChanged 处理标签页切换，实现互斥逻辑
func (ui *MainUI) onTabChanged(tab *container.TabItem) {
	// 检查传输状态
	if ui.sendTab.GetState() == types.TabStateSending {
		dialog.ShowInformation("提示", "正在发送，请先取消后再切换标签页", ui.window)
		ui.tabs.SelectTabIndex(0)
		return
	}
	if ui.receiveTab.GetState() == types.TabStateReceiving {
		dialog.ShowInformation("提示", "正在接收，请先取消后再切换标签页", ui.window)
		ui.tabs.SelectTabIndex(1)
		return
	}

	// 重置所有标签页的活跃状态
	ui.sendTab.SetActive(false)
	ui.receiveTab.SetActive(false)
	ui.historyTab.SetActive(false)

	// 设置当前标签页为活跃状态
	switch ui.tabs.SelectedIndex() {
	case 0:
		ui.sendTab.SetActive(true)
	case 1:
		ui.receiveTab.SetActive(true)
	case 2:
		ui.historyTab.SetActive(true)
		if historyTab, ok := ui.historyTab.(*tabs.HistoryTab); ok {
			historyTab.Refresh()
		}
	}
}

// Close 关闭资源
func (ui *MainUI) Close() {
	if ui.crocManager != nil {
		ui.crocManager.Close()
	}
}
