package trae

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// sensitiveKeys are omitted from schema reports. A real catalogue row carries
// per-account credential material and encrypted prompt blobs; the report exists
// to confirm field names and types, never to expose values.
var sensitiveKeys = map[string]struct{}{
	"ak":                    {},
	"sk":                    {},
	"token":                 {},
	"access_token":          {},
	"refresh_token":         {},
	"session_token":         {},
	"encrypted_model_param": {},
	"encrypted_prompt":      {},
	"encrypted_prompt_set":  {},
	"encrypted_prompt_list": {},
	"custom_config":         {},
	"extra_config":          {},
	"model_extra_config":    {},
	"provider_context":      {},
	"biz_context":           {},
	"api_key":               {},
	"cookie":                {},
	"machine_id":            {},
	"device_id":             {},
	"uid":                   {},
	"email":                 {},
}

// SchemaReport describes the shape of a model_list response without exposing
// values. It is emitted once per credential so an operator can confirm the
// inferred field names against the real upstream payload.
//
// Output shape:
//
//	envelope keys: allow_tenant_user_add_model, config_info_list, code, metadata
//	config_info_list: 54 entries
//	row keys: config_name(string), config_switch(bool), context_window_tokens(object{dev}), …
//	parsed: 14 routable, skipped: 1
type SchemaReport struct {
	EnvelopeKeys []string
	RowCount     int
	RowKeys      []string
	Parsed       int
	Skipped      int
	Anomalies    []string
}

// String renders the report as one log-friendly line set.
func (r SchemaReport) String() string {
	var builder strings.Builder
	builder.WriteString("envelope keys: ")
	builder.WriteString(joinOrNone(r.EnvelopeKeys))
	fmt.Fprintf(&builder, "; config_info_list: %d entries", r.RowCount)
	builder.WriteString("; row keys: ")
	builder.WriteString(joinOrNone(r.RowKeys))
	fmt.Fprintf(&builder, "; parsed: %d, skipped: %d", r.Parsed, r.Skipped)
	if len(r.Anomalies) > 0 {
		builder.WriteString("; anomalies: ")
		builder.WriteString(strings.Join(r.Anomalies, "; "))
	}
	return builder.String()
}

// DescribeModelList builds a value-free schema report from a raw model_list
// body. It never fails: a body it cannot decode yields the anomalies list so the
// log still tells the operator what arrived.
func DescribeModelList(raw []byte) SchemaReport {
	var report SchemaReport
	var envelope map[string]any
	if errUnmarshal := json.Unmarshal(raw, &envelope); errUnmarshal != nil {
		report.Anomalies = append(report.Anomalies, "body is not a JSON object: "+errUnmarshal.Error())
		return report
	}
	report.EnvelopeKeys = sortedKeys(envelope)

	rows, ok := envelope["config_info_list"].([]any)
	if !ok {
		if _, present := envelope["config_info_list"]; present {
			report.Anomalies = append(report.Anomalies, "config_info_list is not an array")
		} else {
			report.Anomalies = append(report.Anomalies, "config_info_list missing from envelope")
		}
		return report
	}
	report.RowCount = len(rows)
	if len(rows) == 0 {
		report.Anomalies = append(report.Anomalies, "config_info_list is empty")
		return report
	}

	// Merge every row's keys so a field that only appears on some rows is still
	// reported, and note the type actually observed for each.
	types := map[string]map[string]struct{}{}
	for index, entry := range rows {
		row, ok := entry.(map[string]any)
		if !ok {
			report.Anomalies = append(report.Anomalies, fmt.Sprintf("row %d is not an object", index))
			continue
		}
		for key, value := range row {
			if _, sensitive := sensitiveKeys[key]; sensitive {
				if types[key] == nil {
					types[key] = map[string]struct{}{}
				}
				types[key]["<redacted>"] = struct{}{}
				continue
			}
			if types[key] == nil {
				types[key] = map[string]struct{}{}
			}
			types[key][describeType(value)] = struct{}{}
		}
	}
	keys := make([]string, 0, len(types))
	for key := range types {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		variants := make([]string, 0, len(types[key]))
		for variant := range types[key] {
			variants = append(variants, variant)
		}
		sort.Strings(variants)
		report.RowKeys = append(report.RowKeys, fmt.Sprintf("%s(%s)", key, strings.Join(variants, "|")))
	}

	for _, entry := range rows {
		row, ok := entry.(map[string]any)
		if !ok {
			report.Skipped++
			continue
		}
		if _, routable := parseConfigInfo(mustMarshal(row)); routable {
			report.Parsed++
		} else {
			report.Skipped++
		}
	}

	// Flag the fields the parser depends on when they are absent, so a schema
	// drift shows up as an explicit anomaly instead of a silent metadata loss.
	if _, present := types["config_name"]; !present {
		report.Anomalies = append(report.Anomalies, "no row carries config_name (the chat routing key)")
	}
	if _, present := types["context_window_tokens"]; !present {
		if _, sizePresent := types["context_window_size"]; !sizePresent {
			report.Anomalies = append(report.Anomalies, "no context window field found")
		}
	}
	return report
}

func describeType(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case float64:
		return "number"
	case string:
		return "string"
	case []any:
		return "array"
	case map[string]any:
		inner := make([]string, 0, len(typed))
		for key := range typed {
			inner = append(inner, key)
		}
		sort.Strings(inner)
		return "object{" + strings.Join(inner, ",") + "}"
	default:
		return fmt.Sprintf("%T", value)
	}
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func joinOrNone(values []string) string {
	if len(values) == 0 {
		return "<none>"
	}
	return strings.Join(values, ", ")
}

func mustMarshal(value any) []byte {
	raw, errMarshal := json.Marshal(value)
	if errMarshal != nil {
		return nil
	}
	return raw
}
