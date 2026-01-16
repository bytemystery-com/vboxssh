package main

import (
	"errors"
	"fmt"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
)

func addNewServer() {
	Gui.ServerSshTab.SetAddNewMode(true)
}

func isServerSelected() bool {
	if Gui.ActiveItemServer != "" && Gui.ActiveItemServer != SERVER_ADD_NEW_UUID && Gui.ActiveItemVm == "" {
		return true
	} else {
		return false
	}
}

func doDeleteServer() {
	deleteServer()
}

func deleteServer() error {
	if !isServerSelected() {
		return errors.New("no server selected")
	}
	s := Data.GetServer(Gui.ActiveItemServer, true)
	if s == nil {
		return errors.New("server not found")
	}
	if s.IsConnected() {
		s.Disonnect(&s.Client.Client)
	}
	Data.RemoveData(s)
	Gui.Tree.Refresh()
	SaveServers()
	SetStatusText(fmt.Sprintf(lang.X("status.server_delete_ok", "Server '%s' where removed."), s.Name), MsgInfo)
	return nil
}

func doConnectServer() {
	connectServer()
	Gui.Tree.Refresh()
}

func canConnectServer() (*vm.VmServer, error) {
	if !isServerSelected() {
		return nil, errors.New("no server selected")
	}
	s := Data.GetServer(Gui.ActiveItemServer, true)
	if s == nil {
		return nil, errors.New("server not found")
	}
	if s.IsLocal() {
		return nil, errors.New("server is local")
	}
	if s.IsConnected() {
		return nil, errors.New("server already connected")
	}
	return s, nil
}

func connectServer() error {
	s, err := canConnectServer()
	if err != nil {
		return err
	}
	go s.Connect(func() {
		setAfterConnectStatus(s, nil)
	}, func(err error) {
		setAfterConnectStatus(s, err)
	})
	return nil
}

func doDisconnectServer() {
	disconnectServer()
	Gui.Tree.Refresh()
	UpdateUI()
}

func canDisconnectServer() (*vm.VmServer, error) {
	if !isServerSelected() {
		return nil, errors.New("no server selected")
	}
	s := Data.GetServer(Gui.ActiveItemServer, true)
	if s == nil {
		return nil, errors.New("server not found")
	}
	if s.IsLocal() {
		return nil, errors.New("server is local")
	}
	if !s.IsConnected() {
		return nil, errors.New("server already disconnected")
	}
	return s, nil
}

func disconnectServer() error {
	s, err := canDisconnectServer()
	if err != nil {
		return err
	}
	err = s.Disonnect(&s.Client.Client)
	if err == nil {
		SetStatusText(fmt.Sprintf(lang.X("status.server_disconnect_ok", "Disconnect for server '%s'."), s.Name), MsgInfo)
	} else {
		SetStatusText(fmt.Sprintf(lang.X("status.server_disconnect_error", "Disconnect for server '%s' failed. (%s)"), s.Name, err.Error()), MsgError)
	}
	return err
}

func doReconnectServer() {
	reconnectServer()
	Gui.Tree.Refresh()
}

func canReconnectServer() (*vm.VmServer, error) {
	if !isServerSelected() {
		return nil, errors.New("no server selected")
	}
	s := Data.GetServer(Gui.ActiveItemServer, true)
	if s == nil {
		return nil, errors.New("server not found")
	}
	if s.IsLocal() {
		return nil, errors.New("server is local")
	}
	return s, nil
}

func reconnectServer() error {
	s, err := canReconnectServer()
	if err != nil {
		return err
	}
	err = s.Reconnect(&s.Client.Client)
	if err == nil {
		SetStatusText(fmt.Sprintf(lang.X("status.server_reconnect_ok", "Reconnect for server '%s'."), s.Name), MsgInfo)
	} else {
		SetStatusText(fmt.Sprintf(lang.X("status.server_reconnect_error", "Reconnect for server '%s' failed. (%s)"), s.Name, err.Error()), MsgError)
	}
	return err
}

func setAfterConnectStatus(s *vm.VmServer, err error) {
	if err == nil {
		SetStatusText(fmt.Sprintf(lang.X("status.server_connect_ok", "Connect for server '%s'."), s.Name), MsgInfo)
		fyne.Do(func() { UpdateUI() })
	} else {
		SetStatusText(fmt.Sprintf(lang.X("status.server_connect_error", "Connect for server '%s' failed. (%s)"), s.Name, err.Error()), MsgError)
		fyne.Do(func() { UpdateUI() })
	}
	treeRefresh()
}
