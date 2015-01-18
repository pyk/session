package session

import (
	"testing"
)

// TestCommandValidLine check validity of general syntax of command
// like every command should terminated with <CLRF>, etc.
func TestCommandValidLine(t *testing.T) {
	cases := []struct {
		line  string
		valid bool
		err   error
	}{
		// command should terminated with <CRLF>
		{"", false, syntaxErr},
		{"\r\n", true, nil},
		{"HELLO", false, syntaxErr},
	}

	for _, input := range cases {
		c := command(input.line)
		got, err := c.ValidLine()
		if got != input.valid {
			t.Errorf("%q.ValidLine() == %t, expected %t", input.line, got, input.valid)
		}

		if (err == nil && input.err != nil) || (err != nil && input.err == nil) {
			t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
		}
		if err != nil && input.err != nil {
			if err.Error() != input.err.Error() {
				t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
			}
		}
	}
}

// TestCommandVerb make sure that every c.Ver() extract the correct verb from valid line
func TestCommandVerb(t *testing.T) {
	cases := []struct {
		valid_line, expected_verb string
	}{
		{"\r\n", "\r\n"},
		{"DATA\r\n", "DATA"},
		{"data\r\n", "DATA"},
		{"RSET\r\n", "RSET"},
		{"rset\r\n", "RSET"},
		{"QUIT\r\n", "QUIT"},
		{"quit\r\n", "QUIT"},
		{"NOOP\r\n", "NOOP"},
		{"noop\r\n", "NOOP"},
		{"EHLO some-string\r\n", "EHLO"},
		{"HELO some-string\r\n", "HELO"},
		{"ehlo some-string\r\n", "EHLO"},
		{"helo some-string\r\n", "HELO"},
		{"NOOP some-string\r\n", "NOOP"},
		{"noop some-string\r\n", "NOOP"},
		{"HELP some-string\r\n", "HELP"},
		{"HELP some-string\r\n", "HELP"},
		{"EXPN some-string\r\n", "EXPN"},
		{"expn some-string\r\n", "EXPN"},
		{"VRFY some-string\r\n", "VRFY"},
		{"vrfy some-string\r\n", "VRFY"},

		{"MAIL FROM: some-string\r\n", "MAIL FROM:"},
		{"mail from: some-string\r\n", "MAIL FROM:"},
		{"RCPT TO: some-string\r\n", "RCPT TO:"},
		{"rcpt to: some-string\r\n", "RCPT TO:"},
	}

	for _, input := range cases {
		verb := command(input.valid_line).Verb()
		if verb != input.expected_verb {
			t.Errorf("%q: got %q, expected %q", input.valid_line, verb, input.expected_verb)
		}
	}
}

// TestCommandArg make sure that c.Arg() extract the correct argument clause
func TestCommandArg(t *testing.T) {
	cases := []struct {
		valid_line, expected_arg string
	}{
		{"\r\n", ""},
		{"EHLO some-string\r\n", "some-string"},
		{"HELO some-string\r\n", "some-string"},
		{"ehlo some-string\r\n", "some-string"},
		{"helo some-string\r\n", "some-string"},
		{"NOOP some-string\r\n", "some-string"},
		{"noop some-string\r\n", "some-string"},
		{"HELP some-string\r\n", "some-string"},
		{"HELP some-string\r\n", "some-string"},
		{"EXPN some-string\r\n", "some-string"},
		{"expn some-string\r\n", "some-string"},
		{"VRFY some-string\r\n", "some-string"},
		{"vrfy some-string\r\n", "some-string"},

		{"MAIL FROM:<reverse-path> <mail-parameter>\r\n", "<reverse-path> <mail-parameter>"},
		{"mail from: <reverse-path> <mail-parameter>\r\n", "<reverse-path> <mail-parameter>"},

		// TODO: validate RCPT TO arg
		// {"RCPT TO: some-string\r\n", "RCPT TO:"},
		// {"rcpt to: some-string\r\n", "RCPT TO:"},
	}

	for _, input := range cases {
		arg := command(input.valid_line).Arg()
		if arg != input.expected_arg {
			t.Errorf("%q: got %q, expected %q", input.valid_line, arg, input.expected_arg)
		}
	}
}

// TestValidityOfHelloCommand test validity of EHLO & HELO command
func TestValidityOfHelloCommand(t *testing.T) {
	cases := []struct {
		line  string
		valid bool
		err   error
	}{
		// EHLO & HELO should only have one argument
		{"EHLO mail.domain.com\r\n", true, nil},
		{"EHLO mail.domain.com \r\n", true, nil},
		{"EHLO\r\n", false, invalidCommandArgErr},
		{"EHLO \r\n", false, invalidCommandArgErr},
		{"EHLO mail.domain.com test\r\n", false, invalidCommandArgErr},
		{"EHLO mail.domain.com test 1 2 3\r\n", false, invalidCommandArgErr},

		{"HELO mail.domain.com\r\n", true, nil},
		{"HELO mail.domain.com \r\n", true, nil},
		{"HELO\r\n", false, invalidCommandArgErr},
		{"HELO \r\n", false, invalidCommandArgErr},
		{"HELO mail.domain.com test\r\n", false, invalidCommandArgErr},
		{"HELO mail.domain.com test 1 2 3\r\n", false, invalidCommandArgErr},
	}

	for _, input := range cases {
		got, err := command(input.line).ValidHello()
		if got != input.valid {
			t.Errorf("%q.Valid() == %t, expected %t", input.line, got, input.valid)
		}

		if (err == nil && input.err != nil) || (err != nil && input.err == nil) {
			t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
		}
		if err != nil && input.err != nil {
			if err.Error() != input.err.Error() {
				t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
			}
		}
	}
}

// TestCommandEmailAddress make sure that c.EmailAddress() extract the correct email address
func TestCommandEmailAddress(t *testing.T) {
	cases := []struct {
		valid_line, expected_email_addr string
	}{
		{"EHLO ubuntu-trusty\r\n", ""},
		{"HELO ubuntu-trusty\r\n", ""},

		{"MAIL FROM:<some@domain.com>\r\n", "some@domain.com"},
		{"MAIL FROM:<some@domain.com> with-extension\r\n", "some@domain.com"},
		{"MAIL FROM:<some12@sub.domain.com> with-extension\r\n", "some12@sub.domain.com"},
		{"MAIL FROM:<some12-ds@sub.domain.com> with-extension\r\n", "some12-ds@sub.domain.com"},
		{"MAIL FROM:<some_rods@sub.domain.com> with-extension\r\n", "some_rods@sub.domain.com"},
		{"MAIL FROM:<some.another@sub.domain.com> with-extension\r\n", "some.another@sub.domain.com"},
		{"MAIL FROM:<some.anot-her@sub.domain.com> with-extension\r\n", "some.anot-her@sub.domain.com"},

		{"RCPT TO:<some@domain.com>\r\n", "some@domain.com"},
		{"RCPT TO:<@host.io:some@domain.com> with-extension\r\n", "some@domain.com"},
		{"RCPT TO:<@host.io,@abchost.com:some@domain.com> with-extension\r\n", "some@domain.com"},
	}

	for _, input := range cases {
		emailAddr := command(input.valid_line).EmailAddress()
		if emailAddr != input.expected_email_addr {
			t.Errorf("%q: got %q, expected %q", input.valid_line, emailAddr, input.expected_email_addr)
		}
	}
}

// TestValidityOfMailCommand test validity of MAIL command
func TestValidityOfMailCommand(t *testing.T) {
	cases := []struct {
		line  string
		valid bool
		err   error
	}{
		// MAIL FROM validation of reverse-path
		{"MAIL FROM:\r\n", false, syntaxErr},
		{"MAIL FROM: \r\n", false, syntaxErr},
		{"MAIL FROM: some invalid argument\r\n", false, invalidCommandArgErr},
		{"MAIL FROM: some@valid.email.com\r\n", false, invalidCommandArgErr},
		{"MAIL FROM:<invalid-email>\r\n", false, invalidCommandArgErr},
		{"MAIL FROM:<some@valid-email.com,twice@appear.this.com>\r\n", false, invalidCommandArgErr},

		{"MAIL FROM:<some@valid.email.com>\r\n", true, nil},
		{"MAIL FROM: <some@valid.email.com>\r\n", true, nil},
		// TODO: add validation with extension
	}

	for _, input := range cases {
		got, err := command(input.line).ValidMail()
		if got != input.valid {
			t.Errorf("%q.Valid() == %t, expected %t", input.line, got, input.valid)
		}

		if (err == nil && input.err != nil) || (err != nil && input.err == nil) {
			t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
		}
		if err != nil && input.err != nil {
			if err.Error() != input.err.Error() {
				t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
			}
		}
	}
}

// TestValidityOfRcptCommand test validity of RCPT command
func TestValidityOfRcptCommand(t *testing.T) {
	cases := []struct {
		line  string
		valid bool
		err   error
	}{
		// syntax error
		{"RCPT TO:\r\n", false, syntaxErr},
		{"RCPT TO: \r\n", false, syntaxErr},
		{"RCPT TO: some random arg\r\n", false, syntaxErr},

		{"RCPT TO:<invalid formatted email>\r\n", false, invalidRcptEmailErr},
		{"RCPT TO:<some@email.com,some@anothermail.com>\r\n", false, invalidRcptEmailErr},
		{"RCPT TO: <@hostA.io,@hostB.io:some@email.com,some@anohet.com>\r\n", false, invalidRcptEmailErr},

		// source route & email
		// syntaxically valid email addres
		// NOTE: every email address should validated, MUST exist on database
		// {"RCPT TO:<some@email.com>\r\n", true, nil},
		// {"RCPT TO: <some@email.com> \r\n", true, nil},
		// {"RCPT TO:<@hosta.com, @hostb.com:some@email.com> \r\n", true, nil},

	}

	for _, input := range cases {
		got, err := command(input.line).ValidRcpt()
		if got != input.valid {
			t.Errorf("%q.ValidRcpt() == %t, expected %t", input.line, got, input.valid)
		}

		if (err == nil && input.err != nil) || (err != nil && input.err == nil) {
			t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
		}
		if err != nil && input.err != nil {
			if err.Error() != input.err.Error() {
				t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
			}
		}
	}
}

// TestValidityOfDataCommand test validity of DATA command
func TestValidityOfDataCommand(t *testing.T) {
	cases := []struct {
		line  string
		valid bool
		err   error
	}{
		{"DATA\r\n", true, nil},
		{"data\r\n", true, nil},
		{"DATA \r\n", true, nil},

		// syntax error
		{"DATA plus argument\r\n", false, syntaxErr},
	}

	for _, input := range cases {
		got, err := command(input.line).ValidData()
		if got != input.valid {
			t.Errorf("%q.ValidData() == %t, expected %t", input.line, got, input.valid)
		}

		if (err == nil && input.err != nil) || (err != nil && input.err == nil) {
			t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
		}
		if err != nil && input.err != nil {
			if err.Error() != input.err.Error() {
				t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
			}
		}
	}
}

func TestValidityOfQuitCommand(t *testing.T) {
	cases := []struct {
		line  string
		valid bool
		err   error
	}{
		// QUIT should not have an argument
		{"QUIT\r\n", true, nil},
		{"QUIT \r\n", true, nil},
		{"QUIT argument\r\n", false, invalidCommandArgErr},
		{"QUIT some argument \r\n", false, invalidCommandArgErr},
	}

	for _, input := range cases {
		got, err := command(input.line).ValidQuit()
		if got != input.valid {
			t.Errorf("%q.ValidQuit() == %t, expected %t", input.line, got, input.valid)
		}

		if (err == nil && input.err != nil) || (err != nil && input.err == nil) {
			t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
		}
		if err != nil && input.err != nil {
			if err.Error() != input.err.Error() {
				t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
			}
		}
	}
}
