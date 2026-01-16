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
	"fyne.io/fyne/v2"
)

/*
func GrayIcon(res *fyne.StaticResource) *fyne.StaticResource {
	img, _, err := image.Decode(bytes.NewReader(res.Content()))
	if err != nil {
		return res
	}

	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, gray)

	return fyne.NewStaticResource(res.Name()+"_gray", buf.Bytes())
}
*/

func UpdateToolbarMenu() {
	s, m := getActiveServerAndVm()

	t := Gui.ToolbarActions["create"]
	if t != nil {
		if s != nil && s.IsConnected() {
			t.Enable()
		} else {
			t.Disable()
		}
	}

	t = Gui.ToolbarActions["delete"]
	if t != nil {
		if s != nil && s.IsConnected() && m != nil {
			t.Enable()
		} else {
			t.Disable()
		}
	}

	actions := []string{"export", "import"}
	for _, a := range actions {
		t := Gui.ToolbarActions[a]
		if t != nil {
			if s != nil && s.IsConnected() {
				var icon *fyne.StaticResource
				switch a {
				case "export":
					icon = Gui.IconExport
				case "import":
					icon = Gui.IconImport
				}
				t.SetIcon(icon)
				t.Enable()
			} else {
				var icon *fyne.StaticResource
				switch a {
				case "export":
					icon = Gui.IconExport_x
				case "import":
					icon = Gui.IconImport_x
				}
				t.SetIcon(icon)
				t.Disable()
			}
		}
	}

	actions = []string{"menu.machine.export", "menu.machine.import", "menu.machine.create"}
	for _, a := range actions {
		m := Gui.MenuItems[a]
		if m != nil {
			if s != nil && s.IsConnected() {
				m.Disabled = false
			} else {
				m.Disabled = true
			}
		}

	}

	if isServerSelected() {
		t := Gui.ToolbarActions["connect"]
		m := Gui.MenuItems["menu.server.connect"]
		if _, err := canConnectServer(); err == nil {
			if t != nil {
				t.Enable()
			}
			if m != nil {
				m.Disabled = false
			}
		} else {
			if t != nil {
				t.Disable()
			}
			if m != nil {
				m.Disabled = true
			}
		}

		t = Gui.ToolbarActions["disconnect"]
		m = Gui.MenuItems["menu.server.disconnect"]
		if _, err := canDisconnectServer(); err == nil {
			if t != nil {
				t.Enable()
			}
			if m != nil {
				m.Disabled = false
			}
		} else {
			if t != nil {
				t.Disable()
			}
			if m != nil {
				m.Disabled = true
			}
		}

		t = Gui.ToolbarActions["reconnect"]
		m = Gui.MenuItems["menu.server.reconnect"]
		if _, err := canReconnectServer(); err == nil {
			if t != nil {
				t.Enable()
			}
			if m != nil {
				m.Disabled = false
			}
		} else {
			if t != nil {
				t.Disable()
			}
			if m != nil {
				m.Disabled = true
			}
		}
	} else {
		actions := []string{"connect", "disconnect", "reconnect"}
		for _, a := range actions {
			t := Gui.ToolbarActions[a]
			if t != nil {
				t.Disable()
			}
		}

		actions = []string{"menu.server.connect", "menu.server.disconnect", "menu.server.reconnect"}
		for _, a := range actions {
			m := Gui.MenuItems[a]
			if m != nil {
				m.Disabled = true
			}
		}
	}
	Gui.MainMenu.Refresh()
}
