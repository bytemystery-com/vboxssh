package main

import (
	"encoding/json"
	"fmt"

	"bytemystery-com/vboxssh/crypt"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2/lang"
)

func saveServers(servers map[string]*vm.VmServer, masterKey string) error {
	var list []vm.VmServer
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
	}
	SetStatusText(fmt.Sprintf(lang.X("data.serverlist.loaded", "Server list with %d entries was loaded"), len(list)), MsgInfo)

	return list, nil
}

func SaveServers() {
	saveServers(Data.ServerMap, Gui.MasterPassword)
}
