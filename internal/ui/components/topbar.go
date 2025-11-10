package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TopBar struct {
	*fyne.Container

	title string

	backBtn    *widget.Button
	titleLabel *widget.Label
}

func NewTopBar(title string, goBack func()) *TopBar {
	backBtn := widget.NewButtonWithIcon("返回", theme.NavigateBackIcon(), goBack)
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	topbar := container.NewStack(container.NewHBox(backBtn), titleLabel)
	return &TopBar{
		Container:  topbar,
		title:      title,
		backBtn:    backBtn,
		titleLabel: titleLabel,
	}
}

func (topbar *TopBar) SetTitle(title string) {
	topbar.titleLabel.SetText(title)
}
