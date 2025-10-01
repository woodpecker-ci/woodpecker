// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT.

// cSpell:words hostmatcher
package hostmatcher

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"syscall"
	"time"
)

// NewDialContext returns a DialContext for Transport, the DialContext will do allow/block list check.
func NewDialContext(usage string, allowList *HostMatchList) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return NewDialContextWithProxy(usage, allowList, nil)
}

func NewDialContextWithProxy(usage string, allowList *HostMatchList, proxy *url.URL) func(ctx context.Context, network, addr string) (net.Conn, error) {
	// How Go HTTP Client works with redirection:
	//   transport.RoundTrip URL=http://domain.com, Host=domain.com
	//   transport.DialContext addrOrHost=domain.com:80
	//   dialer.Control tcp4:11.22.33.44:80
	//   transport.RoundTrip URL=http://www.domain.com/, Host=(empty here, in the direction, HTTP client doesn't fill the Host field)
	//   transport.DialContext addrOrHost=domain.com:80
	//   dialer.Control tcp4:11.22.33.44:80
	return func(ctx context.Context, network, addrOrHost string) (net.Conn, error) {
		// default values are from http.DefaultTransport
		const dialTimeout = 30 * time.Second
		const dialKeepAlive = 30 * time.Second

		dialer := net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: dialKeepAlive,

			Control: func(network, ipAddr string, _ syscall.RawConn) error {
				host, port, err := net.SplitHostPort(addrOrHost)
				if err != nil {
					return err
				}
				if proxy != nil {
					// Always allow the host of the proxy, but only on the specified port.
					if host == proxy.Hostname() && port == proxy.Port() {
						return nil
					}
				}

				// in Control func, the addr was already resolved to IP:PORT format, there is no cost to do ResolveTCPAddr here
				tcpAddr, err := net.ResolveTCPAddr(network, ipAddr)
				if err != nil {
					return fmt.Errorf("%s can only call HTTP servers via TCP, deny '%s(%s:%s)', err=%w", usage, host, network, ipAddr, err)
				}

				// if we have an allow-list, check the allow-list first
				if !allowList.IsEmpty() {
					if !allowList.MatchHostOrIP(host, tcpAddr.IP) {
						return fmt.Errorf("%s can only call allowed HTTP servers (check your %s setting), deny '%s(%s)'", usage, allowList.SettingKeyHint, host, ipAddr)
					}
				}

				return nil
			},
		}
		return dialer.DialContext(ctx, network, addrOrHost)
	}
}
