package vm

func (m *VMachine) SetEnableRde(client *VmSshClient, bRde bool, callBack func(uuid string)) error {
	return m.setProperty(client, "vrde", bRde, callBack)
}

func (m *VMachine) SetRdePorts(client *VmSshClient, ports string, callBack func(uuid string)) error {
	return m.setProperty(client, "vrde-port", ports, callBack)
}

func (m *VMachine) SetRdeMultiConnection(client *VmSshClient, multi bool, callBack func(uuid string)) error {
	return m.setProperty(client, "vrde-multi-con", multi, callBack)
}

func (m *VMachine) SetRdeReuseConnection(client *VmSshClient, reuse bool, callBack func(uuid string)) error {
	return m.setProperty(client, "vrde-reuse-con", reuse, callBack)
}

func (m *VMachine) SetRdeSecurityMethode(client *VmSshClient, security RdpSecurityType, callBack func(uuid string)) error {
	return m.setProperty(client, "vrde-property=Security/Method", security, callBack)
}

func (m *VMachine) SetRdeAuthType(client *VmSshClient, auth RdpAuthType, callBack func(uuid string)) error {
	return m.setProperty(client, "vrde-auth-type", auth, callBack)
}
