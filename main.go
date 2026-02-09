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
	"embed"
	"errors"
	"image/color"
	"net/http"
	"time"

	_ "net/http/pprof"

	"bytemystery-com/vboxssh/crypt"

	"bytemystery-com/vboxssh/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bytemystery-com/colorlabel"
	"github.com/bytemystery-com/picbutton"
)

//go:embed assets/*
var content embed.FS

const (
	NUMBER_OF_NICS = 8
)

type GUI struct {
	App               fyne.App
	MainWindow        fyne.Window
	Toolbar           *widget.Toolbar
	ToolbarActions    map[string]*widget.ToolbarAction
	MainMenu          *fyne.MainMenu
	MenuItems         map[string]*fyne.MenuItem
	OuterBorderLayout *fyne.Container
	StartusBar        *colorlabel.ColorLabel
	ContentLayout     *fyne.Container
	IsDesktop         bool
	Split             *container.Split
	Tree              *widget.Tree
	Details           *widget.Accordion
	DetailsScroll     *container.Scroll
	Icon              *fyne.StaticResource
	FyneSettings      fyne.Settings
	IconRun           *fyne.StaticResource
	IconStop          *fyne.StaticResource
	IconPause         *fyne.StaticResource
	IconSave          *fyne.StaticResource
	IconAbort         *fyne.StaticResource
	IconMeditation    *fyne.StaticResource
	IconUnknown       *fyne.StaticResource
	IconEmpty         *fyne.StaticResource
	IconOk            *fyne.StaticResource
	IconError         *fyne.StaticResource
	PicStartU         *fyne.StaticResource
	PicStartD         *fyne.StaticResource
	PicPauseU         *fyne.StaticResource
	PicPauseD         *fyne.StaticResource
	PicSaveU          *fyne.StaticResource
	PicSaveD          *fyne.StaticResource
	PicOffU           *fyne.StaticResource
	PicOffD           *fyne.StaticResource
	PicShutdownU      *fyne.StaticResource
	PicShutdownD      *fyne.StaticResource
	PicResetU         *fyne.StaticResource
	PicResetD         *fyne.StaticResource
	PicDiscardU       *fyne.StaticResource
	PicDiscardD       *fyne.StaticResource

	IconFloppy *fyne.StaticResource
	IconIde    *fyne.StaticResource
	IconPcie   *fyne.StaticResource
	IconSas    *fyne.StaticResource
	IconSata   *fyne.StaticResource
	IconScsi   *fyne.StaticResource
	IconUsb    *fyne.StaticResource
	IconVirt   *fyne.StaticResource

	IconCd    *fyne.StaticResource
	IconHdd   *fyne.StaticResource
	IconSsd   *fyne.StaticResource
	IconFdd   *fyne.StaticResource
	IconStick *fyne.StaticResource

	IconController     *fyne.StaticResource
	IconMedia          *fyne.StaticResource
	IconSnapshot       *fyne.StaticResource
	IconEject          *fyne.StaticResource
	IconGuestAdditions *fyne.StaticResource

	IconExport   *fyne.StaticResource
	IconImport   *fyne.StaticResource
	IconExport_x *fyne.StaticResource
	IconImport_x *fyne.StaticResource

	IconGlobal    *fyne.StaticResource
	IconReadOnly  *fyne.StaticResource
	IconWriteable *fyne.StaticResource
	IconAutomount *fyne.StaticResource

	Settings         *Preferences
	ActiveItemServer string
	ActiveItemVm     string
	StartButton      *picbutton.PicButton
	PauseButton      *picbutton.PicButton
	SaveButton       *picbutton.PicButton
	ResetButton      *picbutton.PicButton
	OffButton        *picbutton.PicButton
	ShutdownButton   *picbutton.PicButton
	DiscardButton    *picbutton.PicButton
	ButtonLine       *fyne.Container
	ButtonLineBack   *canvas.Rectangle
	MasterPassword   string

	MenuServer *fyne.Menu

	SShServerDetails *widget.AccordionItem
	VmInfoDetails    *widget.AccordionItem
	VmNetworkDetails *widget.AccordionItem
	VmStorageDetails *widget.AccordionItem
	TasksDetails     *widget.AccordionItem

	VmServerTabs       *container.AppTabs
	VmInfoTabs         *container.AppTabs
	VmNetworkTab       *container.AppTabs
	VmStorageContainer *fyne.Container

	ServerSshTab      *ServerSshInfos
	ServerStatTab     *ServerStatInfos
	ServerVmTab       *VmServerInfos
	VmInfoTab         *InfoTab
	VmCpuRamTab       *CpuRamTab
	VmDisplayTab      *DisplayTab
	VmAudioTab        *AudioTab
	VmRdpTab          *RdpTab
	VmSystemTab       *SystemTab
	VmNetworkTabs     []*NetworkTab
	VmStorageContent  *StorageContent
	VmUsbTab          *UsbTab
	VmUsbAttachTab    *UsbAttachTab
	VmSnapshotTab     *SnapshotTab
	VmSharedFolderTab *SharedFolderTab
	TasksInfos        *TasksInfos
	DetailObjs        []DetailsInterface
}

var Gui = GUI{
	ToolbarActions: make(map[string]*widget.ToolbarAction, 10),
	MenuItems:      make(map[string]*fyne.MenuItem, 10),
	DetailObjs:     make([]DetailsInterface, 0, 25),
}

var Data = data.NewVmData()

type forcedVariant struct {
	fyne.Theme

	variant fyne.ThemeVariant
}

func (f *forcedVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}

func main() {
	//  go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	Gui.App = app.NewWithID("com.bytemystery.vboxssh")

	loadTranslations(content, "assets/lang")
	loadIcons()
	loadPreferences()

	Gui.FyneSettings = Gui.App.Settings()

	if _, ok := Gui.App.(desktop.App); ok {
		Gui.IsDesktop = true
	}
	Gui.MainWindow = Gui.App.NewWindow("VBoxSsh")
	Gui.MainWindow.SetIcon(Gui.Icon)

	hMenu := fyne.NewMenu(lang.X("menu.help", "Help"),
		fyne.NewMenuItem(lang.X("menu.help.help", "Help"), func() { doHelp() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(lang.X("menu.help.info", "Info"), showInfoDialog),
	)
	var m *fyne.MenuItem
	m = fyne.NewMenuItem(lang.X("menu.server.add", "Add"), addNewServer)
	m.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyD, Modifier: fyne.KeyModifierControl}
	Gui.MenuItems["menu.server.add"] = m

	m = fyne.NewMenuItem(lang.X("menu.server.remove", "Remove"), doDeleteServer)
	m.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierControl}
	Gui.MenuItems["menu.server.remove"] = m

	m = fyne.NewMenuItem(lang.X("menu.server.connect", "Connect"), doConnectServer)
	m.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl}
	Gui.MenuItems["menu.server.connect"] = m

	m = fyne.NewMenuItem(lang.X("menu.server.reconnect", "Reconnect"), doReconnectServer)
	m.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierControl}
	Gui.MenuItems["menu.server.reconnect"] = m

	m = fyne.NewMenuItem(lang.X("menu.server.disconnect", "Disconnect"), doDisconnectServer)
	Gui.MenuItems["menu.server.disconnect"] = m
	m.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl}

	Gui.MenuServer = fyne.NewMenu(lang.X("menu.server", "Server"),
		Gui.MenuItems["menu.server.add"],
		Gui.MenuItems["menu.server.remove"],
		fyne.NewMenuItemSeparator(),
		Gui.MenuItems["menu.server.connect"],
		Gui.MenuItems["menu.server.reconnect"],
		Gui.MenuItems["menu.server.disconnect"],
	)
	eMenu := fyne.NewMenu(lang.X("menu.edit", "Edit"),
		fyne.NewMenuItem(lang.X("menu.edit.appearance", "Appearance"), showAppearanceDialog))

	Gui.MenuItems["menu.machine.import"] = fyne.NewMenuItem(lang.X("menu.machine.import", "Import"), doImport)
	Gui.MenuItems["menu.machine.export"] = fyne.NewMenuItem(lang.X("menu.machine.export", "Export"), doExport)
	Gui.MenuItems["menu.machine.create"] = fyne.NewMenuItem(lang.X("menu.machine.create", "Create"), doCreateVm)
	Gui.MenuItems["menu.machine.clone"] = fyne.NewMenuItem(lang.X("menu.machine.clone", "Clone"), doCloneVm)
	Gui.MenuItems["menu.machine.delete"] = fyne.NewMenuItem(lang.X("menu.machine.delete", "Delete"), doDeleteVm)

	mMenu := fyne.NewMenu(lang.X("menu.machine", "Machine"),
		Gui.MenuItems["menu.machine.import"],
		Gui.MenuItems["menu.machine.export"],
		fyne.NewMenuItemSeparator(),
		Gui.MenuItems["menu.machine.clone"],
		fyne.NewMenuItemSeparator(),
		Gui.MenuItems["menu.machine.create"],
		Gui.MenuItems["menu.machine.delete"],
	)

	Gui.MainMenu = fyne.NewMainMenu(Gui.MenuServer, eMenu, mMenu, hMenu)
	Gui.MainWindow.SetMainMenu(Gui.MainMenu)

	Gui.StartusBar = colorlabel.NewColorLabel("Status", theme.ColorNameHyperlink, theme.ColorNameDisabled, 1.0)
	Gui.StartusBar.SetTruncate(true)

	Gui.Toolbar = widget.NewToolbar()

	if !Gui.IsDesktop {
		t := widget.NewToolbarAction(theme.ContentAddIcon(), addNewServer)
		Gui.ToolbarActions["addserver"] = t
		Gui.Toolbar.Append(t)
	}

	t := widget.NewToolbarAction(theme.ContentAddIcon(), addNewServer)
	Gui.ToolbarActions["addserver"] = t
	Gui.Toolbar.Append(t)

	t = widget.NewToolbarAction(theme.ContentRemoveIcon(), doDeleteServer)
	Gui.ToolbarActions["deleteserver"] = t
	Gui.Toolbar.Append(t)

	Gui.Toolbar.Append(widget.NewToolbarSeparator())

	t = widget.NewToolbarAction(theme.ConfirmIcon(), doConnectServer)
	Gui.ToolbarActions["connect"] = t
	Gui.Toolbar.Append(t)

	t = widget.NewToolbarAction(theme.ContentClearIcon(), doDisconnectServer)
	Gui.ToolbarActions["disconnect"] = t
	Gui.Toolbar.Append(t)

	t = widget.NewToolbarAction(theme.SearchReplaceIcon(), doReconnectServer)
	Gui.ToolbarActions["reconnect"] = t
	Gui.Toolbar.Append(t)

	Gui.Toolbar.Append(widget.NewToolbarSeparator())

	//	if Gui.IsDesktop {
	t = widget.NewToolbarAction(Gui.IconImport, doImport)
	Gui.ToolbarActions["import"] = t
	Gui.Toolbar.Append(t)
	t = widget.NewToolbarAction(Gui.IconExport, doExport)
	Gui.ToolbarActions["export"] = t
	Gui.Toolbar.Append(t)
	t = widget.NewToolbarAction(theme.DocumentIcon(), doCreateVm)
	Gui.ToolbarActions["create"] = t
	Gui.Toolbar.Append(t)
	t = widget.NewToolbarAction(theme.DeleteIcon(), doDeleteVm)
	Gui.ToolbarActions["delete"] = t
	Gui.Toolbar.Append(t)
	t = widget.NewToolbarAction(theme.ContentCopyIcon(), doCloneVm)
	Gui.ToolbarActions["clone"] = t
	Gui.Toolbar.Append(t)

	Gui.Toolbar.Append(widget.NewToolbarSeparator())
	//	}

	t = widget.NewToolbarAction(theme.LogoutIcon(), LogOut)
	Gui.ToolbarActions["logout"] = t
	Gui.Toolbar.Append(t)

	if Gui.IsDesktop {
		Gui.Toolbar.Append(widget.NewToolbarSpacer())
		t = widget.NewToolbarAction(theme.HelpIcon(), doHelp)
		Gui.ToolbarActions["help"] = t
		Gui.Toolbar.Append(t)
	}

	t = widget.NewToolbarAction(theme.InfoIcon(), showInfoDialog)
	Gui.ToolbarActions["info"] = t
	Gui.Toolbar.Append(t)

	Gui.Tree = widget.NewTree(treeGetChilds, treeIsBranche, createCanvasObject, treeUpdateItem)

	Gui.Tree.OnBranchOpened = treeBranchOpened
	Gui.Tree.OnSelected = treeOnSelected

	Gui.ServerSshTab = NewSshServerTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.ServerSshTab)

	Gui.ServerStatTab = NewServerStatTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.ServerStatTab)

	Gui.ServerVmTab = NewVmServerTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.ServerVmTab)

	Gui.VmInfoTab = NewInfoTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmInfoTab)

	Gui.VmCpuRamTab = NewCpuRamTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmCpuRamTab)

	Gui.VmDisplayTab = NewDisplayTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmDisplayTab)

	Gui.VmAudioTab = NewAudioTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmAudioTab)

	Gui.VmRdpTab = NewRdpTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmRdpTab)

	Gui.VmSystemTab = NewSystemTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmSystemTab)

	Gui.VmStorageContent = NewStorageContent()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmStorageContent)

	Gui.VmUsbTab = NewUsbTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmUsbTab)

	Gui.VmUsbAttachTab = NewUsbAttachTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmUsbAttachTab)

	Gui.VmSnapshotTab = NewSnapshotTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmSnapshotTab)

	Gui.VmSharedFolderTab = NewSharedFolderTab()
	Gui.DetailObjs = append(Gui.DetailObjs, Gui.VmSharedFolderTab)

	Gui.VmServerTabs = container.NewAppTabs(Gui.ServerSshTab.tabItem, Gui.ServerStatTab.tabItem, Gui.ServerVmTab.tabItem)

	Gui.SShServerDetails = widget.NewAccordionItem(lang.X("details.server", "Server"), Gui.VmServerTabs)

	Gui.VmInfoTabs = container.NewAppTabs(
		Gui.VmInfoTab.tabItem, Gui.VmSystemTab.tabItem, Gui.VmCpuRamTab.tabItem,
		Gui.VmDisplayTab.tabItem, Gui.VmRdpTab.tabItem, Gui.VmAudioTab.tabItem, Gui.VmStorageContent.tabItem,
		Gui.VmUsbTab.tabItem, Gui.VmUsbAttachTab.tabItem, Gui.VmSnapshotTab.tabItem, Gui.VmSharedFolderTab.tabItem)
	Gui.VmInfoDetails = widget.NewAccordionItem(lang.X("details.vm_info", "VM - General"), Gui.VmInfoTabs)

	for i := 0; i < NUMBER_OF_NICS; i++ {
		Gui.VmNetworkTabs = append(Gui.VmNetworkTabs, NewNetworkTab(i))
	}

	netTabList := make([]*container.TabItem, 0, NUMBER_OF_NICS)
	for _, item := range Gui.VmNetworkTabs {
		netTabList = append(netTabList, item.tabItem)
		Gui.DetailObjs = append(Gui.DetailObjs, item)
	}
	Gui.VmNetworkTab = container.NewAppTabs(netTabList...)
	Gui.VmNetworkDetails = widget.NewAccordionItem(lang.X("details.vm_network", "VM - Network"), Gui.VmNetworkTab)

	// Gui.VmStorageDetails = widget.NewAccordionItem(lang.X("details.vm_storage", "VM - Storage"), Gui.VmStorageContent.storageContent)
	Gui.TasksInfos = NewTasksInfos()
	Gui.TasksDetails = widget.NewAccordionItem(lang.X("details.tasks", "Tasks"), Gui.TasksInfos.content)

	Gui.Details = widget.NewAccordion(Gui.SShServerDetails, Gui.VmInfoDetails, Gui.VmNetworkDetails, Gui.TasksDetails)
	Gui.Details.MultiOpen = false
	Gui.DetailsScroll = container.NewVScroll(Gui.Details)
	fixScroll(Gui.DetailsScroll)

	padding := true
	/*
		if !Gui.IsDesktop {
			padding = false
		}
	*/

	Gui.StartButton = picbutton.NewPicButtonEx(Gui.PicStartU.StaticContent, Gui.PicStartD.StaticContent, nil, Gui.PicStartD.StaticContent, true, padding, 0, ButtonStart, nil)
	Gui.PauseButton = picbutton.NewPicButtonEx(Gui.PicPauseU.StaticContent, Gui.PicPauseD.StaticContent, nil, Gui.PicPauseD.StaticContent, true, padding, 0, ButtonPause, nil)
	Gui.SaveButton = picbutton.NewPicButtonEx(Gui.PicSaveU.StaticContent, Gui.PicSaveD.StaticContent, nil, Gui.PicSaveD.StaticContent, true, padding, 0, ButtonSave, nil)
	Gui.ResetButton = picbutton.NewPicButtonEx(Gui.PicResetU.StaticContent, Gui.PicResetD.StaticContent, nil, nil, false, padding, 0, ButtonReset, nil)
	Gui.OffButton = picbutton.NewPicButtonEx(Gui.PicOffU.StaticContent, Gui.PicOffD.StaticContent, nil, Gui.PicOffD.StaticContent, true, padding, 0, ButtonOff, nil)
	Gui.ShutdownButton = picbutton.NewPicButtonEx(Gui.PicShutdownU.StaticContent, Gui.PicShutdownD.StaticContent, nil, nil, false, padding, 0, ButtonShutdown, nil)
	Gui.DiscardButton = picbutton.NewPicButtonEx(Gui.PicDiscardU.StaticContent, Gui.PicDiscardD.StaticContent, nil, nil, false, padding, 0, ButtonDiscard, nil)

	Gui.ButtonLine = container.NewHBox(Gui.StartButton, Gui.SaveButton, Gui.ShutdownButton, Gui.PauseButton, Gui.OffButton, Gui.ResetButton, Gui.DiscardButton)
	Gui.ButtonLineBack = canvas.NewRectangle(color.NRGBA{R: 192, G: 192, B: 192, A: 255})
	vbox := container.NewBorder(container.NewStack(Gui.ButtonLineBack, Gui.ButtonLine), nil, nil, nil, Gui.DetailsScroll)

	Gui.Split = container.NewHSplit(Gui.Tree, container.NewHScroll(vbox))
	Gui.Split.Offset = 0.25

	Gui.OuterBorderLayout = container.NewBorder(Gui.Toolbar, Gui.StartusBar, nil, nil, Gui.Split)

	Gui.MainWindow.SetContent(Gui.OuterBorderLayout)
	Gui.MainWindow.Resize(fyne.NewSize(1000, 750))
	Gui.MainWindow.CenterOnScreen()

	showPasswordDialog(func(pass string) {
		pass, err := crypt.Encrypt(crypt.InternPassword, pass)
		if err != nil {
			CloseApp()
		}
		Gui.MasterPassword = pass
		pass = ""
		if !CheckMasterKey() {
			dia := dialog.NewError(errors.New(lang.X("msg.masterpassword_wrong", "Masterpassword is wrong !!")), Gui.MainWindow)
			dia.SetOnClosed(func() {
				CloseApp()
			})
			dia.Show()
		} else {
			LoadData()
			go treeUpdateTimerProc()
			UpdateButtons()
			if Gui.Settings.FirstStart {
				Gui.Settings.FirstStart = false
				Gui.Settings.Store()
			}
		}
	}, func() {
		CloseApp()
	}, Gui.Settings.FirstStart)

	fyne.CurrentApp().Settings().AddListener(func(settings fyne.Settings) {
		switch settings.ThemeVariant() {
		case theme.VariantDark:
			Gui.ButtonLineBack.FillColor = color.NRGBA{R: 92, G: 92, B: 92, A: 255}
		case theme.VariantLight:
			Gui.ButtonLineBack.FillColor = color.NRGBA{R: 192, G: 192, B: 192, A: 255}
		default:
			Gui.ButtonLineBack.FillColor = color.NRGBA{R: 192, G: 192, B: 192, A: 255}
		}
		Gui.ButtonLineBack.Refresh()
		loadIconsForTheme()
		Gui.Tree.Refresh()
		Gui.VmStorageContent.tree.Refresh()
		Gui.VmStorageContent.UpdateToolBarIcons()
		updateToolBarIcons()
		Gui.ServerVmTab.extPackList.Refresh()
		Gui.VmSharedFolderTab.list.Refresh()
	})

	if desk, ok := Gui.App.(desktop.App); ok {
		m := fyne.NewMenu("",
			fyne.NewMenuItem(lang.X("systemtray.show", "Show"), func() {
				Gui.MainWindow.Show()
			}))
		desk.SetSystemTrayMenu(m)
	}

	Gui.MainWindow.SetCloseIntercept(func() {
		Gui.MainWindow.Hide()
	})

	Gui.MainWindow.ShowAndRun()
}

func fixScroll(scroll *container.Scroll) {
	go func() {
		var oldSize fyne.Size
		for {
			time.Sleep(100 * time.Millisecond)
			fyne.Do(func() {
				c := scroll.Content
				if c != nil {
					ms := c.MinSize()
					if oldSize != ms {
						c.Resize(ms)
						scroll.Refresh()
						oldSize = ms
					}
				}
			})
		}
	}()
}
