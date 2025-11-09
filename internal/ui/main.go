package ui

import (
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/shapled/mocroc/internal/crocmgr"
	"github.com/shapled/mocroc/internal/types"
	"github.com/shapled/mocroc/internal/ui/pages"
)

type PageType int

const (
	PageTypeHome PageType = iota
	PageTypeSend
	PageTypeSendDetail
	PageTypeReceive
	PageTypeReceiveDetail
	PageTypeHistory
)

type MainUI struct {
	app    fyne.App
	window fyne.Window

	// Croc 管理器
	crocManager *crocmgr.Manager

	// 页面
	currentPage       PageType
	content           *container.Scroll
	homePage          *pages.HomePage
	sendTab           *pages.SendTab
	receiveTab        *pages.ReceiveTab
	historyTab        *pages.HistoryTab
	sendDetailPage    *pages.SendDetailPage
	receiveDetailPage *pages.ReceiveDetailPage

	// Top Bar
	backBtn *widget.Button
	title   *widget.Label
}

func NewMainUI(a fyne.App, w fyne.Window) fyne.CanvasObject {
	mainUI := &MainUI{
		app:         a,
		window:      w,
		crocManager: crocmgr.NewManager(),
		currentPage: PageTypeHome,
	}

	// 初始化页面
	mainUI.createPages()
	mainUI.buildMainWindow()

	return mainUI.buildMainContainer()
}

func (ui *MainUI) createPages() {
	// 创建后退按钮
	ui.backBtn = widget.NewButtonWithIcon("返回", theme.NavigateBackIcon(), func() {
		ui.goBack()
	})
	ui.backBtn.Hide() // 初始隐藏
	ui.title = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// 创建首页
	ui.homePage = pages.NewHomePage(ui.window,
		func() { ui.navigateTo(PageTypeSend) },
		func() { ui.navigateTo(PageTypeReceive) },
		func() { ui.navigateTo(PageTypeHistory) },
	)

	// 创建详情页面
	ui.sendDetailPage = pages.NewSendDetailPage(ui.window,
		func() { ui.navigateTo(PageTypeSend) },
		func() {
			// 取消发送
			if ui.sendTab != nil {
				ui.sendTab.Cancel()
			}
		},
	)

	ui.receiveDetailPage = pages.NewReceiveDetailPage(ui.window,
		func() { ui.navigateTo(PageTypeReceive) },
		func() {
			// 取消接收
			if ui.receiveTab != nil {
				ui.receiveTab.Cancel()
			}
		},
	)

	// 创建功能页面
	ui.sendTab = pages.NewSendTab(ui.crocManager, ui.window)
	ui.receiveTab = pages.NewReceiveTab(ui.crocManager, ui.window)
	ui.historyTab = pages.NewHistoryTab()

	// 设置导航回调
	ui.sendTab.SetOnNavigateToDetail(func() {
		// 设置详情页数据
		if ui.sendDetailPage != nil {
			fileName, code, _ := ui.sendTab.GetSendData()
			ui.sendDetailPage.SetFileName(fileName)
			ui.sendDetailPage.SetCode(code)
			ui.sendDetailPage.SetState(pages.SendDetailStateWaiting)
			ui.sendDetailPage.SetStatusMessage("等待接收端连接...")
		}
		ui.NavigateToSendDetail()
	})

	// 设置详情页更新回调
	ui.sendTab.SetOnUpdateDetail(func(state string, progress float64, message string) {
		if ui.sendDetailPage != nil {
			switch state {
			case "waiting":
				ui.sendDetailPage.SetStateAndMessage(pages.SendDetailStateWaiting, message)
			case "sending":
				ui.sendDetailPage.SetStateAndMessage(pages.SendDetailStateSending, message)
				ui.sendDetailPage.SetProgress(progress)
			case "completed":
				ui.sendDetailPage.SetStateAndMessage(pages.SendDetailStateCompleted, message)
				ui.sendDetailPage.SetProgress(1.0)
			case "failed":
				ui.sendDetailPage.SetStateAndMessage(pages.SendDetailStateFailed, message)
			case "cancelled":
				ui.sendDetailPage.SetStateAndMessage(pages.SendDetailStateCancelled, message)
			}
			// 刷新详情页显示
			ui.content.Refresh()
		}
	})

	// 设置接收页面导航回调
	ui.receiveTab.SetOnNavigateToDetail(func() {
		// 设置详情页数据
		if ui.receiveDetailPage != nil {
			code, savePath := ui.receiveTab.GetReceiveData()
			ui.receiveDetailPage.SetFileName("获取文件信息中...")
			ui.receiveDetailPage.SetSenderInfo("发送方 (" + code + ")")
			ui.receiveDetailPage.SetSavePath(savePath)
			ui.receiveDetailPage.SetState(pages.ReceiveDetailStateConnecting)
			ui.receiveDetailPage.SetStatusMessage("正在连接发送方...")
		}
		ui.NavigateToReceiveDetail()
	})

	// 创建内容容器 - 使用滚动容器让内容可以填满空间
	ui.content = container.NewScroll(ui.homePage.Build())
}

func (ui *MainUI) buildMainWindow() {
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

func (ui *MainUI) buildMainContainer() fyne.CanvasObject {
	headerContainer := container.NewStack(
		container.NewHBox(ui.backBtn),
		ui.title,
	)

	// 主容器 - 使用 Border 让内容填满剩余空间
	mainContainer := container.NewBorder(headerContainer, nil, nil, nil, ui.content)

	return mainContainer
}

func (ui *MainUI) navigateTo(pageType PageType) {
	// 检查是否可以导航（例如传输过程中不能切换）
	if !ui.canNavigateFromCurrentPage() {
		return
	}

	ui.currentPage = pageType
	ui.updateContent()
}

func (ui *MainUI) canNavigateFromCurrentPage() bool {
	// 检查传输状态
	if ui.sendTab.GetState() == types.TabStateSending {
		dialog.ShowInformation("提示", "正在发送，请先取消后再切换页面", ui.window)
		return false
	}
	if ui.receiveTab.GetState() == types.TabStateReceiving {
		dialog.ShowInformation("提示", "正在接收，请先取消后再切换页面", ui.window)
		return false
	}
	return true
}

func (ui *MainUI) SetTitle(title string) {
	if ui.title != nil {
		ui.title.SetText(title)
	}
}

func (ui *MainUI) updateContent() {
	var content fyne.CanvasObject

	switch ui.currentPage {
	case PageTypeHome:
		ui.SetTitle("MoCroc")
		content = ui.homePage.Build()
		ui.backBtn.Hide()
		ui.sendTab.SetActive(false)
		ui.receiveTab.SetActive(false)
		ui.historyTab.SetActive(false)
	case PageTypeSend:
		ui.SetTitle("发送")
		content = ui.sendTab.Build()
		ui.backBtn.Show()
		ui.sendTab.SetActive(true)
		ui.receiveTab.SetActive(false)
		ui.historyTab.SetActive(false)
	case PageTypeSendDetail:
		ui.SetTitle("发送详情")
		content = ui.sendDetailPage.Build()
		ui.backBtn.Show()
	case PageTypeReceive:
		ui.SetTitle("接收")
		content = ui.receiveTab.Build()
		ui.backBtn.Show()
		ui.sendTab.SetActive(false)
		ui.receiveTab.SetActive(true)
		ui.historyTab.SetActive(false)
	case PageTypeReceiveDetail:
		ui.SetTitle("接收详情")
		content = ui.receiveDetailPage.Build()
		ui.backBtn.Show()
	case PageTypeHistory:
		ui.SetTitle("历史")
		content = ui.historyTab.Build()
		ui.backBtn.Show()
		ui.sendTab.SetActive(false)
		ui.receiveTab.SetActive(false)
		ui.historyTab.SetActive(true)
		ui.historyTab.Refresh()
	default:
		ui.SetTitle("MoCroc")
		content = ui.homePage.Build()
		ui.backBtn.Hide()
	}

	ui.content.Content = content
	ui.content.Refresh()
}

func (ui *MainUI) goBack() {
	switch ui.currentPage {
	case PageTypeSend, PageTypeReceive, PageTypeHistory:
		ui.navigateTo(PageTypeHome)
	case PageTypeSendDetail:
		ui.navigateTo(PageTypeSend)
	case PageTypeReceiveDetail:
		ui.navigateTo(PageTypeReceive)
	default:
		ui.navigateTo(PageTypeHome)
	}
}

// NavigateToSendDetail 导航到发送详情页
func (ui *MainUI) NavigateToSendDetail() {
	ui.navigateTo(PageTypeSendDetail)
}

// NavigateToReceiveDetail 导航到接收详情页
func (ui *MainUI) NavigateToReceiveDetail() {
	ui.navigateTo(PageTypeReceiveDetail)
}

// GetSendDetailPage 获取发送详情页实例
func (ui *MainUI) GetSendDetailPage() *pages.SendDetailPage {
	return ui.sendDetailPage
}

// GetReceiveDetailPage 获取接收详情页实例
func (ui *MainUI) GetReceiveDetailPage() *pages.ReceiveDetailPage {
	return ui.receiveDetailPage
}

// Close 关闭资源
func (ui *MainUI) Close() {
	if ui.crocManager != nil {
		ui.crocManager.Close()
	}
}
