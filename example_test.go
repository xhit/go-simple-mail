package mail

import (
	"testing"
	"time"
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

	host           = "nsaq500awin"
	port           = 25
	username       = "test@example.com"
	password       = "santiago"
	encryptionType = EncryptionNone
	connectTimeout = 10 * time.Second
	sendTimeout    = 10 * time.Second
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

	//smtpClient is a struct that contain *Client, so you can apply to this the commands directly commands

	//NOOP
	smtpClient.Client.Noop()

	if err != nil {
		t.Error("Expected nil, got", err, "connecting to client")
	}

	//Create the email message
	email := NewMSG()

	email.SetFrom("From Example <test@example.com>").
		AddTo("admin@example.com").
		SetSubject("New Go Email")

	email.SetBody("text/html", htmlBody)
	email.AddAlternative("text/plain", "Hello Gophers!")

	//Some additional options to send
	email.SetSender("xhit@test.com")
	email.SetReplyTo("replyto@reply.com")
	email.SetReturnPath("test@example.com")
	email.AddCc("cc@example1.com")
	email.AddBcc("bcccc@example2.com")

	//Add inline too!
	email.AddInline("C:/Users/sdelacruz/Pictures/Gopher.png")

	//Attach a file with path
	email.AddAttachment("C:/Users/sdelacruz/Pictures/Gopher.png")

	//Attach the file with a base64
	email.AddAttachmentBase64("base64string", "filename")

	//Set a different date in header email
	email.SetDate("2015-04-28 10:32:00 CDT")

	//Send with low priority
	email.SetPriority(PriorityLow)

	//Pass the client to the email message to send it
	err = email.Send(smtpClient)

	//Get first error
	email.GetError()

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

	smtpClient.Client.helo()

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

func sendEmail(htmlBody string, to string, smtpClient *SMTPClient) error {
	//Create the email message
	email := NewMSG()

	email.SetFrom("From Example <from.email@example.com>").
		AddTo(to).
		SetSubject("New Go Email")

	//Get from each mail
	email.getFrom()
	email.SetBody("text/html", htmlBody)

	//Send with high priority
	email.SetPriority(PriorityHigh)

	//Pass the client to the email message to send it
	err := email.Send(smtpClient)

	return err
}

func TestPriority(t *testing.T) {
	str := PriorityLow.String()

	if len(str) < 1 {
		t.Error("Expected Low, returned empty string")
	}

	str = PriorityHigh.String()

	if len(str) < 1 {
		t.Error("Expected High, returned empty string")
	}
}
