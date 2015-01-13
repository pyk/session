package session

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

// define replies
// TODO: add host command flags
const (
	REPLY_220      = "220 <host> Maillennia ESMTP ready"
	REPLY_221      = "221 <host> OK bye"
	REPLY_250      = "250 2.0.0 OK"
	REPLY_421      = "421 4.4.2 Bad connection"
	REPLY_453      = "453 5.3.2 System not accepting network message"
	REPLY_503      = "503 5.5.1 Invalid command"
	REPLY_503_ehlo = "503 5.5.1 HELO/EHLO first"

	// transmitted with error
	REPLY_500 = "500 "
	REPLY_501 = "501 "
)

// reply represents a SMTP Replies
type Reply struct {
	w *bufio.Writer
}

// Transmit send a reply to SMTP sender
func (rp *Reply) Transmit(str string) error {
	fmt.Fprintf(rp.w, "%s\r\n", str)
	err := rp.w.Flush()
	if err != nil {
		return errors.New("Error while send a Reply")
	}
	return nil
}

// TransmitWithErr send a reply to SMTP sender with custom error message
func (rp *Reply) TransmitWithErr(str string, err error) error {
	fmt.Fprintf(rp.w, "%s%s\r\n", str, err.Error())
	e := rp.w.Flush()
	if e != nil {
		return errors.New("Error while send a Reply")
	}
	return nil
}

// command represents a SMTP Commands
type command string

// String return a command as string
func (c command) String() string {
	return string(c)
}

// Valid check validality of command
func (c command) Valid() (bool, error) {
	// empty lines
	if c.String() == "" {
		return false, errors.New("Syntax error")
	}

	// every command should terminated with <CRLF>
	if r := strings.HasSuffix(c.String(), "\r\n"); !r {
		return false, errors.New("Command not terminated with <CRLF>")
	}

	// Validate specific command
	s := strings.Split(c.Arg(), " ")

	// HELO & EHLO should have an argument and not more than one
	if c.Verb() == "EHLO" || c.Verb() == "HELO" {
		if c.Arg() == "" || len(s) > 1 {
			return false, errors.New("5.5.4 Invalid command arguments")
		}
	}

	// RCPT, VRFY & EXPN should have an argument
	if c.Verb() == "RCPT TO:" || c.Verb() == "VRFY" || c.Verb() == "EXPN" {
		if c.Arg() == "" {
			return false, errors.New("5.5.4 Invalid command arguments")
		}
	}

	// DATA, RSET & QUIT should not have an argument
	if c.Verb() == "DATA" || c.Verb() == "RSET" || c.Verb() == "QUIT" {
		if c.Arg() != "" {
			return false, errors.New("5.5.4 Invalid command arguments")
		}
	}

	return true, nil
}

// Verb extract a command verb from line
func (c command) Verb() string {
	if c.String() == "\r\n" {
		return "\r\n"
	}

	verb := strings.TrimSpace(c.String())
	if len(verb) == 4 {
		return strings.ToUpper(verb)
	}
	if len(verb) > 4 {
		i := strings.Index(verb, ":")
		if i > 0 {
			return strings.ToUpper(verb[:i+1])
		} else {
			s := strings.Split(verb, " ")
			return strings.ToUpper(s[0])
		}
	}
	return ""
}

// Arg extract argument from command
func (c command) Arg() string {
	if c.String() == "\r\n" {
		return ""
	}
	line := c.String()
	i := len(c.Verb())
	arg := strings.TrimSpace(line[i:])
	return arg
}

}

// Session represents session on new connection
type Session struct {
	Conn       net.Conn
	Validity   bool
	Reader     *bufio.Reader
	Writer     *bufio.Writer
	Reply      *Reply
	Wg         *sync.WaitGroup
	ChanClosed chan bool
}

// New create a new session
func New(conn net.Conn, wg *sync.WaitGroup, chanclosed chan bool) *Session {
	rp := &Reply{
		w: bufio.NewWriter(conn),
	}

	return &Session{
		Conn:       conn,
		Validity:   false,
		Reader:     bufio.NewReader(conn),
		Writer:     bufio.NewWriter(conn),
		Reply:      rp,
		Wg:         wg,
		ChanClosed: chanclosed,
	}
}

// Close close the open connection of session
func (s *Session) Close() error {
	s.Wg.Done()
	// log.Println("session:", s.Conn.RemoteAddr(), "disconnected")

	err := s.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// SetValid mark a session as valid. Session valid if
// initialized with Hello command
func (s *Session) SetValid(valid bool) {
	s.Validity = true
}

// Valid check a validality of session
func (s *Session) Valid() bool {
	return s.Validity
}

// CheckChanClosed check a channel ChanClosed if received then
// reply with 453 and close the connection
func (s *Session) CheckChanClosed() bool {
	// if signal for close the session received
	// then close the session gracefully
	select {
	case <-s.ChanClosed:
		err := s.Reply.Transmit(REPLY_453)
		if err != nil {
			return true
		}
		return true
	default:
		return false
	}
}

// Serve serve connected SMTP sender
func (s *Session) Serve() {
	defer s.Close()

	// log.Println("session:", s.Conn.RemoteAddr(), "connected")
	err := s.Reply.Transmit(REPLY_220)
	if err != nil {
		return
	}

	// reject connection temporary
	// 421 Service not available
	// when is service not available?
	// in what event occurs?

	for {

		// read from connection, return non-escaped string include \r\n
		line, err := s.Reader.ReadString('\n')
		if err != nil {
			err := s.Reply.Transmit(REPLY_453)
			if err != nil {
				return
			}
		}

		// check signal from smtp server
		chanClosed := s.CheckChanClosed()
		if chanClosed {
			return
		}

		c := command(line)
		valid, err := c.Valid()
		if !valid && err != nil {
			// send a reply with custom error
			e := s.Reply.TransmitWithErr(REPLY_501, err)
			if e != nil {
				return
			}
			continue
		}

		switch c.Verb() {
		case "HELO":
			err := s.Reply.Transmit(REPLY_250)
			if err != nil {
				return
			}
		case "EHLO":
			// TODO: implment Extended SMTP
			err := s.Reply.Transmit(REPLY_250)
			if err != nil {
				return
			}
		case "\r\n":
			log.Println("enter")
		case "DATA":
			log.Println(c.Verb())
		case "RSET":
			log.Println(c.Verb())
		case "QUIT":
			err := s.Reply.Transmit(REPLY_221)
			if err != nil {
				return
			}
			return
		case "NOOP":
			log.Println(c.Verb())
		case "HELP":
			log.Println(c.Verb())
		case "EXPN":
			log.Println(c.Verb())
		case "VRFY":
			log.Println(c.Verb())
		case "MAIL FROM:":
			log.Println(c.Verb())
		case "RCPT TO:":
			log.Println(c.Verb())
		default:
			e := s.Reply.Transmit(REPLY_503)
			if e != nil {
				return
			}
		}

	}
}
