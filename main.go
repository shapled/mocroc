package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/shapled/mocroc/internal/ui"
)

func main() {
	// 创建 Fyne 应用
	a := app.NewWithID("com.shapled.mocroc")

	// 创建主窗口
	w := a.NewWindow("Croc Mobile")
	w.SetIcon(nil) // TODO: 添加应用图标

	// 构建主界面
	mainUI := ui.NewMainUI(a, w)

	// 设置窗口内容并显示
	w.SetContent(mainUI)
	w.ShowAndRun()
}
