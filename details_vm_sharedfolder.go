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
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"bytemystery-com/vboxssh/filebrowser"
	"bytemystery-com/vboxssh/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ssfDataType struct {
	name       string
	hostPath   string
	mountPoint string
	autoMount  bool
	readOnly   bool
	global     bool
}
type oldSharedFolderType struct {
	ssfData []*ssfDataType
}

type SharedFolderTab struct {
	oldValues oldSharedFolderType

	list *widget.List

	toolBar           *widget.Toolbar
	toolBarItemAdd    *widget.ToolbarAction
	toolBarItemRemove *widget.ToolbarAction
	toolBarItemEdit   *widget.ToolbarAction

	apply   *widget.Button
	tabItem *container.TabItem

	ssfData []*ssfDataType

	selectedItem *ssfDataType
	isV6         bool
}

var _ DetailsInterface = (*SharedFolderTab)(nil)

func NewSharedFolderTab() *SharedFolderTab {
	ssfTab := SharedFolderTab{}

	ssfTab.apply = widget.NewButton(lang.X("details.vm_ssf.apply", "Apply"), func() {
		ssfTab.Apply()
	})
	ssfTab.apply.Importance = widget.HighImportance

	ssfTab.toolBarItemAdd = widget.NewToolbarAction(theme.ContentAddIcon(), func() { ssfTab.onNewSharedFolder() })
	ssfTab.toolBarItemRemove = widget.NewToolbarAction(theme.ContentRemoveIcon(), func() { ssfTab.onRemoveSharedFolder() })
	ssfTab.toolBarItemEdit = widget.NewToolbarAction(theme.DocumentCreateIcon(), func() { ssfTab.onEditSharedFolder() })
	ssfTab.toolBar = widget.NewToolbar(ssfTab.toolBarItemAdd, ssfTab.toolBarItemEdit, ssfTab.toolBarItemRemove)

	ssfTab.list = widget.NewList(ssfTab.listGetength, ssfTab.listCreateObject, ssfTab.listUpdateItem)
	ssfTab.list.OnSelected = ssfTab.listOnSelected
	ssfTab.list.OnUnselected = ssfTab.listOnUnSelected

	c := container.NewBorder(ssfTab.toolBar, nil, nil, container.NewHBox(container.NewVBox(layout.NewSpacer(), ssfTab.apply, layout.NewSpacer()), util.NewFiller(32, 0)), ssfTab.list)

	ssfTab.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.ssf", "Shared folder"), c)
	return &ssfTab
}

func (ssf *SharedFolderTab) onNewSharedFolder() {
	s := ssfDataType{}
	ssf.doEditSharedFolder(&s, func() {
		ssf.ssfData = append(ssf.ssfData, &s)
		ssf.sort()
		ssf.list.Refresh()
	})
}

func (ssf *SharedFolderTab) onRemoveSharedFolder() {
	if ssf.selectedItem == nil {
		return
	}
	i := slices.Index(ssf.ssfData, ssf.selectedItem)
	if i < 0 {
		return
	}
	ssf.ssfData = slices.Delete(ssf.ssfData, i, i+1)
	ssf.list.UnselectAll()
	ssf.list.Refresh()

	if len(ssf.ssfData) > 0 {
		if len(ssf.ssfData)-1 < i {
			i = len(ssf.ssfData) - 1
		}
		ssf.list.Select(i)
	}
	ssf.updateTaskBarButtons()
}

func (ssf *SharedFolderTab) onEditSharedFolder() {
	if ssf.selectedItem == nil {
		return
	}
	ssf.doEditSharedFolder(ssf.selectedItem, func() {
		ssf.sort()
		ssf.list.Refresh()
	})
}

func (ssf *SharedFolderTab) doBrowse() string {
	return ""
}

func (ssf *SharedFolderTab) doEditSharedFolder(sd *ssfDataType, fOk func()) {
	s, v := getActiveServerAndVm()

	if s == nil || v == nil {
		return
	}

	var dia *dialog.ConfirmDialog

	name := widget.NewEntry()
	name.SetPlaceHolder(lang.X("details.vm_ssf.name.placeholder", "Name"))
	name.SetText(sd.name)
	name.Validator = func(str string) error {
		if str == "" {
			return errors.New("Name is empty")
		}
		return nil
	}
	name.OnSubmitted = func(string) {
		dia.Confirm()
	}

	hostPath := widget.NewEntry()
	hostPath.SetPlaceHolder(lang.X("details.vm_ssf.hostpath.placeholder", "Host path"))
	hostPath.SetText(sd.hostPath)
	hostPath.Validator = func(str string) error {
		if str == "" {
			return errors.New("HostPath is empty")
		}
		return nil
	}
	hostPath.OnSubmitted = func(string) {
		dia.Confirm()
	}

	hostPathBrowse := widget.NewButtonWithIcon(lang.X("details.vm_ssf.hostpath.browse", "Browse..."), theme.SearchIcon(), func() {
		sftp := filebrowser.NewSftpBrowser(s.Client.Client, hostPath.Text, nil,
			lang.X("details.vm_ssf.hostpath.browse.title", "Select folder for sharing"), filebrowser.SftpFileBrowserMode_selectdir)
		sftp.Show(Gui.MainWindow, 0.75, func(file string, fi os.FileInfo, dir string) {
			hostPath.SetText(file)
			if name.Text == "" {
				i := strings.LastIndex(file, "/")
				if i > 0 {
					name.SetText(file[i+1:])
				}
			}
		})
	})

	mountPoint := widget.NewEntry()
	mountPoint.SetPlaceHolder(lang.X("details.vm_ssf.mountpoint.placeholder", "Mountpoint"))
	mountPoint.SetText(sd.mountPoint)
	mountPoint.OnSubmitted = func(string) {
		dia.Confirm()
	}

	readOnly := widget.NewCheck(lang.X("details.vm_ssf.readonly", "Read only"), nil)
	if sd.readOnly {
		readOnly.SetChecked(true)
	}

	autoMount := widget.NewCheck(lang.X("details.vm_ssf.automount", "Auto mount"), nil)
	if sd.autoMount {
		autoMount.SetChecked(true)
	}

	global := widget.NewCheck(lang.X("details.vm_ssf.global", "Global"), nil)
	if sd.global {
		global.SetChecked(true)
	}

	var hb *fyne.Container
	if ssf.isV6 {
		hb = container.NewHBox(autoMount)
	} else {
		hb = container.NewHBox(autoMount, global)
	}

	c := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_ssf.name", "Name")), name,
		widget.NewLabel(lang.X("details.vm_ssf.hostpath", "Host path")), container.NewBorder(nil, nil, nil, hostPathBrowse, hostPath),
		widget.NewLabel(lang.X("details.vm_ssf.mountpoint", "Mount point")), mountPoint,
		readOnly, hb,
	)
	dia = dialog.NewCustomConfirm(lang.X("details.vm_ssf.edit.title", "Edit shared folder"),
		lang.X("import.ok", "Ok"),
		lang.X("import.cancel", "Cancel"), c,
		func(ok bool) {
			if ok {
				err := name.Validate()
				if err != nil {
					dia.Show()
				}
				err = hostPath.Validate()
				if err != nil {
					dia.Show()
				}
				sd.name = name.Text
				sd.hostPath = hostPath.Text
				sd.mountPoint = mountPoint.Text
				sd.global = global.Checked
				sd.autoMount = autoMount.Checked
				sd.readOnly = readOnly.Checked
				fOk()
			}
		}, Gui.MainWindow)
	si := Gui.MainWindow.Canvas().Size()
	var windowScale float32 = 0.65
	dia.Resize(fyne.NewSize(si.Width*windowScale, dia.MinSize().Height*1.1))
	dia.Show()
	Gui.MainWindow.Canvas().Focus(name)
}

func (ssf *SharedFolderTab) listOnSelected(id widget.ListItemID) {
	ssf.selectedItem = ssf.ssfData[id]
	ssf.updateTaskBarButtons()
}

func (ssf *SharedFolderTab) listOnUnSelected(id widget.ListItemID) {
	ssf.selectedItem = nil
	ssf.updateTaskBarButtons()
}

func (ssf *SharedFolderTab) listGetength() int {
	return len(ssf.ssfData)
}

func (ssf *SharedFolderTab) listCreateObject() fyne.CanvasObject {
	icon1 := canvas.NewImageFromResource(theme.QuestionIcon())
	icon1.SetMinSize(fyne.NewSize(16, 16))
	icon1.FillMode = canvas.ImageFillContain
	icon1.Refresh()

	icon2 := canvas.NewImageFromResource(theme.QuestionIcon())
	icon2.SetMinSize(fyne.NewSize(16, 16))
	icon2.FillMode = canvas.ImageFillContain
	icon2.Refresh()

	icon3 := canvas.NewImageFromResource(theme.QuestionIcon())
	icon3.SetMinSize(fyne.NewSize(16, 16))
	icon3.FillMode = canvas.ImageFillContain
	icon3.Refresh()

	text := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	text.Refresh()

	if ssf.isV6 {
		return container.NewHBox(icon1, icon2, text)
	} else {
		return container.NewHBox(icon1, icon2, icon3, text)
	}
}

func (ssf *SharedFolderTab) listUpdateItem(id widget.ListItemID, o fyne.CanvasObject) {
	cont, ok := o.(*fyne.Container)
	if !ok {
		return
	}

	imap := make(map[string]*canvas.Image)

	i := 0
	if !ssf.isV6 {
		icon, ok := cont.Objects[i].(*canvas.Image)
		if !ok {
			return
		}
		imap["global"] = icon
		i++
	}

	icon, ok := cont.Objects[i].(*canvas.Image)
	if !ok {
		return
	}
	imap["readonly"] = icon
	i++

	icon, ok = cont.Objects[i].(*canvas.Image)
	if !ok {
		return
	}
	imap["automount"] = icon
	i++

	text, ok := cont.Objects[i].(*canvas.Text)
	if !ok {
		return
	}
	i++

	item := ssf.ssfData[id]
	if !ssf.isV6 {
		if item.global {
			imap["global"].Resource = Gui.IconGlobal
		} else {
			imap["global"].Resource = theme.HomeIcon()
		}
	}
	if item.readOnly {
		imap["readonly"].Resource = Gui.IconReadOnly
	} else {
		imap["readonly"].Resource = Gui.IconWriteable
	}
	if item.autoMount {
		imap["automount"].Resource = Gui.IconAutomount
	} else {
		imap["automount"].Resource = theme.SettingsIcon()
	}
	for _, icon := range imap {
		icon.SetMinSize(fyne.NewSize(16, 16))
		icon.FillMode = canvas.ImageFillContain
		icon.Refresh()
	}

	text.Text = fmt.Sprintf("%s: %s ⇒ %s", item.name, item.hostPath, item.mountPoint)
	text.Refresh()
}

// calles by selection change
func (ssf *SharedFolderTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()
	ssf.selectedItem = nil
	if s == nil || v == nil {
		ssf.DisableAll()
		ssf.updateTaskBarButtons()
		return
	}

	ssf.apply.Enable()
	v.UpdateStatusEx(&s.Client)
	maj, _ := s.GetVmMajorVersion()
	if maj == 6 {
		ssf.isV6 = true
	} else {
		ssf.isV6 = false
	}

	ssf.ssfData = ssf.ssfData[:0]
	if v.Properties["ssf"] == "yes" {
		index := 1
		for {
			name, ok := v.Properties[fmt.Sprintf("ssfName%d", index)]
			if !ok {
				break
			}
			s := ssfDataType{}
			s.name = name
			s.hostPath = v.Properties[fmt.Sprintf("ssfHostPath%d", index)]
			if v.Properties[fmt.Sprintf("ssfMapping%d", index)] == "global" {
				s.global = true
			}
			if v.Properties[fmt.Sprintf("ssfReadOnly%d", index)] == "true" {
				s.readOnly = true
			}
			if v.Properties[fmt.Sprintf("ssfAutoMount%d", index)] == "true" {
				s.autoMount = true
			}
			s.mountPoint = v.Properties[fmt.Sprintf("ssfMountPoint%d", index)]
			ssf.ssfData = append(ssf.ssfData, &s)
			index++
		}
		ssf.sort()
	}
	ssf.list.UnselectAll()
	ssf.list.Refresh()
	if len(ssf.ssfData) > 0 {
		ssf.list.Select(0)
	}
	ssf.saveOldSharedFolderConfig()

	ssf.updateTaskBarButtons()

	ssf.UpdateByStatus()
}

func (ssf *SharedFolderTab) sort() {
	slices.SortFunc(ssf.ssfData, func(a, b *ssfDataType) int {
		if a.global && !b.global {
			return -1
		}
		if !a.global && b.global {
			return 1
		}
		aN := strings.ToLower(a.name)
		bN := strings.ToLower(a.name)
		if aN < bN {
			return -1
		}
		if aN > bN {
			return 1
		}
		return 0
	})
}

// called from status updates
func (ssf *SharedFolderTab) UpdateByStatus() {
	_, v := getActiveServerAndVm()
	if v != nil {
		state, err := v.GetState()
		_ = state
		if err != nil {
			return
		}
		/*
			switch state {
			case vm.RunState_unknown, vm.RunState_meditation:
				ssf.DisableAll()

			case vm.RunState_running:
				ssf.enabled.Disable()
				ssf.hostDriver.Disable()
				ssf.controller.Disable()
				ssf.codec.Disable()
				if ssf.enabled.Checked {
					ssf.out.Enable()
					ssf.in.Enable()
				} else {
					ssf.out.Disable()
					ssf.in.Disable()
				}

			case vm.RunState_paused:
				ssf.enabled.Disable()
				ssf.controller.Disable()
				ssf.codec.Disable()
				ssf.hostDriver.Disable()
				if ssf.enabled.Checked {
					ssf.out.Enable()
					ssf.in.Enable()
				} else {
					ssf.out.Disable()
					ssf.in.Disable()
				}

			case vm.RunState_saved:
				ssf.enabled.Disable()
				ssf.controller.Disable()
				ssf.codec.Disable()
				if ssf.enabled.Checked {
					ssf.hostDriver.Enable()
					ssf.out.Enable()
					ssf.in.Enable()
				} else {
					ssf.hostDriver.Disable()
					ssf.out.Disable()
					ssf.in.Disable()
				}

			case vm.RunState_off, vm.RunState_aborted:
				ssf.enabled.Enable()
				ssf.enableDisableAudioCtrls(ssf.enabled.Checked)

			default:
				SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
			}
		*/
	} else {
		ssf.DisableAll()
	}
}

func (ssf *SharedFolderTab) DisableAll() {
	ssf.apply.Disable()
}

func (ssf *SharedFolderTab) saveOldSharedFolderConfig() {
	ssf.oldValues.ssfData = ssf.oldValues.ssfData[:0]
	for _, f := range ssf.ssfData {
		s := ssfDataType{}
		s = *f
		ssf.oldValues.ssfData = append(ssf.oldValues.ssfData, &s)
	}
}

func (ssf *SharedFolderTab) updateTaskBarButtons() {
	_, v := getActiveServerAndVm()
	if v == nil {
		ssf.toolBarItemAdd.Disable()
		ssf.toolBarItemEdit.Disable()
		ssf.toolBarItemRemove.Disable()
		return
	}
	ssf.toolBarItemAdd.Enable()
	if ssf.selectedItem == nil {
		ssf.toolBarItemEdit.Disable()
		ssf.toolBarItemRemove.Disable()
	} else {
		ssf.toolBarItemEdit.Enable()
		ssf.toolBarItemRemove.Enable()
	}
}

func (ssf *SharedFolderTab) Apply() {
	s, v := getActiveServerAndVm()
	if v != nil {
		ResetStatus()
	}
	addList := make([]*ssfDataType, 0, len(ssf.ssfData))
	for _, newItem := range ssf.ssfData {
		flag := true
		for _, oldItem := range ssf.oldValues.ssfData {
			if *newItem == *oldItem {
				flag = false
				break
			}
		}
		if flag {
			addList = append(addList, newItem)
		}
	}

	delList := make([]*ssfDataType, 0, len(ssf.ssfData))
	for _, oldItem := range ssf.oldValues.ssfData {
		flag := true
		for _, newItem := range ssf.ssfData {
			if *newItem == *oldItem {
				flag = false
				break
			}
		}
		if flag {
			delList = append(delList, oldItem)
		}
	}

	// Remove
	for _, item := range delList {
		err := v.RemoveSharedFolder(s, item.name, item.global, VMStatusUpdateCallBack)
		if err != nil {
			SetStatusText(fmt.Sprintf(lang.X("details.vm_ssf.removessf.error", "Remove shared folder '%s' from VM '%s' failed"), item.name, v.Name), MsgError)
		}
	}

	for _, item := range addList {
		err := v.AddSharedFolder(s, item.name, item.hostPath, item.mountPoint, item.readOnly, item.autoMount, item.global, VMStatusUpdateCallBack)
		if err != nil {
			SetStatusText(fmt.Sprintf(lang.X("details.vm_ssf.addssf.error", "Add shared folder '%s' to VM '%s' failed"), item.name, v.Name), MsgError)
		}
	}
}
