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
	"errors"
	"regexp"
	"slices"
)

var regex100prozent = regexp.MustCompile(`100%`)

func (m *VMachine) setPropertyInternal(client *VmSshClient, cmds []string, bUpdateStatus bool, callBack func(uuid string)) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	lines, err := m.runCmd(client, VBOXMANAGE_APP, cmds, bUpdateStatus, callBack)
	if err != nil {
		return err
	}
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	if slices.ContainsFunc(lines, regex100prozent.MatchString) {
		return nil
	}
	return errors.New("set property error")
}

func (m *VMachine) setProperty(client *VmSshClient, tag string, value any, callBack func(uuid string)) error {
	strVal, err := argTranslate(value)
	if err != nil {
		return err
	}
	return m.setPropertyInternal(client, []string{"modifyvm", m.UUID, "--" + tag + "=" + strVal}, true, callBack)
}

func (m *VMachine) setPropertyEx(client *VmSshClient, cmd string, tag string, value any, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, cmd, []any{tag, value}, callBack)
}

func (m *VMachine) setPropertyEx2(client *VmSshClient, cmd string, options []any, callBack func(uuid string)) error {
	opStr := make([]string, 0, len(options)+1)
	opStr = append(opStr, cmd)

	for _, value := range options {
		strVal, err := argTranslate(value)
		if err != nil {
			return err
		}
		if strVal != "" {
			opStr = append(opStr, strVal)
		}
	}
	return m.setPropertyInternal(client, opStr, true, callBack)
}
