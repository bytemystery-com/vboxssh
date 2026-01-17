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
	"strconv"

	"bytemystery-com/vboxssh/util"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type oldRdpType struct {
	enabled  bool
	multiple bool
	reusecon bool
	ports    string
	security int
	auth     int
}

type RdpTab struct {
	oldValues oldRdpType

	enabled  *widget.Check
	multiple *widget.Check
	reuseCon *widget.Check
	ports    *widget.Entry
	usedPort *widget.Label
	security *widget.Select
	auth     *widget.Select

	apply   *widget.Button
	tabItem *container.TabItem

	securityMapStringToIndex map[string]int
	securityMapIndexToType   map[int]vm.RdpSecurityType
	authMapStringToIndex     map[string]int
	authMapIndexToType       map[int]vm.RdpAuthType
}

var _ DetailsInterface = (*RdpTab)(nil)

func NewRdpTab() *RdpTab {
	rdpTab := RdpTab{
		securityMapStringToIndex: map[string]int{"tls": 0, "rdp": 1, "negotiate": 2},
		securityMapIndexToType:   map[int]vm.RdpSecurityType{0: vm.RdpSecurity_tls, 1: vm.RdpSecurity_rdp, 2: vm.RdpSecurity_negotiate},
		authMapStringToIndex:     map[string]int{"null": 0, "extrenal": 1, "guest": 2},
		authMapIndexToType:       map[int]vm.RdpAuthType{0: vm.RdpAuth_null, 1: vm.RdpAuth_external, 2: vm.RdpAuth_guest},
	}

	rdpTab.apply = widget.NewButton(lang.X("details.vm_rdp.apply", "Apply"), func() {
		rdpTab.Apply()
	})
	rdpTab.apply.Importance = widget.HighImportance

	formWidth := util.GetFormWidth()

	rdpTab.enabled = widget.NewCheck(lang.X("details.vm_rdp.enabled", "Enabled"), func(checked bool) {
		rdpTab.UpdateByStatus()
	})

	rdpTab.multiple = widget.NewCheck(lang.X("details.vm_rdp.multiple", "Multiple connections"), func(checked bool) {
		rdpTab.UpdateByStatus()
	})
	rdpTab.reuseCon = widget.NewCheck(lang.X("details.vm_rdp.reusecon", "Reuse connection"), func(checked bool) {})
	rdpTab.usedPort = widget.NewLabel("-----")
	rdpTab.ports = widget.NewEntry()
	rdpTab.ports.SetPlaceHolder(lang.X("details.vm_rdp.ports_placeholder", "5000-5010,5012"))
	rdpTab.security = widget.NewSelect([]string{
		lang.X("details.vm_rdp.security.tls", "TLS"),
		lang.X("details.vm_rdp.security.rdp", "RDP"),
		lang.X("details.vm_rdp.security.negotiate", "NEGOTIATE"),
	}, nil)

	rdpTab.auth = widget.NewSelect([]string{
		lang.X("details.vm_rdp.auth.null", "Null"),
		lang.X("details.vm_rdp.auth.external", "External"),
		lang.X("details.vm_rdp.auth.guest", "Guest"),
	}, nil)

	grid1 := container.New(layout.NewFormLayout(),
		rdpTab.enabled, util.NewFiller(0, 0),
	)
	grid2 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_rdp.ports", "Port")), rdpTab.ports,
		widget.NewLabel(lang.X("details.vm_rdp.used_port", "Used port")), rdpTab.usedPort,
		widget.NewLabel(lang.X("details.vm_rdp.security", "Security")), rdpTab.security,
		widget.NewLabel(lang.X("details.vm_rdp.auth", "Authentication")), rdpTab.auth,
		rdpTab.multiple, rdpTab.reuseCon,
	)
	gridWrap1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap2 := container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)

	gridWrap := container.NewVBox(util.NewVFiller(0.5), gridWrap1, gridWrap2)

	c := container.NewVBox(container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), rdpTab.apply, util.NewFiller(32, 0)))
	rdpTab.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.rdp", "RDP"), c)
	return &rdpTab
}

func (rdp *RdpTab) setUsedPort(v *vm.VMachine) {
	// Used port
	str, ok := v.Properties["vrdeport"]
	if ok {
		port, err := strconv.Atoi(str)
		if err == nil {
			if port == -1 {
				rdp.usedPort.SetText(lang.X("details.vm_rdp.used_port.unused", "-----"))
			} else {
				rdp.usedPort.SetText(fmt.Sprintf("%d", port))
			}
		}
	}
}

// calles by selection change
func (rdp *RdpTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()

	if s == nil || v == nil {
		rdp.DisableAll()
		return
	}
	rdp.apply.Enable()

	// Enabled
	if util.CheckFromProperty(rdp.enabled, v, "vrde", "on", &rdp.oldValues.enabled) {
		rdp.enableDisableRdpCtrls(true)
	} else {
		rdp.enableDisableRdpCtrls(false)
	}

	// Multiple
	util.CheckFromProperty(rdp.multiple, v, "vrdemulticon", "on", &rdp.oldValues.multiple)

	// Reuse connections
	util.CheckFromProperty(rdp.reuseCon, v, "vrdereusecon", "on", &rdp.oldValues.reusecon)

	// Security
	util.SelectEntryFromProperty(rdp.security, v, "vrdeproperty[Security/Method]", rdp.securityMapStringToIndex, &rdp.oldValues.security)

	// Authentication
	util.SelectEntryFromProperty(rdp.auth, v, "vrdeauthtype", rdp.authMapStringToIndex, &rdp.oldValues.auth)

	// Ports
	str, ok := v.Properties["vrdeports"]
	if ok {
		rdp.ports.SetText(str)
		rdp.oldValues.ports = str
	}

	rdp.setUsedPort(v)
}

// called from status updates
func (rdp *RdpTab) UpdateByStatus() {
	_, v := getActiveServerAndVm()
	if v != nil {
		state, err := v.GetState()
		if err != nil {
			return
		}
		switch state {
		case vm.RunState_unknown, vm.RunState_meditation:
			rdp.DisableAll()

		case vm.RunState_running, vm.RunState_paused, vm.RunState_saved:
			rdp.enabled.Disable()
			rdp.ports.Disable()
			rdp.reuseCon.Disable()
			rdp.multiple.Disable()
			rdp.security.Disable()
			rdp.auth.Disable()

		case vm.RunState_off, vm.RunState_aborted:
			rdp.enabled.Enable()
			rdp.enableDisableRdpCtrls(rdp.enabled.Checked)

		default:
			SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
		}
		rdp.setUsedPort(v)
	} else {
		rdp.DisableAll()
	}
}

func (rdp *RdpTab) enableDisableRdpCtrls(enable bool) {
	if enable {
		rdp.ports.Enable()
		rdp.multiple.Enable()
		rdp.security.Enable()
		rdp.auth.Enable()
		if rdp.multiple.Checked {
			rdp.reuseCon.Disable()
		} else {
			rdp.reuseCon.Enable()
		}
	} else {
		rdp.reuseCon.Disable()
		rdp.ports.Disable()
		rdp.multiple.Disable()
		rdp.security.Disable()
		rdp.auth.Disable()
	}
}

func (rdp *RdpTab) DisableAll() {
	rdp.enabled.Disable()
	rdp.enableDisableRdpCtrls(false)
	rdp.apply.Disable()
}

func (rdp *RdpTab) Apply() {
	s, v := getActiveServerAndVm()
	if v != nil {
		ResetStatus()
		if !rdp.enabled.Disabled() {
			if rdp.enabled.Checked != rdp.oldValues.enabled {
				go func() {
					err := v.SetEnableRde(&s.Client, rdp.enabled.Checked, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_rdp.enablerdp.error", "Enable RDP for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						rdp.oldValues.enabled = rdp.enabled.Checked
					}
				}()
			}
		}
		if !rdp.ports.Disabled() {
			if rdp.ports.Text != rdp.oldValues.ports {
				go func() {
					err := v.SetRdePorts(s, rdp.ports.Text, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_rdp.setports.error", "Set RDP ports for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						rdp.oldValues.ports = rdp.ports.Text
					}
				}()
			}
		}
		if !rdp.multiple.Disabled() {
			if rdp.multiple.Checked != rdp.oldValues.multiple {
				go func() {
					err := v.SetRdeMultiConnection(s, rdp.multiple.Checked, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_rdp.multiuse.error", "Set RDP multi use for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						rdp.oldValues.multiple = rdp.multiple.Checked
					}
				}()
			}
		}
		if !rdp.reuseCon.Disabled() {
			if rdp.reuseCon.Checked != rdp.oldValues.reusecon {
				go func() {
					err := v.SetRdeReuseConnection(s, rdp.reuseCon.Checked, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_rdp.reuse.error", "Set RDP reuse connection for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						rdp.oldValues.reusecon = rdp.reuseCon.Checked
					}
				}()
			}
		}
		if !rdp.security.Disabled() {
			index := rdp.security.SelectedIndex()
			if index != rdp.oldValues.security {
				if index >= 0 {
					val, ok := rdp.securityMapIndexToType[index]
					if ok {
						go func() {
							err := v.SetRdeSecurityMethode(s, val, VMStatusUpdateCallBack)
							if err != nil {
								SetStatusText(fmt.Sprintf(lang.X("details.vm_rdp.setsecurity.error", "Set security methode for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
							} else {
								rdp.oldValues.security = index
							}
						}()
					}
				}
			}
		}
		if !rdp.auth.Disabled() {
			index := rdp.auth.SelectedIndex()
			if index != rdp.oldValues.auth {
				if index >= 0 {
					val, ok := rdp.authMapIndexToType[index]
					if ok {
						go func() {
							err := v.SetRdeAuthType(s, val, VMStatusUpdateCallBack)
							if err != nil {
								SetStatusText(fmt.Sprintf(lang.X("details.vm_rdp.authtype.error", "Set auth type for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
							} else {
								rdp.oldValues.auth = index
							}
						}()
					}
				}
			}
		}
	}
}
