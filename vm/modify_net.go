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

// ifNumber starts from 1 up to 8
func (m *VMachine) SetNetType(client *VmSshClient, ifNumber int, netType NetType, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("nic%d", ifNumber), netType, callBack)
}

func (m *VMachine) SetNetDevice(v *VmServer, ifNumber int, nicType NicType, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, fmt.Sprintf("nictype%d", ifNumber), nicType, callBack)
	} else {
		return m.setProperty(&v.Client, fmt.Sprintf("nic-type%d", ifNumber), nicType, callBack)
	}
}

func (m *VMachine) SetCableConnected(v *VmServer, ifNumber int, bConnected bool, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, fmt.Sprintf("cableconnected%d", ifNumber), bConnected, callBack)
	} else {
		return m.setProperty(&v.Client, fmt.Sprintf("cable-connected%d", ifNumber), bConnected, callBack)
	}
}

func (m *VMachine) SetPromiscMode(v *VmServer, ifNumber int, promiscType PromiscType, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, fmt.Sprintf("nicpromisc%d", ifNumber), promiscType, callBack)
	} else {
		return m.setProperty(&v.Client, fmt.Sprintf("nic-promisc%d", ifNumber), promiscType, callBack)
	}
}

func (m *VMachine) SetBridgeAdapter(v *VmServer, ifNumber int, adapter string, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, fmt.Sprintf("bridgeadapter%d", ifNumber), adapter, callBack)
	} else {
		return m.setProperty(&v.Client, fmt.Sprintf("bridge-adapter%d", ifNumber), adapter, callBack)
	}
}

func (m *VMachine) SetHostOnlyAdapter(v *VmServer, ifNumber int, adapter string, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, fmt.Sprintf("hostonlyadapter%d", ifNumber), adapter, callBack)
	} else {
		return m.setProperty(&v.Client, fmt.Sprintf("host-only-adapter%d", ifNumber), adapter, callBack)
	}
}

func (m *VMachine) SetNatAdapter(client *VmSshClient, ifNumber int, adapter string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("nat-network%d", ifNumber), adapter, callBack)
}

func (m *VMachine) SetInternalNetworkName(client *VmSshClient, ifNumber int, name string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("intnet%d", ifNumber), name, callBack)
}

func (m *VMachine) SetGenericNetworkName(v *VmServer, ifNumber int, name string, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, fmt.Sprintf("nicgenericdrv%d", ifNumber), name, callBack)
	} else {
		return m.setProperty(&v.Client, fmt.Sprintf("nic-generic-drv%d", ifNumber), name, callBack)
	}
}

func (m *VMachine) SetCloudNetworkName(client *VmSshClient, ifNumber int, name string, callBack func(uuid string)) error {
	return m.setProperty(client, fmt.Sprintf("cloud-network%d", ifNumber), name, callBack)
}

func (m *VMachine) SetMacAddress(v *VmServer, ifNumber int, mac string, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, fmt.Sprintf("macaddress%d", ifNumber), mac, callBack)
	} else {
		return m.setProperty(&v.Client, fmt.Sprintf("mac-address%d", ifNumber), mac, callBack)
	}
}
