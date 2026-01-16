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

package main

import (
	"fmt"
	"strings"

	"bytemystery-com/vboxssh/util"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type oldAudioType struct {
	driver     int
	codec      int
	controller int
	enabled    bool
	in         bool
	out        bool
}

type AudioTab struct {
	oldValues oldAudioType

	enabled    *widget.Check
	hostDriver *widget.Select
	controller *widget.Select
	codec      *widget.Select
	out        *widget.Check
	in         *widget.Check

	apply   *widget.Button
	tabItem *container.TabItem

	hostDriverMapStringToIndex map[string]int
	hostDriverMapIndexToType   map[int]vm.AudioDriverType
	controllerMapStringToIndex map[string]int
	controllerMapIndexToType   map[int]vm.AudioControllerType
	codecMapStringToIndex      map[string]int
	codecMapIndexToType        map[int]vm.AudioCodecType
}

var _ DetailsInterface = (*AudioTab)(nil)

func NewAudioTab() *AudioTab {
	audioTab := AudioTab{
		hostDriverMapStringToIndex: map[string]int{"default": 0, "alsa": 1, "oss": 2, "pulseaudio": 3, "null": 4},
		hostDriverMapIndexToType:   map[int]vm.AudioDriverType{0: vm.AudioDriver_default, 1: vm.AudioDriver_alsa, 2: vm.AudioDriver_oss, 3: vm.AudioDriver_pulse, 4: vm.AudioDriver_null},
		controllerMapStringToIndex: map[string]int{"ac97": 0, "hda": 1, "sb16": 2},
		controllerMapIndexToType:   map[int]vm.AudioControllerType{0: vm.AudioController_ac97, 1: vm.AudioController_hda, 2: vm.AudioController_sb16},
		codecMapStringToIndex:      map[string]int{"stac9700": 0, "ad1980": 1, "stac9221": 2, "sb16": 3},
		codecMapIndexToType:        map[int]vm.AudioCodecType{0: vm.AudioCodec_stac9700, 1: vm.AudioCodec_ad1980, 2: vm.AudioCodec_stac9221, 3: vm.AudioCodec_sb16},
	}

	audioTab.apply = widget.NewButton(lang.X("details.vm_audio.apply", "Apply"), func() {
		audioTab.Apply()
	})
	audioTab.apply.Importance = widget.HighImportance

	formWidth := util.GetFormWidth()
	audioTab.enabled = widget.NewCheck(lang.X("details.vm_audio.enabled", "Enabled"), func(checked bool) {
		audioTab.UpdateByStatus()
	})
	audioTab.out = widget.NewCheck(lang.X("details.vm_audio.out", "Output"), nil)
	audioTab.in = widget.NewCheck(lang.X("details.vm_audio.in", "Input"), nil)
	audioTab.out = widget.NewCheck(lang.X("details.vm_audio.out", "Output"), nil)
	audioTab.in = widget.NewCheck(lang.X("details.vm_audio.in", "Input"), nil)
	audioTab.hostDriver = widget.NewSelect([]string{
		lang.X("details.vm_audio.hostdriver.default", "Default"),
		lang.X("details.vm_audio.hostdriver.alsa", "ALSA"),
		lang.X("details.vm_audio.hostdriver.oss", "OSS"),
		lang.X("details.vm_audio.hostdriver.pulse", "Pulse"),
		lang.X("details.vm_audio.hostdriver.null", "Null"),
	}, nil)

	audioTab.controller = widget.NewSelect([]string{
		lang.X("details.vm_audio.controller.ac97", "ICH AC97"),
		lang.X("details.vm_audio.controller.intelhd", "Intel HD"),
		lang.X("details.vm_audio.controller.soundblaster", "Soundblaster 16 "),
	}, nil)

	audioTab.codec = widget.NewSelect([]string{
		lang.X("details.vm_audio.codec.stac9700", "SigmaTel STAC9700 (AC97)"),
		lang.X("details.vm_audio.codec.ad1980", "Analog Devices AD1980 (AC97)"),
		lang.X("details.vm_audio.codec.stac9221", "Conexant STAC9221 (Intel HD)"),
		lang.X("details.vm_audio.codec.sb16", "Sound Blaster 16 (Soundblaster 16)"),
	}, nil)

	grid1 := container.New(layout.NewFormLayout(),
		audioTab.enabled, util.NewFiller(0, 0),
		widget.NewLabel(lang.X("details.vm_audio.hostdriver", "Host driver")), audioTab.hostDriver,
		widget.NewLabel(lang.X("details.vm_audio.controller", "Controller")), audioTab.controller,
		widget.NewLabel(lang.X("details.vm_audio.codec", "Codec")), audioTab.codec,
		audioTab.out, audioTab.in,
	)
	gridWrap1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap := container.NewVBox(util.NewVFiller(0.5), gridWrap1)

	c := container.NewVBox(container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), audioTab.apply, util.NewFiller(32, 0)))
	audioTab.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.audio", "Audio"), c)
	return &audioTab
}

// calles by selection change
func (audio *AudioTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()

	if s == nil || v == nil {
		audio.DisableAll()
		return
	}
	audio.apply.Enable()
	// v.UpdateStatusEx(&s.Client)

	// Values
	// Enabled
	str, ok := v.Properties["audio"]
	if ok && strings.ToLower(str) != "none" {
		audio.enabled.SetChecked(true)
		audio.enableDisableAudioCtrls(true)
		audio.oldValues.enabled = true
	} else {
		audio.enabled.SetChecked(false)
		audio.enableDisableAudioCtrls(false)
		audio.oldValues.enabled = false
	}

	// Audio out
	util.CheckFromProperty(audio.out, v, "audio_out", "on", &audio.oldValues.out)

	// Audio in
	util.CheckFromProperty(audio.out, v, "audio_in", "on", &audio.oldValues.in)

	// Driver
	util.SelectEntryFromProperty(audio.hostDriver, v, "audio_driver", audio.hostDriverMapStringToIndex, &audio.oldValues.driver)

	// Controller
	util.SelectEntryFromProperty(audio.controller, v, "audio_controller", audio.controllerMapStringToIndex, &audio.oldValues.controller)

	// Codec
	util.SelectEntryFromProperty(audio.codec, v, "audio_codec", audio.codecMapStringToIndex, &audio.oldValues.codec)

	audio.UpdateByStatus()
}

// called from status updates
func (audio *AudioTab) UpdateByStatus() {
	_, v := getActiveServerAndVm()
	if v != nil {
		state, err := v.GetState()
		if err != nil {
			return
		}
		switch state {
		case vm.RunState_unknown, vm.RunState_meditation:
			audio.DisableAll()

		case vm.RunState_running:
			audio.enabled.Disable()
			audio.hostDriver.Disable()
			audio.controller.Disable()
			audio.codec.Disable()
			if audio.enabled.Checked {
				audio.out.Enable()
				audio.in.Enable()
			} else {
				audio.out.Disable()
				audio.in.Disable()
			}

		case vm.RunState_paused:
			audio.enabled.Disable()
			audio.controller.Disable()
			audio.codec.Disable()
			audio.hostDriver.Disable()
			if audio.enabled.Checked {
				audio.out.Enable()
				audio.in.Enable()
			} else {
				audio.out.Disable()
				audio.in.Disable()
			}

		case vm.RunState_saved:
			audio.enabled.Disable()
			audio.controller.Disable()
			audio.codec.Disable()
			if audio.enabled.Checked {
				audio.hostDriver.Enable()
				audio.out.Enable()
				audio.in.Enable()
			} else {
				audio.hostDriver.Disable()
				audio.out.Disable()
				audio.in.Disable()
			}

		case vm.RunState_off, vm.RunState_aborted:
			audio.enabled.Enable()
			audio.enableDisableAudioCtrls(audio.enabled.Checked)

		default:
			SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
		}
	} else {
		audio.DisableAll()
	}
}

func (audio *AudioTab) enableDisableAudioCtrls(enable bool) {
	if enable {
		audio.hostDriver.Enable()
		audio.controller.Enable()
		audio.codec.Enable()
		audio.out.Enable()
		audio.in.Enable()
	} else {
		audio.hostDriver.Disable()
		audio.controller.Disable()
		audio.codec.Disable()
		audio.out.Disable()
		audio.in.Disable()
	}
}

func (audio *AudioTab) DisableAll() {
	audio.enableDisableAudioCtrls(false)
	audio.enabled.Disable()

	audio.apply.Disable()
}

func (audio *AudioTab) Apply() {
	s, v := getActiveServerAndVm()
	if v != nil {
		ResetStatus()
		if !audio.enabled.Disabled() {
			if audio.enabled.Checked != audio.oldValues.enabled {
				err := v.SetAudioEnabled(&s.Client, audio.enabled.Checked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_audio.enableaudio.error", "Enable audio for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					audio.oldValues.enabled = audio.enabled.Checked
				}
			}
		}
		if !audio.out.Disabled() {
			if audio.out.Checked != audio.oldValues.out {
				err := v.SetAudioOutEnabled(&s.Client, audio.out.Checked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_audio.setaudioout.error", "Set audio output for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					audio.oldValues.out = audio.out.Checked
				}
			}
		}
		if !audio.in.Disabled() {
			if audio.in.Checked != audio.oldValues.in {
				err := v.SetAudioInEnabled(&s.Client, audio.in.Checked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_audio.setaudioin.error", "Set audio input for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					audio.oldValues.in = audio.in.Checked
				}
			}
		}

		if !audio.hostDriver.Disabled() {
			index := audio.hostDriver.SelectedIndex()
			if index != audio.oldValues.driver {
				if index >= 0 {
					val, ok := audio.hostDriverMapIndexToType[index]
					if ok {
						err := v.SetAudioDriver(&s.Client, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_audio.sethostdriver.error", "Set audio host driver for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							audio.oldValues.driver = index
						}
					}
				}
			}
		}
		if !audio.controller.Disabled() {
			index := audio.controller.SelectedIndex()
			if index != audio.oldValues.controller {
				if index >= 0 {
					val, ok := audio.controllerMapIndexToType[index]
					if ok {
						err := v.SetAudioController(&s.Client, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_audio.setcontroller.error", "Set audio controller for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							audio.oldValues.controller = index
						}
					}
				}
			}
		}
		if !audio.codec.Disabled() {
			index := audio.codec.SelectedIndex()
			if index != audio.oldValues.codec {
				if index >= 0 {
					val, ok := audio.codecMapIndexToType[index]
					if ok {
						err := v.SetAudioCodec(&s.Client, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_audio.setcodec.error", "Set audio codec for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							audio.oldValues.codec = index
						}
					}
				}
			}
		}
	}
}
