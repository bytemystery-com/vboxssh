package vm

import (
	"errors"
	"strconv"
)

func (m *VMachine) SetUsb(client *VmSshClient, usb UsbType, callBack func(uuid string)) error {
	switch usb {
	case Usb_none:
		return m.setPropertyInternal(client, []string{"modifyvm", m.UUID, "--usb-ohci=off", "--usb-ehci=off", "--usb-xhci=off"}, true, callBack)
	case Usb_1:
		return m.setPropertyInternal(client, []string{"modifyvm", m.UUID, "--usb-ohci=on", "--usb-ehci=off", "--usb-xhci=off"}, true, callBack)
	case Usb_2:
		return m.setPropertyInternal(client, []string{"modifyvm", m.UUID, "--usb-ohci=off", "--usb-ehci=on", "--usb-xhci=off"}, true, callBack)
	case Usb_3:
		return m.setPropertyInternal(client, []string{"modifyvm", m.UUID, "--usb-ohci=off", "--usb-ehci=off", "--usb-xhci=on"}, true, callBack)
	default:
		return errors.New("unknown usbtype")
	}
}

func (m *VMachine) AttachUsbDevice(client *VmSshClient, uuid string, captureFile string, callBack func(uuid string)) error {
	if len(captureFile) == 0 {
		return m.setPropertyInternal(client, []string{"controlvm", m.UUID, "usbattach", uuid}, true, callBack)
	} else {
		return m.setPropertyInternal(client, []string{"controlvm", m.UUID, "usbattach", uuid, "--capturefile=" + captureFile}, true, callBack)
	}
}

func (m *VMachine) DetachUsbDevice(client *VmSshClient, uuid string, callBack func(uuid string)) error {
	return m.setPropertyInternal(client, []string{"controlvm", m.UUID, "usbdetach", uuid}, true, callBack)
}

func getYesNoFromBool(b bool) string {
	if b {
		return "yes"
	} else {
		return "no"
	}
}

func (m *VMachine) AddUsbFilter(client *VmSshClient, index int, name, vendorId, productId, serialNumber, product, manufacturer string, active bool, callBack func(uuid string)) error {
	options := []string{"usbfilter", "add", strconv.Itoa(index), "--target=" + m.UUID, "--name=" + client.quoteArgString(name)}
	if vendorId != "" {
		options = append(options, "--vendorid="+client.quoteArgString(vendorId))
	}
	if productId != "" {
		options = append(options, "--productid="+client.quoteArgString(productId))
	}
	if serialNumber != "" {
		options = append(options, "--serialnumber="+client.quoteArgString(serialNumber))
	}
	if product != "" {
		options = append(options, "--product="+client.quoteArgString(product))
	}
	if manufacturer != "" {
		options = append(options, "--manufacturer="+client.quoteArgString(manufacturer))
	}
	options = append(options, "--active="+client.quoteArgString(getYesNoFromBool(active)))
	return m.setPropertyInternal(client, options, true, callBack)
}

func (m *VMachine) ModifyUsbFilter(client *VmSshClient, index int, name, vendorId, productId, serialNumber, product, manufacturer string, active bool, callBack func(uuid string)) error {
	options := []string{"usbfilter", "modify", strconv.Itoa(index), "--target=" + m.UUID}
	if name != "" {
		options = append(options, "--name="+client.quoteArgString(name))
	}
	if vendorId != "" {
		options = append(options, "--vendorid="+client.quoteArgString(vendorId))
	}
	if productId != "" {
		options = append(options, "--productid="+client.quoteArgString(productId))
	}
	if serialNumber != "" {
		options = append(options, "--serialnumber="+client.quoteArgString(serialNumber))
	}
	if product != "" {
		options = append(options, "--product="+client.quoteArgString(product))
	}
	if manufacturer != "" {
		options = append(options, "--manufacturer="+client.quoteArgString(manufacturer))
	}
	options = append(options, "--active="+client.quoteArgString(getYesNoFromBool(active)))
	return m.setPropertyInternal(client, options, true, callBack)
}

func (m *VMachine) RemoveUsbFilter(client *VmSshClient, index int, callBack func(uuid string)) error {
	return m.setPropertyInternal(client, []string{"usbfilter", "remove", strconv.Itoa(index), "--target=" + m.UUID}, true, callBack)
}
