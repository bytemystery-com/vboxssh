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

const (
	PREF_TREE_UPDATE_TIME_KEY   = "tree.update_time"
	PREF_TREE_UPDATE_TIME_VALUE = 20000

	PREF_TREE_UPDATE_DELAY_KEY   = "tree.update_delay"
	PREF_TREE_UPDATE_DELAY_VALUE = 1000

	PREF_TASKS_MAX_ENTRIES_KEY   = "tasks.max_entries"
	PREF_TASKS_MAX_ENTRIES_VALUE = 15

	PREF_TASKS_FIRST_START_KEY = "firststart"

	PREF_SERVERS_KEY          = "serverlist"
	PREF_MASTERKEY_TEST_KEY   = "mastertest"
	PREF_MASTERKEY_TEST_VALUE = "Reiner"
)

type Preferences struct {
	TreeUpdateTime  int // msec
	TreeDelayTime   int // msec
	TasksMaxEntries int
	ServerList      string // json String
	MasterKeyTest   string
	FirstStart      bool
}

func NewPreferences() *Preferences {
	p := &Preferences{
		TreeUpdateTime:  Gui.App.Preferences().IntWithFallback(PREF_TREE_UPDATE_TIME_KEY, PREF_TREE_UPDATE_TIME_VALUE),
		TreeDelayTime:   Gui.App.Preferences().IntWithFallback(PREF_TREE_UPDATE_DELAY_KEY, PREF_TREE_UPDATE_DELAY_VALUE),
		TasksMaxEntries: Gui.App.Preferences().IntWithFallback(PREF_TASKS_MAX_ENTRIES_KEY, PREF_TASKS_MAX_ENTRIES_VALUE),
		ServerList:      Gui.App.Preferences().StringWithFallback(PREF_SERVERS_KEY, ""),
		MasterKeyTest:   Gui.App.Preferences().StringWithFallback(PREF_MASTERKEY_TEST_KEY, PREF_MASTERKEY_TEST_VALUE),
		FirstStart:      Gui.App.Preferences().BoolWithFallback(PREF_TASKS_FIRST_START_KEY, true),
	}
	return p
}

func (p *Preferences) Store() {
	pref := Gui.App.Preferences()
	pref.SetInt(PREF_TREE_UPDATE_TIME_KEY, p.TreeUpdateTime)
	pref.SetInt(PREF_TREE_UPDATE_DELAY_KEY, p.TreeDelayTime)
	pref.SetInt(PREF_TASKS_MAX_ENTRIES_KEY, p.TasksMaxEntries)
	pref.SetString(PREF_SERVERS_KEY, p.ServerList)
	pref.SetString(PREF_MASTERKEY_TEST_KEY, p.MasterKeyTest)
	pref.SetBool(PREF_TASKS_FIRST_START_KEY, p.FirstStart)
}
