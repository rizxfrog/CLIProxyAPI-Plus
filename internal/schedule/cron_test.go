package schedule

import (
	"strings"
	"testing"
	"time"

	// The embedded time zone database keeps the daylight saving tests independent
	// of the host's copy of it.
	_ "time/tzdata"
)

// cronUTC builds a fixed reference instant for the table-driven tests.
func cronUTC(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

func TestParseCronNormalizesExpression(t *testing.T) {
	cases := []struct {
		name string
		expr string
		want string
	}{
		{"every minute", "* * * * *", "* * * * *"},
		{"daily at nine", "0 9 * * *", "0 9 * * *"},
		{"minute step", "*/15 * * * *", "0,15,30,45 * * * *"},
		{"range step", "0-30/10 * * * *", "0,10,20,30 * * * *"},
		{"sorted list", "15,5,10 * * * *", "5,10,15 * * * *"},
		{"day of month list", "0 0 1,15 * *", "0 0 1,15 * *"},
		{"day of month step", "0 0 */10 * *", "0 0 1,11,21,31 * *"},
		{"month name range", "0 0 * JAN-MAR *", "0 0 * 1-3 *"},
		{"lowercase month name", "0 0 * dec *", "0 0 * 12 *"},
		{"weekday name", "0 0 * * SUN", "0 0 * * 0"},
		{"seven is sunday", "0 0 * * 7", "0 0 * * 0"},
		{"weekday name range", "0 0 * * MON-FRI", "0 0 * * 1-5"},
		{"weekday full alias", "0 0 * * 0-7", "0 0 * * *"},
		{"weekday step", "0 0 * * */2", "0 0 * * 0,2,4,6"},
		{"hour range", "0 8-10 * * *", "0 8-10 * * *"},
		{"surrounding whitespace", "  0   9  *   *   *  ", "0 9 * * *"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			parsed, errParse := ParseCron(testCase.expr)
			if errParse != nil {
				t.Fatalf("ParseCron(%q) returned error: %v", testCase.expr, errParse)
			}
			if got := parsed.String(); got != testCase.want {
				t.Errorf("ParseCron(%q).String() = %q, want %q", testCase.expr, got, testCase.want)
			}
			again, errAgain := ParseCron(parsed.String())
			if errAgain != nil {
				t.Fatalf("ParseCron(%q) of the normalized form returned error: %v", parsed.String(), errAgain)
			}
			if again.String() != testCase.want {
				t.Errorf("normalized %q re-normalized to %q, want %q", testCase.want, again.String(), testCase.want)
			}
		})
	}
}

func TestParseCronRejectsInvalidExpressions(t *testing.T) {
	cases := []struct {
		name  string
		expr  string
		field string
	}{
		{"empty", "", "five fields"},
		{"four fields", "0 9 * *", "five fields"},
		{"six fields", "0 9 * * * *", "five fields"},
		{"minute out of range", "60 * * * *", `cron field minute: "60" is out of range 0-59`},
		{"hour out of range", "* 24 * * *", `cron field hour: "24" is out of range 0-23`},
		{"day of month zero", "* * 0 * *", `cron field day-of-month: "0" is out of range 1-31`},
		{"day of month 32", "* * 32 * *", `cron field day-of-month: "32" is out of range 1-31`},
		{"month zero", "* * * 0 *", `cron field month: "0" is out of range 1-12`},
		{"month 13", "* * * 13 *", `cron field month: "13" is out of range 1-12`},
		{"weekday eight", "* * * * 8", `cron field day-of-week: "8" is out of range 0-7`},
		{"unknown month name", "* * * FOO *", `cron field month: "FOO" is not a number or a known name`},
		{"name in a numeric field", "JAN * * * *", `cron field minute: "JAN" is not a number`},
		{"reversed range", "10-5 * * * *", `cron field minute: range "10-5" is reversed`},
		{"reversed weekday range", "* * * * 7-0", `cron field day-of-week: range "7-0" is reversed`},
		{"zero step", "*/0 * * * *", `cron field minute: step "0" must be greater than zero`},
		{"non numeric step", "*/x * * * *", `cron field minute: step "x" is not a number`},
		{"missing step", "1-5/ * * * *", `cron field minute: "1-5/" is missing a step`},
		{"step without a range", "5/20 * * * *", `cron field minute: "5/20" needs a range or "*" before the step, such as "0-59/20"`},
		{"empty list item", "1,,2 * * * *", `cron field minute: list "1,,2" has an empty item`},
		{"non numeric value", "a * * * *", `cron field minute: "a" is not a number`},
		{"range end missing", "5- * * * *", `cron field minute: "" is not a number`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			parsed, errParse := ParseCron(testCase.expr)
			if errParse == nil {
				t.Fatalf("ParseCron(%q) = %v, want an error mentioning %q", testCase.expr, parsed, testCase.field)
			}
			if !strings.Contains(errParse.Error(), testCase.field) {
				t.Errorf("ParseCron(%q) error = %q, want it to mention %q", testCase.expr, errParse, testCase.field)
			}
		})
	}
}

func TestCronNext(t *testing.T) {
	cases := []struct {
		name  string
		expr  string
		after time.Time
		want  time.Time
	}{
		{
			"later the same day",
			"0 9 * * *",
			cronUTC(2026, time.November, 10, 8, 0),
			cronUTC(2026, time.November, 10, 9, 0),
		},
		{
			"strictly after a matching instant",
			"0 9 * * *",
			cronUTC(2026, time.November, 10, 9, 0),
			cronUTC(2026, time.November, 11, 9, 0),
		},
		{
			"inside the minute before a match",
			"0 9 * * *",
			cronUTC(2026, time.November, 10, 8, 59),
			cronUTC(2026, time.November, 10, 9, 0),
		},
		{
			"minute step",
			"*/20 * * * *",
			cronUTC(2026, time.November, 10, 10, 5),
			cronUTC(2026, time.November, 10, 10, 20),
		},
		{
			"hour step",
			"0 */6 * * *",
			cronUTC(2026, time.November, 10, 7, 0),
			cronUTC(2026, time.November, 10, 12, 0),
		},
		{
			"day of month only",
			"0 0 5 * *",
			cronUTC(2026, time.November, 10, 0, 0),
			cronUTC(2026, time.December, 5, 0, 0),
		},
		{
			"day of week only",
			"0 0 * * 1",
			cronUTC(2026, time.November, 10, 0, 0),
			cronUTC(2026, time.November, 16, 0, 0),
		},
		{
			"either day field, day of week first",
			"0 0 20 * 0",
			cronUTC(2026, time.November, 5, 0, 0),
			cronUTC(2026, time.November, 8, 0, 0),
		},
		{
			"either day field, day of month next",
			"0 0 20 * 0",
			cronUTC(2026, time.November, 15, 0, 0),
			cronUTC(2026, time.November, 20, 0, 0),
		},
		{
			"either day field, day of week after that",
			"0 0 20 * 0",
			cronUTC(2026, time.November, 20, 0, 0),
			cronUTC(2026, time.November, 22, 0, 0),
		},
		{
			"either day field skips a day matching neither",
			"0 0 13 * 5",
			cronUTC(2026, time.November, 11, 0, 0),
			cronUTC(2026, time.November, 13, 0, 0),
		},
		{
			"across a month boundary",
			"0 0 1 * *",
			cronUTC(2026, time.January, 31, 0, 0),
			cronUTC(2026, time.February, 1, 0, 0),
		},
		{
			"across a year boundary",
			"0 0 1 1 *",
			cronUTC(2026, time.June, 15, 12, 0),
			cronUTC(2027, time.January, 1, 0, 0),
		},
		{
			"leap day",
			"0 0 29 2 *",
			cronUTC(2026, time.March, 1, 0, 0),
			cronUTC(2028, time.February, 29, 0, 0),
		},
		{
			"never matches february thirtieth",
			"0 0 30 2 *",
			cronUTC(2026, time.January, 1, 0, 0),
			time.Time{},
		},
		{
			"never matches april thirty first",
			"0 0 31 4 *",
			cronUTC(2026, time.January, 1, 0, 0),
			time.Time{},
		},
		{
			"beyond the search bound",
			"0 0 29 2 *",
			cronUTC(2097, time.March, 1, 0, 0),
			time.Time{},
		},
		{
			"restricted day of month with a step",
			"0 0 */10 * *",
			cronUTC(2026, time.November, 11, 0, 0),
			cronUTC(2026, time.November, 21, 0, 0),
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			parsed, errParse := ParseCron(testCase.expr)
			if errParse != nil {
				t.Fatalf("ParseCron(%q) returned error: %v", testCase.expr, errParse)
			}
			got := parsed.Next(testCase.after)
			if !got.Equal(testCase.want) {
				t.Errorf("Next(%s) for %q = %s, want %s", testCase.after, testCase.expr, got, testCase.want)
			}
			if testCase.want.IsZero() {
				return
			}
			if !got.After(testCase.after) {
				t.Errorf("Next(%s) for %q = %s, want a time after the reference", testCase.after, testCase.expr, got)
			}
			if got.Location() != testCase.after.Location() {
				t.Errorf("Next(%s) for %q = %s in %s, want the reference's location", testCase.after, testCase.expr, got, got.Location())
			}
		})
	}
}

// TestCronNextNeverMatchesWithinBound documents the search bound: an expression
// whose dates cannot be reached inside cronSearchDays resolves to the zero time.
func TestCronNextNeverMatchesWithinBound(t *testing.T) {
	if cronSearchDays != 1830 {
		t.Fatalf("cronSearchDays = %d, want 1830 days (about five years)", cronSearchDays)
	}
	parsed, errParse := ParseCron("0 0 29 2 *")
	if errParse != nil {
		t.Fatalf("ParseCron returned error: %v", errParse)
	}
	// 2100 is not a leap year, so the next February 29 after 2097-03-01 is 2104,
	// which is outside the bound.
	if got := parsed.Next(cronUTC(2097, time.March, 1, 0, 0)); !got.IsZero() {
		t.Errorf("Next across the non-leap century = %s, want the zero time", got)
	}
}

func TestCronNextAcrossDaylightSaving(t *testing.T) {
	newYork, errLoad := time.LoadLocation("America/New_York")
	if errLoad != nil {
		t.Fatalf("LoadLocation(America/New_York) failed: %v", errLoad)
	}

	springForward, errSpring := ParseCron("30 2 * * *")
	if errSpring != nil {
		t.Fatalf("ParseCron(30 2 * * *) returned error: %v", errSpring)
	}
	// 2026-03-08 skips 02:00-03:00, so 02:30 that day never fires.
	gotSpring := springForward.Next(time.Date(2026, time.March, 7, 3, 0, 0, 0, newYork))
	wantSpring := time.Date(2026, time.March, 9, 2, 30, 0, 0, newYork)
	if !gotSpring.Equal(wantSpring) {
		t.Errorf("Next across the spring-forward gap = %s, want %s", gotSpring, wantSpring)
	}
	if gotSpring.Hour() != 2 || gotSpring.Minute() != 30 {
		t.Errorf("Next across the spring-forward gap = %s, want a 02:30 wall clock", gotSpring)
	}

	fallBack, errFall := ParseCron("30 1 * * *")
	if errFall != nil {
		t.Fatalf("ParseCron(30 1 * * *) returned error: %v", errFall)
	}
	firstOccurrence := time.Date(2026, time.November, 1, 1, 30, 0, 0, newYork)
	if firstOccurrence.UTC().Hour() != 5 {
		t.Fatalf("test assumption broken: the first 01:30 on the fall-back day is %s", firstOccurrence.UTC())
	}
	gotFall := fallBack.Next(time.Date(2026, time.October, 31, 2, 0, 0, 0, newYork))
	if !gotFall.Equal(firstOccurrence) {
		t.Errorf("Next across the fall-back overlap = %s (%s UTC), want the first occurrence %s",
			gotFall, gotFall.UTC(), firstOccurrence)
	}
	nextDay := time.Date(2026, time.November, 2, 1, 30, 0, 0, newYork)
	if got := fallBack.Next(firstOccurrence); !got.Equal(nextDay) {
		t.Errorf("Next from the first occurrence = %s, want %s so the repeated hour fires once", got, nextDay)
	}
	if got := fallBack.Next(firstOccurrence.Add(time.Hour)); !got.Equal(nextDay) {
		t.Errorf("Next from the second occurrence = %s, want %s", got, nextDay)
	}
}

func TestParseCronWithinRange(t *testing.T) {
	// A daily schedule carries every minute of the day, and a stepped day field
	// still matches only its own days.
	parsed, errParse := ParseCron("0 0 */5 * 3")
	if errParse != nil {
		t.Fatalf("ParseCron returned error: %v", errParse)
	}
	// Wednesdays and every fifth day of the month, whichever comes first.
	got := parsed.Next(cronUTC(2026, time.November, 10, 0, 0))
	if got.Weekday() != time.Wednesday && got.Day()%5 != 0 {
		t.Errorf("Next = %s, want a Wednesday or a day divisible by five", got)
	}
	if want := cronUTC(2026, time.November, 11, 0, 0); !got.Equal(want) {
		t.Errorf("Next = %s, want %s", got, want)
	}
}

func TestCronNilReceiver(t *testing.T) {
	var cron *Cron
	if got := cron.Next(cronUTC(2026, time.November, 10, 0, 0)); !got.IsZero() {
		t.Errorf("nil Cron.Next = %s, want the zero time", got)
	}
	if got := cron.String(); got != "" {
		t.Errorf("nil Cron.String() = %q, want an empty string", got)
	}
}
