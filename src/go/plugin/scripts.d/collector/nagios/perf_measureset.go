// SPDX-License-Identifier: GPL-3.0-or-later

package nagios

import (
	"sort"
	"strings"

	"github.com/netdata/netdata/go/plugins/pkg/metrix"
)

const (
	perfFieldValue         = "value"
	perfFieldMin           = "min"
	perfFieldMax           = "max"
	perfFieldWarnDefined   = "warn_defined"
	perfFieldWarnInclusive = "warn_inclusive"
	perfFieldWarnLow       = "warn_low"
	perfFieldWarnHigh      = "warn_high"
	perfFieldWarnLowDef    = "warn_low_defined"
	perfFieldWarnHighDef   = "warn_high_defined"
	perfFieldCritDefined   = "crit_defined"
	perfFieldCritInclusive = "crit_inclusive"
	perfFieldCritLow       = "crit_low"
	perfFieldCritHigh      = "crit_high"
	perfFieldCritLowDef    = "crit_low_defined"
	perfFieldCritHighDef   = "crit_high_defined"
)

var perfMeasureSetFieldOrder = []string{
	perfFieldValue,
	perfFieldMin,
	perfFieldMax,
	perfFieldWarnDefined,
	perfFieldWarnInclusive,
	perfFieldWarnLow,
	perfFieldWarnHigh,
	perfFieldWarnLowDef,
	perfFieldWarnHighDef,
	perfFieldCritDefined,
	perfFieldCritInclusive,
	perfFieldCritLow,
	perfFieldCritHigh,
	perfFieldCritLowDef,
	perfFieldCritHighDef,
}

type perfMeasureSet struct {
	name       string
	scriptName string
	unit       string
	values     map[string]metrix.SampleValue
}

func perfMeasureSetFieldSpecs() []metrix.MeasureFieldSpec {
	out := make([]metrix.MeasureFieldSpec, 0, len(perfMeasureSetFieldOrder))
	for _, field := range perfMeasureSetFieldOrder {
		out = append(out, metrix.MeasureFieldSpec{
			Name:  field,
			Float: perfMeasureFieldFloat(field),
		})
	}
	return out
}

func perfMeasureFieldFloat(field string) bool {
	switch field {
	case perfFieldValue, perfFieldMin, perfFieldMax, perfFieldWarnLow, perfFieldWarnHigh, perfFieldCritLow, perfFieldCritHigh:
		return true
	default:
		return false
	}
}

func buildPerfMeasureSets(samples []perfMetricSample) []perfMeasureSet {
	if len(samples) == 0 {
		return nil
	}

	groups := make(map[string]*perfMeasureSet)
	for _, sample := range samples {
		base, field, ok := splitPerfSampleField(sample.name)
		if !ok {
			continue
		}

		group := groups[base]
		if group == nil {
			scriptName, _ := splitPerfMeasureSetSource(base)
			group = &perfMeasureSet{
				name:       base,
				scriptName: scriptName,
				values:     defaultPerfMeasureSetValues(),
			}
			groups[base] = group
		}
		group.values[field] = sample.value
		if field == perfFieldValue {
			group.unit = sample.unit
		} else if group.unit == "" && sample.unit != "state" {
			group.unit = sample.unit
		}
	}

	out := make([]perfMeasureSet, 0, len(groups))
	for _, group := range groups {
		if group.unit == "" {
			group.unit = "generic"
		}
		out = append(out, *group)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out
}

func defaultPerfMeasureSetValues() map[string]metrix.SampleValue {
	values := make(map[string]metrix.SampleValue, len(perfMeasureSetFieldOrder))
	for _, field := range perfMeasureSetFieldOrder {
		values[field] = 0
	}
	return values
}

func splitPerfSampleField(name string) (base string, field string, ok bool) {
	for _, candidate := range perfMeasureSetFieldOrder {
		suffix := "_" + candidate
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSuffix(name, suffix), candidate, true
		}
	}
	return "", "", false
}

func splitPerfMeasureSetSource(base string) (scriptName, family string) {
	if idx := strings.Index(base, "."); idx > 0 && idx < len(base)-1 {
		return base[:idx], base[idx+1:]
	}
	return base, ""
}
