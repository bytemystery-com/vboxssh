package vm

import (
	"errors"
	"io"
	"sync"

	"bytemystery-com/vboxssh/run"

	"golang.org/x/crypto/ssh"
)

const (
	VBOXMANAGE_APP    = "VBoxManage"
	VM_PROP_KEY_STATE = "VMState"
	MAX_LOG_ENTRIES   = 25
	DEBUG             = true
)

type RunState int

const (
	RunState_unknown RunState = iota - 1
	RunState_off
	RunState_running
	RunState_paused
	RunState_saved
	RunState_aborted
	RunState_meditation
)

type ParaVirtProviderType int

const (
	ParaVirtProvider_none ParaVirtProviderType = iota
	ParaVirtProvider_default
	ParaVirtProvider_legacy
	ParaVirtProvider_minimal // for Mac guests
	ParaVirtProvider_kvm     // for Windows guests
	ParaVirtProvider_hyperV  // for Linux guests
)

type ProcessPriorityType int

const (
	ProcessPriority_default ProcessPriorityType = iota
	ParaVirtProvider_flat
	ProcessPriority_low
	ParaVirtProvider_normal
	ParaVirtProvider_high
)

type ChipSetType int

const (
	ChipSet_piix3 ChipSetType = iota
	ChipSet_ich9
)

type TpmType int

const (
	Tpm_none TpmType = iota
	Tpm_12
	Tpm_20
	Tpm_host
)

type MouseType int

const (
	Mouse_ps2 MouseType = iota
	Mouse_usb
	Mouse_usbtablet
	Mouse_usbmultitouch
	Mouse_usbmtscreenpluspad
)

type KeyboardType int

const (
	Keyboard_ps2 KeyboardType = iota
	Keyboard_usb
)

type FirmwareType int

const (
	Firmware_bios FirmwareType = iota
	Firmware_efi
	Firmware_efi32
	Firmware_efi64
)

type BootType int

const (
	Boot_none BootType = iota
	Boot_floppy
	Boot_dvd
	Boot_disk
	Boot_net
)

type VgaType int

const (
	Vga_none     VgaType = iota
	Vga_vboxvga          // legacy
	Vga_vmsvga           // Linux (guest additions needed)
	Vga_vboxsvga         // new Windows (guest additions needed)
)

type NetType int

const (
	Net_none NetType = iota // no interface
	Net_null                // not connected
	Net_nat
	Net_natnetwork
	Net_bridged
	Net_intnet
	Net_hostonly
	Net_generic
	Net_cloudnetwork
)

type NicType int

const (
	Nic_amdpcnetpcii          NicType = iota // Am79C970A
	Nic_amdpcnetfastiii                      // Am79C973
	Nic_intelpro1000mtdesktop                // 82540EM
	Nic_intelpro1000tserver                  // 82543GC
	Nic_intelpro1000mtserver                 // 82545EM
	Nic_intel82583Vgigabit                   // 82583V
	Nic_virtio
	Nic_usbnet
)

type PromiscType int

const (
	Promisc_deny PromiscType = iota
	Promisc_allowvms
	Promisc_allowall
)

type AudioCodecType int

const (
	AudioCodec_stac9700 AudioCodecType = iota // Emuliert den Conexant STAC9700 Codec
	AudioCodec_ad1980                         // Emuliert den Analog Devices AD1980 Codec
	AudioCodec_stac9221                       // Emuliert den Conexant STAC9221 Codec
	AudioCodec_sb16                           // Emuliert die alte Sound Blaster 16 Karte
)

type AudioControllerType int

const (
	AudioController_ac97 AudioControllerType = iota // ad1980
	AudioController_hda                             // stac9700
	AudioController_sb16                            // sb16
)

type AudioDriverType int

const (
	AudioDriver_none    AudioDriverType = iota
	AudioDriver_default                 // automatic
	AudioDriver_null
	AudioDriver_dsound // Windows
	AudioDriver_was    // Windows
	AudioDriver_oss
	AudioDriver_alsa      // Linux
	AudioDriver_pulse     // Linux
	AudioDriver_coreaudio // Mac
)

type UsbType int

const (
	Usb_none UsbType = iota
	Usb_1
	Usb_2
	Usb_3
)

type VMachine struct {
	Name       string
	UUID       string
	Properties map[string]string
	lock       *sync.RWMutex
	logBuffer  [][]string
}

type NicAdapter struct {
	Name string
}

type UsbDevice struct {
	UUID         string
	Name         string
	ProductId    string
	VendorId     string
	SerialNumber string
	Product      string
	Manufacturer string
	Port         int
}

type OsFamily struct {
	FamilyId string
	Family   string
}

type OsType struct {
	ID   string
	Name string
	OsFamily
	Subtype      string
	Architecture string
	Is64Bit      bool
}

type MediaState int

const (
	MediaState_created MediaState = iota
	MediaState_inaccessible
)

type MediaInfo struct {
	UUID     string
	State    MediaState
	Location string
	UsedBy   []string
}

type DvdInfo struct {
	MediaInfo
}

type FloppyInfo struct {
	MediaInfo
}

type UsedByInfo struct {
	UUID                string
	SnapshotDescription string
	SnapshotUUID        string
}

type HddInfo struct {
	UUID     string
	State    MediaState
	Location string
	UsedBy   []*UsedByInfo
	Parent   string
	Childs   []*HddInfo
}

type RdpSecurityType int

const (
	RdpSecurity_tls RdpSecurityType = iota
	RdpSecurity_rdp
	RdpSecurity_negotiate
)

type RdpAuthType int

const (
	RdpAuth_null RdpAuthType = iota
	RdpAuth_external
	RdpAuth_guest
)

type StorageBusType int

const (
	StorageBus_floppy StorageBusType = iota
	StorageBus_ide
	StorageBus_pcie
	StorageBus_sas
	StorageBus_sata
	StorageBus_scsi
	StorageBus_usb
	StorageBus_virtio
)

type StorageChipsetType int

const (
	StorageChipset_BusLogic StorageChipsetType = iota
	StorageChipset_I82078
	StorageChipset_ICH6
	StorageChipset_IntelAHCI
	StorageChipset_LSILogic
	StorageChipset_LSILogicSAS
	StorageChipset_NVMe
	StorageChipset_PIIX3
	StorageChipset_PIIX4
	StorageChipset_USB
	StorageChipset_VirtIO
	// StorageChipset_VirtioSCSI
)

type StorageType int

const (
	Storage_dvddrive StorageType = iota
	Storage_fdd
	Storage_hdd
)

type MediaSpecialType int

const (
	MediaSpecial_none MediaSpecialType = iota
	MediaSpecial_emptydrive
	MediaSpecial_additions
)

type MediaType int

const (
	Media_disk MediaType = iota
	Media_dvd
	Media_floppy
)

type MediaFormatType int

const (
	MediaFormat_vdi MediaFormatType = iota
	MediaFormat_vmdk
	MediaFormat_vhd
)

type OvaFormatType int

const (
	OvaFormat_legacy OvaFormatType = iota
	OvaFormat_0_9
	OvaFormat_1_0
	OvaFormat_2_0
)

type MacExportType int

const (
	MacExport_all MacExportType = iota
	MacExport_nomacsbutnat
	MacExport_nomacs
)

type MacImportType int

const (
	MacImport_all MacImportType = iota
	MacImport_natmacs
)

type ExtPackInfoType struct {
	Name       string
	Version    string
	Revision   string
	Usable     bool
	WhyUnsable string
}

type VmSshClient struct {
	Client  *ssh.Client
	IsLocal bool
}

func (s *VmSshClient) quoteArgString(arg string) string {
	if s.IsLocal {
		return arg
	}
	return "'" + arg + "'"
}

func RunCmd(client *VmSshClient, cmd string, args []string, userWriterOut, userWriterErr io.Writer) ([]string, error) {
	if client == nil {
		return nil, errors.New("null pointer as client")
	}
	var lines []string
	var err error
	if client.IsLocal {
		lines, err = run.RunLocalCmd(cmd, args, userWriterOut, userWriterErr)
	} else if client.Client != nil {
		lines, err = run.RunSshCmd(client.Client, cmd, args, userWriterOut, userWriterErr)
	} else {
		return nil, errors.New("ssh client is null")
	}

	return lines, err
}
