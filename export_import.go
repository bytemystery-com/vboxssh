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
	"regexp"
	"strings"

	"bytemystery-com/vboxssh/filebrowser"
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

type ExportTabItem struct {
	tab         *container.TabItem
	name        *widget.Entry
	product     *widget.Entry
	productURL  *widget.Entry
	vendor      *widget.Entry
	vendorURL   *widget.Entry
	version     *widget.Entry
	description *widget.Entry
	license     *widget.Entry
}

type VmExport struct {
	*vm.VMachine
	isSelected   bool
	isExportable bool
}

type ExportHelper struct {
	browse               *widget.Button
	file                 *widget.Entry
	list                 *widget.List
	mnifest              *widget.Check
	iso                  *widget.Check
	format               *widget.Select
	mac                  *widget.Select
	vms                  []*VmExport
	vmServer             *vm.VmServer
	tabs                 []*ExportTabItem
	formatMapIndexToType map[int]vm.OvaFormatType
	macMapIndexToType    map[int]vm.MacExportType
}

func NewExportHelper(s *vm.VmServer) *ExportHelper {
	return &ExportHelper{
		vmServer:             s,
		formatMapIndexToType: map[int]vm.OvaFormatType{0: vm.OvaFormat_legacy, 1: vm.OvaFormat_0_9, 2: vm.OvaFormat_1_0, 3: vm.OvaFormat_2_0},
		macMapIndexToType:    map[int]vm.MacExportType{0: vm.MacExport_nomacs, 1: vm.MacExport_nomacsbutnat, 3: vm.MacExport_all},
	}
}

func (e *ExportHelper) updateStatus() {
	for index, item := range e.vms {
		err := item.UpdateStatus(&e.vmServer.Client, nil)
		if err == nil {
			state, err := item.GetState()
			if err == nil {
				fyne.Do(func() {
					item.isExportable = (state == vm.RunState_off || state == vm.RunState_aborted || state == vm.RunState_saved)
					e.list.RefreshItem(index)
				})
			}
		}
	}
}

func (e *ExportHelper) createExport2Tab(v *vm.VMachine) *ExportTabItem {
	etab := ExportTabItem{}
	// formWidth := util.GetFormWidth()
	etab.name = widget.NewEntry()
	etab.name.SetText(v.Name)
	etab.name.SetPlaceHolder(lang.X("export.name.placeholder", "Name"))
	etab.product = widget.NewEntry()
	etab.product.SetPlaceHolder(lang.X("export.product.placeholder", "Product informations"))
	etab.productURL = widget.NewEntry()
	etab.productURL.SetPlaceHolder(lang.X("export.producturl.placeholder", "Product URL"))
	etab.vendor = widget.NewEntry()
	etab.vendor.SetPlaceHolder(lang.X("export.vendor.placeholder", "Vendor informations"))
	etab.vendorURL = widget.NewEntry()
	etab.vendorURL.SetPlaceHolder(lang.X("export.vendorURL.placeholder", "Vendor URL"))
	etab.version = widget.NewEntry()
	etab.version.SetPlaceHolder(lang.X("export.version.placeholder", "Version informations"))
	etab.description = widget.NewEntry()
	etab.description.SetPlaceHolder(lang.X("export.description.placeholder", "Further description"))
	description, ok := v.Properties["description"]
	if ok {
		etab.description.SetText(description)
	}
	etab.license = widget.NewEntry()
	etab.license.SetPlaceHolder(lang.X("export.description.license", "License informations"))

	grid := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("export.vsys.name", "Name")), etab.name,
		widget.NewLabel(lang.X("export.vsys.product", "Product")), etab.product,
		widget.NewLabel(lang.X("export.vsys.producturl", "Product URL")), etab.productURL,
		widget.NewLabel(lang.X("export.vsys.vendor", "Vendor")), etab.vendor,
		widget.NewLabel(lang.X("export.vsys.vendorurl", "Vendor URL")), etab.vendorURL,
		widget.NewLabel(lang.X("export.vsys.version", "Version")), etab.version,
		widget.NewLabel(lang.X("export.vsys.description", "Description")), etab.description,
		widget.NewLabel(lang.X("export.vsys.license", "License")), etab.license,
	)
	gridWrap1 := container.NewBorder(nil, nil, nil, nil, grid)
	gridWrap := container.NewVBox(util.NewVFiller(1.0), gridWrap1, util.NewVFiller(1.0))
	etab.tab = container.NewTabItem(v.Name, gridWrap)
	return &etab
}

func (e *ExportHelper) buildVsys() []string {
	vsys := make([]string, 0, 30)
	index := 0
	for _, item := range e.vms {
		if item.isSelected {
			vsys = append(vsys, fmt.Sprintf("--vsys=%d", index))
			etab := e.tabs[index]
			if etab.name.Text != "" {
				vsys = append(vsys, fmt.Sprintf("--vmname='%s'", etab.name.Text))
			}
			if etab.product.Text != "" {
				vsys = append(vsys, fmt.Sprintf("--product='%s'", etab.product.Text))
			}
			if etab.productURL.Text != "" {
				vsys = append(vsys, fmt.Sprintf("--producturl='%s'", etab.productURL.Text))
			}
			if etab.productURL.Text != "" {
				vsys = append(vsys, fmt.Sprintf("--producturl='%s'", etab.productURL.Text))
			}
			if etab.vendor.Text != "" {
				vsys = append(vsys, fmt.Sprintf("--vendor='%s'", etab.vendor.Text))
			}
			if etab.vendorURL.Text != "" {
				vsys = append(vsys, fmt.Sprintf("--vendorurl='%s'", etab.vendorURL.Text))
			}
			if etab.version.Text != "" {
				vsys = append(vsys, fmt.Sprintf("--version='%s'", etab.version.Text))
			}
			if etab.description.Text != "" {
				vsys = append(vsys, fmt.Sprintf("--description='%s'", etab.description.Text))
			}
			if etab.license.Text != "" {
				vsys = append(vsys, fmt.Sprintf("--eula='%s'", etab.license.Text))
			}
			index++
		}
	}
	return vsys
}

func (e *ExportHelper) doOvaExport() {
	vsys := e.buildVsys()
	machines := make([]string, 0, 5)
	for _, item := range e.vms {
		if item.isSelected {
			machines = append(machines, item.UUID)
		}
	}
	go func() {
		uuid := uuid.NewString()

		name := fmt.Sprintf(lang.X("export.task.name", "Export OVA %s"), util.GetFilename(e.file.Text))
		Gui.TasksInfos.AddTask(uuid, name, "")
		OpenTaskDetails()
		err := e.vmServer.ExportOva(&e.vmServer.Client, machines,
			e.formatMapIndexToType[e.format.SelectedIndex()], e.mnifest.Checked, e.iso.Checked,
			e.macMapIndexToType[e.mac.SelectedIndex()], vsys, e.file.Text, util.WriterFunc(func(p []byte) (int, error) {
				Gui.TasksInfos.UpdateTaskStatus(uuid, string(p), true)
				return len(p), nil
			}))
		if err != nil {
			t := fmt.Sprintf(lang.X("export.done.error", "OVA export '%s' from server '%s' failed"), e.file.Text, e.vmServer.Name)
			SetStatusText(t, MsgError)
			Gui.TasksInfos.AbortTask(uuid, t, false)
		} else {
			t := fmt.Sprintf(lang.X("export.done.ok", "OVA export '%s' from server '%s' was created"), e.file.Text, e.vmServer.Name)
			Gui.TasksInfos.FinishTask(uuid, t, false)
			SendNotification(lang.X("export.notification.title", "OVA exported"), t)
		}
	}()
}

func (e *ExportHelper) doExport2() {
	e.tabs = make([]*ExportTabItem, 0, 5)
	anzahl := 0
	tab := container.NewAppTabs()
	for _, item := range e.vms {
		if item.isSelected {
			etab := e.createExport2Tab(item.VMachine)
			e.tabs = append(e.tabs, etab)
			tab.Append(etab.tab)
			anzahl++
		}
	}
	dia := dialog.NewCustomConfirm(lang.X("export.infos.title", "Export VM infos"),
		lang.X("export.ok", "Ok"),
		lang.X("export.cancel", "Cancel"),
		tab, func(ok bool) {
			if ok {
				e.doOvaExport()
			}
		}, Gui.MainWindow)
	si := Gui.MainWindow.Canvas().Size()
	var windowScale float32 = 0.65
	dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
	dia.Show()
}

func (e *ExportHelper) Export() {
	vms, err := vm.GetVMs(&e.vmServer.Client)
	if err != nil {
		return
	}
	for _, item := range vms {
		e.vms = append(e.vms, &VmExport{
			VMachine:     item,
			isExportable: true,
		})
	}

	go e.updateStatus()

	var dia *dialog.CustomDialog
	var ok *widget.Button

	e.list = widget.NewList(e.listNumberOfItems, e.listCreateObject, e.listUpdateItem)
	e.list.OnSelected = e.listOnSelected
	e.format = widget.NewSelect([]string{
		lang.X("export.format.legacy_0_9", "Legacy 0.9"),
		lang.X("export.format.0_9", "0.9"),
		lang.X("export.format.1_", "1.0"),
		lang.X("export.format.2_0", "2.0"),
	}, nil)

	e.mac = widget.NewSelect([]string{
		lang.X("export.mac.nomacs", "No MACs"),
		lang.X("export.format.nomacsbutnat_", "Include only NAT MACs"),
		lang.X("export.format.allmacs", "Include all MACs"),
	}, nil)

	e.mnifest = widget.NewCheck(lang.X("export.manifest", "Export manifest"), nil)
	e.iso = widget.NewCheck(lang.X("export.iso", "Export ISO files"), nil)
	e.file = widget.NewEntry()
	e.file.SetPlaceHolder(lang.X("export.file.placeholder", "Filepath for ova file"))
	e.file.OnChanged = func(s string) {
		if len(s) > 5 {
			ok.Enable()
		} else {
			ok.Disable()
		}
	}
	e.browse = widget.NewButton(lang.X("export.file.browse", "Browse"), e.doBrowse)
	grid1 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("export.format", "Format")), e.format,
		widget.NewLabel(lang.X("export.mac", "MAC")), e.mac,
		widget.NewLabel(lang.X("export.file", "File")), container.NewBorder(nil, nil, nil, e.browse, e.file),
	)
	grid2 := container.New(layout.NewFormLayout(),
		e.mnifest, e.iso,
	)
	gridWrap := container.NewVBox(grid1, grid2)

	ok = widget.NewButtonWithIcon(lang.X("export.ok", "Ok"), theme.ConfirmIcon(), func() {
		dia.Hide()
		if e.file.Text == "" {
			dialog.ShowError(errors.New(lang.X("export.error.nofile", "No export file given")), Gui.MainWindow)
		} else {
			s := filebrowser.SftpHelper{}
			s.DeleteFile(e.vmServer.Client.Client, e.file.Text)
			e.doExport2()
		}
	})
	ok.Importance = widget.HighImportance
	ok.Disable()
	cancel := widget.NewButtonWithIcon(lang.X("export.cancel", "Cancel"), theme.CancelIcon(), func() {
		dia.Hide()
	})
	buttons := container.New(layout.NewGridLayout(6), layout.NewSpacer(), layout.NewSpacer(), cancel, ok, layout.NewSpacer(), layout.NewSpacer())

	c := container.NewBorder(nil, container.NewVBox(util.NewVFiller(0.5), gridWrap, util.NewVFiller(0.5), buttons), nil, nil, e.list)

	dia = dialog.NewCustomWithoutButtons(lang.X("export.title", "Export ova"), c, Gui.MainWindow)
	si := Gui.MainWindow.Canvas().Size()
	var windowScale float32 = 0.65
	dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*0.75))
	dia.Show()
	e.format.SetSelectedIndex(3)
	e.mac.SetSelectedIndex(0)
}

func (e *ExportHelper) doBrowse() {
	sftp := filebrowser.NewSftpBrowser(e.vmServer.Client.Client, e.vmServer.OvaPath, nil,
		lang.X("export.browse.title", "Select file for export"), filebrowser.SftpFileBrowserMode_savefile)
	sftp.Show(Gui.MainWindow, 0.75, func(file string, fi os.FileInfo, dir string) {
		e.vmServer.OvaPath = dir
		SaveServers()
		e.file.SetText(file)
	})
}

func (e *ExportHelper) listOnSelected(id widget.ListItemID) {
	item := e.vms[id]
	if item.isExportable {
		item.isSelected = !item.isSelected
		e.list.RefreshItem(id)
	}
}

func (e *ExportHelper) listNumberOfItems() int {
	return len(e.vms)
}

func (e *ExportHelper) listCreateObject() fyne.CanvasObject {
	text := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	icon := canvas.NewImageFromResource(theme.CheckButtonIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(16, 16))
	icon.Refresh()
	text.Refresh()
	return container.NewHBox(icon, util.NewFiller(16, 0), text)
}

func (e *ExportHelper) listUpdateItem(id widget.ListItemID, o fyne.CanvasObject) {
	cont, ok := o.(*fyne.Container)
	if !ok {
		return
	}
	text, ok := cont.Objects[2].(*canvas.Text)
	if !ok {
		return
	}
	icon, ok := cont.Objects[0].(*canvas.Image)
	if !ok {
		return
	}
	item := e.vms[id]
	text.Text = item.Name
	if item.isExportable {
		text.Color = theme.Color(theme.ColorNameForeground)
		if item.isSelected {
			icon.Resource = theme.CheckButtonCheckedIcon()
		} else {
			icon.Resource = theme.CheckButtonIcon()
		}
	} else {
		text.Color = theme.Color(theme.ColorNameDisabled)
		icon.Resource = theme.ContentClearIcon()
	}
	text.Refresh()
	icon.Refresh()
}

func doImport() {
	s, _ := getActiveServerAndVm()

	if s == nil {
		return
	}

	i := NewImportHelper(s)
	i.Import()
}

func doExport() {
	s, _ := getActiveServerAndVm()

	if s == nil {
		return
	}

	e := NewExportHelper(s)
	e.Export()
}

type ImportHelper struct {
	vdi               *widget.Check
	mac               *widget.Select
	vmServer          *vm.VmServer
	macMapIndexToType map[int]vm.MacImportType
}

func NewImportHelper(s *vm.VmServer) *ImportHelper {
	return &ImportHelper{
		vmServer:          s,
		macMapIndexToType: map[int]vm.MacImportType{0: vm.MacImport_all, 1: vm.MacImport_natmacs},
	}
}

func (i *ImportHelper) Import() {
	r := regexp.MustCompile(`(?i)\.(ova|ovf)$`)
	sftp := filebrowser.NewSftpBrowser(i.vmServer.Client.Client, i.vmServer.OvaPath, r,
		lang.X("import.browse.title", "Select file for import"), filebrowser.SftpFileBrowserMode_openfile)
	sftp.Show(Gui.MainWindow, 0.75, func(file string, fi os.FileInfo, dir string) {
		i.vmServer.OvaPath = dir
		SaveServers()
		i.doImport2(file)
	})
}

func (i *ImportHelper) doImport2(file string) {
	i.mac = widget.NewSelect([]string{
		lang.X("import.mac.all", "Import all MAC"),
		lang.X("import.mac.onlynat", "Import MAC only for NAT networks"),
	}, nil)
	i.vdi = widget.NewCheck(lang.X("import.vdi", "Import hard drives as VDI"), nil)

	grid1 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("import.mac", "MAC addresses")), i.mac,
	)
	grid2 := container.New(layout.NewFormLayout(),
		i.vdi, util.NewFiller(0, 0),
	)

	label := widget.NewLabel(file)

	c := container.NewVBox(label, grid1, grid2)

	dia := dialog.NewCustomConfirm(lang.X("import.title", "Import ova"),
		lang.X("import.ok", "Ok"),
		lang.X("import.cancel", "Cancel"), c,
		func(ok bool) {
			if ok {
				i.doImportOva(file)
			}
		}, Gui.MainWindow)
	si := Gui.MainWindow.Canvas().Size()
	var windowScale float32 = 0.4
	dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
	dia.Show()
	i.mac.SetSelectedIndex(0)
	i.vdi.SetChecked(true)
	Gui.MainWindow.Canvas().Focus(i.mac)
}

func (i *ImportHelper) doImportOva(file string) {
	// Dry run
	count, err := i.vmServer.ImportOvaDryRun(&i.vmServer.Client, file)
	if err != nil {
		return
	}
	vsys := make([]string, 0, count)
	for index := range count {
		vsys = append(vsys, fmt.Sprintf("--vsys=%d", index))
		vsys = append(vsys, "--eula=accept")
	}

	go func() {
		uuid := uuid.NewString()

		name := fmt.Sprintf(lang.X("import.task.name", "Import OVA %s"), util.GetFilename(file))
		Gui.TasksInfos.AddTask(uuid, name, "")
		OpenTaskDetails()
		s := ""
		err := i.vmServer.ImportOva(&i.vmServer.Client, i.macMapIndexToType[i.mac.SelectedIndex()], i.vdi.Checked,
			file, vsys, util.WriterFunc(func(p []byte) (int, error) {
				s += string(p)
				i := strings.LastIndex(s, "\n")
				strStatus := s
				if i >= 0 {
					strStatus = s[i+1:]
					s = s[i+1:]
				}
				if strStatus != "" {
					Gui.TasksInfos.UpdateTaskStatus(uuid, strStatus, false)
				}
				return len(p), nil
			}))
		if err != nil {
			t := fmt.Sprintf(lang.X("import.done.error", "Import of OVA '%s' in server '%s'failed"), file, i.vmServer.Name)
			SetStatusText(t, MsgError)
			Gui.TasksInfos.AbortTask(uuid, t, false)
		} else {
			t := fmt.Sprintf(lang.X("import.done.ok", "OVA '%s' was imported on server '%s'"), file, i.vmServer.Name)
			Gui.TasksInfos.FinishTask(uuid, t, false)
			SendNotification(lang.X("import.notification.title", "OVA imported"), t)
		}
	}()
}
