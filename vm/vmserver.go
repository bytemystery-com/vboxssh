package vm

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"bytemystery-com/vboxssh/server"

	"github.com/google/uuid"
)

var (
	regexNicName = regexp.MustCompile(`^Name:\s*(.*)`)

	regexUsbUUID         = regexp.MustCompile(`^UUID:\s*([0-9a-fA-F-]*)`)
	regexUsbProduct      = regexp.MustCompile(`^Product:\s*(.*)`)
	regexUsbManufacturer = regexp.MustCompile(`^Manufacturer:\s*(.*)`)
	regexUsbProductId    = regexp.MustCompile(`^ProductId:\s*.*\s+\((.*)\)`)
	regexUsbVendorId     = regexp.MustCompile(`^VendorId:\s*.*\s+\((.*)\)`)
	regexUsbPort         = regexp.MustCompile(`^Port:\s*([0-9]+)`)
	regexUsbSerialNumber = regexp.MustCompile(`^SerialNumber:\s*(.*)`)

	regexDvdUUID  = regexp.MustCompile(`^UUID:\s*([0-9a-fA-F-]*)`)
	regexDvdState = regexp.MustCompile(`^State:\s*(.*)`)
	regexDvdUsed1 = regexp.MustCompile(`^In use by VMs:\s*(.*)`)
	// THIK4_BASE (UUID: baaeda96-ad1f-42c2-a5d2-dbef45dbf243)
	regexDvdUsed2    = regexp.MustCompile(`^\s*.*\s*\(UUID:\s*([0-9a-fA-F-]*)\)`)
	regexDvdLocation = regexp.MustCompile(`^Location:\s*(.*)`)

	regexFloppyUUID  = regexp.MustCompile(`^UUID:\s*([0-9a-fA-F-]*)`)
	regexFloppyState = regexp.MustCompile(`^State:\s*(.*)`)
	regexFloppyUsed1 = regexp.MustCompile(`^In use by VMs:\s*(.*)`)
	// THIK4_BASE (UUID: baaeda96-ad1f-42c2-a5d2-dbef45dbf243)
	regexFloppyUsed2    = regexp.MustCompile(`^\s*.*\s*\(UUID:\s*([0-9a-fA-F-]*)\)`)
	regexFloppyLocation = regexp.MustCompile(`^Location:\s*(.*)`)

	regexStartWithSpace = regexp.MustCompile(`^\s+`)
	regexHddUUID        = regexp.MustCompile(`^UUID:\s*([0-9a-fA-F-]+)`)
	regexHddState       = regexp.MustCompile(`^State:\s*(.*)`)
	regexHddUsed1       = regexp.MustCompile(`^In use by VMs:\s*(.*)`)
	// THIK4_BASE (UUID: baaeda96-ad1f-42c2-a5d2-dbef45dbf243)
	regexHddUsed2 = regexp.MustCompile(`^*.?\s*\(UUID:\s*([0-9a-fA-F-]+)\)\s*(.*)`)
	// [vor rus. Lexikon (UUID: a9907a05-4454-4084-a3c5-b5ed5eb3e458)]
	regexHddUsed3      = regexp.MustCompile(`\[(.*)\s*\(UUID:\s*([0-9a-fA-F-]+)\)\]`)
	regexHddLocation   = regexp.MustCompile(`^Location:\s*(.*)`)
	regexHddParentUUID = regexp.MustCompile(`^Parent UUID:\s*(.*)`)

	regexOsId                = regexp.MustCompile(`^ID:\s*(.*)`)
	regexOsDescription       = regexp.MustCompile(`^Description:\s*(.*)`)
	regexOsFamilyId          = regexp.MustCompile(`^Family ID:\s*(.*)`)
	regexOsFamilyDescription = regexp.MustCompile(`^Family Desc:\s*(.*)`)
	regexOsArchitecture      = regexp.MustCompile(`^Architecture:\s*(.*)`)
	regexOsSubType           = regexp.MustCompile(`^OS Subtype:\s*(.*)`)
	regexOs64Bit             = regexp.MustCompile(`^64 bit:\s*(.*)`)

	regexSystemProperties = regexp.MustCompile(`^(.*):\s*(.*)`)
	regexHostInfos        = regexp.MustCompile(`^(.*):\s*(.*)`)

	// Pack no. 0
	regexExtPackNr          = regexp.MustCompile(`^Pack no\.\s+[0-9]+:\s+(.*)`)
	regexExtPackVersion     = regexp.MustCompile(`^Version:\s+(.*)`)
	regexExtPackRevision    = regexp.MustCompile(`^Revision:\s+(.*)`)
	regexExtPackUsable      = regexp.MustCompile(`^Usable:\s+(.*)`)
	regexExtPackWhyUnUsable = regexp.MustCompile(`^Why unusable:\s+(.*)`)

	regexImportDryRunVsys = regexp.MustCompile(`^Virtual system\s+([0-9]+):`)
)

type VmServer struct {
	server.Server
	UUID             string            `json:"server"`
	Client           VmSshClient       `json:"-"`
	Version          string            `json:"-"`
	OsTypes          []*OsType         `json:"-"`
	SystemProperties map[string]string `json:"-"`
	HostInfos        map[string]string `json:"-"`

	BridgeAdapter      []NicAdapter `json:"-"`
	HostOnlyAdapter    []NicAdapter `json:"-"`
	InternalNetAdapter []NicAdapter `json:"-"`
	NatNetAdapter      []NicAdapter `json:"-"`
	CloudAdapter       []NicAdapter `json:"-"`
	UsbDevices         []UsbDevice  `json:"-"`

	FloppyImagesPath string `json:"imgpath"`
	DvdImagesPath    string `json:"isopath"`
	HddImagesPath    string `json:"vdipath"`
	OvaPath          string `json:"ovapath"`
}

func NewVmServer(s server.Server) VmServer {
	v := VmServer{
		Server:           s,
		UUID:             uuid.NewString(),
		SystemProperties: make(map[string]string, 120),
	}
	v.Client.IsLocal = v.IsLocal()
	return v
}

func (v *VmServer) Connect(fOk func(), fErr func(error)) error {
	if v.IsLocal() {
		version, err := v.GetVersion()
		if err != nil {
			if fErr != nil {
				fErr(err)
			}
			return err
		}
		v.Version = version
		if fOk != nil {
			fOk()
		}
		return nil
	}
	client, err := v.Server.Connect()
	if err != nil {
		if fErr != nil {
			fErr(err)
		}
		return err
	}

	v.Client.Client = client
	version, err := v.GetVersion()
	if err != nil {
		if fErr != nil {
			fErr(err)
		}
		return err
	}
	v.Version = version
	if fOk != nil {
		fOk()
	}
	return nil
}

func (v *VmServer) IsLocal() bool {
	if v.Server.Host == "" || v.Server.Port == 0 {
		return true
	}
	return false
}

func (v *VmServer) IsConnected() bool {
	if v.IsLocal() {
		return true
	}
	if v.Client.Client != nil {
		return true
	}
	return false
}

// Version
func (s *VmServer) GetVersion() (string, error) {
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"--version"}, nil, nil)
	if err != nil {
		return "", err
	}
	if len(lines) == 2 {
		return lines[0], nil
	}
	return "", errors.New("get version failed")
}

// NICs
func (s *VmServer) UpdateBridgeAdapters() error {
	adapters, err := s.GetBridgeAdapters(true)
	if err != nil {
		return nil
	}
	s.BridgeAdapter = adapters
	return nil
}

func (s *VmServer) GetBridgeAdapters(update bool) ([]NicAdapter, error) {
	if !update && len(s.BridgeAdapter) > 0 {
		return s.BridgeAdapter, nil
	}
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "bridgedifs"}, nil, nil)
	if err != nil {
		return nil, err
	}
	adapters := getAdapters(regexNicName, lines)
	s.BridgeAdapter = adapters
	return s.BridgeAdapter, nil
}

func (s *VmServer) UpdateHostAdapters() error {
	adapters, err := s.GetHostAdapters(true)
	if err != nil {
		return nil
	}
	s.HostOnlyAdapter = adapters
	return nil
}

func (s *VmServer) GetHostAdapters(update bool) ([]NicAdapter, error) {
	if !update && len(s.HostOnlyAdapter) > 0 {
		return s.HostOnlyAdapter, nil
	}
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "hostonlyifs"}, nil, nil)
	if err != nil {
		return nil, err
	}
	adapters := getAdapters(regexNicName, lines)
	s.HostOnlyAdapter = adapters
	return s.HostOnlyAdapter, nil
}

func (s *VmServer) UpdateInternalAdapters() error {
	adapters, err := s.GetInternalAdapters(true)
	if err != nil {
		return nil
	}
	s.InternalNetAdapter = adapters
	return nil
}

func (s *VmServer) GetInternalAdapters(update bool) ([]NicAdapter, error) {
	if !update && len(s.InternalNetAdapter) > 0 {
		return s.InternalNetAdapter, nil
	}
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "intnets"}, nil, nil)
	if err != nil {
		return nil, err
	}
	adapters := getAdapters(regexNicName, lines)

	s.InternalNetAdapter = adapters
	return s.InternalNetAdapter, nil
}

func (s *VmServer) UpdateNatAdapters() error {
	adapters, err := s.GetNatAdapters(true)
	if err != nil {
		return nil
	}
	s.InternalNetAdapter = adapters
	return nil
}

func (s *VmServer) GetNatAdapters(update bool) ([]NicAdapter, error) {
	if !update && len(s.NatNetAdapter) > 0 {
		return s.NatNetAdapter, nil
	}
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "natnets"}, nil, nil)
	if err != nil {
		return nil, err
	}
	adapters := getAdapters(regexNicName, lines)
	s.NatNetAdapter = adapters
	return s.NatNetAdapter, nil
}

func (s *VmServer) GetCloudAdapters(update bool) ([]NicAdapter, error) {
	if !update && len(s.CloudAdapter) > 0 {
		return s.CloudAdapter, nil
	}
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "cloudnets"}, nil, nil)
	if err != nil {
		return nil, err
	}
	adapters := getAdapters(regexNicName, lines)
	s.CloudAdapter = adapters
	return s.CloudAdapter, nil
}

func (s *VmServer) UpdateAllNetAdapters() error {
	var errs error
	err := s.UpdateBridgeAdapters()
	if err != nil {
		errs = errors.Join(errs, err)
	}
	err = s.UpdateHostAdapters()
	if err != nil {
		errs = errors.Join(errs, err)
	}
	err = s.UpdateInternalAdapters()
	if err != nil {
		errs = errors.Join(errs, err)
	}
	err = s.UpdateNatAdapters()
	if err != nil {
		errs = errors.Join(errs, err)
	}
	return errs
}

func getAdapters(reg *regexp.Regexp, lines []string) []NicAdapter {
	adapters := []NicAdapter{}
	for _, line := range lines {
		items := reg.FindStringSubmatch(line)
		if len(items) == 2 {
			adapters = append(adapters, NicAdapter{
				Name: items[1],
			})
		}
	}
	return adapters
}

// USB
func (s *VmServer) UpdateUsbDevices() error {
	usbDevices, err := s.GetUsbDevices()
	if err != nil {
		return nil
	}
	s.UsbDevices = usbDevices
	return nil
}

func (s *VmServer) GetUsbDevices() ([]UsbDevice, error) {
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "usbhost"}, nil, nil)
	if err != nil {
		return nil, err
	}
	usb := []UsbDevice{}
	usbDevice := UsbDevice{}
	for _, line := range lines {
		if line == "" {
			if usbDevice.UUID != "" {
				usb = append(usb, usbDevice)
			}
			usbDevice = UsbDevice{}
			continue
		}
		items := regexUsbUUID.FindStringSubmatch(line)
		if len(items) == 2 {
			usb = append(usb, UsbDevice{
				UUID: items[1],
			})
			continue
		}
		items = regexUsbManufacturer.FindStringSubmatch(line)
		if len(items) == 2 {
			usb[len(usb)-1].Manufacturer = items[1]
			if len(usb[len(usb)-1].Name) > 0 {
				usb[len(usb)-1].Name = items[1] + "-" + usb[len(usb)-1].Name
			} else {
				usb[len(usb)-1].Name = items[1]
			}
			continue
		}

		items = regexUsbProduct.FindStringSubmatch(line)
		if len(items) == 2 {
			usb[len(usb)-1].Product = items[1]
			if len(usb[len(usb)-1].Name) > 0 {
				usb[len(usb)-1].Name += "-" + items[1]
			} else {
				usb[len(usb)-1].Name = items[1]
			}
			continue
		}
		// ProductId
		items = regexUsbProductId.FindStringSubmatch(line)
		if len(items) == 2 {
			usb[len(usb)-1].ProductId = items[1]
			continue
		}
		// VendorId
		items = regexUsbVendorId.FindStringSubmatch(line)
		if len(items) == 2 {
			usb[len(usb)-1].VendorId = items[1]
			continue
		}
		// Port
		items = regexUsbPort.FindStringSubmatch(line)
		if len(items) == 2 {
			p, err := strconv.Atoi(items[1])
			if err == nil {
				usb[len(usb)-1].Port = p
			}
			continue
		}
		// SerialNumber
		items = regexUsbSerialNumber.FindStringSubmatch(line)
		if len(items) == 2 {
			usb[len(usb)-1].SerialNumber = items[1]
			continue
		}
	}
	if usbDevice.UUID != "" {
		usb = append(usb, usbDevice)
	}
	return usb, nil
}

// Media - DVD
func (s *VmServer) GetDvdMedias() ([]DvdInfo, error) {
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "--long", "dvds"}, nil, nil)
	if err != nil {
		return nil, err
	}
	dvds := []DvdInfo{}
	dvd := DvdInfo{}
	inUsed := false
	for _, line := range lines {
		if line == "" {
			if dvd.MediaInfo.UUID != "" {
				dvds = append(dvds, dvd)
			}
			dvd = DvdInfo{}
			inUsed = false
			continue
		}
		if inUsed {
			items := regexDvdUsed2.FindStringSubmatch(line)
			if len(items) == 2 {
				dvd.UsedBy = append(dvd.UsedBy, items[1])
				continue
			}
		} else {
			items := regexDvdUUID.FindStringSubmatch(line)
			if len(items) == 2 {
				dvd = DvdInfo{
					MediaInfo{
						UUID: items[1],
					},
				}
				continue
			}
			items = regexDvdState.FindStringSubmatch(line)
			if len(items) == 2 {
				if items[1] == "created" {
					dvd.State = MediaState_created
				}
				continue
			}
			items = regexDvdUsed1.FindStringSubmatch(line)
			if len(items) == 2 {
				items = regexDvdUsed2.FindStringSubmatch(items[1])
				if len(items) == 2 {
					dvd.UsedBy = append(dvd.UsedBy, items[1])
					inUsed = true
				}
				continue
			}
			items = regexDvdLocation.FindStringSubmatch(line)
			if len(items) == 2 {
				dvd.Location = items[1]
				continue
			}
		}
	}
	if dvd.UUID != "" {
		dvds = append(dvds, dvd)
	}
	slices.SortFunc(dvds, func(a, b DvdInfo) int {
		A := path.Base(a.Location)
		B := path.Base(b.Location)
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
	return dvds, nil
}

// Media - Floppies
func (s *VmServer) GetFloppyMedias() ([]FloppyInfo, error) {
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "--long", "floppies"}, nil, nil)
	if err != nil {
		return nil, err
	}
	floppies := []FloppyInfo{}
	floppy := FloppyInfo{}
	inUsed := false
	for _, line := range lines {
		if line == "" {
			if floppy.MediaInfo.UUID != "" {
				floppies = append(floppies, floppy)
			}
			floppy = FloppyInfo{}
			inUsed = false
			continue
		}
		if inUsed {
			items := regexFloppyUsed2.FindStringSubmatch(line)
			if len(items) == 2 {
				floppy.UsedBy = append(floppy.UsedBy, items[1])
				continue
			}
		} else {
			items := regexFloppyUUID.FindStringSubmatch(line)
			if len(items) == 2 {
				floppy = FloppyInfo{
					MediaInfo{
						UUID: items[1],
					},
				}
				continue
			}
			items = regexFloppyState.FindStringSubmatch(line)
			if len(items) == 2 {
				if items[1] == "created" {
					floppy.State = MediaState_created
				}
				continue
			}
			items = regexFloppyUsed1.FindStringSubmatch(line)
			if len(items) == 2 {
				items = regexFloppyUsed2.FindStringSubmatch(items[1])
				if len(items) == 2 {
					floppy.UsedBy = append(floppy.UsedBy, items[1])
					inUsed = true
				}
				continue
			}
			items = regexFloppyLocation.FindStringSubmatch(line)
			if len(items) == 2 {
				floppy.Location = items[1]
				continue
			}
		}
	}
	if floppy.UUID != "" {
		floppies = append(floppies, floppy)
	}

	slices.SortFunc(floppies, func(a, b FloppyInfo) int {
		A := path.Base(a.Location)
		B := path.Base(b.Location)
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
	return floppies, nil
}

// Media - Hdds
func (s *VmServer) GetHddMedias() ([]*HddInfo, map[string]*HddInfo, error) {
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "--long", "hdds"}, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	hdds := []*HddInfo{}
	hdd := &HddInfo{}
	inUsed := false
	for _, line := range lines {
		if line == "" {
			if hdd.UUID != "" {
				hdds = append(hdds, hdd)
			}
			hdd = &HddInfo{}
			inUsed = false
			continue
		}
		if inUsed {
			if regexStartWithSpace.MatchString(line) {
				items := regexHddUsed2.FindStringSubmatch(line)
				if len(items) == regexHddUsed2.NumSubexp()+1 {
					u := &UsedByInfo{
						UUID: items[1],
					}
					items2 := regexHddUsed3.FindStringSubmatch(items[2])
					if len(items2) == regexHddUsed3.NumSubexp()+1 {
						u.SnapshotDescription = items2[1]
						u.SnapshotUUID = items2[2]
					}
					hdd.UsedBy = append(hdd.UsedBy, u)
				}
				continue
			} else {
				inUsed = false
			}
		}
		items := regexHddUUID.FindStringSubmatch(line)
		if len(items) == 2 {
			hdd = &HddInfo{
				UUID: items[1],
			}
			continue
		}
		items = regexHddState.FindStringSubmatch(line)
		if len(items) == 2 {
			if items[1] == "created" {
				hdd.State = MediaState_created
			}
			continue
		}
		items = regexHddUsed1.FindStringSubmatch(line)
		if len(items) == regexHddUsed1.NumSubexp()+1 {
			items = regexHddUsed2.FindStringSubmatch(items[1])
			if len(items) == regexHddUsed2.NumSubexp()+1 {
				u := &UsedByInfo{
					UUID: items[1],
				}
				items2 := regexHddUsed3.FindStringSubmatch(items[2])
				if len(items2) == regexHddUsed3.NumSubexp()+1 {
					u.SnapshotDescription = items2[1]
					u.SnapshotUUID = items2[2]
				}
				hdd.UsedBy = append(hdd.UsedBy, u)
				inUsed = true
			}
			continue
		}
		items = regexHddLocation.FindStringSubmatch(line)
		if len(items) == 2 {
			hdd.Location = items[1]
			continue
		}
		items = regexHddParentUUID.FindStringSubmatch(line)
		if len(items) == 2 {
			hdd.Parent = items[1]
			continue
		}
	}
	if hdd.UUID != "" {
		hdds = append(hdds, hdd)
	}

	m := make(map[string]*HddInfo, len(hdds))
	for _, item := range hdds {
		m[item.UUID] = item
	}

	l := make([]*HddInfo, 0, len(hdds))
	for _, item := range hdds {
		if item.Parent == "base" {
			l = append(l, item)
		} else {
			item2, ok := m[item.Parent]
			if ok {
				item2.Childs = append(item2.Childs, item)
			} else {
				fmt.Printf("!!! Something went wrong !!!")
			}
		}
	}

	slices.SortFunc(l, func(a, b *HddInfo) int {
		A := path.Base(a.Location)
		B := path.Base(b.Location)
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

	return l, m, nil
}

// Os
func (s *VmServer) GetOsTypes(update bool) ([]*OsType, error) {
	if !update && len(s.OsTypes) > 0 {
		return s.OsTypes, nil
	}
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "--long", "ostypes"}, nil, nil)
	if err != nil {
		return nil, err
	}
	ostypes := []*OsType{}
	ostype := new(OsType)
	for _, line := range lines {
		if line == "" {
			if ostype.ID != "" {
				ostypes = append(ostypes, ostype)
			}
			ostype = new(OsType)
			continue
		}
		items := regexOsId.FindStringSubmatch(line)
		if len(items) == 2 {
			ostype.ID = items[1]
			continue
		}
		items = regexOsDescription.FindStringSubmatch(line)
		if len(items) == 2 {
			ostype.Name = items[1]
			continue
		}
		items = regexOsFamilyId.FindStringSubmatch(line)
		if len(items) == 2 {
			ostype.FamilyId = items[1]
			continue
		}
		items = regexOsFamilyDescription.FindStringSubmatch(line)
		if len(items) == 2 {
			ostype.Family = items[1]
			continue
		}
		items = regexOsArchitecture.FindStringSubmatch(line)
		if len(items) == 2 {
			ostype.Architecture = items[1]
			continue
		}
		items = regexOsSubType.FindStringSubmatch(line)
		if len(items) == 2 {
			ostype.Subtype = items[1]
			continue
		}
		items = regexOs64Bit.FindStringSubmatch(line)
		if len(items) == 2 {
			ostype.Is64Bit, _ = strconv.ParseBool(items[1])
			continue
		}
	}
	if ostype.ID != "" {
		ostypes = append(ostypes, ostype)
	}
	slices.SortFunc(ostypes, func(a, b *OsType) int {
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

	s.OsTypes = ostypes
	return ostypes, nil
}

func GetOsFamilies(osList []*OsType) ([]*OsFamily, error) {
	fmap := make(map[string]string, 10)
	for _, item := range osList {
		fmap[item.FamilyId] = item.Family
	}
	families := make([]*OsFamily, 0, 10)
	for id, name := range fmap {
		families = append(families, &OsFamily{
			FamilyId: id,
			Family:   name,
		})
	}

	slices.SortFunc(families, func(a, b *OsFamily) int {
		A := a.Family
		B := b.Family
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
	return families, nil
}

func GetOsSubTypes(familyDescription string, osList []*OsType) ([]string, error) {
	stmap := make(map[string]string, 30)
	for _, item := range osList {
		if item.Family == familyDescription && item.Subtype != "" {
			stmap[item.Subtype] = item.Subtype
		}
	}
	subTypes := slices.Collect(maps.Keys(stmap))
	slices.SortFunc(subTypes, func(a, b string) int {
		A := a
		B := b
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
	return subTypes, nil
}

func GetOsVersionTypes(familyDescription, subType string, osList []*OsType) ([]string, error) {
	vermap := make(map[string]string, 50)
	for _, item := range osList {
		if item.Family == familyDescription && (subType == "" || item.Subtype == subType) {
			vermap[item.ID] = item.Name
		}
	}
	versions := slices.Collect(maps.Values(vermap))
	slices.SortFunc(versions, func(a, b string) int {
		A := a
		B := b
		an := strings.ToLower(a)
		bn := strings.ToLower(b)

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
	return versions, nil
}

func (s *VmServer) GetSystemProperties(update bool) (map[string]string, error) {
	if !update && len(s.SystemProperties) > 0 {
		return s.SystemProperties, nil
	}
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "--long", "systemproperties"}, nil, nil)
	if err != nil {
		return nil, err
	}
	sysprop := make(map[string]string, 120)
	for _, line := range lines {
		if line == "" {
			continue
		}
		items := regexSystemProperties.FindStringSubmatch(line)
		if len(items) == 3 {
			sysprop[items[1]] = items[2]
		}
	}
	s.SystemProperties = sysprop
	return s.SystemProperties, nil
}

func (s *VmServer) GetHostInfos(update bool) (map[string]string, error) {
	if !update && len(s.HostInfos) > 0 {
		return s.HostInfos, nil
	}
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "--long", "hostinfo"}, nil, nil)
	if err != nil {
		return nil, err
	}
	hostinfo := make(map[string]string, 120)
	for _, line := range lines {
		if line == "" {
			continue
		}
		items := regexHostInfos.FindStringSubmatch(line)
		if len(items) == 3 {
			hostinfo[items[1]] = items[2]
		}
	}
	s.HostInfos = hostinfo
	return s.HostInfos, nil
}

func (s *VmServer) CreateMedia(client *VmSshClient, media MediaType, size int64, mediaFormat *MediaFormatType, isFixedSize *bool, file string, statusWriter io.Writer) error {
	opt := []any{media}
	if mediaFormat != nil {
		opt = append(opt, "--format")
		opt = append(opt, *mediaFormat)
	}
	opt = append(opt, "--variant")
	if isFixedSize != nil && *isFixedSize {
		opt = append(opt, "Fixed")
	} else {
		opt = append(opt, "Standard")
	}
	opt = append(opt, "--filename="+client.quoteArgString(file))
	opt = append(opt, "--size")
	opt = append(opt, size)

	optS, err := argPreProcess("createmedium", opt)
	if err != nil {
		return nil
	}

	lines, err := RunCmd(client, VBOXMANAGE_APP, optS, nil, statusWriter)
	_ = lines
	return err
}

func (s *VmServer) DeleteMedia(client *VmSshClient, media MediaType, file string) error {
	opt := []any{media, client.quoteArgString(file), "--delete"}
	optS, err := argPreProcess("closemedium", opt)
	if err != nil {
		return nil
	}
	lines, err := RunCmd(client, VBOXMANAGE_APP, optS, nil, nil)
	_ = lines
	return err
}

func (s *VmServer) ExportOva(client *VmSshClient, machines []string, format OvaFormatType, manifest bool, iso bool, macMode MacExportType, vsys []string, file string, statusWriter io.Writer) error {
	opt := []string{"export"}
	opt = append(opt, machines...)

	fStr, err := argTranslate(format)
	if err != nil {
		return nil
	}
	macStr, err := argTranslate(macMode)

	opStr := ""
	if manifest {
		opStr += "manifest"
	}
	if iso {
		if opStr != "" {
			opStr += ","
		}
		opStr += "iso"
	}
	if macStr != "" {
		if opStr != "" {
			opStr += ","
		}
		opStr += macStr
	}
	if opStr != "" {
		opStr = "--options=" + opStr
	}

	opt = append(opt, "--output="+client.quoteArgString(file))
	opt = append(opt, "--"+fStr)

	if opStr != "" {
		opt = append(opt, opStr)
	}

	for _, item := range vsys {
		opt = append(opt, item)
	}

	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, statusWriter)
	_ = lines
	return err
}

func (s *VmServer) ImportOvaDryRun(client *VmSshClient, file string) (int, error) {
	opt := []string{"import"}
	opt = append(opt, "--dry-run")
	opt = append(opt, s.Client.quoteArgString(file))
	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, nil)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, line := range lines {
		items := regexImportDryRunVsys.FindStringSubmatch(line)
		if len(items) == 2 {
			count++
		}
	}
	return count, nil
}

func (s *VmServer) ImportOva(client *VmSshClient, mac MacImportType, isVdi bool, file string, vsys []string, statusWriter io.Writer) error {
	macStr, err := argTranslate(mac)
	if err != nil {
		return err
	}
	opt := []string{"import"}
	opt = append(opt, s.Client.quoteArgString(file))
	opStr := ""
	if isVdi {
		opStr = "importtovdi," + macStr
	} else {
		opStr = macStr
	}
	opStr = "--options=" + opStr

	opt = append(opt, opStr)
	opt = append(opt, vsys...)

	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, statusWriter)
	_ = lines
	return err
}

func (s *VmServer) CreateVm(client *VmSshClient, name string) error {
	opt := []string{"createvm", "--name", s.Client.quoteArgString(name), "--register"}

	lines, err := RunCmd(client, VBOXMANAGE_APP, opt, nil, nil)
	_ = lines
	return err
}

func (s *VmServer) GetExtPackHostInfos() ([]*ExtPackInfoType, error) {
	lines, err := RunCmd(&s.Client, VBOXMANAGE_APP, []string{"list", "--long", "extpacks"}, nil, nil)
	if err != nil {
		return nil, err
	}
	extPackList := make([]*ExtPackInfoType, 0, 1)
	var info ExtPackInfoType
	for _, line := range lines {
		if line == "" {
			continue
		}

		items := regexExtPackNr.FindStringSubmatch(line)
		if len(items) == 2 {
			info = ExtPackInfoType{
				Name: items[1],
			}
			extPackList = append(extPackList, &info)
			continue
		}

		items = regexExtPackVersion.FindStringSubmatch(line)
		if len(items) == 2 {
			info.Version = items[1]
			continue
		}

		items = regexExtPackRevision.FindStringSubmatch(line)
		if len(items) == 2 {
			info.Revision = items[1]
			continue
		}
		items = regexExtPackUsable.FindStringSubmatch(line)
		if len(items) == 2 {
			if strings.ToLower(items[1]) == "true" {
				info.Usable = true
			} else {
				info.Usable = false
			}
			continue
		}
		items = regexExtPackWhyUnUsable.FindStringSubmatch(line)
		if len(items) == 2 {
			info.WhyUnsable = items[1]
			continue
		}
	}
	return extPackList, nil
}
