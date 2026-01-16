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
