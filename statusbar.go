// Copyright (c) 2026 Reiner Pröls
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// SPDX-License-Identifier: MIT
//
// Author: Reiner Pröls

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
