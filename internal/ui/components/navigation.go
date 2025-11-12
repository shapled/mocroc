package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NavigationItem 导航项
type NavigationItem struct {
	Icon     fyne.Resource
	Label    string
	OnTap    func()
	IsActive bool
}

// BottomNavigation 底部导航栏组件
type BottomNavigation struct {
	container *fyne.Container
	items     []*NavigationItem

	// 回调函数
	onNavigate func(pageType string)
}

// NewBottomNavigation 创建新的底部导航栏
func NewBottomNavigation(onNavigate func(pageType string)) *BottomNavigation {
	nav := &BottomNavigation{
		onNavigate: onNavigate,
		items:      make([]*NavigationItem, 0),
	}

	nav.createNavigation()
	return nav
}

// createNavigation 创建导航栏
func (nav *BottomNavigation) createNavigation() {
	// 创建导航按钮 - 使用一致的图标主题
	nav.items = []*NavigationItem{
		{
			Icon:  theme.HomeIcon(),
			Label: "首页",
			OnTap: func() {
				if nav.onNavigate != nil {
					nav.onNavigate("home")
				}
			},
		},
		{
			Icon:  theme.UploadIcon(),
			Label: "发送",
			OnTap: func() {
				if nav.onNavigate != nil {
					nav.onNavigate("send")
				}
			},
		},
		{
			Icon:  theme.DownloadIcon(),
			Label: "接收",
			OnTap: func() {
				if nav.onNavigate != nil {
					nav.onNavigate("receive")
				}
			},
		},
		{
			Icon:  theme.HistoryIcon(),
			Label: "历史",
			OnTap: func() {
				if nav.onNavigate != nil {
					nav.onNavigate("history")
				}
			},
		},
	}

	// 设置默认激活状态
	nav.items[0].IsActive = true

	// 创建导航按钮
	navButtons := make([]fyne.CanvasObject, len(nav.items))
	for i, item := range nav.items {
		navButtons[i] = nav.createNavButton(item)
	}

	// 创建导航容器 - 使用水平布局，居中对齐
	buttonsContainer := container.NewHBox(navButtons...)

	// 添加内边距和背景
	paddedContainer := container.NewPadded(buttonsContainer)

	// 创建导航栏背景
	navBackground := widget.NewCard("", "", paddedContainer)

	nav.container = container.NewMax(navBackground)
}

// createNavButton 创建单个导航按钮
func (nav *BottomNavigation) createNavButton(item *NavigationItem) fyne.CanvasObject {
	icon := widget.NewIcon(item.Icon)
	label := widget.NewLabel(item.Label)

	// 设置标签样式
	label.Alignment = fyne.TextAlignCenter

	// 创建按钮容器
	btnContainer := container.NewVBox(
		container.NewCenter(icon),
		container.NewCenter(label),
	)

	// 设置按钮大小 - 确保至少44px的触摸区域
	btnContainer.Resize(fyne.NewSize(80, 70))

	// 创建按钮用于交互
	btn := widget.NewButton("", item.OnTap)
	btn.Importance = widget.MediumImportance

	// 根据激活状态设置样式
	nav.updateButtonStyle(label, item.IsActive)

	// 创建包含按钮和样式的容器
	wrapper := container.NewStack(
		btnContainer,
		btn,
	)

	// 设置最小尺寸
	wrapper.Resize(fyne.NewSize(80, 70))

	return wrapper
}

// updateButtonStyle 更新按钮样式
func (nav *BottomNavigation) updateButtonStyle(label *widget.Label, isActive bool) {
	if isActive {
		// 激活状态：加粗文本
		label.TextStyle = fyne.TextStyle{Bold: true}
		label.Importance = widget.HighImportance
	} else {
		// 非激活状态
		label.TextStyle = fyne.TextStyle{Bold: false}
		label.Importance = widget.MediumImportance
	}

	label.Refresh()
}

// SetActivePage 设置当前激活的页面
func (nav *BottomNavigation) SetActivePage(pageType string) {
	// 重置所有激活状态
	for _, item := range nav.items {
		item.IsActive = false
	}

	// 设置新的激活状态
	switch pageType {
	case "home":
		nav.items[0].IsActive = true
	case "send":
		nav.items[1].IsActive = true
	case "receive":
		nav.items[2].IsActive = true
	case "history":
		nav.items[3].IsActive = true
	}

	// 重新创建导航以更新样式
	nav.createNavigation()
}

// Container 返回导航容器
func (nav *BottomNavigation) Container() *fyne.Container {
	return nav.container
}

// Hide 隐藏导航栏
func (nav *BottomNavigation) Hide() {
	nav.container.Hide()
}

// Show 显示导航栏
func (nav *BottomNavigation) Show() {
	nav.container.Show()
}