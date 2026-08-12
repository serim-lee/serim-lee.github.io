package main

import (
	"encoding/json"
	"fmt"
)

func visibleInCV(raw json.RawMessage, contextLabel string) (bool, error) {
	fields, err := decodeObject(raw)
	if err != nil {
		return false, err
	}
	value, exists := fields["visible_in_CV"]
	if !exists {
		return true, nil
	}
	var visible *bool
	if err := json.Unmarshal(value, &visible); err != nil || visible == nil {
		return false, fmt.Errorf("%s의 visible_in_CV는 Boolean이어야 합니다", contextLabel)
	}
	return *visible, nil
}

func mustVisibleInCV(raw json.RawMessage) bool {
	visible, err := visibleInCV(raw, "항목")
	return err == nil && visible
}

func withVisibleInCV(raw json.RawMessage, requested *bool) (json.RawMessage, error) {
	if requested == nil {
		return cloneRawMessage(raw), nil
	}
	fields, err := decodeObject(raw)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(*requested)
	if err != nil {
		return nil, err
	}
	fields["visible_in_CV"] = encoded
	return json.Marshal(fields)
}

func visibleInCVForSave(requested *bool, isNew bool) *bool {
	if requested != nil || !isNew {
		return requested
	}
	value := true
	return &value
}
