package session

import (
	"errors"
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
		{"EHLO ubuntu-trusty\r\n", true, nil},
	}

	for _, input := range cases {
		c := command(input.line)
		got, err := c.Valid()
		// check validality
		if got != input.valid {
			t.Errorf("%q.Valid() == %t, expected %t", input.line, got, input.valid)
		}
		// check error message
		if err != input.err {
			if err.Error() != input.err.Error() {
				t.Errorf("got %v, expected %v", err, input.err)
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
