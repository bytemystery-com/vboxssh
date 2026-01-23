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

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
)

func ButtonStart() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}
	SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.start.started", "VM '%s' was started ..."), v.Name), MsgInfo)
	headless := false
	startMode, err := v.GetStartInWindow(s)
	if err == nil {
		if startMode == vm.StartInWindow_no || (startMode == vm.StartInWindow_default && !s.IsLocal()) {
			headless = true
		}
	}
	go v.Start(&s.Client, headless, func(err error) {
		if err != nil {
			SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.start.error", "Start of VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
			fyne.Do(func() {
				Gui.StartButton.SetDown(false)
			})
		} else {
			ResetStatus()
		}
	}, VMStatusUpdateCallBack)
}

func ButtonPause() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}
	go func() {
		err := v.Pause(&s.Client, VMStatusUpdateCallBack)
		if err != nil {
			SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.pause.error", "Pause of VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
		}
	}()
}

func ButtonSave() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}
	SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.save.started", "Save status of VM '%s' started ..."), v.Name), MsgInfo)
	go v.Save(&s.Client, func(err error) {
		if err != nil {
			SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.save.error", "Save status of VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
		} else {
			ResetStatus()
		}
	}, VMStatusUpdateCallBack)
}

func ButtonReset() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}
	go func() {
		err := v.Reset(&s.Client, VMStatusUpdateCallBack)
		if err != nil {
			SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.reset.error", "Reset of VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
		}
	}()
}

func ButtonOff() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}
	go func() {
		err := v.Off(&s.Client, VMStatusUpdateCallBack)
		if err != nil {
			SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.off.error", "Power off for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
		}
	}()
}

func ButtonShutdown() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}
	SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.shutdown.started", "Shutdown of VM '%s' started ..."), v.Name), MsgInfo)
	go v.Shutdown(&s.Client, func(err error) {
		if err != nil {
			SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.shutdown.error", "Shutdown of VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
		} else {
			ResetStatus()
		}
	}, VMStatusUpdateCallBack)
}

func ButtonDiscard() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}
	go func() {
		err := v.DiscardSaveState(&s.Client, VMStatusUpdateCallBack)
		if err != nil {
			SetStatusText(fmt.Sprintf(lang.X("details.vm_ctrl.discard.error", "Discard of VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
		}
	}()
}

func UpdateButtons() {
	v := Data.GetVm(Gui.ActiveItemServer, Gui.ActiveItemVm, true)
	if v != nil {
		state, err := v.GetState()
		if err != nil {
			return
		}
		switch state {
		case vm.RunState_unknown:
			disableAllButtons()

		case vm.RunState_running:
			Gui.StartButton.SetDown(true)
			Gui.StartButton.SetEnabled(false)
			Gui.PauseButton.SetDown(false)
			Gui.PauseButton.SetEnabled(true)
			Gui.SaveButton.SetDown(false)
			Gui.SaveButton.SetEnabled(true)
			Gui.OffButton.SetDown(false)
			Gui.OffButton.SetEnabled(true)
			Gui.ResetButton.SetDown(false)
			Gui.ResetButton.SetEnabled(true)
			Gui.ShutdownButton.SetDown(false)
			Gui.ShutdownButton.SetEnabled(true)
			Gui.DiscardButton.SetDown(false)
			Gui.DiscardButton.SetEnabled(false)

		case vm.RunState_off:
			Gui.StartButton.SetDown(false)
			Gui.StartButton.SetEnabled(true)
			Gui.PauseButton.SetDown(false)
			Gui.PauseButton.SetEnabled(false)
			Gui.SaveButton.SetDown(false)
			Gui.SaveButton.SetEnabled(false)
			Gui.OffButton.SetDown(true)
			Gui.OffButton.SetEnabled(false)
			Gui.ResetButton.SetDown(false)
			Gui.ResetButton.SetEnabled(false)
			Gui.ShutdownButton.SetDown(false)
			Gui.ShutdownButton.SetEnabled(false)
			Gui.DiscardButton.SetDown(false)
			Gui.DiscardButton.SetEnabled(false)

		case vm.RunState_paused:
			Gui.StartButton.SetDown(true)
			Gui.StartButton.SetEnabled(false)
			Gui.PauseButton.SetDown(true)
			Gui.PauseButton.SetEnabled(true)
			Gui.SaveButton.SetDown(false)
			Gui.SaveButton.SetEnabled(true)
			Gui.OffButton.SetDown(false)
			Gui.OffButton.SetEnabled(true)
			Gui.ResetButton.SetDown(false)
			Gui.ResetButton.SetEnabled(false)
			Gui.ShutdownButton.SetDown(false)
			Gui.ShutdownButton.SetEnabled(false)
			Gui.DiscardButton.SetDown(false)
			Gui.DiscardButton.SetEnabled(false)

		case vm.RunState_saved:
			Gui.StartButton.SetDown(false)
			Gui.StartButton.SetEnabled(true)
			Gui.PauseButton.SetDown(false)
			Gui.PauseButton.SetEnabled(false)
			Gui.SaveButton.SetDown(true)
			Gui.SaveButton.SetEnabled(false)
			Gui.OffButton.SetDown(false)
			Gui.OffButton.SetEnabled(false)
			Gui.ResetButton.SetDown(false)
			Gui.ResetButton.SetEnabled(false)
			Gui.ShutdownButton.SetDown(false)
			Gui.ShutdownButton.SetEnabled(false)
			Gui.DiscardButton.SetDown(false)
			Gui.DiscardButton.SetEnabled(true)

		case vm.RunState_aborted:
			Gui.StartButton.SetDown(false)
			Gui.StartButton.SetEnabled(true)
			Gui.PauseButton.SetDown(false)
			Gui.PauseButton.SetEnabled(false)
			Gui.SaveButton.SetDown(false)
			Gui.SaveButton.SetEnabled(false)
			Gui.OffButton.SetDown(false)
			Gui.OffButton.SetEnabled(false)
			Gui.ResetButton.SetDown(false)
			Gui.ResetButton.SetEnabled(false)
			Gui.ShutdownButton.SetDown(false)
			Gui.ShutdownButton.SetEnabled(false)
			Gui.DiscardButton.SetDown(false)
			Gui.DiscardButton.SetEnabled(false)

		case vm.RunState_meditation:
			Gui.StartButton.SetDown(false)
			Gui.StartButton.SetEnabled(true)
			Gui.PauseButton.SetDown(false)
			Gui.PauseButton.SetEnabled(false)
			Gui.SaveButton.SetDown(false)
			Gui.SaveButton.SetEnabled(false)
			Gui.OffButton.SetDown(false)
			Gui.OffButton.SetEnabled(true)
			Gui.ResetButton.SetDown(false)
			Gui.ResetButton.SetEnabled(false)
			Gui.ShutdownButton.SetDown(false)
			Gui.ShutdownButton.SetEnabled(false)
			Gui.DiscardButton.SetDown(false)
			Gui.DiscardButton.SetEnabled(true)
		default:
			SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
		}
	} else {
		disableAllButtons()
	}
}

func disableAllButtons() {
	Gui.StartButton.SetDown(false)
	Gui.StartButton.SetEnabled(false)
	Gui.PauseButton.SetDown(false)
	Gui.PauseButton.SetEnabled(false)
	Gui.SaveButton.SetDown(false)
	Gui.SaveButton.SetEnabled(false)
	Gui.OffButton.SetDown(false)
	Gui.OffButton.SetEnabled(false)
	Gui.ResetButton.SetDown(false)
	Gui.ResetButton.SetEnabled(false)
	Gui.ShutdownButton.SetDown(false)
	Gui.ShutdownButton.SetEnabled(false)
	Gui.DiscardButton.SetDown(false)
	Gui.DiscardButton.SetEnabled(false)
}
