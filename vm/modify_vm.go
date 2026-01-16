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

package vm

import (
	"fmt"
)

func (m *VMachine) SetCpus(client *VmSshClient, cpus int, callBack func(uuid string)) error {
	return m.setProperty(client, "cpus", cpus, callBack)
}

func (m *VMachine) SetRam(client *VmSshClient, memory int, callBack func(uuid string)) error {
	return m.setProperty(client, "memory", memory, callBack)
}

func (m *VMachine) SetPae(client *VmSshClient, pae bool, callBack func(uuid string)) error {
	return m.setProperty(client, "pae", pae, callBack)
}

func (m *VMachine) SetNestedVirt(client *VmSshClient, nestedVirt bool, callBack func(uuid string)) error {
	return m.setProperty(client, "nested-hw-virt", nestedVirt, callBack)
}

func (m *VMachine) SetX2Acpi(client *VmSshClient, x2acpi bool, callBack func(uuid string)) error {
	return m.setProperty(client, "x2apic", x2acpi, callBack)
}

func (m *VMachine) SetNestedPaging(client *VmSshClient, nestedPaging bool, callBack func(uuid string)) error {
	return m.setProperty(client, "nested-paging", nestedPaging, callBack)
}

func (m *VMachine) SetParaVirtProvider(client *VmSshClient, paraVirtProvider ParaVirtProviderType, callBack func(uuid string)) error {
	return m.setProperty(client, "paravirt-provider", paraVirtProvider, callBack)
}

func (m *VMachine) SetCPUExecCap(client *VmSshClient, cpuExecCap int, callBack func(uuid string)) error {
	return m.setProperty(client, "cpu-execution-cap", cpuExecCap, callBack)
}

func (m *VMachine) SetChipset(client *VmSshClient, chipSet ChipSetType, callBack func(uuid string)) error {
	return m.setProperty(client, "chipset", chipSet, callBack)
}

func (m *VMachine) SetTpm(client *VmSshClient, tpm TpmType, callBack func(uuid string)) error {
	return m.setProperty(client, "tpm-type", tpm, callBack)
}

func (m *VMachine) SetMouse(client *VmSshClient, mouse MouseType, callBack func(uuid string)) error {
	return m.setProperty(client, "mouse", mouse, callBack)
}

func (m *VMachine) SetKeyboard(client *VmSshClient, keyboard KeyboardType, callBack func(uuid string)) error {
	return m.setProperty(client, "keyboard", keyboard, callBack)
}

func (m *VMachine) SetAcpi(client *VmSshClient, acpi bool, callBack func(uuid string)) error {
	return m.setProperty(client, "acpi", acpi, callBack)
}

func (m *VMachine) SetIoApic(client *VmSshClient, ioAcpi bool, callBack func(uuid string)) error {
	return m.setProperty(client, "ioacpi", ioAcpi, callBack)
}

func (m *VMachine) SetHPet(client *VmSshClient, hpet bool, callBack func(uuid string)) error {
	return m.setProperty(client, "hpet", hpet, callBack)
}

func (m *VMachine) SetUseUtc(client *VmSshClient, useUtc bool, callBack func(uuid string)) error {
	return m.setProperty(client, "rtc-use-utc", useUtc, callBack)
}

func (m *VMachine) SetFirmware(client *VmSshClient, firmware FirmwareType, callBack func(uuid string)) error {
	return m.setProperty(client, "firmware", firmware, callBack)
}

// n starts at 1 !!!
func (m *VMachine) SetBootOrder(client *VmSshClient, device int, boot BootType, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("boot%d", device), boot, callBack)
}

func (m *VMachine) SetVideoRamSize(client *VmSshClient, size int, callBack func(uuid string)) error {
	return m.setProperty(client, "vram", size, callBack)
}

func (m *VMachine) SetMonitorCounts(client *VmSshClient, monitors int, callBack func(uuid string)) error {
	return m.setProperty(client, "monitor-count", monitors, callBack)
}

func (m *VMachine) SetVgaController(client *VmSshClient, vga VgaType, callBack func(uuid string)) error {
	return m.setProperty(client, "graphicscontroller", vga, callBack)
}

func (m *VMachine) SetAccelerate3D(client *VmSshClient, bAccel bool, callBack func(uuid string)) error {
	return m.setProperty(client, "accelerate-3d", bAccel, callBack)
}

func (m *VMachine) SetAccelerate2D(client *VmSshClient, bAccel bool, callBack func(uuid string)) error {
	return m.setProperty(client, "accelerate-2d", bAccel, callBack)
}

func (m *VMachine) SetName(client *VmSshClient, name string, callBack func(uuid string)) error {
	return m.setProperty(client, "name", client.quoteArgString(name), callBack)
}

func (m *VMachine) SetOsType(client *VmSshClient, osType string, callBack func(uuid string)) error {
	return m.setProperty(client, "ostype", osType, callBack)
}

func (m *VMachine) SetDescription(client *VmSshClient, description string, callBack func(uuid string)) error {
	if client.IsLocal {
		// description = description
	} else {
		description = "$'" + description + "'"
	}
	return m.setProperty(client, "description", description, callBack)
}

func (m *VMachine) SetProcessPriority(client *VmSshClient, processPriority ProcessPriorityType, callBack func(uuid string)) error {
	return m.setPropertyEx(client, "controlvm", "vm-process-priority", processPriority, callBack)
}

func (m *VMachine) SetBiosTimeOffset(client *VmSshClient, offset int, callBack func(uuid string)) error {
	return m.setProperty(client, "bios-system-time-offset", offset, callBack)
}

func (m *VMachine) DeleteVm(client *VmSshClient, del bool) error {
	opt := []string{"unregistervm", m.UUID}
	if del {
		opt = append(opt, "--delete-all")
	}

	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, nil)
	_ = lines
	return err
}
