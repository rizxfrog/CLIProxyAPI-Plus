package schedule

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

// curlRequest builds the expected normalized request of a converted command. Every
// expectation therefore carries the defaults NormalizeRequest applies.
func curlRequest(method, target string, headers map[string]string, body string, form []FormField, follow bool) Request {
	return Request{
		Method:          method,
		URL:             target,
		Headers:         headers,
		Body:            body,
		Form:            form,
		TimeoutMS:       DefaultTimeoutMS,
		FollowRedirects: follow,
	}
}

// parseCurlsCase pairs a pasted blob with everything ParseCurls must make of it.
type parseCurlsCase struct {
	name     string
	input    string
	commands []CurlCommand
	problems []CurlError
}

func TestParseCurlsCommands(t *testing.T) {
	cases := []parseCurlsCase{
		{
			name:  "bare url",
			input: "curl https://api.example.com/v1/ping",
			commands: []CurlCommand{{
				Name:    "GET api.example.com/v1/ping",
				Request: curlRequest("GET", "https://api.example.com/v1/ping", nil, "", nil, false),
			}},
		},
		{
			name:  "url without a path",
			input: "curl https://api.example.com",
			commands: []CurlCommand{{
				Name:    "GET api.example.com",
				Request: curlRequest("GET", "https://api.example.com", nil, "", nil, false),
			}},
		},
		{
			name:  "explicit method headers and body",
			input: `curl -X POST 'https://api.example.com/v1/items?draft=1' -H 'Content-Type: application/json' -H 'Authorization: Bearer REDACTED' -d '{"a":1}'`,
			commands: []CurlCommand{{
				Name: "POST api.example.com/v1/items",
				Request: curlRequest("POST", "https://api.example.com/v1/items?draft=1",
					map[string]string{
						"Content-Type":  "application/json",
						"Authorization": "Bearer REDACTED",
					},
					`{"a":1}`, nil, false),
			}},
		},
		{
			name:  "grouped short flags",
			input: `curl -sS -XPOST -H'X-Trace: abc' -d 'a=1' https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name: "POST api.example.com/v1/items",
				Request: curlRequest("POST", "https://api.example.com/v1/items",
					map[string]string{"X-Trace": "abc"}, "a=1", nil, false),
			}},
		},
		{
			name:  "url before the flags",
			input: `curl https://api.example.com/v1/items -X PUT -d 'a=1'`,
			commands: []CurlCommand{{
				Name: "PUT api.example.com/v1/items",
				Request: curlRequest("PUT", "https://api.example.com/v1/items",
					nil, "a=1", nil, false),
			}},
		},
		{
			name:  "data implies post",
			input: `curl -d 'a=1' https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name:    "POST api.example.com/v1/items",
				Request: curlRequest("POST", "https://api.example.com/v1/items", nil, "a=1", nil, false),
			}},
		},
		{
			name:  "repeated data joins with an ampersand",
			input: `curl -d 'a=1' --data 'b=2' https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name:    "POST api.example.com/v1/items",
				Request: curlRequest("POST", "https://api.example.com/v1/items", nil, "a=1&b=2", nil, false),
			}},
		},
		{
			name:  "data raw and data binary",
			input: `curl --data-raw 'raw text' --data-binary 'binary' https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name:    "POST api.example.com/v1/items",
				Request: curlRequest("POST", "https://api.example.com/v1/items", nil, "raw text&binary", nil, false),
			}},
		},
		{
			name:  "data urlencode name equals value",
			input: `curl --data-urlencode 'q=hello world' https://api.example.com/v1/search`,
			commands: []CurlCommand{{
				Name:    "POST api.example.com/v1/search",
				Request: curlRequest("POST", "https://api.example.com/v1/search", nil, "q=hello%20world", nil, false),
			}},
		},
		{
			name:  "data urlencode leading equals encodes the value alone",
			input: `curl --data-urlencode '=hello world' https://api.example.com/v1/search`,
			commands: []CurlCommand{{
				Name:    "POST api.example.com/v1/search",
				Request: curlRequest("POST", "https://api.example.com/v1/search", nil, "hello%20world", nil, false),
			}},
		},
		{
			name:  "data urlencode bare value",
			input: `curl --data-urlencode 'hello world' https://api.example.com/v1/search`,
			commands: []CurlCommand{{
				Name:    "POST api.example.com/v1/search",
				Request: curlRequest("POST", "https://api.example.com/v1/search", nil, "hello%20world", nil, false),
			}},
		},
		{
			name:  "data urlencode reserved characters and multi byte text",
			input: `curl --data-urlencode 'name=a&b' --data-urlencode 'city=São Paulo' https://api.example.com/v1/search`,
			commands: []CurlCommand{{
				Name: "POST api.example.com/v1/search",
				Request: curlRequest("POST", "https://api.example.com/v1/search",
					nil, "name=a%26b&city=S%C3%A3o%20Paulo", nil, false),
			}},
		},
		{
			name:  "data urlencode mixes with data",
			input: `curl -d 'a=1' --data-urlencode 'b=c d' https://api.example.com/v1/search`,
			commands: []CurlCommand{{
				Name:    "POST api.example.com/v1/search",
				Request: curlRequest("POST", "https://api.example.com/v1/search", nil, "a=1&b=c%20d", nil, false),
			}},
		},
		{
			name:  "json sets the body and the content type",
			input: `curl --json '{"a":1}' https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name: "POST api.example.com/v1/items",
				Request: curlRequest("POST", "https://api.example.com/v1/items",
					map[string]string{"Content-Type": "application/json"}, `{"a":1}`, nil, false),
			}},
		},
		{
			name:  "explicit content type beats json",
			input: `curl --json '{"a":1}' -H 'Content-Type: text/plain' https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name: "POST api.example.com/v1/items",
				Request: curlRequest("POST", "https://api.example.com/v1/items",
					map[string]string{"Content-Type": "text/plain"}, `{"a":1}`, nil, false),
			}},
		},
		{
			name:  "form fields imply post",
			input: `curl -F 'file=hello' --form-string 'name=world' https://api.example.com/v1/upload`,
			commands: []CurlCommand{{
				Name: "POST api.example.com/v1/upload",
				Request: curlRequest("POST", "https://api.example.com/v1/upload", nil, "",
					[]FormField{{Name: "file", Value: "hello"}, {Name: "name", Value: "world"}}, false),
			}},
		},
		{
			name:  "basic auth from user and password",
			input: `curl -u 'user:pass' https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name: "GET api.example.com/v1/items",
				Request: curlRequest("GET", "https://api.example.com/v1/items",
					map[string]string{"Authorization": "Basic dXNlcjpwYXNz"}, "", nil, false),
			}},
		},
		{
			name:  "basic auth without a password",
			input: `curl --user user https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name: "GET api.example.com/v1/items",
				Request: curlRequest("GET", "https://api.example.com/v1/items",
					map[string]string{"Authorization": "Basic dXNlcjo="}, "", nil, false),
			}},
		},
		{
			name:  "cookie agent referer and redirects",
			input: `curl -b 'session=abc' -A 'agent/1.0' -e 'https://ref.example.com' -L https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name: "GET api.example.com/v1/items",
				Request: curlRequest("GET", "https://api.example.com/v1/items",
					map[string]string{
						"Cookie":     "session=abc",
						"User-Agent": "agent/1.0",
						"Referer":    "https://ref.example.com",
					}, "", nil, true),
			}},
		},
		{
			name: "long flag spellings",
			input: "curl --request PATCH --header 'X-A: b' --user 'u:p' --cookie 'c=1' " +
				"--user-agent 'agent' --referer 'https://r.example.com' --location https://api.example.com/v1/items",
			commands: []CurlCommand{{
				Name: "PATCH api.example.com/v1/items",
				Request: curlRequest("PATCH", "https://api.example.com/v1/items",
					map[string]string{
						"X-A":           "b",
						"Authorization": "Basic dTpw",
						"Cookie":        "c=1",
						"User-Agent":    "agent",
						"Referer":       "https://r.example.com",
					}, "", nil, true),
			}},
		},
		{
			name: "ignored flags",
			input: "curl -sS -k -i -I -v -g --compressed --http1.1 --http2 --http2-prior-knowledge " +
				"--no-progress-meter --globoff --silent --include --head --verbose --insecure " +
				"-o /dev/null -w '%{http_code}' --output=/dev/null --write-out='%{time_total}' " +
				"https://api.example.com/v1/items",
			commands: []CurlCommand{{
				Name:    "GET api.example.com/v1/items",
				Request: curlRequest("GET", "https://api.example.com/v1/items", nil, "", nil, false),
			}},
		},
		{
			name:  "separators inside quotes are data",
			input: `curl 'https://api.example.com/a?x=1;y=2' -H 'X-Note: a;b' -d 'k=v;w'`,
			commands: []CurlCommand{{
				Name: "POST api.example.com/a",
				Request: curlRequest("POST", "https://api.example.com/a?x=1;y=2",
					map[string]string{"X-Note": "a;b"}, "k=v;w", nil, false),
			}},
		},
		{
			name: "several commands split on newlines continuations semicolons and and",
			input: "curl -X PUT 'https://a.example.com/first' \\\n" +
				"  -H 'X-A: b' \\\n" +
				"  -d 'x=1'\n" +
				"curl https://a.example.com/second; curl 'https://a.example.com/third?q=a;b' && curl -s https://a.example.com/fourth",
			commands: []CurlCommand{
				{
					Name: "PUT a.example.com/first",
					Request: curlRequest("PUT", "https://a.example.com/first",
						map[string]string{"X-A": "b"}, "x=1", nil, false),
				},
				{
					Name:    "GET a.example.com/second",
					Request: curlRequest("GET", "https://a.example.com/second", nil, "", nil, false),
				},
				{
					Name:    "GET a.example.com/third",
					Request: curlRequest("GET", "https://a.example.com/third?q=a;b", nil, "", nil, false),
				},
				{
					Name:    "GET a.example.com/fourth",
					Request: curlRequest("GET", "https://a.example.com/fourth", nil, "", nil, false),
				},
			},
		},
		{
			name: "a multi line command without continuations stays one command",
			input: "curl -X PATCH\n" +
				"  -H 'X-A: b'\n" +
				"  https://a.example.com/patch",
			commands: []CurlCommand{{
				Name: "PATCH a.example.com/patch",
				Request: curlRequest("PATCH", "https://a.example.com/patch",
					map[string]string{"X-A": "b"}, "", nil, false),
			}},
		},
		{
			name:  "double dash ends the flags",
			input: `curl -- https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name:    "GET api.example.com/v1/items",
				Request: curlRequest("GET", "https://api.example.com/v1/items", nil, "", nil, false),
			}},
		},
		{
			name:  "quoted empty body",
			input: `curl -d '' https://api.example.com/v1/items`,
			commands: []CurlCommand{{
				Name:    "POST api.example.com/v1/items",
				Request: curlRequest("POST", "https://api.example.com/v1/items", nil, "", nil, false),
			}},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			commands, problems := ParseCurls(testCase.input)
			if len(problems) != 0 {
				t.Fatalf("ParseCurls(%q) reported problems: %+v", testCase.input, problems)
			}
			if !reflect.DeepEqual(commands, testCase.commands) {
				t.Errorf("ParseCurls(%q)\n got %s\nwant %s", testCase.input,
					curlCommandsString(commands), curlCommandsString(testCase.commands))
			}
		})
	}
}

func TestParseCurlsSkipsEmptyStatements(t *testing.T) {
	for _, input := range []string{"", "   \n\t\n", ";;", " && ; \n ; ", "\n\n\n"} {
		commands, problems := ParseCurls(input)
		if len(commands) != 0 || len(problems) != 0 {
			t.Errorf("ParseCurls(%q) = %d commands, %d problems, want none", input, len(commands), len(problems))
		}
	}
}

// parseCurlsProblemCase pairs a blob that must fail with the problem it must report.
type parseCurlsProblemCase struct {
	name    string
	input   string
	index   int
	command string
	// messageContains is a fragment of the expected message. Messages raised by
	// the request validator live in spec.go, so only their distinctive wording is
	// asserted here; the whole message is asserted for the ones this file raises.
	messageContains string
	commands        int
}

func TestParseCurlsProblems(t *testing.T) {
	cases := []parseCurlsProblemCase{
		{
			name:            "unknown long flag",
			input:           `curl --bogus https://api.example.com/x`,
			command:         `curl --bogus https://api.example.com/x`,
			messageContains: `unknown flag "--bogus"`,
			commands:        0,
		},
		{
			name:            "undocumented long flag",
			input:           `curl --location-trusted https://api.example.com/x`,
			command:         `curl --location-trusted https://api.example.com/x`,
			messageContains: `unknown flag "--location-trusted"`,
			commands:        0,
		},
		{
			name:            "unknown short flag",
			input:           `curl -Z https://api.example.com/x`,
			command:         `curl -Z https://api.example.com/x`,
			messageContains: "unknown flag -Z",
			commands:        0,
		},
		{
			name:            "short flag without a value",
			input:           `curl -H`,
			command:         `curl -H`,
			messageContains: "-H requires a value",
			commands:        0,
		},
		{
			name:            "long flag without a value",
			input:           `curl --header`,
			command:         `curl --header`,
			messageContains: "--header requires a value",
			commands:        0,
		},
		{
			name:            "missing url",
			input:           `curl -X POST`,
			command:         `curl -X POST`,
			messageContains: "command has no URL",
			commands:        0,
		},
		{
			name:            "malformed header",
			input:           `curl -H 'nocolon' https://api.example.com/x`,
			command:         `curl -H 'nocolon' https://api.example.com/x`,
			messageContains: `header "nocolon" must be written as 'Name: value'`,
			commands:        0,
		},
		{
			name:            "header without a name",
			input:           `curl -H ': value' https://api.example.com/x`,
			command:         `curl -H ': value' https://api.example.com/x`,
			messageContains: `header ": value" has an empty name`,
			commands:        0,
		},
		{
			name:            "unsupported scheme",
			input:           `curl ftp://example.com/x`,
			command:         `curl ftp://example.com/x`,
			messageContains: `request.url`,
			commands:        0,
		},
		{
			name:            "relative url",
			input:           `curl api.example.com/x`,
			command:         `curl api.example.com/x`,
			messageContains: `request.url`,
			commands:        0,
		},
		{
			name:            "form field without a value",
			input:           `curl -F 'justname' https://api.example.com/x`,
			command:         `curl -F 'justname' https://api.example.com/x`,
			messageContains: `form field "justname" must be written as 'name=value'`,
			commands:        0,
		},
		{
			name:            "body and form together",
			input:           `curl -d 'a=1' -F 'b=2' https://api.example.com/x`,
			command:         `curl -d 'a=1' -F 'b=2' https://api.example.com/x`,
			messageContains: "mutually exclusive",
			commands:        0,
		},
		{
			name:            "explicit get with a body",
			input:           `curl -X GET -d 'a=1' https://api.example.com/x`,
			command:         `curl -X GET -d 'a=1' https://api.example.com/x`,
			messageContains: "cannot carry a body",
			commands:        0,
		},
		{
			name:            "unterminated single quote",
			input:           `curl 'https://api.example.com/x`,
			command:         `curl 'https://api.example.com/x`,
			messageContains: "unterminated single quote",
			commands:        0,
		},
		{
			name:            "unterminated double quote",
			input:           `curl "https://api.example.com/x`,
			command:         `curl "https://api.example.com/x`,
			messageContains: "unterminated double quote",
			commands:        0,
		},
		{
			name:            "not a curl command",
			input:           `echo hello`,
			command:         `echo hello`,
			messageContains: `command "echo hello" must start with curl`,
			commands:        0,
		},
		{
			name:            "two urls",
			input:           `curl https://a.example.com/1 https://a.example.com/2`,
			command:         `curl https://a.example.com/1 https://a.example.com/2`,
			messageContains: `more than one URL in the command: "https://a.example.com/1" and "https://a.example.com/2"`,
			commands:        0,
		},
		{
			name:            "a broken command does not stop the next one",
			input:           `curl --bogus https://a.example.com/1 && curl -X POST https://a.example.com/2`,
			command:         `curl --bogus https://a.example.com/1`,
			messageContains: `unknown flag "--bogus"`,
			commands:        1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			commands, problems := ParseCurls(testCase.input)
			if len(problems) != 1 {
				t.Fatalf("ParseCurls(%q) reported %d problems (%+v), want 1", testCase.input, len(problems), problems)
			}
			problem := problems[0]
			if problem.Index != testCase.index {
				t.Errorf("problem index = %d, want %d", problem.Index, testCase.index)
			}
			if problem.Command != testCase.command {
				t.Errorf("problem command = %q, want %q", problem.Command, testCase.command)
			}
			if !strings.Contains(problem.Message, testCase.messageContains) {
				t.Errorf("problem message = %q, want it to contain %q", problem.Message, testCase.messageContains)
			}
			if len(commands) != testCase.commands {
				t.Errorf("ParseCurls(%q) converted %d commands, want %d", testCase.input, len(commands), testCase.commands)
			}
		})
	}
}

func TestParseCurlsProblemIndicesFollowCommands(t *testing.T) {
	input := "curl -X POST -d 'a=1' https://a.example.com/ok\n" +
		"curl -H 'broken' https://a.example.com/bad\n" +
		"curl https://a.example.com/ok2"
	commands, problems := ParseCurls(input)
	if len(commands) != 2 {
		t.Fatalf("converted %d commands, want 2", len(commands))
	}
	if len(problems) != 1 {
		t.Fatalf("reported %d problems, want 1", len(problems))
	}
	if problems[0].Index != 1 {
		t.Errorf("problem index = %d, want 1 (the second command)", problems[0].Index)
	}
	if commands[0].Request.URL != "https://a.example.com/ok" || commands[1].Request.URL != "https://a.example.com/ok2" {
		t.Errorf("converted urls = %q, %q", commands[0].Request.URL, commands[1].Request.URL)
	}
}

func TestParseCurlsTruncatesName(t *testing.T) {
	asciiCommand := "curl https://api.example.com/" + strings.Repeat("p", 300)
	asciiCommands, asciiProblems := ParseCurls(asciiCommand)
	if len(asciiProblems) != 0 || len(asciiCommands) != 1 {
		t.Fatalf("ParseCurls returned %d commands and %d problems, want 1 and 0", len(asciiCommands), len(asciiProblems))
	}
	asciiName := asciiCommands[0].Name
	if runes := utf8.RuneCountInString(asciiName); runes != MaxNameLength {
		t.Errorf("name has %d characters, want %d", runes, MaxNameLength)
	}
	if !strings.HasPrefix(asciiName, "GET api.example.com/") {
		t.Errorf("name = %q, want it to start with the method and host", asciiName)
	}

	multibyteCommand := "curl https://api.example.com/" + strings.Repeat("ü", 300)
	multibyteCommands, multibyteProblems := ParseCurls(multibyteCommand)
	if len(multibyteProblems) != 0 || len(multibyteCommands) != 1 {
		t.Fatalf("ParseCurls returned %d commands and %d problems, want 1 and 0", len(multibyteCommands), len(multibyteProblems))
	}
	multibyteName := multibyteCommands[0].Name
	if runes := utf8.RuneCountInString(multibyteName); runes != MaxNameLength {
		t.Errorf("name has %d characters, want %d", runes, MaxNameLength)
	}
	if !utf8.ValidString(multibyteName) {
		t.Errorf("name %q is not valid UTF-8, so it was cut inside a character", multibyteName)
	}
}

// curlCommandsString renders commands for failure messages.
func curlCommandsString(commands []CurlCommand) string {
	var builder strings.Builder
	builder.WriteString("[")
	for index, command := range commands {
		if index > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(command.Name)
		builder.WriteString(" ")
		builder.WriteString(command.Request.Method)
		builder.WriteString(" ")
		builder.WriteString(command.Request.URL)
	}
	builder.WriteString("]")
	return builder.String()
}
