// SPDX-License-Identifier: GPL-3.0-or-later

package secretstore

import "encoding/json"

func cloneAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch t := v.(type) {
		case map[string]any:
			out[k] = cloneAnyMap(t)
		case []any:
			out[k] = cloneAnySlice(t)
		default:
			out[k] = t
		}
	}
	return out
}

func cloneAnySlice(in []any) []any {
	if len(in) == 0 {
		return nil
	}
	out := make([]any, 0, len(in))
	for _, v := range in {
		switch t := v.(type) {
		case map[string]any:
			out = append(out, cloneAnyMap(t))
		case []any:
			out = append(out, cloneAnySlice(t))
		default:
			out = append(out, t)
		}
	}
	return out
}

func marshalCanonicalJSON(v any) (string, error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func canonicalizeConfig(raw Config, providerCfg any) (Config, error) {
	out := Config{}

	if providerCfg != nil {
		bs, err := json.Marshal(providerCfg)
		if err != nil {
			return nil, err
		}
		if len(bs) > 0 && string(bs) != "null" {
			var payload map[string]any
			if err := json.Unmarshal(bs, &payload); err != nil {
				return nil, err
			}
			for k, v := range payload {
				out[k] = v
			}
		}
	}

	out.SetName(raw.Name())
	out.SetKind(raw.Kind())
	out.SetSource(raw.Source())
	out.SetSourceType(raw.SourceType())
	return out, nil
}

func cloneConfig(in Config) Config {
	if len(in) == 0 {
		return nil
	}
	return Config(cloneAnyMap(map[string]any(in)))
}

func cloneStoreStatus(status StoreStatus) StoreStatus {
	out := status
	out.LastValidation = cloneValidationStatus(status.LastValidation)
	return out
}

func cloneValidationStatus(status *ValidationStatus) *ValidationStatus {
	if status == nil {
		return nil
	}
	copyStatus := *status
	return &copyStatus
}
