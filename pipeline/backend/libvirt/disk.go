package libvirt

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
	"github.com/rs/zerolog/log"
)

func (e *libvirt) CreateSharedDisk(ctx context.Context, guestOS string, domainType string, diskSize string, taskUUID string) (string, string, error) {
	libvirtImgDir := e.c.String("backend-libvirt-image-dir")

	var diskUuid string
	var fstype string
	if guestOS == "windows" {
		fstype = "ntfs"
	} else { // TODO: UFS for BSD?
		fstype = "ext4"
	}

	// create a disk from scratch based on the domain type
	// kvm and qemu support qcow2, for everything else we use raw image
	if (domainType == "kvm" || domainType == "qemu") && hasCommand("guestfish") && hasCommand("qemu-img") { // qemu with libguest stuff
		disk := fmt.Sprintf("%s/shared_%s.qcow2", libvirtImgDir, taskUUID)
		// create qcow2 image

		// create the disk
		{
			cmd := exec.Command("qemu-img", "create", "-f", "qcow2", disk, diskSize)
			err := cmd.Run()
			if err != nil {
				return "", "", fmt.Errorf("command %s failed with: %s", cmd.String(), err.Error())
			}
		}

		// Format the disk.
		// On redhat systems, this will fail on 'ntfs', because they do not support ntfs inside libguest.
		// On other systems, this may work if the host has the appropriate tools.
		// May need 'libguestfs-winsupport' installed.
		{
			cmd := exec.Command("guestfish", "-a", disk, "--", "run", ":",
				"part-disk", "/dev/sda", "mbr", ":",
				"mkfs", fstype, "/dev/sda1")
			err := cmd.Run()
			if err != nil {
				return "", "", fmt.Errorf("command %s failed with: %s", cmd.String(), err.Error())
			}
		}

		// get uuid
		{
			cmd := exec.Command("virt-filesystems", "--long", "--csv", "-a", disk, "--uuid")
			bytes, err := cmd.Output()
			if err != nil {
				return "", "", fmt.Errorf("command %s failed with: %s", cmd.String(), err.Error())
			}

			r := csv.NewReader(strings.NewReader(string(bytes)))
			header, err := r.Read()
			if err != nil {
				return "", "", err
			}
			i := -1
			for ix, e := range header {
				if e == "UUID" {
					i = ix
				}
			}

			if i == -1 {
				return "", "", fmt.Errorf("Could not find UUID in: %s", header)
			}

			data, err := r.Read()
			if err != nil {
				return "", "", err
			}
			diskUuid = data[i]

		}

		return disk, diskUuid, nil

	} else { // raw images (basic qemu without libguest, bhyve, etc.)
		// this only works on a linux host!
		if e.os != "linux" {
			return "", "", fmt.Errorf("Cannot automatically create raw disk images on %s. Either use a linux host or define 'disk_config' yourself.", e.os)
		}

		disk := fmt.Sprintf("%s/shared_%s.raw", libvirtImgDir, taskUUID)

		// create disk
		{
			cmd := exec.Command("truncate", "-s", diskSize, disk)
			err := cmd.Run()
			if err != nil {
				return "", "", fmt.Errorf("command %s failed with: %s", cmd.String(), err.Error())
			}
		}

		// mount loop device
		var loop_dev []byte
		{
			cmd := CmdViaSudo("losetup", "--find", "--show", disk)
			var err error
			loop_dev, err = cmd.Output()
			if err != nil {
				return "", "", fmt.Errorf("command %s failed with: %s", cmd.String(), err.Error())
			}
		}

		// format
		{
			cmd := CmdViaSudo(fmt.Sprintf("mkfs.%s", fstype), "-E", "lazy_itable_init=1,lazy_journal_init=1", string(loop_dev))
			err := cmd.Run()
			if err != nil {
				return "", "", fmt.Errorf("command %s failed with: %s", cmd.String(), err.Error())
			}
		}

		// detach loop device
		{
			cmd := CmdViaSudo("blkid", "-o", "value", "--match-tag", "UUID", string(loop_dev))
			bytes, err := cmd.Output()
			if err != nil {
				return "", "", fmt.Errorf("command %s failed with: %s", cmd.String(), err.Error())
			}

			diskUuid = string(bytes)
		}

		// detach loop device
		{
			cmd := CmdViaSudo("losetup", "--detach", string(loop_dev))
			err := cmd.Run()
			if err != nil {
				return "", "", fmt.Errorf("command %s failed with: %s", cmd.String(), err.Error())
			}
		}

		return disk, diskUuid, nil
	}
}

func (e *libvirt) FromBaseImage(ctx context.Context, baseImage string, targetImg string, overwrite bool) (string, error) {
	libvirtImgDir := e.c.String("backend-libvirt-image-dir")
	libvirtImg := filepath.Join(libvirtImgDir, targetImg)

	if !overwrite {
		_, err := os.Stat(libvirtImg)
		if err == nil {
			return libvirtImg, nil
		}
	}

	log.Debug().Msgf("Copying base image %s to %s ", baseImage, libvirtImg)

	if filepath.Ext(baseImage) == ".qcow2" {
		cmd := exec.Command("qemu-img", "create", "-o", "backing_fmt=qcow2", "-b", baseImage, "-f", "qcow2", libvirtImg)
		_, err := cmd.Output()
		if err != nil {
			log.Debug().Msgf("Error copying base image %s via qemu-img...falling back to manual copy", string(err.(*exec.ExitError).Stderr))
			err := CopyFile(baseImage, libvirtImg, false)
			if err != nil {
				return "", err
			}
		}
	} else { // TODO: maybe other formats can be efficiently copied too?
		err := CopyFile(baseImage, libvirtImg, false)
		if err != nil {
			return "", err
		}
	}

	return libvirtImg, nil
}

func (e *libvirt) EphemeralizeSharedDisk(ctx context.Context, shared_disk *SharedDisk, taskUUID string) (string, error) {
	if shared_disk.DiskConfig == "" {
		return "", fmt.Errorf("No image specified")
	}

	xml := shared_disk.DiskConfig
	log.Debug().Msgf("Shared disk config: %s", xml)

	// parse the XML into etree
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return "", err
	}

	// process the XML, replacing the disk
	el := doc.FindElements("//disk/source[@file]")
	for _, child := range el {
		attr := child.SelectAttr("file")
		if attr == nil {
			continue
		}

		sharedImg := fmt.Sprintf(sharedDiskPattern, taskUUID, filepath.Ext(attr.Value))

		newImg, err := e.FromBaseImage(ctx, attr.Value, sharedImg, false)
		if err != nil {
			return "", err
		}
		child.CreateAttr("file", newImg)
	}

	newXml, err := doc.WriteToString()
	if err != nil {
		return "", err
	}
	return newXml, nil
}
