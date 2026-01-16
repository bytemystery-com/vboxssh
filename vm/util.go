package vm

import (
	"errors"
	"regexp"
	"slices"
)

var regex100prozent = regexp.MustCompile(`100%`)

func (m *VMachine) setPropertyInternal(client *VmSshClient, cmds []string, bUpdateStatus bool, callBack func(uuid string)) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	lines, err := m.runCmd(client, VBOXMANAGE_APP, cmds, bUpdateStatus, callBack)
	if err != nil {
		return err
	}
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	if slices.ContainsFunc(lines, regex100prozent.MatchString) {
		return nil
	}
	return errors.New("set property error")
}

func (m *VMachine) setProperty(client *VmSshClient, tag string, value any, callBack func(uuid string)) error {
	strVal, err := argTranslate(value)
	if err != nil {
		return err
	}
	return m.setPropertyInternal(client, []string{"modifyvm", m.UUID, "--" + tag + "=" + strVal}, true, callBack)
}

func (m *VMachine) setPropertyEx(client *VmSshClient, cmd string, tag string, value any, callBack func(uuid string)) error {
	return m.setPropertyEx2(client, cmd, []any{tag, value}, callBack)
}

func (m *VMachine) setPropertyEx2(client *VmSshClient, cmd string, options []any, callBack func(uuid string)) error {
	opStr := make([]string, 0, len(options)+1)
	opStr = append(opStr, cmd)

	for _, value := range options {
		strVal, err := argTranslate(value)
		if err != nil {
			return err
		}
		if strVal != "" {
			opStr = append(opStr, strVal)
		}
	}
	return m.setPropertyInternal(client, opStr, true, callBack)
}
