package codearts

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// ledgerNumber is a credit amount decoded from the plugin statistics ledger.
//
// The route does not keep a stable JSON type for these fields: the same metric
// arrives as an integer (0), a float (100.0) or a numeric string ("100")
// depending on the account's package, and Go's int64 decoder rejects every form
// but the first. Decoding leniently here keeps a formatting change upstream from
// turning a readable balance into "invalid plugin statistics response".
//
// Only integral amounts are meaningful (credits are whole units), so the value is
// rounded toward zero. A non-finite number is treated as absent rather than
// silently becoming a wild balance.
type ledgerNumber int64

// UnmarshalJSON accepts a JSON number, a numeric string, or null.
func (n *ledgerNumber) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*n = 0
		return nil
	}
	// A quoted value: the ledger sometimes stringifies amounts.
	if trimmed[0] == '"' {
		var text string
		if errUnquote := json.Unmarshal(trimmed, &text); errUnquote != nil {
			return errUnquote
		}
		return n.setFromText(text)
	}
	return n.setFromText(string(trimmed))
}

// setFromText parses either an integer or a float literal.
func (n *ledgerNumber) setFromText(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		*n = 0
		return nil
	}
	if integer, errInt := strconv.ParseInt(text, 10, 64); errInt == nil {
		*n = ledgerNumber(integer)
		return nil
	}
	float, errFloat := strconv.ParseFloat(text, 64)
	if errFloat != nil {
		// An unparseable amount is not a balance: leave it absent so the caller
		// reports "unknown" instead of a fabricated zero.
		return errFloat
	}
	if math.IsNaN(float) || math.IsInf(float, 0) {
		*n = 0
		return nil
	}
	*n = ledgerNumber(int64(float))
	return nil
}

// Int64 returns the amount as a plain integer.
func (n ledgerNumber) Int64() int64 { return int64(n) }
