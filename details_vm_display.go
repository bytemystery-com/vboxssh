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

type oldDisplayType struct {
	vgaRam        int
	accel3D       bool
	vgaController int
}

type DisplayTab struct {
	oldValues oldDisplayType

	ram        *widget.Slider
	ramEntry   *widget.Entry
	a3D        *widget.Check
	controller *widget.Select

	apply   *widget.Button
	tabItem *container.TabItem

	controllerMapStringToIndex map[string]int
	controllerMapIndexToType   map[int]vm.VgaType

	regexRam *regexp.Regexp
}

var _ DetailsInterface = (*DisplayTab)(nil)

func NewDisplayTab() *DisplayTab {
	displayTab := DisplayTab{
		controllerMapStringToIndex: map[string]int{"none": 0, "vboxvga": 1, "vmsvga": 2, "vboxsvga": 3},
		controllerMapIndexToType:   map[int]vm.VgaType{0: vm.Vga_none, 1: vm.Vga_vboxvga, 2: vm.Vga_vmsvga, 3: vm.Vga_vboxsvga},
		regexRam:                   regexp.MustCompile(`^([0-9]+)\s*.*`),
	}

	displayTab.apply = widget.NewButton(lang.X("details.vm_display.apply", "Apply"), func() {
		displayTab.Apply()
	})
	displayTab.apply.Importance = widget.HighImportance

	formWidth := util.GetFormWidth()
	displayTab.ramEntry = widget.NewEntry()

	displayTab.ram = widget.NewSlider(1, 256)
	displayTab.ram.Step = 1
	displayTab.ram.OnChanged = func(val float64) {
		displayTab.ramEntry.SetText(fmt.Sprintf("%.0f", val))
	}
	displayTab.ramEntry.OnChanged = util.GetNumberFilter(displayTab.ramEntry, func(s string) {
		val, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		displayTab.ram.SetValue(float64(val))
	})

	displayTab.ramEntry.SetText(fmt.Sprintf("%.0f", displayTab.ram.Min))

	displayTab.controller = widget.NewSelect([]string{
		lang.X("details.vm_display.vga.none", "None"),
		lang.X("details.vm_display.vga.vboxvga", "VBoxVGA"),
		lang.X("details.vm_display.vga.legacy", "VMSVGA"),
		lang.X("details.vm_display.vga.minimal", "VBoxSVGA"),
	}, nil)

	displayTab.a3D = widget.NewCheck(lang.X("details.vm_display.3dacceleration", "3D acceleration"), func(checked bool) {})

	unitSize := util.GetDefaultTextSize("XXXXXX")
	entrySize := util.GetDefaultTextSize("XXXXXXXX")
	entrySize.Height = displayTab.ramEntry.MinSize().Height
	grid1 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_display.ram", "RAM")),
		container.NewBorder(nil, nil, nil, container.NewHBox(container.NewGridWrap(entrySize, displayTab.ramEntry),
			container.NewGridWrap(unitSize, widget.NewLabel(lang.X("details.vm_display.ram.mb", "MB")))),
			displayTab.ram),
	)

	grid2 := container.New(layout.NewFormLayout(),
		displayTab.a3D, util.NewFiller(0, 0),
		widget.NewLabel(lang.X("details.vm_display.vga", "Graphics controller")),
		displayTab.controller,
	)

	gridWrap1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap2 := container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)

	gridWrap := container.NewVBox(util.NewVFiller(0.5), gridWrap1, gridWrap2)

	c := container.NewVBox(container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), displayTab.apply, util.NewFiller(32, 0)))
	displayTab.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.display", "Display"), c)
	return &displayTab
}

// calles by selection change
func (display *DisplayTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		display.DisableAll()
		return
	}
	display.apply.Enable()

	sysprop, err := s.GetSystemProperties(false)
	if err != nil {
		return
	}
	_ = v

	// min video RAM
	str, ok := sysprop["Minimum video RAM size"]
	if ok {
		items := display.regexRam.FindStringSubmatch(str)
		if len(items) == 2 {
			min, err := strconv.Atoi(items[1])
			if err == nil {
				display.ram.Min = float64(min)
				display.ram.Refresh()
			}
		}
	}

	// max video RAM
	str, ok = sysprop["Maximum video RAM size"]
	if ok {
		items := display.regexRam.FindStringSubmatch(str)
		if len(items) == 2 {
			max, err := strconv.Atoi(items[1])
			if err == nil {
				display.ram.Max = float64(max)
				display.ram.Refresh()
			}
		}
	}

	// Values
	// VGA RAM
	str, ok = v.Properties["vram"]
	if ok {
		val, err := strconv.Atoi(str)
		if err == nil {
			display.ram.SetValue(float64(val))
			display.oldValues.vgaRam = val
		}
	}

	// 3D
	util.CheckFromProperty(display.a3D, v, "accelerate3d", "on", &display.oldValues.accel3D)

	// Graphics controller
	util.SelectEntryFromProperty(display.controller, v, "graphicscontroller", display.controllerMapStringToIndex, &display.oldValues.vgaController)
}

// called from status updates
func (display *DisplayTab) UpdateByStatus() {
	_, v := getActiveServerAndVm()
	if v != nil {
		state, err := v.GetState()
		if err != nil {
			return
		}
		switch state {
		case vm.RunState_unknown, vm.RunState_running, vm.RunState_paused, vm.RunState_meditation, vm.RunState_saved:
			display.DisableAll()

		case vm.RunState_off, vm.RunState_aborted:
			display.ram.Enable()
			display.ramEntry.Enable()
			display.a3D.Enable()
			display.controller.Enable()
			display.apply.Enable()
		default:
			SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
		}
	} else {
		display.DisableAll()
	}
}

func (display *DisplayTab) DisableAll() {
	display.ram.Disable()
	display.ramEntry.Disable()
	display.a3D.Disable()
	display.controller.Disable()
	display.apply.Disable()
}

func (display *DisplayTab) Apply() {
	s, v := getActiveServerAndVm()
	if v != nil {
		ResetStatus()
		if !display.ramEntry.Disabled() {
			n, err := strconv.Atoi(display.ramEntry.Text)
			if err == nil && n != display.oldValues.vgaRam {
				err = v.SetVideoRamSize(&s.Client, n, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_display.setvgaram.error", "Set vram size for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					display.oldValues.vgaRam = n
				}
			}
		}
		if !display.a3D.Disabled() {
			if display.a3D.Checked != display.oldValues.accel3D {
				err := v.SetAccelerate3D(&s.Client, display.a3D.Checked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_display.set3daccel.error", "Set 3D acceleration for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					display.oldValues.accel3D = display.a3D.Checked
				}
			}
		}
		if !display.controller.Disabled() {
			index := display.controller.SelectedIndex()
			if index != display.oldValues.vgaController {
				if index >= 0 {
					val, ok := display.controllerMapIndexToType[index]
					if ok {
						err := v.SetVgaController(&s.Client, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_display.setvgacontroller.error", "Set VGA controller for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							display.oldValues.vgaController = index
						}
					}
				}
			}
		}
	}
}
