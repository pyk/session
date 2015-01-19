package session

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
)

// define replies
// TODO: add host command flags
const (
	REPLY_220      = "220 <host> Maillennia ESMTP ready"
	REPLY_221      = "221 2.0.0 Bye"
	REPLY_250      = "250 2.0.0 OK"
	REPLY_250_RCPT = "250 2.1.5 OK"
	REPLY_354      = "354 Go ahead"
	REPLY_421      = "421 4.4.2 Bad connection"
	REPLY_453      = "453 5.3.2 System not accepting network message"
	REPLY_503      = "503 5.5.1 Invalid command"
)

// predefined regex
var (
	rArgSyntax = regexp.MustCompile(`<(.+)>`)
	rMailAddr  = regexp.MustCompile(`[a-zA-Z0-9._-]+@(?:[a-zA-Z0-9._-]+\.)+[a-zA-Z]{2,}`)
	rRcptArg   = regexp.MustCompile(`<(?:@(?:[a-zA-Z0-9._-]+\.)+[a-zA-Z]{2,},?)*:?[a-zA-Z0-9._-]+@(?:[a-zA-Z0-9._-]+\.)+[a-zA-Z]{2,}>`)
	rMailArg   = regexp.MustCompile(`<[a-zA-Z0-9._-]+@(?:[a-zA-Z0-9._-]+\.)+[a-zA-Z]{2,}>`)
)

// error replies
var (
	ehloFirstErr         = errors.New("503 5.5.1 HELO/EHLO first")
	badSeqErr            = errors.New("503 5.5.1 Bad sequence of commands") // TODO: improve err reply of bad sequence command
	syntaxErr            = errors.New("555 5.5.2 Syntax error")
	invalidCommandArgErr = errors.New("501 5.5.4 Invalid command arguments")

	invalidRcptEmailErr = errors.New("553-5.1.2 Invalid recipient email address.\r\n" +
		"553-5.1.2 Please Check for any spelling errors\r\n" +
		"553-5.1.2 make sure before & after recipient email address\r\n" +
		"553 5.1.2 doesn't contain periods, spaces, or other punctuation.")

	emailNotExistErr = errors.New("550-5.1.1 Recipient email address doesn't exist.\r\n" +
		"550-5.1.1 Please Check for any spelling errors\r\n" +
		"550-5.1.1 make sure before & after recipient email address\r\n" +
		"550 5.1.1 doesn't contain periods, spaces, or other punctuation.")
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

// TransmitErr send a reply to SMTP sender with custom error message
func (rp *Reply) TransmitErr(err error) error {
	fmt.Fprintf(rp.w, "%s\r\n", err.Error())
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

// Valid check validity general of command. syntax, arg, etc.
func (c command) ValidLine() (bool, error) {
	// empty lines
	if c.String() == "" {
		return false, syntaxErr
	}

	// every command should terminated with <CRLF>
	if r := strings.HasSuffix(c.String(), "\r\n"); !r {
		return false, syntaxErr
	}

	return true, nil
}

// ValidHello check validity of EHLO & HELO command
func (c command) ValidHello() (bool, error) {
	// HELO & EHLO should have an argument and not more than one
	args := strings.Split(c.Arg(), " ")
	if c.Arg() == "" || len(args) > 1 {
		return false, invalidCommandArgErr
	}

	return true, nil
}

// ValidMail check validity of MAIL command
func (c command) ValidMail() (bool, error) {
	if c.Arg() == "" {
		return false, syntaxErr
	}

	if !rMailArg.MatchString(c.Arg()) {
		return false, invalidCommandArgErr
	}

	return true, nil
}

// ValidRcpt check validity of RCPT command
func (c command) ValidRcpt() (bool, error) {

	if c.Arg() == "" || !rArgSyntax.MatchString(c.Arg()) {
		return false, syntaxErr
	}

	if !rRcptArg.MatchString(c.Arg()) {
		return false, invalidRcptEmailErr
	}
	// TODO: email address shoule exist on database

	return true, nil
}

// ValidData check validity of DATA command
func (c command) ValidData() (bool, error) {
	if c.Arg() != "" {
		return false, syntaxErr
	}
	return true, nil
}

// ValidQuit check validity of QUIT command
func (c command) ValidQuit() (bool, error) {
	if c.Verb() == "QUIT" {
		// QUIT should not have an argument
		if c.Arg() != "" {
			return false, invalidCommandArgErr
		}
	}

	return true, nil
}

// EmailAddress extract email address from command arguments
func (c command) EmailAddress() string {
	return rMailAddr.FindString(c.Arg())
}

// Envelopes represents envelope for mail object
// on each session
type Envelope struct {
	OriginatorAddress string
	RecipientAddress  []string
	Extension         string
}

func NewEnvelope() *Envelope {
	return &Envelope{}
}

// SessionValidity represent validity of each session
type SessionValidity struct {
	HeloFirst bool
	MailFirst bool
	RcptFirst bool
}

// Session represents session on new connection
type Session struct {
	Conn       net.Conn
	Validity   *SessionValidity
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

	validity := &SessionValidity{
		HeloFirst: false,
		MailFirst: false,
		RcptFirst: false,
	}

	return &Session{
		Conn:       conn,
		Validity:   validity,
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

// SetHeloFirst mark a session as valid. Session valid if
// initialized with Hello command
func (s *Session) SetHeloFirst(heloFirst bool) {
	s.Validity.HeloFirst = heloFirst
}

// SetMailFirst mark a session as valid if MAIL command
// appear before RCPT command
func (s *Session) SetMailFirst(mailFirst bool) {
	s.Validity.MailFirst = mailFirst
}

// SetRcptFirst mark a session as valid if MAIL command
// appear before RCPT command
func (s *Session) SetRcptFirst(rcptFirst bool) {
	s.Validity.RcptFirst = rcptFirst
}

func (s *Session) Valid(c command) (bool, error) {
	// check validity of line
	_, err := c.ValidLine()
	if err != nil {
		return false, err
	}

	// validation for EHLO & HELO command
	if c.Verb() == "EHLO" || c.Verb() == "HELO" {
		_, err := c.ValidHello()
		if err != nil {
			return false, err
		}

		s.SetHeloFirst(true)
		return true, nil
	}

	// validation for MAIL command
	if c.Verb() == "MAIL FROM:" {
		// MUST appear after EHLO/HELO
		if !s.Validity.HeloFirst {
			return false, ehloFirstErr
		}

		// syntax MUST valid
		_, err := c.ValidMail()
		if err != nil {
			return false, err
		}

		s.SetMailFirst(true)
		return true, nil
	}

	// validation for RCPT command
	if c.Verb() == "RCPT TO:" {
		// MUST appear after EHLO/HELO
		if !s.Validity.HeloFirst {
			return false, ehloFirstErr
		}

		// MUST appear after MAIL
		if !s.Validity.MailFirst {
			return false, badSeqErr
		}

		_, err := c.ValidRcpt()
		if err != nil {
			return false, err
		}

		s.SetRcptFirst(true)
		return true, nil
	}

	// validation for DATA command
	if c.Verb() == "DATA" {
		// MUST appear after EHLO/HELO
		if !s.Validity.HeloFirst {
			return false, ehloFirstErr
		}

		// MUST appear after MAIL & RCPT
		if !s.Validity.MailFirst || !s.Validity.RcptFirst {
			return false, badSeqErr
		}

		_, err := c.ValidData()
		if err != nil {
			return false, err
		}

		return true, nil
	}

	// validation for QUIT command
	if c.Verb() == "QUIT" {
		_, err := c.ValidQuit()
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return true, nil
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

	// create new envelope
	envl := NewEnvelope()

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

		// check validity of session like valid line,
		// command sequences, command syntax, command argument, etc.
		c := command(line)
		valid, err := s.Valid(c)
		if !valid && err != nil {
			// reply with custom error
			e := s.Reply.TransmitErr(err)
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
		case "MAIL FROM:":
			// fill the OriginatorAddress & Extension of envelope here
			envl.OriginatorAddress = c.EmailAddress()
			// envl.Extension = "extension"

			err := s.Reply.Transmit(REPLY_250)
			if err != nil {
				return
			}
		case "RCPT TO:":
			envl.RecipientAddress = append(envl.RecipientAddress, c.EmailAddress())
			err := s.Reply.Transmit(REPLY_250_RCPT)
			if err != nil {
				return
			}
		case "DATA":
			err := s.Reply.Transmit(REPLY_354)
			if err != nil {
				return
			}
			// receive message data here
		case "\r\n":
			log.Println("enter")
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
		default:
			e := s.Reply.Transmit(REPLY_503)
			if e != nil {
				return
			}
		}

	}
}
