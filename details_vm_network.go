package main

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"bytemystery-com/vboxssh/util"

	"bytemystery-com/vboxssh/vm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type oldNetworkType struct {
	enabled     bool
	network     int
	name        string
	adapter     int
	promiscuous int
	mac         string
	connected   bool
}

type NetworkTab struct {
	number         int
	oldValues      oldNetworkType
	enabled        *widget.Check
	network        *widget.Select
	name           *widget.Select
	nameEntry      *widget.SelectEntry
	adapter        *widget.Select
	promiscuous    *widget.Select
	mac            *widget.Entry
	newMac         *widget.Button
	cableConnected *widget.Check

	apply   *widget.Button
	tabItem *container.TabItem

	networkMapStringToIndex     map[string]int
	networkMapIndexToType       map[int]vm.NetType
	adapterMapStringToIndex     map[string]int
	adapterMapIndexToType       map[int]vm.NicType
	promiscuousMapStringToIndex map[string]int
	promiscuousMapIndexToType   map[int]vm.PromiscType
}

var _ DetailsInterface = (*NetworkTab)(nil)

func NewNetworkTab(index int) *NetworkTab {
	netTab := NetworkTab{
		number:                      index,
		networkMapStringToIndex:     map[string]int{"none": -1, "nat": 0, "bridged": 1, "intnet": 2, "hostonly": 3, "generic": 4, "natnetwork": 5, "cloudnetwork": 6, "null": 7},
		networkMapIndexToType:       map[int]vm.NetType{0: vm.Net_nat, 1: vm.Net_bridged, 2: vm.Net_intnet, 3: vm.Net_hostonly, 4: vm.Net_generic, 5: vm.Net_natnetwork, 6: vm.Net_cloudnetwork, 7: vm.Net_null},
		adapterMapStringToIndex:     map[string]int{"am79c970a": 0, "am79c973": 1, "82540em": 2, "82543gc": 3, "82545em": 4, "82583v": 5, "virtio": 6, "usbnet": 7},
		adapterMapIndexToType:       map[int]vm.NicType{0: vm.Nic_amdpcnetpcii, 1: vm.Nic_amdpcnetfastiii, 2: vm.Nic_intelpro1000mtdesktop, 3: vm.Nic_intelpro1000tserver, 4: vm.Nic_intelpro1000mtserver, 5: vm.Nic_intel82583Vgigabit, 6: vm.Nic_virtio, 7: vm.Nic_usbnet},
		promiscuousMapStringToIndex: map[string]int{"deny": 0, "allow-vms": 1, "allow-all": 2},
		promiscuousMapIndexToType:   map[int]vm.PromiscType{0: vm.Promisc_deny, 1: vm.Promisc_allowvms, 2: vm.Promisc_allowall},
	}

	netTab.apply = widget.NewButton(lang.X("details.vm_network.apply", "Apply"), func() {
		netTab.Apply()
	})
	netTab.apply.Importance = widget.HighImportance

	netTab.enabled = widget.NewCheck(lang.X("details.vm_network.enabled", "Enabled"), func(checked bool) {
		netTab.adjustNameField()
	})
	netTab.network = widget.NewSelect([]string{
		lang.X("details.vm_network.attach.nat", "NAT"),
		lang.X("details.vm_network.attach.bridge", "Bridged"),
		lang.X("details.vm_network.attach.internal", "Internal"),
		lang.X("details.vm_network.attach.hostonly", "Host only"),
		lang.X("details.vm_network.attach.generic", "Generic"),
		lang.X("details.vm_network.attach.natnetwork", "NAT network"),
		lang.X("details.vm_network.attach.cloud", "Cloud"),
		lang.X("details.vm_network.attach.notattached", "Not attached"),
	}, func(s string) {
		netTab.adjustNameField()
		netTab.UpdateByStatus()
	})

	netTab.nameEntry = widget.NewSelectEntry([]string{})
	netTab.name = widget.NewSelect([]string{}, nil)
	netTab.nameEntry.Hide()

	netTab.adapter = widget.NewSelect([]string{
		lang.X("details.vm_network.adapter.am79C970a", "PCnet-PCI II (Am79C970A)"),
		lang.X("details.vm_network.adapter.am79C973", "PCnet-FAST III (Am79C973)"),
		lang.X("details.vm_network.adapter.82540em", "Intel PRO/1000 MT Desktop (82540EM)"),
		lang.X("details.vm_network.adapter.82543gc", "Intel PRO/1000 T Server (82543GC)"),
		lang.X("details.vm_network.adapter.82545em", "Intel PRO/1000 MT Server (82545EM)"),
		lang.X("details.vm_network.adapter.82583v", "Intel 82583V Gigabit Network Connection (82583V)"),
		lang.X("details.vm_network.adapter.virt", "Paravirtualized Network (virtio-net)"),
		lang.X("details.vm_network.adapter.usbnet", "Ethernet over USB (usbnet)"),
	}, nil)

	netTab.promiscuous = widget.NewSelect([]string{
		lang.X("details.vm_network.promiscuous.deny", "Deny"),
		lang.X("details.vm_network.promiscuous.allowvms", "Allow VMs"),
		lang.X("details.vm_network.promiscuous.allowall", "Allow All"),
	}, nil)

	netTab.mac = widget.NewEntry()
	netTab.newMac = widget.NewButtonWithIcon(lang.X("details.vm_network.newmac", "New"), theme.SearchReplaceIcon(), func() {
		netTab.NewMac()
	})

	netTab.cableConnected = widget.NewCheck(lang.X("details.vm_network.connected", "Cable connected"), nil)

	grid1 := container.New(layout.NewFormLayout(),
		netTab.enabled, util.NewFiller(0, 0),
	)

	grid2 := container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("details.vm_network.attached", "Attached to")), netTab.network,
		widget.NewLabel(lang.X("details.vm_network.name", "Name")), container.NewStack(netTab.name, netTab.nameEntry),
		widget.NewLabel(lang.X("details.vm_network.adapter", "Adapter")), netTab.adapter,
		widget.NewLabel(lang.X("details.vm_network.promiscuous", "Promiscuous Mode")), netTab.promiscuous,
		widget.NewLabel(lang.X("details.vm_network.mac", "MAC address")), container.NewBorder(nil, nil, nil, netTab.newMac, netTab.mac),
		util.NewFiller(0, 0), netTab.cableConnected,
	)

	formWidth := util.GetFormWidth()
	gridWrap1 := container.NewGridWrap(fyne.NewSize(formWidth, grid1.MinSize().Height), grid1)
	gridWrap2 := container.NewGridWrap(fyne.NewSize(formWidth, grid2.MinSize().Height), grid2)

	gridWrap := container.NewVBox(util.NewVFiller(0.5), gridWrap1, gridWrap2)

	c := container.NewVBox(container.NewHBox(gridWrap),
		container.NewHBox(layout.NewSpacer(), netTab.apply, util.NewFiller(32, 0)))

	netTab.tabItem = container.NewTabItem(lang.X("details.vm_network.tab",
		fmt.Sprintf(lang.X("details.vm_network.tab.header", "Adapter %d"), index+1)), c)
	return &netTab
}

func (n *NetworkTab) geNicName() string {
	return fmt.Sprintf("nic%d", n.number+1)
}

func (n *NetworkTab) UpdateBySelect() {
	s, v := getActiveServerAndVm()

	if s == nil || v == nil {
		n.DisableAll()
		return
	}
	n.apply.Enable()

	nicName := n.geNicName()

	if v.Properties[n.geNicName()] == "none" {
		n.enabled.SetChecked(false)
		n.adjustNameField()
		n.oldValues.enabled = false
	} else {
		util.SelectEntryFromProperty(n.network, v, nicName, n.networkMapStringToIndex, &n.oldValues.network)
		n.oldValues.enabled = true
		n.enabled.SetChecked(true)

		n.adjustNameField()

		index := n.network.SelectedIndex()
		name := v.Properties[nicName+"_name"]
		if index == 2 || index == 4 {
			n.nameEntry.SetText(name)
		} else {
			fyne.Do(func() {
				n.name.SetSelected(name)
			})
		}
		n.oldValues.name = name

		util.SelectEntryFromProperty(n.adapter, v, fmt.Sprintf("nictype%d", n.number+1), n.adapterMapStringToIndex, &n.oldValues.adapter)

		util.SelectEntryFromProperty(n.promiscuous, v, nicName+"_promiscuous", n.promiscuousMapStringToIndex, &n.oldValues.promiscuous)

		n.mac.SetText(v.Properties[nicName+"_mac"])
		n.oldValues.mac = n.mac.Text

		util.CheckFromProperty(n.cableConnected, v, nicName+"_connected", "on", &n.oldValues.connected)
	}
	n.UpdateByStatus()
}

func (n *NetworkTab) UpdateByStatus() {
	_, v := getActiveServerAndVm()
	if v != nil {
		state, err := v.GetState()
		if err != nil {
			return
		}
		switch state {
		case vm.RunState_unknown, vm.RunState_meditation:
			n.DisableAll()

		case vm.RunState_running, vm.RunState_paused, vm.RunState_saved:
			n.adjustEnable(true)

		case vm.RunState_off, vm.RunState_aborted:
			n.enabled.Enable()
			n.adjustEnable(false)

		default:
			SetStatusText(lang.X("status.unknown_vm_state", "!!! Unknown VM state !!!"), MsgError)
		}
	} else {
		n.DisableAll()
	}
}

func (n *NetworkTab) adjustNameField() {
	s, v := getActiveServerAndVm()

	if s == nil || v == nil {
		n.DisableAll()
		return
	}

	index := n.network.SelectedIndex()
	var adapters []vm.NicAdapter
	var err error

	switch index {
	case 1:
		adapters, err = s.GetBridgeAdapters(false)
	case 2:
		adapters, err = s.GetInternalAdapters(false)
	case 3:
		adapters, err = s.GetHostAdapters(false)
	case 5:
		adapters, err = s.GetNatAdapters(false)
	case 6:
		adapters, err = s.GetCloudAdapters(false)
	}
	if err != nil {
		return
	}
	list := make([]string, 0, len(adapters))
	for _, item := range adapters {
		list = append(list, item.Name)
	}
	n.nameEntry.SetOptions(nil)
	n.name.SetOptions(nil)
	n.nameEntry.SetText("")
	n.name.ClearSelected()

	if index == 2 {
		n.nameEntry.SetOptions(list)
		if len(list) == 1 {
			fyne.Do(func() {
				n.nameEntry.SetText(list[0])
			})
		}
	} else {
		n.name.SetOptions(list)
		if len(list) == 1 {
			fyne.Do(func() {
				n.name.SetSelectedIndex(0)
			})
		}
	}
	n.adjustEnable(false)
	if n.mac.Text == "" {
		n.NewMac()
	}
}

func (n *NetworkTab) adjustEnable(running bool) {
	if n.enabled.Checked {
		if running {
			n.enabled.Disable()
			n.adapter.Disable()
			n.mac.Disable()
			n.newMac.Disable()
		} else {
			n.enabled.Enable()
			n.adapter.Enable()
			n.mac.Enable()
			n.newMac.Enable()
		}
		n.network.Enable()
		n.name.Enable()
		n.nameEntry.Enable()
		n.promiscuous.Enable()
		n.cableConnected.Enable()

		index := n.network.SelectedIndex()

		if index == 0 || index == 7 || index == 4 {
			n.promiscuous.Disable()
		}
		switch index {
		case 0, 7:
			n.nameEntry.Hide()
			n.name.Show()
			n.name.Disable()
			if index == 7 {
				n.promiscuous.Disable()
			}
		case 1, 3, 5, 6:
			n.nameEntry.Hide()
			n.name.Show()
			n.name.Enable()
		case 2, 4:
			n.name.Hide()
			n.nameEntry.Show()
			n.nameEntry.Enable()
		default:
			n.nameEntry.Hide()
			n.name.Show()
			n.name.Disable()
		}
	} else {
		n.network.Disable()
		n.name.Disable()
		n.nameEntry.Disable()
		n.adapter.Disable()
		n.promiscuous.Disable()
		n.mac.Disable()
		n.newMac.Disable()
		n.cableConnected.Disable()
	}
}

func (n *NetworkTab) DisableAll() {
	n.enabled.Disable()
	n.network.Disable()
	n.name.Disable()
	n.nameEntry.Disable()
	n.adapter.Disable()
	n.promiscuous.Disable()
	n.mac.Disable()
	n.newMac.Disable()
	n.cableConnected.Disable()
	n.apply.Disable()
}

func (n *NetworkTab) NewMac() {
	base := []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}
	str := "080027"
	for range 6 {
		n, err := rand.Int(rand.Reader, big.NewInt(16))
		if err != nil {
			return
		}
		str = str + string(base[n.Uint64()])
	}
	n.mac.SetText(str)
}

func (n *NetworkTab) Apply() {
	s, v := getActiveServerAndVm()
	if v != nil {
		ResetStatus()
		if !n.enabled.Disabled() {
			if n.enabled.Checked != n.oldValues.enabled {
				var err error
				if !n.enabled.Checked {
					err = v.SetNetType(&s.Client, n.number+1, vm.Net_none, VMStatusUpdateCallBack)
					if err != nil {
						SetStatusText(fmt.Sprintf(lang.X("details.vm_net.disable.error", "Disable network for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
					} else {
						n.oldValues.enabled = false
					}
				}
			}
		}
		if !n.network.Disabled() {
			index := n.network.SelectedIndex()
			if index != n.oldValues.network {
				if index >= 0 {
					val, ok := n.networkMapIndexToType[index]
					if ok {
						err := v.SetNetType(&s.Client, n.number+1, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_net.network.error", "Set network for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							n.oldValues.network = index
						}
					}
				}
			}
		}
		if !n.name.Disabled() && !n.name.Hidden {
			val := n.name.Selected
			if val != n.oldValues.name {
				var err error
				switch n.network.SelectedIndex() {
				case 1: // Bridged
					err = v.SetBridgeAdapter(&s.Client, n.number+1, val, VMStatusUpdateCallBack)
				case 3: // Hostonly
					err = v.SetHostOnlyAdapter(&s.Client, n.number+1, val, VMStatusUpdateCallBack)
				case 5: // NatNetwork
					if val != "" {
						err = v.SetNatAdapter(&s.Client, n.number+1, val, VMStatusUpdateCallBack)
					}
				case 6: // CloudNetwork
					if val != "" {
						err = v.SetCloudNetworkName(&s.Client, n.number+1, val, VMStatusUpdateCallBack)
					}
				default:
					fmt.Println("!!! Unhandeld")
				}
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_net.name.error", "Set name fot network for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					n.oldValues.name = val
				}
			}
		}
		if !n.nameEntry.Disabled() && !n.nameEntry.Hidden {
			val := n.nameEntry.Text
			if val != n.oldValues.name {
				var err error
				switch n.network.SelectedIndex() {
				case 2: // Internal
					err = v.SetInternalNetworkName(&s.Client, n.number+1, val, VMStatusUpdateCallBack)
				case 4: // Generic
					err = v.SetGenericNetworkName(&s.Client, n.number+1, val, VMStatusUpdateCallBack)
				default:
					fmt.Println("!!! Unhandeld")
				}

				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_net.name.error", "Set name fot network for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					n.oldValues.name = val
				}
			}
		}
		if !n.adapter.Disabled() {
			index := n.adapter.SelectedIndex()
			if index != n.oldValues.adapter {
				if index >= 0 {
					val, ok := n.adapterMapIndexToType[index]
					if ok {
						err := v.SetNetDevice(&s.Client, n.number+1, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_net.device.error", "Set net device for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							n.oldValues.adapter = index
						}
					}
				}
			}
		}
		if !n.promiscuous.Disabled() {
			index := n.promiscuous.SelectedIndex()
			if index != n.oldValues.promiscuous {
				if index >= 0 {
					val, ok := n.promiscuousMapIndexToType[index]
					if ok {
						err := v.SetPromiscMode(&s.Client, n.number+1, val, VMStatusUpdateCallBack)
						if err != nil {
							SetStatusText(fmt.Sprintf(lang.X("details.vm_net.promiscuous.error", "Set net promiscuous for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
						} else {
							n.oldValues.promiscuous = index
						}
					}
				}
			}
		}
		if !n.mac.Disabled() && n.mac.Text != n.oldValues.mac {
			err := v.SetMacAddress(&s.Client, n.number+1, n.mac.Text, VMStatusUpdateCallBack)
			if err != nil {
				SetStatusText(fmt.Sprintf(lang.X("details.vm_info.setdescription.error", "Set MAC for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
			}
		}
		if !n.cableConnected.Disabled() {
			if n.cableConnected.Checked != n.oldValues.connected {
				err := v.SetCableConnected(&s.Client, n.number+1, n.cableConnected.Checked, VMStatusUpdateCallBack)
				if err != nil {
					SetStatusText(fmt.Sprintf(lang.X("details.vm_net.connected.error", "Set net cable connected for VM '%s' failed with: %s"), v.Name, err.Error()), MsgError)
				} else {
					n.oldValues.connected = n.cableConnected.Checked
				}
			}
		}
	}
}
