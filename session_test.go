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
		{"", false, errors.New("Syntax error")},
		{"\r\n", true, nil},
		{"HELLO", false, errors.New("Command not terminated with <CRLF>")},

		// EHLO & HELO should only have one argument
		{"EHLO mail.domain.com\r\n", true, nil},
		{"EHLO mail.domain.com \r\n", true, nil},
		{"EHLO\r\n", false, errors.New("5.5.4 Invalid command arguments")},
		{"EHLO \r\n", false, errors.New("5.5.4 Invalid command arguments")},
		{"EHLO mail.domain.com test\r\n", false, errors.New("5.5.4 Invalid command arguments")},
		{"EHLO mail.domain.com test 1 2 3\r\n", false, errors.New("5.5.4 Invalid command arguments")},

		{"HELO mail.domain.com\r\n", true, nil},
		{"HELO mail.domain.com \r\n", true, nil},
		{"HELO\r\n", false, errors.New("5.5.4 Invalid command arguments")},
		{"HELO \r\n", false, errors.New("5.5.4 Invalid command arguments")},
		{"HELO mail.domain.com test\r\n", false, errors.New("5.5.4 Invalid command arguments")},
		{"HELO mail.domain.com test 1 2 3\r\n", false, errors.New("5.5.4 Invalid command arguments")},

		// MAIL may have an reverse-path
		{"MAIL FROM:\r\n", true, nil},
		{"MAIL FROM: \r\n", true, nil},
		{"MAIL FROM: reverse-path\r\n", true, nil},
		{"MAIL FROM: reverse-path \r\n", true, nil},
		{"MAIL FROM: reverse-path optional-param\r\n", true, nil},
		{"MAIL FROM: reverse-path optional-param \r\n", true, nil},

		// RCPT should have an argument and may have optional params
		{"RCPT TO:\r\n", false, errors.New("Syntax error: command should have an argument")},
		{"RCPT TO: \r\n", false, errors.New("Syntax error: command should have an argument")},
		{"RCPT TO: forward-path\r\n", true, nil},
		{"RCPT TO: forward-path \r\n", true, nil},
		{"RCPT TO: forward-path optional params\r\n", true, nil},
		{"RCPT TO: forward-path optional params \r\n", true, nil},

		// DATA should not have an argument
		{"DATA\r\n", true, nil},
		{"DATA \r\n", true, nil},
		{"DATA argument\r\n", false, errors.New("Syntax error: command should not have an argument")},
		{"DATA some argument \r\n", false, errors.New("Syntax error: command should not have an argument")},

		// RSET should not have an argument
		{"RSET\r\n", true, nil},
		{"RSET \r\n", true, nil},
		{"RSET argument\r\n", false, errors.New("Syntax error: command should not have an argument")},
		{"RSET some argument \r\n", false, errors.New("Syntax error: command should not have an argument")},

		// VRFY should have an argument
		{"VRFY mail.domain.com\r\n", true, nil},
		{"VRFY mail.domain.com \r\n", true, nil},
		{"VRFY\r\n", false, errors.New("Syntax error: command should have an argument")},
		{"VRFY \r\n", false, errors.New("Syntax error: command should have an argument")},

		// EXPN should have an argument
		{"EXPN mail.domain.com\r\n", true, nil},
		{"EXPN mail.domain.com \r\n", true, nil},
		{"EXPN\r\n", false, errors.New("Syntax error: command should have an argument")},
		{"EXPN \r\n", false, errors.New("Syntax error: command should have an argument")},

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
		{"QUIT argument\r\n", false, errors.New("Syntax error: command should not have an argument")},
		{"QUIT some argument \r\n", false, errors.New("Syntax error: command should not have an argument")},
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
