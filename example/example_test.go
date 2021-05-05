package example

import (
	"testing"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

// Some variables to connect and the body.
var (
	htmlBody = `<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<title>Hello Gophers!</title>
	</head>
	<body>
		<p>This is the <b>Go gopher</b>.</p>
		<p><img src="cid:Gopher.png" alt="Go gopher" /></p>
		<p>Image created by Renee French</p>
	</body>
</html>`

	host           = "localhost"
	port           = 25
	username       = "test@example.com"
	password       = "santiago"
	encryptionType = mail.EncryptionNone
	connectTimeout = 10 * time.Second
	sendTimeout    = 10 * time.Second
)

// TestSendMailWithAttachment send a simple html email.
func TestSendMail(t *testing.T) {
	client := mail.NewSMTPClient()

	//SMTP Client
	client.Host = host
	client.Port = port
	client.Username = username
	client.Password = password
	client.Encryption = encryptionType
	client.ConnectTimeout = connectTimeout
	client.SendTimeout = sendTimeout
	client.KeepAlive = false

	//Connect to client
	smtpClient, err := client.Connect()

	if err != nil {
		t.Error("Expected nil, got", err, "connecting to client")
	}

	//NOOP command, optional, used for avoid timeout when KeepAlive is true and you aren't sending mails.
	//Execute this command each 30 seconds is ideal for persistent connection
	err = smtpClient.Noop()

	if err != nil {
		t.Error("Expected nil, got", err, "noop to client")
	}

	//Create the email message
	email := mail.NewMSG()

	email.SetFrom("From Example <test@example.com>").
		AddTo("admin@example.com").
		SetSubject("New Go Email")

	email.SetBody(mail.TextHTML, htmlBody)
	email.AddAlternative(mail.TextPlain, "Hello Gophers!")

	//Some additional options to send
	email.SetSender("xhit@test.com")
	email.SetReplyTo("replyto@reply.com")
	email.SetReturnPath("test@example.com")
	email.AddCc("cc@example1.com")
	email.AddBcc("bcccc@example2.com")

	//Add inline too!
	email.Attach(&mail.File{FilePath: "C:/Users/sdelacruz/Pictures/Gopher.png", Inline: true})

	//Attach a file with path
	email.Attach(&mail.File{FilePath: "C:/Users/sdelacruz/Pictures/Gopher.png"})

	//Attach the file with a base64
	email.Attach(&mail.File{B64Data: "Zm9v", Name: "filename"})

	//Set a different date in header email
	email.SetDate("2015-04-28 10:32:00 CDT")

	//Send with low priority
	email.SetPriority(mail.PriorityLow)

	// always check error after send
	if email.Error != nil {
		t.Error("Expected nil, got", email.Error, "generating email")
	}

	//Pass the client to the email message to send it
	err = email.Send(smtpClient)

	//Get first error
	email.GetError()

	if err != nil {
		t.Error("Expected nil, got", err, "sending email")
	}
}

// TestSendMultipleEmails send multiple emails in same connection.
func TestSendMultipleEmails(t *testing.T) {
	client := mail.NewSMTPClient()

	//SMTP Client
	client.Host = host
	client.Port = port
	client.Username = username
	client.Password = password
	client.Encryption = encryptionType
	client.ConnectTimeout = connectTimeout
	client.SendTimeout = sendTimeout

	//For authentication you can use AuthPlain, AuthLogin or AuthCRAMMD5
	client.Authentication = mail.AuthPlain

	//KeepAlive true because the connection need to be open for multiple emails
	//For avoid inactivity timeout, every 30 second you can send a NO OPERATION command to smtp client
	//use smtpClient.Client.Noop() after 30 second of inactivity in this example
	client.KeepAlive = true

	//Connect to client
	smtpClient, err := client.Connect()

	if err != nil {
		t.Error("Expected nil, got", err, "connecting to client")
	}

	toList := [3]string{"to1@example1.com", "to3@example2.com", "to4@example3.com"}

	for _, to := range toList {
		err = sendEmail(htmlBody, to, smtpClient)
		if err != nil {
			t.Error("Expected nil, got", err, "sending email")
		}
	}
}

func sendEmail(htmlBody string, to string, smtpClient *mail.SMTPClient) error {
	//Create the email message
	email := mail.NewMSG()

	email.SetFrom("From Example <from.email@example.com>").
		AddTo(to).
		SetSubject("New Go Email")

	//Get from each mail
	email.GetFrom()
	email.SetBody(mail.TextHTML, htmlBody)

	//Send with high priority
	email.SetPriority(mail.PriorityHigh)

	// always check error after send
	if email.Error != nil {
		return email.Error
	}

	//Pass the client to the email message to send it
	return email.Send(smtpClient)
}

// TestWithTLS using gmail port 587.
func TestWithTLS(t *testing.T) {
	client := mail.NewSMTPClient()

	//SMTP Client
	client.Host = "smtp.gmail.com"
	client.Port = 587
	client.Username = "aaa@gmail.com"
	client.Password = "asdfghh"
	client.Encryption = mail.EncryptionSTARTTLS
	client.ConnectTimeout = 10 * time.Second
	client.SendTimeout = 10 * time.Second

	//KeepAlive is not settted because by default is false

	//Connect to client
	smtpClient, err := client.Connect()

	if err != nil {
		t.Error("Expected nil, got", err, "connecting to client")
	}

	err = sendEmail(htmlBody, "bbb@gmail.com", smtpClient)
	if err != nil {
		t.Error("Expected nil, got", err, "sending email")
	}
}

// TestWithTLS using gmail port 465.
func TestWithSSL(t *testing.T) {
	client := mail.NewSMTPClient()

	//SMTP Client
	client.Host = "smtp.gmail.com"
	client.Port = 465
	client.Username = "aaa@gmail.com"
	client.Password = "asdfghh"
	client.Encryption = mail.EncryptionSSLTLS
	client.ConnectTimeout = 10 * time.Second
	client.SendTimeout = 10 * time.Second

	//KeepAlive is not settted because by default is false

	//Connect to client
	smtpClient, err := client.Connect()

	if err != nil {
		t.Error("Expected nil, got", err, "connecting to client")
	}

	err = sendEmail(htmlBody, "bbb@gmail.com", smtpClient)
	if err != nil {
		t.Error("Expected nil, got", err, "sending email")
	}
}
