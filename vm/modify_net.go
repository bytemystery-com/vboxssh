package vm

import (
	"fmt"
)

// ifNumber starts from 1 up to 8
func (m *VMachine) SetNetType(client *VmSshClient, ifNumber int, netType NetType, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("nic%d", ifNumber), netType, callBack)
}

func (m *VMachine) SetNetDevice(client *VmSshClient, ifNumber int, nicType NicType, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("nic-type%d", ifNumber), nicType, callBack)
}

func (m *VMachine) SetCableConnected(client *VmSshClient, ifNumber int, bConnected bool, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("cable-connected%d", ifNumber), bConnected, callBack)
}

func (m *VMachine) SetPromiscMode(client *VmSshClient, ifNumber int, promiscType PromiscType, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("nic-promisc%d", ifNumber), promiscType, callBack)
}

func (m *VMachine) SetBridgeAdapter(client *VmSshClient, ifNumber int, adapter string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("bridge-adapter%d", ifNumber), adapter, callBack)
}

func (m *VMachine) SetHostOnlyAdapter(client *VmSshClient, ifNumber int, adapter string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("host-only-adapter%d", ifNumber), adapter, callBack)
}

func (m *VMachine) SetNatAdapter(client *VmSshClient, ifNumber int, adapter string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("nat-network%d", ifNumber), adapter, callBack)
}

func (m *VMachine) SetInternalNetworkName(client *VmSshClient, ifNumber int, name string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("intnet%d", ifNumber), name, callBack)
}

func (m *VMachine) SetGenericNetworkName(client *VmSshClient, ifNumber int, name string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("nic-generic-drv%d", ifNumber), name, callBack)
}

func (m *VMachine) SetCloudNetworkName(client *VmSshClient, ifNumber int, name string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("cloud-network%d", ifNumber), name, callBack)
}

func (m *VMachine) SetMacAddress(client *VmSshClient, ifNumber int, mac string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("mac-address%d", ifNumber), mac, callBack)
}
