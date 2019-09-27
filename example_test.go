package mail

import (
	"testing"
)

//Some variables to connect and the body
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

	host           = "example.com"
	port           = 25
	username       = "test@example.com"
	password       = "examplepass"
	encryptionType = EncryptionTLS
	connectTimeout = 10
	sendTimeout    = 10
)

//TestSendMailWithAttachment send a simple html email
func TestSendMail(t *testing.T) {

	client := NewSMTPClient()

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

	//Create the email message
	email := NewMSG()

	email.SetFrom("From Example <from.email@example.com>").
		AddTo("to.email@example.com").
		SetSubject("New Go Email")

	email.SetBody("text/html", htmlBody)

	//Pass the client to the email message to send it
	err = email.Send(smtpClient)

	if err != nil {
		t.Error("Expected nil, got", err, "sending email")
	}

}

//TestSendMailWithAttachment send email with attachment
func TestSendMailWithAttachment(t *testing.T) {

	client := NewSMTPClient()

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

	//Create the email message
	email := NewMSG()

	email.SetFrom("From Example <from.email@example.com>").
		AddTo("to.email@example.com").
		SetSubject("New Go Email")

	//Some additional options to send
	email.SetSender("sender@sender.com")
	email.SetReplyTo("replyto@reply.com")
	email.SetReturnPath("returnpath@info.com")
	email.AddCc("cc@example.com")
	email.AddBcc("bcccc@example.com")

	email.SetBody("text/html", htmlBody)
	email.AddInline("path/to/inline/Gopher.png")

	//Attach the file
	email.AddAttachment("path/to/file")

	//Attach the file with a base64
	email.AddAttachmentBase64("base64string", "filename")

	//Pass the client to the email message to send it
	err = email.Send(smtpClient)

	if err != nil {
		t.Error("Expected nil, got", err, "sending email")
	}

}

//TestSendMultipleEmails send multiple emails in same connection
func TestSendMultipleEmails(t *testing.T) {

	client := NewSMTPClient()

	//SMTP Client
	client.Host = host
	client.Port = port
	client.Username = username
	client.Password = password
	client.Encryption = encryptionType
	client.ConnectTimeout = connectTimeout
	client.SendTimeout = sendTimeout

	//KeepAlive true because the connection need to be open for multiple emails
	//For avoid inactivity timeout, every 30 second you can send a NO OPERATION command to smtp client
	//use smtpClient.Client.Noop() after 30 second of inactivity in this example
	client.KeepAlive = true

	//Connect to client
	smtpClient, err := client.Connect()

	if err != nil {
		t.Error("Expected nil, got", err, "connecting to client")
	}

	toList := [3]string{"to1@example.com", "to3@example.com", "to4@example.com"}

	for _, to := range toList {
		err = sendEmail(htmlBody, to, smtpClient)
		if err != nil {
			t.Error("Expected nil, got", err, "sending email")
		}
	}

}

func sendEmail(htmlBody string, to string, smtpClient *SMTPClient) error {
	//Create the email message
	email := NewMSG()

	email.SetFrom("From Example <from.email@example.com>").
		AddTo(to).
		SetSubject("New Go Email")

	email.SetBody("text/html", htmlBody)

	//Pass the client to the email message to send it
	err := email.Send(smtpClient)

	return err
}
