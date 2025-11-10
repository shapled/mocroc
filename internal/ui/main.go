package ui

import (
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/shapled/mocroc/internal/crocmgr"
	"github.com/shapled/mocroc/internal/storage"
	"github.com/shapled/mocroc/internal/ui/components"
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

	// Croc 管理器和存储
	crocManager    *crocmgr.Manager
	historyStorage *storage.HistoryStorage

	// 公共属性
	currentPage PageType
	topBar      *components.TopBar
	content     *container.Scroll

	// 页面
	homePage          *pages.HomePage
	sendPage          *pages.SendPage
	receivePage       *pages.ReceivePage
	historyPage       *pages.HistoryPage
	sendDetailPage    *pages.SendDetailPage
	receiveDetailPage *pages.ReceiveDetailPage
}

func NewMainUI(a fyne.App, w fyne.Window) fyne.CanvasObject {
	mainUI := &MainUI{
		app:            a,
		window:         w,
		crocManager:    crocmgr.NewManager(),
		historyStorage: storage.NewHistoryStorage(a),
		currentPage:    PageTypeHome,
	}

	// 初始化页面
	mainUI.createPages()
	mainUI.buildMainWindow()

	return container.NewBorder(mainUI.topBar.Container, nil, nil, nil, mainUI.content)
}

func (ui *MainUI) createPages() {
	// 创建后退按钮
	ui.topBar = components.NewTopBar("MoCroc", func() { ui.goBack() })
	ui.topBar.Hide()

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
			if ui.sendPage != nil {
				ui.sendPage.Cancel()
			}
		},
	)

	ui.receiveDetailPage = pages.NewReceiveDetailPage(ui.window,
		func() { ui.navigateTo(PageTypeReceive) },
		func() {
			// 取消接收
			if ui.receivePage != nil {
				ui.receivePage.Cancel()
			}
		},
	)

	// 创建功能页面
	ui.sendPage = pages.NewSendTab(ui.crocManager, ui.window, ui.app)
	ui.receivePage = pages.NewReceiveTab(ui.crocManager, ui.window, ui.historyStorage)
	ui.historyPage = pages.NewHistoryPage(ui.historyStorage)

	// 设置导航回调
	ui.sendPage.SetOnNavigateToDetail(func() {
		// 设置详情页数据
		if ui.sendDetailPage != nil {
			fileName, code, _ := ui.sendPage.GetSendData()
			ui.sendDetailPage.SetFileName(fileName)
			ui.sendDetailPage.SetCode(code)
			ui.sendDetailPage.SetState(pages.SendDetailStateWaiting)
			ui.sendDetailPage.SetStatusMessage("等待接收端连接...")
		}
		ui.NavigateToSendDetail()
	})

	// 设置详情页更新回调
	ui.sendPage.SetOnUpdateDetail(func(state string, progress float64, message string) {
		fyne.Do(func() {
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
	})

	// 设置接收页面导航回调
	ui.receivePage.SetOnNavigateToDetail(func() {
		// 设置详情页数据
		if ui.receiveDetailPage != nil {
			code, savePath := ui.receivePage.GetReceiveData()
			ui.receiveDetailPage.SetFileName("获取文件信息中...")
			ui.receiveDetailPage.SetSenderInfo("发送方 (" + code + ")")
			ui.receiveDetailPage.SetSavePath(savePath)
			ui.receiveDetailPage.SetState(pages.ReceiveDetailStateConnecting)
			ui.receiveDetailPage.SetStatusMessage("正在连接发送方...")
		}
		ui.NavigateToReceiveDetail()
	})

	// 设置接收详情页更新回调
	ui.receivePage.SetOnUpdateDetail(func(state string, progress float64, message string) {
		fyne.Do(func() {
			if ui.receiveDetailPage != nil {
				switch state {
				case "connecting":
					ui.receiveDetailPage.SetState(pages.ReceiveDetailStateConnecting)
					ui.receiveDetailPage.SetStatusMessage(message)
					ui.receiveDetailPage.SetProgress(0.0)
				case "receiving":
					ui.receiveDetailPage.SetState(pages.ReceiveDetailStateReceiving)
					ui.receiveDetailPage.SetStatusMessage(message)
					ui.receiveDetailPage.SetProgress(progress)
				case "completed":
					ui.receiveDetailPage.SetState(pages.ReceiveDetailStateCompleted)
					ui.receiveDetailPage.SetStatusMessage(message)
					ui.receiveDetailPage.SetProgress(1.0)
				case "failed":
					ui.receiveDetailPage.SetState(pages.ReceiveDetailStateFailed)
					ui.receiveDetailPage.SetStatusMessage(message)
					ui.receiveDetailPage.SetProgress(0.0)
				case "cancelled":
					ui.receiveDetailPage.SetState(pages.ReceiveDetailStateCancelled)
					ui.receiveDetailPage.SetStatusMessage(message)
					ui.receiveDetailPage.SetProgress(0.0)
				}
				// 刷新详情页显示
				ui.content.Refresh()
			}
		})
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

func (ui *MainUI) navigateTo(pageType PageType) {
	ui.currentPage = pageType
	ui.updateContent()
}

func (ui *MainUI) updateContent() {
	var content fyne.CanvasObject

	switch ui.currentPage {
	case PageTypeSend:
		ui.topBar.SetTitle("发送")
		ui.topBar.Show()
		content = ui.sendPage.Build()
	case PageTypeSendDetail:
		ui.topBar.SetTitle("发送详情")
		ui.topBar.Show()
		content = ui.sendDetailPage.Build()
	case PageTypeReceive:
		ui.topBar.SetTitle("接收")
		ui.topBar.Show()
		content = ui.receivePage.Build()
	case PageTypeReceiveDetail:
		ui.topBar.SetTitle("接收详情")
		ui.topBar.Show()
		content = ui.receiveDetailPage.Build()
	case PageTypeHistory:
		ui.topBar.SetTitle("历史")
		ui.topBar.Show()
		content = ui.historyPage.Build()
		ui.historyPage.Refresh()
	case PageTypeHome:
		fallthrough
	default:
		ui.topBar.SetTitle("MoCroc")
		ui.topBar.Hide()
		content = ui.homePage.Build()
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
