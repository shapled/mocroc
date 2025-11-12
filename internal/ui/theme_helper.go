package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// ThemeHelper 提供一致的UI样式辅助函数
type ThemeHelper struct{}

// NewThemeHelper 创建主题助手
func NewThemeHelper() *ThemeHelper {
	return &ThemeHelper{}
}

// CreatePrimaryButton 创建主要按钮（高重要性，蓝色主题）
func (t *ThemeHelper) CreatePrimaryButton(text string, icon fyne.Resource, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon(text, icon, onTap)
	btn.Importance = widget.HighImportance
	return btn
}

// CreateSecondaryButton 创建次要按钮（中等重要性，灰色主题）
func (t *ThemeHelper) CreateSecondaryButton(text string, icon fyne.Resource, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon(text, icon, onTap)
	btn.Importance = widget.MediumImportance
	return btn
}

// CreateDangerButton 创建危险按钮（取消等操作，红色主题）
func (t *ThemeHelper) CreateDangerButton(text string, icon fyne.Resource, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon(text, icon, onTap)
	btn.Importance = widget.DangerImportance
	return btn
}

// CreateStandardEntry 创建标准输入框
func (t *ThemeHelper) CreateStandardEntry(placeholder string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	return entry
}

// CreateCardWithPadding 创建带内边距的卡片
func (t *ThemeHelper) CreateCardWithPadding(title, subtitle string, content fyne.CanvasObject) *widget.Card {
	paddedContent := widget.NewCard(title, subtitle, content)
	// 通过容器添加内边距
	return paddedContent
}

// CreateSectionTitle 创建章节标题
func (t *ThemeHelper) CreateSectionTitle(text string) *widget.Label {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.Importance = widget.HighImportance
	return label
}

// CreateSubTitle 创建副标题
func (t *ThemeHelper) CreateSubTitle(text string) *widget.Label {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{})
	label.Importance = widget.MediumImportance
	return label
}

// CreateStatusMessage 创建状态消息
func (t *ThemeHelper) CreateStatusMessage(text string, isError bool) *widget.Label {
	label := widget.NewLabel(text)
	if isError {
		label.Importance = widget.DangerImportance
	} else {
		label.Importance = widget.MediumImportance
	}
	return label
}

// 标准按钮尺寸
const (
	StandardButtonHeight = 56
	StandardButtonWidth  = 280
	SmallButtonHeight    = 44
	SmallButtonWidth     = 200
)

// ApplyStandardButtonSize 应用标准按钮尺寸
func (t *ThemeHelper) ApplyStandardButtonSize(btn *widget.Button, isSmall bool) {
	if isSmall {
		btn.Resize(fyne.NewSize(SmallButtonWidth, SmallButtonHeight))
	} else {
		btn.Resize(fyne.NewSize(StandardButtonWidth, StandardButtonHeight))
	}
}

// ApplyStandardEntrySize 应用标准输入框尺寸
func (t *ThemeHelper) ApplyStandardEntrySize(entry *widget.Entry) {
	entry.Resize(fyne.NewSize(StandardButtonWidth, SmallButtonHeight))
}