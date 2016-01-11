package mail

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/textproto"
	"net/mail"
	"net/smtp"
	"path/filepath"
	"time"
)

// email represents an email message.
type email struct {
	from           string
	sender         string
	replyTo        string
	returnPath     string
	recipients     []string
	headers        textproto.MIMEHeader
	parts          []part
	attachments    []*file
	inlines        []*file
	Charset        string
	Encoding       encoding
	Encryption     encryption
	Username       string
	Password       string
	TLSConfig      *tls.Config
	ConnectTimeout int
}

// part represents the different content parts of an email body.
type part struct {
	contentType string
	body        *bytes.Buffer
}

// file represents the files that can be added to the email message.
type file struct {
	filename string
	mimeType string
	data     []byte
}

type encryption int

const (
	EncryptionTLS encryption = iota
	EncryptionSSL
	EncryptionNone
)

var encryptionTypes = [...]string{"TLS", "SSL", "None"}

func (encryption encryption) String() string {
	return encryptionTypes[encryption]
}

type encoding int

const (
	EncodingQuotedPrintable encoding = iota
	EncodingBase64
	EncodingNone
)

var encodingTypes = [...]string{"quoted-printable", "base64", "binary"}

func (encoding encoding) String() string {
	return encodingTypes[encoding]
}

// New creates a new email. It uses UTF-8 by default.
func New() *email {
	email := &email{
		headers:    make(textproto.MIMEHeader),
		Charset:    "UTF-8",
		Encoding:   EncodingQuotedPrintable,
		Encryption: EncryptionNone,
		TLSConfig:  new(tls.Config),
	}

	email.AddHeader("MIME-Version", "1.0")

	return email
}

// SetFrom sets the From address.
func (email *email) SetFrom(address string) error {
	return email.AddAddresses("From", address)
}

// SetSender sets the Sender address.
func (email *email) SetSender(address string) error {
	return email.AddAddresses("Sender", address)
}

// SetReplyTo sets the Reply-To address.
func (email *email) SetReplyTo(address string) error {
	return email.AddAddresses("Reply-To", address)
}

// SetReturnPath sets the Return-Path address. This is most often used
// to send bounced emails to a different email address.
func (email *email) SetReturnPath(address string) error {
	return email.AddAddresses("Return-Path", address)
}

// AddTo adds a To address. You can provide multiple
// addresses at the same time.
func (email *email) AddTo(addresses ...string) error {
	return email.AddAddresses("To", addresses...)
}

// AddCc adds a Cc address. You can provide multiple
// addresses at the same time.
func (email *email) AddCc(addresses ...string) error {
	return email.AddAddresses("Cc", addresses...)
}

// AddBcc adds a Bcc address. You can provide multiple
// addresses at the same time.
func (email *email) AddBcc(addresses ...string) error {
	return email.AddAddresses("Bcc", addresses...)
}

// AddAddresses allows you to add addresses to the specified address header.
func (email *email) AddAddresses(header string, addresses ...string) error {
	found := false

	// check for a valid address header
	for _, h := range []string{"To", "Cc", "Bcc", "From", "Sender", "Reply-To", "Return-Path"} {
		if header == h {
			found = true
		}
	}

	if !found {
		return errors.New("Mail Error: Invalid address header; Header: [" + header + "]")
	}

	// check to see if the addresses are valid
	for i := range addresses {
		address, err := mail.ParseAddress(addresses[i])
		if err != nil {
			return errors.New("Mail Error: " + err.Error() + "; Header: [" + header + "] Address: [" + addresses[i] + "]")
		}

		// check for more than one address
		switch {
		case header == "From" && len(email.from) > 0:
			fallthrough
		case header == "Sender" && len(email.sender) > 0:
			fallthrough
		case header == "Reply-To" && len(email.replyTo) > 0:
			fallthrough
		case header == "Return-Path" && len(email.returnPath) > 0:
			return errors.New("Mail Error: There can only be one \"" + header + "\" address; Header: [" + header + "] Address: [" + addresses[i] + "]")
		default:
			// other address types can have more than one address
		}

		// save the address
		switch header {
		case "From":
			email.from = address.Address
		case "Sender":
			email.sender = address.Address
		case "Reply-To":
			email.replyTo = address.Address
		case "Return-Path":
			email.returnPath = address.Address
		default:
			// check that the address was added to the recipients list
			email.recipients, err = addAddress(email.recipients, address.Address)
			if err != nil {
				return errors.New("Mail Error: " + err.Error() + "; Header: [" + header + "] Address: [" + addresses[i] + "]")
			}
		}

		// make sure the from and sender addresses are different
		if email.from != "" && email.sender != "" && email.from == email.sender {
			email.sender = ""
			email.headers.Del("Sender")
			return errors.New("Mail Error: From and Sender should not be set to the same address.")
		}

		// add all addresses to the headers except for Bcc and Return-Path
		if header != "Bcc" && header != "Return-Path" {
			// add the address to the headers
			email.headers.Add(header, address.String())
		}
	}

	return nil
}

// addAddress adds an address to the address list if it hasn't already been added
func addAddress(addressList []string, address string) ([]string, error) {
	// loop through the address list to check for dups
	for _, a := range addressList {
		if address == a {
			return addressList, errors.New("Mail Error: Address: [" + address + "] has already been added")
		}
	}

	return append(addressList, address), nil
}

type priority int

const (
	PriorityHigh priority = iota
	PriorityLow
)

var priorities = [...]string{"High", "Low"}

func (priority priority) String() string {
	return priorities[priority]
}

// SetPriority sets the email message priority. Use with
// either "High" or "Low".
func (email *email) SetPriority(priority priority) error {
	switch priority {
	case PriorityHigh:
		return email.AddHeaders(textproto.MIMEHeader{
			"X-Priority":        {"1 (Highest)"},
			"X-MSMail-Priority": {"High"},
			"Importance":        {"High"},
		})
	case PriorityLow:
		return email.AddHeaders(textproto.MIMEHeader{
			"X-Priority":        {"5 (Lowest)"},
			"X-MSMail-Priority": {"Low"},
			"Importance":        {"Low"},
		})
	default:
	}

	return nil
}

// SetDate sets the date header to the provided date/time.
// The format of the string should be YYYY-MM-DD HH:MM:SS Time Zone.
//
// Example: SetDate("2015-04-28 10:32:00 CDT")
func (email *email) SetDate(dateTime string) error {
	const dateFormat = "2006-01-02 15:04:05 MST"

	// Try to parse the provided date/time
	dt, err := time.Parse(dateFormat, dateTime)
	if err != nil {
		return errors.New("Mail Error: Setting date failed with: " + err.Error())
	}

	email.headers.Set("Date", dt.Format(time.RFC1123Z))

	return nil
}

// SetSubject sets the subject of the email message.
func (email *email) SetSubject(subject string) error {
	return email.AddHeader("Subject", subject)
}

// SetBody sets the body of the email message.
func (email *email) SetBody(contentType, body string) {
	email.parts = []part{
		part{
			contentType: contentType,
			body:        bytes.NewBufferString(body),
		},
	}
}

// Header adds the given "header" with the passed "value".
func (email *email) AddHeader(header string, values ...string) error {
	// check that there is actually a value
	if len(values) < 1 {
		return errors.New("Mail Error: no value provided; Header: [" + header + "]")
	}

	switch header {
	case "Sender":
		fallthrough
	case "From":
		fallthrough
	case "To":
		fallthrough
	case "Bcc":
		fallthrough
	case "Cc":
		fallthrough
	case "Reply-To":
		fallthrough
	case "Return-Path":
		return email.AddAddresses(header, values...)
	case "Date":
		if len(values) > 1 {
			return errors.New("Mail Error: To many dates provided")
		}
		return email.SetDate(values[0])
	default:
		email.headers[header] = values
	}

	return nil
}

// Headers is used to add mulitple headers at once
func (email *email) AddHeaders(headers textproto.MIMEHeader) error {
	for header, values := range headers {
		if err := email.AddHeader(header, values...); err != nil {
			return err
		}
	}

	return nil
}

// Alternative allows you to add alternative parts to the body
// of the email message. This is most commonly used to add an
// html version in addition to a plain text version that was
// already added with SetBody.
func (email *email) AddAlternative(contentType, body string) {
	email.parts = append(email.parts,
		part{
			contentType: contentType,
			body:        bytes.NewBufferString(body),
		},
	)
}

// Attach allows you to add an attachment to the email message.
// You can optionally provide a different name for the file.
func (email *email) AddAttachment(file string, name ...string) error {
	if len(name) > 1 {
		return errors.New("Mail Error: Attach can only have a file and an optional name")
	}

	return email.attach(file, false, name...)
}

// Inline allows you to add an inline attachment to the email message.
// You can optionally provide a different name for the file.
func (email *email) AddInline(file string, name ...string) error {
	if len(name) > 1 {
		return errors.New("Mail Error: Inline can only have a file and an optional name")
	}

	return email.attach(file, true, name...)
}

// attach does the low level attaching of the files
func (email *email) attach(f string, inline bool, name ...string) error {
	// Get the file data
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return errors.New("Mail Error: Failed to add file with following error: " + err.Error())
	}

	// get the file mime type
	mimeType := mime.TypeByExtension(filepath.Ext(f))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// get the filename
	_, filename := filepath.Split(f)

	// if an alternative filename was provided, use that instead
	if len(name) == 1 {
		filename = name[0]
	}

	if inline {
		email.inlines = append(email.inlines, &file{
			filename: filename,
			mimeType: mimeType,
			data:     data,
		})
	} else {
		email.attachments = append(email.attachments, &file{
			filename: filename,
			mimeType: mimeType,
			data:     data,
		})
	}

	return nil
}

// getFrom returns the sender of the email, if any
func (email *email) getFrom() string {
	from := email.returnPath
	if from == "" {
		from = email.sender
		if from == "" {
			from = email.from
			if from == "" {
				from = email.replyTo
			}
		}
	}

	return from
}

func (email *email) hasMixedPart() bool {
	return (len(email.parts) > 0 && len(email.attachments) > 0) || len(email.attachments) > 1
}

func (email *email) hasRelatedPart() bool {
	return (len(email.parts) > 0 && len(email.inlines) > 0) || len(email.inlines) > 1
}

func (email *email) hasAlternativePart() bool {
	return len(email.parts) > 1
}

// GetMessage builds and returns the email message
func (email *email) GetMessage() string {
	msg := newMessage(email)

	if email.hasMixedPart() {
		msg.openMultipart("mixed")
	}

	if email.hasRelatedPart() {
		msg.openMultipart("related")
	}

	if email.hasAlternativePart() {
		msg.openMultipart("alternative")
	}

	for _, part := range email.parts {
		msg.addBody(part.contentType, part.body.Bytes())
	}

	if email.hasAlternativePart() {
		msg.closeMultipart()
	}

	msg.addFiles(email.inlines, true)
	if email.hasRelatedPart() {
		msg.closeMultipart()
	}

	msg.addFiles(email.attachments, false)
	if email.hasMixedPart() {
		msg.closeMultipart()
	}

	return msg.getHeaders() + msg.body.String()
}

// Send sends the composed email
func (email *email) Send(address string) error {
	var auth smtp.Auth

	from := email.getFrom()
	if from == "" {
		return errors.New(`Mail Error: No "From" address specified.`)
	}

	if len(email.recipients) < 1 {
		return errors.New("Mail Error: No recipient specified.")
	}

	msg := email.GetMessage()

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return errors.New("Mail Error: " + err.Error())
	}

	if email.Username != "" || email.Password != "" {
		auth = smtp.PlainAuth("", email.Username, email.Password, host)
	}

	return send(host, port, from, email.recipients, msg, auth, email.Encryption, email.TLSConfig, email.ConnectTimeout)
}

// dial connects to the smtp server with the request encryption type
func dial(host string, port string, encryption encryption, config *tls.Config) (*smtp.Client, error) {
	var conn net.Conn
	var err error
	
	address := host + ":" + port
	
	// do the actual dial
	switch encryption {
		case EncryptionSSL:
			conn, err = tls.Dial("tcp", address, config)
		default:
			conn, err = net.Dial("tcp", address)
	}
	
	if err != nil {
		return nil, errors.New("Mail Error on dailing with encryption type " + encryption.String() + ": " + err.Error())
	}

	c, err := smtp.NewClient(conn, host)
	
	if err != nil {
		return nil, errors.New("Mail Error on smtp dial: " + err.Error())
	}
	
	return c, err
}

// smtpConnect connects to the smtp server and starts TLS and passes auth
// if necessary
func smtpConnect(host string, port string, from string, to []string, msg string, auth smtp.Auth, encryption encryption, config *tls.Config) (*smtp.Client, error) {
	// connect to the mail server
	c, err := dial(host, port, encryption, config)

	if err != nil {
		return nil, err
	}

	// send Hello
	if err = c.Hello("localhost"); err != nil {
		c.Close()
		return nil, errors.New("Mail Error on Hello: " + err.Error())
	}

	// start TLS if necessary
	if encryption == EncryptionTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if config.ServerName == "" {
				config = &tls.Config{ServerName: host}
			} 
			
			if err = c.StartTLS(config); err != nil {
				c.Close()
				return nil, errors.New("Mail Error on Start TLS: " + err.Error())
			}
		}
	}

	// pass the authentication if necessary
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				c.Close()
				return nil, errors.New("Mail Error on Auth: " + err.Error())
			}
		}
	}

	return c, nil
}

type smtpConnectErrorChannel struct {
	client *smtp.Client
	err    error
}

// send does the low level sending of the email
func send(host string, port string, from string, to []string, msg string, auth smtp.Auth, encryption encryption, config *tls.Config, connectTimeout int) error {
	var smtpConnectChannel chan smtpConnectErrorChannel
	var c *smtp.Client = nil
	var err error
	
	// set the timeout value
	timeout := time.Duration(connectTimeout) * time.Second

	// if there is a timeout, setup the channel and do the connect under a goroutine
	if timeout != 0 {
		smtpConnectChannel = make(chan smtpConnectErrorChannel, 2)
		go func() {
			c, err = smtpConnect(host, port, from, to, msg, auth, encryption, config)
			// send the result
			smtpConnectChannel <- smtpConnectErrorChannel{
				client: c,
				err:    err,
			}
		}()
	}

	if timeout == 0 {
		// no timeout, just fire the connect
		c, err = smtpConnect(host, port, from, to, msg, auth, encryption, config)
	} else {
		// get the connect result or timeout result, which ever happens first
		select {
		case result := <-smtpConnectChannel:
			c = result.client
			err = result.err
		case <-time.After(timeout):
			return errors.New("Mail Error: SMTP Connection timed out")
		}
	}

	// check for connect error
	if err != nil {
		return err
	}
	
	defer c.Close()

	// Set the sender
	if err := c.Mail(from); err != nil {
		return err
	}

	// Set the recipients
	for _, address := range to {
		if err := c.Rcpt(address); err != nil {
			return err
		}
	}

	// Send the data command
	w, err := c.Data()
	if err != nil {
		return err
	}

	// write the message
	_, err = fmt.Fprint(w, msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}