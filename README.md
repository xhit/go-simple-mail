# Go Simple Mail

The best way to send emails in Go with SMTP Keep Alive and Timeout for Connect and Send.

<a href="https://goreportcard.com/badge/github.com/xhit/go-simple-mail/v2"><img src="https://goreportcard.com/report/github.com/xhit/go-simple-mail" alt="Go Report Card"></a>
<a href="https://pkg.go.dev/github.com/xhit/go-simple-mail/v2?tab=doc"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white" alt="go.dev"></a>


# IMPORTANT

Examples in this readme is for v2.2.0 and above. 

Versions <= v2.1.3 use strings "text/html" and "text/plain" in `SetBody` and `AddAlternative`.

Examples for versions <= v2.0.0 go here: https://gist.github.com/xhit/54516917473420a8db1b6fff68a21c99

The base of this package is mail package from **Joe Grasse** https://github.com/joegrasse/mail and mime package of same author https://github.com/joegrasse/mime that are not supported since Jun 29, 2018 and Oct 1, 2015 respectively. A lot of changes in Go Simple Mail were sent with not response.

# Introduction

Go Simple Mail is a simple and efficient package to send emails. It is well tested and
documented.

Go Simple Mail can only send emails using an SMTP server. But the API is flexible and it
is easy to implement other methods for sending emails using a local Postfix, an API, etc.

## Features

Go Simple Mail supports:
- Multiple Attachments with path
- Multiple Attachments in base64
- Multiple Recipients
- Priority
- Reply to
- Set other sender
- Set other from
- Embedded images
- HTML and text templates
- Automatic encoding of special characters
- SSL and TLS
- Unencrypted connection (not recommended)
- Sending multiple emails with the same SMTP connection (Keep Alive or Persistent Connection)
- Timeout for connect to a SMTP Server
- Timeout for send an email
- Return Path
- Alternaive Email Body
- CC and BCC
- Add Custom Headers in Message
- Send NOOP, RESET, QUIT and CLOSE to SMTP client
- PLAIN, LOGIN and CRAM-MD5 Authentication ( >= v2.3.0)

## Documentation

https://pkg.go.dev/github.com/xhit/go-simple-mail/v2?tab=doc

## Download

This package use go modules.

```bash
go get github.com/xhit/go-simple-mail/v2
```

# Usage

```go
package main

import (
	"github.com/xhit/go-simple-mail/v2"
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

	/*
	In version >=2.3.0 you can specified authentication type: PLAIN, LOGIN, CRAM-MD5
	if not specified, default is mail.AuthPlain

	code:
	server.Authentication = mail.AuthPlain
	*/
	
	//Variable to keep alive connection
	server.KeepAlive = false
	
	//Timeout for connect to SMTP Server
	server.ConnectTimeout = 10 * time.Second
	
	//Timeout for send the data and wait respond
	server.SendTimeout = 10 * time.Second
	
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

	email.SetBody(mail.TextHTML, htmlBody)

	email.AddInline("/path/to/image.png", "Gopher.png")

	//Call Send and pass the client
	err = email.Send(smtpClient)

	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email Sent")
	}
}
```

## Send multiple emails in same connection

```go
	//Set your smtpClient struct to keep alive connection
	server.KeepAlive = true

	toList := [3]string{"to1@example1.com", "to3@example2.com", "to4@example3.com"}

	for _, to := range toList {
		//New email simple html with inline and CC
		email := mail.NewMSG()

		email.SetFrom("From Example <nube@example.com>").
			AddTo(to).
			SetSubject("New Go Email")

		email.SetBody(mail.TextHTML, htmlBody)

		email.AddInline("/path/to/image.png", "Gopher.png")

		//Call Send and pass the client
		err = email.Send(smtpClient)

		if err != nil {
			log.Println(err)
		} else {
			log.Println("Email Sent")
		}
	}
```