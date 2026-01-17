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
	"io"
)

func (m *VMachine) TakeSnapshot(client *VmSshClient, name, description string, live bool, statusWriter io.Writer) error {
	opt := []string{"snapshot", m.UUID, "take", client.quoteArgString(name)}
	if description != "" {
		if !client.IsLocal {
			description = "$'" + description + "'"
		}
	}
	opt = append(opt, description)
	if live {
		opt = append(opt, "--live")
	}
	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, statusWriter)
	if err != nil {
		m.addLogEntry(lines, false)
	}
	return err
}

func (m *VMachine) DeleteSnapshot(client *VmSshClient, uuid string, statusWriter io.Writer) error {
	opt := []string{"snapshot", m.UUID, "delete", client.quoteArgString(uuid)}
	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, statusWriter)
	if err != nil {
		m.addLogEntry(lines, false)
	}
	return err
}

func (m *VMachine) RestoreSnapshot(client *VmSshClient, uuid string, statusWriter io.Writer) error {
	opt := []string{"snapshot", m.UUID, "restore", client.quoteArgString(uuid)}
	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, statusWriter)
	if err != nil {
		m.addLogEntry(lines, false)
	}
	return err
}
