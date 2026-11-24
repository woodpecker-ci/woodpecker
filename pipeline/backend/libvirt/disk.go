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

//go:build cgo && libvirt

package libvirt

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

func (e *libvirt) CreateSharedDisk(ctx context.Context, guestOS, domainType, diskSize, taskUUID string) (string, string, error) {
	libvirtImgDir := e.c.String("backend-libvirt-image-dir")

	var diskUUID string
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

		_, err := os.Stat(disk)
		if err != nil {
			// create the disk
			{
				cmd := exec.Command("qemu-img", "create", "-f", "qcow2", disk, diskSize)
				log.Debug().Msgf("Create shared disk: %s", cmd.Args)
				_, err := cmd.Output()
				if err != nil {
					return "", "", cmdError(cmd, err)
				}
			}

			// Format the disk.
			// On redhat systems, this will fail on 'ntfs', because they do not support ntfs inside libguest.
			// On other systems, this may work if the host has the appropriate tools.
			// May need 'libguestfs-winsupport' installed.
			{
				var cmd *exec.Cmd
				if fstype == "ntfs" {
					cmd = exec.Command(
						"guestfish", "-a", disk, "--", "run", ":",
						"part-init", "/dev/sda", "mbr", ":", // initialize MBR
						"part-add", "/dev/sda", "p", "2048", "-1", ":", // create primary partition
						"part-set-mbr-id", "/dev/sda", "1", "0x07", ":", // sets NTFS type
						"mkfs", "ntfs", "/dev/sda1", // format NTFS
					)
				} else {
					cmd = exec.Command(
						"guestfish", "-a", disk, "--", "run", ":",
						"mkfs", fstype, "/dev/sda", // format
					)
				}
				log.Debug().Msgf("Format disk: %s", cmd.Args)
				_, err := cmd.Output()
				if err != nil {
					return "", "", cmdError(cmd, err)
				}
			}
		}

		// get uuid
		diskUUID, err = GetQemuDiskUUID(disk)
		if err != nil {
			return "", "", err
		}

		return disk, diskUUID, nil
	} else { // raw images (basic qemu without libguest, e.g. bhyve, etc.)
		// this only works on a linux host!
		if e.os != "linux" {
			return "", "", fmt.Errorf("cannot automatically create raw disk images on %s. Either use a linux host or define 'disk' in the backend options", e.os)
		}

		disk := fmt.Sprintf("%s/shared_%s.raw", libvirtImgDir, taskUUID)
		var loopDev []byte

		_, statErr := os.Stat(disk)
		if statErr != nil {
			// create disk
			{
				cmd := exec.Command("truncate", "-s", diskSize, disk)
				log.Debug().Msgf("Create shared disk: %s", cmd.Args)
				_, err := cmd.Output()
				if err != nil {
					return "", "", cmdError(cmd, err)
				}
			}
		}

		// mount loop device
		{
			var err error
			loopDev, err = MountLoopDev(disk)
			if err != nil {
				return "", "", err
			}
		}

		if statErr != nil {
			// format
			// TODO: this probably doesn't work for NTFS
			{
				cmd := CmdViaSudo(fmt.Sprintf("mkfs.%s", fstype), "-E", "lazy_itable_init=1,lazy_journal_init=1", string(loopDev))
				log.Debug().Msgf("Format device: %s", cmd.Args)
				_, err := cmd.Output()
				if err != nil {
					return "", "", cmdError(cmd, err)
				}
			}
		}

		// Get disk UUID
		{
			var err error
			diskUUID, err = GetBhyveDiskUUIDLoop(loopDev)
			if err != nil {
				return "", "", err
			}
		}

		// detach loop device
		{
			err := UnmountLoopDev(loopDev)
			if err != nil {
				return "", "", err
			}
		}

		return disk, diskUUID, nil
	}
}

func (e *libvirt) FromBaseImage(ctx context.Context, baseImage, targetImg string, overwrite bool) (string, error) {
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
			log.Debug().Msgf("Error copying base image %s via qemu-img...falling back to manual copy", cmdError(cmd, err))
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

func GetQemuDiskUUID(disk string) (diskUUID string, err error) {
	cmd := exec.Command("virt-filesystems", "--long", "--csv", "-a", disk, "--uuid")
	log.Debug().Msgf("Get UUID: %s", cmd.Args)
	bytes, err := cmd.Output()
	if err != nil {
		return "", cmdError(cmd, err)
	}

	r := csv.NewReader(strings.NewReader(string(bytes)))
	header, err := r.Read()
	if err != nil {
		return "", fmt.Errorf("failed parsing CVS output for UUID: %w", err)
	}
	i := -1
	for ix, e := range header {
		if e == "UUID" {
			i = ix
		}
	}

	if i == -1 {
		return "", fmt.Errorf("could not find UUID in: %s", header)
	}

	data, err := r.Read()
	if err != nil {
		return "", fmt.Errorf("failed parsing CVS output for UUID: %w", err)
	}
	diskUUID = data[i]

	return diskUUID, nil
}

func GetBhyveDiskUUIDLoop(loopDev []byte) (diskUUID string, err error) {
	cmd := CmdViaSudo("blkid", "-o", "value", "--match-tag", "UUID", string(loopDev))
	log.Debug().Msgf("Get UUID: %s", cmd.Args)
	bytes, err := cmd.Output()
	if err != nil {
		return "", cmdError(cmd, err)
	}

	diskUUID = string(bytes)

	return diskUUID, nil
}

func GetBhyveDiskUUID(disk string) (diskUUID string, err error) {
	var loopDev []byte
	// mount loop device
	{
		var err error
		loopDev, err = MountLoopDev(disk)
		if err != nil {
			return "", err
		}
	}

	// Get disk UUID
	{
		var err error
		diskUUID, err = GetBhyveDiskUUIDLoop(loopDev)
		if err != nil {
			return "", err
		}
	}

	// detach loop device
	{
		err = UnmountLoopDev(loopDev)
		if err != nil {
			return "", err
		}
	}

	return diskUUID, nil
}

func MountLoopDev(disk string) ([]byte, error) {
	cmd := CmdViaSudo("losetup", "--find", "--show", disk)
	log.Debug().Msgf("Mount loop device: %s", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		return nil, cmdError(cmd, err)
	}
	return out, nil
}

func UnmountLoopDev(loopDev []byte) error {
	cmd := CmdViaSudo("losetup", "--detach", string(loopDev))
	log.Debug().Msgf("Detach loop device: %s", cmd.Args)
	_, err := cmd.Output()
	if err != nil {
		return cmdError(cmd, err)
	}
	return nil
}

func GetDiskUUID(disk, domainType string) (diskUUID string, err error) {
	if (domainType == "kvm" || domainType == "qemu") && hasCommand("guestfish") && hasCommand("qemu-img") { // qemu with libguest stuff
		return GetQemuDiskUUID(disk)
	} else {
		return GetBhyveDiskUUID(disk)
	}
}
