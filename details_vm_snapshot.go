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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
)

type SnapshotItem struct {
	uuid        string
	name        string
	description string
	childs      []*SnapshotItem
	isCurrent   bool
}

type SnapshotTab struct {
	tree    *widget.Tree
	tabItem *container.TabItem

	toolTake    *widget.ToolbarAction
	toolDelete  *widget.ToolbarAction
	toolRestore *widget.ToolbarAction

	toolBar *widget.Toolbar

	snapshots   []*SnapshotItem
	snapshotMap map[string]*SnapshotItem

	selectedItem *SnapshotItem
}

var _ DetailsInterface = (*SnapshotTab)(nil)

func NewSnapshotTab() *SnapshotTab {
	snapshot := SnapshotTab{
		snapshotMap: make(map[string]*SnapshotItem, 10),
	}

	snapshot.tree = widget.NewTree(snapshot.treeGetChilds, snapshot.treeIsBranch, snapshot.treeCreate, snapshot.treeUpdate)
	snapshot.tree.OnSelected = snapshot.treeSelected
	// snapshot.tree.OnUnselected = snapshot.treeUnselected

	snapshot.toolTake = widget.NewToolbarAction(theme.MediaPhotoIcon(), snapshot.take)
	snapshot.toolDelete = widget.NewToolbarAction(theme.DeleteIcon(), snapshot.delete)
	snapshot.toolRestore = widget.NewToolbarAction(theme.ContentUndoIcon(), snapshot.restore)

	snapshot.toolBar = widget.NewToolbar(snapshot.toolTake, snapshot.toolRestore, snapshot.toolDelete)

	gridWrap := container.NewBorder(snapshot.toolBar, nil, nil, util.NewFiller(32, 0), snapshot.tree)

	snapshot.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.snapshot", "Snapshot"), gridWrap)
	snapshot.updateToolbarButtons()
	return &snapshot
}

func (snap *SnapshotTab) updateAfterSnapshotAction(s *vm.VmServer, v *vm.VMachine) {
	go v.UpdateStatus(&s.Client, func(uuid string) {
		fyne.Do(func() {
			snap.UpdateBySelect()
			Gui.Details.CloseAll()
			Gui.Details.Open(1)
		})
	})
}

func (snap *SnapshotTab) take() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder(lang.X("snapshot.take.name.placeholder", "Name for the snapshot"))
	dia := dialog.NewCustomConfirm(lang.X("snapshot.take.title", "Take a snapshot"),
		lang.X("snapshot.take.ok", "Ok"),
		lang.X("snapshot.take.cancel", "Cancel"),
		container.New(layout.NewFormLayout(),
			widget.NewLabel(lang.X("snapshot.take.name", "Name")), nameEntry,
		), func(ok bool) {
			go func() {
				uuid := uuid.NewString()
				name := fmt.Sprintf(lang.X("snapshot.take.msg", "Take snapshot of '%s'"), v.Name)
				Gui.TasksInfos.AddTask(uuid, name, "")
				OpenTaskDetails()
				err := v.TakeSnapshot(&s.Client, nameEntry.Text, "", false, util.WriterFunc(func(p []byte) (int, error) {
					Gui.TasksInfos.UpdateTaskStatus(uuid, string(p), true)
					return len(p), nil
				}))
				if err != nil {
					t := fmt.Sprintf(lang.X("snapshot.take.done.error", "snapshot '%s' of '%s' failed"), nameEntry.Text, v.Name)
					SetStatusText(t, MsgError)
					Gui.TasksInfos.AbortTask(uuid, t, false)
				} else {
					t := fmt.Sprintf(lang.X("snapshot.take.done.ok", "Snapshot '%s' of '%s' was created"), nameEntry.Text, v.Name)
					Gui.TasksInfos.FinishTask(uuid, t, false)
					SendNotification(lang.X("snapsot.take.notification.title", "Snapshot taken"), t)
					snap.updateAfterSnapshotAction(s, v)
				}
			}()
		}, Gui.MainWindow)
	// si := Gui.MainWindow.Canvas().Size()
	// var windowScale float32 = 1.5
	dia.Resize(fyne.NewSize(dia.MinSize().Height*1.9, dia.MinSize().Height*1.2))
	dia.Show()
	Gui.MainWindow.Canvas().Focus(nameEntry)
}

func (snap *SnapshotTab) restore() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil || snap.selectedItem == nil || snap.selectedItem.isCurrent {
		return
	}
	dialog.ShowConfirm(lang.X("snapsot.restore.title", "Restore snapshot"),
		fmt.Sprintf(lang.X("snapsot.restore.msg", "Do you really want to restore to the snapshot\n'%s' for the virtual machine\n'%s' on the server '%s' ?"),
			snap.selectedItem.name, v.Name, util.GetServerAddressAsString(s.Client.Client)), func(oK bool) {
			if !oK {
				return
			}
			go func() {
				uuid := uuid.NewString()
				name := fmt.Sprintf(lang.X("snapshot.restore.msg", "Restore to snapshot '%s'"), snap.selectedItem.name)
				Gui.TasksInfos.AddTask(uuid, name, "")
				OpenTaskDetails()
				err := v.RestoreSnapshot(&s.Client, snap.selectedItem.uuid, util.WriterFunc(func(p []byte) (int, error) {
					Gui.TasksInfos.UpdateTaskStatus(uuid, string(p), true)
					return len(p), nil
				}))
				if err != nil {
					t := fmt.Sprintf(lang.X("snapshot.restore.done.error", "Restoring to snapshot '%s' of '%s' failed"), snap.selectedItem, v.Name)
					SetStatusText(t, MsgError)
					Gui.TasksInfos.AbortTask(uuid, t, false)
				} else {
					t := fmt.Sprintf(lang.X("snapshot.restore.done.ok", "Restored to snapshot '%s' of '%s'"), snap.selectedItem.name, v.Name)
					Gui.TasksInfos.FinishTask(uuid, t, false)
					SendNotification(lang.X("snapsot.restore.notification.title", "Snapshot restored"), t)
					snap.updateAfterSnapshotAction(s, v)
				}
			}()
		}, Gui.MainWindow)
}

func (snap *SnapshotTab) delete() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil || snap.selectedItem == nil || snap.selectedItem.isCurrent {
		return
	}

	dialog.ShowConfirm(lang.X("snapsot.delete.title", "Delete snapshot"),
		fmt.Sprintf(lang.X("snapsot.delete.msg", "Do you really want to delete the snapshot\n'%s' tof the virtual machine\n'%s' on the server '%s' ?"),
			snap.selectedItem.name, v.Name, util.GetServerAddressAsString(s.Client.Client)), func(oK bool) {
			if !oK {
				return
			}
			go func() {
				uuid := uuid.NewString()
				name := fmt.Sprintf(lang.X("snapshot.delete.msg", "Delete snapshot '%s'"), snap.selectedItem.name)
				Gui.TasksInfos.AddTask(uuid, name, "")
				OpenTaskDetails()
				err := v.DeleteSnapshot(&s.Client, snap.selectedItem.uuid, util.WriterFunc(func(p []byte) (int, error) {
					Gui.TasksInfos.UpdateTaskStatus(uuid, string(p), true)
					return len(p), nil
				}))
				if err != nil {
					t := fmt.Sprintf(lang.X("snapshot.delete.done.error", "Deleting snapshot '%s' of '%s' failed"), snap.selectedItem, name, v.Name)
					SetStatusText(t, MsgError)
					Gui.TasksInfos.AbortTask(uuid, t, false)
				} else {
					t := fmt.Sprintf(lang.X("snapshot.delete.done.ok", "Snapshot '%s' was deletd from '%s'"), snap.selectedItem.name, v.Name)
					Gui.TasksInfos.FinishTask(uuid, t, false)
					SendNotification(lang.X("snapsot.delete.notification.title", "Snapshot deleted"), t)
					snap.updateAfterSnapshotAction(s, v)
				}
			}()
		}, Gui.MainWindow)
}

func (snap *SnapshotTab) treeSelected(id widget.TreeNodeID) {
	snap.selectedItem = snap.snapshotMap[id]
	snap.updateToolbarButtons()
}

func (snap *SnapshotTab) treeUnSelected(id widget.TreeNodeID) {
	snap.selectedItem = nil
	snap.updateToolbarButtons()
}

func (snap *SnapshotTab) treeGetChilds(id widget.TreeNodeID) []widget.TreeNodeID {
	if id == "" {
		if len(snap.snapshots) > 0 {
			return []widget.TreeNodeID{snap.snapshots[0].uuid}
		} else {
			return nil
		}
	}
	s := snap.snapshotMap[id]
	if s == nil || len(s.childs) <= 0 {
		return nil
	}
	list := make([]widget.TreeNodeID, 0, len(s.childs))
	for _, item := range s.childs {
		list = append(list, item.uuid)
	}
	return list
}

func (snap *SnapshotTab) treeIsBranch(id widget.TreeNodeID) bool {
	if id == "" {
		if len(snap.snapshots) > 0 {
			return true
		} else {
			return false
		}
	}
	s := snap.snapshotMap[id]
	if s == nil {
		return false
	}
	return len(s.childs) > 0
}

func (snap *SnapshotTab) treeCreate(isBranche bool) fyne.CanvasObject {
	text := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	text.Refresh()
	return text
}

func (snap *SnapshotTab) treeUpdate(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
	text, ok := o.(*canvas.Text)
	if !ok {
		return
	}
	s := snap.snapshotMap[id]
	if s == nil {
		return
	}

	text.Text = s.name
	if s.isCurrent {
		text.Color = theme.Color(theme.ColorNamePrimary)
		text.TextStyle = fyne.TextStyle{
			Bold: true,
		}
	} else {
		text.TextStyle = fyne.TextStyle{}
		text.Color = theme.Color(theme.ColorNameForeground)
	}

	text.Refresh()
}

func (snap *SnapshotTab) getChilds(v *vm.VMachine, token string, snapMap map[string]*SnapshotItem) []*SnapshotItem {
	index := 1
	list := make([]*SnapshotItem, 0, 10)
	for {
		token2 := fmt.Sprintf("%s-%d", token, index)
		str, ok := v.Properties["SnapshotName"+token2]
		if !ok {
			break
		}
		uuid, ok := v.Properties["SnapshotUUID"+token2]
		if !ok {
			break
		}
		description := v.Properties["SnapshotDescription"+token2]

		childs := snap.getChilds(v, token2, snapMap)
		newItem := SnapshotItem{
			name:        str,
			description: description,
			uuid:        uuid,
			childs:      childs,
		}
		list = append(list, &newItem)
		snapMap[uuid] = &newItem

		index++
	}
	return list
}

// calles by selection change
func (snap *SnapshotTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}
	ss := SnapshotItem{}
	ss.name = lang.X("details.vm_snapshot.current", "Current state")
	ss.uuid = v.Properties["CurrentSnapshotUUID"]
	ss.isCurrent = true

	snap.snapshots = snap.snapshots[:0]
	clear(snap.snapshotMap)

	str, ok := v.Properties["SnapshotName"]
	if ok {
		newItem := SnapshotItem{
			name: str,
			uuid: v.Properties["SnapshotUUID"],
		}
		newItem.childs = snap.getChilds(v, "", snap.snapshotMap)
		snap.snapshots = append(snap.snapshots, &newItem)
		snap.snapshotMap[newItem.uuid] = &newItem
	}

	p := snap.snapshotMap[ss.uuid]
	ss.uuid = uuid.NewString()
	if p != nil {
		p.childs = append(p.childs, &ss)
	} else {
		snap.snapshots = append(snap.snapshots, &ss)
	}
	snap.snapshotMap[ss.uuid] = &ss

	snap.tree.Refresh()
	snap.tree.OpenAllBranches()

	snap.updateToolbarButtons()
}

// called from status updates
func (snap *SnapshotTab) UpdateByStatus() {
	snap.updateToolbarButtons()
}

func (snap *SnapshotTab) DisableAll() {
	snap.toolTake.Disable()
	snap.toolRestore.Disable()
	snap.toolDelete.Disable()
}

func (snap *SnapshotTab) updateToolbarButtons() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		snap.toolDelete.Disable()
		snap.toolRestore.Disable()
		snap.toolTake.Disable()
		return
	}
	state, err := v.GetState()
	if err == nil {
		if (state == vm.RunState_aborted || state == vm.RunState_off) &&
			snap.selectedItem != nil && !snap.selectedItem.isCurrent {
			snap.toolRestore.Enable()
		} else {
			snap.toolRestore.Disable()
		}
	} else {
		snap.toolRestore.Disable()
	}
	if snap.selectedItem == nil {
		snap.toolDelete.Disable()
	} else {
		if snap.selectedItem.isCurrent {
			snap.toolDelete.Disable()
		} else {
			snap.toolDelete.Enable()
		}
	}
}

func (snap *SnapshotTab) Apply() {
}
