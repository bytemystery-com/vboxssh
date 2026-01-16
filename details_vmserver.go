package main

import (
	"fmt"
	"image/color"
	"strings"

	"bytemystery-com/vboxssh/util"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bytemystery-com/colorlabel"
)

type ExtPackDataType struct {
	vm.ExtPackInfoType
}

type VmServerInfos struct {
	vmVersion   *colorlabel.ColorLabel
	os          *colorlabel.ColorLabel
	osVersion   *colorlabel.ColorLabel
	cpuCores    *colorlabel.ColorLabel
	ramSize     *colorlabel.ColorLabel
	freeRam     *colorlabel.ColorLabel
	extPackList *widget.List

	tabItem     *container.TabItem
	extPackData []*ExtPackDataType
}

var _ DetailsInterface = (*VmServerInfos)(nil)

func NewVmServerTab() *VmServerInfos {
	srv := VmServerInfos{}

	srv.vmVersion = colorlabel.NewColorLabel("", theme.ColorNamePrimary, nil, 1)
	srv.osVersion = colorlabel.NewColorLabel("", theme.ColorNamePrimary, nil, 1)
	srv.os = colorlabel.NewColorLabel("", theme.ColorNamePrimary, nil, 1)
	srv.cpuCores = colorlabel.NewColorLabel("", theme.ColorNamePrimary, nil, 1)
	srv.ramSize = colorlabel.NewColorLabel("", theme.ColorNamePrimary, nil, 1)
	srv.freeRam = colorlabel.NewColorLabel("", theme.ColorNamePrimary, nil, 1)

	formWidth := util.GetFormWidth() / 2
	// formWidth := util.GetFormWidth()

	dummy := canvas.NewRectangle(color.Transparent)
	dummy.SetMinSize(widget.NewLabel("X").MinSize())

	grid1 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.srvv.version", "VM version:")), srv.vmVersion,
		widget.NewLabel(lang.X("details.srvv.os", "OS:")), srv.os,
		widget.NewLabel(lang.X("details.srvv.cores", "CPU cores:")), srv.cpuCores,
		dummy, dummy,
	)

	srv.extPackList = widget.NewList(srv.extPackListGetLength, srv.extPackListCreate, srv.extPackListUpdateItem)

	grid2 := container.New(layout.NewFormLayout(),
		dummy, dummy,
		widget.NewLabel(lang.X("details.srvv.osversion", "OS version:")), srv.osVersion,
		widget.NewLabel(lang.X("details.srvv.ram", "RAM:")), srv.ramSize,
		widget.NewLabel(lang.X("details.srvv.freeram", "Free RAM:")), srv.freeRam,
	)

	i1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	i2 := container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)

	label := widget.NewLabel(lang.X("details.vm_info.extpacks.label", "Extension packs:"))
	content := container.NewVBox(util.NewVFiller(0.5), container.NewHBox(i1, i2), util.NewVFiller(0.5), label)

	content = container.NewBorder(content, nil, nil, nil, srv.extPackList)

	srv.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.vm", "VM"), content)

	return &srv
}

func (srv *VmServerInfos) extPackListGetLength() int {
	return len(srv.extPackData)
}

func (srv *VmServerInfos) extPackListCreate() fyne.CanvasObject {
	icon := canvas.NewImageFromResource(Gui.IconOk)
	icon.SetMinSize(fyne.NewSize(16, 16))
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	text := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	text.Refresh()

	return container.NewHBox(icon, util.NewFiller(16, 16), text)
}

func (srv *VmServerInfos) extPackListUpdateItem(id widget.ListItemID, o fyne.CanvasObject) {
	c, ok := o.(*fyne.Container)
	if !ok {
		return
	}

	text, ok := c.Objects[2].(*canvas.Text)
	if !ok {
		return
	}

	icon, ok := c.Objects[0].(*canvas.Image)
	if !ok {
		return
	}
	item := srv.extPackData[id]

	text.Text = fmt.Sprintf("%s - %s (%s)", item.Name, item.Version, item.Revision)
	if item.Usable {
		icon.Resource = Gui.IconOk
	} else {
		icon.Resource = Gui.IconError
	}
	text.Color = theme.Color(theme.ColorNamePrimary)
	icon.Refresh()
	text.Refresh()
}

func (srv *VmServerInfos) UpdateBySelect() {
	s, _ := getActiveServerAndVm()
	srv.reset()
	if s == nil {
		return
	}
	srv.vmVersion.SetText(s.Version)

	hostInfos, err := s.GetHostInfos(true)
	if err != nil {
		return
	}

	srv.os.SetText(strings.TrimSpace(hostInfos["Operating system"]))
	srv.osVersion.SetText(strings.TrimSpace(hostInfos["Operating system version"]))

	str, ok := hostInfos["Processor online count"]
	if ok {
		srv.cpuCores.SetText(str)
	}

	str, ok = hostInfos["Memory size"]
	if ok {
		srv.ramSize.SetText(str)
	}

	str, ok = hostInfos["Memory available"]
	if ok {
		srv.freeRam.SetText(str)
	}

	extPackList, err := s.GetExtPackHostInfos()
	if err != nil {
		return
	}
	srv.extPackData = make([]*ExtPackDataType, 0, len(extPackList))
	for _, item := range extPackList {
		i := ExtPackDataType{
			ExtPackInfoType: *item,
		}
		srv.extPackData = append(srv.extPackData, &i)
	}
}

func (srv *VmServerInfos) reset() {
	srv.vmVersion.SetText("")
	srv.os.SetText("")
	srv.osVersion.SetText("")
	srv.cpuCores.SetText("")
	srv.ramSize.SetText("")
	srv.freeRam.SetText("")
}

func (srv *VmServerInfos) DisableAll() {
}

func (srv *VmServerInfos) UpdateByStatus() {
}

func (srv *VmServerInfos) Apply() {
}
