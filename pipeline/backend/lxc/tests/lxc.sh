#!/bin/sh -ex

: ${CLASS_C:=10.0.7}

apt-get install -y -qq git libvirt0 libpam-cgfs bridge-utils uidmap dnsmasq-base dnsmasq dnsmasq-utils qemu-user-static
systemctl stop dnsmasq
systemctl disable dnsmasq
apt-get install -y -qq lxc
systemctl stop lxc-net
cat >> /etc/default/lxc-net <<EOF
LXC_ADDR="$CLASS_C.1"
LXC_NETMASK="255.255.255.0"
LXC_NETWORK="$CLASS_C.0/24"
LXC_DHCP_RANGE="$CLASS_C.2,$CLASS_C.254"
LXC_DHCP_MAX="253"
EOF
if ! systemctl start lxc-net ; then
    journalctl --unit lxc-net --no-pager
    exit 1
fi
systemctl status
