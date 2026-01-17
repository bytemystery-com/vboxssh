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

func (m *VMachine) EjectMedia(client *VmSshClient, controllerName string, storageType StorageType, port, device int, medium MediaSpecialType, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, "storageattach", []any{m.UUID, "--storagectl=" + client.quoteArgString(controllerName), "--type", storageType, "--port", port, "--device", device, "--medium", medium, "--forceunmount"}, callBack)
}

func (m *VMachine) DetachMedia(client *VmSshClient, controllerName string, port, device int, medium MediaSpecialType, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, "storageattach", []any{m.UUID, "--storagectl=" + client.quoteArgString(controllerName), "--port", port, "--device", device, "--medium", medium}, callBack)
}

func (m *VMachine) AttachMedia(client *VmSshClient, controllerName string, storageType StorageType, port, device int, medium string, isLive *bool, isSsd *bool, callBack func(uuid string)) error {
	opt := []any{m.UUID, "--storagectl=" + client.quoteArgString(controllerName), "--type", storageType, "--port", port, "--device", device, "--medium", medium}
	if isLive != nil {
		opt = append(opt, "--tempeject")
		opt = append(opt, *isLive)
	}
	if isSsd != nil {
		opt = append(opt, "--nonrotational")
		opt = append(opt, *isSsd)
	}
	return m.setPropertyEx2(client, "storageattach", opt, callBack)
}

func (m *VMachine) RemoveStorageController(client *VmSshClient, controllerName string, chipSet StorageChipsetType, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, "storagectl", []any{m.UUID, "--name=" + client.quoteArgString(controllerName), "--controller", chipSet, "--remove"}, callBack)
}

func (m *VMachine) AttachGuestAdditions(client *VmSshClient, controllerName string, port, device int, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, "storageattach", []any{m.UUID, "--storagectl=" + client.quoteArgString(controllerName), "--port", port, "--device", device, "--type", "dvddrive", "--medium", MediaSpecial_additions}, callBack)
}

func (m *VMachine) RenameStorageController(client *VmSshClient, controllerOldName, controllerNewName string, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, "storagectl", []any{m.UUID, "--name", client.quoteArgString(controllerOldName), "--rename", client.quoteArgString(controllerNewName)}, callBack)
}

func (m *VMachine) SetStorageControllerBootable(client *VmSshClient, controllerName string, bootable bool, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, "storagectl", []any{m.UUID, "--name=" + client.quoteArgString(controllerName), "--bootable", bootable}, callBack)
}

func (m *VMachine) AddStorageController(client *VmSshClient, controllerName string, bus StorageBusType, chipSet StorageChipsetType, ports int, bootable bool, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, "storagectl", []any{m.UUID, "--name=" + client.quoteArgString(controllerName), "--add", bus, "--controller", chipSet, "--portcount", ports, "--bootable", bootable}, callBack)
}
