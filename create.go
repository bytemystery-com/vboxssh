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

	"bytemystery-com/vboxssh/util"
	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bytemystery-com/colorlabel"
	"github.com/google/uuid"
)

func doCreateVm() {
	s, _ := getActiveServerAndVm()

	if s == nil {
		return
	}

	name := widget.NewEntry()
	name.SetPlaceHolder(lang.X("create.name.placeholder", "Name of the new VM"))
	item := widget.NewFormItem(lang.X("create.name", "Name"), name)

	dia := dialog.NewForm(lang.X("create.title", "Create new VM"),
		lang.X("create.create", "Create"),
		lang.X("create.cancel", "Cancel"), []*widget.FormItem{item}, func(ok bool) {
			err := s.CreateVm(&s.Client, name.Text)
			if err != nil {
				SetStatusText(fmt.Sprintf(lang.X("create.failed", "Creating VM width name '%s' failed"), name.Text), MsgError)
			} else {
				SetStatusText(fmt.Sprintf(lang.X("create.created", "VM width name '%s' was created"), name.Text), MsgInfo)
				go treeUpdateVmList(s.UUID)
			}
		}, Gui.MainWindow)
	var windowScale float32 = 1.25
	// si := Gui.MainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(dia.MinSize().Width*windowScale, dia.MinSize().Height*windowScale))
	dia.Show()
}

func doDeleteVm() {
	s, m := getActiveServerAndVm()

	if s == nil || m == nil {
		return
	}

	label := widget.NewLabel(m.Name)
	del := widget.NewCheck(lang.X("delete.all", "Delete all files"), nil)
	item1 := widget.NewFormItem("", del)
	item2 := widget.NewFormItem("", label)

	dia := dialog.NewForm(lang.X("delete.title", "Delete VM"),
		lang.X("delete.delete", "Delete"),
		lang.X("delete.cancel", "Cancel"), []*widget.FormItem{item2, item1}, func(ok bool) {
			if ok {
				dialog.ShowConfirm(lang.X("delete.confirm.title", "Confirm deleting VM"),
					fmt.Sprintf(lang.X("delete.confirm.msg", "Do you really want to delete the VM\n'%s'\nfrom server '%s' ?"), m.Name, s.Name),
					func(ok bool) {
						if ok {
							err := m.DeleteVm(s, del.Checked)
							if err != nil {
								SetStatusText(fmt.Sprintf(lang.X("delete.failed", "Deleting VM width name '%s' failed"), m.Name), MsgError)
							} else {
								SetStatusText(fmt.Sprintf(lang.X("delete.deleted", "VM width name '%s' was deleted"), m.Name), MsgInfo)
								go treeUpdateVmList(s.UUID)
							}
						}
					}, Gui.MainWindow)
			}
		}, Gui.MainWindow)

	var windowScale float32 = 0.25
	si := Gui.MainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
	dia.Show()
}

func doCloneVm() {
	s, m := getActiveServerAndVm()

	if s == nil || m == nil {
		return
	}

	cloneModeMap := map[int]vm.CloneModeType{0: vm.CloneMode_machine, 1: vm.CloneMode_all}
	linkMap := map[int]vm.CloneOptionsType{0: vm.CloneOption_none, 1: vm.CloneOption_link}
	macMap := map[int]vm.CloneOptionsType{0: vm.CloneOption_keepallmacs, 1: vm.CloneOption_keepnatmacs, 2: vm.CloneOption_none}
	diskMap := map[bool]vm.CloneOptionsType{false: vm.CloneOption_none, true: vm.CloneOption_keepdiscnames}
	hwIdMap := map[bool]vm.CloneOptionsType{false: vm.CloneOption_none, true: vm.CloneOption_keephwuuids}

	label := colorlabel.NewColorLabel(m.Name, theme.ColorNamePrimary, nil, 1.0)
	name := widget.NewEntry()
	name.SetText(m.Name + " " + lang.X("clone.clone.append", "Clone"))

	name.SetPlaceHolder(lang.X("clone.name.placeholder", "Name for the cloned VM"))
	var cloneType *widget.Select
	var cloneModeType *widget.Select
	cloneType = widget.NewSelect([]string{
		lang.X("clone.type.full", "Full clone"),
		lang.X("clone.type.link", "Linked clone"),
	}, func(string) {
		if linkMap[cloneType.SelectedIndex()] == vm.CloneOption_link {
			cloneModeType.SetSelectedIndex(0)
			cloneModeType.Disable()
		} else {
			cloneModeType.Enable()
		}
	})
	cloneModeType = widget.NewSelect([]string{
		lang.X("clone.mode.current", "Current"),
		lang.X("clone.mode.all", "All"),
	}, nil)
	mac := widget.NewSelect([]string{
		lang.X("clone.mac.all", "All"),
		lang.X("clone.mac.onlynat", "Only NAT"),
		lang.X("clone.mac.allnew", "Create new for all"),
	}, nil)

	disk := widget.NewCheck(lang.X("clone.keepdisknames", "Keep disk names"), nil)
	hwuuid := widget.NewCheck(lang.X("clone.keephwuuids", "Keep hardware UUIDs"), nil)

	c := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("clone.name", "Name for the cloned VM")), name,
		widget.NewLabel(lang.X("clone.type", "Clone type")), cloneType,
		widget.NewLabel(lang.X("clone.state", "Snapshots")), cloneModeType,
		widget.NewLabel(lang.X("clone.mac", "Keep MAC addresses")), mac,
		disk, hwuuid,
	)

	dia := dialog.NewCustomConfirm(lang.X("clone.title", "Clone VM"),
		lang.X("clone.clone", "Clone"),
		lang.X("clone.cancel", "Cancel"), container.NewVBox(label, c), func(ok bool) {
			if ok {
				go func() {
					uuid := uuid.NewString()
					tname := fmt.Sprintf(lang.X("clone.msg", "Clone of '%s'"), m.Name)
					Gui.TasksInfos.AddTask(uuid, tname, "")
					OpenTaskDetails()
					ResetStatus()
					snapshotname := fmt.Sprintf("Linked base for %s and %s", m.Name, name.Text)
					var err, errSnap error
					if linkMap[cloneType.SelectedIndex()] == vm.CloneOption_link {
						err = m.TakeSnapshot(&s.Client, snapshotname, "", false, nil)
						errSnap = err
					}
					if err == nil {
						err = m.CloneVm(s, name.Text, cloneModeMap[cloneModeType.SelectedIndex()], linkMap[cloneType.SelectedIndex()],
							macMap[mac.SelectedIndex()], diskMap[disk.Checked], hwIdMap[hwuuid.Checked], snapshotname, util.WriterFunc(func(p []byte) (int, error) {
								Gui.TasksInfos.UpdateTaskStatus(uuid, string(p), true)
								return len(p), nil
							}))
					}
					if err != nil {
						t := fmt.Sprintf(lang.X("clone.done.error", "Clone of '%s' failed"), m.Name)
						SetStatusText(t, MsgError)
						Gui.TasksInfos.AbortTask(uuid, t, false)
					} else {
						t := fmt.Sprintf(lang.X("clone.done.ok", "Clone '%s' of '%s' was created"), name.Text, m.Name)
						Gui.TasksInfos.FinishTask(uuid, t, false)
						SendNotification(lang.X("clone.notification.title", "Clone was made"), t)
						treeUpdateVmList(s.UUID)
					}
					if errSnap == nil && linkMap[cloneType.SelectedIndex()] == vm.CloneOption_link {
						Gui.VmSnapshotTab.updateAfterSnapshotAction(s, m)
					}
				}()
			}
		}, Gui.MainWindow)

	var windowScale float32 = 0.55
	si := Gui.MainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
	dia.Show()
	cloneType.SetSelectedIndex(0)
	cloneModeType.SetSelectedIndex(0)
	mac.SetSelectedIndex(1)
	Gui.MainWindow.Canvas().Focus(name)
}
