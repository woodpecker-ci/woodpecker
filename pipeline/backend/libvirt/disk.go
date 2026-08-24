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
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

		_, err := os.Stat(disk)
		if err != nil {
			// create the disk
			{
				cmd := exec.Command("qemu-img", "create", "-f", "qcow2", disk, diskSize)
				log.Debug().Msgf("Create shared disk: %s", cmd.Args)
				_, err := cmd.Output()
				if err != nil {
					return "", "", fmt.Errorf("command %s failed with %s and stderr: %s", cmd.String(), err.Error(), err.(*exec.ExitError).Stderr)
				}
			}

			// Format the disk.
			// On redhat systems, this will fail on 'ntfs', because they do not support ntfs inside libguest.
			// On other systems, this may work if the host has the appropriate tools.
			// May need 'libguestfs-winsupport' installed.
			{
				var cmd *exec.Cmd
				if fstype == "ntfs" {
					cmd = exec.Command("guestfish", "-a", disk, "--", "run", ":",
						"part-init", "/dev/sda", "mbr", ":", // initialize MBR
						"part-add", "/dev/sda", "p", "2048", "-1", ":", // create primary partition
						"part-set-mbr-id", "/dev/sda", "1", "0x07", ":", // sets NTFS type
						"mkfs", "ntfs", "/dev/sda1", // format NTFS
					)

				} else {
					cmd = exec.Command("guestfish", "-a", disk, "--", "run", ":",
						"mkfs", fstype, "/dev/sda", // format
					)

				}
				log.Debug().Msgf("Format disk: %s", cmd.Args)
				_, err := cmd.Output()
				if err != nil {
					return "", "", fmt.Errorf("command %s failed with %s and stderr: %s", cmd.String(), err.Error(), err.(*exec.ExitError).Stderr)
				}
			}

		}

		// get uuid
		diskUuid, err = GetQemuDiskUUID(disk)
		if err != nil {
			return "", "", err
		}

		return disk, diskUuid, nil

	} else { // raw images (basic qemu without libguest, e.g. bhyve, etc.)
		// this only works on a linux host!
		if e.os != "linux" {
			return "", "", fmt.Errorf("Cannot automatically create raw disk images on %s. Either use a linux host or define 'disk' in the backend options.", e.os)
		}

		disk := fmt.Sprintf("%s/shared_%s.raw", libvirtImgDir, taskUUID)
		var loop_dev []byte

		_, statErr := os.Stat(disk)
		if statErr != nil {
			// create disk
			{
				cmd := exec.Command("truncate", "-s", diskSize, disk)
				log.Debug().Msgf("Create shared disk: %s", cmd.Args)
				_, err := cmd.Output()
				if err != nil {
					return "", "", fmt.Errorf("command %s failed with %s and stderr: %s", cmd.String(), err.Error(), err.(*exec.ExitError).Stderr)
				}
			}
		}

		// mount loop device
		{
			var err error
			loop_dev, err = MountLoopDev(disk)
			if err != nil {
				return "", "", err
			}
		}

		if statErr != nil {
			// format
			// TODO: this probably doesn't work for NTFS
			{
				cmd := CmdViaSudo(fmt.Sprintf("mkfs.%s", fstype), "-E", "lazy_itable_init=1,lazy_journal_init=1", string(loop_dev))
				log.Debug().Msgf("Format device: %s", cmd.Args)
				_, err := cmd.Output()
				if err != nil {
					return "", "", fmt.Errorf("command %s failed with %s and stderr: %s", cmd.String(), err.Error(), err.(*exec.ExitError).Stderr)
				}
			}
		}

		// Get disk UUID
		{
			var err error
			diskUuid, err = GetBhyveDiskUUIDLoop(loop_dev)
			if err != nil {
				return "", "", err
			}
		}

		// detach loop device
		{
			err := UnmountLoopDev(loop_dev)
			if err != nil {
				return "", "", err
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

func GetQemuDiskUUID(disk string) (diskUuid string, err error) {
	cmd := exec.Command("virt-filesystems", "--long", "--csv", "-a", disk, "--uuid")
	log.Debug().Msgf("Get UUID: %s", cmd.Args)
	bytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command %s failed with %s and stderr: %s", cmd.String(), err.Error(), err.(*exec.ExitError).Stderr)
	}

	r := csv.NewReader(strings.NewReader(string(bytes)))
	header, err := r.Read()
	if err != nil {
		return "", fmt.Errorf("Failed parsing CVS output for UUID: %s", err.Error())
	}
	i := -1
	for ix, e := range header {
		if e == "UUID" {
			i = ix
		}
	}

	if i == -1 {
		return "", fmt.Errorf("Could not find UUID in: %s", header)
	}

	data, err := r.Read()
	if err != nil {
		return "", fmt.Errorf("Failed parsing CVS output for UUID: %s", err.Error())
	}
	diskUuid = data[i]

	return diskUuid, nil
}

func GetBhyveDiskUUIDLoop(loop_dev []byte) (diskUuid string, err error) {
	cmd := CmdViaSudo("blkid", "-o", "value", "--match-tag", "UUID", string(loop_dev))
	log.Debug().Msgf("Get UUID: %s", cmd.Args)
	bytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command %s failed with %s and stderr: %s", cmd.String(), err.Error(), err.(*exec.ExitError).Stderr)
	}

	diskUuid = string(bytes)

	return diskUuid, nil
}

func GetBhyveDiskUUID(disk string) (diskUuid string, err error) {
	var loop_dev []byte
	// mount loop device
	{
		var err error
		loop_dev, err = MountLoopDev(disk)
		if err != nil {
			return "", err
		}
	}

	// Get disk UUID
	{
		var err error
		diskUuid, err = GetBhyveDiskUUIDLoop(loop_dev)
		if err != nil {
			return "", err
		}
	}

	// detach loop device
	{
		err = UnmountLoopDev(loop_dev)
		if err != nil {
			return "", err
		}
	}

	return diskUuid, nil
}

func MountLoopDev(disk string) ([]byte, error) {
	cmd := CmdViaSudo("losetup", "--find", "--show", disk)
	log.Debug().Msgf("Mount loop device: %s", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command %s failed with %s and stderr: %s", cmd.String(), err.Error(), err.(*exec.ExitError).Stderr)
	}
	return out, nil
}

func UnmountLoopDev(loop_dev []byte) error {
	cmd := CmdViaSudo("losetup", "--detach", string(loop_dev))
	log.Debug().Msgf("Detach loop device: %s", cmd.Args)
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("command %s failed with %s and stderr: %s", cmd.String(), err.Error(), err.(*exec.ExitError).Stderr)
	}
	return nil
}

func GetDiskUUID(disk string, domainType string) (diskUuid string, err error) {
	if (domainType == "kvm" || domainType == "qemu") && hasCommand("guestfish") && hasCommand("qemu-img") { // qemu with libguest stuff
		return GetQemuDiskUUID(disk)
	} else {
		return GetBhyveDiskUUID(disk)
	}
}
