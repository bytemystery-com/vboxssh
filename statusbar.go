package main

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MsgType int

const (
	MsgInfo MsgType = iota
	MsgWarning
	MsgError
)

var (
	colorInfo    any = theme.ColorNameForeground
	colorWarning any = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	colorError   any = color.NRGBA{R: 255, G: 0, B: 9, A: 255}
)

func SetStatusText(msg string, t MsgType) {
	msg = strings.ReplaceAll(msg, "\r\n", ". ")
	msg = strings.ReplaceAll(msg, "\n", ". ")
	fyne.Do(func() {
		Gui.StartusBar.SetText(msg)
		switch t {
		case MsgInfo:
			Gui.StartusBar.SetTextWithColor(msg, theme.ColorNameForeground)
		case MsgWarning:
			Gui.StartusBar.SetTextWithColor(msg, colorWarning)
		case MsgError:
			Gui.StartusBar.SetTextWithColor(msg, colorError)
		}
	})
}

func ResetStatus() {
	SetStatusText("", MsgInfo)
}
