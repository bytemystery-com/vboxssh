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
	"regexp"
	"strconv"

	"bytemystery-com/vboxssh/util"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type oldCpuRamType struct {
	cpus              int
	cpuCap            int
	ram               int
	pae               bool
	nestedVt          bool
	nestedPaging      bool
	paraVirtInterface int
	x2apic            bool
}

type CpuRamTab struct {
	oldValues oldCpuRamType

	cpus              *widget.Slider
	cpusEntry         *widget.Entry
	cpuCap            *widget.Slider
	cpuCapEntry       *widget.Entry
	memory            *widget.Slider
	memoryEntry       *widget.Entry
	pae               *widget.Check
	nestedVT          *widget.Check
	x2Apic            *widget.Check
	nestedPaging      *widget.Check
	paraVirtInterface *widget.Select

	apply   *widget.Button
	tabItem *container.TabItem

	paraVirtMapStringToIndex map[string]int
	paraVirtMapIndexToType   map[int]vm.ParaVirtProviderType

	regexRam *regexp.Regexp
}

var _ DetailsInterface = (*CpuRamTab)(nil)

func NewCpuRamTab() *CpuRamTab {
	cpuRamTab := CpuRamTab{
		paraVirtMapStringToIndex: map[string]int{"none": 0, "default": 1, "legacy": 2, "minimal": 3, "hyperv": 4, "kvm": 5},
		paraVirtMapIndexToType: map[int]vm.ParaVirtProviderType{
			0: vm.ParaVirtProvider_none, 1: vm.ParaVirtProvider_default,
			2: vm.ParaVirtProvider_legacy, 3: vm.ParaVirtProvider_minimal, 4: vm.ParaVirtProvider_hyperV, 5: vm.ParaVirtProvider_kvm,
		},
		regexRam: regexp.MustCompile(`^([0-9]+)\s*.*`),
	}

	cpuRamTab.apply = widget.NewButton(lang.X("details.vm_vpuram.apply", "Apply"), func() {
		cpuRamTab.Apply()
	})
	cpuRamTab.apply.Importance = widget.HighImportance

	formWidth := util.GetFormWidth()
	cpuRamTab.cpusEntry = widget.NewEntry()
	cpuRamTab.memoryEntry = widget.NewEntry()
	cpuRamTab.cpuCapEntry = widget.NewEntry()

	cpuRamTab.cpus = widget.NewSlider(1, 2)
	cpuRamTab.cpus.Step = 1
	cpuRamTab.cpus.OnChanged = func(val float64) {
		cpuRamTab.cpusEntry.SetText(fmt.Sprintf("%.0f", val))
	}
	cpuRamTab.cpusEntry.OnChanged = util.GetNumberFilter(cpuRamTab.cpusEntry, func(s string) {
		val, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		cpuRamTab.cpus.SetValue(float64(val))
	})

	cpuRamTab.cpuCap = widget.NewSlider(1, 100)
	cpuRamTab.cpuCap.Step = 1
	cpuRamTab.cpuCap.OnChanged = func(val float64) {
		cpuRamTab.cpuCapEntry.SetText(fmt.Sprintf("%.0f", val))
	}
	cpuRamTab.cpuCapEntry.OnChanged = util.GetNumberFilter(cpuRamTab.cpuCapEntry, func(s string) {
		val, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		cpuRamTab.cpuCap.SetValue(float64(val))
	})

	cpuRamTab.memory = widget.NewSlider(4, 128)
	cpuRamTab.memory.Step = 1
	cpuRamTab.memory.OnChanged = func(val float64) {
		cpuRamTab.memoryEntry.SetText(fmt.Sprintf("%.0f", val))
	}
	cpuRamTab.memoryEntry.OnChanged = util.GetNumberFilter(cpuRamTab.memoryEntry, func(s string) {
		val, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		cpuRamTab.memory.SetValue(float64(val))
	})

	cpuRamTab.cpusEntry.SetText(fmt.Sprintf("%.0f", cpuRamTab.cpus.Min))
	cpuRamTab.cpuCapEntry.SetText(fmt.Sprintf("%.0f", cpuRamTab.cpuCap.Min))
	cpuRamTab.memoryEntry.SetText(fmt.Sprintf("%.0f", cpuRamTab.memory.Min))

	unitSize := util.GetDefaultTextSize("XXXXXX")
	entrySize := util.GetDefaultTextSize("XXXXXXXX")
	entrySize.Height = cpuRamTab.cpusEntry.MinSize().Height

	grid1 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_cpuram.cpus", "CPUs")),
		container.NewBorder(nil, nil, nil, container.NewHBox(container.NewGridWrap(entrySize, cpuRamTab.cpusEntry),
			container.NewGridWrap(unitSize, widget.NewLabel(lang.X("details.vm_cpuram.cpus.cores", "cores")))),
			cpuRamTab.cpus),

		widget.NewLabel(lang.X("details.vm_cpuram.cpucap", "CPU cap")),
		container.NewBorder(nil, nil, nil, container.NewHBox(container.NewGridWrap(entrySize, cpuRamTab.cpuCapEntry),
			container.NewGridWrap(unitSize, widget.NewLabel(lang.X("details.vm_cpuram.cpucap.percent", "%")))),
			cpuRamTab.cpuCap),

		widget.NewLabel(lang.X("details.vm_cpuram.ram", "RAM")),
		container.NewBorder(nil, nil, nil, container.NewHBox(container.NewGridWrap(entrySize, cpuRamTab.memoryEntry),
			container.NewGridWrap(unitSize, widget.NewLabel(lang.X("details.vm_cpuram.ram.mb", "MB")))),
			cpuRamTab.memory),
	)
	cpuRamTab.nestedPaging = widget.NewCheck(lang.X("details.vm_cpuram.nestedpage", "Nested paging"), nil)
	cpuRamTab.nestedVT = widget.NewCheck(lang.X("details.vm_cpuram.nestedvt", "Nested VT"), nil)
	cpuRamTab.pae = widget.NewCheck(lang.X("details.vm_cpuram.pae", "PAE"), nil)
	cpuRamTab.x2Apic = widget.NewCheck(lang.X("details.vm_cpuram.x2apic", "x2APIC"), nil)
	cpuRamTab.paraVirtInterface = widget.NewSelect([]string{
		lang.X("details.vm_cpuram.paravirt.none", "None"),
		lang.X("details.vm_cpuram.paravirt.default", "Default"),
		lang.X("details.vm_cpuram.paravirt.legacy", "Legacy"),
		lang.X("details.vm_cpuram.paravirt.minimal", "Minimal"),
		lang.X("details.vm_cpuram.paravirt.hyperv", "Hyper-V"),
		lang.X("details.vm_cpuram.paravirt.kvm", "KVM"),
	}, nil)

	grid2 := container.New(layout.NewFormLayout(),
		cpuRamTab.nestedPaging, cpuRamTab.nestedVT,
		cpuRamTab.pae, cpuRamTab.x2Apic,

		widget.NewLabel(lang.X("details.vm_cpuram.paravirt", "Paravirtualization interface")), cpuRamTab.paraVirtInterface,
	)

	gridWrap1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap2 := container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)
	gridWrap := container.NewVBox(util.NewVFiller(0.5), gridWrap1, util.NewFiller(0, 10), widget.NewSeparator(), util.NewFiller(0, 10), gridWrap2)

	c := container.NewVBox(container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), cpuRamTab.apply, util.NewFiller(32, 0)))
	cpuRamTab.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.cpuram", "CPU/RAM"), c)
	return &cpuRamTab
}

// calles by selection change
func (cr *CpuRamTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		cr.DisableAll()
		return
	}
	cr.apply.Enable()

	hostInfos, err := s.GetHostInfos(false)
	if err != nil {
		return
	}
	sysprop, err := s.GetSystemProperties(false)
	if err != nil {
		return
	}
	_ = v

	// CPU
	str, ok := hostInfos["Processor online count"]
	if ok {
		max, err := strconv.Atoi(str)
		if err == nil {
			sysprop, err := s.GetSystemProperties(false)
			if err == nil {
				str, ok := sysprop["Maximum guest CPU count"]
				if ok {
					max2, err := strconv.Atoi(str)
					if err == nil {
						if max > max2 {
							max = max2
						}
					}
				}
			}
			cr.cpus.Max = float64(max)
			cr.cpus.Refresh()
		}
	}

	// cr.memoryEntry.Validator = nil

	// RAM
	str, ok = hostInfos["Memory size"]
	if ok {
		items := cr.regexRam.FindStringSubmatch(str)
		if len(items) == 2 {
			max, err := strconv.Atoi(items[1])
			if err == nil {
				sysprop, err := s.GetSystemProperties(false)
				if err == nil {
					str, ok := sysprop["Maximum guest RAM size"]
					if ok {
						items := cr.regexRam.FindStringSubmatch(str)
						if len(items) == 2 {
							max2, err := strconv.Atoi(items[1])
							if err == nil {
								if max > max2 {
									max = max2
								}
							}
						}
					}
				}
				cr.memory.Max = float64(max)
				cr.memory.Refresh()
			}
		}
	}

	str, ok = sysprop["Minimum guest RAM size"]
	if ok {
		items := cr.regexRam.FindStringSubmatch(str)
		if len(items) == 2 {
			min, err := strconv.Atoi(items[1])
			if err == nil {
				cr.memory.Min = float64(min)
				cr.memory.Refresh()
			}
		}
	}

	// Values
	// CPU
	str, ok = v.Properties["cpus"]
	if ok {
		val, err := strconv.Atoi(str)
		if err == nil {
			cr.cpus.SetValue(float64(val))
			cr.oldValues.cpus = val
		}
	}

	// CPUcap
	str, ok = v.Properties["cpuexecutioncap"]
	if ok {
		val, err := strconv.Atoi(str)
		if err == nil {
			cr.cpuCap.SetValue(float64(val))
			cr.oldValues.cpuCap = val
		}
	}

	// Memory
	str, ok = v.Properties["memory"]
	if ok {
		val, err := strconv.Atoi(str)
		if err == nil {
			cr.memory.SetValue(float64(val))
			cr.oldValues.ram = val
		}
	}

	// Pae
	util.CheckFromProperty(cr.pae, v, "pae", "on", &cr.oldValues.pae)

	// Nested VT
	util.CheckFromProperty(cr.nestedVT, v, "nested-hw-virt", "on", &cr.oldValues.nestedVt)

	// Nested Paging
	util.CheckFromProperty(cr.nestedPaging, v, "nestedpaging", "on", &cr.oldValues.nestedPaging)

	// x2APIC
	util.CheckFromProperty(cr.x2Apic, v, "x2apic", "on", &cr.oldValues.x2apic)

	// Paravirt Interface
	util.SelectEntryFromProperty(cr.paraVirtInterface, v, "paravirtprovider", cr.paraVirtMapStringToIndex, &cr.oldValues.paraVirtInterface)

	/*
		cr.memoryEntry.Validator = func(s string) error {
			val, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			min := cr.memory.Min
			max := cr.memory.Max
			if float64(val) > max {
				return errors.New("bigger than max")
			}
			if float64(val) < min {
				return errors.New("lesser than min")
			}
			return nil
		}
	*/
}

// called from status updates
func (cr *CpuRamTab) UpdateByStatus() {
	_, v := getActiveServerAndVm()
	if v != nil {
		state, err := v.GetState()
		if err != nil {
			return
		}
		switch state {
		case vm.RunState_unknown, vm.RunState_running, vm.RunState_paused, vm.RunState_meditation, vm.RunState_saved:
			cr.DisableAll()

		case vm.RunState_off, vm.RunState_aborted:
			cr.cpus.Enable()
			cr.cpusEntry.Enable()
			cr.cpuCap.Enable()
			cr.cpuCapEntry.Enable()
			cr.memory.Enable()
			cr.memoryEntry.Enable()
			cr.apply.Enable()
			cr.pae.Enable()
			cr.x2Apic.Enable()
			cr.nestedPaging.Enable()
			cr.nestedVT.Enable()
			cr.paraVirtInterface.Enable()

		default:
			SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
		}
	} else {
		cr.DisableAll()
	}
}

func (cr *CpuRamTab) DisableAll() {
	cr.cpus.Disable()
	cr.cpusEntry.Disable()
	cr.cpuCap.Disable()
	cr.cpuCapEntry.Disable()
	cr.memory.Disable()
	cr.memoryEntry.Disable()

	cr.pae.Disable()
	cr.x2Apic.Disable()
	cr.nestedPaging.Disable()
	cr.nestedVT.Disable()
	cr.paraVirtInterface.Disable()

	cr.apply.Disable()
}

func (cr *CpuRamTab) Apply() {
	s, v := getActiveServerAndVm()
	if v != nil {
		ResetStatus()
		if !cr.cpusEntry.Disabled() {
			n, err := strconv.Atoi(cr.cpusEntry.Text)
			if err == nil && n != cr.oldValues.cpus {
				go func() {
					err = v.SetCpus(&s.Client, n, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_cpuram.cpus.error", "Set number of CPUs for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						cr.oldValues.cpus = n
					}
				}()
			}
		}
		if !cr.cpuCapEntry.Disabled() {
			val, err := strconv.Atoi(cr.cpuCapEntry.Text)
			if err == nil && val != cr.oldValues.cpuCap {
				go func() {
					err = v.SetCPUExecCap(s, val, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_cpuram.cpucap.error", "Set cap for CPUs for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						cr.oldValues.cpuCap = val
					}
				}()
			}
		}
		if !cr.memoryEntry.Disabled() {
			val, err := strconv.Atoi(cr.memoryEntry.Text)
			if err == nil && val != cr.oldValues.ram {
				go func() {
					err = v.SetRam(&s.Client, val, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_cpuram.ram.error", "Set RAM for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						cr.oldValues.ram = val
					}
				}()
			}
		}
		if !cr.pae.Disabled() {
			if cr.pae.Checked != cr.oldValues.pae {
				go func() {
					err := v.SetPae(&s.Client, cr.pae.Checked, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_cpuram.pae.error", "Set PAE for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						cr.oldValues.pae = cr.pae.Checked
					}
				}()
			}
		}
		if !cr.nestedPaging.Disabled() {
			if cr.nestedPaging.Checked != cr.oldValues.nestedPaging {
				go func() {
					err := v.SetNestedPaging(s, cr.nestedPaging.Checked, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_cpuram.nestedpaging.error", "Set nested paging for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						cr.oldValues.nestedPaging = cr.nestedPaging.Checked
					}
				}()
			}
		}
		if !cr.nestedVT.Disabled() {
			if cr.nestedVT.Checked != cr.oldValues.nestedVt {
				go func() {
					err := v.SetNestedVirt(&s.Client, cr.nestedVT.Checked, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_cpuram.nestedvt.error", "Set nested VT-x/AMD-V for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						cr.oldValues.nestedVt = cr.nestedVT.Checked
					}
				}()
			}
		}
		if !cr.x2Apic.Disabled() {
			if cr.x2Apic.Checked != cr.oldValues.x2apic {
				go func() {
					err := v.SetX2Acpi(&s.Client, cr.x2Apic.Checked, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_cpuram.x2apic.error", "Set x2APIC for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						cr.oldValues.x2apic = cr.x2Apic.Checked
					}
				}()
			}
		}
		if !cr.paraVirtInterface.Disabled() {
			index := cr.paraVirtInterface.SelectedIndex()
			if index != cr.oldValues.paraVirtInterface {
				if index >= 0 {
					val, ok := cr.paraVirtMapIndexToType[index]
					if ok {
						go func() {
							err := v.SetParaVirtProvider(s, val, VMStatusUpdateCallBack)
							if err != nil {
								SetStatusText(fmt.Sprintf(lang.X("details.vm_cpuram.paravirt.error", "Set para virt. for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
							} else {
								cr.oldValues.paraVirtInterface = index
							}
						}()
					}
				}
			}
		}
	}
}
