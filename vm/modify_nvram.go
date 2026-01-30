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

func (m *VMachine) SetSecureBoot(client *VmSshClient, secureBoot, doEnroll bool, callBack func(uuid string)) error {
	if secureBoot && doEnroll {
		err := m.setPropertyEx(client, "modifynvram", "enrollmssignatures", nil, callBack)
		if err != nil {
			// return err
		}
		err = m.setPropertyEx(client, "modifynvram", "enrollorclpk", nil, callBack)
		if err != nil {
			// return err
		}
	}
	s := ""
	if secureBoot {
		s = "--enable"
	} else {
		s = "--disable"
	}
	return m.setPropertyEx(client, "modifynvram", "secureboot", s, callBack)
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
