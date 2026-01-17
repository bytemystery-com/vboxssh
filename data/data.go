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
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"bytemystery-com/vboxssh/server"
	"bytemystery-com/vboxssh/vm"
)

const (
	DEFAULT_NUMBER_OF_SERVERS        = 10
	DEFAULT_NUMBER_OF_VMS_PER_SERVER = 10
)

type VmData struct {
	ServerMap      map[string]*vm.VmServer            // server UUID -> *VMserver
	ServerMapVmMap map[string]map[string]*vm.VMachine // server UUID, VMs
	Lock           *sync.RWMutex
}

func NewVmData() *VmData {
	return &VmData{
		Lock:           new(sync.RWMutex),
		ServerMap:      make(map[string]*vm.VmServer, DEFAULT_NUMBER_OF_SERVERS),
		ServerMapVmMap: make(map[string]map[string]*vm.VMachine, DEFAULT_NUMBER_OF_VMS_PER_SERVER),
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
		v.ServerMap[vmNew.UUID] = &vmNew
		v.ServerMapVmMap[vmNew.UUID] = make(map[string]*vm.VMachine, DEFAULT_NUMBER_OF_VMS_PER_SERVER)
	}
	for _, s := range v.ServerMap {
		go s.Connect(nil, nil)
	}
}

func (v *VmData) AddData(server server.Server, fOk func(), fErr func(error)) *vm.VmServer {
	v.Lock.Lock()
	defer v.Lock.Unlock()
	vmNew := vm.NewVmServer(server)
	v.ServerMap[vmNew.UUID] = &vmNew
	v.ServerMapVmMap[vmNew.UUID] = make(map[string]*vm.VMachine, DEFAULT_NUMBER_OF_VMS_PER_SERVER)
	go vmNew.Connect(fOk, fErr)
	return &vmNew
}

func (v *VmData) RemoveData(s *vm.VmServer) {
	v.Lock.Lock()
	defer v.Lock.Unlock()
	delete(v.ServerMap, s.UUID)
	delete(v.ServerMapVmMap, s.UUID)
}

func (v *VmData) GetServer(uuid string, lock bool) *vm.VmServer {
	if lock {
		v.Lock.RLock()
		defer v.Lock.RUnlock()
	}
	s, ok := v.ServerMap[uuid]
	if !ok {
		return nil
	}
	return s
}

func (v *VmData) GetVm(serverUuid, vmUuid string, lock bool) *vm.VMachine {
	if lock {
		v.Lock.RLock()
		defer v.Lock.RUnlock()
	}
	vma := v.ServerMapVmMap[serverUuid][vmUuid]
	return vma
}

func (v *VmData) GetVms(serverUuid string, sorted, lock bool) []*vm.VMachine {
	if lock {
		v.Lock.RLock()
	}
	data, ok := v.ServerMapVmMap[serverUuid]
	if lock {
		v.Lock.RUnlock()
	}
	if !ok {
		return nil
	}
	vms := slices.Collect(maps.Values(data))
	if vms == nil {
		return nil
	}
	if !sorted {
		return vms
	}

	slices.SortFunc(vms, func(a, b *vm.VMachine) int {
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
	return vms
}

func (v *VmData) GetServers(lock bool) []*vm.VmServer {
	if lock {
		v.Lock.RLock()
	}
	servers := slices.Collect(maps.Values(v.ServerMap))
	if lock {
		v.Lock.RUnlock()
	}
	if servers == nil {
		return nil
	}

	slices.SortFunc(servers, func(a, b *vm.VmServer) int {
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
	return servers
}

func (v *VmData) UpdateVmList(serverUuid string) error {
	// TODO
	if v.Lock.TryLock() {
		defer v.Lock.Unlock()
		fmt.Println("update vm list for", serverUuid)
		s, ok := v.ServerMap[serverUuid]
		if !ok {
			return errors.New("unknown server uuid")
		}
		if !s.IsConnected() {
			return errors.New("not connected")
		}
		vms, err := vm.GetVMs(&s.Client)
		if err != nil {
			return err
		}
		m := v.ServerMapVmMap[serverUuid]
		clear(m)
		for _, item := range vms {
			m[item.UUID] = item
		}
		return nil
	} else {
		return errors.New("already locked")
	}
}
