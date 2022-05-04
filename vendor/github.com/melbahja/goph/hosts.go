// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"errors"
	"fmt"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// DefaultKnownHosts returns host key callback from default known hosts path, and error if any.
func DefaultKnownHosts() (ssh.HostKeyCallback, error) {

	path, err := DefaultKnownHostsPath()
	if err != nil {
		return nil, err
	}

	return KnownHosts(path)
}

// KnownHosts returns host key callback from a custom known hosts path.
func KnownHosts(file string) (ssh.HostKeyCallback, error) {
	return knownhosts.New(file)
}

// CheckKnownHost checks is host in known hosts file.
// it returns is the host found in known_hosts file and error, if the host found in
// known_hosts file and error not nil that means public key mismatch, maybe MAN IN THE MIDDLE ATTACK! you should not handshake.
func CheckKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (found bool, err error) {

	var keyErr *knownhosts.KeyError

	// Fallback to default known_hosts file
	if knownFile == "" {
		path, err := DefaultKnownHostsPath()
		if err != nil {
			return false, err
		}

		knownFile = path
	}

	// Get host key callback
	callback, err := KnownHosts(knownFile)

	if err != nil {
		return false, err
	}

	// check if host already exists.
	err = callback(host, remote, key)

	// Known host already exists.
	if err == nil {
		return true, nil
	}

	// Make sure that the error returned from the callback is host not in file error.
	// If keyErr.Want is greater than 0 length, that means host is in file with different key.
	if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
		return true, keyErr
	}

	// Some other error occurred and safest way to handle is to pass it back to user.
	if err != nil {
		return false, err
	}

	// Key is not trusted because it is not in the file.
	return false, nil
}

// AddKnownHost add a a host to known hosts file.
func AddKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (err error) {

	// Fallback to default known_hosts file
	if knownFile == "" {
		path, err := DefaultKnownHostsPath()
		if err != nil {
			return err
		}

		knownFile = path
	}

	f, err := os.OpenFile(knownFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	remoteNormalized := knownhosts.Normalize(remote.String())
	hostNormalized := knownhosts.Normalize(host)
	addresses := []string{remoteNormalized}

	if hostNormalized != remoteNormalized {
		addresses = append(addresses, hostNormalized)
	}

	_, err = f.WriteString(knownhosts.Line(addresses, key) + "\n")

	return err
}

// DefaultKnownHostsPath returns default user knows hosts file.
func DefaultKnownHostsPath() (string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/.ssh/known_hosts", home), err
}
