The best way to send emails in Go with SMTP Keep Alive and Timeout for Connect and Send.

Inspired in joegrasse package github.com/joegrasse/mail Thanks

**IMPORTANT**
This example is for version 2.1.0 and above, for v2.0.0 example go here https://gist.github.com/xhit/54516917473420a8db1b6fff68a21c99

**Download**

```bash
go get -u github.com/xhit/go-simple-mail
```

**Usage**

```go
package main

import (
	"github.com/xhit/go-simple-mail"
	"log"
)

func main() {

	htmlBody :=
`<html>
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

	server := mail.NewSMTPClient()
	
	//SMTP Server
	server.Host = "smtp.example.com"
	server.Port = 587
	server.Username = "test@example.com"
	server.Password = "examplepass"
	server.Encryption = mail.EncryptionTLS
	
	//Variable to keep alive connection
	server.KeepAlive = true
	
	//Timeout for connect to SMTP Server
	server.ConnectTimeout = 10
	
	//Timeout for send the data and wait respond
	server.SendTimeout = 10
	
	//SMTP client
	smtpClient,err :=server.Connect()
	
	if err != nil{
		log.Fatal(err)
	}

	//New email simple html with inline and CC
	email := mail.NewMSG()

	email.SetFrom("From Example <nube@example.com>").
		AddTo("xhit@example.com").
		AddCc("otherto@example.com").
		SetSubject("New Go Email")

	email.SetBody("text/html", htmlBody)

	email.AddInline("/path/to/image.png", "Gopher.png")

	//Call Send and pass the client
	err = email.Send(smtpClient)

	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email Sent")
	}


	//Other email with same connection and attachments
	email = mail.NewMSG()
	
	email.SetFrom("HELLO <nube@example.com>").
		AddTo("xhit@example.com").
		SetSubject("dfgdfgdf")

	email.SetBody("text/plain", "Hello Gophers!")
	email.AddAlternative("text/html", htmlBody)

	email.AddAttachment("path/to/file","filename test")
	email.AddAttachment("path/to/file2")

	// also you can attach a base64 instead a file path
	email.AddAttachmentBase64("SGVsbG8gZ29waGVycyE=", "hello.txt")

	//Call Send and pass the client
	err = email.Send(smtpClient)

	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email Sent")
	}
}
```