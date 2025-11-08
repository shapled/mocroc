package tabs

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/schollz/croc/v10/src/croc"
	"github.com/shapled/mocroc/internal/crocmgr"
	"github.com/shapled/mocroc/internal/types"
)

const (
	sendFileMode = "文件"
	sendTextMode = "文本"
)

type SendTab struct {
	crocManager *crocmgr.Manager
	window      fyne.Window

	// UI 组件
	modeRadio         *widget.RadioGroup
	fileContent       *fyne.Container
	textContent       *fyne.Container
	addFilesBtn       *widget.Button
	textEntry         *widget.Entry
	fileList          *widget.List
	sendBtn           *widget.Button
	cancelBtn         *widget.Button
	codeLabel         *widget.Label
	progressBar       *widget.ProgressBar
	statusLabel       *widget.Label
	advancedCheck     *widget.Check
	disableLocalCheck *widget.Check

	// 配置组件
	preSendCard   *widget.Card
	postSendCard  *widget.Card
	advancedCard  *widget.Card
	compressCheck *widget.Check
	relayEntry    *widget.Entry
	passwordEntry *widget.Entry

	// 数据
	selectedFiles  []string
	sendText       string
	codePhrase     string
	currentMode    string
	isTransferring bool
	isActive       bool

	// 容器
	content fyne.CanvasObject
}

func NewSendTab(crocManager *crocmgr.Manager, window fyne.Window) *SendTab {
	rand.Seed(time.Now().UnixNano())
	tab := &SendTab{
		crocManager: crocManager,
		window:      window,
		currentMode: sendFileMode,
	}
	tab.createWidgets()
	tab.buildContent()
	return tab
}

func (tab *SendTab) createWidgets() {
	// --- File Widgets ---
	tab.fileList = widget.NewList(
		func() int { return len(tab.selectedFiles) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel(""), widget.NewButtonWithIcon("", theme.DeleteIcon(), nil))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			label := c.Objects[0].(*widget.Label)
			btn := c.Objects[1].(*widget.Button)
			if id < len(tab.selectedFiles) {
				label.SetText(filepath.Base(tab.selectedFiles[id]))
			}
			btn.OnTapped = func() { tab.deleteFile(id) }
		},
	)
	tab.addFilesBtn = widget.NewButtonWithIcon("选择文件或文件夹", theme.FileIcon(), tab.onAddFiles)

	// --- Text Widgets ---
	tab.textEntry = widget.NewMultiLineEntry()
	tab.textEntry.SetPlaceHolder("输入要发送的文本内容...")
	tab.textEntry.OnChanged = func(s string) {
		tab.sendText = s
		tab.updateSendButton()
	}

	// --- Common Widgets ---
	tab.sendBtn = widget.NewButtonWithIcon("开始发送", theme.MailSendIcon(), tab.onSend)
	tab.cancelBtn = widget.NewButtonWithIcon("取消发送", theme.CancelIcon(), tab.onCancel)
	tab.cancelBtn.Hide()
	tab.sendBtn.Disable()

	tab.codeLabel = widget.NewLabel("等待生成接收码...")
	tab.progressBar = widget.NewProgressBar()
	tab.statusLabel = widget.NewLabel("准备就绪")

	// --- Advanced Options ---
	tab.disableLocalCheck = widget.NewCheck("禁用本地传输", nil)
	tab.compressCheck = widget.NewCheck("自动压缩文件夹", nil)
	tab.relayEntry = widget.NewEntry()
	tab.relayEntry.SetText("croc.schollz.com")
	tab.passwordEntry = widget.NewPasswordEntry()
	tab.passwordEntry.SetText("pass123")

	relayForm := widget.NewForm(
		&widget.FormItem{Text: "中继地址:", Widget: tab.relayEntry},
		&widget.FormItem{Text: "密码:", Widget: tab.passwordEntry},
	)

	tab.advancedCard = widget.NewCard("", "", container.NewVBox(
		tab.compressCheck,
		tab.disableLocalCheck,
		relayForm,
	))
	tab.advancedCard.Hide()

	tab.advancedCheck = widget.NewCheck("高级选项", func(checked bool) {
		if checked {
			tab.advancedCard.Show()
		} else {
			tab.advancedCard.Hide()
		}
	})

	// --- Mode Selection (at the end) ---
	tab.modeRadio = widget.NewRadioGroup([]string{sendFileMode, sendTextMode}, func(selected string) {
		tab.currentMode = selected
		tab.updateSendModeUI()
		tab.updateSendButton()
	})
}

func (tab *SendTab) buildContent() {
	// --- File Content Area ---
	tab.fileContent = container.NewVBox(
		tab.addFilesBtn,
		widget.NewLabel("已选择的文件:"),
		tab.fileList,
	)

	// --- Text Content Area ---
	tab.textContent = container.NewVBox(
		widget.NewLabel("输入文本:"),
		tab.textEntry,
	)
	tab.textContent.Hide() // Initially hidden

	// --- Pre-Send Card ---
	tab.preSendCard = widget.NewCard("发送前设置", "", container.NewVBox(
		widget.NewCard("选择模式", "", tab.modeRadio),
		tab.fileContent,
		tab.textContent,
		tab.advancedCheck,
		tab.advancedCard,
		tab.sendBtn,
	))

	// --- Post-Send Card ---
	qrSection := container.NewVBox(
		widget.NewLabel("接收码:"),
		tab.codeLabel,
		widget.NewButton("显示二维码", tab.onShowQRCode),
	)

	tab.postSendCard = widget.NewCard("发送中", "", container.NewVBox(
		widget.NewCard("接收信息", "", qrSection),
		container.NewHBox(tab.cancelBtn),
		widget.NewSeparator(),
		widget.NewCard("传输状态", "", container.NewVBox(
			tab.progressBar,
			tab.statusLabel,
		)),
	))
	tab.postSendCard.Hide()

	// --- Final Layout ---
	tab.content = container.NewVScroll(container.NewVBox(
		tab.preSendCard,
		tab.postSendCard,
	))

	// Set initial state after all content is built
	tab.modeRadio.SetSelected(sendFileMode)
}

func (tab *SendTab) updateSendModeUI() {
	if tab.currentMode == sendFileMode {
		tab.fileContent.Show()
		tab.textContent.Hide()
		tab.compressCheck.Show()
	} else {
		tab.fileContent.Hide()
		tab.textContent.Show()
		tab.compressCheck.Hide()
	}
}

func (tab *SendTab) Build() fyne.CanvasObject {
	return tab.content
}

func (tab *SendTab) GetState() types.TabState {
	if tab.isTransferring {
		return types.TabStateSending
	}
	return types.TabStateIdle
}

func (tab *SendTab) IsActive() bool {
	return tab.isActive
}

func (tab *SendTab) SetActive(active bool) {
	tab.isActive = active
}

func (tab *SendTab) Cancel() error {
	if !tab.isTransferring {
		return fmt.Errorf("没有正在进行的发送任务")
	}
	tab.onCancel()
	return nil
}

// --- Event Handlers ---

func (tab *SendTab) onAddFiles() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		path := reader.URI().Path()
		tab.selectedFiles = append(tab.selectedFiles, path)
		fyne.Do(func() {
			tab.fileList.Refresh()
			tab.updateSendButton()
			tab.statusLabel.SetText(fmt.Sprintf("已添加 %d 个文件", len(tab.selectedFiles)))
		})
	}, tab.window)
}

func (tab *SendTab) onSend() {
	if (tab.currentMode == sendFileMode && len(tab.selectedFiles) == 0) ||
		(tab.currentMode == sendTextMode && tab.sendText == "") {
		tab.statusLabel.SetText("请先选择文件或输入文本")
		return
	}

	tab.isTransferring = true
	tab.preSendCard.Hide()
	tab.postSendCard.Show()
	tab.progressBar.SetValue(0.0)
	tab.statusLabel.SetText("正在初始化...")

	code, err := tab.generateCode()
	if err != nil {
		fyne.Do(func() {
			tab.statusLabel.SetText("生成接收码失败: " + err.Error())
			tab.resetSendState()
		})
		return
	}
	tab.codePhrase = code
	fyne.Do(func() {
		tab.codeLabel.SetText("接收码: " + code)
	})

	go tab.startSending()
}

func (tab *SendTab) onCancel() {
	if !tab.isTransferring {
		return
	}
	tab.statusLabel.SetText("正在取消发送...")
	tab.crocManager.Cancel()
	tab.resetSendState()
	fyne.Do(func() {
		tab.statusLabel.SetText("发送已取消")
	})
}

func (tab *SendTab) resetSendState() {
	tab.isTransferring = false
	fyne.Do(func() {
		tab.preSendCard.Show()
		tab.postSendCard.Hide()
		tab.progressBar.SetValue(0.0)
		tab.codePhrase = ""
		tab.codeLabel.SetText("等待生成接收码...")
	})
}

func (tab *SendTab) deleteFile(index int) {
	if index < 0 || index >= len(tab.selectedFiles) {
		return
	}
	tab.selectedFiles = append(tab.selectedFiles[:index], tab.selectedFiles[index+1:]...)
	fyne.Do(func() {
		tab.fileList.Refresh()
		tab.updateSendButton()
		tab.statusLabel.SetText(fmt.Sprintf("已删除文件，剩余 %d 个", len(tab.selectedFiles)))
	})
}

func (tab *SendTab) onShowQRCode() {
	// Placeholder
}

// --- Helper Functions ---

func (tab *SendTab) updateSendButton() {
	enabled := false
	if !tab.isTransferring {
		if tab.currentMode == sendFileMode && len(tab.selectedFiles) > 0 {
			enabled = true
		} else if tab.currentMode == sendTextMode && strings.TrimSpace(tab.sendText) != "" {
			enabled = true
		}
	}

	if enabled {
		tab.sendBtn.Enable()
	} else {
		tab.sendBtn.Disable()
	}
}

func (tab *SendTab) generateCode() (string, error) {
	adjectives := []string{"red", "blue", "green", "yellow", "orange", "purple", "pink", "brown"}
	animals := []string{"cat", "dog", "frog", "bird", "fish", "lion", "tiger", "bear"}
	adj := adjectives[rand.Intn(len(adjectives))]
	ani := animals[rand.Intn(len(animals))]
	num := rand.Intn(9000) + 1000
	return fmt.Sprintf("%s-%s-%d", adj, ani, num), nil
}

func (tab *SendTab) startSending() {
	defer tab.resetSendState()

	var sendFiles []string
	var err error

	if tab.currentMode == sendTextMode {
		var tmpFile *os.File
		tmpFile, err = tab.createTextFile(tab.sendText)
		if err != nil {
			fyne.Do(func() { tab.statusLabel.SetText("文本发送失败: " + err.Error()) })
			return
		}
		defer os.Remove(tmpFile.Name())
		sendFiles = []string{tmpFile.Name()}
	} else {
		sendFiles = make([]string, len(tab.selectedFiles))
		copy(sendFiles, tab.selectedFiles)
	}

	options := tab.buildCrocOptions()
	client, err := tab.crocManager.CreateCrocClient(options)
	if err != nil {
		fyne.Do(func() { tab.statusLabel.SetText("创建客户端失败: " + err.Error()) })
		return
	}

	fyne.Do(func() { tab.statusLabel.SetText("等待接收方连接...") })

	filesInfo, emptyFolders, totalNumberFolders, err := croc.GetFilesInfo(sendFiles, tab.compressCheck.Checked, false, []string{})
	if err != nil {
		fyne.Do(func() { tab.statusLabel.SetText("获取文件信息失败: " + err.Error()) })
		return
	}

	err = client.Send(filesInfo, emptyFolders, totalNumberFolders)
	if err != nil {
		fyne.Do(func() { tab.statusLabel.SetText("发送失败: " + err.Error()) })
		return
	}

	go tab.monitorProgress()

	fyne.Do(func() {
		if tab.currentMode == sendTextMode {
			tab.statusLabel.SetText("正在发送文本...")
		} else {
			tab.statusLabel.SetText("正在发送文件...")
		}
	})
}

func (tab *SendTab) monitorProgress() {
	ctx := tab.crocManager.GetContext()
	steps := []float64{0.1, 0.3, 0.5, 0.7, 0.9}
	for _, progress := range steps {
		select {
		case <-ctx.Done():
			return // Canceled
		case <-time.After(500 * time.Millisecond):
			fyne.Do(func() { tab.progressBar.SetValue(progress) })
		}
	}
	fyne.Do(func() {
		tab.progressBar.SetValue(1.0)
		tab.statusLabel.SetText("发送完成！")
	})
}

func (tab *SendTab) buildCrocOptions() croc.Options {
	return croc.Options{
		IsSender:       true,
		SharedSecret:   tab.codePhrase,
		Debug:          false,
		NoPrompt:       true,
		Stdout:         false,
		HashAlgorithm:  "xxhash",
		Curve:          "p256",
		ZipFolder:      tab.compressCheck.Checked,
		OnlyLocal:      false,
		DisableLocal:   tab.disableLocalCheck.Checked,
		RelayAddress:   tab.relayEntry.Text,
		RelayPorts:     []string{"9009", "9010", "9011", "9012", "9013"},
		RelayPassword:  tab.passwordEntry.Text,
	}
}

func (tab *SendTab) createTextFile(textContent string) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "mocroc-text-*.txt")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	if _, err := tmpFile.WriteString(textContent); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("写入临时文件失败: %w", err)
	}
	tmpFile.Close()
	return tmpFile, nil
}
