package pages

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/schollz/croc/v10/src/croc"
	"github.com/shapled/mocroc/internal/crocmgr"
	"github.com/shapled/mocroc/internal/storage"
)

type ReceivePage struct {
	crocManager    *crocmgr.Manager
	window         fyne.Window
	historyStorage *storage.HistoryStorage

	// 回调函数
	onNavigateToDetail func()
	onUpdateDetail   func(state string, progress float64, message string)

	// UI 组件
	scanBtn       *widget.Button
	codeEntry     *widget.Entry
	downloadBtn   *widget.Button
	cancelBtn     *widget.Button
	savePathBtn   *widget.Button
	savePathLabel *widget.Label
	progressBar   *widget.ProgressBar
	statusLabel   *widget.Label

	// 数据
	receiveCode  string
	savePath     string
	isReceiving  bool
	currentItemID string // 当前传输记录的ID

	// 容器
	content fyne.CanvasObject
}

func NewReceiveTab(crocManager *crocmgr.Manager, window fyne.Window, historyStorage *storage.HistoryStorage) *ReceivePage {
	tab := &ReceivePage{
		crocManager:    crocManager,
		window:         window,
		historyStorage: historyStorage,
		savePath:       getDefaultSavePath(),
	}
	tab.createWidgets()
	tab.buildContent()
	tab.content.Refresh()
	return tab
}

func (page *ReceivePage) SetOnNavigateToDetail(callback func()) {
	page.onNavigateToDetail = callback
}

func (page *ReceivePage) SetOnUpdateDetail(callback func(state string, progress float64, message string)) {
	page.onUpdateDetail = callback
}

// GetReceiveData 获取接收数据用于详情页
func (page *ReceivePage) GetReceiveData() (code string, savePath string) {
	return page.receiveCode, page.savePath
}

// GetIsReceiving 获取接收状态
func (page *ReceivePage) GetIsReceiving() bool {
	return page.isReceiving
}

func (page *ReceivePage) createWidgets() {
	// 接收方式选择
	page.scanBtn = widget.NewButtonWithIcon("📷 扫描二维码", theme.SearchIcon(), page.onScanQR)
	page.codeEntry = widget.NewEntry()
	page.codeEntry.SetPlaceHolder("或手动输入接收码")

	// 保存位置
	page.savePathLabel = widget.NewLabel(page.savePath)
	page.savePathBtn = widget.NewButtonWithIcon("选择保存位置", theme.FolderIcon(), page.onSelectSavePath)

	// 下载和取消按钮
	page.downloadBtn = widget.NewButtonWithIcon("开始接收", theme.DownloadIcon(), page.onDownload)
	page.cancelBtn = widget.NewButtonWithIcon("取消接收", theme.CancelIcon(), page.onCancel)
	page.cancelBtn.Hide()

	// 进度显示
	page.progressBar = widget.NewProgressBar()
	page.statusLabel = widget.NewLabel("等待接收码...")
}

func (page *ReceivePage) buildPreReceiveContent() fyne.CanvasObject {
	// 接收码输入区域
	codeSection := container.NewVBox(
		page.scanBtn,
		widget.NewForm(
			&widget.FormItem{Text: "接收码:", Widget: page.codeEntry},
		),
	)

	// 保存位置选择
	saveSection := container.NewHBox(
		page.savePathBtn,
		page.savePathLabel,
	)

	// 操作按钮
	actionSection := container.NewVBox(
		page.downloadBtn,
	)

	// 主要内容
	mainContent := container.NewVBox(
		widget.NewCard("接收方式", "", codeSection),
		widget.NewCard("保存设置", "", saveSection),
		widget.NewCard("操作", "", actionSection),
	)

	// 添加一些垂直间距
	contentWithSpacing := container.NewVBox(
		mainContent,
	)

	return container.NewScroll(contentWithSpacing)
}

func (page *ReceivePage) buildPostReceiveContent() fyne.CanvasObject {
	// 传输状态卡片
	statusCard := widget.NewCard("传输状态", "", container.NewVBox(
		page.progressBar,
		page.statusLabel,
	))

	// 操作按钮
	actionSection := container.NewVBox(
		page.cancelBtn,
	)

	// 主要内容
	mainContent := container.NewVBox(
		statusCard,
		widget.NewCard("操作", "", actionSection),
	)

	return container.NewScroll(mainContent)
}

func (page *ReceivePage) buildContent() {
	if page.isReceiving {
		page.content = page.buildPostReceiveContent()
	} else {
		page.content = page.buildPreReceiveContent()
	}
}

func (page *ReceivePage) Build() fyne.CanvasObject {
	return page.content
}

func (page *ReceivePage) Cancel() error {
	if !page.isReceiving {
		return fmt.Errorf("没有正在进行的接收任务")
	}
	page.onCancel()
	return nil
}

func (page *ReceivePage) refreshDisplay() {
	page.buildContent()
	page.content.Refresh()
}

// 事件处理器
func (page *ReceivePage) onScanQR() {
	if page.isReceiving {
		page.statusLabel.SetText("接收中，无法扫描二维码")
		return
	}
	// TODO: 实现二维码扫描
	page.statusLabel.SetText("二维码扫描功能待实现")
}

func (page *ReceivePage) onSelectSavePath() {
	if page.isReceiving {
		page.statusLabel.SetText("接收中，无法更改保存位置")
		return
	}

	dialog.ShowFolderOpen(func(reader fyne.ListableURI, err error) {
		if err != nil || reader == nil {
			return
		}

		page.savePath = reader.Path()
		page.savePathLabel.SetText(page.savePath)
		page.statusLabel.SetText("保存位置已更新")
	}, page.window)
}

func (page *ReceivePage) onDownload() {
	if page.isReceiving {
		page.statusLabel.SetText("正在接收中，请等待完成")
		return
	}

	code := strings.TrimSpace(page.codeEntry.Text)
	if code == "" {
		page.statusLabel.SetText("请先输入接收码")
		return
	}

	page.receiveCode = code

	// 创建历史记录
	itemID, err := page.createReceiveHistoryItem(code)
	if err != nil {
		page.statusLabel.SetText("创建历史记录失败: " + err.Error())
		return
	}
	page.currentItemID = itemID

	// 先导航到详情页（此时状态还是 Idle，允许导航）
	if page.onNavigateToDetail != nil {
		page.onNavigateToDetail()
	}

	// 然后设置接收状态
	page.isReceiving = true

	// 启动接收协程
	go page.startReceiving()
}

func (page *ReceivePage) onCancel() {
	if !page.isReceiving {
		return
	}

	page.statusLabel.SetText("正在取消接收...")
	page.crocManager.Cancel()

	// 更新历史记录状态为已取消
	page.updateHistoryItemStatus("cancelled")

	// 重置状态
	fyne.Do(func() {
		page.isReceiving = false
		page.refreshDisplay()
		page.progressBar.SetValue(0.0)
		page.statusLabel.SetText("接收已取消")
		page.receiveCode = ""
		page.currentItemID = ""
	})

}

// 辅助函数
func getDefaultSavePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// 如果获取用户主目录失败，使用临时目录
		if runtime.GOOS == "windows" {
			return os.Getenv("TEMP")
		}
		return "/tmp"
	}

	downloads := filepath.Join(home, "Downloads")
	if _, err := os.Stat(downloads); os.IsNotExist(err) {
		// 如果 Downloads 目录不存在，使用主目录
		return home
	}

	return downloads
}

func (page *ReceivePage) startReceiving() {
	startTime := time.Now()

	defer func() {
		fyne.Do(func() {
			page.isReceiving = false
			page.refreshDisplay()
		})
	}()

	// 更新历史记录状态为进行中
	page.updateHistoryItemStatus("in_progress")

	// 通知详情页更新状态
	if page.onUpdateDetail != nil {
		page.onUpdateDetail("connecting", 0.0, "正在连接发送方...")
	}

	// 创建 Croc 选项 - 根据文档中的正确配置
	options := croc.Options{
		IsSender:       false,
		SharedSecret:   page.receiveCode,
		Debug:          false,
		NoPrompt:       true, // 对应命令行的 --yes 参数
		Stdout:         false,
		NoMultiplexing: false,
		HashAlgorithm:  "xxhash",
		Curve:          "p256", // 必须小写，不是 "P-256"
		ZipFolder:      false,
		Exclude:        []string{},
		GitIgnore:      false,
		Overwrite:      false,
	}

	// 接收端必须设置中继服务器配置才能正常工作
	options.RelayAddress = "croc.schollz.com"
	options.RelayPorts = []string{"9009", "9010", "9011", "9012", "9013"}
	options.RelayPassword = "pass123"
	options.OnlyLocal = false
	options.DisableLocal = false

	// 创建 Croc 客户端
	client, err := page.crocManager.CreateCrocClient(options)
	if err != nil {
		fyne.Do(func() {
			page.statusLabel.SetText("创建客户端失败: " + err.Error())
		})
		// 更新历史记录状态为失败
		page.updateHistoryItemStatus("failed")
		// 通知详情页更新状态
		if page.onUpdateDetail != nil {
			page.onUpdateDetail("failed", 0.0, "创建客户端失败: "+err.Error())
		}
		return
	}

	page.crocManager.Log("开始接收文件...")

	// 通知详情页更新状态为接收中
	if page.onUpdateDetail != nil {
		page.onUpdateDetail("receiving", 0.1, "正在接收文件...")
	}

	// 启动接收
	err = client.Receive()
	if err != nil {
		fyne.Do(func() {
			page.statusLabel.SetText("接收失败: " + err.Error())
		})
		page.crocManager.Log("接收失败: " + err.Error())
		// 更新历史记录状态为失败
		page.updateHistoryItemStatus("failed")
		// 通知详情页更新状态
		if page.onUpdateDetail != nil {
			page.onUpdateDetail("failed", 0.0, "接收失败: "+err.Error())
		}
		return
	}

	// 计算传输耗时
	duration := int64(time.Since(startTime).Seconds())

	// 接收完成 - 更新历史记录状态为已完成，并记录耗时
	page.updateHistoryItemCompleted(duration)

	// 通知详情页更新状态为完成
	if page.onUpdateDetail != nil {
		page.onUpdateDetail("completed", 1.0, "接收完成！文件保存在: "+page.savePath)
	}

	fyne.Do(func() {
		page.progressBar.SetValue(1.0)
		page.statusLabel.SetText("接收完成！文件保存在: " + page.savePath)
	})
	page.crocManager.Log("接收完成")

	// 清空当前记录ID
	page.currentItemID = ""
}

// createReceiveHistoryItem 创建接收历史记录
func (page *ReceivePage) createReceiveHistoryItem(code string) (string, error) {
	item := storage.HistoryItem{
		Type:       "receive",
		FileName:   "等待接收文件信息",
		FileSize:   "未知",
		Code:       code,
		Status:     "in_progress",
		Timestamp:  time.Now(),
		Duration:   0,
		ClientInfo: "接收端",
		NumFiles:   0,
	}

	return page.historyStorage.Add(item)
}

// updateHistoryItemStatus 更新历史记录状态
func (page *ReceivePage) updateHistoryItemStatus(status string) {
	if page.currentItemID == "" {
		return
	}

	err := page.historyStorage.Update(page.currentItemID, func(item *storage.HistoryItem) {
		item.Status = status
	})
	if err != nil {
		page.crocManager.Log("更新历史记录状态失败: " + err.Error())
	}
}

// updateHistoryItemCompleted 更新历史记录为完成状态
func (page *ReceivePage) updateHistoryItemCompleted(duration int64) {
	if page.currentItemID == "" {
		return
	}

	err := page.historyStorage.Update(page.currentItemID, func(item *storage.HistoryItem) {
		item.Status = "completed"
		item.Duration = duration
		// 这里可以进一步更新文件信息，但需要更复杂的实现
		// 目前保持基础信息
	})
	if err != nil {
		page.crocManager.Log("更新历史记录完成状态失败: " + err.Error())
	}
}
