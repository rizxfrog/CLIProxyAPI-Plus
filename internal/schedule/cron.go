package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Cron is a parsed five-field cron expression: minute, hour, day of month, month,
// and day of week.
//
// Every field accepts "*", a single value, an inclusive "a-b" range, a step
// ("*/n" or "a-b/n", which needs a range or "*" before it), and comma-separated
// lists of those. Minute, hour, day of month, and month are numeric; month and day
// of week also accept the case-insensitive three-letter names JAN..DEC and
// SUN..SAT. Day of week accepts 0..7, where both 0 and 7 mean Sunday.
//
// Daily selection follows Vixie cron: when both day-of-month and day-of-week are
// restricted (neither is written as "*"), a date matches when either field
// matches; when only one of them is restricted, only that one constrains the
// date. A field written as "*" is unrestricted even after a step such as "*/2",
// which is a restriction.
//
// Next evaluates wall-clock time in the location of the reference instant, so it
// tracks daylight saving transitions: a wall-clock time skipped by a spring-forward
// transition never fires, and a wall-clock time repeated by a fall-back transition
// fires on its first occurrence only.
type Cron struct {
	minutes     uint64
	hours       uint64
	daysOfMonth uint64
	months      uint64
	daysOfWeek  uint64

	// domUnrestricted and dowUnrestricted record fields literally written as "*",
	// which is what selects the either-field day rule.
	domUnrestricted bool
	dowUnrestricted bool

	// times holds every matching wall clock of a day, ordered by hour then minute.
	times []cronTime
}

// cronTime is one matching wall clock during a day.
type cronTime struct {
	hour   int
	minute int
}

// cronSearchDays bounds how far ahead Next looks for a match. The window is about
// five years, which covers the leap-year cycle, so an expression that can match at
// all matches inside it; one that cannot match any date (February 30, for example)
// resolves to the zero time.
const cronSearchDays = 366 * 5

// cronFieldSpec describes one field of the expression: its display name for error
// messages, the inclusive value range it accepts, and the names it understands.
type cronFieldSpec struct {
	name  string
	min   int
	max   int
	names map[string]int
}

var (
	// cronMonthNames maps the three-letter month names to their 1-based numbers.
	cronMonthNames = map[string]int{
		"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
		"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
	}
	// cronWeekdayNames maps the three-letter day names to the weekday numbers used
	// by time.Weekday, where Sunday is 0.
	cronWeekdayNames = map[string]int{
		"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
	}
)

// ParseCron parses a five-field cron expression.
//
// It rejects a wrong field count, out-of-range values, reversed ranges, a step
// that is missing, non-numeric, zero, or not preceded by a range or "*", and
// unknown names, with an error that names the offending field.
func ParseCron(expr string) (*Cron, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf(
			"cron expression %q must have five fields (minute hour day-of-month month day-of-week), got %d",
			expr, len(fields))
	}
	specs := [5]cronFieldSpec{
		{name: "minute", min: 0, max: 59},
		{name: "hour", min: 0, max: 23},
		{name: "day-of-month", min: 1, max: 31},
		{name: "month", min: 1, max: 12, names: cronMonthNames},
		{name: "day-of-week", min: 0, max: 7, names: cronWeekdayNames},
	}
	var bits [5]uint64
	for index, spec := range specs {
		parsed, errField := parseCronField(fields[index], spec)
		if errField != nil {
			return nil, errField
		}
		bits[index] = parsed
	}

	cron := &Cron{
		minutes:         bits[0],
		hours:           bits[1],
		daysOfMonth:     bits[2],
		months:          bits[3],
		daysOfWeek:      bits[4],
		domUnrestricted: fields[2] == "*",
		dowUnrestricted: fields[4] == "*",
	}
	for hour := range 24 {
		if cron.hours&(1<<uint(hour)) == 0 {
			continue
		}
		for minute := range 60 {
			if cron.minutes&(1<<uint(minute)) == 0 {
				continue
			}
			cron.times = append(cron.times, cronTime{hour: hour, minute: minute})
		}
	}
	return cron, nil
}

// Next returns the first instant strictly after the reference instant whose
// wall-clock time matches the expression, evaluated in the reference's location.
// It returns the zero time when no match exists within cronSearchDays (about five
// years), which is the case for an expression such as "0 0 30 2 *".
//
// A matching wall clock skipped by a spring-forward transition is passed over, and
// one repeated by a fall-back transition is reported at its first occurrence.
func (c *Cron) Next(after time.Time) time.Time {
	if c == nil {
		return time.Time{}
	}
	location := after.Location()
	if location == nil {
		location = time.UTC
	}
	startYear, startMonth, startDay := after.Date()
	midnight := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, location)
	for offset := range cronSearchDays {
		day := midnight.AddDate(0, 0, offset)
		if !c.matchesDay(day) {
			continue
		}
		year, month, dayOfMonth := day.Date()
		for _, candidate := range c.times {
			instant, exists := resolveWallClock(year, month, dayOfMonth, candidate.hour, candidate.minute, location)
			if !exists {
				continue
			}
			if instant.After(after) {
				return instant
			}
		}
	}
	return time.Time{}
}

// String returns the normalized expression: fields joined by single spaces, each
// rendered as "*" when it covers every value it accepts and otherwise as
// comma-separated values with runs collapsed to "a-b" ranges.
func (c *Cron) String() string {
	if c == nil {
		return ""
	}
	fields := [5]string{
		formatCronField(c.minutes, 0, 59),
		formatCronField(c.hours, 0, 23),
		formatCronField(c.daysOfMonth, 1, 31),
		formatCronField(c.months, 1, 12),
		formatCronField(c.daysOfWeek, 0, 6),
	}
	return strings.Join(fields[:], " ")
}

// matchesDay reports whether the date component of a day matches the expression.
func (c *Cron) matchesDay(day time.Time) bool {
	if c.months&(1<<uint(int(day.Month()))) == 0 {
		return false
	}
	domMatches := c.daysOfMonth&(1<<uint(day.Day())) != 0
	dowMatches := c.daysOfWeek&(1<<uint(int(day.Weekday()))) != 0
	switch {
	case c.domUnrestricted && c.dowUnrestricted:
		return true
	case c.domUnrestricted:
		return dowMatches
	case c.dowUnrestricted:
		return domMatches
	default:
		return domMatches || dowMatches
	}
}

// parseCronField parses one field into a bit set of the values it selects.
func parseCronField(text string, spec cronFieldSpec) (uint64, error) {
	var bits uint64
	for _, item := range strings.Split(text, ",") {
		if item == "" {
			return 0, fmt.Errorf("cron field %s: list %q has an empty item", spec.name, text)
		}
		base := item
		step := 1
		stepText := ""
		if slash := strings.IndexByte(item, '/'); slash >= 0 {
			base = item[:slash]
			stepText = item[slash+1:]
			if stepText == "" {
				return 0, fmt.Errorf("cron field %s: %q is missing a step after %q", spec.name, item, "/")
			}
			parsed, errStep := strconv.Atoi(stepText)
			if errStep != nil {
				return 0, fmt.Errorf("cron field %s: step %q is not a number", spec.name, stepText)
			}
			if parsed <= 0 {
				return 0, fmt.Errorf("cron field %s: step %q must be greater than zero", spec.name, stepText)
			}
			step = parsed
		}
		if stepText != "" && base != "*" && !strings.Contains(base, "-") {
			return 0, fmt.Errorf("cron field %s: %q needs a range or \"*\" before the step, such as \"%d-%d/%s\"",
				spec.name, item, spec.min, spec.max, stepText)
		}
		low, high, errRange := cronFieldRange(base, item, spec)
		if errRange != nil {
			return 0, errRange
		}
		for value := low; value <= high; value += step {
			bits |= 1 << uint(cronFieldValue(value, spec))
		}
	}
	return bits, nil
}

// cronFieldRange resolves the "base" part of one list item, which is "*", a single
// value, or an inclusive "a-b" range, into its low and high bounds.
func cronFieldRange(base, item string, spec cronFieldSpec) (int, int, error) {
	if base == "*" {
		return spec.min, spec.max, nil
	}
	if dash := strings.IndexByte(base, '-'); dash > 0 {
		low, errLow := cronFieldValueFromText(base[:dash], spec)
		if errLow != nil {
			return 0, 0, errLow
		}
		high, errHigh := cronFieldValueFromText(base[dash+1:], spec)
		if errHigh != nil {
			return 0, 0, errHigh
		}
		if low > high {
			return 0, 0, fmt.Errorf("cron field %s: range %q is reversed", spec.name, item)
		}
		return low, high, nil
	}
	value, errValue := cronFieldValueFromText(base, spec)
	if errValue != nil {
		return 0, 0, errValue
	}
	return value, value, nil
}

// cronFieldValueFromText parses one value or name of a field and checks its range.
func cronFieldValueFromText(text string, spec cronFieldSpec) (int, error) {
	upper := strings.ToUpper(text)
	if spec.names != nil {
		if value, ok := spec.names[upper]; ok {
			return value, nil
		}
	}
	value, errAtoi := strconv.Atoi(text)
	if errAtoi != nil {
		if spec.names != nil {
			return 0, fmt.Errorf("cron field %s: %q is not a number or a known name", spec.name, text)
		}
		return 0, fmt.Errorf("cron field %s: %q is not a number", spec.name, text)
	}
	if value < spec.min || value > spec.max {
		return 0, fmt.Errorf("cron field %s: %q is out of range %d-%d", spec.name, text, spec.min, spec.max)
	}
	return value, nil
}

// cronFieldValue folds the accepted values onto the bit positions used at
// evaluation time, where day of week 7 is the same day as Sunday 0.
func cronFieldValue(value int, spec cronFieldSpec) int {
	if spec.names != nil && spec.name == "day-of-week" && value == 7 {
		return 0
	}
	return value
}

// formatCronField renders one field of the normalized expression.
func formatCronField(bits uint64, low, high int) string {
	covers := true
	for value := low; value <= high; value++ {
		if bits&(1<<uint(value)) == 0 {
			covers = false
			break
		}
	}
	if covers {
		return "*"
	}
	parts := make([]string, 0, 8)
	for value := low; value <= high; value++ {
		if bits&(1<<uint(value)) == 0 {
			continue
		}
		end := value
		for end+1 <= high && bits&(1<<uint(end+1)) != 0 {
			end++
		}
		if end == value {
			parts = append(parts, strconv.Itoa(value))
		} else {
			parts = append(parts, strconv.Itoa(value)+"-"+strconv.Itoa(end))
			value = end
		}
	}
	return strings.Join(parts, ",")
}

// resolveWallClock returns the first instant on the given date whose wall clock is
// hour:minute. It reports false when that wall clock does not exist, which happens
// when a spring-forward transition skips it.
//
// time.Date resolves a time that a fall-back transition repeats to its first
// occurrence; the backward probe keeps that guarantee even if the standard library
// ever resolves such a time the other way, so a repeated wall clock fires once.
func resolveWallClock(year int, month time.Month, day, hour, minute int, location *time.Location) (time.Time, bool) {
	instant := time.Date(year, month, day, hour, minute, 0, 0, location)
	if instant.Year() != year || instant.Month() != month || instant.Day() != day ||
		instant.Hour() != hour || instant.Minute() != minute {
		return time.Time{}, false
	}
	for _, shift := range []time.Duration{15 * time.Minute, 30 * time.Minute, time.Hour, 2 * time.Hour} {
		earlier := instant.Add(-shift)
		if earlier.Day() == day && earlier.Hour() == hour && earlier.Minute() == minute && earlier.Before(instant) {
			instant = earlier
		}
	}
	return instant, true
}
