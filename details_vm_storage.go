package main

import (
	"fmt"
	"image/color"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"bytemystery-com/vboxssh/util"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
)

type StorageMediumStateType int

const (
	StorageMediumState_new StorageMediumStateType = iota
	StorageMediumState_changed
	StorageMediumState_unchanged
	StorageMediumState_removed
	StorageMediumState_invalid
)

type StorageMedium struct {
	uuid          string
	file          string
	port          int
	device        int
	nonrotational bool
	hotpluggable  bool
	discard       bool
	isLive        bool

	// will be used only by Apply
	state StorageMediumStateType
}

func (m *StorageMedium) getStorageType(c *StorageController) vm.StorageType {
	var t vm.StorageType
	if c.busType == vm.StorageBus_floppy {
		t = vm.Storage_fdd
	} else if m.isImage() {
		t = vm.Storage_dvddrive
	} else {
		t = vm.Storage_hdd
	}
	return t
}

func (m *StorageMedium) getIsLive(c *StorageController) *bool {
	if c.busType != vm.StorageBus_floppy && m.isImage() {
		return &m.isLive
	}
	return nil
}

func (m *StorageMedium) getIsSsd() *bool {
	if !m.isImage() {
		return &m.nonrotational
	}
	return nil
}

func (m *StorageMedium) getUuidOrFile() string {
	if m.uuid == "" || m.uuid[0] == 'X' {
		return m.file
	} else {
		return m.uuid
	}
}

func (m1 *StorageMedium) isEqual(m2 *StorageMedium) bool {
	return !(m1.uuid != m2.uuid || m1.file != m2.file || m1.port != m2.port || m1.device != m2.device ||
		m1.nonrotational != m2.nonrotational || m1.hotpluggable != m2.hotpluggable ||
		m1.discard != m2.discard || m1.isLive != m2.isLive)
}

func (m *StorageMedium) isImage() bool {
	ext := strings.ToLower(filepath.Ext(m.file))
	if ext == ".iso" || ext == ".img" {
		return true
	}
	return false
}

type StorageControllerStateType int

const (
	StorageControllerState_new StorageControllerStateType = iota
	StorageControllerState_changed
	StorageControllerState_unchanged
	StorageControllerState_removed
	StorageControllerState_invalid
)

type StorageController struct {
	uuid           string
	name           string
	bootable       bool
	controllerType vm.StorageChipsetType
	busType        vm.StorageBusType
	mediums        []*StorageMedium

	// will be used only by Apply
	state   StorageControllerStateType
	oldItem *StorageController
}

func (s1 *StorageController) isEqualDetails(s2 *StorageController) bool {
	return !(s1.name != s2.name || s1.bootable != s2.bootable)
}

type StorageOldValues struct {
	storageControllers []*StorageController
}

type StorageContent struct {
	tree           *widget.Tree
	formMedium     *fyne.Container
	formController *fyne.Container
	formEmpty      *fyne.Container
	border         *fyne.Layout
	storageContent *container.Split
	tabItem        *container.TabItem
	apply          *widget.Button

	ssd    *widget.Check
	isLive *widget.Check
	device *widget.Entry
	port   *widget.Entry

	name     *widget.Entry
	bootable *widget.Check

	toolBar                      *widget.Toolbar
	toolBarItemAddCtrl           *widget.ToolbarAction
	toolBarItemRemoveCtrl        *widget.ToolbarAction
	toolBarItemAddMedium         *widget.ToolbarAction
	toolBarItemRemoveMedium      *widget.ToolbarAction
	toolBarItemAddGuestAdditions *widget.ToolbarAction
	toolBarItemEjectMedium       *widget.ToolbarAction

	storageControllers []*StorageController

	oldValues StorageOldValues

	lastSelectedCtrl   string
	lastSelectedMedium string

	chipsetMapStringToType     map[string]vm.StorageChipsetType
	busMapcontrollerTypeToType map[vm.StorageChipsetType]vm.StorageBusType

	snapshotRegEx1 *regexp.Regexp
	snapshotRegEx2 *regexp.Regexp
}

func NewStorageContent() *StorageContent {
	st := StorageContent{
		chipsetMapStringToType: map[string]vm.StorageChipsetType{
			"BusLogic": vm.StorageChipset_BusLogic,
			"I82078":   vm.StorageChipset_I82078, "ICH6": vm.StorageChipset_ICH6,
			"IntelAhci": vm.StorageChipset_IntelAHCI, "LsiLogic": vm.StorageChipset_LSILogic,
			"LsiLogicSas": vm.StorageChipset_LSILogicSAS, "NVMe": vm.StorageChipset_NVMe,
			"PIIX3": vm.StorageChipset_PIIX3, "PIIX4": vm.StorageChipset_PIIX4,
			"USB": vm.StorageChipset_USB, "VirtioSCSI": vm.StorageChipset_VirtIO,
		},
		busMapcontrollerTypeToType: map[vm.StorageChipsetType]vm.StorageBusType{
			vm.StorageChipset_BusLogic: vm.StorageBus_scsi,
			vm.StorageChipset_I82078:   vm.StorageBus_floppy, vm.StorageChipset_ICH6: vm.StorageBus_ide,
			vm.StorageChipset_IntelAHCI: vm.StorageBus_sata, vm.StorageChipset_LSILogic: vm.StorageBus_scsi,
			vm.StorageChipset_LSILogicSAS: vm.StorageBus_sas, vm.StorageChipset_NVMe: vm.StorageBus_pcie,
			vm.StorageChipset_PIIX3: vm.StorageBus_ide, vm.StorageChipset_PIIX4: vm.StorageBus_ide,
			vm.StorageChipset_USB: vm.StorageBus_usb, vm.StorageChipset_VirtIO: vm.StorageBus_virtio,
		},
		snapshotRegEx1: regexp.MustCompile(`{[0-9a-fA-F-]+}\..*$`),
		snapshotRegEx2: regexp.MustCompile(`.*/(.+)/.*/({[0-9a-fA-F-]+}\..*)$`),
	}

	st.apply = widget.NewButton(lang.X("details.vm_storage", "Apply"), func() {
		st.Apply()
	})
	st.apply.Importance = widget.HighImportance

	grid1 := container.New(layout.NewFormLayout())

	formWidth := util.GetFormWidth() / 2

	st.ssd = widget.NewCheck(lang.X("details.vm_storage.ssd", "Solid state drive"), st.onSssdChanged)
	st.isLive = widget.NewCheck(lang.X("details.vm_storage.islive", "Live CD/DVD"), st.onIsLiveChanged)
	st.device = widget.NewEntry()
	st.device.OnChanged = util.GetNumberFilter(st.device, st.onDeviceChanged)
	st.port = widget.NewEntry()
	st.port.OnChanged = util.GetNumberFilter(st.port, st.onPortChanged)
	grid2 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_storage.port", "Port")), st.port,
		widget.NewLabel(lang.X("details.vm_storage.device", "Device")), st.device,
		st.ssd, st.isLive,
	)
	gridWrap1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap2 := container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)

	gridWrap := container.NewVBox(util.NewVFiller(0.5), gridWrap1, gridWrap2)

	st.formMedium = container.NewVBox(container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), st.apply, util.NewFiller(32, 0)))

	grid1 = container.New(layout.NewFormLayout())

	st.name = widget.NewEntry()
	st.name.OnChanged = st.setControllerName

	st.bootable = widget.NewCheck(lang.X("details.vm_storage.bootable", "Bootable"), st.setControllerBootable)

	grid2 = container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_storage.name", "Name")), st.name,
		st.bootable, util.NewFiller(0, 0),
	)
	gridWrap1 = container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap2 = container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)

	gridWrap = container.NewVBox(util.NewVFiller(0.5), gridWrap1, gridWrap2)

	st.formController = container.NewVBox(container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), st.apply, util.NewFiller(32, 0)))

	// Empty
	grid1 = container.New(layout.NewFormLayout())

	grid2 = container.New(layout.NewFormLayout())
	gridWrap1 = container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap2 = container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)

	gridWrap = container.NewVBox(util.NewVFiller(0.5), gridWrap1, gridWrap2)

	st.formEmpty = container.NewVBox(container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), st.apply, util.NewFiller(32, 0)))

	st.tree = widget.NewTree(st.treeGetChilds, st.treeIsBranche, st.createCanvasObject, st.treeUpdateItem)
	st.tree.OnSelected = st.treeOnSelected

	st.formController.Hide()
	st.formMedium.Hide()
	st.formEmpty.Show()

	st.toolBarItemAddCtrl = widget.NewToolbarAction(Gui.IconController, func() { st.onNewController() })
	st.toolBarItemRemoveCtrl = widget.NewToolbarAction(theme.ContentRemoveIcon(), func() { st.onRemoveController() })

	st.toolBarItemAddMedium = widget.NewToolbarAction(Gui.IconMedia, func() { st.onNewMedia() })
	st.toolBarItemRemoveMedium = widget.NewToolbarAction(theme.ContentRemoveIcon(), func() { st.onRemoveMedia() })
	st.toolBarItemAddGuestAdditions = widget.NewToolbarAction(Gui.IconGuestAdditions, func() { st.onAddGuestAdditions() })
	st.toolBarItemEjectMedium = widget.NewToolbarAction(Gui.IconEject, func() { st.onEjectMedia() })

	st.toolBar = widget.NewToolbar(st.toolBarItemAddCtrl, st.toolBarItemRemoveCtrl, widget.NewToolbarSeparator(),
		st.toolBarItemAddMedium, st.toolBarItemRemoveMedium, widget.NewToolbarSeparator(), st.toolBarItemEjectMedium)

	st.storageContent = container.NewHSplit(container.NewBorder(st.toolBar, nil, nil, nil, st.tree), container.NewStack(st.formEmpty, st.formController, st.formMedium))
	st.storageContent.SetOffset(0.6)

	st.tabItem = container.NewTabItem(lang.X("details.vm_info.tab.storage", "Storage"), st.storageContent)
	return &st
}

func (st *StorageContent) buildText(s StorageMedium) (string, bool) {
	name := filepath.Base(s.file)
	if !st.snapshotRegEx1.MatchString(name) {
		return name, false
	} else {
		items := st.snapshotRegEx2.FindStringSubmatch(s.file)
		if len(items) == st.snapshotRegEx2.NumSubexp()+1 {
			return items[1] + "/…/" + items[2], true
		}
	}
	return s.file, false
}

func (st *StorageContent) isBusTypeAlreadyPresent(bus vm.StorageBusType) bool {
	for _, item := range st.storageControllers {
		if item.busType == bus {
			return true
		}
	}
	return false
}

func (st *StorageContent) onRemoveController() {
	c, m := st.getActiveControllerAndMedium()
	if c != nil && m == nil {
		for index, item := range st.storageControllers {
			if item == c {
				st.storageControllers = append(st.storageControllers[:index], st.storageControllers[index+1:]...)
				st.lastSelectedCtrl = ""
				st.lastSelectedMedium = ""
				break
			}
		}
		st.tree.Refresh()
		st.treeOnSelected("")
	}
}

func (st *StorageContent) getMaxNumberOfMedias(bus vm.StorageBusType) int {
	switch bus {
	case vm.StorageBus_floppy:
		return 1
	case vm.StorageBus_ide:
		return 4
	case vm.StorageBus_sata:
		return 30
	case vm.StorageBus_sas:
		return 255
	case vm.StorageBus_scsi:
		return 16
	case vm.StorageBus_usb:
		return 8
	case vm.StorageBus_pcie:
		return 255
	case vm.StorageBus_virtio:
		return 256

	}
	return 0
}

func (st *StorageContent) onRemoveMedia() {
	c, m := st.getActiveControllerAndMedium()
	for index, item := range c.mediums {
		if item == m {
			c.mediums = append(c.mediums[:index], c.mediums[index+1:]...)
			st.lastSelectedMedium = ""
			st.tree.Refresh()
			st.tree.Select(st.buildStorageMediumID(c.uuid, ""))
			break
		}
	}
}

func (st *StorageContent) onEjectMedia() {
	c, m := st.getActiveControllerAndMedium()
	s, v := getActiveServerAndVm()
	if m == nil || !m.isImage() || v == nil || s == nil {
		return
	}
	ResetStatus()
	storageType := vm.Storage_dvddrive
	if c.busType == vm.StorageBus_floppy {
		storageType = vm.Storage_fdd
	}
	err := v.EjectMedia(&s.Client, c.name, storageType, m.port, m.device, vm.MediaSpecial_emptydrive, VMStatusUpdateCallBack)
	if err != nil {
		SetStatusText(fmt.Sprintf(lang.X("details.vm_storge.eject.error", "Eject failed for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
	} else {
		for index, item := range c.mediums {
			if item == m {
				c.mediums = append(c.mediums[:index], c.mediums[index+1:]...)
				st.tree.Refresh()
				st.tree.Select(st.buildStorageMediumID(c.uuid, ""))
				break
			}
		}
	}
}

func (st *StorageContent) findFreePortDevice(c *StorageController, dvd bool) (int, int) {
	port := -1
	device := -1
	if c.busType == vm.StorageBus_ide {
		port, device = st.findFreeIDEPortAndDevice(c, dvd)
	} else {
		device = 0
		port = st.findFreePortNonIDE(c)
	}
	return port, device
}

func (st *StorageContent) findFreePortNonIDE(c *StorageController) int {
	max := st.getMaxNumberOfMedias(c.busType)
	for index := range max {
		flag := false
		for _, item := range c.mediums {
			if item.port == index {
				flag = true
				break
			}
		}
		if !flag {
			return index
		}
	}
	return -1
}

func (st *StorageContent) findFreeIDEPortAndDevice(c *StorageController, dvd bool) (int, int) {
	type PortDevice struct {
		p int
		d int
	}
	var l []PortDevice
	if dvd {
		l = []PortDevice{{1, 0}, {1, 1}, {0, 1}, {0, 0}}
	} else {
		l = []PortDevice{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
	}
	for _, item := range c.mediums {
		for index, item2 := range l {
			if item2.p == item.port && item2.d == item.device {
				l = slices.Delete(l, index, index+1)
				break
			}
		}
	}
	if len(l) > 0 {
		return l[0].p, l[0].d
	}
	return -1, -1
}

func (st *StorageContent) adNewMedium(s *vm.VmServer, v *vm.VMachine, c *StorageController, dvd bool) {
	m := NewMediaHelper(Gui.MainWindow, 0.85, 0.75)
	if dvd {
		m.SelectFloppyOrDvdImage(s, false, func(m vm.MediaInfo) {
			port, device := st.findFreePortDevice(c, true)
			if port >= 0 && device >= 0 {
				if m.UUID == "" {
					m.UUID = "X" + uuid.NewString()
				}
				c.mediums = append(c.mediums, &StorageMedium{
					uuid:   m.UUID,
					file:   m.Location,
					port:   port,
					device: device,
				})
				st.tree.Refresh()
				st.tree.OpenBranch(c.uuid)
				st.tree.Select(st.buildStorageMediumID(c.uuid, m.UUID))
			}
		})
	} else {
		m.SelectHddImage(s, v, func(m *vm.HddInfo) {
			port, device := st.findFreePortDevice(c, false)
			if port >= 0 && device >= 0 {
				c.mediums = append(c.mediums, &StorageMedium{
					uuid:   m.UUID,
					file:   m.Location,
					port:   port,
					device: device,
				})
				st.tree.Refresh()
				st.tree.OpenBranch(c.uuid)
				st.tree.Select(st.buildStorageMediumID(c.uuid, m.UUID))
			}
		})
	}
}

func (st *StorageContent) onAddGuestAdditions() {
	c, _ := st.getActiveControllerAndMedium()
	s, v := getActiveServerAndVm()
	if c != nil && s != nil {
		if len(c.mediums) < st.getMaxNumberOfMedias(c.busType) {
			if c.busType != vm.StorageBus_floppy {
				port, device := st.findFreePortDevice(c, true)
				err := v.AttachGuestAdditions(&s.Client, c.name, port, device, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_storage.attachguestadditions.error", "Attach guest additions to storage controller '%s' for VM '%s' failed with: %s"), c.name, v.Name, err.Error()), MsgError)
				}
				st.tree.Refresh()
			}
		}
	}
}

func (st *StorageContent) onNewMedia() {
	c, _ := st.getActiveControllerAndMedium()
	s, v := getActiveServerAndVm()
	if c != nil && s != nil {
		if len(c.mediums) < st.getMaxNumberOfMedias(c.busType) {
			m := NewMediaHelper(Gui.MainWindow, 0.85, 0.75)
			switch c.busType {
			case vm.StorageBus_floppy:
				m.SelectFloppyOrDvdImage(s, true, func(m vm.MediaInfo) {
					if m.UUID == "" {
						m.UUID = "X" + uuid.NewString()
					}
					c.mediums = append(c.mediums, &StorageMedium{
						uuid:   m.UUID,
						file:   m.Location,
						port:   0,
						device: 0,
					})
					st.tree.Refresh()
				})
			default:
				state, err := v.GetState()
				if err != nil {
					return
				}
				if state == vm.RunState_aborted || state == vm.RunState_off {
					m.ShowDvdHddOptionDialog(func(dvd bool) {
						st.adNewMedium(s, v, c, dvd)
					})
				} else {
					st.adNewMedium(s, v, c, true)
				}
			}
		}
	}
}

func (st *StorageContent) onNewController() {
	list := []vm.StorageChipsetType{}
	for _, val := range st.chipsetMapStringToType {
		bus, ok := st.busMapcontrollerTypeToType[val]
		if ok {
			if (bus == vm.StorageBus_floppy || bus == vm.StorageBus_ide || bus == vm.StorageBus_usb) &&
				st.isBusTypeAlreadyPresent(bus) {
				continue
			}
			list = append(list, val)
		}
	}
	type ControllerEntry struct {
		name           string
		controllerType vm.StorageChipsetType
	}
	selectList := []ControllerEntry{}
	for _, item := range list {
		str := ""
		switch item {
		case vm.StorageChipset_I82078:
			str = lang.X("details.vm_storage.floppy", "Floppy")
		case vm.StorageChipset_ICH6:
			str = lang.X("details.vm_storage.ich6", "IDE: ICH6")
		case vm.StorageChipset_PIIX3:
			str = lang.X("details.vm_storage.piix3", "IDE: PIIX3")
		case vm.StorageChipset_PIIX4:
			str = lang.X("details.vm_storage.piix4", "IDE: PIIX4")
		case vm.StorageChipset_LSILogicSAS:
			str = lang.X("details.vm_storage.sas", "SAS: LSILogic")
		case vm.StorageChipset_IntelAHCI:
			str = lang.X("details.vm_storage.ahci", "SATA: Intel AHCI")
		case vm.StorageChipset_BusLogic:
			str = lang.X("details.vm_storage.buslogic", "SCSI: BusLogic")
		case vm.StorageChipset_LSILogic:
			str = lang.X("details.vm_storage.lsi", "SCSI: LSILogic")
		case vm.StorageChipset_NVMe:
			str = lang.X("details.vm_storage.nvme", "NVMe")
		case vm.StorageChipset_USB:
			str = lang.X("details.vm_storage.usb", "USB")
		case vm.StorageChipset_VirtIO:
			str = lang.X("details.vm_storage.virtio", "VirtIO")
		}
		if str != "" {
			selectList = append(selectList, ControllerEntry{
				name:           str,
				controllerType: item,
			})
		}
	}

	slices.SortFunc(selectList, func(a, b ControllerEntry) int {
		an := strings.ToLower(a.name)
		bn := strings.ToLower(b.name)

		if an == bn {
			if a.name < b.name {
				return -1
			}
			if a.name > b.name {
				return 1
			}
			return 0
		}
		if an < bn {
			return -1
		}
		return 1
	})
	l := make([]string, 0, len(selectList))
	for _, item := range selectList {
		l = append(l, item.name)
	}
	sel := widget.NewSelect(l, nil)
	var dia *dialog.ConfirmDialog
	dia = dialog.NewCustomConfirm(lang.X("details.vm_storage.addctrl.title", "Add new controller"),
		lang.X("details.vm_storage.addctrl.add", "Add"),
		lang.X("details.vm_storage.addctrl.cancel", "Cancel"),
		container.NewVBox(sel, util.NewVFiller(1.0)),
		func(ok bool) {
			if ok {
				index := sel.SelectedIndex()
				if index >= 0 {
					item := selectList[index]
					name := strings.Split(item.name, ":")
					st.addController(item.controllerType, strings.TrimSpace(name[len(name)-1]))
				}
			}
			// dia.Hide()
		}, Gui.MainWindow)
	dia.Show()
	Gui.MainWindow.Canvas().Focus(sel)
}

func (st *StorageContent) addController(chipset vm.StorageChipsetType, name string) {
	bus, ok := st.busMapcontrollerTypeToType[chipset]
	if ok {
		u := uuid.NewString()
		st.storageControllers = append(st.storageControllers, &StorageController{
			uuid:           u,
			name:           name,
			controllerType: chipset,
			busType:        bus,
			mediums:        make([]*StorageMedium, 0, 2),
		})
		st.tree.Refresh()
		st.tree.Select(u)
	}
}

func (st *StorageContent) setControllerName(name string) {
	c, m := st.getActiveControllerAndMedium()
	if c != nil && m == nil {
		c.name = name
		st.tree.Refresh()
	}
}

func (st *StorageContent) setControllerBootable(checked bool) {
	c, m := st.getActiveControllerAndMedium()
	if c != nil && m == nil {
		c.bootable = checked
		st.tree.Refresh()
	}
}

func (st *StorageContent) onDeviceChanged(s string) {
	_, m := st.getActiveControllerAndMedium()
	if m == nil {
		return
	}
	val, err := strconv.Atoi(s)
	if err == nil {
		return
	}
	m.device = val
	st.tree.Refresh()
}

func (st *StorageContent) onPortChanged(s string) {
	_, m := st.getActiveControllerAndMedium()
	if m == nil {
		return
	}
	val, err := strconv.Atoi(s)
	if err == nil {
		return
	}
	m.port = val
	st.tree.Refresh()
}

func (st *StorageContent) onSssdChanged(checked bool) {
	_, m := st.getActiveControllerAndMedium()
	if m == nil {
		return
	}
	m.nonrotational = checked
	st.tree.Refresh()
}

func (st *StorageContent) onIsLiveChanged(checked bool) {
	_, m := st.getActiveControllerAndMedium()
	if m == nil {
		return
	}
	m.isLive = checked
	st.tree.Refresh()
}

func (st *StorageContent) buildStorageMediumID(cUuid, mUuid string) string {
	if mUuid == "" {
		return cUuid
	} else {
		return cUuid + "/" + mUuid
	}
}

func (st *StorageContent) getStorageAndMediumUUIDFromID(id widget.TreeNodeID) (string, string) {
	s := string(id)
	if s != "" {
		ids := strings.Split(s, "/")
		if len(ids) == 2 {
			return ids[0], ids[1]
		}
		return s, ""
	}
	return "", ""
}

func (st *StorageContent) getControllerAndMediumFromUuid(cUuid, mUuid string) (*StorageController, *StorageMedium) {
	for _, itemCtl := range st.storageControllers {
		if itemCtl.uuid == cUuid {
			if mUuid == "" {
				return itemCtl, nil
			}
			for _, itemMedium := range itemCtl.mediums {
				if itemMedium.uuid == mUuid {
					return itemCtl, itemMedium
				}
			}
			return itemCtl, nil
		}
	}
	return nil, nil
}

func (st *StorageContent) getActiveControllerAndMedium() (*StorageController, *StorageMedium) {
	return st.getControllerAndMediumFromUuid(st.lastSelectedCtrl, st.lastSelectedMedium)
}

// return childs
func (st *StorageContent) treeGetChilds(id widget.TreeNodeID) []widget.TreeNodeID {
	if id == "" {
		list := make([]widget.TreeNodeID, 0, len(st.storageControllers))
		for _, item := range st.storageControllers {
			list = append(list, st.buildStorageMediumID(item.uuid, ""))
		}
		return list
	} else {
		c, m := st.getStorageAndMediumUUIDFromID(id)
		if m != "" {
			return nil
		}
		if c != "" {
			ctrl, _ := st.getControllerAndMediumFromUuid(c, "")
			if ctrl == nil {
				return nil
			}
			list := make([]widget.TreeNodeID, 0, len(ctrl.mediums))
			for _, item := range ctrl.mediums {
				list = append(list, st.buildStorageMediumID(c, item.uuid))
			}
			return list
		}
	}
	return nil
}

// return is Branch
func (st *StorageContent) treeIsBranche(id widget.TreeNodeID) bool {
	if id == "" {
		return true
	}
	c, m := st.getStorageAndMediumUUIDFromID(id)
	if m != "" {
		return false
	}
	if c != "" {
		return true
	}
	return false
}

// create canvasObject
func (st *StorageContent) createCanvasObject(branche bool) fyne.CanvasObject {
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
func (st *StorageContent) treeUpdateItem(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
	cont, ok := o.(*fyne.Container)
	if !ok {
		return
	}
	text, ok := cont.Objects[0].(*canvas.Text)
	if !ok {
		return
	}

	icon, ok := cont.Objects[2].(*canvas.Image)
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

	c, m := st.getStorageAndMediumUUIDFromID(id)
	if m == "" && c != "" {
		ctrl, _ := st.getControllerAndMediumFromUuid(c, "")
		if ctrl == nil {
			return
		}
		text.Text = ctrl.name

		switch ctrl.busType {
		case vm.StorageBus_floppy:
			icon.Resource = Gui.IconFloppy

		case vm.StorageBus_ide:
			icon.Resource = Gui.IconIde

		case vm.StorageBus_pcie:
			icon.Resource = Gui.IconPcie

		case vm.StorageBus_sas:
			icon.Resource = Gui.IconSas

		case vm.StorageBus_sata:
			icon.Resource = Gui.IconSata

		case vm.StorageBus_scsi:
			icon.Resource = Gui.IconScsi

		case vm.StorageBus_usb:
			icon.Resource = Gui.IconUsb

		case vm.StorageBus_virtio:
			icon.Resource = Gui.IconVirt

		default:
			icon.Resource = Gui.IconUnknown
		}
	} else if m != "" {
		ctrl, medium := st.getControllerAndMediumFromUuid(c, m)
		if ctrl == nil || medium == nil {
			return
		}

		t, isSnapshot := st.buildText(*medium)
		text.Text = t

		icon.Resource = Gui.IconUnknown
		switch ctrl.busType {
		case vm.StorageBus_floppy:
			icon.Resource = Gui.IconFdd
		case vm.StorageBus_usb:
			if isSnapshot {
				icon.Resource = Gui.IconSnapshot
			} else {
				icon.Resource = Gui.IconStick
			}
		default:
			if medium.isImage() {
				icon.Resource = Gui.IconCd
			} else if medium.nonrotational {
				if isSnapshot {
					icon.Resource = Gui.IconSnapshot
				} else {
					icon.Resource = Gui.IconSsd
				}
			} else {
				if isSnapshot {
					icon.Resource = Gui.IconSnapshot
				} else {
					icon.Resource = Gui.IconHdd
				}
			}
		}
	}
	icon.SetMinSize(fyne.NewSize(16, 16)) // gewünschte Größe
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	text.Refresh()
}

func (st *StorageContent) saveOldStorageConfig() {
	st.oldValues.storageControllers = st.oldValues.storageControllers[:0]
	for _, c := range st.storageControllers {
		newController := StorageController{}
		newController = *c
		newController.mediums = make([]*StorageMedium, 0, len(c.mediums))

		for _, m := range c.mediums {
			newMedium := StorageMedium{}
			newMedium = *m
			newController.mediums = append(newController.mediums, &newMedium)
		}
		st.oldValues.storageControllers = append(st.oldValues.storageControllers, &newController)
	}
}

func (st *StorageContent) updateBySelect() {
	c, m := st.getActiveControllerAndMedium()

	st.toolBarItemAddCtrl.Enable()

	if c != nil && m == nil {
		st.toolBarItemRemoveCtrl.Enable()
	} else {
		st.toolBarItemRemoveCtrl.Disable()
	}

	if c != nil && m != nil {
		st.toolBarItemRemoveMedium.Enable()
	} else {
		st.toolBarItemRemoveMedium.Disable()
	}

	if c != nil && len(c.mediums) < st.getMaxNumberOfMedias(c.busType) {
		if c.busType != vm.StorageBus_floppy {
			st.toolBarItemAddGuestAdditions.Enable()
		} else {
			st.toolBarItemAddGuestAdditions.Disable()
		}
		st.toolBarItemAddMedium.Enable()
	} else {
		st.toolBarItemAddMedium.Disable()
		st.toolBarItemAddGuestAdditions.Disable()
	}

	if m != nil && m.isImage() {
		st.toolBarItemEjectMedium.Enable()
	} else {
		st.toolBarItemEjectMedium.Disable()
	}

	if c == nil && m == nil {
		st.formController.Hide()
		st.formMedium.Hide()
		st.formEmpty.Show()
		return
	}
	if c != nil && m == nil {
		st.formEmpty.Hide()
		st.formController.Show()
		st.formMedium.Hide()
		st.name.SetText(c.name)
		st.name.Enable()
		st.bootable.SetChecked(c.bootable)
		st.bootable.Enable()
		return
	}
	if c != nil && m != nil {
		st.formEmpty.Hide()
		st.formController.Hide()
		st.formMedium.Show()
		st.port.SetText(strconv.Itoa(m.port))
		st.device.SetText(strconv.Itoa(m.device))
		st.ssd.SetChecked(m.nonrotational)
		st.isLive.SetChecked(m.isLive)
		st.port.Enable()
		if c.busType == vm.StorageBus_ide {
			st.device.Enable()
		} else {
			st.device.Disable()
		}
		if c.busType == vm.StorageBus_floppy {
			st.ssd.Disable()
			st.isLive.Disable()
			st.isLive.SetChecked(false)
			st.port.Disable()
		} else {
			st.port.Enable()
			st.ssd.Enable()
			if m.isImage() {
				st.isLive.Enable()
			} else {
				st.isLive.Disable()
				st.isLive.SetChecked(false)
			}
		}
		return
	}
}

func (st *StorageContent) treeOnSelected(id widget.TreeNodeID) {
	c, m := st.getStorageAndMediumUUIDFromID(id)
	st.lastSelectedCtrl = c
	st.lastSelectedMedium = m
	st.updateBySelect()
	st.UpdateByStatus()
}

// calles by selection change
func (st *StorageContent) UpdateBySelect() {
	s, v := getActiveServerAndVm()
	st.lastSelectedCtrl = ""
	st.lastSelectedMedium = ""

	if s == nil || v == nil {
		st.DisableAll()
		return
	}
	st.apply.Enable()

	// Chipset
	index := 0
	st.storageControllers = st.storageControllers[:0]
	for {
		str, ok := v.Properties[fmt.Sprintf("storagecontrollername%d", index)]
		if ok {
			ctrl := new(StorageController)
			ctrl.name = str
			ctrl.uuid = uuid.NewString()
			str, ok := v.Properties[fmt.Sprintf("storagecontrollertype%d", index)]
			if ok {
				chip, ok := st.chipsetMapStringToType[str]
				if ok {
					ctrl.controllerType = chip
					bus, ok := st.busMapcontrollerTypeToType[ctrl.controllerType]
					if ok {
						ctrl.busType = bus
					}
				}
			}

			str, ok = v.Properties[fmt.Sprintf("storagecontrollerbootable%d", index)]
			if ok {
				if str == "on" {
					ctrl.bootable = true
				}
			}

			maxPorts := -1
			str, ok = v.Properties[fmt.Sprintf("storagecontrollerportcount%d", index)]
			if ok {
				val, err := strconv.Atoi(str)
				if err == nil {
					maxPorts = val
				}
			}
			maxDevice := -1
			switch ctrl.busType {
			case vm.StorageBus_ide:
				maxDevice = 2
			default:
				maxDevice = 1
			}

			for port := range maxPorts {
				for device := range maxDevice {
					str, ok := v.Properties[fmt.Sprintf("\"%s-%d-%d\"", ctrl.name, port, device)]
					if ok && str != "none" && str != "emptydrive" {
						medium := new(StorageMedium)
						medium.device = device
						medium.port = port
						medium.file = str
						str, ok := v.Properties[fmt.Sprintf("\"%s-ImageUUID-%d-%d\"", ctrl.name, port, device)]
						if ok {
							medium.uuid = str
						}
						str, ok = v.Properties[fmt.Sprintf("\"%s-nonrotational-%d-%d\"", ctrl.name, port, device)]
						if ok {
							if str == "on" {
								medium.nonrotational = true
							}
						}
						str, ok = v.Properties[fmt.Sprintf("\"%s-hot-pluggable-%d-%d\"", ctrl.name, port, device)]
						if ok {
							if str == "on" {
								medium.hotpluggable = true
							}
						}
						str, ok = v.Properties[fmt.Sprintf("\"%s-discard-%d-%d\"", ctrl.name, port, device)]
						if ok {
							if str == "on" {
								medium.discard = true
							}
						}
						str, ok = v.Properties[fmt.Sprintf("\"%s-tempeject-%d-%d\"", ctrl.name, port, device)]
						if ok {
							if str == "on" {
								medium.isLive = true
							}
						}
						ctrl.mediums = append(ctrl.mediums, medium)

					}
				}
			}
			st.storageControllers = append(st.storageControllers, ctrl)
		} else {
			break
		}

		index += 1
	}
	st.saveOldStorageConfig()
	st.tree.Refresh()
	st.updateBySelect()
	st.UpdateByStatus()
}

// called from status updates
func (st *StorageContent) UpdateByStatus() {
	st.updateBySelect()

	_, v := getActiveServerAndVm()
	if v == nil {
		st.DisableAll()
		return
	}
	state, err := v.GetState()
	if err != nil {
		return
	}
	switch state {
	case vm.RunState_unknown, vm.RunState_meditation:
		st.DisableAll()

	case vm.RunState_running, vm.RunState_paused:
		st.toolBarItemAddCtrl.Disable()
		st.toolBarItemRemoveCtrl.Disable()
		st.toolBarItemRemoveMedium.Disable()
		st.port.Disable()
		st.device.Disable()
		st.ssd.Disable()

	case vm.RunState_saved:
		st.toolBarItemAddCtrl.Disable()
		st.toolBarItemRemoveCtrl.Disable()
		st.toolBarItemRemoveMedium.Disable()
		st.toolBarItemAddMedium.Disable()
		st.port.Disable()
		st.device.Disable()
		st.ssd.Disable()

	case vm.RunState_off, vm.RunState_aborted:
		st.updateBySelect()

	default:
		SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
	}
}

func (st *StorageContent) DisableAll() {
	st.apply.Disable()

	st.ssd.Disable()
	st.isLive.Disable()
	st.device.Disable()
	st.port.Disable()

	st.name.Disable()
	st.bootable.Disable()

	st.toolBarItemAddCtrl.Disable()
	st.toolBarItemRemoveCtrl.Disable()
	st.toolBarItemAddMedium.Disable()
	st.toolBarItemRemoveMedium.Disable()
}

func (st *StorageContent) Apply() {
	s, v := getActiveServerAndVm()
	if s == nil || v == nil {
		return
	}

	// Controller
	for _, itemNew := range st.storageControllers {
		itemNew.state = StorageControllerState_invalid
	}
	for _, itemOld := range st.oldValues.storageControllers {
		itemOld.state = StorageControllerState_invalid
	}

	for _, itemNew := range st.storageControllers {
		itemNew.state = StorageControllerState_new
		for _, itemOld := range st.oldValues.storageControllers {
			if itemNew.controllerType == itemOld.controllerType && itemOld.state == StorageControllerState_invalid {
				if !itemNew.isEqualDetails(itemOld) {
					itemNew.state = StorageControllerState_changed
					itemOld.state = StorageControllerState_changed
					itemNew.oldItem = itemOld
				} else {
					itemNew.state = StorageControllerState_unchanged
					itemOld.state = StorageControllerState_unchanged
					itemNew.oldItem = itemOld
				}
				break
			}
		}
	}

	for _, itemOld := range st.oldValues.storageControllers {
		if itemOld.state == StorageControllerState_invalid {
			itemOld.state = StorageControllerState_removed
		}
	}

	// Media
	for _, itemNew := range st.storageControllers {
		if itemNew.state == StorageControllerState_changed || itemNew.state == StorageControllerState_unchanged {
			for _, medium := range itemNew.mediums {
				medium.state = StorageMediumState_invalid
			}
			if itemNew.oldItem != nil {
				for _, medium := range itemNew.oldItem.mediums {
					medium.state = StorageMediumState_invalid
				}
			}
		}
	}

	for _, itemNew := range st.storageControllers {
		if itemNew.state == StorageControllerState_changed || itemNew.state == StorageControllerState_unchanged {
			for _, mediumNew := range itemNew.mediums {
				mediumNew.state = StorageMediumState_new
				if itemNew.oldItem != nil {
					for _, mediumOld := range itemNew.oldItem.mediums {
						if mediumOld.uuid == mediumNew.uuid {
							if mediumOld.isEqual(mediumNew) {
								mediumNew.state = StorageMediumState_unchanged
								mediumOld.state = StorageMediumState_unchanged
							} else {
								mediumNew.state = StorageMediumState_changed
								mediumOld.state = StorageMediumState_changed
							}
							break
						}
					}
				}
			}
		}
	}

	for _, itemNew := range st.storageControllers {
		if itemNew.state == StorageControllerState_changed || itemNew.state == StorageControllerState_unchanged {
			if itemNew.oldItem != nil {
				for _, mediumOld := range itemNew.oldItem.mediums {
					if mediumOld.state == StorageMediumState_invalid {
						mediumOld.state = StorageMediumState_removed
					}
				}
			}
		}
	}

	ResetStatus()
	for _, itemOld := range st.oldValues.storageControllers {
		if itemOld.state == StorageControllerState_removed {
			// ??? -- needed ???
			for _, medium := range itemOld.mediums {
				err := v.DetachMedia(&s.Client, itemOld.name, medium.port, medium.device, vm.MediaSpecial_none, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_storage.removemedia.error", "Remove media from controller '%s' from VM '%s' failed with: %s"), itemOld.name, v.Name, err.Error()), MsgError)
				}
			}
			err := v.RemoveStorageController(&s.Client, itemOld.name, VMStatusUpdateCallBack)
			if err != nil {
				SetStatusText(fmt.Sprintf(lang.X("details.vm_storage.removectrl.error", "Remove storage controller '%s' from VM '%s' failed with: %s"), itemOld.name, v.Name, err.Error()), MsgError)
			}
		}
	}

	// Changed
	for _, itemNew := range st.storageControllers {
		if itemNew.state == StorageControllerState_changed {
			if itemNew.name != itemNew.oldItem.name {
				err := v.RenameStorageController(&s.Client, itemNew.oldItem.name, itemNew.name, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_storage.changename.error", "Change storage controller name to '%s' for VM '%s' failed with: %s"), itemNew.name, v.Name, err.Error()), MsgError)
				}
			}
			if itemNew.bootable != itemNew.oldItem.bootable {
				err := v.SetStorageControllerBootable(&s.Client, itemNew.name, itemNew.bootable, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_storge.setbootable.error", "Set bootable for storage controller '%s' for VM '%s' failed with: %s"), itemNew.name, v.Name, err.Error()), MsgError)
				}
			}
		}
	}

	// New
	for _, itemNew := range st.storageControllers {
		if itemNew.state == StorageControllerState_new {
			ports := st.getMaxNumberOfMedias(itemNew.busType)
			if itemNew.busType == vm.StorageBus_ide {
				ports /= 2
			}
			err := v.AddStorageController(&s.Client, itemNew.name, itemNew.busType, itemNew.controllerType, ports, itemNew.bootable, VMStatusUpdateCallBack)
			if err != nil {
				SetStatusText(fmt.Sprintf(lang.X("details.vm_storge.removectrl.error", "Add storage controller '%s' to VM '%s' failed with: %s"), itemNew.name, v.Name, err.Error()), MsgError)
			}
		}
	}

	// Media
	// New
	for _, itemNew := range st.storageControllers {
		if itemNew.state == StorageControllerState_new {
			for _, medium := range itemNew.mediums {
				err := v.AttachMedia(&s.Client, itemNew.name, medium.getStorageType(itemNew), medium.port, medium.device, medium.getUuidOrFile(), medium.getIsLive(itemNew), medium.getIsSsd(), VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_storge.attachmedia.error", "Attach media to storage controller '%s' for VM '%s' failed with: %s"), itemNew.name, v.Name, err.Error()), MsgError)
				}
			}
		}
	}

	// Changed & Unchanged
	for _, itemNew := range st.storageControllers {
		if itemNew.state == StorageControllerState_changed || itemNew.state == StorageControllerState_unchanged {
			// remove
			for _, medium := range itemNew.oldItem.mediums {
				if medium.state == StorageMediumState_removed {
					err := v.DetachMedia(&s.Client, itemNew.name, medium.port, medium.device, vm.MediaSpecial_none, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_storge.dettach.error", "Dettach media from storage controller '%s' for VM '%s' failed with: %s"), itemNew.name, v.Name, err.Error()), MsgError)
					}
				}
			}

			// changed
			for _, medium := range itemNew.mediums {
				if medium.state == StorageMediumState_changed {
					err := v.DetachMedia(&s.Client, itemNew.name, medium.port, medium.device, vm.MediaSpecial_none, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_storge.dettach.error", "Dettach media from storage controller '%s' for VM '%s' failed with: %s"), itemNew.name, v.Name, err.Error()), MsgError)
					}
					err = v.AttachMedia(&s.Client, itemNew.name, medium.getStorageType(itemNew), medium.port, medium.device, medium.uuid, medium.getIsLive(itemNew), medium.getIsSsd(), VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_storge.dettach.error", "Attach media to storage controller '%s' for VM '%s' failed with: %s"), itemNew.name, v.Name, err.Error()), MsgError)
					}
				}
			}

			// new
			for _, medium := range itemNew.mediums {
				if medium.state == StorageMediumState_new {
					err := v.AttachMedia(&s.Client, itemNew.name, medium.getStorageType(itemNew), medium.port, medium.device, medium.getUuidOrFile(), medium.getIsLive(itemNew), medium.getIsSsd(), VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_storge.attach.error", "Attach media to storage controller '%s' for VM '%s' failed with: %s"), itemNew.name, v.Name, err.Error()), MsgError)
					}
				}
			}
		}
	}
	st.saveOldStorageConfig()
}

func (st *StorageContent) UpdateToolBarIcons() {
	st.toolBarItemAddCtrl.SetIcon(Gui.IconController)
	st.toolBarItemAddMedium.SetIcon(Gui.IconMedia)
	st.toolBarItemAddGuestAdditions.SetIcon(Gui.IconGuestAdditions)
	st.toolBarItemEjectMedium.SetIcon(Gui.IconEject)
	st.toolBar.Refresh()
}
