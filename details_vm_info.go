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
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/bytemystery-com/colorlabel"
)

type oldInfoType struct {
	name        string
	osversion   string
	description string
}

type InfoTab struct {
	oldValues   oldInfoType
	version     *widget.Label
	cfgLocation *colorlabel.ColorLabel
	name        *widget.Entry
	os          *widget.Select
	osVersion   *widget.Select
	description *widget.Entry

	apply   *widget.Button
	tabItem *container.TabItem
}

var _ DetailsInterface = (*InfoTab)(nil)

func NewInfoTab() *InfoTab {
	infoTab := InfoTab{}

	// Info
	infoTab.version = widget.NewLabel("")
	infoTab.cfgLocation = colorlabel.NewColorLabel("", "", "", 1.0)
	infoTab.cfgLocation.SetTruncateMode(colorlabel.Begin)
	infoTab.name = widget.NewEntry()
	infoTab.name.SetPlaceHolder(lang.X("details.vm_info.name_placeholder", "Name of VM"))
	infoTab.os = widget.NewSelect([]string{}, func(sel string) {
		s, v := getActiveServerAndVm()
		infoTab.setVersionTypes(s, v)
	})

	infoTab.osVersion = widget.NewSelect([]string{}, nil)
	infoTab.description = widget.NewMultiLineEntry()
	infoTab.description.SetPlaceHolder(lang.X("details.vm_info.description_placeholder", "Description of VM"))
	infoTab.description.SetMinRowsVisible(7)
	infoTab.apply = widget.NewButton(lang.X("details.vm_info", "Apply"), func() {
		infoTab.Apply()
	})
	infoTab.apply.Importance = widget.HighImportance

	formWidth := util.GetFormWidth()

	grid := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_info.version", "VM Version")), infoTab.version,
		widget.NewLabel(lang.X("details.vm_info.cfglocation", "Location")), infoTab.cfgLocation,
		widget.NewLabel(lang.X("details.vm_info.name", "Name")), infoTab.name,
		widget.NewLabel(lang.X("details.vm_info.os", "Operating system")), infoTab.os,
		widget.NewLabel(lang.X("details.vm_info.os_version", "OS Version")), infoTab.osVersion,
		widget.NewLabel(lang.X("details.vm_info.os_description", "Description")), infoTab.description,
	)

	gridWrap := container.NewGridWrap(fyne.NewSize(formWidth, grid.MinSize().Height), grid)

	c := container.NewVBox(util.NewVFiller(0.5), container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), infoTab.apply, util.NewFiller(32, 0)))
	infoTab.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.info", "Info"), c)
	return &infoTab
}

func (info *InfoTab) setOsTypes(s *vm.VmServer, v *vm.VMachine) {
	if s == nil || !s.IsConnected() {
		return
	}
	types, err := s.GetOsTypes(false)
	if err != nil {
		return
	}
	families, err := vm.GetOsFamilies(types)
	if err != nil {
		return
	}
	f := make([]string, 0, len(families))
	for _, item := range families {
		f = append(f, item.Family)
	}
	info.os.SetOptions(f)
	if v == nil {
		return
	}
	ostype := v.Properties["ostype"]
	if ostype == "" {
		return
	}
	info.osVersion.ClearSelected()
	for _, item := range types {
		if item.Name == ostype {
			info.os.SetSelected(item.Family)
			break
		}
	}
}

func (info *InfoTab) setVersionTypes(s *vm.VmServer, v *vm.VMachine) {
	if s == nil || !s.IsConnected() {
		return
	}
	types, err := s.GetOsTypes(false)
	if err != nil {
		return
	}

	versions, err := vm.GetOsVersionTypes(info.os.Selected, "", types)
	if err != nil {
		return
	}

	f := make([]string, 0, len(versions))
	for _, item := range versions {
		f = append(f, item)
	}
	info.osVersion.SetOptions(f)

	if v == nil {
		return
	}
	ostype := v.Properties["ostype"]
	if ostype == "" {
		return
	}

	info.osVersion.ClearSelected()
	for _, item := range types {
		if item.Name == ostype {
			info.oldValues.osversion = item.Name
			info.osVersion.SetSelected(item.Name)
			break
		}
	}
}

// calles by selection change
func (info *InfoTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()
	if s != nil {
		info.version.SetText(s.Version)
	} else {
		info.version.SetText("")
	}

	if s == nil || v == nil {
		info.DisableAll()
		return
	}
	info.apply.Enable()

	info.setOsTypes(s, v)
	info.setVersionTypes(s, v)
	location := v.Properties["CfgFile"]
	info.cfgLocation.SetText(location)

	name := v.Properties["name"]
	info.name.SetText(name)
	info.oldValues.name = name

	description := v.Properties["description"]
	info.description.SetText(description)
	info.oldValues.description = description
}

// called from status updates
func (info *InfoTab) UpdateByStatus() {
	_, v := getActiveServerAndVm()
	if v != nil {
		state, err := v.GetState()
		if err != nil {
			return
		}
		switch state {
		case vm.RunState_unknown:
			info.DisableAll()

		case vm.RunState_running, vm.RunState_paused, vm.RunState_meditation:
			info.name.Disable()
			info.os.Disable()
			info.osVersion.Disable()
			info.description.Enable()
			info.apply.Enable()

		case vm.RunState_saved:
			info.name.Enable()
			info.os.Disable()
			info.osVersion.Disable()
			info.description.Enable()
			info.apply.Enable()

		case vm.RunState_off, vm.RunState_aborted:
			info.name.Enable()
			info.os.Enable()
			info.osVersion.Enable()
			info.description.Enable()
			info.apply.Enable()

		default:
			SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
		}
	} else {
		info.DisableAll()
	}
}

func (info *InfoTab) DisableAll() {
	info.description.Disable()
	info.os.Disable()
	info.name.Disable()
	info.osVersion.Disable()
	info.apply.Disable()
}

func (info *InfoTab) Apply() {
	s, v := getActiveServerAndVm()
	if v != nil {
		ResetStatus()
		if !info.name.Disabled() && info.name.Text != info.oldValues.name {
			go func() {
				err := v.SetName(&s.Client, info.name.Text, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_info.setname.error", "Set name for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					info.oldValues.name = info.name.Text
				}
			}()
		}
		if !info.osVersion.Disabled() && info.osVersion.Selected != info.oldValues.osversion {
			osTypes, err := s.GetOsTypes(false)
			if err == nil {
				for _, item := range osTypes {
					if item.Name == info.osVersion.Selected {
						go func() {
							err := v.SetOsType(&s.Client, item.ID, VMStatusUpdateCallBack)
							if err != nil {
								SetStatusText(fmt.Sprintf(lang.X("details.vm_info.setos.error", "Set OS for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
							} else {
								info.oldValues.osversion = info.osVersion.Selected
							}
						}()
						break
					}
				}
			}
		}
		if !info.description.Disabled() && info.description.Text != info.oldValues.description {
			go func() {
				err := v.SetDescription(&s.Client, info.description.Text, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_info.setdescription.error", "Set description for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					info.oldValues.description = info.description.Text
				}
			}()
		}
	}
}
