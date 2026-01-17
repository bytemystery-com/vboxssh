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
	"slices"

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
	"github.com/bytemystery-com/colorlabel"
)

type oldUsbType struct {
	usbType vm.UsbType
	filters []*UsbFilter
}

type UsbFilter struct {
	name         string
	isChecked    bool
	productId    string
	vendorId     string
	serialNumber string
	product      string
	manufacturer string
}

type UsbTab struct {
	oldValues oldUsbType

	enabled *widget.Check
	usbType *widget.Select
	list    *widget.List
	up      *widget.Button
	down    *widget.Button

	apply   *widget.Button
	tabItem *container.TabItem

	toolNew    *widget.ToolbarAction
	toolAdd    *widget.ToolbarAction
	toolEdit   *widget.ToolbarAction
	toolDelete *widget.ToolbarAction
	toolCheck  *widget.ToolbarAction

	toolBar *widget.Toolbar
	label   *colorlabel.ColorLabel

	selectedItem *UsbFilter

	filters []*UsbFilter

	mapIndexToValue map[int]vm.UsbType
}

var _ DetailsInterface = (*UsbTab)(nil)

func NewUsbTab() *UsbTab {
	usbTab := UsbTab{}
	usbTab.filters = make([]*UsbFilter, 0, 3)
	usbTab.mapIndexToValue = map[int]vm.UsbType{-1: vm.Usb_none, 0: vm.Usb_1, 1: vm.Usb_2, 2: vm.Usb_3}

	usbTab.apply = widget.NewButton(lang.X("details.vm_usb.apply", "Apply"), func() {
		usbTab.Apply()
	})
	usbTab.apply.Importance = widget.HighImportance

	formWidth := util.GetFormWidth()
	usbTab.enabled = widget.NewCheck(lang.X("details.vm_usb.enabled", "Enable USB"), func(checked bool) {
		usbTab.UpdateByStatus()
		usbTab.updateToolbarButtons()
	})
	usbTab.usbType = widget.NewSelect([]string{
		lang.X("details.vm_usb.usb1", "USB 1.1 (OHCI)"),
		lang.X("details.vm_usb.usb2", "USB 2.0 (OHCI + EHCI)"),
		lang.X("details.vm_usb.usb3", "USB 3.0 xHCI)"),
	}, nil)

	usbTab.list = widget.NewList(usbTab.listLength, usbTab.listCreateItem, usbTab.listUpdateItem)
	usbTab.list.OnSelected = usbTab.listSelected
	usbTab.list.OnUnselected = usbTab.listUnSelected

	usbTab.up = widget.NewButtonWithIcon("", theme.MoveUpIcon(), usbTab.listMoveUp)
	usbTab.down = widget.NewButtonWithIcon("", theme.MoveDownIcon(), usbTab.listMoveDown)

	grid1 := container.New(layout.NewFormLayout(),
		usbTab.enabled, util.NewFiller(0, 0),
		widget.NewLabel(lang.X("details.vm_usb.type", "USB type")), usbTab.usbType,
	)
	gridWrap1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)

	usbTab.toolCheck = widget.NewToolbarAction(theme.CheckButtonCheckedIcon(), usbTab.listCheck)
	usbTab.toolNew = widget.NewToolbarAction(theme.ContentAddIcon(), usbTab.listNew)
	usbTab.toolAdd = widget.NewToolbarAction(theme.DocumentIcon(), usbTab.listAdd)
	usbTab.toolEdit = widget.NewToolbarAction(theme.DocumentCreateIcon(), usbTab.listEdit)
	usbTab.toolDelete = widget.NewToolbarAction(theme.DeleteIcon(), usbTab.listDelete)
	usbTab.toolBar = widget.NewToolbar(usbTab.toolCheck, widget.NewToolbarSeparator(), usbTab.toolNew, usbTab.toolAdd, usbTab.toolEdit, usbTab.toolDelete)
	usbTab.label = colorlabel.NewColorLabel(lang.X("details.vm_usb.filters.title", "USB device filters"), theme.ColorNamePrimary, nil, 1.0)

	gridWrap := container.NewBorder(container.NewVBox(gridWrap1, usbTab.label, usbTab.toolBar),
		nil, nil, container.NewHBox(container.NewVBox(layout.NewSpacer(), usbTab.up, usbTab.down, layout.NewSpacer(), usbTab.apply, layout.NewSpacer(), layout.NewSpacer()), util.NewFiller(32, 0)), usbTab.list)

	usbTab.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.usb", "USB"), gridWrap)
	usbTab.updateToolbarButtons()
	return &usbTab
}

func (usb *UsbTab) listNew() {
	f := UsbFilter{
		name: lang.X("details.vm_info.tab.usb.newfilter.name", "New filter"),
	}
	usb.filters = append(usb.filters, &f)
	usb.list.Refresh()
	usb.list.Select(len(usb.filters) - 1)
}

func (usb *UsbTab) listAdd() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}

	usbList, err := s.GetUsbDevices()
	if err != nil {
		return
	}

	usbSelect := func(index int) {
		item := usbList[index]
		f := UsbFilter{
			name:         item.Name,
			productId:    item.ProductId,
			vendorId:     item.VendorId,
			serialNumber: item.SerialNumber,
			product:      item.Product,
			manufacturer: item.Manufacturer,
			isChecked:    true,
		}
		usb.filters = append(usb.filters, &f)
		usb.list.Refresh()
		usb.list.Select(len(usb.filters) - 1)
	}

	menuItems := make([]*fyne.MenuItem, 0, len(usbList))
	for index, item := range usbList {
		menuItems = append(menuItems, fyne.NewMenuItem(item.Name, func() { usbSelect(index) }))
	}

	menu := fyne.NewMenu("XXX", menuItems...)
	popup := widget.NewPopUpMenu(menu, Gui.MainWindow.Canvas())

	popup.ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(usb.toolBar))
	usb.list.Refresh()
}

func (usb *UsbTab) listDelete() {
	if usb.selectedItem == nil {
		return
	}
	index := slices.Index(usb.filters, usb.selectedItem)
	if index < 0 {
		return
	}
	usb.filters = append(usb.filters[:index], usb.filters[index+1:]...)
	usb.list.Refresh()
	if index >= len(usb.filters) {
		index = len(usb.filters) - 1
	}
	usb.list.UnselectAll()
	if len(usb.filters) > 0 {
		usb.list.Select(index)
	}
	usb.updateToolbarButtons()
}

func (usb *UsbTab) listEdit() {
	active := widget.NewCheck(lang.X("details.vm_usb.filter.edit.active", "Active"), nil)

	name := widget.NewEntry()
	name.SetPlaceHolder(lang.X("details.vm_usb.filter.edit.name.placeholder", "Name of the filter"))

	productid := widget.NewEntry()
	productid.SetPlaceHolder(lang.X("details.vm_usb.filter.edit.productid.placeholder", "Product ID of device"))

	vendorid := widget.NewEntry()
	vendorid.SetPlaceHolder(lang.X("details.vm_usb.filter.edit.vendoridplaceholder", "Vendor ID of device"))

	manufacturer := widget.NewEntry()
	manufacturer.SetPlaceHolder(lang.X("details.vm_usb.filter.edit.manufacturer.placeholder", "Manufacturer of device"))

	product := widget.NewEntry()
	product.SetPlaceHolder(lang.X("details.vm_usb.filter.edit.product.placeholder", "Product name of device"))

	serialNumber := widget.NewEntry()
	serialNumber.SetPlaceHolder(lang.X("details.vm_usb.filter.edit.serialNumber.placeholder", "Serial number of device"))

	if usb.selectedItem != nil {
		active.SetChecked(usb.selectedItem.isChecked)
		name.Text = usb.selectedItem.name
		productid.Text = usb.selectedItem.productId
		vendorid.Text = usb.selectedItem.vendorId
		manufacturer.Text = usb.selectedItem.manufacturer
		product.Text = usb.selectedItem.product
		serialNumber.Text = usb.selectedItem.serialNumber
	}

	grid1 := container.New(layout.NewFormLayout(),
		active, util.NewFiller(0, 0),
		widget.NewLabel(lang.X("details.vm_usb.filter.edit.name", "Name")), name,
		widget.NewLabel(lang.X("details.vm_usb.filter.edit.vendorid", "Vendor ID")), vendorid,
		widget.NewLabel(lang.X("details.vm_usb.filter.edit.productid", "Product ID")), productid,
		widget.NewLabel(lang.X("details.vm_usb.filter.edit.manufacturer", "Manufacturer")), manufacturer,
		widget.NewLabel(lang.X("details.vm_usb.filter.edit.product", "Product")), product,
		widget.NewLabel(lang.X("details.vm_usb.filter.edit.serialnumber", "Serial No.")), serialNumber,
	)

	dia := dialog.NewCustomConfirm(lang.X("details.vm_usb.filter.edit.title", "USB filter details"),
		lang.X("details.vm_usb.filter.edit.ok", "Ok"),
		lang.X("details.vm_usb.filter.edit.cancel", "Cancel"), grid1, func(ok bool) {
			usb.selectedItem.isChecked = active.Checked
			usb.selectedItem.name = name.Text
			usb.selectedItem.productId = productid.Text
			usb.selectedItem.vendorId = vendorid.Text
			usb.selectedItem.manufacturer = manufacturer.Text
			usb.selectedItem.product = product.Text
			usb.selectedItem.serialNumber = serialNumber.Text
			usb.list.Refresh()
		}, Gui.MainWindow)
	dia.Show()

	si := Gui.MainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width*.45, si.Height*.60))
	dia.Show()
}

func (usb *UsbTab) listMoveUp() {
	if usb.selectedItem == nil {
		return
	}

	index := slices.Index(usb.filters, usb.selectedItem)
	if index <= 0 {
		return
	}
	usb.filters[index], usb.filters[index-1] = usb.filters[index-1], usb.filters[index]
	usb.list.Refresh()
	usb.list.Select(index - 1)
}

func (usb *UsbTab) listMoveDown() {
	if usb.selectedItem == nil {
		return
	}

	index := slices.Index(usb.filters, usb.selectedItem)
	if index >= len(usb.filters)-1 {
		return
	}
	usb.filters[index], usb.filters[index+1] = usb.filters[index+1], usb.filters[index]
	usb.list.Refresh()
	usb.list.Select(index + 1)
}

func (usb *UsbTab) listCheck() {
	if usb.selectedItem == nil {
		return
	}
	usb.selectedItem.isChecked = !usb.selectedItem.isChecked
	usb.list.Refresh()
}

func (usb *UsbTab) listSelected(id widget.ListItemID) {
	usb.selectedItem = usb.filters[id]
	usb.updateToolbarButtons()
}

func (usb *UsbTab) listUnSelected(id widget.ListItemID) {
	usb.selectedItem = nil
	usb.updateToolbarButtons()
}

func (usb *UsbTab) listLength() int {
	return len(usb.filters)
}

func (usb *UsbTab) listCreateItem() fyne.CanvasObject {
	icon := canvas.NewImageFromResource(theme.QuestionIcon())
	icon.SetMinSize(fyne.NewSize(16, 16))
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	text := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	text.Refresh()

	return container.NewHBox(util.NewFiller(6, 0), icon, util.NewFiller(theme.Padding(), 0), text)
}

func (usb *UsbTab) listUpdateItem(id widget.ListItemID, o fyne.CanvasObject) {
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

	filter := usb.filters[id]
	if filter.isChecked {
		icon.Resource = theme.CheckButtonCheckedIcon()
	} else {
		icon.Resource = theme.CheckButtonIcon()
	}

	icon.SetMinSize(fyne.NewSize(16, 16))
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	text.Text = filter.name
	text.Refresh()
}

// calles by selection change
func (usb *UsbTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()

	if s == nil || v == nil {
		usb.DisableAll()
		return
	}
	usb.apply.Enable()

	err := v.UpdateStatusEx(&s.Client)
	if err != nil {
		return
	}

	var usb1 bool
	var usb2 bool
	var usb3 bool
	var usbEnabled bool
	var usbType int

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

	usbEnabled = usb1 || usb2 || usb3
	usbType = -1
	if usb3 {
		usbType = 2
	} else if usb2 {
		usbType = 1
	} else if usb1 {
		usbType = 0
	}

	// Values
	// Enabled
	if usbEnabled {
		usb.enabled.SetChecked(true)
		usb.enableDisableUsbCtrls(true)
	} else {
		usb.enabled.SetChecked(false)
		usb.enableDisableUsbCtrls(false)
		usb.oldValues.usbType = vm.Usb_none
	}

	usb.oldValues.usbType = usb.mapIndexToValue[usbType]
	if usbType >= 0 {
		usb.usbType.SetSelectedIndex(usbType)
	} else {
		usb.oldValues.usbType = vm.Usb_none
	}

	// Filters
	usb.selectedItem = nil
	usb.list.UnselectAll()
	index := 1
	usb.filters = usb.filters[:0]
	for {
		str, ok := v.Properties[fmt.Sprintf("USBFilterActive%d", index)]

		if !ok {
			break
		}
		filter := UsbFilter{}
		filter.isChecked = (str == "on")
		filter.name = v.Properties[fmt.Sprintf("USBFilterName%d", index)]
		filter.productId = v.Properties[fmt.Sprintf("USBFilterProductId%d", index)]
		filter.vendorId = v.Properties[fmt.Sprintf("USBFilterVendorId%d", index)]
		filter.serialNumber = v.Properties[fmt.Sprintf("USBFilterSerialNumber%d", index)]
		filter.product = v.Properties[fmt.Sprintf("USBFilterProduct%d", index)]
		filter.manufacturer = v.Properties[fmt.Sprintf("USBFilterManufacturer%d", index)]
		usb.filters = append(usb.filters, &filter)
		index++
	}
	usb.saveOldFilterConfig()

	usb.list.Refresh()
	if len(usb.filters) > 0 {
		usb.list.Select(0)
	}

	usb.UpdateByStatus()
	usb.updateToolbarButtons()
}

// called from status updates
func (usb *UsbTab) UpdateByStatus() {
	_, v := getActiveServerAndVm()
	if v != nil {
		state, err := v.GetState()
		if err != nil {
			return
		}
		switch state {
		case vm.RunState_unknown, vm.RunState_meditation, vm.RunState_running, vm.RunState_paused, vm.RunState_saved:
			usb.DisableAll()

		case vm.RunState_off, vm.RunState_aborted:
			usb.enabled.Enable()
			usb.enableDisableUsbCtrls(usb.enabled.Checked)

		default:
			SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
		}
	} else {
		usb.DisableAll()
	}
}

func (usb *UsbTab) enableDisableUsbCtrls(enable bool) {
	if enable {
		usb.usbType.Enable()
	} else {
		usb.usbType.Disable()
	}
}

func (usb *UsbTab) DisableAll() {
	usb.enableDisableUsbCtrls(false)
	usb.enabled.Disable()
	usb.usbType.Disable()

	usb.apply.Disable()
}

func (usb *UsbTab) saveOldFilterConfig() {
	usb.oldValues.filters = make([]*UsbFilter, 0, len(usb.filters))
	for _, item := range usb.filters {
		newItem := UsbFilter{}
		newItem = *item
		usb.oldValues.filters = append(usb.oldValues.filters, &newItem)
	}
}

func (usb *UsbTab) updateToolbarButtons() {
	_, v := getActiveServerAndVm()
	if usb.enabled.Checked {
		usb.list.Show()
		usb.toolBar.Show()
		usb.up.Show()
		usb.down.Show()
		usb.label.Show()
	} else {
		usb.list.Hide()
		usb.toolBar.Hide()
		usb.up.Hide()
		usb.down.Hide()
		usb.label.Hide()
	}
	if v == nil || !usb.enabled.Checked {
		usb.up.Disable()
		usb.down.Disable()
		usb.toolDelete.Disable()
		usb.toolEdit.Disable()
		usb.toolAdd.Disable()
		usb.toolNew.Disable()
		usb.toolCheck.Disable()
		return
	} else {
		usb.toolAdd.Enable()
		usb.toolNew.Enable()
	}
	if usb.selectedItem == nil {
		usb.up.Disable()
		usb.down.Disable()
		usb.toolDelete.Disable()
		usb.toolEdit.Disable()
		usb.toolCheck.Disable()
	} else {
		usb.toolDelete.Enable()
		usb.toolEdit.Enable()
		usb.toolCheck.Enable()
		index := slices.Index(usb.filters, usb.selectedItem)
		if index < 0 {
			usb.up.Disable()
			usb.down.Disable()
		} else {
			if index == 0 {
				usb.up.Disable()
				usb.down.Enable()
			} else if index == len(usb.filters)-1 {
				usb.up.Enable()
				usb.down.Disable()
			} else {
				usb.up.Enable()
				usb.down.Enable()
			}
		}
	}
}

func (usb *UsbTab) Apply() {
	s, v := getActiveServerAndVm()
	if s != nil && v != nil {
		ResetStatus()
		if !usb.enabled.Disabled() {
			var usbType vm.UsbType
			if !usb.enabled.Checked {
				usbType = usb.mapIndexToValue[-1]
			} else {
				usbType = usb.mapIndexToValue[usb.usbType.SelectedIndex()]
			}
			if usbType != usb.oldValues.usbType {
				go func() {
					err := v.SetUsb(s, usbType, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_usb.setusb.error", "Setting USB for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						usb.oldValues.usbType = usbType
					}
				}()
			}
		}

		go func() {
			// Remove all old
			for range usb.oldValues.filters {
				err := v.RemoveUsbFilter(&s.Client, 0, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_usb.removeusbfilter.error", "Remove USB filter for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				}
			}

			// Add all new
			for index, item := range usb.filters {
				err := v.AddUsbFilter(&s.Client, index, item.name, item.vendorId, item.productId, item.serialNumber, item.product, item.manufacturer, item.isChecked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_usb.addusbfilter.error", "Add USB filter for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				}
			}
			usb.saveOldFilterConfig()
		}()
	}
}
