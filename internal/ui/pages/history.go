package pages

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/shapled/mocroc/internal/storage"
	"github.com/shapled/mocroc/internal/types"
)

type HistoryTab struct {
	// 存储管理器
	storage *storage.HistoryManager

	// UI 组件
	historyList *widget.List
	statsCard   *widget.Card
	clearBtn    *widget.Button
	noDataLabel *widget.Label

	// 状态
	isActive bool

	// 容器
	content fyne.CanvasObject
}

type HistoryItem = storage.HistoryItem

func NewHistoryTab() *HistoryTab {
	tab := &HistoryTab{
		storage: storage.NewHistoryManager(),
	}
	tab.createWidgets()
	tab.buildContent()
	return tab
}

func (tab *HistoryTab) createWidgets() {
	// 历史记录列表
	tab.historyList = widget.NewList(
		func() int {
			return len(tab.storage.GetAll())
		},
		func() fyne.CanvasObject {
			return widget.NewCard("", "", widget.NewLabel(""))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			card := obj.(*widget.Card)
			items := tab.storage.GetAll()
			if id >= len(items) {
				return
			}
			item := items[id]

			statusIcon := tab.getStatusIcon(item.Status)
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
	tab.statsCard = tab.buildStatsCard()

	// 清除按钮
	tab.clearBtn = widget.NewButtonWithIcon("清除历史", theme.DeleteIcon(), tab.onClearHistory)

	// 无数据显示
	tab.noDataLabel = widget.NewLabel("暂无传输记录")
}

func (tab *HistoryTab) buildStatsCard() *widget.Card {
	total, completed, failed, inProgress := tab.storage.GetStats()

	statsText := widget.NewRichTextFromMarkdown(
		"📊 **传输统计**\n" +
			"总计: " + fmt.Sprintf("%d", total) + " | 成功: " + fmt.Sprintf("%d", completed) +
			" | 失败: " + fmt.Sprintf("%d", failed) + " | 进行中: " + fmt.Sprintf("%d", inProgress),
	)

	return widget.NewCard("", "", statsText)
}

func (tab *HistoryTab) buildContent() {
	items := tab.storage.GetAll()
	if len(items) == 0 {
		tab.content = container.NewVBox(
			widget.NewCard("历史记录", "", tab.noDataLabel),
		)
	} else {
		vbox := container.NewVBox(
			tab.statsCard,
			widget.NewSeparator(),
			widget.NewLabel("传输记录:"),
			tab.historyList,
			widget.NewSeparator(),
			tab.clearBtn,
		)
		tab.content = container.NewVScroll(vbox)
	}
}

func (tab *HistoryTab) Build() fyne.CanvasObject {
	return tab.content
}

// TabInterface 实现
func (tab *HistoryTab) GetState() types.TabState {
	return types.TabStateIdle // 历史记录页面不会有传输状态
}

func (tab *HistoryTab) Cancel() error {
	return fmt.Errorf("历史记录页面没有可取消的操作")
}

func (tab *HistoryTab) IsActive() bool {
	return tab.isActive
}

func (tab *HistoryTab) SetActive(active bool) {
	tab.isActive = active
	if active {
		tab.Refresh()
	}
}

// 辅助方法
func (tab *HistoryTab) getStatusIcon(status string) string {
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
func (tab *HistoryTab) onClearHistory() {
	tab.storage.Clear()
	tab.refresh()
}

func (tab *HistoryTab) refresh() {
	tab.statsCard = tab.buildStatsCard()
	tab.buildContent()
	tab.historyList.Refresh()
}

// Refresh 公开的刷新方法
func (tab *HistoryTab) Refresh() {
	tab.refresh()
}
