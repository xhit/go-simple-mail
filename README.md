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
