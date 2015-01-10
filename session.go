package session

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

// define replies
const (
	REPLY_220 = "220 Maillennia ready"
	REPLY_221 = "221 OK bye"
	REPLY_440 = "440 Command not received. Please try again"
	REPLY_502 = "502 5.5.1 Unrecognized command."

	// transmitted with error
	REPLY_500 = "500 "
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

func (c command) String() string {
	return string(c)
}

func (c command) Valid() (bool, error) {
	// empty lines
	if c.String() == "" {
		return false, errors.New("Syntax error")
	}

	// every command should terminated with <CRLF>
	if r := strings.HasSuffix(c.String(), "\r\n"); !r {
		return false, errors.New("Command not terminated with <CRLF>")
	}

	return true, nil
}
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
	return "NOT YET"
}

func (c command) Arg() string {
	// TODO: extract argument from command string
	s := strings.Split(string(c), " ")
	if len(s) > 1 {
		return s[1]
	}
	return ""
}

// Session represents session on new connection
type Session struct {
	Conn     net.Conn
	Validity bool
	Reader   *bufio.Reader
	Writer   *bufio.Writer
	Reply    *Reply
}

// New create a new session
func New(conn net.Conn) *Session {
	rp := &Reply{
		w: bufio.NewWriter(conn),
	}

	return &Session{
		Conn:     conn,
		Validity: false,
		Reader:   bufio.NewReader(conn),
		Writer:   bufio.NewWriter(conn),
		Reply:    rp,
	}
}

// Close close the open connection of session
func (s *Session) Close() error {
	err := s.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) SetValid(valid bool) {
	s.Validity = true
}

func (s *Session) Valid() bool {
	return s.Validity
}

// Serve serve connected SMTP sender
func (s *Session) Serve() {
	defer s.Close()

	err := s.Reply.Transmit(REPLY_220)
	if err != nil {
		return
	}

	for {
		// read from connection, return non-escaped string include \r\n
		line, err := s.Reader.ReadString('\n')
		if err != nil {
			err := s.Reply.Transmit(REPLY_440)
			if err != nil {
				return
			}
		}

		c := command(line)
		valid, err := c.Valid()
		if !valid && err != nil {
			// send a reply with custom error
			e := s.Reply.TransmitWithErr(REPLY_500, err)
			if e != nil {
				return
			}
		}

		switch c.Verb() {
		case "\r\n":
			log.Println("enter")
		case "EHLO":
			log.Println(c.Verb())
		case "HELO":
			log.Println(c.Verb())
		case "DATA":
			log.Println(c.Verb())
		case "RSET":
			log.Println(c.Verb())
		case "QUIT":
			log.Println(c.Verb())
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
			err := errors.New("Command unrecognized")
			e := s.Reply.TransmitWithErr(REPLY_500, err)
			if e != nil {
				return
			}
		}

	}
}
