// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Auth represents ssh auth methods.
type Auth []ssh.AuthMethod

// Password returns password auth method.
func Password(pass string) Auth {
	return Auth{
		ssh.Password(pass),
	}
}

// Key returns auth method from private key with or without passphrase.
func Key(prvFile string, passphrase string) (Auth, error) {

	signer, err := GetSigner(prvFile, passphrase)

	if err != nil {
		return nil, err
	}

	return Auth{
		ssh.PublicKeys(signer),
	}, nil
}

// HasAgent checks if ssh agent exists.
func HasAgent() bool {
	return os.Getenv("SSH_AUTH_SOCK") != ""
}

// UseAgent auth via ssh agent, (Unix systems only)
func UseAgent() (Auth, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, fmt.Errorf("could not find ssh agent: %w", err)
	}
	return Auth{
		ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
	}, nil
}

// GetSigner returns ssh signer from private key file.
func GetSigner(prvFile string, passphrase string) (ssh.Signer, error) {

	var (
		err    error
		signer ssh.Signer
	)

	privateKey, err := ioutil.ReadFile(prvFile)

	if err != nil {

		return nil, err

	} else if passphrase != "" {

		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))

	} else {

		signer, err = ssh.ParsePrivateKey(privateKey)
	}

	return signer, err
}
