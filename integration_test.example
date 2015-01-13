package session

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

type SMTPserver struct {
	Listener *net.TCPListener
	Ready    chan bool
	Stoped   chan bool
	Wg       *sync.WaitGroup
}

func (smtps *SMTPserver) Run() {
	defer smtps.Wg.Done()
	smtps.Ready <- true

	for {
		select {
		case <-smtps.Stoped:
			// log.Println("smtpserver: stopping listening on", smtps.Listener.Addr())
			smtps.Listener.Close()
			return
		default:
		}

		// make sure listener.AcceptTCP() doesn't block forever
		// so it can read a stopped channel
		smtps.Listener.SetDeadline(time.Now().Add(1e9))
		conn, err := smtps.Listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
		}

		smtps.Wg.Add(1)
		s := New(conn, smtps.Wg, smtps.Stoped)
		go s.Serve()
	}
}

func (smtps *SMTPserver) Stop() {
	close(smtps.Stoped)
	smtps.Wg.Wait()
}

type Client struct {
	Conn   *net.TCPConn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

var server *SMTPserver

func TestMain(m *testing.M) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	server = &SMTPserver{
		Listener: listener,
		Ready:    make(chan bool),
		Stoped:   make(chan bool),
		Wg:       &sync.WaitGroup{},
	}

	server.Wg.Add(1)
	go server.Run()
	<-server.Ready

	os.Exit(m.Run())
}

func TestGreetingAndQuit(t *testing.T) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	client := Client{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
		Writer: bufio.NewWriter(conn),
	}

	greet, err := client.Reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
		return
	}
	greetings := "220 <host> Maillennia ESMTP ready\r\n"
	if greet != greetings {
		t.Errorf("got: %q, expected: %q", greet, greetings)
	}

	fmt.Fprint(client.Writer, "QUIT\r\n")
	client.Writer.Flush()
	reply, err := client.Reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
		return
	}
	quitmsg := "221 <host> OK bye\r\n"
	if reply != quitmsg {
		t.Errorf("got: %q, expected: %q", reply, quitmsg)
	}

	server.Stop()
}
