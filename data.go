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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"bytemystery-com/vboxssh/crypt"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/storage"
)

func saveServers(servers []*vm.VmServer, masterKey string) error {
	list := make([]vm.VmServer, 0, len(servers))
	pass, err := crypt.Decrypt(crypt.InternPassword, masterKey)
	if err != nil {
		return err
	}
	for _, ss := range servers {
		s := vm.VmServer{
			Server:           ss.Server,
			FloppyImagesPath: ss.FloppyImagesPath,
			DvdImagesPath:    ss.DvdImagesPath,
			HddImagesPath:    ss.HddImagesPath,
			OvaPath:          ss.OvaPath,
		}
		x, err := crypt.Encrypt(pass, s.Password)
		if err != nil {
			return err
		}
		s.Password = x
		list = append(list, s)
	}
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	j := string(b)
	Gui.Settings.ServerList = j
	Gui.Settings.Store()

	SetStatusText(lang.X("data.serverlist.saved", "Server list was saved"), MsgInfo)
	return nil
}

func readKeyFile(file string) ([]byte, error) {
	// Windows: needed ???
	if filepath.VolumeName(file) != "" {
		data, err := os.ReadFile(file)
		return data, err
	}

	// URI -> use Fyne storage
	uri, err := storage.ParseURI(file)
	if err == nil && uri.Scheme() != "" {
		// URI -> Fyne storage will be used
		r, err := storage.Reader(uri)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	// Fallback
	data, err := os.ReadFile(file)
	return data, err
}

func loadServers(masterKey string) ([]vm.VmServer, error) {
	var list []vm.VmServer
	pass, err := crypt.Decrypt(crypt.InternPassword, masterKey)
	if err != nil {
		return nil, err
	}
	j := Gui.Settings.ServerList
	err = json.Unmarshal([]byte(j), &list)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(list); i++ {
		x, err := crypt.Decrypt(pass, list[i].Password)
		if err == nil {
			list[i].Password = x
		} else {
			list[i].Password = ""
			fmt.Println("!!! Unable to decrypt !!!")
		}
		list[i].KeyFileReader = readKeyFile
		list[i].HostFileReader = readKeyFile
	}
	SetStatusText(fmt.Sprintf(lang.X("data.serverlist.loaded", "Server list with %d entries was loaded"), len(list)), MsgInfo)

	return list, nil
}

func SaveServers() {
	saveServers(Data.GetServers(true), Gui.MasterPassword)
}
