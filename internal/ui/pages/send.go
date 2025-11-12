package pages

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
	"github.com/shapled/mocroc/internal/storage"
)

const (
	sendFileMode = "文件"
	sendTextMode = "文本"
)

type SendPage struct {
	crocManager *crocmgr.Manager
	window      fyne.Window
	storage     *storage.HistoryStorage

	// 回调函数
	onNavigateToDetail func()
	onUpdateDetail     func(state string, progress float64, message string)

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
	isCancelled    bool // 取消标志

	// 历史记录相关
	currentHistoryID string
	sendStartTime    time.Time

	// 容器
	content fyne.CanvasObject
}

func NewSendTab(crocManager *crocmgr.Manager, window fyne.Window, a fyne.App) *SendPage {
	rand.Seed(time.Now().UnixNano())
	tab := &SendPage{
		crocManager: crocManager,
		window:      window,
		storage:     storage.NewHistoryStorage(a),
		currentMode: sendFileMode,
	}
	tab.createWidgets()
	tab.buildContent()
	return tab
}

func (page *SendPage) SetOnNavigateToDetail(callback func()) {
	page.onNavigateToDetail = callback
}

func (page *SendPage) SetOnUpdateDetail(callback func(state string, progress float64, message string)) {
	page.onUpdateDetail = callback
}

// GetSendData 获取发送数据用于详情页
func (page *SendPage) GetSendData() (fileName string, code string, isText bool) {
	if page.currentMode == sendTextMode {
		fileName = "文本内容"
		isText = true
	} else if len(page.selectedFiles) > 0 {
		fileName = filepath.Base(page.selectedFiles[0])
		isText = false
	}
	code = page.codePhrase
	return
}

// GetIsTransferring 获取传输状态
func (page *SendPage) GetIsTransferring() bool {
	return page.isTransferring
}

func (page *SendPage) createWidgets() {
	// --- File Widgets ---
	page.fileList = widget.NewList(
		func() int { return len(page.selectedFiles) },
		func() fyne.CanvasObject {
			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			deleteBtn.Resize(fyne.NewSize(44, 44)) // 移动端最小触摸区域
			return container.NewHBox(widget.NewLabel(""), deleteBtn)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			label := c.Objects[0].(*widget.Label)
			btn := c.Objects[1].(*widget.Button)
			if id < len(page.selectedFiles) {
				label.SetText(filepath.Base(page.selectedFiles[id]))
			}
			btn.OnTapped = func() { page.deleteFile(id) }
		},
	)
	page.fileList.Resize(fyne.NewSize(400, 200)) // 设置最小高度
	page.addFilesBtn = widget.NewButtonWithIcon("选择文件或文件夹", theme.FileIcon(), page.onAddFiles)
	page.addFilesBtn.Resize(fyne.NewSize(280, 56)) // 移动端标准尺寸
	page.addFilesBtn.Importance = widget.HighImportance

	// --- Text Widgets ---
	page.textEntry = widget.NewMultiLineEntry()
	page.textEntry.SetPlaceHolder("输入要发送的文本内容...")
	page.textEntry.Resize(fyne.NewSize(400, 150)) // 设置一个合适的高度
	page.textEntry.OnChanged = func(s string) {
		page.sendText = s
		page.updateSendButton()
	}

	// --- Common Widgets ---
	page.sendBtn = widget.NewButtonWithIcon("开始发送", theme.MailSendIcon(), page.onSend)
	page.sendBtn.Resize(fyne.NewSize(280, 56)) // 移动端标准尺寸
	page.sendBtn.Importance = widget.HighImportance

	page.cancelBtn = widget.NewButtonWithIcon("取消发送", theme.CancelIcon(), page.onCancel)
	page.cancelBtn.Resize(fyne.NewSize(280, 56)) // 移动端标准尺寸
	page.cancelBtn.Importance = widget.MediumImportance
	page.cancelBtn.Hide()
	page.sendBtn.Disable()

	page.codeLabel = widget.NewLabel("等待生成接收码...")
	page.progressBar = widget.NewProgressBar()
	page.statusLabel = widget.NewLabel("准备就绪")

	// --- Advanced Options ---
	page.disableLocalCheck = widget.NewCheck("禁用本地传输", nil)
	page.compressCheck = widget.NewCheck("自动压缩文件夹", nil)
	page.relayEntry = widget.NewEntry()
	page.relayEntry.SetText("croc.schollz.com")
	page.passwordEntry = widget.NewPasswordEntry()
	page.passwordEntry.SetText("pass123")

	relayForm := widget.NewForm(
		&widget.FormItem{Text: "中继地址:", Widget: page.relayEntry},
		&widget.FormItem{Text: "密码:", Widget: page.passwordEntry},
	)

	page.advancedCard = widget.NewCard("", "", container.NewVBox(
		page.compressCheck,
		page.disableLocalCheck,
		relayForm,
	))
	page.advancedCard.Hide()

	page.advancedCheck = widget.NewCheck("高级选项", func(checked bool) {
		if checked {
			page.advancedCard.Show()
		} else {
			page.advancedCard.Hide()
		}
	})

	// --- Mode Selection (at the end) ---
	page.modeRadio = widget.NewRadioGroup([]string{sendFileMode, sendTextMode}, func(selected string) {
		page.currentMode = selected
		page.updateSendModeUI()
		page.updateSendButton()
	})
}

func (page *SendPage) buildContent() {
	// --- File Content Area ---
	page.fileContent = container.NewVBox(
		page.addFilesBtn,
		widget.NewLabel(""), // 间距
		widget.NewLabelWithStyle("已选择的文件:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""), // 小间距
		page.fileList,
	)

	// --- Text Content Area ---
	page.textContent = container.NewVBox(
		widget.NewLabelWithStyle("输入文本:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""), // 小间距
		page.textEntry,
	)
	page.textContent.Hide() // Initially hidden

	// --- Pre-Send Card ---
	page.preSendCard = widget.NewCard("发送设置", "", container.NewPadded(container.NewVBox(
		widget.NewCard("传输模式", "", container.NewPadded(page.modeRadio)),
		widget.NewLabel(""), // 间距
		page.fileContent,
		page.textContent,
		widget.NewLabel(""), // 间距
		page.advancedCheck,
		widget.NewLabel(""), // 小间距
		page.advancedCard,
		widget.NewLabel(""), // 大间距
		container.NewCenter(page.sendBtn),
	)))

	// --- Post-Send Card ---
	qrSection := container.NewVBox(
		widget.NewLabelWithStyle("接收码:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""), // 小间距
		page.codeLabel,
		widget.NewLabel(""), // 间距
		widget.NewButtonWithIcon("显示二维码", theme.InfoIcon(), page.onShowQRCode),
	)

	// 发送状态图标
	sendIcon := widget.NewLabelWithStyle("📤", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	sendIconContainer := container.NewCenter(sendIcon)

	// 增强的传输状态显示
	progressDetails := container.NewVBox(
		page.statusLabel,
		widget.NewLabel(""), // 间距
		page.progressBar,
	)

	page.postSendCard = widget.NewCard("发送中", "", container.NewPadded(container.NewVBox(
		sendIconContainer,
		widget.NewLabel(""), // 间距
		widget.NewCard("接收信息", "", container.NewPadded(qrSection)),
		widget.NewLabel(""), // 间距
		container.NewCenter(page.cancelBtn),
		widget.NewLabel(""), // 间距
		widget.NewSeparator(),
		widget.NewLabel(""), // 间距
		widget.NewCard("传输状态", "", container.NewPadded(progressDetails)),
	)))
	page.postSendCard.Hide()

	// --- Final Layout ---
	mainContent := container.NewVBox(
		page.preSendCard,
		page.postSendCard,
	)

	// 添加内边距并使用滚动容器
	paddedContent := container.NewPadded(mainContent)
	page.content = container.NewScroll(paddedContent)

	// Set initial state after all content is built
	page.modeRadio.SetSelected(sendFileMode)
}

func (page *SendPage) updateSendModeUI() {
	if page.currentMode == sendFileMode {
		page.fileContent.Show()
		page.textContent.Hide()
		page.compressCheck.Show()
	} else {
		page.fileContent.Hide()
		page.textContent.Show()
		page.compressCheck.Hide()
	}
}

func (page *SendPage) Build() fyne.CanvasObject {
	return page.content
}

func (page *SendPage) Cancel() error {
	if !page.isTransferring {
		return fmt.Errorf("没有正在进行的发送任务")
	}
	page.onCancel()
	return nil
}

// --- Event Handlers ---

func (page *SendPage) onAddFiles() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		path := reader.URI().Path()
		page.selectedFiles = append(page.selectedFiles, path)
		fyne.Do(func() {
			page.fileList.Refresh()
			page.updateSendButton()
			page.statusLabel.SetText(fmt.Sprintf("已添加 %d 个文件", len(page.selectedFiles)))
		})
	}, page.window)
}

func (page *SendPage) onSend() {
	if page.currentMode == sendFileMode && len(page.selectedFiles) == 0 {
		page.statusLabel.SetText("请先选择文件")
		return
	}

	if page.currentMode == sendTextMode && page.sendText == "" {
		page.statusLabel.SetText("请先输入文本")
		return
	}

	// 生成接收码
	code, err := page.generateCode()
	if err != nil {
		page.statusLabel.SetText("生成接收码失败: " + err.Error())
		return
	}
	page.codePhrase = code

	// 创建历史记录
	historyItem := page.createHistoryItem(code)
	if historyItem == nil {
		page.statusLabel.SetText("创建历史记录失败")
		return
	}

	// 先导航到详情页（此时状态还是 Idle，允许导航）
	if page.onNavigateToDetail != nil {
		page.onNavigateToDetail()
	}

	// 然后设置传输状态
	page.isTransferring = true
	page.sendStartTime = time.Now()

	// 在后台开始发送
	go page.startSending()
}

func (page *SendPage) onCancel() {
	if !page.isTransferring {
		return
	}
	page.statusLabel.SetText("正在取消发送...")
	// 更新详情页状态为取消中
	if page.onUpdateDetail != nil {
		page.onUpdateDetail("cancelled", 0.0, "正在取消发送...")
	}

	// 设置取消标志
	page.isCancelled = true

	// 取消 croc 管理器的 context
	page.crocManager.Cancel()

	// 不立即重置状态，等待发送进程检测到取消信号
	fyne.Do(func() {
		page.statusLabel.SetText("发送已取消")
		// 更新详情页状态为已取消
		if page.onUpdateDetail != nil {
			page.onUpdateDetail("cancelled", 0.0, "发送已取消")
		}
	})
}

func (page *SendPage) resetSendState() {
	page.isTransferring = false
	page.isCancelled = false // 重置取消标志
	fyne.Do(func() {
		page.preSendCard.Show()
		page.postSendCard.Hide()
		page.progressBar.SetValue(0.0)
		page.codePhrase = ""
		page.codeLabel.SetText("等待生成接收码...")
	})
}

func (page *SendPage) deleteFile(index int) {
	if index < 0 || index >= len(page.selectedFiles) {
		return
	}
	page.selectedFiles = append(page.selectedFiles[:index], page.selectedFiles[index+1:]...)
	fyne.Do(func() {
		page.fileList.Refresh()
		page.updateSendButton()
		page.statusLabel.SetText(fmt.Sprintf("已删除文件，剩余 %d 个", len(page.selectedFiles)))
	})
}

func (page *SendPage) onShowQRCode() {
	// Placeholder
}

// --- Helper Functions ---

func (page *SendPage) updateSendButton() {
	enabled := false
	if !page.isTransferring {
		if page.currentMode == sendFileMode && len(page.selectedFiles) > 0 {
			enabled = true
		} else if page.currentMode == sendTextMode && strings.TrimSpace(page.sendText) != "" {
			enabled = true
		}
	}

	if enabled {
		page.sendBtn.Enable()
	} else {
		page.sendBtn.Disable()
	}
}

func (page *SendPage) generateCode() (string, error) {
	// 英语词根列表 - 第一部分 (前缀/形容词)
	wordRoots1 := []string{
		"act", "ask", "big", "bold", "bright", "calm", "clear", "cool", "dark", "deep",
		"easy", "fast", "fine", "flat", "free", "full", "good", "grand", "great", "green",
		"hard", "high", "honest", "hot", "huge", "kind", "large", "late", "light", "long",
		"loud", "low", "mad", "main", "new", "nice", "old", "open", "plain", "pure",
		"quick", "quiet", "rare", "real", "rich", "round", "safe", "sharp", "slow", "soft",
		"sore", "square", "star", "still", "sweet", "thick", "thin", "tight", "true", "vast",
		"warm", "weak", "white", "wild", "wise", "young",
	}

	// 英语词根列表 - 第二部分 (名词/动作)
	wordRoots2 := []string{
		"art", "ball", "band", "bank", "base", "bell", "bird", "boat", "body", "book",
		"box", "boy", "bug", "camp", "car", "card", "care", "case", "cat", "chair",
		"chance", "change", "charge", "city", "class", "cloud", "coat", "code", "coin", "come",
		"cook", "copper", "copy", "corn", "cost", "cottage", "cotton", "count", "cover", "crack",
		"cream", "crop", "cross", "crowd", "crown", "cry", "cup", "curve", "cut", "dance",
		"day", "deal", "deer", "design", "door", "draw", "dream", "dress", "drop", "drum",
		"duck", "dust", "earth", "edge", "engine", "event", "face", "fact", "fall", "family",
		"farm", "father", "fear", "field", "fire", "fish", "flag", "flower", "fly", "forest",
		"form", "fountain", "fox", "friend", "fruit", "game", "garden", "gate", "giant", "gift",
		"girl", "glass", "glove", "gold", "grass", "group", "guide", "hair", "hand", "head",
		"heart", "hill", "history", "home", "hope", "horn", "horse", "hour", "house", "hunter",
		"iron", "island", "jack", "jam", "jar", "jet", "job", "join", "judge", "key",
		"kick", "king", "kiss", "kite", "knife", "lake", "lamp", "land", "language", "leaf",
		"leg", "letter", "life", "light", "line", "lion", "lock", "look", "love", "machine",
		"man", "map", "mark", "mask", "match", "meal", "meat", "milk", "mind", "mine",
		"minute", "mirror", "money", "moon", "morning", "mother", "mountain", "mouth", "music", "name",
		"nation", "nature", "nerve", "news", "night", "noise", "north", "nose", "note", "number",
		"ocean", "offer", "office", "orange", "order", "page", "paint", "paper", "park", "part",
		"pen", "pencil", "person", "picture", "pie", "pilot", "pipe", "place", "plane", "plant",
		"plate", "play", "point", "pond", "post", "pot", "price", "prince", "prison", "problem",
		"process", "produce", "queen", "question", "rain", "range", "rate", "ray", "reason", "record",
		"rest", "rice", "ring", "river", "road", "rock", "roll", "roof", "room", "root",
		"rose", "rule", "salt", "sand", "scale", "school", "science", "sea", "seat", "seed",
		"serve", "shade", "shake", "shape", "share", "sheep", "sheet", "ship", "shirt", "shoe",
		"shop", "show", "side", "sign", "silk", "silver", "sing", "size", "skin", "skirt",
		"sky", "sleep", "slave", "snow", "soap", "soldier", "son", "song", "sort", "sound",
		"south", "space", "spare", "speak", "spring", "square", "stamp", "star", "state", "steam",
		"steel", "step", "stick", "stone", "stop", "store", "storm", "story", "street", "study",
		"substance", "sugar", "summer", "support", "surprise", "system", "table", "tail", "teacher", "team",
		"teeth", "temperature", "test", "text", "than", "that", "theft", "theory", "there", "thick",
		"thing", "thought", "thread", "thrill", "throat", "thumb", "thunder", "ticket", "time", "tin",
		"tire", "title", "today", "together", "tomorrow", "tone", "tongue", "tooth", "top", "touch",
		"tower", "town", "trade", "train", "transport", "tray", "tree", "trick", "trip", "trouble",
		"trousers", "truck", "turn", "twist", "umbrella", "uncle", "under", "unit", "value", "verse",
		"vessel", "view", "voice", "walk", "wall", "war", "wash", "watch", "water", "wave",
		"weather", "week", "weight", "west", "wheel", "whip", "whistle", "white", "wide", "wife",
		"wind", "window", "wing", "winter", "wire", "wise", "woman", "women", "wood", "word",
		"work", "world", "worm", "wound", "write", "wrong", "year", "yesterday", "young", "youth",
	}

	// 随机选择词根
	root1 := wordRoots1[rand.Intn(len(wordRoots1))]
	root2 := wordRoots2[rand.Intn(len(wordRoots2))]

	// 组合成单词
	word1 := root1 + root2

	// 生成第二个随机单词
	root3 := wordRoots1[rand.Intn(len(wordRoots1))]
	root4 := wordRoots2[rand.Intn(len(wordRoots2))]
	word2 := root3 + root4

	// 生成随机数字 (100-999)
	num := rand.Intn(9000) + 1000

	return fmt.Sprintf("%s-%s-%d", word1, word2, num), nil
}

func (page *SendPage) startSending() {
	defer page.resetSendState()

	var sendFiles []string
	var err error

	if page.currentMode == sendTextMode {
		var tmpFile *os.File
		tmpFile, err = page.createTextFile(page.sendText)
		if err != nil {
			fyne.Do(func() { page.statusLabel.SetText("文本发送失败: " + err.Error()) })
			return
		}
		defer os.Remove(tmpFile.Name())
		sendFiles = []string{tmpFile.Name()}
	} else {
		sendFiles = make([]string, len(page.selectedFiles))
		copy(sendFiles, page.selectedFiles)
	}

	options := page.buildCrocOptions()
	client, err := page.crocManager.CreateCrocClient(options)
	if err != nil {
		fyne.Do(func() { page.statusLabel.SetText("创建客户端失败: " + err.Error()) })
		return
	}

	fyne.Do(func() {
		page.statusLabel.SetText("等待接收方连接...")
		// 更新详情页状态为等待连接
		if page.onUpdateDetail != nil {
			page.onUpdateDetail("waiting", 0.0, "等待接收方连接...")
		}
		// 更新历史记录状态
		page.updateHistoryStatus("waiting", 0.0, "等待接收方连接...")
	})

	filesInfo, emptyFolders, totalNumberFolders, err := croc.GetFilesInfo(sendFiles, page.compressCheck.Checked, false, []string{})
	if err != nil {
		fyne.Do(func() {
			page.statusLabel.SetText("获取文件信息失败: " + err.Error())
			if page.onUpdateDetail != nil {
				page.onUpdateDetail("failed", 0.0, "获取文件信息失败: "+err.Error())
			}
		})
		page.resetSendState()
		return
	}

	// 开始发送，更新状态
	fyne.Do(func() {
		if page.currentMode == sendTextMode {
			page.statusLabel.SetText("正在发送文本...")
			if page.onUpdateDetail != nil {
				page.onUpdateDetail("sending", 0.0, "正在发送文本...")
			}
		} else {
			page.statusLabel.SetText("正在发送文件...")
			if page.onUpdateDetail != nil {
				page.onUpdateDetail("sending", 0.0, "正在发送文件...")
			}
		}
		// 更新历史记录状态为发送中
		page.updateHistoryStatus("sending", 0.0, "正在发送数据...")
	})

	// 在单独的 goroutine 中执行发送，以便可以响应取消
	go func() {
		defer page.resetSendState() // 确保状态被重置

		err := client.Send(filesInfo, emptyFolders, totalNumberFolders)
		if err != nil {
			// 检查是否是因为取消导致的错误
			if page.isCancelled {
				fyne.Do(func() {
					page.statusLabel.SetText("发送已取消")
					if page.onUpdateDetail != nil {
						page.onUpdateDetail("cancelled", 0.0, "发送已取消")
					}
					// 更新历史记录状态为已取消
					page.updateHistoryStatus("cancelled", 0.0, "发送已取消")
				})
			} else {
				fyne.Do(func() {
					page.statusLabel.SetText("发送失败: " + err.Error())
					if page.onUpdateDetail != nil {
						page.onUpdateDetail("failed", 0.0, "发送失败: "+err.Error())
					}
					// 更新历史记录状态为失败
					page.updateHistoryStatus("failed", 0.0, "发送失败: "+err.Error())
				})
			}
			return
		}

		// 发送成功
		fyne.Do(func() {
			page.progressBar.SetValue(1.0)
			page.statusLabel.SetText("发送完成！")
			if page.onUpdateDetail != nil {
				page.onUpdateDetail("completed", 1.0, "发送完成！")
			}
			// 更新历史记录状态为完成
			page.updateHistoryStatus("completed", 1.0, "发送完成！")
		})
	}()

	// 启动进度监控
	go page.monitorProgress()
}

func (page *SendPage) monitorProgress() {
	// 简化的进度监控 - 只模拟进度，实际状态由发送 goroutine 处理
	ctx := page.crocManager.GetContext()
	steps := []float64{0.1, 0.3, 0.5, 0.7, 0.9}

	for _, progress := range steps {
		// 检查取消标志
		if page.isCancelled {
			return
		}

		select {
		case <-ctx.Done():
			// 用户取消
			return
		case <-time.After(500 * time.Millisecond):
			// 再次检查取消标志
			if page.isCancelled {
				return
			}

			// 只有在未取消时才更新进度
			if !page.isCancelled {
				fyne.Do(func() {
					page.progressBar.SetValue(progress)
					// 更新详情页进度
					if page.onUpdateDetail != nil {
						page.onUpdateDetail("sending", progress, fmt.Sprintf("发送中... %.1f%%", progress*100))
					}
				})
			}
		}
	}
}

func (page *SendPage) buildCrocOptions() croc.Options {
	return croc.Options{
		IsSender:      true,
		SharedSecret:  page.codePhrase,
		Debug:         false,
		NoPrompt:      true,
		Stdout:        false,
		HashAlgorithm: "xxhash",
		Curve:         "p256",
		ZipFolder:     page.compressCheck.Checked,
		OnlyLocal:     false,
		DisableLocal:  page.disableLocalCheck.Checked,
		RelayAddress:  page.relayEntry.Text,
		RelayPorts:    []string{"9009", "9010", "9011", "9012", "9013"},
		RelayPassword: page.passwordEntry.Text,
	}
}

func (page *SendPage) createTextFile(textContent string) (*os.File, error) {
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

// createHistoryItem 创建历史记录
func (page *SendPage) createHistoryItem(code string) *storage.HistoryItem {
	var fileName string
	var fileSize string
	var numFiles int

	if page.currentMode == sendFileMode {
		if len(page.selectedFiles) == 0 {
			return nil
		}

		// 获取文件信息
		var totalSize int64
		for _, filePath := range page.selectedFiles {
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			totalSize += info.Size()
		}

		if len(page.selectedFiles) == 1 {
			fileName = filepath.Base(page.selectedFiles[0])
		} else {
			fileName = fmt.Sprintf("%d 个文件", len(page.selectedFiles))
		}

		fileSize = formatFileSize(totalSize)
		numFiles = len(page.selectedFiles)
	} else if page.currentMode == sendTextMode {
		fileName = "文本内容"
		fileSize = formatFileSize(int64(len(page.sendText)))
		numFiles = 1
	} else {
		return nil
	}

	// 创建历史记录项
	historyItem := storage.HistoryItem{
		Type:       "send",
		FileName:   fileName,
		FileSize:   fileSize,
		Code:       code,
		Status:     "in_progress",
		Timestamp:  time.Now(),
		Duration:   0,
		ClientInfo: "MoCroc",
		NumFiles:   numFiles,
	}

	// 保存到存储
	recordID, err := page.storage.Add(historyItem)
	if err != nil {
		fmt.Printf("保存历史记录失败: %v\n", err)
		return nil
	}

	// 保存记录ID供后续更新使用
	page.currentHistoryID = recordID

	return &historyItem
}

// updateHistoryStatus 更新历史记录状态
func (page *SendPage) updateHistoryStatus(status string, progress float64, message string) {
	if page.currentHistoryID == "" {
		return
	}

	duration := time.Since(page.sendStartTime).Seconds()

	err := page.storage.Update(page.currentHistoryID, func(item *storage.HistoryItem) {
		item.Status = status
		item.Duration = int64(duration)
	})

	if err != nil {
		fmt.Printf("更新历史记录失败: %v\n", err)
	}
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
