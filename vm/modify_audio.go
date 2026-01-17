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

func (m *VMachine) SetAudioEnabled(v *VmServer, audioEnabled bool, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		if !audioEnabled {
			return m.SetAudioDriver(v, AudioDriver_none, callBack)
		} else {
			return nil
		}
	} else {
		return m.setProperty(&v.Client, "audio-enabled", audioEnabled, callBack)
	}
}

func (m *VMachine) SetAudioController(v *VmServer, audioController AudioControllerType, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, "audiocontroller", audioController, callBack)
	} else {
		return m.setProperty(&v.Client, "audio-controller", audioController, callBack)
	}
}

func (m *VMachine) SetAudioCodec(v *VmServer, audioCodec AudioCodecType, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, "audiocodec", audioCodec, callBack)
	} else {
		return m.setProperty(&v.Client, "audio-codec", audioCodec, callBack)
	}
}

func (m *VMachine) SetAudioInEnabled(v *VmServer, audioInEnabled bool, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, "audioin", audioInEnabled, callBack)
	} else {
		return m.setProperty(&v.Client, "audio-in", audioInEnabled, callBack)
	}
}

func (m *VMachine) SetAudioOutEnabled(v *VmServer, audioOutEnabled bool, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, "audioout", audioOutEnabled, callBack)
	} else {
		return m.setProperty(&v.Client, "audio-out", audioOutEnabled, callBack)
	}
}

func (m *VMachine) SetAudioDriver(v *VmServer, audioDriver AudioDriverType, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		return m.setProperty(&v.Client, "audio", audioDriver, callBack)
	} else {
		return m.setProperty(&v.Client, "audio-driver", audioDriver, callBack)
	}
}
