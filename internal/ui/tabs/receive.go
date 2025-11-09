package tabs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/schollz/croc/v10/src/croc"
	"github.com/shapled/mocroc/internal/crocmgr"
	"github.com/shapled/mocroc/internal/types"
)

type ReceiveTab struct {
	crocManager *crocmgr.Manager
	window      fyne.Window

	// 回调函数
	onNavigateToDetail func()

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
		receiveCode string
		savePath    string
		isReceiving bool
		isActive    bool

	// 容器
	content fyne.CanvasObject
}

func NewReceiveTab(crocManager *crocmgr.Manager, window fyne.Window) *ReceiveTab {
	tab := &ReceiveTab{
		crocManager: crocManager,
		window:      window,
		savePath:    getDefaultSavePath(),
	}
	tab.createWidgets()
	tab.buildContent()
	tab.content.Refresh()
	return tab
}

func (tab *ReceiveTab) SetOnNavigateToDetail(callback func()) {
	tab.onNavigateToDetail = callback
}

// GetReceiveData 获取接收数据用于详情页
func (tab *ReceiveTab) GetReceiveData() (code string, savePath string) {
	return tab.receiveCode, tab.savePath
}

// GetIsReceiving 获取接收状态
func (tab *ReceiveTab) GetIsReceiving() bool {
	return tab.isReceiving
}

func (tab *ReceiveTab) createWidgets() {
	// 接收方式选择
	tab.scanBtn = widget.NewButtonWithIcon("📷 扫描二维码", theme.SearchIcon(), tab.onScanQR)
	tab.codeEntry = widget.NewEntry()
	tab.codeEntry.SetPlaceHolder("或手动输入接收码")

	// 保存位置
	tab.savePathLabel = widget.NewLabel(tab.savePath)
	tab.savePathBtn = widget.NewButtonWithIcon("选择保存位置", theme.FolderIcon(), tab.onSelectSavePath)

	// 下载和取消按钮
	tab.downloadBtn = widget.NewButtonWithIcon("开始接收", theme.DownloadIcon(), tab.onDownload)
	tab.cancelBtn = widget.NewButtonWithIcon("取消接收", theme.CancelIcon(), tab.onCancel)
	tab.cancelBtn.Hide()

	// 进度显示
	tab.progressBar = widget.NewProgressBar()
	tab.statusLabel = widget.NewLabel("等待接收码...")
}

func (tab *ReceiveTab) buildPreReceiveContent() fyne.CanvasObject {
	// 接收码输入区域
	codeSection := container.NewVBox(
		tab.scanBtn,
		widget.NewForm(
			&widget.FormItem{Text: "接收码:", Widget: tab.codeEntry},
		),
	)

	// 保存位置选择
	saveSection := container.NewHBox(
		tab.savePathBtn,
		tab.savePathLabel,
	)

	// 操作按钮
	actionSection := container.NewVBox(
		tab.downloadBtn,
	)

	// 主要内容
	mainContent := container.NewVBox(
		widget.NewCard("接收方式", "", codeSection),
		widget.NewCard("保存设置", "", saveSection),
		widget.NewCard("操作", "", actionSection),
	)

	// 添加一些垂直间距
	contentWithSpacing := container.NewVBox(
		widget.NewLabel(""), // 顶部间距
		mainContent,
		widget.NewLabel(""), // 底部间距
	)

	return container.NewScroll(contentWithSpacing)
}

func (tab *ReceiveTab) buildPostReceiveContent() fyne.CanvasObject {
	// 传输状态卡片
	statusCard := widget.NewCard("传输状态", "", container.NewVBox(
		tab.progressBar,
		tab.statusLabel,
	))

	// 操作按钮
	actionSection := container.NewVBox(
		tab.cancelBtn,
	)

	// 主要内容
	mainContent := container.NewVBox(
		widget.NewLabel(""), // 顶部间距
		statusCard,
		widget.NewCard("操作", "", actionSection),
		widget.NewLabel(""), // 底部间距
	)

	return container.NewScroll(mainContent)
}

func (tab *ReceiveTab) buildContent() {
	if tab.isReceiving {
		tab.content = tab.buildPostReceiveContent()
	} else {
		tab.content = tab.buildPreReceiveContent()
	}
}

func (tab *ReceiveTab) Build() fyne.CanvasObject {
	return tab.content
}

// TabInterface 实现
func (tab *ReceiveTab) GetState() types.TabState {
	if tab.isReceiving {
		return types.TabStateReceiving
	}
	return types.TabStateIdle
}

func (tab *ReceiveTab) Cancel() error {
	if !tab.isReceiving {
		return fmt.Errorf("没有正在进行的接收任务")
	}
	tab.onCancel()
	return nil
}

func (tab *ReceiveTab) IsActive() bool {
	return tab.isActive
}

func (tab *ReceiveTab) SetActive(active bool) {
	tab.isActive = active
	if active {
		tab.refreshDisplay()
	}
}

func (tab *ReceiveTab) refreshDisplay() {
	tab.buildContent()
	tab.content.Refresh()
}

// 事件处理器
func (tab *ReceiveTab) onScanQR() {
	if tab.isReceiving {
		tab.statusLabel.SetText("接收中，无法扫描二维码")
		return
	}
	// TODO: 实现二维码扫描
	tab.statusLabel.SetText("二维码扫描功能待实现")
}

func (tab *ReceiveTab) onSelectSavePath() {
	if tab.isReceiving {
		tab.statusLabel.SetText("接收中，无法更改保存位置")
		return
	}

	dialog.ShowFolderOpen(func(reader fyne.ListableURI, err error) {
		if err != nil || reader == nil {
			return
		}

		tab.savePath = reader.Path()
		tab.savePathLabel.SetText(tab.savePath)
		tab.statusLabel.SetText("保存位置已更新")
	}, tab.window)
}

func (tab *ReceiveTab) onDownload() {
	if tab.isReceiving {
		tab.statusLabel.SetText("正在接收中，请等待完成")
		return
	}

	code := strings.TrimSpace(tab.codeEntry.Text)
	if code == "" {
		tab.statusLabel.SetText("请先输入接收码")
		return
	}

	tab.receiveCode = code

	// 先导航到详情页（此时状态还是 Idle，允许导航）
	if tab.onNavigateToDetail != nil {
		tab.onNavigateToDetail()
	}

	// 然后设置接收状态
	tab.isReceiving = true

	// 启动接收协程
	go tab.startReceiving()
}

func (tab *ReceiveTab) onCancel() {
	if !tab.isReceiving {
		return
	}

	tab.statusLabel.SetText("正在取消接收...")
	tab.crocManager.Cancel()

	// 重置状态
	fyne.Do(func() {
		tab.isReceiving = false
		tab.refreshDisplay()
		tab.progressBar.SetValue(0.0)
		tab.statusLabel.SetText("接收已取消")
		tab.receiveCode = ""
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

func (tab *ReceiveTab) startReceiving() {
	defer func() {
		fyne.Do(func() {
			tab.isReceiving = false
			tab.refreshDisplay()
		})
	}()

	// 创建 Croc 选项 - 根据文档中的正确配置
	options := croc.Options{
		IsSender:       false,
		SharedSecret:   tab.receiveCode,
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
	client, err := tab.crocManager.CreateCrocClient(options)
	if err != nil {
		fyne.Do(func() {
			tab.statusLabel.SetText("创建客户端失败: " + err.Error())
		})
		return
	}

	tab.crocManager.Log("开始接收文件...")
	fyne.Do(func() {
		tab.statusLabel.SetText("正在连接发送方...")
	})

	// 启动接收
	err = client.Receive()
	if err != nil {
		fyne.Do(func() {
			tab.statusLabel.SetText("接收失败: " + err.Error())
		})
		tab.crocManager.Log("接收失败: " + err.Error())
		return
	}

	// 接收完成
	fyne.Do(func() {
		tab.progressBar.SetValue(1.0)
		tab.statusLabel.SetText("接收完成！文件保存在: " + tab.savePath)
	})
	tab.crocManager.Log("接收完成")
}

func (tab *ReceiveTab) simulateProgress() {
	steps := 10
	for i := 0; i <= steps; i++ {
		select {
		case <-tab.crocManager.GetContext().Done():
			return
		default:
			progress := float64(i) / float64(steps)
			tab.progressBar.SetValue(progress)

			if i < steps {
				tab.statusLabel.SetText(fmt.Sprintf("接收进度: %.1f%%", progress*100))
			}

			// 模拟接收延迟
			// time.Sleep(time.Millisecond * 300)
		}
	}
}
