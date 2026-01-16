package vm

func (m *VMachine) SetSecureBoot(client *VmSshClient, secureBoot, doEnroll bool, callBack func(uuid string)) error {
	if secureBoot && doEnroll {
		err := m.setPropertyEx(client, "modifynvram", "enrollmssignatures", nil, callBack)
		if err != nil {
			return err
		}
		err = m.setPropertyEx(client, "modifynvram", "enrollorclpk", nil, callBack)
		if err != nil {
			return err
		}
	}
	return m.setPropertyEx(client, "modifynvram", "secureboot", secureBoot, callBack)
}

func (m *VMachine) EnrollDefPlatformKey(client *VmSshClient, callBack func(uuid string)) error {
	return m.setPropertyEx(client, "modifynvram", "enrollorclpk", nil, callBack)
}

func (m *VMachine) EnrollMsSignatures(client *VmSshClient, callBack func(uuid string)) error {
	return m.setPropertyEx(client, "modifynvram", "enrollmssignatures", nil, callBack)
}

func (m *VMachine) InitUefiVarStore(client *VmSshClient, callBack func(uuid string)) error {
	return m.setPropertyEx(client, "modifynvram", "inituefivarstore", nil, callBack)
}
