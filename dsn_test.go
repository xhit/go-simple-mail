package mail

import (
	"bufio"
	"bytes"
	"net/textproto"
	"strings"
	"testing"
)

func TestDSNString(t *testing.T) {
	tests := []struct {
		dsn  DSN
		want string
	}{
		{NEVER, "NEVER"},
		{FAILURE, "FAILURE"},
		{DELAY, "DELAY"},
		{SUCCESS, "SUCCESS"},
	}
	for _, tc := range tests {
		if got := tc.dsn.String(); got != tc.want {
			t.Errorf("DSN(%d).String() = %q, want %q", int(tc.dsn), got, tc.want)
		}
	}
}

func TestSendMailProcessDSN(t *testing.T) {
	tests := []struct {
		name     string
		hasDSN   bool
		dsn      []DSN
		preserve bool
		wantRCPT string
	}{
		{
			name:     "notify success and failure",
			hasDSN:   true,
			dsn:      []DSN{SUCCESS, FAILURE},
			wantRCPT: "RCPT TO:<rcpt@bar> NOTIFY=SUCCESS,FAILURE",
		},
		{
			name:     "never overrides other notifications",
			hasDSN:   true,
			dsn:      []DSN{SUCCESS, NEVER, FAILURE},
			wantRCPT: "RCPT TO:<rcpt@bar> NOTIFY=NEVER",
		},
		{
			name:     "original recipient preserved",
			hasDSN:   true,
			dsn:      []DSN{DELAY},
			preserve: true,
			wantRCPT: "RCPT TO:<rcpt@bar> NOTIFY=DELAY ORCPT=rfc822;rcpt@bar",
		},
		{
			name:     "no dsn set",
			hasDSN:   true,
			wantRCPT: "RCPT TO:<rcpt@bar>",
		},
		{
			name:     "server without dsn extension",
			hasDSN:   false,
			dsn:      []DSN{SUCCESS},
			wantRCPT: "RCPT TO:<rcpt@bar>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := strings.Join(strings.Split(
				"250 Sender OK\n250 Receiver OK\n354 Go ahead\n250 Data OK\n", "\n"), "\r\n")

			var cmdbuf bytes.Buffer
			bcmdbuf := bufio.NewWriter(&cmdbuf)
			var fake faker
			fake.ReadWriter = bufio.NewReadWriter(bufio.NewReader(strings.NewReader(server)), bcmdbuf)
			sc := &smtpClient{text: textproto.NewConn(fake), localName: "localhost", didHello: true}
			client := &SMTPClient{
				Client:                    sc,
				HasDSNExt:                 tc.hasDSN,
				dsn:                       tc.dsn,
				preserveOriginalRecipient: tc.preserve,
			}

			if err := sendMailProcess("sender@bar", []string{"rcpt@bar"}, "Subject: test\r\n\r\nbody\r\n", client); err != nil {
				t.Fatalf("sendMailProcess failed: %v", err)
			}
			if err := bcmdbuf.Flush(); err != nil {
				t.Fatalf("flushing captured commands failed: %v", err)
			}

			var rcpt string
			for _, line := range strings.Split(cmdbuf.String(), "\r\n") {
				if strings.HasPrefix(line, "RCPT TO:") {
					rcpt = line
				}
			}
			if rcpt != tc.wantRCPT {
				t.Fatalf("RCPT command = %q, want %q (full transcript:\n%s)", rcpt, tc.wantRCPT, cmdbuf.String())
			}
		})
	}
}
