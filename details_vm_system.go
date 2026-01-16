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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type oldSystemType struct {
	chipset        int
	tpm            int
	mouse          int
	keyboard       int
	ioapic         bool
	clockInUtc     bool
	firmware       int
	secureboot     bool
	acpi           bool
	hpet           bool
	biosTimeOffset int
	bootEntries    []*bootEntry
}

type bootEntry struct {
	name      string
	checked   bool
	entryType vm.BootType
}

type SystemTab struct {
	oldValues oldSystemType

	chipset  *widget.Select
	mouse    *widget.Select
	keyboard *widget.Select

	ioApic          *widget.Check
	acpi            *widget.Check
	hpet            *widget.Check
	clockInUtc      *widget.Check
	firmware        *widget.Select
	secureBoot      *widget.Check
	biosTimeOffset  *widget.Entry
	bootUp          *widget.Button
	bootDown        *widget.Button
	bootList        *widget.List
	bootListOverlay *widget.Button

	apply   *widget.Button
	tabItem *container.TabItem

	bootEntries       []*bootEntry
	selectedBootEntry int

	chipsetMapStringToIndex  map[string]int
	chipsetMapIndexToType    map[int]vm.ChipSetType
	mouseMapStringToIndex    map[string]int
	mouseMapIndexToType      map[int]vm.MouseType
	firmwareMapStringToIndex map[string]int
	firmwareMapIndexToType   map[int]vm.FirmwareType
	keyboardMapStringToIndex map[string]int
	keyboardMapIndexToType   map[int]vm.KeyboardType
}

var _ DetailsInterface = (*SystemTab)(nil)

func NewSystemTab() *SystemTab {
	sysTab := SystemTab{
		chipsetMapStringToIndex:  map[string]int{"piix3": 0, "ich9": 1},
		chipsetMapIndexToType:    map[int]vm.ChipSetType{0: vm.ChipSet_piix3, 1: vm.ChipSet_ich9},
		mouseMapStringToIndex:    map[string]int{"usbmouse": 0, "ps2mouse": 1, "usbtablet": 2, "usbmultitouch": 3, "unknown": 4},
		mouseMapIndexToType:      map[int]vm.MouseType{0: vm.Mouse_usb, 1: vm.Mouse_ps2, 2: vm.Mouse_usbtablet, 3: vm.Mouse_usbmultitouch, 4: vm.Mouse_usbmtscreenpluspad},
		firmwareMapStringToIndex: map[string]int{"bios": 0, "efi": 1, "efi32": 2, "efi64": 3},
		firmwareMapIndexToType:   map[int]vm.FirmwareType{0: vm.Firmware_bios, 1: vm.Firmware_efi, 2: vm.Firmware_efi32, 3: vm.Firmware_efi64},
		keyboardMapStringToIndex: map[string]int{"ps2kbd": 0, "usbkbd": 1},
		keyboardMapIndexToType:   map[int]vm.KeyboardType{0: vm.Keyboard_ps2, 1: vm.Keyboard_usb},
	}
	sysTab.apply = widget.NewButton(lang.X("details.vm_system.apply", "Apply"), func() {
		sysTab.Apply()
	})
	sysTab.apply.Importance = widget.HighImportance

	formWidth := util.GetFormWidth()
	sysTab.chipset = widget.NewSelect([]string{
		lang.X("details.vm_system.chipset.piix3", "PIIX3"),
		lang.X("details.vm_system.chipset.ICH9", "ICH9"),
	}, func(s string) {})

	sysTab.mouse = widget.NewSelect([]string{
		lang.X("details.vm_system.mouse.usbmouse", "USB Mouse"),
		lang.X("details.vm_system.mouse.ps2mouse", "PS/2 Mouse"),
		lang.X("details.vm_system.mouse.usbtablet", "USB Tablet"),
		lang.X("details.vm_system.mouse.usbmultitouchtablet", "USB Multi-Touch Tablet"),
		lang.X("details.vm_system.mouse.usbmultitouchtouchscreentouchpad", "USB Multi-Touch TouchScreen and TouchPad"),
	}, nil)

	sysTab.keyboard = widget.NewSelect([]string{
		lang.X("details.vm_system.keyboard.ps2", "PS/2"),
		lang.X("details.vm_system.keyboard.usb", "USB"),
	}, nil)

	sysTab.ioApic = widget.NewCheck(lang.X("details.vm_system.ioapic", "I/O APIC"), nil)
	sysTab.acpi = widget.NewCheck(lang.X("details.vm_system.acpi", "ACPI"), nil)
	sysTab.hpet = widget.NewCheck(lang.X("details.vm_system.hpet", "High Precision Event Timer (HPET)"), nil)
	sysTab.clockInUtc = widget.NewCheck(lang.X("details.vm_system.clockinutc", "Clock in UTC"), nil)
	sysTab.firmware = widget.NewSelect([]string{
		lang.X("details.vm_system.firmware.bios", "BIOS"),
		lang.X("details.vm_system.firmware.efi", "EFI"),
		lang.X("details.vm_system.firmware.efi32", "EFI32"),
		lang.X("details.vm_system.firmware.efi64", "EFI64"),
	}, func(s string) {
		sysTab.UpdateByStatus()
		sysTab.setEnableDisableBootOptions()
	})
	sysTab.secureBoot = widget.NewCheck(lang.X("details.vm_system.secureboot", "Secure Boot"), nil)
	sysTab.biosTimeOffset = widget.NewEntry()
	sysTab.biosTimeOffset.SetPlaceHolder(lang.X("details.vm_system.biostiemoffset_placeholder", "> 0 guest VM time runs ahead"))
	sysTab.biosTimeOffset.OnChanged = util.GetNumberFilterPlusMinus(sysTab.biosTimeOffset, nil)

	grid1 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_system.chipset", "Chipset")), sysTab.chipset,
		widget.NewLabel(lang.X("details.vm_system.keyboard", "Keyboard")), sysTab.keyboard,
		widget.NewLabel(lang.X("details.vm_system.mouse", "Mouse")), sysTab.mouse,
		widget.NewLabel(lang.X("details.vm_system.firmware", "Firmware")), sysTab.firmware,
		sysTab.acpi, sysTab.hpet,
		sysTab.ioApic, sysTab.clockInUtc,
		sysTab.secureBoot, util.NewFiller(0, 0),
	)

	sysTab.bootUp = widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		if sysTab.selectedBootEntry > 0 {
			sysTab.bootEntries[sysTab.selectedBootEntry], sysTab.bootEntries[sysTab.selectedBootEntry-1] = sysTab.bootEntries[sysTab.selectedBootEntry-1], sysTab.bootEntries[sysTab.selectedBootEntry]
			sysTab.bootList.Refresh()
			sysTab.selectedBootEntry -= 1
			sysTab.bootList.Select(sysTab.selectedBootEntry)
		}
	})
	sysTab.bootDown = widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		if sysTab.selectedBootEntry < len(sysTab.bootEntries)-1 {
			sysTab.bootEntries[sysTab.selectedBootEntry], sysTab.bootEntries[sysTab.selectedBootEntry+1] = sysTab.bootEntries[sysTab.selectedBootEntry+1], sysTab.bootEntries[sysTab.selectedBootEntry]
			sysTab.bootList.Refresh()
			sysTab.selectedBootEntry += 1
			sysTab.bootList.Select(sysTab.selectedBootEntry)
		}
	})
	sysTab.bootList = widget.NewList(func() int {
		return len(sysTab.bootEntries)
	}, func() fyne.CanvasObject {
		return container.NewBorder(nil, nil, widget.NewCheck("", nil), nil, widget.NewLabel(""))
	}, func(id widget.ListItemID, obj fyne.CanvasObject) {
		c, ok := obj.(*fyne.Container)
		if ok {
			w := c.Objects
			l, _ := w[0].(*widget.Label)
			ch, _ := w[1].(*widget.Check)
			l.SetText(sysTab.bootEntries[id].name)
			ch.SetChecked(sysTab.bootEntries[id].checked)
			ch.OnChanged = func(changed bool) {
				sysTab.bootEntries[id].checked = changed
			}
		}
	})

	sysTab.bootList.OnSelected = func(id widget.ListItemID) {
		sysTab.selectedBootEntry = id
	}
	sysTab.bootListOverlay = widget.NewButton("", nil)
	sysTab.bootListOverlay.Disable()

	listContent := container.NewStack(sysTab.bootList, sysTab.bootListOverlay)
	sysTab.bootListOverlay.Hide()

	sysTab.bootEntries = make([]*bootEntry, 4)
	sysTab.bootEntries = []*bootEntry{
		{
			name:      lang.X("details.vm_system.bootentry.floppy", "Floppy"),
			checked:   true,
			entryType: vm.Boot_floppy,
		},
		{
			name:      lang.X("details.vm_system.bootentry.optical", "Optical"),
			checked:   true,
			entryType: vm.Boot_dvd,
		},
		{
			name:      lang.X("details.vm_system.bootentry.harddisk", "Harddisk"),
			checked:   true,
			entryType: vm.Boot_disk,
		},
		{
			name:      lang.X("details.vm_system.bootentry.network", "Network"),
			checked:   true,
			entryType: vm.Boot_net,
		},
	}

	grid2 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_system.biostimeoffset", "BIOS system time offset")), container.NewBorder(nil, nil, nil, widget.NewLabel("msec"), sysTab.biosTimeOffset),
		widget.NewLabel(lang.X("details.vm_system.biosbootorder", "BIOS boot order")), container.NewBorder(nil, nil, nil, container.NewVBox(sysTab.bootUp, sysTab.bootDown, util.NewFiller(0, 2.1*sysTab.bootDown.MinSize().Height)), listContent),
	)
	gridWrap1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap2 := container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)

	gridWrap := container.NewVBox(util.NewVFiller(0.5), gridWrap1, gridWrap2)

	c := container.NewVBox(container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), sysTab.apply, util.NewFiller(32, 0)))

	sysTab.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.system", "System"), c)

	fyne.Do(func() {
		sysTab.bootList.Select(0)
	})

	return &sysTab
}

// calles by selection change
func (sys *SystemTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()

	if s == nil || v == nil {
		sys.DisableAll()
		return
	}
	sys.apply.Enable()

	// Chipset
	util.SelectEntryFromProperty(sys.chipset, v, "chipset", sys.chipsetMapStringToIndex, &sys.oldValues.chipset)

	// Keyboard
	util.SelectEntryFromProperty(sys.keyboard, v, "hidkeyboard", sys.keyboardMapStringToIndex, &sys.oldValues.keyboard)

	// Mouse
	util.SelectEntryFromProperty(sys.mouse, v, "hidpointing", sys.mouseMapStringToIndex, &sys.oldValues.mouse)

	// ACPI
	util.CheckFromProperty(sys.acpi, v, "acpi", "on", &sys.oldValues.acpi)

	// HPET
	util.CheckFromProperty(sys.hpet, v, "hpet", "on", &sys.oldValues.hpet)

	// I/O APIC
	util.CheckFromProperty(sys.ioApic, v, "ioapic", "on", &sys.oldValues.ioapic)

	// UTC
	util.CheckFromProperty(sys.clockInUtc, v, "rtcuseutc", "on", &sys.oldValues.clockInUtc)

	// Firmware
	util.SelectEntryFromProperty(sys.firmware, v, "firmware", sys.firmwareMapStringToIndex, &sys.oldValues.firmware)

	// Secure Boot
	util.CheckFromProperty(sys.secureBoot, v, "SecureBoot", "on", &sys.oldValues.secureboot)

	// Time offset
	str, ok := v.Properties["biossystemtimeoffset"]
	if ok {
		sys.biosTimeOffset.SetText(str)
		sys.oldValues.biosTimeOffset, _ = strconv.Atoi(str)
	}

	for _, item := range sys.bootEntries {
		item.checked = false
	}

	for i := range len(sys.bootEntries) {
		str, ok = v.Properties[fmt.Sprintf("boot%d", i+1)]
		switch str {
		case "dvd":
			sys.setBootEntry(vm.Boot_dvd, i)
		case "floppy":
			sys.setBootEntry(vm.Boot_floppy, i)
		case "disk":
			sys.setBootEntry(vm.Boot_disk, i)
		case "net":
			sys.setBootEntry(vm.Boot_net, i)
		}
	}
	sys.setOldValuesBootEntries()
	sys.bootList.Refresh()

	sys.UpdateByStatus()
}

func (sys *SystemTab) setOldValuesBootEntries() {
	sys.oldValues.bootEntries = make([]*bootEntry, 0, len(sys.bootEntries))
	for _, item := range sys.bootEntries {
		sys.oldValues.bootEntries = append(sys.oldValues.bootEntries, &bootEntry{
			checked:   item.checked,
			entryType: item.entryType,
		})
	}
}

func (sys *SystemTab) hasBootEntriesChanged() bool {
	if len(sys.bootEntries) != len(sys.oldValues.bootEntries) {
		return true
	}
	for index, item := range sys.bootEntries {
		item2 := sys.oldValues.bootEntries[index]
		if item.entryType != item2.entryType {
			return true
		}
		if item.checked != item2.checked {
			return true
		}
	}
	return false
}

func (sys *SystemTab) findBootEntry(t vm.BootType) int {
	for index, item := range sys.bootEntries {
		if item.entryType == t {
			return index
		}
	}
	return -1
}

func (sys *SystemTab) setBootEntry(t vm.BootType, index int) {
	i := sys.findBootEntry(t)
	sys.bootEntries[i].checked = true
	if i == index {
		return
	}
	sys.bootEntries[i], sys.bootEntries[index] = sys.bootEntries[index], sys.bootEntries[i]
}

// called from status updates
func (sys *SystemTab) UpdateByStatus() {
	_, v := getActiveServerAndVm()
	if v != nil {
		state, err := v.GetState()
		if err != nil {
			return
		}
		switch state {
		case vm.RunState_unknown, vm.RunState_meditation:
			sys.DisableAll()

		case vm.RunState_running, vm.RunState_paused, vm.RunState_saved:
			sys.chipset.Disable()
			sys.mouse.Disable()
			sys.keyboard.Disable()

			sys.ioApic.Disable()
			sys.clockInUtc.Disable()
			sys.firmware.Disable()
			sys.secureBoot.Disable()
			sys.acpi.Disable()
			sys.hpet.Disable()
			sys.biosTimeOffset.Disable()

		case vm.RunState_off, vm.RunState_aborted:
			sys.chipset.Enable()
			sys.mouse.Enable()
			sys.keyboard.Enable()
			sys.biosTimeOffset.Enable()

			sys.acpi.Enable()
			sys.hpet.Enable()
			sys.ioApic.Enable()
			sys.clockInUtc.Enable()
			sys.firmware.Enable()
			if sys.firmware.SelectedIndex() > 0 {
				sys.secureBoot.Enable()
			} else {
				sys.secureBoot.Disable()
			}

		default:
			SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
		}
	} else {
		sys.DisableAll()
	}
}

func (sys *SystemTab) DisableAll() {
	sys.chipset.Disable()
	sys.acpi.Disable()
	sys.hpet.Disable()
	sys.mouse.Disable()
	sys.keyboard.Disable()
	sys.ioApic.Disable()
	sys.clockInUtc.Disable()
	sys.firmware.Disable()
	sys.secureBoot.Disable()
	sys.biosTimeOffset.Disable()
	sys.bootUp.Disable()
	sys.bootDown.Disable()
	sys.bootListOverlay.Show()

	sys.apply.Disable()
}

func (sys *SystemTab) setEnableDisableBootOptions() {
	index := sys.firmware.SelectedIndex()
	if index >= 0 {
		val, ok := sys.firmwareMapIndexToType[index]
		if ok {
			if val == vm.Firmware_bios {
				sys.bootUp.Enable()
				sys.bootDown.Enable()
				sys.bootListOverlay.Hide()
			} else {
				sys.bootUp.Disable()
				sys.bootDown.Disable()
				sys.bootListOverlay.Show()
			}
		}
	}
}

func (sys *SystemTab) Apply() {
	s, v := getActiveServerAndVm()
	if v != nil {
		ResetStatus()

		// Chipset
		if !sys.chipset.Disabled() {
			index := sys.chipset.SelectedIndex()
			if index != sys.oldValues.chipset {
				if index >= 0 {
					val, ok := sys.chipsetMapIndexToType[index]
					if ok {
						err := v.SetChipset(&s.Client, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_system.chipset.error", "Set chipset for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							sys.oldValues.chipset = index
						}
					}
				}
			}
		}

		// Keyboard
		if !sys.keyboard.Disabled() {
			index := sys.keyboard.SelectedIndex()
			if index != sys.oldValues.keyboard {
				if index >= 0 {
					val, ok := sys.keyboardMapIndexToType[index]
					if ok {
						err := v.SetKeyboard(&s.Client, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_system.keyboard.error", "Set keyboard for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							sys.oldValues.keyboard = index
						}
					}
				}
			}
		}

		// Mouse
		if !sys.mouse.Disabled() {
			index := sys.mouse.SelectedIndex()
			if index != sys.oldValues.mouse {
				if index >= 0 {
					val, ok := sys.mouseMapIndexToType[index]
					if ok {
						err := v.SetMouse(&s.Client, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_system.mouse.error", "Set mouse for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							sys.oldValues.mouse = index
						}
					}
				}
			}
		}

		// I/O APIC
		if !sys.ioApic.Disabled() {
			if sys.ioApic.Checked != sys.oldValues.ioapic {
				err := v.SetIoApic(&s.Client, sys.ioApic.Checked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_system.ioapic.error", "Set I/O API for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					sys.oldValues.ioapic = sys.ioApic.Checked
				}
			}
		}

		// ACPI
		if !sys.acpi.Disabled() {
			if sys.acpi.Checked != sys.oldValues.acpi {
				err := v.SetAcpi(&s.Client, sys.acpi.Checked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_system.acpi.error", "Set ACPI for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					sys.oldValues.acpi = sys.acpi.Checked
				}
			}
		}

		// HPET
		if !sys.hpet.Disabled() {
			if sys.hpet.Checked != sys.oldValues.hpet {
				err := v.SetHPet(&s.Client, sys.hpet.Checked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_system.hpet.error", "Set HPET for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					sys.oldValues.hpet = sys.hpet.Checked
				}
			}
		}

		// UTC
		if !sys.clockInUtc.Disabled() {
			if sys.clockInUtc.Checked != sys.oldValues.clockInUtc {
				err := v.SetUseUtc(&s.Client, sys.clockInUtc.Checked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_system.utc.error", "Set UTC for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					sys.oldValues.clockInUtc = sys.clockInUtc.Checked
				}
			}
		}

		// UEFI
		if !sys.firmware.Disabled() {
			index := sys.firmware.SelectedIndex()
			if index != sys.oldValues.firmware {
				if index >= 0 {
					val, ok := sys.firmwareMapIndexToType[index]
					if ok {
						err := v.SetFirmware(&s.Client, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_system.firmware.error", "Set firmware for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							sys.oldValues.firmware = index
						}
					}
				}
			}
		}

		// Secure boot
		if !sys.secureBoot.Disabled() {
			if sys.secureBoot.Checked != sys.oldValues.secureboot {
				err := v.SetSecureBoot(&s.Client, sys.secureBoot.Checked, true, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_system.secureboot.error", "Set secure boot for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					sys.oldValues.secureboot = sys.secureBoot.Checked
				}
			}
		}

		// BIOS offset time
		if !sys.biosTimeOffset.Disabled() {
			val, err := strconv.Atoi(sys.biosTimeOffset.Text)
			if err == nil {
				if val != sys.oldValues.biosTimeOffset {
					err := v.SetBiosTimeOffset(&s.Client, val, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_system.biostimeoffset.error", "Set BIOS time offset for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						sys.oldValues.biosTimeOffset = val
					}
				}
			}
		}
		// BIOS boot order
		if sys.bootListOverlay.Hidden && sys.hasBootEntriesChanged() {
			bError := false
			i := 1
			for _, item := range sys.bootEntries {
				if item.checked {
					err := v.SetBootOrder(&s.Client, i, item.entryType, VMStatusUpdateCallBack)
					i += 1
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_system.biosbootorder.error", "Set BIOS boot order for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						bError = true
					}
				}
			}
			for _, item := range sys.bootEntries {
				if !item.checked {
					err := v.SetBootOrder(&s.Client, i, vm.Boot_none, VMStatusUpdateCallBack)
					i += 1
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_system.biosbootorder.error", "Set BIOS boot order for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						bError = true
					}
				}
			}
			if !bError {
				sys.setOldValuesBootEntries()
			}
		}
	}
}
