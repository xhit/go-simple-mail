// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mail

import (
	"errors"
)

// Auth is implemented by an SMTP authentication mechanism.
type Auth interface {
	// Start begins an authentication with a server.
	// It returns the name of the authentication protocol
	// and optionally data to include in the initial AUTH message
	// sent to the server. It can return proto == "" to indicate
	// that the authentication should be skipped.
	// If it returns a non-nil error, the SMTP client aborts
	// the authentication attempt and closes the connection.
	Start(server *ServerInfo) (proto string, toServer []byte, err error)

	// Next continues the authentication. The server has just sent
	// the fromServer data. If more is true, the server expects a
	// response, which Next should return as toServer; otherwise
	// Next should return toServer == nil.
	// If Next returns a non-nil error, the SMTP client aborts
	// the authentication attempt and closes the connection.
	Next(fromServer []byte, more bool) (toServer []byte, err error)
}

// ServerInfo records information about an SMTP server.
type ServerInfo struct {
	Name string   // SMTP server name
	TLS  bool     // using TLS, with valid certificate for Name
	Auth []string // advertised authentication mechanisms
}

type plainAuth struct {
	identity, username, password string
	host                         string
}

// PlainAuth returns an Auth that implements the PLAIN authentication
// mechanism as defined in RFC 4616. The returned Auth uses the given
// username and password to authenticate to host and act as identity.
// Usually identity should be the empty string, to act as username.
//
// PlainAuth will only send the credentials if the connection is using TLS
// or is connected to localhost. Otherwise authentication will fail with an
// error, without sending the credentials.
func PlainAuth(identity, username, password, host string) Auth {
	return &plainAuth{identity, username, password, host}
}


func (a *plainAuth) Start(server *ServerInfo) (string, []byte, error) {
	// Must have TLS, or else localhost server. Unencrypted connection is permitted here too but is not recommended
	// Note: If TLS is not true, then we can't trust ANYTHING in ServerInfo.
	// In particular, it doesn't matter if the server advertises PLAIN auth.
	// That might just be the attacker saying
	// "it's ok, you can trust me with your password."
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}
	resp := []byte(a.identity + "\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *plainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		// We've already sent everything.
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}