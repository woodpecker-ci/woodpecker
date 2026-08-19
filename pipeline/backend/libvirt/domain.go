package libvirt

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/cenkalti/backoff/v7"
	"github.com/rs/zerolog/log"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	virt "libvirt.org/go/libvirt"
)

func (e *libvirt) LoadDomain(ctx context.Context, image string, env map[string]string, ephemeral bool, sharedDisk SharedDisk, taskUUID string, stepUUID string) (outDomain *virt.Domain, guestOS string, uuid string, err error) {
	domain, err := e.conn.LookupDomainByName(image)

	if err != nil {
		return nil, "", "", err
	}

	defer domain.Free()

	guestOS, err = getGuestOS(domain)
	if err != nil {
		return nil, "", "", err
	}

	// get the XML from the loaded domain
	xml, err := domain.GetXMLDesc(virt.DOMAIN_XML_INACTIVE)
	if err != nil {
		return nil, "", "", err
	}
	// parse the XML into etree
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return nil, "", "", err
	}

	if ephemeral {
		log.Debug().Msgf("Setting up ephemeral disks")
		// replacing all the disks with temporary ones
		el := doc.FindElements("/domain/devices/disk[@type='file'][@device='disk']")
		for _, child := range el {
			// get target to disambiguate
			targetEl := child.FindElement("./target[@dev]")
			if targetEl == nil {
				continue
			}
			devAttr := targetEl.SelectAttr("dev")
			if devAttr == nil {
				continue
			}

			// get source to disambiguate
			sourceEl := child.FindElement("./source[@file]")
			if sourceEl == nil {
				continue
			}
			fileAttr := sourceEl.SelectAttr("file")
			if fileAttr == nil {
				continue
			}

			baseImgName := fmt.Sprintf(baseImagePattern, devAttr.Value, taskUUID, stepUUID, filepath.Ext(fileAttr.Value))
			newImg, err := e.FromBaseImage(ctx, fileAttr.Value, baseImgName, false)
			if err != nil {
				return nil, "", "", err
			}
			sourceEl.CreateAttr("file", newImg)
		}
	}

	// get devices element
	devices := doc.FindElement("/domain/devices")
	if devices == nil {
		return nil, "", "", fmt.Errorf("Could not find devices in domain XML")
	}

	// insert the shared disk if any
	if sharedDisk.DiskConfig != "" {
		log.Debug().Msgf("Inserting shared disk: %s", sharedDisk.DiskConfig)

		// generate the XML with the base image
		sharedXml, err := e.EphemeralizeSharedDisk(ctx, &sharedDisk, taskUUID)
		if err != nil {
			return nil, "", "", err
		}
		sharedXmlDoc := etree.NewDocument()
		if err := sharedXmlDoc.ReadFromString(sharedXml); err != nil {
			return nil, "", "", err
		}

		uuid = sharedDisk.UUID

		// insert
		devices.AddChild(sharedXmlDoc.Root())
	} else { // in absence of a config, create a disk from scratch (ntfs for windows, ext4 otherwise)

		domainType, err := GetDomainType(ctx, domain)
		if err != nil {
			return nil, "", "", err
		}

		diskSize, ok := env["LIBVIRT_DISK_SIZE"]
		if !ok {
			diskSize = "10G"
		}

		disk, diskUuid, err := e.CreateSharedDisk(ctx, guestOS, domainType, diskSize, taskUUID)
		if err != nil {
			return nil, "", "", err
		}

		// 20 is the max amount of chars allowed for the serial number
		// on windows, so we just limit to that
		serial := stepUUID[:20]

		// now cook up an XML config
		var newXml string
		if domainType == "kvm" || domainType == "qemu" {
			newXml = fmt.Sprintf(`
			  <disk type='file' device='disk'>
          <driver name='qemu' type='qcow2'/>
          <source file='%s'/>
          <target dev='sdz' bus='sata'/>
          <serial>%s</serial>
        </disk>`, disk, serial)
		} else {
			newXml = fmt.Sprintf(`
        <disk type='file'>
          <driver name='file' type='raw'/>
          <source file='%s'/>
          <target dev='sdz' bus='sata'/>
          <serial>%s</serial>
        </disk>`, disk, serial)
		}

		// on windows we mount via the serial, which is stepUUID
		if guestOS == "windows" {
			uuid = serial
		} else {
			// on unix we mount via the disk uuid, which we must discover
			uuid = diskUuid
		}

		// insert
		sharedXmlDoc := etree.NewDocument()
		if err := sharedXmlDoc.ReadFromString(newXml); err != nil {
			return nil, "", "", err
		}
		devices.AddChild(sharedXmlDoc.Root())
	}

	newXml, err := doc.WriteToString()
	log.Debug().Msgf("New domain XML: %s", newXml)
	if err != nil {
		return nil, "", "", err
	}

	domain, err = e.conn.DomainCreateXML(newXml, virt.DOMAIN_NONE)

	return domain, uuid, guestOS, err
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
			return nil, err
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
		xml, err := domain.GetXMLDesc(virt.DOMAIN_XML_INACTIVE)
		if err != nil {
			return "", err
		}

		// parse the XML into etree
		doc := etree.NewDocument()
		if err := doc.ReadFromString(xml); err != nil {
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
	dErr := fmt.Errorf("Could not determine domain type!")

	xml, err := domain.GetXMLDesc(virt.DOMAIN_XML_INACTIVE)
	if err != nil {
		return "", err
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return "", err
	}
	el := doc.FindElement("/domain[@type]")
	if el == nil {
		return "", dErr
	}
	domainType := el.SelectAttr("type")
	if domainType == nil {
		return "", dErr
	}

	return domainType.Value, nil
}
