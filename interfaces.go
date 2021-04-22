//

package mail

import (
	"net/textproto"
)

// EmailInterface ...
type EmailInterface interface {
	// GetError returns the first email error encountered
	GetError() error
	// SetFrom sets the From address.
	SetFrom(address string) EmailInterface
	// SetSender sets the Sender address.
	SetSender(address string) EmailInterface
	// SetReplyTo sets the Reply-To address.
	SetReplyTo(address string) EmailInterface
	// SetReturnPath sets the Return-Path address. This is most often used
	// to send bounced emails to a different email address.
	SetReturnPath(address string) EmailInterface
	// AddTo adds a To address. You can provide multiple
	// addresses at the same time.
	AddTo(addresses ...string) EmailInterface
	// AddCc adds a Cc address. You can provide multiple
	// addresses at the same time.
	AddCc(addresses ...string) EmailInterface
	// AddBcc adds a Bcc address. You can provide multiple
	// addresses at the same time.
	AddBcc(addresses ...string) EmailInterface
	// AddAddresses allows you to add addresses to the specified address header.
	AddAddresses(header string, addresses ...string) EmailInterface
	// SetPriority sets the email message priority. Use with
	// either "High" or "Low".
	SetPriority(priority priority) EmailInterface
	// SetDate sets the date header to the provided date/time.
	// The format of the string should be YYYY-MM-DD HH:MM:SS Time Zone.
	//
	// Example: SetDate("2015-04-28 10:32:00 CDT")
	SetDate(dateTime string) EmailInterface
	// SetSubject sets the subject of the email message.
	SetSubject(subject string) EmailInterface
	// SetBody sets the body of the email message.
	SetBody(contentType contentType, body string) EmailInterface
	// AddHeader adds the given "header" with the passed "value".
	AddHeader(header string, values ...string) EmailInterface
	// AddHeaders is used to add multiple headers at once
	AddHeaders(headers textproto.MIMEHeader) EmailInterface
	// AddAlternative allows you to add alternative parts to the body
	// of the email message. This is most commonly used to add an
	// html version in addition to a plain text version that was
	// already added with SetBody.
	AddAlternative(contentType contentType, body string) EmailInterface
	// AddAttachment allows you to add an attachment to the email message.
	// You can optionally provide a different name for the file.
	AddAttachment(file string, name ...string) EmailInterface
	// AddAttachmentData allows you to add an in-memory attachment to the email message.
	AddAttachmentData(data []byte, filename, mimeType string) EmailInterface
	// AddAttachmentBase64 allows you to add an attachment in base64 to the email message.
	// You need provide a name for the file.
	AddAttachmentBase64(b64File string, name string) EmailInterface
	// AddInline allows you to add an inline attachment to the email message.
	// You can optionally provide a different name for the file.
	AddInline(file string, name ...string) EmailInterface
	// AddInlineData allows you to add an inline in-memory attachment to the email message.
	AddInlineData(data []byte, filename, mimeType string) EmailInterface
	// AddInlineBase64 allows you to add an inline in-memory base64 encoded attachment to the email message.
	// You need provide a name for the file. If mimeType is an empty string, attachment mime type will be deduced
	// from the file name extension and defaults to application/octet-stream.
	AddInlineBase64(b64File string, name string, mimeType string) EmailInterface
	// GetFrom returns the sender of the email, if any
	GetFrom() string
	// GetRecipients returns a slice of recipients emails
	GetRecipients() []string
	// GetMessage builds and returns the email message (RFC822 formatted message)
	GetMessage() string
	// Send sends the composed email
	Send(client SMTPClientInterface) error
	// SendEnvelopeFrom sends the composed email with envelope
	// sender. 'from' must be an email address.
	SendEnvelopeFrom(from string, client SMTPClientInterface) error
}

// SMTPClientInterface ...
type SMTPClientInterface interface {
	// Reset send RSET command to smtp client
	Reset() error
	// Noop send NOOP command to smtp client
	Noop() error
	// Quit send QUIT command to smtp client
	Quit() error
	// Close closes the connection
	Close() error
}

// SMTPServerInterface ...
type SMTPServerInterface interface {
	// GetEncryptionType returns the encryption type used to connect to SMTP server
	GetEncryptionType() Encryption
	//Connect returns the smtp client
	Connect() (SMTPClientInterface, error)
}
