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
	err := mail.New().SetFrom("From Example <from@example.com>").AddTo("to@example.com").SetSubject("New Go Email").SetBody("text/plain", "Hello Gophers!").Send("smtp.example.com:25")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Email Sent")
	}
}
```

**More Advanced Usage**

```go
package main

import (
	"fmt"

	"github.com/joegrasse/mail"
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

	email := mail.New()
	email.SetPriority(mail.PriorityHigh)
	email.SetFrom("From Example <from@example.com>").AddTo("to@example.com").AddCc("otherto@example.com").SetSubject("New Go Email")

	email.SetBody("text/plain", "Hello Gophers!")
	email.AddAlternative("text/html", htmlBody)

	email.AddInline("/path/to/image.png", "Gopher.png")

	err := email.Send("smtp.example.com:25")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Email Sent")
	}
}
```