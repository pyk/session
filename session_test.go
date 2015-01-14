package session

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckValidalityOfCommand(t *testing.T) {
	cases := []struct {
		line  string
		valid bool
		err   error
	}{
		// command should terminated with <CRLF>
		{"", false, errors.New("555 5.5.2 Syntax error")},
		{"\r\n", true, nil},
		{"HELLO", false, errors.New("555 5.5.2 Syntax error")},

		// EHLO & HELO should only have one argument
		{"EHLO mail.domain.com\r\n", true, nil},
		{"EHLO mail.domain.com \r\n", true, nil},
		{"EHLO\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"EHLO \r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"EHLO mail.domain.com test\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"EHLO mail.domain.com test 1 2 3\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},

		{"HELO mail.domain.com\r\n", true, nil},
		{"HELO mail.domain.com \r\n", true, nil},
		{"HELO\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"HELO \r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"HELO mail.domain.com test\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"HELO mail.domain.com test 1 2 3\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},

		// MAIL FROM validation of reverse-path
		{"MAIL FROM:\r\n", false, errors.New("555 5.5.2 Syntax error")},
		{"MAIL FROM: \r\n", false, errors.New("555 5.5.2 Syntax error")},
		{"MAIL FROM: some invalid argument\r\n", false, errors.New("555 5.5.2 Syntax error")},
		{"MAIL FROM: some@valid.email.com\r\n", false, errors.New("555 5.5.2 Syntax error")},
		{"MAIL FROM:<invalid-email>\r\n", false, errors.New("555 5.5.2 Syntax error")},
		{"MAIL FROM:<some@valid.email.com>\r\n", true, nil},
		{"MAIL FROM: <some@valid.email.com>\r\n", true, nil},

		// RCPT should have an argument and may have optional params
		{"RCPT TO:\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"RCPT TO: \r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"RCPT TO: forward-path\r\n", true, nil},
		{"RCPT TO: forward-path \r\n", true, nil},
		{"RCPT TO: forward-path optional params\r\n", true, nil},
		{"RCPT TO: forward-path optional params \r\n", true, nil},

		// DATA should not have an argument
		{"DATA\r\n", true, nil},
		{"DATA \r\n", true, nil},
		{"DATA argument\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"DATA some argument \r\n", false, errors.New("501 5.5.4 Invalid command arguments")},

		// RSET should not have an argument
		{"RSET\r\n", true, nil},
		{"RSET \r\n", true, nil},
		{"RSET argument\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"RSET some argument \r\n", false, errors.New("501 5.5.4 Invalid command arguments")},

		// VRFY should have an argument
		{"VRFY mail.domain.com\r\n", true, nil},
		{"VRFY mail.domain.com \r\n", true, nil},
		{"VRFY\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"VRFY \r\n", false, errors.New("501 5.5.4 Invalid command arguments")},

		// EXPN should have an argument
		{"EXPN mail.domain.com\r\n", true, nil},
		{"EXPN mail.domain.com \r\n", true, nil},
		{"EXPN\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"EXPN \r\n", false, errors.New("501 5.5.4 Invalid command arguments")},

		// HELP may have an argument may
		{"HELP\r\n", true, nil},
		{"HELP \r\n", true, nil},
		{"HELP COMMAND\r\n", true, nil},
		{"HELP COMMAND \r\n", true, nil},

		// NOOP argument should be ignored
		{"NOOP\r\n", true, nil},
		{"NOOP \r\n", true, nil},
		{"NOOP COMMAND\r\n", true, nil},
		{"NOOP COMMAND \r\n", true, nil},

		// QUIT should not have an argument
		{"QUIT\r\n", true, nil},
		{"QUIT \r\n", true, nil},
		{"QUIT argument\r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
		{"QUIT some argument \r\n", false, errors.New("501 5.5.4 Invalid command arguments")},
	}

	for _, input := range cases {
		c := command(input.line)

		line := c.String()
		i := len(c.Verb())
		arg := strings.TrimSpace(line[i:])
		args := strings.Split(arg, " ")
		lns := len(args)

		got, err := c.Valid()
		// check validality
		if got != input.valid {
			t.Logf("line: %q, len(c.Verb): %d, arg: %q, args: %v, len(args): %d", line, i, arg, args, lns)
			t.Errorf("%q.Valid() == %t, expected %t", input.line, got, input.valid)
		}

		if (err == nil && input.err != nil) || (err != nil && input.err == nil) {
			t.Logf("line: %q, len(c.Verb): %d, arg: %q, args: %v, len(args): %d", line, i, arg, args, lns)
			t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
		}
		if err != nil && input.err != nil {
			if err.Error() != input.err.Error() {
				t.Logf("line: %q, len(c.Verb): %d, arg: %q, args: %v, len(args): %d", line, i, arg, args, lns)
				t.Errorf("from: %q => got: %v, expected: %v", input.line, err, input.err)
			}
		}
	}
}

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
	}

	for _, input := range cases {
		emailAddr := command(input.valid_line).EmailAddress()
		if emailAddr != input.expected_email_addr {
			t.Errorf("%q: got %q, expected %q", input.valid_line, emailAddr, input.expected_email_addr)
		}
	}
}
