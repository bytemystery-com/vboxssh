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
