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
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UsbAttachItem struct {
	vm.UsbDevice
	isAttached bool
}

type UsbAttachTab struct {
	list    *widget.List
	tabItem *container.TabItem

	toolAttach  *widget.ToolbarAction
	toolDetach  *widget.ToolbarAction
	toolRefresh *widget.ToolbarAction

	toolBar *widget.Toolbar

	devices []*UsbAttachItem

	selectedItem *UsbAttachItem
	usbEnabled   bool
}

var _ DetailsInterface = (*UsbAttachTab)(nil)

func NewUsbAttachTab() *UsbAttachTab {
	usbTab := UsbAttachTab{}

	// formWidth := util.GetFormWidth()
	usbTab.list = widget.NewList(usbTab.listLength, usbTab.listCreateItem, usbTab.listUpdateItem)
	usbTab.list.OnSelected = usbTab.listSelected
	usbTab.list.OnUnselected = usbTab.listUnSelected

	usbTab.toolAttach = widget.NewToolbarAction(theme.ContentAddIcon(), usbTab.listAttach)
	usbTab.toolDetach = widget.NewToolbarAction(theme.ContentRemoveIcon(), usbTab.listDetach)
	usbTab.toolRefresh = widget.NewToolbarAction(theme.ViewRefreshIcon(), usbTab.listRefresh)

	usbTab.toolBar = widget.NewToolbar(usbTab.toolAttach, usbTab.toolDetach, widget.NewToolbarSeparator(), usbTab.toolRefresh)

	gridWrap := container.NewBorder(usbTab.toolBar, nil, nil, util.NewFiller(32, 0), usbTab.list)

	usbTab.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.usbattach", "Attach USB"), gridWrap)
	usbTab.updateToolbarButtons()
	return &usbTab
}

func (usb *UsbAttachTab) listAttach() {
	s, v := getActiveServerAndVm()
	if v == nil || s == nil || usb.selectedItem == nil || usb.selectedItem.isAttached {
		return
	}
	ResetStatus()
	err := v.AttachUsbDevice(&s.Client, usb.selectedItem.UUID, "", VMStatusUpdateCallBack)
	if err != nil {
		SetStatusText(fmt.Sprintf(lang.X("details.vm_usb.attachusb.error", "Attach USB device to VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
	} else {
		usb.selectedItem.isAttached = true
	}
	usb.list.Refresh()
}

func (usb *UsbAttachTab) listDetach() {
	s, v := getActiveServerAndVm()
	if v == nil || s == nil || usb.selectedItem == nil || !usb.selectedItem.isAttached {
		return
	}
	ResetStatus()
	err := v.DetachUsbDevice(&s.Client, usb.selectedItem.UUID, VMStatusUpdateCallBack)
	if err != nil {
		SetStatusText(fmt.Sprintf(lang.X("details.vm_usb.detachusb.error", "Detach USB device from VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
	} else {
		usb.selectedItem.isAttached = false
	}
	usb.list.Refresh()
}

func (usb *UsbAttachTab) listRefresh() {
	index := usb.updateDevices()
	usb.list.Refresh()
	usb.list.UnselectAll()
	if index >= 0 {
		usb.list.Select(index)
	}
}

func (usb *UsbAttachTab) updateDevices() int {
	s, v := getActiveServerAndVm()
	devices, err := s.GetUsbDevices()
	if err != nil {
		return -1
	}

	usb.devices = usb.devices[:0]
	selectedIndex := -1
	for index, item := range devices {
		usbItem := UsbAttachItem{
			UsbDevice: item,
		}
		if usb.selectedItem != nil && usbItem.UUID == usb.selectedItem.UUID {
			selectedIndex = index
		}
		usb.devices = append(usb.devices, &usbItem)
	}

	index := 1
	for {
		str, ok := v.Properties[fmt.Sprintf("USBAttachActive%d", index)]
		if !ok {
			break
		}
		item := usb.getItem(str)
		if item != nil {
			item.isAttached = true
		}
		index++
	}
	return selectedIndex
}

func (usb *UsbAttachTab) listSelected(id widget.ListItemID) {
	usb.selectedItem = usb.devices[id]
	usb.updateToolbarButtons()
}

func (usb *UsbAttachTab) listUnSelected(id widget.ListItemID) {
	usb.selectedItem = nil
	usb.updateToolbarButtons()
}

func (usb *UsbAttachTab) listLength() int {
	return len(usb.devices)
}

func (usb *UsbAttachTab) listCreateItem() fyne.CanvasObject {
	icon := canvas.NewImageFromResource(theme.QuestionIcon())
	icon.SetMinSize(fyne.NewSize(16, 16))
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	text := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	text.Refresh()

	return container.NewHBox(util.NewFiller(6, 0), icon, util.NewFiller(theme.Padding(), 0), text)
}

func (usb *UsbAttachTab) listUpdateItem(id widget.ListItemID, o fyne.CanvasObject) {
	c, ok := o.(*fyne.Container)
	if !ok {
		return
	}

	text, ok := c.Objects[3].(*canvas.Text)
	if !ok {
		return
	}

	icon, ok := c.Objects[1].(*canvas.Image)
	if !ok {
		return
	}

	item := usb.devices[id]
	if item.isAttached {
		icon.Resource = theme.CheckButtonCheckedIcon()
	} else {
		icon.Resource = theme.CheckButtonIcon()
	}

	icon.SetMinSize(fyne.NewSize(16, 16))
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	text.Text = item.Name
	text.Refresh()
}

func (usb *UsbAttachTab) getItem(uuid string) *UsbAttachItem {
	for _, item := range usb.devices {
		if item.UUID == uuid {
			return item
		}
	}
	return nil
}

// calles by selection change
func (usb *UsbAttachTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()

	usb.selectedItem = nil
	usb.list.UnselectAll()

	if s == nil || v == nil {
		usb.DisableAll()
		return
	}

	var usb1 bool
	var usb2 bool
	var usb3 bool

	str := v.Properties["usb1"]
	if str == "on" {
		usb1 = true
	}
	str = v.Properties["usb2"]
	if str == "on" {
		usb2 = true
	}
	str = v.Properties["usb3"]
	if str == "on" {
		usb3 = true
	}
	usb.usbEnabled = usb1 || usb2 || usb3
	if !usb.usbEnabled {
		usb.DisableAll()
		return
	}

	usb.updateDevices()
	usb.list.Refresh()
	usb.updateToolbarButtons()
}

// called from status updates
func (usb *UsbAttachTab) UpdateByStatus() {
	usb.updateToolbarButtons()
}

func (usb *UsbAttachTab) DisableAll() {
	usb.toolAttach.Disable()
	usb.toolDetach.Disable()
	usb.toolRefresh.Disable()
}

func (usb *UsbAttachTab) updateToolbarButtons() {
	_, v := getActiveServerAndVm()
	if v == nil || !usb.usbEnabled {
		usb.toolAttach.Disable()
		usb.toolDetach.Disable()
		usb.toolRefresh.Disable()
	} else {
		state, err := v.GetState()
		if err != nil {
			return
		}

		usb.toolRefresh.Enable()
		if state == vm.RunState_paused || state == vm.RunState_running {
			if usb.selectedItem == nil || !usb.selectedItem.isAttached {
				usb.toolDetach.Disable()
				usb.toolAttach.Enable()
			} else {
				usb.toolDetach.Enable()
				usb.toolAttach.Disable()
			}
		} else {
			usb.toolAttach.Disable()
			usb.toolDetach.Disable()
		}
	}
}

func (usb *UsbAttachTab) Apply() {
}
