The best way to send emails in Go.

**Download**

```bash
go get github.com/joegrasse/mail
```

**Basic Usage**

```go
package main

import (
	"fmt"
	"github.com/joegrasse/mail"
)

func main() {
	var err error
	email := mail.New()

	email.SetFrom("From Example <from@example.com>")
	email.AddTo("to@example.com")
	email.SetSubject("New Go Email")

	email.SetBody("text/plain", "Hello Gophers!")

	html_body :=
		`<html>
	<head>
	   <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
	   <title>Html Email Test!</title>
	</head>
	<body>
	   <p>Hello <b>Gophers</b>!</p>
	   <p><img src="cid:image.jpg" alt="image" /></p>
	</body>
	</html>`
	email.AddAlternative("text/html", html_body)

	email.AddInline("/path/to/image.jpg")

	err = email.Send("smtp.example.com:25")
	
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Email Sent!")
	}
}
```
