package pages

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/shapled/mocroc/internal/storage"
)

type HistoryPage struct {
	// 存储管理器
	storage *storage.HistoryStorage

	// UI 组件
	historyList *widget.List
	statsCard   *widget.Card
	clearBtn    *widget.Button
	noDataLabel *widget.Label

	// 容器
	content fyne.CanvasObject
}

type HistoryItem = storage.HistoryItem

func NewHistoryPage(storage *storage.HistoryStorage) *HistoryPage {
	tab := &HistoryPage{
		storage: storage,
	}
	tab.createWidgets()
	tab.buildContent()
	return tab
}

func (page *HistoryPage) createWidgets() {
	// 历史记录列表
	page.historyList = widget.NewList(
		func() int {
			items, _ := page.storage.GetAll()
			return len(items)
		},
		func() fyne.CanvasObject {
			return widget.NewCard("", "", widget.NewLabel(""))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			card := obj.(*widget.Card)
			items, err := page.storage.GetAll()
			if err != nil {
				card.SetContent(widget.NewLabel("加载失败: " + err.Error()))
				return
			}
			if id >= len(items) {
				return
			}
			item := items[id]

			statusIcon := page.getStatusIcon(item.Status)
			description := widget.NewRichTextFromMarkdown(
				"**" + item.FileName + "**\n" +
					"📁 " + item.FileSize + " | 🔑 " + item.Code + "\n" +
					"🕒 " + item.Timestamp.Format("2006-01-02 15:04") + " | " +
					statusIcon + " " + item.Status,
			)

			card.SetTitle("")
			card.SetContent(description)
		},
	)

	// 统计信息
	page.statsCard = page.buildStatsCard()

	// 清除按钮
	page.clearBtn = widget.NewButtonWithIcon("清除历史", theme.DeleteIcon(), page.onClearHistory)

	// 无数据显示
	page.noDataLabel = widget.NewLabel("暂无传输记录")
}

func (page *HistoryPage) buildStatsCard() *widget.Card {
	total, completed, failed, inProgress, err := page.storage.GetStats()

	if err != nil {
		statsText := widget.NewLabel("获取统计信息失败: " + err.Error())
		return widget.NewCard("", "", statsText)
	}

	statsText := widget.NewRichTextFromMarkdown(
		"📊 **传输统计**\n" +
			"总计: " + fmt.Sprintf("%d", total) + " | 成功: " + fmt.Sprintf("%d", completed) +
			" | 失败: " + fmt.Sprintf("%d", failed) + " | 进行中: " + fmt.Sprintf("%d", inProgress),
	)

	return widget.NewCard("", "", statsText)
}

func (page *HistoryPage) buildContent() {
	items, err := page.storage.GetAll()
	if err != nil {
		page.content = container.NewVBox(
			widget.NewCard("历史记录", "", widget.NewLabel("加载历史记录失败: "+err.Error())),
		)
		return
	}

	if len(items) == 0 {
		page.content = container.NewVBox(
			widget.NewCard("历史记录", "", page.noDataLabel),
		)
	} else {
		vbox := container.NewVBox(
			page.statsCard,
			widget.NewSeparator(),
			widget.NewLabel("传输记录:"),
			page.historyList,
			widget.NewSeparator(),
			page.clearBtn,
		)
		page.content = container.NewVScroll(vbox)
	}
}

func (page *HistoryPage) Build() fyne.CanvasObject {
	return page.content
}

// 辅助方法
func (page *HistoryPage) getStatusIcon(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "in_progress":
		return "⏳"
	default:
		return "❓"
	}
}

// 事件处理器
func (page *HistoryPage) onClearHistory() {
	page.storage.Clear()
	page.refresh()
}

func (page *HistoryPage) refresh() {
	page.statsCard = page.buildStatsCard()
	page.buildContent()
	page.historyList.Refresh()
}

// Refresh 公开的刷新方法
func (page *HistoryPage) Refresh() {
	page.refresh()
}
