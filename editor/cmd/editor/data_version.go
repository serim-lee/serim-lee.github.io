package main

import (
	"encoding/json"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/dataversion"
)

// Legacy schemas are retained only for the old input/migration path. Public
// editor responses expose the data format version rather than that counter.
func (document SettingsDocument) MarshalJSON() ([]byte, error) {
	type plain SettingsDocument
	raw, err := json.Marshal(plain(document))
	if err != nil {
		return nil, err
	}
	if document.Version == "" {
		return raw, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	delete(fields, "schema_version")
	return json.Marshal(fields)
}

func settingsVersion(document SettingsDocument) string {
	if document.Version != "" {
		return document.Version
	}
	return dataversion.Current
}
