package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Legacy settings used separate catalogs. Merge by stable ID, retaining all
// references, project order and any explicitly configured color or extra fields.
func captureUnifiedTopics(document *SettingsDocument, root map[string]json.RawMessage) error {
	if raw, exists := root["topics"]; exists {
		if document.Topics == nil {
			return errors.New("topics는 JSON 배열이어야 합니다")
		}
		if err := captureTaxonomyExtraFields(document.Topics, raw); err != nil {
			return err
		}
	} else {
		var lists [][]TaxonomyItem
		for _, key := range []string{"project_themes", "publication_topics"} {
			var items []TaxonomyItem
			if raw, exists := root[key]; exists {
				if err := json.Unmarshal(raw, &items); err != nil || items == nil {
					return fmt.Errorf("%s는 JSON 배열이어야 합니다", key)
				}
				if err := captureTaxonomyExtraFields(items, raw); err != nil {
					return err
				}
				if err := validateLoadedTaxonomy(items, key, reservedPublicationTopicIDs); err != nil {
					return err
				}
			}
			lists = append(lists, items)
		}
		if lists[0] == nil || lists[1] == nil {
			return errors.New("settings.json에 topics 배열이 필요합니다")
		}
		document.Topics = mergeLegacyTopics(lists...)
	}
	fallbackLoaded := false
	for _, key := range []string{"topic_fallback", "project_theme_fallback", "publication_topic_fallback"} {
		if raw, exists := root[key]; exists {
			var fallback BilingualLabel
			if err := json.Unmarshal(raw, &fallback); err != nil {
				return err
			}
			fields, err := decodeObject(raw)
			if err != nil {
				return err
			}
			fallback.extraFields = withoutKnownFields(fields, "label_en", "label_ko")
			if !fallbackLoaded {
				document.TopicFallback = fallback
				fallbackLoaded = true
			} else {
				for name, value := range fallback.extraFields {
					if _, exists := document.TopicFallback.extraFields[name]; !exists {
						document.TopicFallback.extraFields[name] = cloneRawMessage(value)
					}
				}
			}
			if key == "topic_fallback" {
				break
			}
		}
	}
	return nil
}

func mergeLegacyTopics(lists ...[]TaxonomyItem) []TaxonomyItem {
	result := []TaxonomyItem{}
	positions := make(map[string]int)
	for _, list := range lists {
		for _, item := range list {
			if position, exists := positions[item.ID]; exists {
				current := &result[position]
				if current.Color == "" {
					current.Color = item.Color
					current.colorProvided = item.colorProvided
				}
				extras := cloneExtraFields(item.extraFields)
				if extras == nil {
					extras = make(map[string]json.RawMessage)
				}
				for key, value := range current.extraFields {
					extras[key] = cloneRawMessage(value)
				}
				current.extraFields = extras
				continue
			}
			positions[item.ID] = len(result)
			result = append(result, item)
		}
	}
	return result
}

func combinedTopicUsage(usage SettingsUsage) map[string]int {
	result := make(map[string]int)
	for _, counts := range []map[string]int{usage.ProjectThemes, usage.PublicationTopics} {
		for id, count := range counts {
			if id != "" {
				result[id] += count
			}
		}
	}
	return result
}
