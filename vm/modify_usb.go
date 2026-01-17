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

package vm

import (
	"errors"
	"strconv"
)

func (m *VMachine) SetUsb(v *VmServer, usb UsbType, callBack func(uuid string)) error {
	maj, _, _ := v.getVmVersion()
	if maj == 6 {
		switch usb {
		case Usb_none:
			return m.setPropertyInternal(&v.Client, []string{"modifyvm", m.UUID, "--usbohci=off", "--usbehci=off", "--usbxhci=off"}, true, callBack)
		case Usb_1:
			return m.setPropertyInternal(&v.Client, []string{"modifyvm", m.UUID, "--usbohci=on", "--usbehci=off", "--usbxhci=off"}, true, callBack)
		case Usb_2:
			return m.setPropertyInternal(&v.Client, []string{"modifyvm", m.UUID, "--usbohci=off", "--usbehci=on", "--usbxhci=off"}, true, callBack)
		case Usb_3:
			return m.setPropertyInternal(&v.Client, []string{"modifyvm", m.UUID, "--usbohci=off", "--usbehci=off", "--usbxhci=on"}, true, callBack)
		default:
			return errors.New("unknown usbtype")
		}
	} else {
		switch usb {
		case Usb_none:
			return m.setPropertyInternal(&v.Client, []string{"modifyvm", m.UUID, "--usb-ohci=off", "--usb-ehci=off", "--usb-xhci=off"}, true, callBack)
		case Usb_1:
			return m.setPropertyInternal(&v.Client, []string{"modifyvm", m.UUID, "--usb-ohci=on", "--usb-ehci=off", "--usb-xhci=off"}, true, callBack)
		case Usb_2:
			return m.setPropertyInternal(&v.Client, []string{"modifyvm", m.UUID, "--usb-ohci=off", "--usb-ehci=on", "--usb-xhci=off"}, true, callBack)
		case Usb_3:
			return m.setPropertyInternal(&v.Client, []string{"modifyvm", m.UUID, "--usb-ohci=off", "--usb-ehci=off", "--usb-xhci=on"}, true, callBack)
		default:
			return errors.New("unknown usbtype")
		}
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
	options := []string{"usbfilter", "add", strconv.Itoa(index), "--target", m.UUID, "--name", client.quoteArgString(name)}
	if vendorId != "" {
		options = append(options, "--vendorid", client.quoteArgString(vendorId))
	}
	if productId != "" {
		options = append(options, "--productid", client.quoteArgString(productId))
	}
	if serialNumber != "" {
		options = append(options, "--serialnumber", client.quoteArgString(serialNumber))
	}
	if product != "" {
		options = append(options, "--product", client.quoteArgString(product))
	}
	if manufacturer != "" {
		options = append(options, "--manufacturer", client.quoteArgString(manufacturer))
	}
	options = append(options, "--active", client.quoteArgString(getYesNoFromBool(active)))
	return m.setPropertyInternal(client, options, true, callBack)
}

func (m *VMachine) ModifyUsbFilter(client *VmSshClient, index int, name, vendorId, productId, serialNumber, product, manufacturer string, active bool, callBack func(uuid string)) error {
	options := []string{"usbfilter", "modify", strconv.Itoa(index), "--target", m.UUID}
	if name != "" {
		options = append(options, "--name", client.quoteArgString(name))
	}
	if vendorId != "" {
		options = append(options, "--vendorid", client.quoteArgString(vendorId))
	}
	if productId != "" {
		options = append(options, "--productid", client.quoteArgString(productId))
	}
	if serialNumber != "" {
		options = append(options, "--serialnumber", client.quoteArgString(serialNumber))
	}
	if product != "" {
		options = append(options, "--product", client.quoteArgString(product))
	}
	if manufacturer != "" {
		options = append(options, "--manufacturer", client.quoteArgString(manufacturer))
	}
	options = append(options, "--active", client.quoteArgString(getYesNoFromBool(active)))
	return m.setPropertyInternal(client, options, true, callBack)
}

func (m *VMachine) RemoveUsbFilter(client *VmSshClient, index int, callBack func(uuid string)) error {
	return m.setPropertyInternal(client, []string{"usbfilter", "remove", strconv.Itoa(index), "--target", m.UUID}, true, callBack)
}
