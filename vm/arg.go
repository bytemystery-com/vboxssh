package vm

import (
	"errors"
	"strconv"
)

func argTranslate(value any) (string, error) {
	var strVal string = ""
	switch v := value.(type) {
	case string:
		strVal = v
	case int:
		strVal = strconv.Itoa(v)
	case int64:
		strVal = strconv.FormatInt(v, 10)
	case bool:
		if v {
			strVal = "on"
		} else {
			strVal = "off"
		}
	case ParaVirtProviderType:
		switch v {
		case ParaVirtProvider_none:
			strVal = "none"
		case ParaVirtProvider_default:
			strVal = "default"
		case ParaVirtProvider_legacy:
			strVal = "legacy"
		case ParaVirtProvider_minimal:
			strVal = "minimal"
		case ParaVirtProvider_kvm:
			strVal = "kvm"
		case ParaVirtProvider_hyperV:
			strVal = "hyperv"
		default:
			return "", errors.New("wrong ParaVirtProvider type")
		}

	case ChipSetType:
		switch v {
		case ChipSet_piix3:
			strVal = "piix3"
		case ChipSet_ich9:
			strVal = "ich9"
		default:
			return "", errors.New("wrong ChipSet type")
		}
	case TpmType:
		switch v {
		case Tpm_none:
			strVal = "none"
		case Tpm_12:
			strVal = "1.2"
		case Tpm_20:
			strVal = "2.0"
		case Tpm_host:
			strVal = "host"
		default:
			return "", errors.New("wrong Tpm type")
		}
	case MouseType:
		switch v {
		case Mouse_ps2:
			strVal = "ps2"
		case Mouse_usb:
			strVal = "usb"
		case Mouse_usbtablet:
			strVal = "usbtablet"
		case Mouse_usbmultitouch:
			strVal = "usbmultitouch"
		case Mouse_usbmtscreenpluspad:
			strVal = "usbmtscreenpluspad"
		default:
			return "", errors.New("wrong Mouse type")
		}
	case KeyboardType:
		switch v {
		case Keyboard_ps2:
			strVal = "ps2"
		case Keyboard_usb:
			strVal = "usb"
		default:
			return "", errors.New("wrong keyboard type")
		}
	case FirmwareType:
		switch v {
		case Firmware_bios:
			strVal = "bios"
		case Firmware_efi:
			strVal = "efi"
		case Firmware_efi32:
			strVal = "efi32"
		case Firmware_efi64:
			strVal = "efi64"
		default:
			return "", errors.New("wrong firmware type")
		}
	case BootType:
		switch v {
		case Boot_none:
			strVal = "none"
		case Boot_floppy:
			strVal = "floppy"
		case Boot_dvd:
			strVal = "dvd"
		case Boot_disk:
			strVal = "disk"
		case Boot_net:
			strVal = "net"
		default:
			return "", errors.New("wrong boot type")
		}
	case VgaType:
		switch v {
		case Vga_none:
			strVal = "none"
		case Vga_vboxvga:
			strVal = "vboxvga"
		case Vga_vmsvga:
			strVal = "vmsvga"
		case Vga_vboxsvga:
			strVal = "vboxsvga"
		default:
			return "", errors.New("wrong vga type")
		}
	case NetType:
		switch v {
		case Net_none:
			strVal = "none"
		case Net_null:
			strVal = "null"
		case Net_nat:
			strVal = "nat"
		case Net_natnetwork:
			strVal = "natnetwork"
		case Net_bridged:
			strVal = "bridged"
		case Net_intnet:
			strVal = "intnet"
		case Net_hostonly:
			strVal = "hostonly"
		case Net_generic:
			strVal = "generic"
		case Net_cloudnetwork:
			strVal = "cloud"
		default:
			return "", errors.New("wrong nettype type")
		}
	case NicType:
		switch v {
		case Nic_amdpcnetpcii:
			strVal = "Am79C970A"
		case Nic_amdpcnetfastiii:
			strVal = "Am79C973"
		case Nic_intelpro1000mtdesktop:
			strVal = "82540EM"
		case Nic_intelpro1000tserver:
			strVal = "82543GC"
		case Nic_intelpro1000mtserver:
			strVal = "82545EM"
		case Nic_intel82583Vgigabit:
			strVal = "82583V"
		case Nic_virtio:
			strVal = "virtio"
		case Nic_usbnet:
			strVal = "usbnet"
		default:
			return "", errors.New("wrong nictype type")
		}
	case PromiscType:
		switch v {
		case Promisc_deny:
			strVal = "deny"
		case Promisc_allowvms:
			strVal = "allow-vms"
		case Promisc_allowall:
			strVal = "allow-all"
		default:
			return "", errors.New("wrong promisctype type")
		}
	case AudioCodecType:
		switch v {
		case AudioCodec_stac9700:
			strVal = "stac9700"
		case AudioCodec_ad1980:
			strVal = "ad1980"
		case AudioCodec_stac9221:
			strVal = "stac9221"
		case AudioCodec_sb16:
			strVal = "sb16"
		default:
			return "", errors.New("wrong audiocodectype type")
		}
	case AudioControllerType:
		switch v {
		case AudioController_ac97:
			strVal = "ac97"
		case AudioController_hda:
			strVal = "hda"
		case AudioController_sb16:
			strVal = "sb16"
		default:
			return "", errors.New("wrong audiocontrollertype type")
		}
	case AudioDriverType:
		switch v {
		case AudioDriver_none:
			strVal = "none"
		case AudioDriver_default:
			strVal = "default"
		case AudioDriver_null:
			strVal = "null"
		case AudioDriver_dsound:
			strVal = "dsound"
		case AudioDriver_was:
			strVal = "was"
		case AudioDriver_oss:
			strVal = "oss"
		case AudioDriver_alsa:
			strVal = "alsa"
		case AudioDriver_pulse:
			strVal = "pulse"
		case AudioDriver_coreaudio:
			strVal = "coreaudio"
		default:
			return "", errors.New("wrong audiodrivertype type")
		}
	case RdpSecurityType:
		switch v {
		case RdpSecurity_tls:
			strVal = "TLS"
		case RdpSecurity_rdp:
			strVal = "RDP"
		case RdpSecurity_negotiate:
			strVal = "NEGOTIATE"
		default:
			return "", errors.New("wrong rdpsecurity type")
		}
	case RdpAuthType:
		switch v {
		case RdpAuth_null:
			strVal = "null"
		case RdpAuth_external:
			strVal = "external"
		case RdpAuth_guest:
			strVal = "guest"
		default:
			return "", errors.New("wrong rdpauth type")
		}
	case ProcessPriorityType:
		switch v {
		case ProcessPriority_default:
			strVal = "default"
		case ParaVirtProvider_flat:
			strVal = "flat"
		case ProcessPriority_low:
			strVal = "low"
		case ParaVirtProvider_normal:
			strVal = "normal"
		case ParaVirtProvider_high:
			strVal = "high"
		default:
			return "", errors.New("wrong ProcessPriority type")
		}
	case StorageBusType:
		switch v {
		case StorageBus_floppy:
			strVal = "floppy"
		case StorageBus_ide:
			strVal = "ide"
		case StorageBus_pcie:
			strVal = "pcie"
		case StorageBus_sas:
			strVal = "sas"
		case StorageBus_sata:
			strVal = "sata"
		case StorageBus_scsi:
			strVal = "scsi"
		case StorageBus_usb:
			strVal = "usb"
		case StorageBus_virtio:
			strVal = "virtio"
		default:
			return "", errors.New("wrong Storage bus type")
		}
	case StorageChipsetType:
		switch v {
		case StorageChipset_BusLogic:
			strVal = "BusLogic"
		case StorageChipset_I82078:
			strVal = "I82078"
		case StorageChipset_ICH6:
			strVal = "ICH6"
		case StorageChipset_IntelAHCI:
			strVal = "IntelAHCI"
		case StorageChipset_LSILogic:
			strVal = "LSILogic"
		case StorageChipset_LSILogicSAS:
			strVal = "LSILogicSAS"
		case StorageChipset_NVMe:
			strVal = "NVMe"
		case StorageChipset_PIIX3:
			strVal = "PIIX3"
		case StorageChipset_PIIX4:
			strVal = "PIIX4"
		case StorageChipset_USB:
			strVal = "USB"
		case StorageChipset_VirtIO:
			strVal = "VirtIO"
		default:
			return "", errors.New("wrong Storage chipset type")
		}
	case StorageType:
		switch v {
		case Storage_dvddrive:
			strVal = "dvddrive"
		case Storage_fdd:
			strVal = "fdd"
		case Storage_hdd:
			strVal = "hdd"
		default:
			return "", errors.New("wrong Storage type type")
		}
	case MediaSpecialType:
		switch v {
		case MediaSpecial_none:
			strVal = "none"
		case MediaSpecial_emptydrive:
			strVal = "emptydrive"
		case MediaSpecial_additions:
			strVal = "additions"
		default:
			return "", errors.New("wrong Media special type")
		}
	case MediaType:
		switch v {
		case Media_disk:
			strVal = "disk"
		case Media_dvd:
			strVal = "dvd"
		case Media_floppy:
			strVal = "floppy"
		default:
			return "", errors.New("wrong Media type")
		}

	case MediaFormatType:
		switch v {
		case MediaFormat_vdi:
			strVal = "VDI"
		case MediaFormat_vmdk:
			strVal = "VMDK"
		case MediaFormat_vhd:
			strVal = "VHD"
		default:
			return "", errors.New("wrong Media format type")
		}
	case OvaFormatType:
		switch v {
		case OvaFormat_legacy:
			strVal = "legacy09"
		case OvaFormat_0_9:
			strVal = "ovf09"
		case OvaFormat_1_0:
			strVal = "ovf10"
		case OvaFormat_2_0:
			strVal = "ovf20"
		default:
			return "", errors.New("wrong OVA format type")
		}
	case MacExportType:
		switch v {
		case MacExport_all:
			strVal = ""
		case MacExport_nomacsbutnat:
			strVal = "nomacsbutnat"
		case MacExport_nomacs:
			strVal = "nomacs"
		default:
			return "", errors.New("wrong MAC export format type")
		}

	case MacImportType:
		switch v {
		case MacImport_all:
			strVal = "keepallmacs"
		case MacImport_natmacs:
			strVal = "keepnatmacs"
		default:
			return "", errors.New("wrong MAC import format type")
		}
	default:
		return "", errors.New("wrong value type")
	}
	return strVal, nil
}

func argPreProcess(cmd string, options []any) ([]string, error) {
	opStr := make([]string, 0, len(options)+1)
	if cmd != "" {
		opStr = append(opStr, cmd)
	}

	for _, value := range options {
		strVal, err := argTranslate(value)
		if err != nil {
			return nil, err
		}
		if strVal != "" {
			opStr = append(opStr, strVal)
		}
	}
	return opStr, nil
}
