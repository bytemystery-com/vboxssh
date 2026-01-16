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

import "fyne.io/fyne/v2"

type DetailsInterface interface {
	UpdateBySelect()
	UpdateByStatus()
	Apply()
	DisableAll()
}

// wil be valled if selection has changed
func UpdateDetails() {
	if (Gui.ActiveItemServer != "" && Gui.ActiveItemServer != SERVER_ADD_NEW_UUID) && Gui.ActiveItemVm == "" {
		Gui.Details.CloseAll()
		Gui.Details.Open(0)
	} else if Gui.ActiveItemServer != "" && Gui.ActiveItemServer != SERVER_ADD_NEW_UUID && Gui.ActiveItemVm != "" {
		Gui.Details.Close(0)
		s, v := getActiveServerAndVm()
		if s != nil && v != nil {
			v.UpdateStatusEx(&s.Client)
		}
		flag := false
		for _, item := range Gui.Details.Items {
			if item.Open {
				flag = true
				break
			}
		}
		if !flag {
			Gui.Details.Open(1)
		}
	}

	for _, item := range Gui.DetailObjs {
		item.UpdateBySelect()
	}
}

// called when status where updated
func UpdateDetailsStatus() {
	if Gui.ActiveItemServer != "" && Gui.ActiveItemVm != "" {
		for _, item := range Gui.DetailObjs {
			item.UpdateByStatus()
		}
	}
}

func OpenTaskDetails() {
	fyne.Do(func() {
		Gui.Details.CloseAll()
		Gui.Details.Open(3)
	})
}
