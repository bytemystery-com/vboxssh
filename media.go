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
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
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
	"github.com/bytemystery-com/colorlabel"
	"github.com/google/uuid"
)

type CreateNewMedia struct {
	size      *widget.Slider
	sizeEntry *widget.Entry
	fixedSize *widget.Check
	format    *widget.Select
}

type MediaHelper struct {
	hdds   []*vm.HddInfo
	medias []vm.MediaInfo

	uuidMapToHdd  map[string]*vm.HddInfo
	snapshotRegEx *regexp.Regexp
	selectedHd    *vm.HddInfo
	vmServer      *vm.VmServer
	vmMachine     *vm.VMachine

	sel            *widget.Select
	tree           *widget.Tree
	windowScale    float32
	windowScaleNew float32
	mainWindow     fyne.Window

	createNewMedia          CreateNewMedia
	hddFormatMapIndexToType map[int]vm.MediaFormatType
}

func NewMediaHelper(mainWindow fyne.Window, windowScale, windowScaleNew float32) *MediaHelper {
	return &MediaHelper{
		snapshotRegEx:           regexp.MustCompile(`.*/(.+)/.*/({[0-9a-fA-F-]+}\..*)$`),
		windowScale:             windowScale,
		windowScaleNew:          windowScaleNew,
		mainWindow:              mainWindow,
		hddFormatMapIndexToType: map[int]vm.MediaFormatType{0: vm.MediaFormat_vdi, 1: vm.MediaFormat_vmdk, 2: vm.MediaFormat_vhd},
	}
}

func (m *MediaHelper) buildText(h *vm.HddInfo) string {
	if h.Parent == "base" {
		return filepath.Base(h.Location)
	} else {
		items := m.snapshotRegEx.FindStringSubmatch(h.Location)
		if len(items) == m.snapshotRegEx.NumSubexp()+1 {
			return items[1] + "/…/" + items[2]
		}
	}
	return h.Location
}

func (m *MediaHelper) ShowDvdHddOptionDialog(fOk func(dvd bool)) {
	var radio *widget.RadioGroup
	rSelected := 0
	radio = widget.NewRadioGroup([]string{lang.X("media.dialog.mediatype.dvd", "DVD"), lang.X("media.dialog.mediatype.hdd", "HDD")}, func(value string) {
		for index, item := range radio.Options {
			if value == item {
				rSelected = index
				break
			}
		}
	})
	radio.Horizontal = true
	radio.SetSelected(radio.Options[0])
	var c *fyne.Container
	c = container.NewVBox(radio, util.NewVFiller(1.0))

	dia := dialog.NewCustomConfirm(lang.X("details.vm_storage.choosemediatype.title", "Media type"),
		lang.X("details.vm_storage.choosemediatype.ok", "Ok"),
		lang.X("details.vm_storage.choosemediatype.cancel", "Cancel"),
		c, func(ok bool) {
			if ok {
				fOk(rSelected == 0)
			}
		}, m.mainWindow)
	dia.Show()
	// Gui.MainWindow.Canvas().Focus(radio)
}

func (m *MediaHelper) selectImage(fOk func(vm.MediaInfo), fAddNewMedia func(), floppy bool) {
	m.sel = widget.NewSelect(nil, nil)
	m.setMediaList(floppy, nil)
	add := widget.NewButtonWithIcon(lang.X("details.vm_storage.addmedia.add", "Add new"), theme.ContentAddIcon(), fAddNewMedia)
	c := container.NewVBox(container.NewBorder(nil, nil, nil, add, m.sel), util.NewVFiller(1.0))
	dia := dialog.NewCustomConfirm(lang.X("details.vm_storage.addmedia.title", "Add media"),
		lang.X("details.vm_storage.addctrl.add", "Add"),
		lang.X("details.vm_storage.addctrl.cancel", "Cancel"),
		c, func(ok bool) {
			if ok {
				index := m.sel.SelectedIndex()
				if index >= 0 {
					fOk(m.medias[index])
				}
			}
		}, m.mainWindow)
	dia.Show()
	m.mainWindow.Canvas().Focus(m.sel)
	si := m.mainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width*m.windowScale, dia.MinSize().Height*1.2))
}

func (m *MediaHelper) addNewFddMedia() {
	r := regexp.MustCompile(`(?i)\.(img)$`)
	sftp := filebrowser.NewSftpBrowser(m.vmServer.Client.Client, m.vmServer.FloppyImagesPath, r,
		lang.X("details.vm_storage.addmedia.floppy.title", "Select floppy image file"), filebrowser.SftpFileBrowserMode_openfile)
	sftp.Show(m.mainWindow, m.windowScaleNew, func(file string, fi os.FileInfo, dir string) {
		me := vm.MediaInfo{
			Location: file,
		}
		m.vmServer.FloppyImagesPath = dir
		SaveServers()
		m.setMediaList(true, &me)
	})
}

func (m *MediaHelper) addNewDvdMedia() {
	r := regexp.MustCompile(`(?i)\.(iso)$`)
	sftp := filebrowser.NewSftpBrowser(m.vmServer.Client.Client, m.vmServer.DvdImagesPath, r,
		lang.X("details.vm_storage.addmedia.dvd.title", "Select CD/DVD image file"), filebrowser.SftpFileBrowserMode_openfile)
	sftp.Show(m.mainWindow, m.windowScaleNew, func(file string, fi os.FileInfo, dir string) {
		me := vm.MediaInfo{
			Location: file,
		}
		m.vmServer.DvdImagesPath = dir
		SaveServers()
		m.setMediaList(false, &me)
	})
}

func (m *MediaHelper) addNewHddMedia() {
	r := regexp.MustCompile(`(?i)\.(vdi|vmdk|hdd|vhd)$`)
	sftp := filebrowser.NewSftpBrowser(m.vmServer.Client.Client, m.vmServer.HddImagesPath, r,
		lang.X("details.vm_storage.addmedia.hdd.title", "Select HDD image file"), filebrowser.SftpFileBrowserMode_openfile)
	sftp.Show(m.mainWindow, m.windowScaleNew, func(file string, fi os.FileInfo, dir string) {
		m.vmServer.HddImagesPath = dir
		SaveServers()
		hdd := vm.HddInfo{
			Location: file,
			Parent:   "base",
			UUID:     "X" + uuid.NewString(),
		}
		m.hdds = append(m.hdds, &hdd)
		m.uuidMapToHdd[hdd.UUID] = &hdd

		slices.SortFunc(m.hdds, func(a, b *vm.HddInfo) int {
			A := path.Base(a.Location)
			B := path.Base(b.Location)
			an := strings.ToLower(A)
			bn := strings.ToLower(B)

			if an == bn {
				if A < B {
					return -1
				}
				if A > B {
					return 1
				}
				return 0
			}
			if an < bn {
				return -1
			}
			return 1
		})
		m.tree.Refresh()
		m.tree.Select(hdd.UUID)
	})
}

func (m *MediaHelper) ShowNewHddPropertyDialog(title string, file string, fOk func(vm.MediaFormatType, int64, bool)) {
	m.createNewMedia.format = widget.NewSelect([]string{
		lang.X("details.vm_storage.addmedia.create.format.vdi", "VDI"),
		lang.X("details.vm_storage.addmedia.create.format.vmdk", "VMDK"),
		lang.X("details.vm_storage.addmedia.create.format.vhd", "VHD"),
	}, func(selected string) {
		switch m.createNewMedia.format.SelectedIndex() {
		case 0:
			m.createNewMedia.size.Max = 24000000
		default:
			m.createNewMedia.size.Max = 2000000
		}
		m.createNewMedia.size.Refresh()
	})

	m.createNewMedia.sizeEntry = widget.NewEntry()
	m.createNewMedia.size = widget.NewSlider(4, 2000000)
	m.createNewMedia.size.Step = 1
	m.createNewMedia.size.OnChanged = func(val float64) {
		m.createNewMedia.sizeEntry.SetText(fmt.Sprintf("%.0f", val))
	}
	m.createNewMedia.sizeEntry.OnChanged = util.GetNumberFilter(m.createNewMedia.sizeEntry, func(s string) {
		val, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		m.createNewMedia.size.SetValue(float64(val))
	})

	m.createNewMedia.fixedSize = widget.NewCheck(lang.X("details.vm_storage.addmedia.create.fixedsize", "Fixed size"), nil)

	unitSize := util.GetDefaultTextSize("XXXXXX")
	entrySize := util.GetDefaultTextSize("XXXXXXXXXXX")
	entrySize.Height = m.createNewMedia.sizeEntry.MinSize().Height
	formWidth := util.GetFormWidth()

	fileLabel := colorlabel.NewColorLabel(file, nil, nil, 1.0)
	fileLabel.SetTruncateMode(colorlabel.Begin)

	grid1 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_storage.addmedia.create.file", "File")), fileLabel,
		widget.NewLabel(lang.X("details.vm_storage.addmedia.create.size", "Size")),
		container.NewBorder(nil, nil, nil, container.NewHBox(container.NewGridWrap(entrySize, m.createNewMedia.sizeEntry),
			container.NewGridWrap(unitSize, widget.NewLabel(lang.X("details.vm_storage.addmedia.create.mb", "MB")))),
			m.createNewMedia.size),
	)

	grid2 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_storage.addmedia.create.format", "Format")), m.createNewMedia.format,
		m.createNewMedia.fixedSize, util.NewFiller(0, 0),
	)

	gridWrap1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap2 := container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)
	gridWrap := container.NewVBox(gridWrap1, gridWrap2)

	dia := dialog.NewCustomConfirm(title,
		lang.X("details.vm_storage.addmedia.create.add", "Add"),
		lang.X("details.vm_storage.addmedia.create.cancel", "Cancel"),
		gridWrap, func(ok bool) {
			if ok {
				size, err := strconv.ParseInt(m.createNewMedia.sizeEntry.Text, 10, 64)
				if err != nil {
					return
				}
				format, ok := m.hddFormatMapIndexToType[m.createNewMedia.format.SelectedIndex()]
				if !ok {
					return
				}
				fOk(format, size, m.createNewMedia.fixedSize.Checked)

			}
		}, m.mainWindow)
	dia.Show()
	m.createNewMedia.format.SetSelectedIndex(0)
	m.createNewMedia.sizeEntry.SetText(fmt.Sprintf("%.0f", m.createNewMedia.size.Min))
}

func (m *MediaHelper) createNewHddMedia() {
	s, _ := getActiveServerAndVm()
	if s == nil {
		return
	}

	r := regexp.MustCompile(`(?i)\.(vdi|vmdk|hdd|vhd)$`)
	startFolder := m.vmServer.HddImagesPath
	if m.vmMachine != nil {
		s, ok := m.vmMachine.Properties["CfgFile"]
		if ok {
			startFolder = path.Dir(s)
		}
	}

	sftp := filebrowser.NewSftpBrowser(m.vmServer.Client.Client, startFolder, r,
		lang.X("details.vm_storage.addmedia.hdd.create.title", "Create new HDD image file"), filebrowser.SftpFileBrowserMode_savefile)
	sftp.Show(m.mainWindow, m.windowScaleNew, func(file string, fi os.FileInfo, dir string) {
		if fi != nil {
			err := s.DeleteMedia(&m.vmServer.Client, vm.Media_disk, file)
			if err != nil {
				s := filebrowser.SftpHelper{}
				s.DeleteFile(m.vmServer.Client.Client, file)
			}
		}
		m.ShowNewHddPropertyDialog(lang.X("details.vm_storage.addmedia.create.title", "New HDD properties"), file,
			func(format vm.MediaFormatType, size int64, fixedSize bool) {
				uuid := uuid.NewString()

				name := lang.X("details.vm_storge.createmedium.task.name", "Create media")
				Gui.TasksInfos.AddTask(uuid, name, "")
				OpenTaskDetails()
				ResetStatus()

				go func() {
					err := s.CreateMedia(&s.Client, vm.Media_disk, size, &format, &fixedSize, file, util.WriterFunc(func(p []byte) (int, error) {
						Gui.TasksInfos.UpdateTaskStatus(uuid, string(p), true)
						return len(p), nil
					}))
					if err != nil {
						t := fmt.Sprintf(lang.X("details.vm_storge.createmedium.done.error", "Create medium on server '%s' failed"), s.Name)
						SetStatusText(t, MsgError)
						Gui.TasksInfos.AbortTask(uuid, t, false)
					} else {
						t := fmt.Sprintf(lang.X("details.vm_storge.createmedium.done.ok", "Medium on server '%s' was created"), s.Name)
						Gui.TasksInfos.FinishTask(uuid, t, false)
						SendNotification(lang.X("details.vm_storge.createmedium.notification.title", "Medium created exported"), t)
						fyne.Do(func() {
							m.hdds, m.uuidMapToHdd, err = s.GetHddMedias()
							m.tree.Refresh()
						})
					}
				}()
			})
	})
}

func (m *MediaHelper) setMediaList(floppy bool, newMedium *vm.MediaInfo) {
	if floppy {
		list, err := m.vmServer.GetFloppyMedias()
		if err != nil {
			return
		}
		m.medias = make([]vm.MediaInfo, 0, len(list))
		for _, item := range list {
			m.medias = append(m.medias, item.MediaInfo)
		}
	} else {
		list, err := m.vmServer.GetDvdMedias()
		if err != nil {
			return
		}
		m.medias = make([]vm.MediaInfo, 0, len(list))
		for _, item := range list {
			m.medias = append(m.medias, item.MediaInfo)
		}
	}
	if newMedium != nil {
		m.medias = append(m.medias, *newMedium)
	}

	slices.SortFunc(m.medias, func(a, b vm.MediaInfo) int {
		A := path.Base(a.Location)
		B := path.Base(b.Location)
		an := strings.ToLower(A)
		bn := strings.ToLower(B)

		if an == bn {
			if A < B {
				return -1
			}
			if A > B {
				return 1
			}
			return 0
		}
		if an < bn {
			return -1
		}
		return 1
	})

	l := make([]string, 0, len(m.medias))
	for _, item := range m.medias {
		l = append(l, path.Base(item.Location))
	}
	m.sel.SetOptions(l)

	if newMedium != nil {
		m.sel.SetSelected(path.Base(newMedium.Location))
	} else {
		m.sel.ClearSelected()
	}
	m.mainWindow.Canvas().Focus(m.sel)
}

func (m *MediaHelper) SelectFloppyOrDvdImage(s *vm.VmServer, floppy bool, fOk func(vm.MediaInfo)) {
	m.vmServer = s
	if floppy {
		m.selectImage(fOk, m.addNewFddMedia, true)
	} else {
		m.selectImage(fOk, m.addNewDvdMedia, false)
	}
}

func (m *MediaHelper) SelectHddImage(s *vm.VmServer, v *vm.VMachine, fOk func(*vm.HddInfo)) {
	m.vmServer = s
	m.vmMachine = v
	var err error
	m.hdds, m.uuidMapToHdd, err = s.GetHddMedias()
	if err != nil {
		return
	}

	m.tree = widget.NewTree(m.treeGetChilds, m.treeIsBranche, m.treeCreateCanvasObject, m.treeUpdateItem)
	m.tree.OnSelected = m.treeOnSelected
	add := widget.NewButtonWithIcon(lang.X("details.vm_storage.addmedia.add", "Add new"), theme.ContentAddIcon(), m.addNewHddMedia)
	create := widget.NewButtonWithIcon(lang.X("details.vm_storage.addmedia.create", "Create new"), theme.DocumentCreateIcon(), m.createNewHddMedia)
	c := container.NewBorder(container.NewHBox(add, create), util.NewVFiller(1.0), nil, nil, m.tree)
	dia := dialog.NewCustomConfirm(lang.X("details.vm_storage.addhdd.title", "Add media"),
		lang.X("details.vm_storage.addhdd.add", "Add"),
		lang.X("details.vm_storage.addhdd.cancel", "Cancel"),
		c, func(ok bool) {
			if ok {
				if m.selectedHd != nil {
					fOk(m.selectedHd)
				}
			}
		}, m.mainWindow)

	si := m.mainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width*.8, si.Height*.8))
	dia.Show()
}

// return childs
func (m *MediaHelper) treeGetChilds(id widget.TreeNodeID) []widget.TreeNodeID {
	if id == "" {
		list := make([]widget.TreeNodeID, 0, len(m.hdds))
		for _, item := range m.hdds {
			list = append(list, item.UUID)
		}
		return list
	} else {
		h, ok := m.uuidMapToHdd[id]
		if !ok {
			return nil
		}
		list := make([]widget.TreeNodeID, 0, len(h.Childs))
		for _, item := range h.Childs {
			list = append(list, item.UUID)
		}
		return list
	}
}

// return is Branch
func (m *MediaHelper) treeIsBranche(id widget.TreeNodeID) bool {
	if id == "" {
		return true
	}
	h, ok := m.uuidMapToHdd[id]
	if !ok {
		return false
	}
	if len(h.Childs) == 0 {
		return false
	}
	return true
}

// create canvasObject
func (m *MediaHelper) treeCreateCanvasObject(branche bool) fyne.CanvasObject {
	var tColor color.Color
	var tScale float32
	var tStyle fyne.TextStyle
	if branche {
		tColor = theme.Color(theme.ColorNamePrimary)
		tScale = 1.0
		tStyle = fyne.TextStyle{Bold: true}
	} else {
		tColor = theme.Color(theme.ColorNameForeground)
		tScale = 1.0
		tStyle = fyne.TextStyle{}
	}
	text := canvas.NewText("", tColor)
	text.TextSize = theme.TextSize() * tScale
	text.TextStyle = tStyle
	text.Refresh()
	icon := canvas.NewImageFromResource(theme.QuestionIcon())
	icon.SetMinSize(fyne.NewSize(16, 16)) // gewünschte Größe
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	return container.NewHBox(text, layout.NewSpacer(), icon, util.NewFiller(24, 0))
}

// update
func (m *MediaHelper) treeUpdateItem(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
	cont, ok := o.(*fyne.Container)
	if !ok {
		return
	}
	text, ok := cont.Objects[0].(*canvas.Text)
	if !ok {
		return
	}

	icon, ok := cont.Objects[2].(*canvas.Image)
	if !ok {
		return
	}

	var tStyle fyne.TextStyle
	var tColor color.Color
	var tScale float32
	h, ok := m.uuidMapToHdd[id]
	if !ok {
		return
	}
	if h.Parent == "base" {
		tColor = theme.Color(theme.ColorNamePrimary)
		tScale = 1.0
		tStyle = fyne.TextStyle{Bold: true}
		icon.Resource = Gui.IconHdd
	} else {
		tColor = theme.Color(theme.ColorNameForeground)
		tScale = 1.0
		tStyle = fyne.TextStyle{}
		icon.Resource = Gui.IconSnapshot
	}

	text.TextSize = theme.TextSize() * tScale
	text.TextStyle = tStyle
	text.Color = tColor
	text.Text = m.buildText(h)
	text.Refresh()

	icon.SetMinSize(fyne.NewSize(16, 16)) // gewünschte Größe
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()
}

func (m *MediaHelper) treeOnSelected(id widget.TreeNodeID) {
	h, ok := m.uuidMapToHdd[id]
	if ok {
		m.selectedHd = h
	} else {
		m.selectedHd = nil
	}
}
