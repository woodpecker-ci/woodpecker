# Libvirt

Woodpecker supports a new experimental backend for executing steps within a virtual machine.

This is achieved by utilizing [libvirt](https://libvirt.org/).

## Architecture

Libvirt is an abstraction over different hypervisors. Currently QEMU and bhyve
have been extensively tested.

In order to fit virtual machines ("domains" in libvirt speak) into a docker-like workflow,
we had to make a couple of choices. Most of these prioritize portability.

Some methods don't work well on some host platforms or hypervisors, so experience may
vary depending on the setup.

### Command execution

Commands are executed over an SSH session and fed into standard input. So there
should be no limitation for script size. Environment variables are injected directly
into the script. We do not rely on SSH's `AcceptEnv` feature.

### Connecting to the VM

Connections to the VMs are established via the [libvirt nss](https://libvirt.org/nss.html)
module (libvirt must be compiled with that feature). This maps the domain names to hostnames.

Alternatively it is possible to specify the name of the guest internal interface on which
SSH is listening. See the [backend options](#backend-options) for more information.

### Ephemeral execution and shared volumes

Just like containers, we want the VM state to be gone after a step has finished. This is achieved
by copying the base image of the domain.

Similarly, a new empty disk is created ad-hoc, formatted, connected to the VM and mounted
at `$CI_WORKSPACE`.
This has the side-effect that two VMs can't access the same disk at the same time without
filesystem corruption, so step parallelization does not work currently. We ensure only one
step is running at a time via mutex locks.

An alternative to ad-hoc disks would be [virtiofs](https://virtio-fs.gitlab.io/). It has the advantage that we get shared access
to a directory on the host, but the disadvantage that it requires more configuration inside the guest
and is possibly less portable.

After the workflow is finished, all the copied base images and shared disks are deleted, unless
the agent was started with `--backend-libvirt-keep-tmp`.

### Domain definitions remain untouched

Although we manipulate the domain XML to inject a shared disk, change the base image
and so on... none of these changes are persistent.

## Prerequisites

These are the requirements for the host that runs the woodpecker agent:

- hardware that supports virtualization
- libvirt supported operating system
- libvirt C libraries installed
- libvirtd service started and operational (including networking)
- sufficient privileges for the user running the agent to interact with libvirtd and write to the images store (`/var/lib/libvirt/images` by default)
- [libvirt nss](https://libvirt.org/nss.html) set up, so that we can connect to the machine by name

## VM creation

Libvirt has an [ecosystem of tools](https://www.virt-tools.org/).

To create a VM conveniently you can use [virt-manager's](https://virt-manager.org/) GUI.
It also provides a command line tool called `virt-install`.

There is also more advanced methods like [cloud-init](https://cloud-init.io/).

### Configuration for all VM guests

Regardless of guest, you need an sshd server running, which must be configured to accept
password authentication.

### Configuration for windows guests

Make sure the login account is an admin account. It usually needs to be a local account with a password, not a pin.

```
net localgroup Administrators <username> /add
```

OpenSSH server needs to be [installed and started](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse?tabs=gui&pivots=windows-11).

Then we need some further configuration to allow remote SSH connections.

Disable UAC remote token filtering:

```powershell
New-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" `
    -Name "LocalAccountTokenFilterPolicy" -Value 1 -PropertyType DWord -Force
```

Set your network to private:

```powershell
Get-NetIPInterface
Set-NetConnectionProfile -InterfaceIndex <index> -NetworkCategory Private
```

Start WinRM service:

```powershell
winrm quickconfig
```

We also need to turn off bitlocker disk encryption. Windows 11 can break our shared
volumes by encrypting them without asking.

```powershell
New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\BitLocker" `
    -Name "PreventDeviceEncryption" -Value 1 -PropertyType DWord -Force
```

## Backend options

```yaml
    backend_options:
      libvirt:
        # Setting this to 'true' will make the changes to the VM
        # visible to other steps/workflows and break isolation.
        # This can be useful however if VM creation is automated
        # in some other way.
        persistent: false
        # This is optional and should only be used if the backend
        # can't figure out how to create and format a disk as the
        # shared volume.
        shared_disk:
          # The UUID is needed for mounting on linux.
          # On windows we create a serial number on the fly and use
          # that for mounting.
          uuid: bd5ea041-e765-4d3a-9cda-083e4e52b187
          # The disk image to use. Must be a filename. Paths
          # are ignored. Both qcow2 and raw img are fine.
          disk_image: shared.qcow2
        # The SSH configuration. This is required.
        ssh_config:
          # required
          username: root
          # Optional. Timeout before giving up on connecting via ssh. If
          # your VM needs a long time to start up, you may want to increase this.
          # Takes a duration string, such as "2m" for two minutes.
          timeout: 1m
          # Optional. Only use this if libvirt NSS doesn't work. This
          # must be the interface inside the guest on which we can connect to sshd.
          interface: eth0
```

The SSH password has to be configured via the environment variable `LIBVIRT_SSH_PW`,
so that we can use the `from_secret` feature. See the [example workflow](#example-workflow).

## Example workflow

```yaml
when:
  - event: push
  - event: manual

labels:
  hypervisor: qemu

skip_clone: true

steps:
 - name: build
   image: win11
   commands:
     # install ghcup and msys2
     - Set-ExecutionPolicy Bypass -Scope Process -Force;[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; try { & ([ScriptBlock]::Create((Invoke-WebRequest https://raw.githubusercontent.com/haskell/ghcup-hs/refs/heads/no-pty/scripts/bootstrap/bootstrap-haskell.ps1 -UseBasicParsing))) -DisableCurl -Msys2Env CLANG64 -Minimal } catch { Write-Error $_ }
     # clone the repository
     - C:/ghcup/msys64/usr/bin/bash.exe --login -e -c "ls -lah && pwd && pacman --noconfirm -S --needed --overwrite '*' bash git && git clone ${CI_REPO_CLONE_URL} --branch ${CI_COMMIT_BRANCH} . && ghcup install ghc --set $env:GHC && ghcup install cabal --set $env:CABAL"
     # build
     - C:/ghcup/msys64/usr/bin/bash.exe --login -e .github/scripts/build.sh --ghc ghc-$env:GHC --verbose --os Windows --check-linking --strip
   environment:
     GHC: 9.10.3
     CABAL: 3.18.1.0
     CABAL_DIR: "C:/cabal"
     MSYSTEM: CLANG64
     CHERE_INVOKING: 1
     GHCUP_MSYS2_ENV: CLANG64
     GHCUP_INSTALL_BASE_PREFIX: "C:/"
     LIBVIRT_SSH_PW:
       from_secret: win11_ssh_pw
   backend_options:
     libvirt:
       ssh_config:
         username: maerwald
```

## Caveats

Plugins don't work. They are docker specific. That also means that automatic cloning
doesn't work and we always need `skip_clone: true`.
