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
	"image/color"
	"strconv"
	"time"

	"bytemystery-com/vboxssh/util"

	"bytemystery-com/vboxssh/server"
	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

const (
	SERVER_ADD_NEW_UUID = "-1"
)

type ServerSshInfos struct {
	content       fyne.CanvasObject
	name          *widget.Entry
	user          *widget.Entry
	host          *widget.Entry
	port          *widget.Entry
	pass          *widget.Entry
	keyFile       *widget.Entry
	keyFileBrowse *widget.Button
	apply         *widget.Button

	updateTicker       *time.Ticker
	updateTickerCancel chan bool

	tabItem *container.TabItem
}

var _ DetailsInterface = (*ServerSshInfos)(nil)

func NewSshServerTab() *ServerSshInfos {
	srv := ServerSshInfos{}

	srv.name = widget.NewEntry()
	srv.name.SetPlaceHolder(lang.X("details.srvssh.name_placeholder", "Name for display"))
	srv.user = widget.NewEntry()
	srv.user.SetPlaceHolder(lang.X("details.srvssh.user_placeholder", "SSH user"))
	srv.host = widget.NewEntry()
	srv.host.SetPlaceHolder(lang.X("details.srvssh.host_placeholder", "SSH host (xy.com)"))

	srv.port = widget.NewEntry()
	srv.port.SetPlaceHolder(lang.X("details.srvssh.port_placeholder", "SSH port (22)"))
	srv.port.OnChanged = util.GetNumberFilter(srv.port, nil)

	srv.pass = widget.NewPasswordEntry()
	srv.pass.SetPlaceHolder(lang.X("details.srvssh.pass_placeholder", "SSH / key password"))
	srv.keyFile = widget.NewEntry()
	srv.keyFile.SetPlaceHolder(lang.X("details.srvssh.keyfile_placeholder", "SSH keyfile"))
	srv.keyFileBrowse = widget.NewButton(lang.X("details.srvssh.browse", "Browse"), func() {
		srv.browseKeyFile()
	})
	srv.apply = widget.NewButton(lang.X("details.srvssh.apply", "Apply"), func() {
		srv.Apply()
	})
	srv.apply.Importance = widget.HighImportance

	formWidth := util.GetFormWidth() / 2
	labelWidth := util.GetDefaultTextWidth("XXXXXXXXXX")

	dummy := canvas.NewRectangle(color.Transparent)
	dummy.SetMinSize(widget.NewLabel("X").MinSize())

	grid1 := container.New(layout.NewFormLayout(),
		container.NewGridWrap(fyne.NewSize(labelWidth, 1),
			widget.NewLabel(lang.X("details.srvssh.name", "Name"))), srv.name,
		widget.NewLabel(lang.X("details.srvssh.host", "Host")), srv.host,
		widget.NewLabel(lang.X("details.srvssh.user", "User")), srv.user)

	grid2 := container.New(layout.NewFormLayout(),
		dummy, dummy,
		widget.NewLabel(lang.X("details.srvssh.port", "Port")), srv.port,
		widget.NewLabel(lang.X("details.srvssh.password", "Password")), srv.pass)

	grid3 := container.New(layout.NewFormLayout(),
		container.NewGridWrap(fyne.NewSize(labelWidth, 1),
			widget.NewLabel(lang.X("details.srvssh.keyfile", "Keyfile"))), srv.keyFile)

	i1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	i2 := container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)
	i3 := container.NewGridWrap(fyne.NewSize(2*formWidth, grid3.MinSize().Height), grid3)

	content := container.NewVBox(util.NewVFiller(0.5), container.NewHBox(i1, i2),
		container.NewHBox(i3, srv.keyFileBrowse), container.NewHBox(layout.NewSpacer(), srv.apply, util.NewFiller(32, 0)))

	srv.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.ssh", "SSH"), content)

	srv.updateTicker = time.NewTicker(time.Duration(500) * time.Millisecond)

	return &srv
}

func (srv *ServerSshInfos) UpdateBySelect() {
	if Gui.ActiveItemServer == SERVER_ADD_NEW_UUID {
		return
	} else {
		srv.apply.SetText(lang.X("details.srvssh.apply", "Apply"))
	}

	if Gui.ActiveItemServer == "" {
		srv.reset()
		return
	}
	s := Data.GetServer(Gui.ActiveItemServer, true)
	if s == nil {
		srv.reset()
		return
	}
	srv.name.SetText(s.Name)
	srv.host.SetText(s.Host)
	srv.user.SetText(s.User)
	srv.port.SetText(strconv.Itoa(s.Port))
	srv.pass.SetText(s.Password)
	srv.keyFile.SetText(s.KeyFile)
}

func (srv *ServerSshInfos) reset() {
	srv.name.SetText("")
	srv.user.SetText("")
	srv.pass.SetText("")
	srv.host.SetText("")
	srv.port.SetText("")
	srv.keyFile.SetText("")
}

func (srv *ServerSshInfos) browseKeyFile() {
	dia := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			return
		}
		if r == nil {
			return
		}
		r.Close()
		srv.keyFile.SetText(r.URI().String())
	}, Gui.MainWindow)

	u, err := storage.ParseURI(srv.keyFile.Text)
	if err == nil {
		parent, err := storage.Parent(u)
		if err != nil {
			return
		}

		lister, err := storage.ListerForURI(parent)
		if err != nil {
			return
		}
		dia.SetLocation(lister)
	}

	dia.SetView(dialog.ListView)
	ms := Gui.MainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(ms.Width*.8, ms.Height*.8))
	dia.Show()
}

func (srv *ServerSshInfos) add() {
	var p int
	var err error
	if len(srv.port.Text) > 0 {
		p, err = strconv.Atoi(srv.port.Text)
	}
	if err != nil {
		return
	}
	s := server.Server{
		Port:          p,
		Name:          srv.name.Text,
		Host:          srv.host.Text,
		User:          srv.user.Text,
		Password:      srv.pass.Text,
		KeyFile:       srv.keyFile.Text,
		KeyFileReader: readKeyFile,
	}
	var vms *vm.VmServer
	vms = Data.AddData(s, func() {
		setAfterConnectStatus(vms, nil)
	}, func(err error) {
		setAfterConnectStatus(vms, err)
	})
	SaveServers()
	srv.SetAddNewMode(false)
	Gui.Tree.Refresh()
	Gui.Tree.Select(vms.UUID)
	treeSetSelectedItem(vms.UUID, "")
}

func (srv *ServerSshInfos) Apply() {
	if Gui.ActiveItemServer == SERVER_ADD_NEW_UUID {
		srv.add()
	} else {

		if Gui.ActiveItemServer == "" {
			return
		}
		s := Data.GetServer(Gui.ActiveItemServer, true)
		if s == nil {
			return
		}
		p, err := strconv.Atoi(srv.port.Text)
		if err != nil {
			return
		}
		s.Port = p
		s.Name = srv.name.Text
		s.Host = srv.host.Text
		s.User = srv.user.Text
		s.Password = srv.pass.Text
		s.KeyFile = srv.keyFile.Text
		Gui.Tree.Refresh()
		SaveServers()
		SetStatusText(fmt.Sprintf(lang.X("status.server_add_ok", "Server '%s' where added."), s.Name), MsgInfo)
		s.Disonnect(&s.Client.Client)
		s.Connect(func() {
			setAfterConnectStatus(s, nil)
		}, func(err error) {
			setAfterConnectStatus(s, err)
		})
	}
}

func (srv *ServerSshInfos) SetAddNewMode(add bool) {
	if add {
		Gui.ActiveItemServer = SERVER_ADD_NEW_UUID
		srv.apply.SetText(lang.X("details.srvssh.add", "Add"))
		Gui.Details.Open(0)
		srv.reset()
	} else {
		Gui.ActiveItemServer = ""
		srv.apply.SetText(lang.X("details.srvssh.apply", "Apply"))
	}
}

func (srv *ServerSshInfos) DisableAll() {
	srv.name.Disable()
	srv.user.Disable()
	srv.host.Disable()
	srv.port.Disable()
	srv.pass.Disable()
	srv.keyFile.Disable()
	srv.keyFileBrowse.Disable()
	srv.apply.Disable()
}

func (srv *ServerSshInfos) UpdateByStatus() {
}
