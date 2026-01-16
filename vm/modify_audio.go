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

func (m *VMachine) SetAudioEnabled(client *VmSshClient, audioEnabled bool, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-enabled", audioEnabled, callBack)
}

func (m *VMachine) SetAudioController(client *VmSshClient, audioController AudioControllerType, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-controller", audioController, callBack)
}

func (m *VMachine) SetAudioCodec(client *VmSshClient, audioCodec AudioCodecType, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-codec", audioCodec, callBack)
}

func (m *VMachine) SetAudioInEnabled(client *VmSshClient, audioInEnabled bool, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-in", audioInEnabled, callBack)
}

func (m *VMachine) SetAudioOutEnabled(client *VmSshClient, audioOutEnabled bool, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-out", audioOutEnabled, callBack)
}

func (m *VMachine) SetAudioDriver(client *VmSshClient, audioDriver AudioDriverType, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-driver", audioDriver, callBack)
}
