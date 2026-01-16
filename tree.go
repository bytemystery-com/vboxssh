package main

import (
	"image/color"
	"strings"
	"sync"
	"time"

	"bytemystery-com/vboxssh/util"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func buildVmTreeNodeID(serverUuid, vmUuid string) widget.TreeNodeID {
	return serverUuid + "/" + vmUuid
}

// var count int = 0

func getVmAndServerUuidFromVmTreeNodeID(id widget.TreeNodeID) (string, string) {
	s := string(id)
	ids := strings.Split(s, "/")
	if len(ids) == 2 {
		return ids[0], ids[1]
	}
	return s, ""
}

// return childs
func treeGetChilds(id widget.TreeNodeID) []widget.TreeNodeID {
	/*
		fmt.Println(count, "GetChilds", id)
		count++
	*/
	if id == "" {
		childs := make([]widget.TreeNodeID, 0, len(Data.ServerMap))
		for _, item := range Data.GetServers(true) {
			childs = append(childs, item.UUID)
		}
		return childs
	} else {
		sid, vmid := getVmAndServerUuidFromVmTreeNodeID(id)
		if vmid != "" {
			return nil
		}
		Data.Lock.RLock()
		vms := Data.GetVms(sid, true, false)
		Data.Lock.RUnlock()

		if vms == nil {
			// fmt.Println("No Vms yet", id)
			go treeUpdateVmList(sid)
			return nil
		}

		childs := make([]widget.TreeNodeID, 0, len(vms))
		for _, v := range vms {
			childs = append(childs, buildVmTreeNodeID(sid, v.UUID))
		}
		return childs
	}
}

// return is Branch
func treeIsBranche(id widget.TreeNodeID) bool {
	/*.Println(count, "IsBranch", id)
	count++
	*/
	if id == "" {
		return true
	}
	s := Data.GetServer(string(id), true)
	if s == nil {
		return false
	} else {
		return s.IsConnected()
	}
}

// create canvasObject
func createCanvasObject(branche bool) fyne.CanvasObject {
	var tColor color.Color
	var tScale float32
	var tStyle fyne.TextStyle
	if branche {
		tColor = theme.Color(theme.ColorNamePrimary)
		tScale = 1.0
		tStyle = fyne.TextStyle{Bold: true}
	} else {
		tColor = theme.Color(theme.ColorNameForeground)
		tScale = 1.0
		tStyle = fyne.TextStyle{}
	}
	text := canvas.NewText("", tColor)
	text.TextSize = theme.TextSize() * tScale
	text.TextStyle = tStyle
	text.Refresh()
	icon := canvas.NewImageFromResource(theme.QuestionIcon())
	icon.SetMinSize(fyne.NewSize(16, 16)) // gewünschte Größe
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	return container.NewHBox(text, layout.NewSpacer(), icon, util.NewFiller(24, 0))
}

// update
func treeUpdateItem(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
	c, ok := o.(*fyne.Container)
	if !ok {
		return
	}
	text, ok := c.Objects[0].(*canvas.Text)
	if !ok {
		return
	}

	icon, ok := c.Objects[2].(*canvas.Image)
	if !ok {
		return
	}

	var tStyle fyne.TextStyle
	var tColor color.Color
	var tScale float32
	if branch {
		tColor = theme.Color(theme.ColorNamePrimary)
		tScale = 1.0
		tStyle = fyne.TextStyle{Bold: true}
	} else {
		tColor = theme.Color(theme.ColorNameForeground)
		tScale = 1.0
		tStyle = fyne.TextStyle{}
	}
	text.TextSize = theme.TextSize() * tScale
	text.TextStyle = tStyle
	text.Color = tColor

	sid, vmid := getVmAndServerUuidFromVmTreeNodeID(id)
	Data.Lock.RLock()
	if vmid == "" {
		s := Data.GetServer(sid, false)
		text.Text = s.Name
		text.Refresh()
		if s.IsConnected() {
			icon.Resource = Gui.IconOk
		} else {
			icon.Resource = Gui.IconError
		}
		icon.SetMinSize(fyne.NewSize(16, 16))
		icon.FillMode = canvas.ImageFillContain
		icon.Refresh()
		Data.Lock.RUnlock()
	} else {
		if vmid == "" {
			Data.Lock.RUnlock()
			return
		}
		vma := Data.GetVm(sid, vmid, false)
		if vma == nil {
			icon.Resource = Gui.IconUnknown
			Data.Lock.RUnlock()
		} else {
			text.Text = vma.Name
			text.Refresh()
			state, err := vma.GetState()
			if err != nil {
				icon.Resource = Gui.IconUnknown
				Data.Lock.RUnlock()
				if vma.Name != "<inaccessible>" {
					go treeUpdateVmStatus(sid, vmid, true)
				}
			} else {
				switch state {
				case vm.RunState_off:
					icon.Resource = Gui.IconStop
				case vm.RunState_running:
					icon.Resource = Gui.IconRun
				case vm.RunState_paused:
					icon.Resource = Gui.IconPause
				case vm.RunState_saved:
					icon.Resource = Gui.IconSave
				case vm.RunState_aborted:
					icon.Resource = Gui.IconAbort
				case vm.RunState_meditation:
					icon.Resource = Gui.IconMeditation
				default:
					icon.Resource = Gui.IconUnknown
				}
				Data.Lock.RUnlock()

			}
		}
		icon.SetMinSize(fyne.NewSize(16, 16))
		icon.FillMode = canvas.ImageFillContain
		icon.Refresh()
	}
}

func treeUpdateVmStatus(serverUuid, vmUuid string, lock bool) {
	if lock {
		Data.Lock.RLock()
		defer Data.Lock.RUnlock()
	}
	s := Data.GetServer(serverUuid, lock)
	if s == nil {
		return
	}
	v := Data.GetVm(serverUuid, vmUuid, lock)
	if v == nil {
		return
	}

	err := v.UpdateStatus(&s.Client, VMStatusUpdateCallBack)
	if err != nil {
		return
	}
}

func treeUpdateAllVms(s *vm.VmServer, delay int) {
	Data.Lock.RLock()
	vms := Data.GetVms(s.UUID, false, false)
	Data.Lock.RUnlock()
	if len(vms) > 0 {
		delay /= len(vms)
		for _, vma := range vms {
			treeUpdateVmStatus(s.UUID, vma.UUID, false)
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}
	}
}

func treeUpdateAll(delay int) bool {
	var wg sync.WaitGroup
	wasDone := false
	for _, s := range Data.ServerMap {
		if Gui.Tree.IsBranchOpen(s.UUID) && s.IsConnected() {
			// fmt.Println("Status update for", s.Name)
			wasDone = true
			wg.Go(func() {
				treeUpdateAllVms(s, delay)
			})
		}
	}
	wg.Wait()
	treeRefresh()
	return wasDone
}

func treeUpdateVmList(serverUuid string) {
	err := Data.UpdateVmList(serverUuid)
	if err != nil {
		return
	}
	// fmt.Println("Vms fetched - Refresh")
	treeRefresh()
}

func treeBranchOpened(id widget.TreeNodeID) {
	s := Data.GetServer(string(id), true)
	if s != nil {
		go treeUpdateVmList(s.UUID)
	}
}

func treeOnSelected(id widget.TreeNodeID) {
	sid, vmid := getVmAndServerUuidFromVmTreeNodeID(id)
	treeSetSelectedItem(sid, vmid)
}

func treeUpdateTimerProc() {
	for {
		// fmt.Println("Update tree status")
		if !treeUpdateAll(Gui.Settings.TreeUpdateTime) {
			time.Sleep(time.Duration(Gui.Settings.TreeDelayTime) * time.Millisecond)
		}
	}
}

func treeSetSelectedItem(sid string, vmid string) {
	Gui.ActiveItemServer = sid
	Gui.ActiveItemVm = vmid
	treeUpdateVmStatus(sid, vmid, true)
	UpdateUI()
}

func treeRefresh() {
	fyne.Do(func() {
		Gui.Tree.Refresh()
	})
}
