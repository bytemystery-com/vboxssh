package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

func doCreateVm() {
	s, _ := getActiveServerAndVm()

	if s == nil {
		return
	}

	name := widget.NewEntry()
	name.SetPlaceHolder(lang.X("create.name.placeholder", "Name for the new VM"))
	item := widget.NewFormItem(lang.X("create.name", "Name"), name)

	dia := dialog.NewForm(lang.X("create.title", "Create new VM"),
		lang.X("create.create", "Create"),
		lang.X("create.cancel", "Cancel"), []*widget.FormItem{item}, func(ok bool) {
			err := s.CreateVm(&s.Client, name.Text)
			if err != nil {
				SetStatusText(fmt.Sprintf(lang.X("create.failed", "Creating VM width name '%s' failed"), name.Text), MsgError)
			} else {
				SetStatusText(fmt.Sprintf(lang.X("create.created", "VM width name '%s' was created"), name.Text), MsgInfo)
				go treeUpdateVmList(s.UUID)
			}
		}, Gui.MainWindow)
	var windowScale float32 = 0.25
	si := Gui.MainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
	dia.Show()
}

func doDeleteVm() {
	s, m := getActiveServerAndVm()

	if s == nil || m == nil {
		return
	}

	label := widget.NewLabel(m.Name)
	del := widget.NewCheck(lang.X("delete.all", "Delete all files"), nil)
	item1 := widget.NewFormItem("", del)
	item2 := widget.NewFormItem("", label)

	dia := dialog.NewForm(lang.X("delete.title", "Delete VM"),
		lang.X("delete.delete", "Delete"),
		lang.X("delete.cancel", "Cancel"), []*widget.FormItem{item2, item1}, func(ok bool) {
			if ok {
				dialog.ShowConfirm(lang.X("delete.confirm.title", "Confirm deleting VM"),
					fmt.Sprintf(lang.X("delete.confirm.msg", "Do you really want to delete the VM\n'%s'\nfrom server '%s' ?"), m.Name, s.Name),
					func(ok bool) {
						if ok {
							err := m.DeleteVm(&s.Client, del.Checked)
							if err != nil {
								SetStatusText(fmt.Sprintf(lang.X("delete.failed", "Deleting VM width name '%s' failed"), m.Name), MsgError)
							} else {
								SetStatusText(fmt.Sprintf(lang.X("delete.deleted", "VM width name '%s' was deleted"), m.Name), MsgInfo)
								go treeUpdateVmList(s.UUID)
							}
						}
					}, Gui.MainWindow)
			}
		}, Gui.MainWindow)

	var windowScale float32 = 0.25
	si := Gui.MainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
	dia.Show()
}
