package vm

import (
	"errors"
	"io"
	"strings"
)

func (m *VMachine) TakeSnapshot(client *VmSshClient, name, description string, live bool, statusWriter io.Writer) error {
	opt := []string{"snapshot", m.UUID, "take", client.quoteArgString(name)}
	if description != "" {
		if !client.IsLocal {
			description = "$'" + description + "'"
		}
	}
	opt = append(opt, description)
	if live {
		opt = append(opt, "--live")
	}
	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, statusWriter)
	if err != nil {
		m.addLogEntry(lines, false)
		err = errors.Join(err, errors.New(strings.Join(lines, ".")))
	}
	return err
}

func (m *VMachine) DeleteSnapshot(client *VmSshClient, uuid string, statusWriter io.Writer) error {
	opt := []string{"snapshot", m.UUID, "delete", client.quoteArgString(uuid)}
	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, statusWriter)
	if err != nil {
		m.addLogEntry(lines, false)
		err = errors.Join(err, errors.New(strings.Join(lines, ".")))
	}
	return err
}

func (m *VMachine) RestoreSnapshot(client *VmSshClient, uuid string, statusWriter io.Writer) error {
	opt := []string{"snapshot", m.UUID, "restore", client.quoteArgString(uuid)}
	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, statusWriter)
	if err != nil {
		m.addLogEntry(lines, false)
		err = errors.Join(err, errors.New(strings.Join(lines, ".")))
	}
	return err
}
