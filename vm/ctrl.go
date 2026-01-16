package vm

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
)

var (
	regexVMList             = regexp.MustCompile(`\"(.*)\"\s*{([0-9-a-fA-F]*)}`)
	regexVMInfoKeyValue     = regexp.MustCompile(`(.*)="(.*)"`)
	regexVMInfoKeyValue2    = regexp.MustCompile(`(.*)=(.*)`)
	regexVMInfoKeyValueDesc = regexp.MustCompile(`(.*)="(.*)`)
	regexVMStart            = regexp.MustCompile(`successfully\s+started`)
	regexVMSave             = regexp.MustCompile(`100%`)
	regexVMPowerOff         = regexp.MustCompile(`100%`)
	//(Driver: ALSA, Controller: HDA, Codec: STAC9221)
	regexVMInfo2Audio1 = regexp.MustCompile(`^Audio:\s+(.*)`)
	regexVMInfo2Audio2 = regexp.MustCompile(`\(Driver:\s+(.*), Controller:\s+(.*), Codec:\s+(.*)\).*`)

	// NIC 1:                       MAC: 0800278ACCAC, Attachment: none, Cable connected: on, Trace: off (file: none), Type: 82583V, Reported speed: 0 Mbps, Boot priority: 0, Promisc Policy: allow-vms, Bandwidth group: none
	// NIC 1:                       MAC: 0800278ACCAC, Attachment: Bridged Interface 'eno1', Cable connected: on, Trace: off (file: none), Type: 82583V, Reported speed: 0 Mbps, Boot priority: 0, Promisc Policy: allow-vms, Bandwidth group: none
	regexVMInfo2Network1 = regexp.MustCompile(`^NIC ([0-9]+):\s*(.*)`)
	regexVMInfo2Network2 = regexp.MustCompile(`MAC:\s*([0-9a-zA-Z]{12}),\s*Attachment:\s*(.*),\s*Cable connected:\s*(.*),\s*Trace:\s*(.*),\s*Type:\s*(.*),\s*Reported speed:\s*(.*),\s*Boot priority:\s*(.*),\s*Promisc Policy:\s*(.*),\s*Bandwidth group:\s*(.*)`)
	// Internal Network 'intnet'
	regexVMInfo2Network3 = regexp.MustCompile(`.*'(.*)'`)

	// USB
	regexVMInfo2USb1 = regexp.MustCompile(`^OHCI USB:\s+(.+)`)
	regexVMInfo2USb2 = regexp.MustCompile(`^EHCI USB:\s+(.+)`)
	regexVMInfo2USb3 = regexp.MustCompile(`^xHCI USB:\s+(.+)`)
)

func (m *VMachine) GetState() (RunState, error) {
	m.lock.RLock()
	state, ok := m.Properties[VM_PROP_KEY_STATE]
	m.lock.RUnlock()
	if ok {
		switch state {
		case "running":
			return RunState_running, nil
		case "poweroff":
			return RunState_off, nil
		case "saved":
			return RunState_saved, nil
		case "aborted":
			return RunState_aborted, nil
		case "paused":
			return RunState_paused, nil
		case "gurumeditation":
			return RunState_meditation, nil
		default:
			return RunState_unknown, nil
		}
	} else {
		return RunState_unknown, errors.New("key not found")
	}
}

func (m *VMachine) GetStateAsString() string {
	state, err := m.GetState()
	if err != nil {
		return ""
	}
	switch state {
	case RunState_unknown:
		return "unknown"
	case RunState_off:
		return "off"
	case RunState_running:
		return "running"
	case RunState_paused:
		return "paused"
	case RunState_saved:
		return "saved"
	case RunState_aborted:
		return "aborted"
	default:
		return "???"
	}
}

func (m *VMachine) runCmd(client *VmSshClient, cmd string, args []string, bUpdateStatus bool, callBack func(uuid string)) ([]string, error) {
	if DEBUG {
		fmt.Println("VM - runCmd", cmd, args)
	}
	lines, err := RunCmd(client, cmd, args, nil, nil)

	if err == nil && bUpdateStatus {
		go m.UpdateStatus(client, callBack)
	}
	if err != nil {
		m.addLogEntry(lines, false)
		err = errors.Join(err, errors.New(strings.Join(lines, ".")))
	}
	return lines, err
}

func (m *VMachine) Start(client *VmSshClient, headless bool, doneCallBack func(err error), callBack func(uuid string)) error {
	c := []string{"startvm", m.UUID}
	if headless {
		c = append(c, "--type", "headless")
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	lines, err := m.runCmd(client, VBOXMANAGE_APP, c, true, callBack)
	if err != nil {
		if doneCallBack != nil {
			doneCallBack(err)
		}
		return err
	}
	if slices.ContainsFunc(lines, regexVMStart.MatchString) {
		if doneCallBack != nil {
			doneCallBack(nil)
		}
		return nil
	}
	err = errors.New("start error")
	if doneCallBack != nil {
		doneCallBack(err)
	}
	return err
}

func (m *VMachine) Pause(client *VmSshClient, callBack func(uuid string)) error {
	cmd := "pause"
	m.UpdateStatus(client, nil)
	state, _ := m.GetState()
	if state == RunState_paused {
		cmd = "resume"
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	lines, err := m.runCmd(client, VBOXMANAGE_APP, []string{"controlvm", m.UUID, cmd}, true, callBack)
	if err != nil {
		return err
	}
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return errors.New("pause error")
}

func (m *VMachine) Save(client *VmSshClient, doneCallBack func(err error), callBack func(uuid string)) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	lines, err := m.runCmd(client, VBOXMANAGE_APP, []string{"controlvm", m.UUID, "savestate"}, true, callBack)
	if err != nil {
		if doneCallBack != nil {
			doneCallBack(err)
		}
		return err
	}
	if len(lines) == 2 && regexVMSave.MatchString(lines[0]) {
		if doneCallBack != nil {
			doneCallBack(nil)
		}
		return nil
	}
	err = errors.New("save error")
	if doneCallBack != nil {
		doneCallBack(err)
	}
	return err
}

func (m *VMachine) Reset(client *VmSshClient, callBack func(uuid string)) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	lines, err := m.runCmd(client, VBOXMANAGE_APP, []string{"controlvm", m.UUID, "reset"}, true, callBack)
	if err != nil {
		return err
	}
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return errors.New("reset error")
}

func (m *VMachine) Off(client *VmSshClient, callBack func(uuid string)) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	lines, err := m.runCmd(client, VBOXMANAGE_APP, []string{"controlvm", m.UUID, "poweroff"}, true, callBack)
	if err != nil {
		return err
	}
	if len(lines) == 2 && regexVMPowerOff.MatchString(lines[0]) {
		return nil
	}
	return errors.New("off error")
}

func (m *VMachine) Shutdown(client *VmSshClient, doneCallBack func(err error), callBack func(uuid string)) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	lines, err := m.runCmd(client, VBOXMANAGE_APP, []string{"controlvm", m.UUID, "acpipowerbutton"}, true, callBack)
	if err != nil {
		if doneCallBack != nil {
			doneCallBack(err)
		}
		return err
	}
	if len(lines) == 1 && lines[0] == "" {
		if doneCallBack != nil {
			doneCallBack(nil)
		}
		return nil
	}
	err = errors.New("powerbutton error")
	if doneCallBack != nil {
		doneCallBack(err)
	}
	return err
}

func (m *VMachine) DiscardSaveState(client *VmSshClient, callBack func(uuid string)) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	lines, err := m.runCmd(client, VBOXMANAGE_APP, []string{"discardstate", m.UUID}, true, callBack)
	if err != nil {
		return err
	}
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return errors.New("powerbutton error")
}

func (m *VMachine) updateStatusEx(client *VmSshClient) error {
	lines, err := m.runCmd(client, VBOXMANAGE_APP, []string{"showvminfo", m.UUID}, false, nil)
	if err != nil {
		return err
	}
	if m.Properties == nil {
		m.Properties = make(map[string]string, len(lines))
	}
	for _, line := range lines {
		items := regexVMInfo2Audio1.FindStringSubmatch(line)
		if len(items) == 2 {
			items = regexVMInfo2Audio2.FindStringSubmatch(items[1])
			if len(items) == 4 {
				m.Properties["audio_driver"] = strings.ToLower(items[1])
				m.Properties["audio_controller"] = strings.ToLower(items[2])
				m.Properties["audio_codec"] = strings.ToLower(items[3])
			}
			continue
		}

		// Network
		items = regexVMInfo2Network1.FindStringSubmatch(line)
		if len(items) == 3 {
			index, err := strconv.Atoi(items[1])
			if err == nil {
				items = regexVMInfo2Network2.FindStringSubmatch(items[2])
				if len(items) == 10 {
					m.Properties[fmt.Sprintf("nic%d_mac", index)] = items[1]
					items2 := regexVMInfo2Network3.FindStringSubmatch(items[2])
					if len(items2) == 2 {
						m.Properties[fmt.Sprintf("nic%d_name", index)] = items2[1]
					}
					m.Properties[fmt.Sprintf("nic%d_connected", index)] = strings.ToLower(items[3])
					m.Properties[fmt.Sprintf("nic%d_promiscuous", index)] = strings.ToLower(items[8])
				}
			}
			continue
		}
		// USB
		items = regexVMInfo2USb1.FindStringSubmatch(line)
		if len(items) == 2 {
			if items[1] == "enabled" {
				m.Properties["usb1"] = "on"
			} else {
				m.Properties["usb1"] = "off"
			}
			continue
		}
		items = regexVMInfo2USb2.FindStringSubmatch(line)
		if len(items) == 2 {
			if items[1] == "enabled" {
				m.Properties["usb2"] = "on"
			} else {
				m.Properties["usb2"] = "off"
			}
			continue
		}
		items = regexVMInfo2USb3.FindStringSubmatch(line)
		if len(items) == 2 {
			if items[1] == "enabled" {
				m.Properties["usb3"] = "on"
			} else {
				m.Properties["usb3"] = "off"
			}
			continue
		}
	}
	return nil
}

func (m *VMachine) UpdateStatusEx(client *VmSshClient) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.updateStatusEx(client)
}

func (m *VMachine) UpdateStatus(client *VmSshClient, callBack func(uuid string)) error {
	if m.lock.TryLock() {
		defer m.lock.Unlock()
		lines, err := m.runCmd(client, VBOXMANAGE_APP, []string{"showvminfo", "--machinereadable", m.UUID}, false, nil)
		if err != nil {
			return err
		}
		if m.Properties == nil {
			m.Properties = make(map[string]string, len(lines))
		} else {
			clear(m.Properties)
		}
		lastWasDesc := false
		for _, line := range lines {
			if line == "" {
				lastWasDesc = false
				continue
			}
			items := regexVMInfoKeyValue.FindStringSubmatch(line)
			if len(items) == 3 {
				lastWasDesc = false
				m.Properties[items[1]] = items[2]
				continue
			}
			items = regexVMInfoKeyValue2.FindStringSubmatch(line)
			if len(items) == 3 {
				if items[1] == "description" {
					lastWasDesc = true
					m.Properties[items[1]] = items[2][1:]
				} else {
					lastWasDesc = false
					m.Properties[items[1]] = items[2]
				}
				continue
			}
			items = regexVMInfoKeyValueDesc.FindStringSubmatch(line)
			if len(items) == 3 {
				if items[1] == "description" {
					lastWasDesc = true
					m.Properties[items[1]] = items[2]
				} else {
					lastWasDesc = false
					m.Properties[items[1]] = items[2]
				}
				continue
			}
			if lastWasDesc {
				// lastWasDesc = false
				if line[len(line)-1] == '"' {
					line = line[:len(line)-1]
				}
				m.Properties["description"] = m.Properties["description"] + "\n" + line
				continue
			}
		}
		if callBack != nil {
			callBack(m.UUID)
		}
		return nil
	} else {
		return errors.New("update is already running")
	}
}

func GetVMs(client *VmSshClient) ([]*VMachine, error) {
	vm := VMachine{}
	lines, err := vm.runCmd(client, VBOXMANAGE_APP, []string{"list", "vms"}, false, nil)
	if err != nil {
		return nil, err
	}

	machines := make([]*VMachine, 0, len(lines))
	for _, line := range lines {
		item, err := newFromVMList(line)
		if err == nil {
			machines = append(machines, item)
		}
	}

	slices.SortFunc(machines, func(a, b *VMachine) int {
		A := a.Name
		B := b.Name
		an := strings.ToLower(A)
		bn := strings.ToLower(B)

		if an == bn {
			if A < B {
				return -1
			}
			if A > B {
				return 1
			}
			return 0
		}
		if an < bn {
			return -1
		}
		return 1
	})

	return machines, nil
}

func newFromVMList(str string) (*VMachine, error) {
	items := regexVMList.FindStringSubmatch(str)
	if items == nil || len(items) != 3 {
		return nil, errors.New("unable to parse")
	}

	return &VMachine{
		Name:      items[1],
		UUID:      items[2],
		lock:      new(sync.RWMutex),
		logBuffer: make([][]string, 0, MAX_LOG_ENTRIES+1),
	}, nil
}
