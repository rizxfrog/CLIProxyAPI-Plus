package schedule

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/MoeForge/moego/base64"
)

// CurlCommand is one curl invocation converted into a task draft the caller can
// store as-is: Request has already been passed through NormalizeRequest.
type CurlCommand struct {
	// Name is a readable suggestion of the form "<METHOD> <host><path>", at most
	// MaxNameLength characters.
	Name string `json:"name"`
	// Request is the parsed request, normalized and validated.
	Request Request `json:"request"`
}

// CurlError reports one command of a pasted blob that could not be converted. The
// remaining commands of the blob are still converted.
//
// The JSON shape is the one the management API documents, "{index, message}":
// Command is available to Go callers but is not sent, because the caller already
// has the text it pasted.
type CurlError struct {
	// Index is the 0-based position of the command among the blob's commands.
	Index int `json:"index"`
	// Command is the text of the offending command, trimmed.
	Command string `json:"-"`
	// Message describes what could not be understood.
	Message string `json:"message"`
}

// ParseCurls converts a blob of pasted text containing one or more curl commands
// into storable requests. It never fails as a whole: a command it cannot
// understand becomes a CurlError at that command's index and the rest of the blob
// still parses.
//
// Commands are separated by newlines that begin a new curl invocation, by ";", and
// by "&&". Backslash line continuations are honoured and quoting is respected, so a
// separator inside quotes is data.
//
// Supported flags: -X/--request, the positional URL, -H/--header, -d/--data,
// --data-raw, --data-binary, --data-urlencode, --json, -F/--form, --form-string,
// -u/--user, -b/--cookie, -A/--user-agent, -e/--referer, and -L/--location.
// Silently ignored: -s, -S, -i, -I, -v, -k, -g, -o, -w, --compressed, --http1.1,
// --http2, --http2-prior-knowledge, --no-progress-meter, and the long spellings of
// the same flags with or without "=value". Any other flag, a missing argument, a
// missing URL, or a malformed header is reported as a CurlError instead of being
// guessed at.
func ParseCurls(text string) ([]CurlCommand, []CurlError) {
	var commands []CurlCommand
	var failures []CurlError
	for _, statement := range splitCurlStatements(text) {
		command, errCommand := parseCurlStatement(statement)
		if errCommand != nil {
			failures = append(failures, CurlError{
				Index:   statement.index,
				Command: statement.text,
				Message: errCommand.Error(),
			})
			continue
		}
		commands = append(commands, command)
	}
	return commands, failures
}

// curlStatement is one command of the blob, before it is tokenized.
type curlStatement struct {
	index int
	text  string
	// err is set when the text itself cannot be tokenized, such as an
	// unterminated quote.
	err error
}

// splitCurlStatements cuts the blob into commands. It tracks quoting so that a
// separator inside quotes stays data, and it cuts on a newline only when the next
// line starts a new curl invocation, so a pasted multi-line command the operator
// forgot to join with backslashes still parses as one command.
func splitCurlStatements(text string) []curlStatement {
	var statements []curlStatement
	start := 0
	var state byte // 0 outside quotes, '\'' inside single quotes, '"' inside double quotes
	i := 0
	cut := func(end int, errStatement error) {
		raw := strings.TrimSpace(text[start:end])
		if raw == "" && errStatement == nil {
			return
		}
		statements = append(statements, curlStatement{index: len(statements), text: raw, err: errStatement})
	}
	for i = 0; i < len(text); i++ {
		switch state {
		case '\'':
			if text[i] == '\'' {
				state = 0
			}
		case '"':
			switch text[i] {
			case '"':
				state = 0
			case '\\':
				i++
			}
		default:
			switch text[i] {
			case '\'':
				state = '\''
			case '"':
				state = '"'
			case '\\':
				i++
			case ';':
				cut(i, nil)
				start = i + 1
			case '&':
				if i+1 < len(text) && text[i+1] == '&' {
					cut(i, nil)
					i++
					start = i + 1
				}
			case '\n':
				next := i + 1
				for next < len(text) && (text[next] == ' ' || text[next] == '\t' || text[next] == '\r') {
					next++
				}
				if beginsCurlInvocation(text, next) {
					cut(i, nil)
					start = next
					i = next - 1
				}
			}
		}
	}
	if state != 0 {
		cut(len(text), fmt.Errorf("unterminated %s quote", quoteName(state)))
		return statements
	}
	cut(len(text), nil)
	return statements
}

// quoteName names a quote character for an error message.
func quoteName(state byte) string {
	if state == '\'' {
		return "single"
	}
	return "double"
}

// beginsCurlInvocation reports whether the text at offset starts with the curl
// word, which makes a preceding newline a command separator.
func beginsCurlInvocation(text string, offset int) bool {
	if offset+4 > len(text) || !strings.EqualFold(text[offset:offset+4], "curl") {
		return false
	}
	if offset+4 == len(text) {
		return true
	}
	switch text[offset+4] {
	case ' ', '\t', '\r', '\n', '\'', '"', '\\':
		return true
	default:
		return false
	}
}

// tokenizeCurl splits one command into words, resolving quoting and escaping the
// way a POSIX shell does: single quotes are literal, double quotes keep most
// characters while honouring "\" escapes, and a backslash outside quotes escapes
// the next character. Backslash-newline continuations disappear everywhere.
func tokenizeCurl(text string) ([]string, error) {
	var tokens []string
	var token strings.Builder
	started := false
	var state byte
	flush := func() {
		if started {
			tokens = append(tokens, token.String())
			token.Reset()
			started = false
		}
	}
	for i := 0; i < len(text); i++ {
		char := text[i]
		switch state {
		case '\'':
			if char == '\'' {
				state = 0
				continue
			}
			token.WriteByte(char)
		case '"':
			switch {
			case char == '"':
				state = 0
			case char == '\\' && i+1 < len(text):
				next := text[i+1]
				switch {
				case next == '"' || next == '\\' || next == '$' || next == '`':
					token.WriteByte(next)
					i++
				case next == '\n':
					i++
				case next == '\r' && i+2 < len(text) && text[i+2] == '\n':
					i += 2
				default:
					token.WriteByte('\\')
				}
			default:
				token.WriteByte(char)
			}
		default:
			switch {
			case char == ' ' || char == '\t' || char == '\r' || char == '\n':
				flush()
			case char == '\'':
				state = '\''
				started = true
			case char == '"':
				state = '"'
				started = true
			case char == '\\':
				switch {
				case i+1 >= len(text):
					token.WriteByte('\\')
					started = true
				case text[i+1] == '\n':
					i++
				case text[i+1] == '\r' && i+2 < len(text) && text[i+2] == '\n':
					i += 2
				default:
					token.WriteByte(text[i+1])
					started = true
					i++
				}
			default:
				token.WriteByte(char)
				started = true
			}
		}
	}
	if state != 0 {
		return nil, fmt.Errorf("unterminated %s quote", quoteName(state))
	}
	flush()
	return tokens, nil
}

// curlHeader is one header as written, kept in order so a later value for the same
// name wins once the list becomes a map.
type curlHeader struct {
	name  string
	value string
}

// curlOptions accumulates the flags of one command.
type curlOptions struct {
	method      string
	targetURL   string
	hasURL      bool
	headers     []curlHeader
	contentType bool
	data        []string
	jsonBody    bool
	form        []FormField
	follow      bool
	// head records -I/--head, which asks for the headers only.
	head bool
}

// parseCurlStatement turns one tokenized command into a normalized request.
func parseCurlStatement(statement curlStatement) (CurlCommand, error) {
	if statement.err != nil {
		return CurlCommand{}, statement.err
	}
	tokens, errTokens := tokenizeCurl(statement.text)
	if errTokens != nil {
		return CurlCommand{}, errTokens
	}
	if len(tokens) == 0 {
		return CurlCommand{}, fmt.Errorf("command is empty")
	}
	if !isCurlInvocation(tokens[0]) {
		return CurlCommand{}, fmt.Errorf("command %q must start with curl", strings.Join(tokens, " "))
	}
	var options curlOptions
	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		switch {
		case token == "--":
			for _, positional := range tokens[i+1:] {
				if errURL := options.setURL(positional); errURL != nil {
					return CurlCommand{}, errURL
				}
			}
			return options.build()
		case strings.HasPrefix(token, "--"):
			name, inline, hasInline := strings.Cut(token, "=")
			value := func() (string, error) {
				if hasInline {
					return inline, nil
				}
				if i+1 >= len(tokens) {
					return "", fmt.Errorf("%s requires a value", name)
				}
				i++
				return tokens[i], nil
			}
			if errLong := options.applyLongFlag(name, value); errLong != nil {
				return CurlCommand{}, errLong
			}
		case len(token) > 1 && token[0] == '-':
			if errShort := options.applyShortFlags(token[1:], tokens, &i); errShort != nil {
				return CurlCommand{}, errShort
			}
		default:
			if errURL := options.setURL(token); errURL != nil {
				return CurlCommand{}, errURL
			}
		}
	}
	return options.build()
}

// isCurlInvocation reports whether a command word names the curl program, accepting
// a path to it and the Windows executable suffix.
func isCurlInvocation(word string) bool {
	base := word
	if index := strings.LastIndexAny(base, `/\`); index >= 0 {
		base = base[index+1:]
	}
	base = strings.TrimSuffix(base, ".exe")
	return strings.EqualFold(base, "curl")
}

// applyLongFlag handles one long flag, including its ignored spellings.
func (o *curlOptions) applyLongFlag(name string, value func() (string, error)) error {
	switch name {
	case "--request":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		o.method = text
	case "--url":
		// Browsers and API clients export "--url" instead of a positional URL;
		// it names the same thing.
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		return o.setURL(text)
	case "--header":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		return o.addHeaderText(text)
	case "--data", "--data-raw", "--data-binary":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		o.addData(text)
	case "--data-urlencode":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		o.addData(encodeDataURLArgument(text))
	case "--json":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		o.addData(text)
		o.jsonBody = true
	case "--form", "--form-string":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		field, errField := parseFormField(text)
		if errField != nil {
			return errField
		}
		o.form = append(o.form, field)
	case "--user":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		o.setUser(text)
	case "--cookie":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		o.addHeader("Cookie", text)
	case "--user-agent":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		o.addHeader("User-Agent", text)
	case "--referer":
		text, errValue := value()
		if errValue != nil {
			return errValue
		}
		o.addHeader("Referer", text)
	case "--location":
		o.follow = true
	case "--output", "--write-out":
		// The value is a local file or format string, which has no meaning here.
		if _, errValue := value(); errValue != nil {
			return errValue
		}
	case "--head":
		// Asks for the response headers only, which curl performs with HEAD.
		o.head = true
	case "--silent", "--show-error", "--include", "--verbose", "--insecure",
		"--compressed", "--http1.1", "--http2", "--http2-prior-knowledge",
		"--no-progress-meter", "--globoff":
		// Accepted and ignored: they change only curl's own output or transport.
	default:
		return fmt.Errorf("unknown flag %q", name)
	}
	return nil
}

// applyShortFlags handles a cluster of short flags such as "-sL" or "-XPOST",
// advancing the caller's token index past any value it takes from the next token.
// A flag that takes a value absorbs the rest of the cluster, so "-Hfoo: bar" sends
// one header, exactly as curl groups them.
func (o *curlOptions) applyShortFlags(cluster string, tokens []string, index *int) error {
	for position := range len(cluster) {
		flag := cluster[position]
		switch flag {
		case 'L':
			o.follow = true
			continue
		case 'I':
			// -I asks for the response headers only, which curl performs with HEAD.
			o.head = true
			continue
		case 's', 'S', 'i', 'v', 'k', 'g':
			// Accepted and ignored: they change only curl's own output or transport.
			continue
		case 'X', 'H', 'd', 'F', 'u', 'b', 'A', 'e', 'o', 'w':
			// Flags below take a value.
		default:
			return fmt.Errorf("unknown flag -%c", flag)
		}
		rest := cluster[position+1:]
		text := rest
		if text == "" {
			if *index+1 >= len(tokens) {
				return fmt.Errorf("-%c requires a value", flag)
			}
			*index = *index + 1
			text = tokens[*index]
		}
		switch flag {
		case 'X':
			o.method = text
		case 'H':
			if errHeader := o.addHeaderText(text); errHeader != nil {
				return errHeader
			}
		case 'd':
			o.addData(text)
		case 'F':
			field, errField := parseFormField(text)
			if errField != nil {
				return errField
			}
			o.form = append(o.form, field)
		case 'u':
			o.setUser(text)
		case 'b':
			o.addHeader("Cookie", text)
		case 'A':
			o.addHeader("User-Agent", text)
		case 'e':
			o.addHeader("Referer", text)
		case 'o', 'w':
			// The value is a local file or format string, which has no meaning here.
		}
		if rest != "" {
			return nil
		}
	}
	return nil
}

// setURL records the positional URL, rejecting a second one.
func (o *curlOptions) setURL(target string) error {
	if o.hasURL {
		return fmt.Errorf("more than one URL in the command: %q and %q", o.targetURL, target)
	}
	o.targetURL = target
	o.hasURL = true
	return nil
}

// addHeader records one header, ignoring a repeated name in favour of the last.
func (o *curlOptions) addHeader(name, value string) {
	o.headers = append(o.headers, curlHeader{name: name, value: value})
	if strings.EqualFold(name, "Content-Type") {
		o.contentType = true
	}
}

// addHeaderText records a "Name: value" header, the form the -H flag uses.
func (o *curlOptions) addHeaderText(text string) error {
	name, value, found := strings.Cut(text, ":")
	if !found {
		return fmt.Errorf("header %q must be written as 'Name: value'", text)
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("header %q has an empty name", text)
	}
	o.addHeader(trimmed, strings.TrimSpace(value))
	return nil
}

// addData appends one body piece; pieces are joined with "&" the way curl does.
func (o *curlOptions) addData(piece string) {
	o.data = append(o.data, piece)
}

// setUser turns "-u user:password" into a Basic authorization header. A value
// without a colon gets an empty password, which is what curl sends when it cannot
// prompt.
func (o *curlOptions) setUser(user string) {
	name, password, found := strings.Cut(user, ":")
	if !found {
		password = ""
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(name + ":" + password))
	o.addHeader("Authorization", "Basic "+encoded)
}

// build assembles the request, applies curl's implied POST, and normalizes it.
func (o *curlOptions) build() (CurlCommand, error) {
	if !o.hasURL {
		return CurlCommand{}, fmt.Errorf("command has no URL")
	}
	method := o.method
	if method == "" && o.head {
		method = "HEAD"
	}
	if method == "" && (len(o.data) > 0 || len(o.form) > 0) {
		method = "POST"
	}
	var headers map[string]string
	if len(o.headers) > 0 || (o.jsonBody && !o.contentType) {
		headers = make(map[string]string, len(o.headers)+1)
		for _, header := range o.headers {
			headers[header.name] = header.value
		}
	}
	if o.jsonBody && !o.contentType {
		headers["Content-Type"] = "application/json"
	}
	normalized, errNormalize := NormalizeRequest(Request{
		Method:          method,
		URL:             o.targetURL,
		Headers:         headers,
		Body:            strings.Join(o.data, "&"),
		Form:            o.form,
		FollowRedirects: o.follow,
	})
	if errNormalize != nil {
		return CurlCommand{}, errNormalize
	}
	return CurlCommand{Name: curlCommandName(normalized), Request: normalized}, nil
}

// parseFormField splits one -F/--form-string argument into a form entry. A file
// reference is kept as a literal value because the scheduler has no file to read.
func parseFormField(text string) (FormField, error) {
	name, value, found := strings.Cut(text, "=")
	if !found {
		return FormField{}, fmt.Errorf("form field %q must be written as 'name=value'", text)
	}
	if strings.TrimSpace(name) == "" {
		return FormField{}, fmt.Errorf("form field %q has an empty name", text)
	}
	return FormField{Name: name, Value: value}, nil
}

// encodeDataURLArgument applies one of curl's three --data-urlencode forms: a
// "name=value" pair whose value is encoded, a "=value" pair encoded whole, and a
// bare value encoded whole.
func encodeDataURLArgument(argument string) string {
	if remainder, ok := strings.CutPrefix(argument, "="); ok {
		return escapeDataURL(remainder)
	}
	if name, value, ok := strings.Cut(argument, "="); ok {
		return name + "=" + escapeDataURL(value)
	}
	return escapeDataURL(argument)
}

// escapeDataURL percent-encodes everything outside the unreserved set, which is
// what curl does to the argument of --data-urlencode.
func escapeDataURL(value string) string {
	const hexDigits = "0123456789ABCDEF"
	var escaped strings.Builder
	escaped.Grow(len(value))
	for i := 0; i < len(value); i++ {
		char := value[i]
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z', char >= '0' && char <= '9',
			char == '-', char == '.', char == '_', char == '~':
			escaped.WriteByte(char)
		default:
			escaped.WriteByte('%')
			escaped.WriteByte(hexDigits[char>>4])
			escaped.WriteByte(hexDigits[char&0x0f])
		}
	}
	return escaped.String()
}

// curlCommandName suggests a task name from a normalized request.
func curlCommandName(request Request) string {
	target := request.URL
	if parsed, errParse := url.Parse(request.URL); errParse == nil && parsed.Host != "" {
		target = parsed.Host + parsed.Path
	}
	return truncateRunes(request.Method+" "+target, MaxNameLength)
}

// truncateRunes cuts a string to at most limit characters without splitting one.
func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
