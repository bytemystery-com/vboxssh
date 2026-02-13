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

package data

import (
	"errors"
	"strings"
	"sync"

	"bytemystery-com/vboxssh/omap"
	"bytemystery-com/vboxssh/server"
	"bytemystery-com/vboxssh/vm"
)

const (
	DEFAULT_NUMBER_OF_SERVERS        = 10
	DEFAULT_NUMBER_OF_VMS_PER_SERVER = 10
)

type VmData struct {
	Lock       *sync.RWMutex
	ServerList omap.OMap[string, *vm.VmServer]             // server UUID -> vmserver
	VmList     map[string]*omap.OMap[string, *vm.VMachine] // server UUID -> List of VMs
}

func NewVmData() *VmData {
	return &VmData{
		Lock:       new(sync.RWMutex),
		ServerList: omap.NewOMap[string, *vm.VmServer](DEFAULT_NUMBER_OF_SERVERS),
		VmList:     make(map[string]*omap.OMap[string, *vm.VMachine], DEFAULT_NUMBER_OF_SERVERS),
	}
}

func (v *VmData) LoadData(servers []vm.VmServer) {
	v.Lock.Lock()
	defer v.Lock.Unlock()
	for _, item := range servers {
		vmNew := vm.NewVmServer(item.Server)
		vmNew.FloppyImagesPath = item.FloppyImagesPath
		vmNew.DvdImagesPath = item.DvdImagesPath
		vmNew.HddImagesPath = item.HddImagesPath
		vmNew.OvaPath = item.OvaPath
		v.ServerList.Add(vmNew.UUID, &vmNew)
		m := omap.NewOMap[string, *vm.VMachine](DEFAULT_NUMBER_OF_VMS_PER_SERVER)
		v.VmList[vmNew.UUID] = &m
	}
	v.sortServerList(false)
	for _, s := range v.ServerList.GetValues() {
		go s.Connect(nil, nil)
	}
}

func (v *VmData) sortVmList(lock bool, uuid string) {
	if lock {
		v.Lock.Lock()
		defer v.Lock.Unlock()
	}
	v.VmList[uuid].Sort(func(a, b *vm.VMachine) int {
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
}

func (v *VmData) sortServerList(lock bool) {
	if lock {
		v.Lock.Lock()
		defer v.Lock.Unlock()
	}
	v.ServerList.Sort(func(a, b *vm.VmServer) int {
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
}

func (v *VmData) GetNumberOfServers(lock bool) int {
	if lock {
		v.Lock.RLock()
		defer v.Lock.RUnlock()
	}
	return v.ServerList.Len()
}

func (v *VmData) AddData(server server.Server, fOk func(), fErr func(error)) *vm.VmServer {
	v.Lock.Lock()
	defer v.Lock.Unlock()
	vmNew := vm.NewVmServer(server)
	v.ServerList.Add(vmNew.UUID, &vmNew)
	m := omap.NewOMap[string, *vm.VMachine](DEFAULT_NUMBER_OF_VMS_PER_SERVER)
	v.VmList[vmNew.UUID] = &m

	v.sortServerList(false)
	go vmNew.Connect(fOk, fErr)
	return &vmNew
}

func (v *VmData) RemoveData(s *vm.VmServer) {
	v.Lock.Lock()
	defer v.Lock.Unlock()
	v.ServerList.RemoveByKey(s.UUID)
	delete(v.VmList, s.UUID)
}

// Called from tree
func (v *VmData) GetServer(uuid string, lock bool) *vm.VmServer {
	if lock {
		v.Lock.RLock()
		defer v.Lock.RUnlock()
	}
	x, _ := v.ServerList.GetByKey(uuid)
	return x
}

// Called from tree (update)
func (v *VmData) GetVm(serverUuid, vmUuid string, lock bool) *vm.VMachine {
	if lock {
		v.Lock.RLock()
		defer v.Lock.RUnlock()
	}
	m := v.VmList[serverUuid]
	if m == nil {
		return nil
	}
	x, _ := m.GetByKey(vmUuid)
	return x
}

// Called from tree
func (v *VmData) GetVms(serverUuid string, lock bool) []*vm.VMachine {
	if lock {
		v.Lock.RLock()
		defer v.Lock.RUnlock()
	}
	m := v.VmList[serverUuid]
	if m == nil {
		return nil
	}
	return m.GetValues()
}

// Called from tree
func (v *VmData) GetServers(lock bool) []*vm.VmServer {
	if lock {
		v.Lock.RLock()
		defer v.Lock.RUnlock()
	}
	return v.ServerList.GetValues()
}

func (v *VmData) UpdateVmList(serverUuid string) error {
	// TODO
	if v.Lock.TryLock() {
		defer v.Lock.Unlock()
		// fmt.Println("update vm list for", serverUuid)
		s := v.GetServer(serverUuid, false)
		if s == nil {
			return errors.New("unknown server uuid")
		}
		if !s.IsConnected() {
			return errors.New("not connected")
		}
		vms, err := vm.GetVMs(&s.Client)
		if err != nil {
			return err
		}

		m := v.VmList[serverUuid]
		if m == nil {
			return errors.New("unknown server uuid 2")
		}
		m.Clear()
		for _, item := range vms {
			m.Add(item.UUID, item)
		}
		v.sortVmList(false, serverUuid)
		return nil
	} else {
		return errors.New("already locked")
	}
}
