// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mail

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

type authTest struct {
	auth       auth
	challenges []string
	name       string
	responses  []string
}

var authTests = []authTest{
	{plainAuthfn("", "user", "pass", "testserver"), []string{}, "PLAIN", []string{"\x00user\x00pass"}},
	{plainAuthfn("foo", "bar", "baz", "testserver"), []string{}, "PLAIN", []string{"foo\x00bar\x00baz"}},
	{loginAuthfn("", "bar", "baz", "testserver"), []string{}, "LOGIN", []string{"bar"}},
	{loginAuthfn("foo", "bar", "baz", "testserver"), []string{}, "LOGIN", []string{"bar"}},
	{cramMD5Authfn("user", "pass"), []string{"<123456.1322876914@testserver>"}, "CRAM-MD5", []string{"", "user 287eb355114cf5c471c26a875f1ca4ae"}},
}

func TestAuth(t *testing.T) {
testLoop:
	for i, test := range authTests {
		name, resp, err := test.auth.start(&serverInfo{"testserver", true, nil})
		if name != test.name {
			t.Errorf("#%d got name %s, expected %s", i, name, test.name)
		}
		if !bytes.Equal(resp, []byte(test.responses[0])) {
			t.Errorf("#%d got response %s, expected %s", i, resp, test.responses[0])
		}
		if err != nil {
			t.Errorf("#%d error: %s", i, err)
		}
		for j := range test.challenges {
			challenge := []byte(test.challenges[j])
			expected := []byte(test.responses[j+1])
			resp, err := test.auth.next(challenge, true)
			if err != nil {
				t.Errorf("#%d error: %s", i, err)
				continue testLoop
			}
			if !bytes.Equal(resp, expected) {
				t.Errorf("#%d got %s, expected %s", i, resp, expected)
				continue testLoop
			}
		}
	}
}

func TestAuthPlain(t *testing.T) {

	tests := []struct {
		authName string
		server   *serverInfo
		err      string
	}{
		{
			authName: "servername",
			server:   &serverInfo{name: "servername", tls: true},
		},
		{
			// OK to use plainAuthfn on localhost without TLS
			authName: "localhost",
			server:   &serverInfo{name: "localhost", tls: false},
		},
		{
			authName: "servername",
			server:   &serverInfo{name: "attacker", tls: true},
			err:      "wrong host name",
		},
	}
	for i, tt := range tests {
		auth := plainAuthfn("foo", "bar", "baz", tt.authName)
		_, _, err := auth.start(tt.server)
		got := ""
		if err != nil {
			got = err.Error()
		}
		if got != tt.err {
			t.Errorf("%d. got error = %q; want %q", i, got, tt.err)
		}
	}
}

func TestAuthLogin(t *testing.T) {

	tests := []struct {
		authName string
		server   *serverInfo
		err      string
	}{
		{
			authName: "servername",
			server:   &serverInfo{name: "servername", tls: true},
		},
		{
			// OK to use loginAuthfn on localhost without TLS
			authName: "localhost",
			server:   &serverInfo{name: "localhost", tls: false},
		},
		{
			authName: "servername",
			server:   &serverInfo{name: "attacker", tls: true},
			err:      "wrong host name",
		},
	}
	for i, tt := range tests {
		auth := loginAuthfn("foo", "bar", "baz", tt.authName)
		_, _, err := auth.start(tt.server)
		got := ""
		if err != nil {
			got = err.Error()
		}
		if got != tt.err {
			t.Errorf("%d. got error = %q; want %q", i, got, tt.err)
		}
	}
}

// Issue https://github.com/golang/go/issues/17794: don't send a trailing space on AUTH command when there's no password.
func TestClientAuthTrimSpace(t *testing.T) {
	server := "220 hello world\r\n" +
		"200 some more"
	var wrote bytes.Buffer
	var fake faker
	fake.ReadWriter = struct {
		io.Reader
		io.Writer
	}{
		strings.NewReader(server),
		&wrote,
	}
	c, err := newClient(fake, "fake.host")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.tls = true
	c.didHello = true
	c.authenticate(toServerEmptyAuth{})
	c.close()
	if got, want := wrote.String(), "AUTH FOOAUTH\r\n*\r\nQUIT\r\n"; got != want {
		t.Errorf("wrote %q; want %q", got, want)
	}
}

// toServerEmptyAuth is an implementation of Auth that only implements
// the Start method, and returns "FOOAUTH", nil, nil. Notably, it returns
// zero bytes for "toServer" so we can test that we don't send spaces at
// the end of the line. See TestClientAuthTrimSpace.
type toServerEmptyAuth struct{}

func (toServerEmptyAuth) start(server *serverInfo) (proto string, toServer []byte, err error) {
	return "FOOAUTH", nil, nil
}

func (toServerEmptyAuth) next(fromServer []byte, more bool) (toServer []byte, err error) {
	panic("unexpected call")
}

type faker struct {
	io.ReadWriter
}

func (f faker) Close() error                     { return nil }
func (f faker) LocalAddr() net.Addr              { return nil }
func (f faker) RemoteAddr() net.Addr             { return nil }
func (f faker) SetDeadline(time.Time) error      { return nil }
func (f faker) SetReadDeadline(time.Time) error  { return nil }
func (f faker) SetWriteDeadline(time.Time) error { return nil }

func TestBasic(t *testing.T) {
	server := strings.Join(strings.Split(basicServer, "\n"), "\r\n")
	client := strings.Join(strings.Split(basicClient, "\n"), "\r\n")

	var cmdbuf bytes.Buffer
	bcmdbuf := bufio.NewWriter(&cmdbuf)
	var fake faker
	fake.ReadWriter = bufio.NewReadWriter(bufio.NewReader(strings.NewReader(server)), bcmdbuf)
	c := &smtpClient{text: textproto.NewConn(fake), localName: "localhost"}

	if err := c.helo(); err != nil {
		t.Fatalf("HELO failed: %s", err)
	}
	if err := c.ehlo(); err == nil {
		t.Fatalf("Expected first EHLO to fail")
	}
	if err := c.ehlo(); err != nil {
		t.Fatalf("Second EHLO failed: %s", err)
	}

	c.didHello = true
	if ok, args := c.extension("aUtH"); !ok || args != "LOGIN PLAIN" {
		t.Fatalf("Expected AUTH supported")
	}
	if ok, _ := c.extension("DSN"); ok {
		t.Fatalf("Shouldn't support DSN")
	}

	if err := c.mail("user@gmail.com"); err == nil {
		t.Fatalf("MAIL should require authentication")
	}

	if err := c.verify("user1@gmail.com"); err == nil {
		t.Fatalf("First VRFY: expected no verification")
	}
	if err := c.verify("user2@gmail.com>\r\nDATA\r\nAnother injected message body\r\n.\r\nQUIT\r\n"); err == nil {
		t.Fatalf("VRFY should have failed due to a message injection attempt")
	}
	if err := c.verify("user2@gmail.com"); err != nil {
		t.Fatalf("Second VRFY: expected verification, got %s", err)
	}

	// fake TLS so authentication won't complain
	c.tls = true
	c.serverName = "smtp.google.com"
	if err := c.authenticate(plainAuthfn("", "user", "pass", "smtp.google.com")); err != nil {
		t.Fatalf("AUTH failed: %s", err)
	}

	if err := c.rcpt("golang-nuts@googlegroups.com>\r\nDATA\r\nInjected message body\r\n.\r\nQUIT\r\n"); err == nil {
		t.Fatalf("RCPT should have failed due to a message injection attempt")
	}
	if err := c.mail("user@gmail.com>\r\nDATA\r\nAnother injected message body\r\n.\r\nQUIT\r\n"); err == nil {
		t.Fatalf("MAIL should have failed due to a message injection attempt")
	}
	if err := c.mail("user@gmail.com"); err != nil {
		t.Fatalf("MAIL failed: %s", err)
	}
	if err := c.rcpt("golang-nuts@googlegroups.com"); err != nil {
		t.Fatalf("RCPT failed: %s", err)
	}
	msg := `From: user@gmail.com
To: golang-nuts@googlegroups.com
Subject: Hooray for Go

Line 1
.Leading dot line .
Goodbye.`
	w, err := c.data()
	if err != nil {
		t.Fatalf("DATA failed: %s", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		t.Fatalf("Data write failed: %s", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Bad data response: %s", err)
	}

	if err := c.quit(); err != nil {
		t.Fatalf("QUIT failed: %s", err)
	}

	bcmdbuf.Flush()
	actualcmds := cmdbuf.String()
	if client != actualcmds {
		t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
	}
}

var basicServer = `250 mx.google.com at your service
502 Unrecognized command.
250-mx.google.com at your service
250-SIZE 35651584
250-AUTH LOGIN PLAIN
250 8BITMIME
530 Authentication required
252 Send some mail, I'll try my best
250 User is valid
235 Accepted
250 Sender OK
250 Receiver OK
354 Go ahead
250 Data OK
221 OK
`

var basicClient = `HELO localhost
EHLO localhost
EHLO localhost
MAIL FROM:<user@gmail.com> BODY=8BITMIME
VRFY user1@gmail.com
VRFY user2@gmail.com
AUTH PLAIN AHVzZXIAcGFzcw==
MAIL FROM:<user@gmail.com> BODY=8BITMIME
RCPT TO:<golang-nuts@googlegroups.com>
DATA
From: user@gmail.com
To: golang-nuts@googlegroups.com
Subject: Hooray for Go

Line 1
..Leading dot line .
Goodbye.
.
QUIT
`

func TestExtensions(t *testing.T) {
	faker := func(server string) (c *smtpClient, bcmdbuf *bufio.Writer, cmdbuf *bytes.Buffer) {
		server = strings.Join(strings.Split(server, "\n"), "\r\n")

		cmdbuf = &bytes.Buffer{}
		bcmdbuf = bufio.NewWriter(cmdbuf)
		var fake faker
		fake.ReadWriter = bufio.NewReadWriter(bufio.NewReader(strings.NewReader(server)), bcmdbuf)
		c = &smtpClient{text: textproto.NewConn(fake), localName: "localhost"}

		return c, bcmdbuf, cmdbuf
	}

	t.Run("helo", func(t *testing.T) {
		const (
			basicServer = `250 mx.google.com at your service
250 Sender OK
221 Goodbye
`

			basicClient = `HELO localhost
MAIL FROM:<user@gmail.com>
QUIT
`
		)

		c, bcmdbuf, cmdbuf := faker(basicServer)

		if err := c.helo(); err != nil {
			t.Fatalf("HELO failed: %s", err)
		}
		c.didHello = true
		if err := c.mail("user@gmail.com"); err != nil {
			t.Fatalf("MAIL FROM failed: %s", err)
		}
		if err := c.quit(); err != nil {
			t.Fatalf("QUIT failed: %s", err)
		}

		bcmdbuf.Flush()
		actualcmds := cmdbuf.String()
		client := strings.Join(strings.Split(basicClient, "\n"), "\r\n")
		if client != actualcmds {
			t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
		}
	})

	t.Run("ehlo", func(t *testing.T) {
		const (
			basicServer = `250-mx.google.com at your service
250 SIZE 35651584
250 Sender OK
221 Goodbye
`

			basicClient = `EHLO localhost
MAIL FROM:<user@gmail.com>
QUIT
`
		)

		c, bcmdbuf, cmdbuf := faker(basicServer)

		if err := c.hi("localhost"); err != nil {
			t.Fatalf("EHLO failed: %s", err)
		}
		if ok, _ := c.extension("8BITMIME"); ok {
			t.Fatalf("Shouldn't support 8BITMIME")
		}
		if ok, _ := c.extension("SMTPUTF8"); ok {
			t.Fatalf("Shouldn't support SMTPUTF8")
		}
		if err := c.mail("user@gmail.com"); err != nil {
			t.Fatalf("MAIL FROM failed: %s", err)
		}
		if err := c.quit(); err != nil {
			t.Fatalf("QUIT failed: %s", err)
		}

		bcmdbuf.Flush()
		actualcmds := cmdbuf.String()
		client := strings.Join(strings.Split(basicClient, "\n"), "\r\n")
		if client != actualcmds {
			t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
		}
	})

	t.Run("ehlo size", func(t *testing.T) {
		const (
			basicServer = `250-mx.google.com at your service
250-SIZE 35651584
250 user@gmail.com OK; can accommodate 500000 byte message
250 Sender OK
221 Goodbye
`

			basicClient = `EHLO localhost
MAIL FROM:<user@gmail.com> SIZE=50000
QUIT
`
		)

		c, bcmdbuf, cmdbuf := faker(basicServer)
		cmdArgs := make(map[string]string)

		if err := c.hi("localhost"); err != nil {
			t.Fatalf("EHLO failed: %s", err)
		}
		if ok, _ := c.extension("SIZE"); !ok {
			t.Fatalf("Should support SIZE")
		}
		if ok, _ := c.extension("SMTPUTF8"); ok {
			t.Fatalf("Shouldn't support SMTPUTF8")
		}

		if ok, _ := c.extension("SIZE"); ok {
			cmdArgs["SIZE"] = "50000"
		}
		if err := c.mail("user@gmail.com", cmdArgs); err != nil {
			t.Fatalf("MAIL FROM failed: %s", err)
		}
		if err := c.quit(); err != nil {
			t.Fatalf("QUIT failed: %s", err)
		}

		bcmdbuf.Flush()
		actualcmds := cmdbuf.String()
		client := strings.Join(strings.Split(basicClient, "\n"), "\r\n")
		if client != actualcmds {
			t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
		}
	})

	t.Run("ehlo 8bitmime", func(t *testing.T) {
		const (
			basicServer = `250-mx.google.com at your service
250-SIZE 35651584
250 8BITMIME
250 Sender OK
221 Goodbye
`

			basicClient = `EHLO localhost
MAIL FROM:<user@gmail.com> BODY=8BITMIME
QUIT
`
		)

		c, bcmdbuf, cmdbuf := faker(basicServer)

		if err := c.hi("localhost"); err != nil {
			t.Fatalf("EHLO failed: %s", err)
		}
		if ok, _ := c.extension("8BITMIME"); !ok {
			t.Fatalf("Should support 8BITMIME")
		}
		if ok, _ := c.extension("SMTPUTF8"); ok {
			t.Fatalf("Shouldn't support SMTPUTF8")
		}
		if err := c.mail("user@gmail.com"); err != nil {
			t.Fatalf("MAIL FROM failed: %s", err)
		}
		if err := c.quit(); err != nil {
			t.Fatalf("QUIT failed: %s", err)
		}

		bcmdbuf.Flush()
		actualcmds := cmdbuf.String()
		client := strings.Join(strings.Split(basicClient, "\n"), "\r\n")
		if client != actualcmds {
			t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
		}
	})

	t.Run("ehlo smtputf8", func(t *testing.T) {
		const (
			basicServer = `250-mx.google.com at your service
250-SIZE 35651584
250 SMTPUTF8
250 Sender OK
221 Goodbye
`

			basicClient = `EHLO localhost
MAIL FROM:<user+📧@gmail.com> SMTPUTF8
QUIT
`
		)

		c, bcmdbuf, cmdbuf := faker(basicServer)

		if err := c.hi("localhost"); err != nil {
			t.Fatalf("EHLO failed: %s", err)
		}
		if ok, _ := c.extension("8BITMIME"); ok {
			t.Fatalf("Shouldn't support 8BITMIME")
		}
		if ok, _ := c.extension("SMTPUTF8"); !ok {
			t.Fatalf("Should support SMTPUTF8")
		}
		if err := c.mail("user+📧@gmail.com"); err != nil {
			t.Fatalf("MAIL FROM failed: %s", err)
		}
		if err := c.quit(); err != nil {
			t.Fatalf("QUIT failed: %s", err)
		}

		bcmdbuf.Flush()
		actualcmds := cmdbuf.String()
		client := strings.Join(strings.Split(basicClient, "\n"), "\r\n")
		if client != actualcmds {
			t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
		}
	})

	t.Run("ehlo 8bitmime smtputf8", func(t *testing.T) {
		const (
			basicServer = `250-mx.google.com at your service
250-SIZE 35651584
250-8BITMIME
250 SMTPUTF8
250 Sender OK
221 Goodbye
	`

			basicClient = `EHLO localhost
MAIL FROM:<user+📧@gmail.com> BODY=8BITMIME SMTPUTF8
QUIT
`
		)

		c, bcmdbuf, cmdbuf := faker(basicServer)

		if err := c.hi("localhost"); err != nil {
			t.Fatalf("EHLO failed: %s", err)
		}
		c.didHello = true
		if ok, _ := c.extension("8BITMIME"); !ok {
			t.Fatalf("Should support 8BITMIME")
		}
		if ok, _ := c.extension("SMTPUTF8"); !ok {
			t.Fatalf("Should support SMTPUTF8")
		}
		if err := c.mail("user+📧@gmail.com"); err != nil {
			t.Fatalf("MAIL FROM failed: %s", err)
		}
		if err := c.quit(); err != nil {
			t.Fatalf("QUIT failed: %s", err)
		}

		bcmdbuf.Flush()
		actualcmds := cmdbuf.String()
		client := strings.Join(strings.Split(basicClient, "\n"), "\r\n")
		if client != actualcmds {
			t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
		}
	})
}

func TestNewClient(t *testing.T) {
	server := strings.Join(strings.Split(newClientServer, "\n"), "\r\n")
	client := strings.Join(strings.Split(newClientClient, "\n"), "\r\n")

	var cmdbuf bytes.Buffer
	bcmdbuf := bufio.NewWriter(&cmdbuf)
	out := func() string {
		bcmdbuf.Flush()
		return cmdbuf.String()
	}
	var fake faker
	fake.ReadWriter = bufio.NewReadWriter(bufio.NewReader(strings.NewReader(server)), bcmdbuf)
	c, err := newClient(fake, "fake.host")
	if err != nil {
		t.Fatalf("NewClient: %v\n(after %v)", err, out())
	}
	defer c.close()
	if ok, args := c.extension("aUtH"); !ok || args != "LOGIN PLAIN" {
		t.Fatalf("Expected AUTH supported")
	}
	if ok, _ := c.extension("DSN"); ok {
		t.Fatalf("Shouldn't support DSN")
	}
	if err := c.quit(); err != nil {
		t.Fatalf("QUIT failed: %s", err)
	}

	actualcmds := out()
	if client != actualcmds {
		t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
	}
}

var newClientServer = `220 hello world
250-mx.google.com at your service
250-SIZE 35651584
250-AUTH LOGIN PLAIN
250 8BITMIME
221 OK
`

var newClientClient = `EHLO localhost
QUIT
`

func TestNewClient2(t *testing.T) {
	server := strings.Join(strings.Split(newClient2Server, "\n"), "\r\n")
	client := strings.Join(strings.Split(newClient2Client, "\n"), "\r\n")

	var cmdbuf bytes.Buffer
	bcmdbuf := bufio.NewWriter(&cmdbuf)
	var fake faker
	fake.ReadWriter = bufio.NewReadWriter(bufio.NewReader(strings.NewReader(server)), bcmdbuf)
	c, err := newClient(fake, "fake.host")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer c.close()
	if ok, _ := c.extension("DSN"); ok {
		t.Fatalf("Shouldn't support DSN")
	}
	if err := c.quit(); err != nil {
		t.Fatalf("QUIT failed: %s", err)
	}

	bcmdbuf.Flush()
	actualcmds := cmdbuf.String()
	if client != actualcmds {
		t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
	}
}

var newClient2Server = `220 hello world
502 EH?
250-mx.google.com at your service
250-SIZE 35651584
250-AUTH LOGIN PLAIN
250 8BITMIME
221 OK
`

var newClient2Client = `EHLO localhost
HELO localhost
QUIT
`

func TestNewClientWithTLS(t *testing.T) {
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		t.Fatalf("loadcert: %v", err)
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &config)
	if err != nil {
		ln, err = tls.Listen("tcp", "[::1]:0", &config)
		if err != nil {
			t.Fatalf("server: listen: %v", err)
		}
	}

	go func() {
		conn, errAccept := ln.Accept()
		if errAccept != nil {
			t.Errorf("server: accept: %v", errAccept)
			return
		}
		defer conn.Close()

		_, errWrite := conn.Write([]byte("220 SIGNS\r\n"))
		if errWrite != nil {
			t.Errorf("server: write: %v", errWrite)
			return
		}
	}()

	config.InsecureSkipVerify = true
	conn, err := tls.Dial("tcp", ln.Addr().String(), &config)
	if err != nil {
		t.Fatalf("client: dial: %v", err)
	}
	defer conn.Close()

	client, err := newClient(conn, ln.Addr().String())
	if err != nil {
		t.Fatalf("smtp: newclient: %v", err)
	}
	if !client.tls {
		t.Errorf("client.tls Got: %t Expected: %t", client.tls, true)
	}
}

func TestHello(t *testing.T) {

	if len(helloServer) != len(helloClient) {
		t.Fatalf("Hello server and client size mismatch")
	}

	for i := 0; i < len(helloServer); i++ {
		server := strings.Join(strings.Split(baseHelloServer+helloServer[i], "\n"), "\r\n")
		client := strings.Join(strings.Split(baseHelloClient+helloClient[i], "\n"), "\r\n")
		var cmdbuf bytes.Buffer
		bcmdbuf := bufio.NewWriter(&cmdbuf)
		var fake faker
		fake.ReadWriter = bufio.NewReadWriter(bufio.NewReader(strings.NewReader(server)), bcmdbuf)
		c, err := newClient(fake, "fake.host")
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		defer c.close()
		c.localName = "customhost"
		err = nil

		switch i {
		case 0:
			err = c.hi("hostinjection>\n\rDATA\r\nInjected message body\r\n.\r\nQUIT\r\n")
			if err == nil {
				t.Errorf("Expected Hello to be rejected due to a message injection attempt")
			}
			err = c.hi("customhost")
		case 1:
			err = c.startTLS(nil)
			if err.Error() == "502 Not implemented" {
				err = nil
			}
		case 2:
			err = c.verify("test@example.com")
		case 3:
			c.tls = true
			c.serverName = "smtp.google.com"
			err = c.authenticate(plainAuthfn("", "user", "pass", "smtp.google.com"))
		case 4:
			err = c.mail("test@example.com")
		case 5:
			ok, _ := c.extension("feature")
			if ok {
				t.Errorf("Expected FEATURE not to be supported")
			}
		case 6:
			err = c.reset()
		case 7:
			err = c.quit()
		case 8:
			err = c.verify("test@example.com")
			if err != nil {
				err = c.hi("customhost")
				if err != nil {
					t.Errorf("Want error, got none")
				}
			}
		case 9:
			err = c.noop()
		default:
			t.Fatalf("Unhandled command")
		}

		if err != nil {
			t.Errorf("Command %d failed: %v", i, err)
		}

		bcmdbuf.Flush()
		actualcmds := cmdbuf.String()
		if client != actualcmds {
			t.Errorf("Got:\n%s\nExpected:\n%s", actualcmds, client)
		}
	}
}

// verify checks the validity of an email address on the server.
// If verify returns nil, the address is valid. A non-nil return
// does not necessarily indicate an invalid address. Many servers
// will not verify addresses for security reasons.
func (c *smtpClient) verify(addr string) error {
	if err := validateLine(addr); err != nil {
		return err
	}
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "VRFY %s", addr)
	return err
}

var baseHelloServer = `220 hello world
502 EH?
250-mx.google.com at your service
250 FEATURE
`

var helloServer = []string{
	"",
	"502 Not implemented\n",
	"250 User is valid\n",
	"235 Accepted\n",
	"250 Sender ok\n",
	"",
	"250 Reset ok\n",
	"221 Goodbye\n",
	"250 Sender ok\n",
	"250 ok\n",
}

var baseHelloClient = `EHLO customhost
HELO customhost
`

var helloClient = []string{
	"",
	"STARTTLS\n",
	"VRFY test@example.com\n",
	"AUTH PLAIN AHVzZXIAcGFzcw==\n",
	"MAIL FROM:<test@example.com>\n",
	"",
	"RSET\n",
	"QUIT\n",
	"VRFY test@example.com\n",
	"NOOP\n",
}

func TestAuthFailed(t *testing.T) {
	server := strings.Join(strings.Split(authFailedServer, "\n"), "\r\n")
	client := strings.Join(strings.Split(authFailedClient, "\n"), "\r\n")
	var cmdbuf bytes.Buffer
	bcmdbuf := bufio.NewWriter(&cmdbuf)
	var fake faker
	fake.ReadWriter = bufio.NewReadWriter(bufio.NewReader(strings.NewReader(server)), bcmdbuf)
	c, err := newClient(fake, "fake.host")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer c.close()

	c.tls = true
	c.serverName = "smtp.google.com"
	err = c.authenticate(plainAuthfn("", "user", "pass", "smtp.google.com"))

	if err == nil {
		t.Error("Auth: expected error; got none")
	} else if err.Error() != "535 Invalid credentials\nplease see www.example.com" {
		t.Errorf("Auth: got error: %v, want: %s", err, "535 Invalid credentials\nplease see www.example.com")
	}

	bcmdbuf.Flush()
	actualcmds := cmdbuf.String()
	if client != actualcmds {
		t.Errorf("Got:\n%s\nExpected:\n%s", actualcmds, client)
	}
}

var authFailedServer = `220 hello world
250-mx.google.com at your service
250 AUTH LOGIN PLAIN
535-Invalid credentials
535 please see www.example.com
221 Goodbye
`

var authFailedClient = `EHLO localhost
AUTH PLAIN AHVzZXIAcGFzcw==
*
QUIT
`

func TestTLSConnState(t *testing.T) {
	ln := newLocalListener(t)
	defer ln.Close()
	clientDone := make(chan bool)
	serverDone := make(chan bool)
	go func() {
		defer close(serverDone)
		c, err := ln.Accept()
		if err != nil {
			t.Errorf("Server accept: %v", err)
			return
		}
		defer c.Close()
		if err := serverHandle(c, t); err != nil {
			t.Errorf("server error: %v", err)
		}
	}()
	go func() {
		defer close(clientDone)
		c, err := dialer(ln.Addr().String())
		if err != nil {
			t.Errorf("Client dial: %v", err)
			return
		}
		defer c.quit()
		cfg := &tls.Config{ServerName: "example.com"}
		testHookStartTLS(cfg) // set the RootCAs
		if err := c.startTLS(cfg); err != nil {
			t.Errorf("StartTLS: %v", err)
			return
		}
		cs, ok := c.tlsConnectionState()
		if !ok {
			t.Errorf("TLSConnectionState returned ok == false; want true")
			return
		}
		if cs.Version == 0 || !cs.HandshakeComplete {
			t.Errorf("ConnectionState = %#v; expect non-zero Version and HandshakeComplete", cs)
		}
	}()
	<-clientDone
	<-serverDone
}

var testHookStartTLS func(*tls.Config)

// dialer returns a new Client connected to an SMTP server at addr.
// The addr must include a port, as in "mail.example.com:smtp".
func dialer(addr string) (*smtpClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return newClient(conn, host)
}

// TLSConnectionState returns the client's TLS connection state.
// The return values are their zero values if StartTLS did
// not succeed.
func (c *smtpClient) tlsConnectionState() (state tls.ConnectionState, ok bool) {
	tc, ok := c.conn.(*tls.Conn)
	if !ok {
		return
	}
	return tc.ConnectionState(), true
}

func newLocalListener(t *testing.T) net.Listener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		ln, err = net.Listen("tcp6", "[::1]:0")
	}
	if err != nil {
		t.Fatal(err)
	}
	return ln
}

type smtpSender struct {
	w io.Writer
}

func (s smtpSender) send(f string) {
	s.w.Write([]byte(f + "\r\n"))
}

// smtp server, finely tailored to deal with our own client only!
func serverHandle(c net.Conn, t *testing.T) error {
	send := smtpSender{c}.send
	send("220 127.0.0.1 ESMTP service ready")
	s := bufio.NewScanner(c)
	for s.Scan() {
		switch s.Text() {
		case "EHLO localhost":
			send("250-127.0.0.1 ESMTP offers a warm hug of welcome")
			send("250-STARTTLS")
			send("250 Ok")
		case "STARTTLS":
			send("220 Go ahead")
			keypair, err := tls.X509KeyPair(localhostCert, localhostKey)
			if err != nil {
				return err
			}
			config := &tls.Config{Certificates: []tls.Certificate{keypair}}
			c = tls.Server(c, config)
			defer c.Close()
			return serverHandleTLS(c, t)
		default:
			t.Fatalf("unrecognized command: %q", s.Text())
		}
	}
	return s.Err()
}

func serverHandleTLS(c net.Conn, t *testing.T) error {
	send := smtpSender{c}.send
	s := bufio.NewScanner(c)
	for s.Scan() {
		switch s.Text() {
		case "EHLO localhost":
			send("250 Ok")
		case "MAIL FROM:<joe1@example.com>":
			send("250 Ok")
		case "RCPT TO:<joe2@example.com>":
			send("250 Ok")
		case "DATA":
			send("354 send the mail data, end with .")
			send("250 Ok")
		case "Subject: test":
		case "":
		case "howdy!":
		case ".":
		case "QUIT":
			send("221 127.0.0.1 Service closing transmission channel")
			return nil
		default:
			t.Fatalf("unrecognized command during TLS: %q", s.Text())
		}
	}
	return s.Err()
}

func init() {
	testRootCAs := x509.NewCertPool()
	testRootCAs.AppendCertsFromPEM(localhostCert)
	testHookStartTLS = func(config *tls.Config) {
		config.RootCAs = testRootCAs
	}
}

// localhostCert is a PEM-encoded TLS cert generated from src/crypto/tls:
// go run generate_cert.go --rsa-bits 1024 --host 127.0.0.1,::1,example.com \
// 		--ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var localhostCert = []byte(`
-----BEGIN CERTIFICATE-----
MIICFDCCAX2gAwIBAgIRAK0xjnaPuNDSreeXb+z+0u4wDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAw
gYkCgYEA0nFbQQuOWsjbGtejcpWz153OlziZM4bVjJ9jYruNw5n2Ry6uYQAffhqa
JOInCmmcVe2siJglsyH9aRh6vKiobBbIUXXUU1ABd56ebAzlt0LobLlx7pZEMy30
LqIi9E6zmL3YvdGzpYlkFRnRrqwEtWYbGBf3znO250S56CCWH2UCAwEAAaNoMGYw
DgYDVR0PAQH/BAQDAgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQF
MAMBAf8wLgYDVR0RBCcwJYILZXhhbXBsZS5jb22HBH8AAAGHEAAAAAAAAAAAAAAA
AAAAAAEwDQYJKoZIhvcNAQELBQADgYEAbZtDS2dVuBYvb+MnolWnCNqvw1w5Gtgi
NmvQQPOMgM3m+oQSCPRTNGSg25e1Qbo7bgQDv8ZTnq8FgOJ/rbkyERw2JckkHpD4
n4qcK27WkEDBtQFlPihIM8hLIuzWoi/9wygiElTy/tVL3y7fGCvY2/k1KBthtZGF
tN8URjVmyEo=
-----END CERTIFICATE-----`)

// localhostKey is the private key for localhostCert.
var localhostKey = []byte(testingKey(`
-----BEGIN RSA TESTING KEY-----
MIICXgIBAAKBgQDScVtBC45ayNsa16NylbPXnc6XOJkzhtWMn2Niu43DmfZHLq5h
AB9+Gpok4icKaZxV7ayImCWzIf1pGHq8qKhsFshRddRTUAF3np5sDOW3QuhsuXHu
lkQzLfQuoiL0TrOYvdi90bOliWQVGdGurAS1ZhsYF/fOc7bnRLnoIJYfZQIDAQAB
AoGBAMst7OgpKyFV6c3JwyI/jWqxDySL3caU+RuTTBaodKAUx2ZEmNJIlx9eudLA
kucHvoxsM/eRxlxkhdFxdBcwU6J+zqooTnhu/FE3jhrT1lPrbhfGhyKnUrB0KKMM
VY3IQZyiehpxaeXAwoAou6TbWoTpl9t8ImAqAMY8hlULCUqlAkEA+9+Ry5FSYK/m
542LujIcCaIGoG1/Te6Sxr3hsPagKC2rH20rDLqXwEedSFOpSS0vpzlPAzy/6Rbb
PHTJUhNdwwJBANXkA+TkMdbJI5do9/mn//U0LfrCR9NkcoYohxfKz8JuhgRQxzF2
6jpo3q7CdTuuRixLWVfeJzcrAyNrVcBq87cCQFkTCtOMNC7fZnCTPUv+9q1tcJyB
vNjJu3yvoEZeIeuzouX9TJE21/33FaeDdsXbRhQEj23cqR38qFHsF1qAYNMCQQDP
QXLEiJoClkR2orAmqjPLVhR3t2oB3INcnEjLNSq8LHyQEfXyaFfu4U9l5+fRPL2i
jiC0k/9L5dHUsF0XZothAkEA23ddgRs+Id/HxtojqqUT27B8MT/IGNrYsp4DvS/c
qgkeluku4GjxRlDMBuXk94xOBEinUs+p/hwP1Alll80Tpg==
-----END RSA TESTING KEY-----`))

func testingKey(s string) string { return strings.ReplaceAll(s, "TESTING KEY", "PRIVATE KEY") }
