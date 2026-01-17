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

func (m *VMachine) SetEnableRde(client *VmSshClient, bRde bool, callBack func(uuid string)) error {
	return m.setProperty(client, "vrde", bRde, callBack)
}

func (m *VMachine) SetRdePorts(s *VmServer, ports string, callBack func(uuid string)) error {
	maj, _, _ := s.getVmVersion()
	if maj == 6 {
		return m.setProperty(&s.Client, "vrdeport", ports, callBack)
	} else {
		return m.setProperty(&s.Client, "vrde-port", ports, callBack)
	}
}

func (m *VMachine) SetRdeMultiConnection(s *VmServer, multi bool, callBack func(uuid string)) error {
	maj, _, _ := s.getVmVersion()
	if maj == 6 {
		return m.setProperty(&s.Client, "vrdemulticon", multi, callBack)
	} else {
		return m.setProperty(&s.Client, "vrde-multi-con", multi, callBack)
	}
}

func (m *VMachine) SetRdeReuseConnection(s *VmServer, reuse bool, callBack func(uuid string)) error {
	maj, _, _ := s.getVmVersion()
	if maj == 6 {
		return m.setProperty(&s.Client, "vrdereusecon", reuse, callBack)
	} else {
		return m.setProperty(&s.Client, "vrde-reuse-con", reuse, callBack)
	}
}

func (m *VMachine) SetRdeSecurityMethode(s *VmServer, security RdpSecurityType, callBack func(uuid string)) error {
	maj, _, _ := s.getVmVersion()
	if maj == 6 {
		return m.setProperty(&s.Client, "vrdeproperty=Security/Method", security, callBack)
	} else {
		return m.setProperty(&s.Client, "vrde-property=Security/Method", security, callBack)
	}
}

func (m *VMachine) SetRdeAuthType(s *VmServer, auth RdpAuthType, callBack func(uuid string)) error {
	maj, _, _ := s.getVmVersion()
	if maj == 6 {
		return m.setProperty(&s.Client, "vrdeauthtype", auth, callBack)
	} else {
		return m.setProperty(&s.Client, "vrde-auth-type", auth, callBack)
	}
}
