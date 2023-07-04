package mail

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"
)

func TestSendRace(t *testing.T) {
	port := 56666
	port2 := 56667
	timeout := 1 * time.Second

	responses := []string{
		`220 test connected`,
		`250 after helo`,
		`250 after mail from`,
		`250 after rcpt to`,
		`354 after data`,
	}

	startService(port, responses, 5*time.Second)
	startService(port2, responses, 0)

	server := NewSMTPClient()
	server.ConnectTimeout = timeout
	server.SendTimeout = timeout
	server.KeepAlive = false
	server.Host = `localhost`
	server.Port = port

	smtpClient, err := server.Connect()
	if err != nil {
		log.Fatalf("couldn't connect: %s", err.Error())
	}
	defer smtpClient.Close()

	// create another server in other port to test timeouts
	server.Port = port2
	smtpClient2, err := server.Connect()
	if err != nil {
		log.Fatalf("couldn't connect: %s", err.Error())
	}
	defer smtpClient2.Close()

	msg := NewMSG().
		SetFrom(`foo@bar`).
		AddTo(`rcpt@bar`).
		SetSubject("subject").
		SetBody(TextPlain, "body")

	// the smtpClient2 has not timeout, so the err, if exists, should be the last message sent in listener service, otherwise is an error
	err = msg.Send(smtpClient2)
	if err != nil && err.Error() != "221 after quit" {
		log.Fatalf("couldn't send: %s", err.Error())
	}

	// the smtpClient send to listener with the last response is after SendTimeout, so when this error is retorned so test is success.
	err = msg.Send(smtpClient)
	if err != nil && err.Error() != "Mail Error: SMTP Send timed out" {
		log.Fatalf("couldn't send: %s", err.Error())
	}
}

func startService(port int, responses []string, timeout time.Duration) {
	log.Printf("starting service at %d...\n", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("couldn't listen to port %d: %s", port, err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalf("couldn't listen accept the request in port %d", port)
			}
			go respond(conn, responses, timeout)
		}
	}()
}

func respond(conn net.Conn, responses []string, timeout time.Duration) {
	buf := make([]byte, 1024)
	for i, resp := range responses {
		conn.Write([]byte(resp + "\n"))
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("couldn't read data")
			return
		}
		readStr := string(buf[:n])
		log.Printf("READ:%s", string(readStr))

		if i == len(responses)-1 {
			break
		}
	}
	if timeout > 0 {
		time.Sleep(timeout)
	}

	conn.Write([]byte(`221 after quit` + "\n"))
	conn.Close()
}
