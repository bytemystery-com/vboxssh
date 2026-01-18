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
	"fmt"
	"image/png"
	"net/url"
	"os"
	"runtime"
	"runtime/debug"

	"bytemystery-com/vboxssh/crypt"
	"bytemystery-com/vboxssh/util"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func showInfoDialog() {
	vgo := runtime.Version()[2:]
	vfyne := ""
	os := runtime.GOOS
	arch := runtime.GOARCH
	info, _ := debug.ReadBuildInfo()
	for _, dep := range info.Deps {
		if dep.Path == "fyne.io/fyne/v2" {
			vfyne = dep.Version[1:]
		}
	}
	s := fyne.CurrentApp().Settings()
	t := s.ThemeVariant()
	thema := ""
	col := s.PrimaryColor()
	b := s.BuildType()
	_ = b
	scale := s.Scale()
	switch t {
	case theme.VariantDark:
		thema = lang.X("info.thema_dark", "Dark")
	case theme.VariantLight:
		thema = lang.X("info.thema_light", "Light")
	default:
		thema = lang.X("info.thema_unknown", "Unknown")
	}

	build := ""
	switch b {
	case fyne.BuildStandard:
		build = lang.X("info.build_standard", "Standard")
	case fyne.BuildDebug:
		build = lang.X("info.build_debug", "Debug")
	case fyne.BuildRelease:
		build = lang.X("info.build_release", "Release")
	default:
		build = lang.X("info.build_unknown", "Unknown")
	}

	m := Gui.App.Metadata()
	v := fmt.Sprintf("%s (%d)", m.Version, m.Build)
	n := m.Name
	if n == "" {
		n = "VBoxSsh"
	}
	tsStr := ""
	ts := m.Custom["buildts"]
	if ts != "" {
		tsStr = "Build: " + ts + "\n"
	}
	wSize := Gui.MainWindow.Canvas().Size()

	msg := fmt.Sprintf(lang.X("info.msg", "\n%s\n\nVersion: %s  \n%sAuthor: Reiner Pröls\n\nGo version: %s\n\nFyne version: %s\nBuild: %s\nThema: %s\nPrimary color: %s\nScale: %.2f\nWindow size: %.0fx%.0f\n\nPlatform: %s\nArchitecture: %s"),
		n, v, tsStr, vgo, vfyne, build, thema, col, scale, wSize.Width, wSize.Height, os, arch)
	dialog.ShowInformation(lang.X("info.title", "Info"), msg, Gui.MainWindow)
}

func showPasswordDialog(fOk func(pass string), fCancel func(), withConfirm bool) {
	var dia *dialog.ConfirmDialog

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder(lang.X("masterpasswd.dialog.passwdplaceholder", "Master password"))
	passEntry.OnSubmitted = func(string) {
		dia.Confirm()
	}

	passEntryConfirm := widget.NewPasswordEntry()
	passEntryConfirm.SetPlaceHolder(lang.X("masterpasswd.dialog.confirm.passwdplaceholder", "Retype master password"))
	passEntryConfirm.OnSubmitted = func(string) {
		dia.Confirm()
	}

	validator := func(str string) error {
		if len(passEntry.Text) < 3 {
			return errors.New("Passwords is too short")
		}
		return nil
	}

	validatorConfirm := func(str string) error {
		if passEntryConfirm.Text != passEntry.Text {
			return errors.New("Passwords does not match")
		}
		return nil
	}

	if withConfirm {
		passEntry.Validator = validator
		passEntryConfirm.Validator = validatorConfirm
	}

	confirm := func(confirm bool) {
		if confirm {
			err := passEntry.Validate()
			if err != nil {
				dia.Show()
				return
			}
			if withConfirm {
				err := passEntryConfirm.Validate()
				if err != nil {
					dia.Show()
					return
				}
			}

			fOk(passEntry.Text)
		} else {
			fCancel()
		}
	}
	var c *fyne.Container
	if withConfirm {
		c = container.NewVBox(passEntry, widget.NewLabel(lang.X("masterpasswd.dialog.confirm", "Confirm password")), passEntryConfirm, util.NewVFiller(1.0))
	} else {
		c = container.NewVBox(passEntry, util.NewVFiller(1.0))
	}
	t := ""
	if withConfirm {
		t = lang.X("masterpasswd.dialog.title.new", "New master password")
	} else {
		t = lang.X("masterpasswd.dialog.title", "Master password")
	}

	dia = dialog.NewCustomConfirm(t, lang.X("ok", "Ok"), lang.X("cancel", "Cancel"),
		c, confirm, Gui.MainWindow)
	dia.Show()

	Gui.MainWindow.Canvas().Focus(passEntry)
	si := Gui.MainWindow.Canvas().Size()
	var windowScale float32 = .3
	dia.Resize(fyne.NewSize(si.Width*windowScale, dia.MinSize().Height))
}

func showAppearanceDialog() {
	appearance := settings.NewSettings().LoadAppearanceScreen(Gui.MainWindow)
	dialog.ShowCustom(lang.X("caption.fyne.appearance", "Fyne theme settings"), lang.X("ok", "Ok"), appearance, Gui.MainWindow)
}

func UpdateUI() {
	UpdateButtons()
	UpdateDetails()
	UpdateToolbarMenu()
}

func loadPreferences() {
	Gui.Settings = NewPreferences()
}

func loadIcon(path, name string) *fyne.StaticResource {
	data, err := content.ReadFile(path)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource(name, data)
}

func loadTranslations(fs embed.FS, dir string) {
	lang.AddTranslationsFS(fs, dir)
}

func VMStatusUpdateCallBack(uuid string) {
	fyne.Do(func() {
		Gui.Tree.Refresh()
		if Gui.ActiveItemVm == uuid {
			UpdateButtons()
			UpdateDetailsStatus()
		}
	})
}

func loadIcons() {
	Gui.Icon = loadIcon("assets/icons/icon.png", "icon")
	Gui.App.SetIcon(Gui.Icon)

	Gui.IconEmpty = loadIcon("assets/icons/empty.png", "icon_empty")

	Gui.PicStartU = loadIcon("assets/start_u.png", "run_u")
	Gui.PicStartD = loadIcon("assets/start_d.png", "run_d")
	Gui.PicPauseU = loadIcon("assets/pause_u.png", "pause_u")
	Gui.PicPauseD = loadIcon("assets/pause_d.png", "pause_d")
	Gui.PicSaveU = loadIcon("assets/save_u.png", "save_u")
	Gui.PicSaveD = loadIcon("assets/save_d.png", "save_d")
	Gui.PicOffU = loadIcon("assets/off_u.png", "off_u")
	Gui.PicOffD = loadIcon("assets/off_d.png", "off_d")
	Gui.PicResetU = loadIcon("assets/reset_u.png", "reset_u")
	Gui.PicResetD = loadIcon("assets/reset_d.png", "reset_d")
	Gui.PicShutdownU = loadIcon("assets/shutdown_u.png", "shutdown_u")
	Gui.PicShutdownD = loadIcon("assets/shutdown_d.png", "shutdown_d")
	Gui.PicDiscardU = loadIcon("assets/discard_u.png", "discard_u")
	Gui.PicDiscardD = loadIcon("assets/discard_d.png", "discard_d")

	loadIconsForTheme()
}

func loadIconsForTheme() {
	dir := ""
	switch fyne.CurrentApp().Settings().ThemeVariant() {
	case theme.VariantDark:
		dir = "dark"
	case theme.VariantLight:
		dir = "light"
	default:
		dir = "light"
	}
	Gui.IconRun = loadIcon("assets/icons/"+dir+"/run.png", "icon_run")
	Gui.IconStop = loadIcon("assets/icons/"+dir+"/stop.png", "icon_stop")
	Gui.IconPause = loadIcon("assets/icons/"+dir+"/pause.png", "icon_pause")
	Gui.IconAbort = loadIcon("assets/icons/"+dir+"/abort.png", "icon_abort")
	Gui.IconSave = loadIcon("assets/icons/"+dir+"/save.png", "icon_save")
	Gui.IconMeditation = loadIcon("assets/icons/"+dir+"/meditation.png", "icon_meditation")
	Gui.IconUnknown = loadIcon("assets/icons/"+dir+"/unknown.png", "icon_unknown")
	Gui.IconOk = loadIcon("assets/icons/"+dir+"/ok.png", "icon_ok")
	Gui.IconError = loadIcon("assets/icons/"+dir+"/error.png", "icon_error")

	Gui.IconFloppy = loadIcon("assets/icons/"+dir+"/floppy.png", "icon_floppy")
	Gui.IconIde = loadIcon("assets/icons/"+dir+"/ide.png", "icon_ide")
	Gui.IconPcie = loadIcon("assets/icons/"+dir+"/pcie.png", "icon_pcie")
	Gui.IconSas = loadIcon("assets/icons/"+dir+"/sas.png", "icon_sas")
	Gui.IconSata = loadIcon("assets/icons/"+dir+"/sata.png", "icon_sata")
	Gui.IconScsi = loadIcon("assets/icons/"+dir+"/scsi.png", "icon_scsi")
	Gui.IconUsb = loadIcon("assets/icons/"+dir+"/usb.png", "icon_usb")
	Gui.IconVirt = loadIcon("assets/icons/"+dir+"/virt.png", "icon_virt")

	Gui.IconCd = loadIcon("assets/icons/"+dir+"/cd.png", "icon_cd")
	Gui.IconHdd = loadIcon("assets/icons/"+dir+"/hdd.png", "icon_hdd")
	Gui.IconSsd = loadIcon("assets/icons/"+dir+"/ssd.png", "icon_ssd")
	Gui.IconFdd = loadIcon("assets/icons/"+dir+"/fdd.png", "icon_fdd")
	Gui.IconStick = loadIcon("assets/icons/"+dir+"/stick.png", "icon_stick")

	Gui.IconController = loadIcon("assets/icons/"+dir+"/controller.png", "icon_controller")
	Gui.IconMedia = loadIcon("assets/icons/"+dir+"/media.png", "icon_media")
	Gui.IconSnapshot = loadIcon("assets/icons/"+dir+"/snapshot.png", "icon_snapshot")
	Gui.IconEject = loadIcon("assets/icons/"+dir+"/eject.png", "icon_eject")

	Gui.IconGuestAdditions = loadIcon("assets/icons/"+dir+"/guestadditions.png", "icon_guestadditions")

	Gui.IconExport = loadIcon("assets/icons/"+dir+"/export.png", "icon_export")
	Gui.IconImport = loadIcon("assets/icons/"+dir+"/import.png", "icon_import")

	Gui.IconExport_x = loadIcon("assets/icons/"+dir+"/export_x.png", "icon_export_x")
	Gui.IconImport_x = loadIcon("assets/icons/"+dir+"/import_x.png", "icon_import_x")
}

func CloseApp() {
	if Gui.IsDesktop {
		Gui.MainWindow.Close()
		Gui.App.Quit()
	} else {
		LogOut()
	}
}

func CheckMasterKey() bool {
	pass, err := crypt.Decrypt(crypt.InternPassword, Gui.MasterPassword)
	if err != nil {
		return false
	}
	if Gui.Settings.MasterKeyTest == PREF_MASTERKEY_TEST_VALUE {
		x, err := crypt.Encrypt(pass, PREF_MASTERKEY_TEST_VALUE)
		if err != nil {
			return false
		}
		Gui.Settings.MasterKeyTest = x
		Gui.Settings.Store()
		return true
	}
	t, err := crypt.Decrypt(pass, Gui.Settings.MasterKeyTest)
	if err != nil {
		return false
	}
	if t == PREF_MASTERKEY_TEST_VALUE {
		return true
	}
	return false
}

func LoadData() {
	servers, _ := loadServers(Gui.MasterPassword)
	Data.LoadData(servers)
	Gui.Tree.Refresh()
	vms := Data.GetServers(true)
	if vms != nil {
		Gui.Tree.Select(vms[0].UUID)
	}
	// saveServers(Data.ServerMap, Gui.MasterPassword)
}

func LogOut() {
	showPasswordDialog(func(pass string) {
		pass, err := crypt.Encrypt(crypt.InternPassword, pass)
		if err != nil {
			return
		}
		Gui.MasterPassword = pass
		pass = ""
		if !CheckMasterKey() {
			dia := dialog.NewError(errors.New(lang.X("msg.masterpassword_wrong", "Masterpassword is wrong !!")), Gui.MainWindow)
			dia.Show()
			dia.SetOnClosed(func() {
				LogOut()
			})
		} else {
		}
	}, func() {
		LogOut()
	}, false)
}

func getActiveServerAndVm() (*vm.VmServer, *vm.VMachine) {
	if Gui.ActiveItemServer == "" || Gui.ActiveItemServer == SERVER_ADD_NEW_UUID {
		return nil, nil
	}
	Data.Lock.RLock()
	defer Data.Lock.RUnlock()
	s := Data.GetServer(Gui.ActiveItemServer, false)
	if s == nil {
		return nil, nil
	}
	if Gui.ActiveItemVm != "" {
		v := Data.GetVm(Gui.ActiveItemServer, Gui.ActiveItemVm, false)
		return s, v
	}
	return s, nil
}

func SendNotification(title, msg string) {
	fyne.Do(func() {
		n := fyne.NewNotification(title, msg)
		Gui.App.SendNotification(n)
	})
}

func updateToolBarIcons() {
	t, ok := Gui.ToolbarActions["export"]
	if ok {
		t.SetIcon(Gui.IconExport)
	}
	t, ok = Gui.ToolbarActions["import"]
	if ok {
		t.SetIcon(Gui.IconImport)
	}
	Gui.Toolbar.Refresh()
}

func doHelp() {
	u := url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "/bytemystery-com/vboxssh",
	}
	Gui.App.OpenURL(&u)
}

func MakeScreenShot() {
	img := Gui.MainWindow.Canvas().Capture()
	f, err := os.Create("screenshot.png")
	if err != nil {
		return
	}
	defer f.Close()

	png.Encode(f, img)
}
