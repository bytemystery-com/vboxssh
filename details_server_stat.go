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
	"fmt"
	"time"

	"bytemystery-com/vboxssh/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ServerStatInfos struct {
	content fyne.CanvasObject
	total   *widget.Label
	read    *widget.Label
	write   *widget.Label

	unitTotal *widget.Label
	unitRead  *widget.Label
	unitWrite *widget.Label

	updateTicker       *time.Ticker
	updateTickerCancel chan bool

	tabItem *container.TabItem
}

var _ DetailsInterface = (*ServerStatInfos)(nil)

func NewServerStatTab() *ServerStatInfos {
	srv := ServerStatInfos{}

	srv.total = widget.NewLabel("")
	srv.total.Importance = widget.HighImportance
	srv.read = widget.NewLabel("")
	srv.write = widget.NewLabel("")

	labelTotal := widget.NewLabel(lang.X("details.srvstat.total", "Total"))
	labelTotal.Importance = widget.HighImportance
	labelRead := widget.NewLabel(lang.X("details.srvstat.read", "Read"))
	labelWrite := widget.NewLabel(lang.X("details.srvstat.write", "Write"))

	srv.unitTotal = widget.NewLabel("")
	srv.unitTotal.Importance = widget.HighImportance
	srv.unitRead = widget.NewLabel("")
	srv.unitWrite = widget.NewLabel("")

	formWidth := util.GetFormWidth()
	fieldSize := util.GetDefaultTextSize("XXXXXXX")

	c1 := container.New(layout.NewFormLayout(),
		labelTotal, container.NewHBox(container.NewGridWrap(fieldSize, srv.total), srv.unitTotal),
		labelRead, container.NewHBox(container.NewGridWrap(fieldSize, srv.read), srv.unitRead),
		labelWrite, container.NewHBox(container.NewGridWrap(fieldSize, srv.write), srv.unitWrite),
	)

	content := container.NewGridWrap(fyne.NewSize(formWidth, c1.MinSize().Height), c1)

	srv.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.stat", "Stat"), content)

	srv.updateTicker = time.NewTicker(time.Duration(500) * time.Millisecond)

	return &srv
}

func (srv *ServerStatInfos) UpdateDisplay() {
	s := Data.GetServer(Gui.ActiveItemServer, true)
	if s == nil {
		return
	}
	if s.IsLocal() {
		t := lang.X("details.srvstat.local", "Only available for SSH connections")
		srv.total.SetText(t)
		srv.read.SetText(t)
		srv.write.SetText(t)
		t = ""
		srv.unitWrite.SetText(t)
		srv.unitRead.SetText(t)
		srv.unitTotal.SetText(t)
	} else {
		r, w := s.GetStatistic()
		val, unit := srv.formatBytesDisplay(r)
		srv.read.SetText(val)
		srv.unitRead.SetText(unit)

		val, unit = srv.formatBytesDisplay(w)
		srv.write.SetText(val)
		srv.unitWrite.SetText(unit)

		val, unit = srv.formatBytesDisplay(r + w)
		srv.total.SetText(val)
		srv.unitTotal.SetText(unit)
	}
}

func (srv *ServerStatInfos) UpdateBySelect() {
	srv.UpdateDisplay()
}

func (srv *ServerStatInfos) Apply() {
}

func (srv *ServerStatInfos) DisableAll() {
}

func (srv *ServerStatInfos) UpdateByStatus() {
	srv.UpdateDisplay()
}

func (srv *ServerStatInfos) formatBytesDisplay(v uint64) (string, string) {
	val := float64(v)
	if val < 1000*1000 {
		return fmt.Sprintf("%.0f", float64(val/1000.0)), "kByte"
	} else if v < 1000*1000*1000 {
		return fmt.Sprintf("%.1f", float64(val/(1000.0*1000.0))), "MByte"
	} else {
		return fmt.Sprintf("%.2f", float64(val/(1000.0*1000.0*1000.0))), "GByte"
	}
}
