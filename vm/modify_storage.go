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

func (m *VMachine) RemoveStorageController(client *VmSshClient, controllerName string, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, "storagectl", []any{m.UUID, "--name=" + client.quoteArgString(controllerName), "--remove"}, callBack)
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
	return m.setPropertyEx2(client, "storagectl", []any{m.UUID, "--name=" + client.quoteArgString(controllerName), controllerName, "--add", bus, "--controller", chipSet, "--portcount", ports, "--bootable", bootable}, callBack)
}
