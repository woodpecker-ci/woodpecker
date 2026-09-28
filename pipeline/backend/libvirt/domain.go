// Copyright 2026 Julian Ospald
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package libvirt

import (
	"context"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/cenkalti/backoff/v7"
	"github.com/rs/zerolog/log"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	virt "libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
)

func (e *libvirt) LoadDomain(ctx context.Context, image string, env map[string]string, persistent bool, sharedDisk SharedDisk, taskUUID string, stepUUID string) (outDomain *virt.Domain, newName string, guestOS string, uuid string, err error) {
	libvirtImgDir := e.c.String("backend-libvirt-image-dir")
	domain, err := e.conn.LookupDomainByName(image)

	if err != nil {
		return nil, "", "", "", err
	}

	defer domain.Free()

	guestOS, err = getGuestOS(domain)
	if err != nil {
		return nil, "", "", "", err
	}

	// get the XML from the loaded domain
	domXml, err := domain.GetXMLDesc(virt.DOMAIN_XML_INACTIVE)
	if err != nil {
		return nil, "", "", "", err
	}

	domcfg := &libvirtxml.Domain{}
	{
		err := domcfg.Unmarshal(domXml)
		if err != nil {
			return nil, "", "", "", err
		}
	}
	newName = fmt.Sprintf("%s-%s", image, stepUUID)
	domcfg.Name = newName
	domcfg.UUID = ""

	if domcfg.Devices == nil {
		return nil, "", "", "", fmt.Errorf("No devices in domain")
	}

	if !persistent {
		log.Debug().Msgf("Setting up ephemeral disks")
		// replacing all the disks with temporary ones
		for ix, disk := range domcfg.Devices.Disks {
			if disk.Target == nil {
				continue
			}
			// get target to disambiguate
			dev := disk.Target.Dev

			if disk.Source == nil || disk.Source.File == nil {
				continue
			}
			// get source to disambiguate
			file := disk.Source.File.File

			baseImgName := fmt.Sprintf(baseImagePattern, dev, taskUUID, stepUUID, filepath.Ext(file))
			newImg, err := e.FromBaseImage(ctx, file, baseImgName, false)
			if err != nil {
				return nil, "", "", "", err
			}

			disk.Source.File.File = newImg
			domcfg.Devices.Disks[ix].Source.File.File = newImg
		}

		// replacing nvram, if any
		if domcfg.OS != nil && domcfg.OS.NVRam != nil {
			nvramDest := filepath.Join(libvirtImgDir, fmt.Sprintf(nvramImagePattern, taskUUID, stepUUID))
			err := CopyFile(domcfg.OS.NVRam.NVRam, nvramDest, false)
			if err != nil {
				return nil, "", "", "", err
			}
			domcfg.OS.NVRam.NVRam = nvramDest
		}

		// generate random network macs
		for ix, interf := range domcfg.Devices.Interfaces {
			if interf.MAC != nil {
				rmac, err := GenerateRandomMACAddress()
				if err != nil {
					return nil, "", "", "", err
				}
				domcfg.Devices.Interfaces[ix].MAC.Address = rmac
			}
		}
	}

	domainType, err := GetDomainType(ctx, domain)
	if err != nil {
		return nil, "", "", "", err
	}

	var disk string
	var diskUuid string

	// insert the shared disk if any
	if sharedDisk.Disk != "" {
		// absolute paths will be ignored
		_, diskFile := filepath.Split(sharedDisk.Disk)
		diskPath := filepath.Join(libvirtImgDir, diskFile)

		disk, err = e.FromBaseImage(ctx, diskPath, fmt.Sprintf(sharedDiskPattern, taskUUID, filepath.Ext(sharedDisk.Disk)), false)
		if err != nil {
			return nil, "", "", "", err
		}

		log.Debug().Msgf("Inserting shared disk: %s", sharedDisk.Disk)

		if sharedDisk.UUID != "" {
			diskUuid = sharedDisk.UUID
		}
	} else { // in absence of a config, create a disk from scratch (ntfs for windows, ext4 otherwise)

		diskSize, ok := env["LIBVIRT_DISK_SIZE"]
		if !ok {
			diskSize = "10G"
		}

		disk, diskUuid, err = e.CreateSharedDisk(ctx, guestOS, domainType, diskSize, taskUUID)
		if err != nil {
			return nil, "", "", "", err
		}
	}

	// 20 is the max amount of chars allowed for the serial number
	// on windows, so we just limit to that
	serial := stepUUID[:20]

	// now cook up the domain config for the shared disk
	var domainDisk libvirtxml.DomainDisk
	zero := uint(0)
	{
		// It's better to attach the shared disk to an existing SATA controller
		// otherwise the topology might change and cause boot loader issues (e.g. on mac).
		// If we can't find one, we just omit it and libvirt will create one.
		var firstSata *libvirtxml.DomainController
		// get the first sata controller with the lowest index
		for _, controller := range domcfg.Devices.Controllers {
			if controller.Type == "sata" {
				if firstSata == nil {
					firstSata = &controller
				} else if *firstSata.Index > *controller.Index {
					firstSata = &controller
				}
			}
		}
		// find a free unit
		var units []uint
		for _, disk := range domcfg.Devices.Disks {
			if disk.Address != nil && disk.Address.Drive != nil && disk.Address.Drive.Unit != nil {
				if firstSata != nil && *disk.Address.Drive.Controller == *firstSata.Index {
					units = append(units, *disk.Address.Drive.Unit)
				}
			}
		}
		var freeUnit uint
		if len(units) > 0 {
			freeUnit = slices.Max(units) + 1
		}

		if domainType == "kvm" || domainType == "qemu" {
			domainDisk = libvirtxml.DomainDisk{
				XMLName: xml.Name{Local: "disk"},
				Device:  "disk",
				Driver: &libvirtxml.DomainDiskDriver{
					Name: "qemu",
					Type: "qcow2",
				},
				Source: &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: disk,
					},
				},
				Target: &libvirtxml.DomainDiskTarget{
					Dev: "sdz",
					Bus: "sata",
				},
				Serial: serial,
			}
		} else {
			domainDisk = libvirtxml.DomainDisk{
				XMLName: xml.Name{Local: "disk"},
				Device:  "disk",
				Driver: &libvirtxml.DomainDiskDriver{
					Name: "file",
					Type: "raw",
				},
				Source: &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: disk,
					},
				},
				Target: &libvirtxml.DomainDiskTarget{
					Dev: "sdz",
					Bus: "sata",
				},
				Serial: serial,
			}
		}

		if firstSata != nil {
			domainDisk.Address = &libvirtxml.DomainAddress{
				Drive: &libvirtxml.DomainAddressDrive{
					Controller: firstSata.Index,
					Bus:        &zero,
					Target:     &zero,
					Unit:       &freeUnit,
				},
			}
		}
	}

	// on windows we mount via the serial, which is stepUUID
	if guestOS == "windows" {
		uuid = serial
	} else {
		// if we don't have a disk UUID by now, get it explicitly
		if diskUuid == "" {
			diskUuid, err = GetDiskUUID(disk, domainType)
			if err != nil {
				return nil, "", "", "", err
			}
		}

		// on unix we mount via the disk uuid, which we must discover
		uuid = diskUuid
	}

	// insert
	domcfg.Devices.Disks = append(domcfg.Devices.Disks, domainDisk)

	newXMLdoc, err := domcfg.Marshal()
	if err != nil {
		return nil, "", "", "", err
	}

	domain, err = e.conn.DomainCreateXML(newXMLdoc, virt.DOMAIN_NONE)

	return domain, newName, uuid, guestOS, err
}

// Get the ip address we can SSH connect to.
func GetDomainIP(ctx context.Context, domain *virt.Domain, interfaceName string) (*virt.DomainIPAddress, error) {
	var interfaces []virt.DomainInterface
	var err error

	// DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT requires qemu-guest-agent
	interfaces, err = domain.ListAllInterfaceAddresses(virt.DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT)
	if err != nil {
		log.Debug().Msgf("Failed to retrieve network interfaces via qemu-agent: %s", err.Error())
		// use lease as fallback
		interfaces, err = domain.ListAllInterfaceAddresses(virt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
		if err != nil {
			return nil, err
		}
	}
	for _, interFace := range interfaces {
		log.Debug().Msgf("Checking interface %s ", interFace.Name)
		if interFace.Name == interfaceName {
			for _, address := range interFace.Addrs {
				log.Debug().Msgf("Checking address %s ", address.Addr)
				if address.Type == virt.IP_ADDR_TYPE_IPV4 {
					return &address, nil
				}
			}
		}
	}

	return nil, &pipeline_errors.PipelineError{
		Message: fmt.Sprintf("Could not find an IPv4 address for the interface %s to connect via ssh to", interfaceName),
		Type:    pipeline_errors.PipelineErrorTypeGeneric,
	}
}

func ShutdownVM(ctx context.Context, domain *virt.Domain) error {
	{
		err := domain.Shutdown()
		if err != nil {
			return err
		}

	}

	// wait 2 minutes for clean shutdown
	// otherwise force via 'Destroy()'
	maxBackOff, _ := time.ParseDuration("45s")
	log.Debug().Msgf("Checking if domain has shutdown for %.1f minutes", maxBackOff.Minutes())
	_, err := backoff.Retry(ctx, func() (any, error) {
		info, err := domain.GetInfo()
		if err != nil {
			// sometimes we race and the domain is shut down already
			if lverr, ok := err.(virt.Error); ok && lverr.Code == virt.ERR_NO_DOMAIN {
				return nil, nil
			} else {
				return nil, err
			}
		}
		switch info.State {
		case virt.DOMAIN_RUNNING:
			return nil, fmt.Errorf("VM still running, retrying")
		case virt.DOMAIN_SHUTDOWN:
			return nil, fmt.Errorf("VM is being shutdown, retrying")
		case virt.DOMAIN_SHUTOFF:
			return nil, nil
		default:
			return nil, backoff.Permanent(fmt.Errorf("Unexpected state: %d", info.State))
		}
	}, backoff.WithMaxElapsedTime(maxBackOff))

	if err != nil {
		log.Debug().Msgf("Could not shut down domain gracefully... destroying cold. Error was: %s", err.Error())
		err := domain.Destroy()
		if err != nil {
			return err
		}
	}

	domain.Free()

	return nil
}

func getGuestOS(domain *virt.Domain) (string, error) {
	// does not work on all hypervisors
	guestInfo, err := domain.GetGuestInfo(virt.DOMAIN_GUEST_INFO_OS, 0)
	if err != nil {
		// read the XML metadata, which sometimes carries libosinfo stuff
		domXml, err := domain.GetXMLDesc(virt.DOMAIN_XML_INACTIVE)
		if err != nil {
			return "", err
		}

		// parse the XML into etree
		doc := etree.NewDocument()
		if err := doc.ReadFromString(domXml); err != nil {
			return "", err
		}

		// TODO: there should be a library parsing this properly
		el := doc.FindElement("/domain/metadata/libosinfo:libosinfo/libosinfo:os")
		if el == nil {
			return "", fmt.Errorf("Could not determine guest platform, add libosinfo metadata to the domain")
		}
		fileAttr := el.SelectAttr("id")
		if fileAttr == nil {
			return "", fmt.Errorf("Could not determine guest platform, add libosinfo metadata to the domain")
		}
		switch {
		case strings.Contains(fileAttr.Value, "/win/"):
			return "windows", nil
		case strings.Contains(fileAttr.Value, "linux"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "ubuntu"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "suse"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "slackware"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "fedora"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "redhat"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "popos"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "oracle"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "nixos"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "guix"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "mandriva"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "mageia"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "gentoo"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "debian"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "centos"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "apple"):
			return "darwin", nil
		case strings.Contains(fileAttr.Value, "openbsd"):
			return "openbsd", nil
		case strings.Contains(fileAttr.Value, "freebsd"):
			return "freebsd", nil
		case strings.Contains(fileAttr.Value, "netbsd"):
			return "netbsd", nil
		case strings.Contains(fileAttr.Value, "dragonflybsd"):
			return "dragonflybsd", nil
		default:
			return "", fmt.Errorf("Could not determine guest platform, add libosinfo metadata to the domain")
		}

	} else {

		// TODO: what does 'ID' return?
		return guestInfo.OS.ID, nil
	}
}

func GetDomainType(ctx context.Context, domain *virt.Domain) (string, error) {
	domXml, err := domain.GetXMLDesc(virt.DOMAIN_XML_INACTIVE)
	if err != nil {
		return "", err
	}
	domcfg := &libvirtxml.Domain{}
	{
		err := domcfg.Unmarshal(domXml)
		if err != nil {
			return "", err
		}
	}

	return domcfg.Type, nil
}
