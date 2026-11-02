package dashboardmigration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// TransformCards prepares a legacy card document for the default dashboard.
// Unknown YAML fields are retained. It does not write or publish the result.
//
//nolint:gocognit // Validate the whole document before returning any transformed output.
func TransformCards(raw []byte) ([]byte, error) {
	var decoded map[string]any
	if err := yaml.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	var document yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("multiple YAML documents")
		}
		return nil, err
	}
	if hasAliases(&document) {
		return nil, errors.New("card document contains YAML aliases or merges")
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("card document must be a mapping")
	}
	root := document.Content[0]
	version := mappingValue(root, "schema_version")
	if version == nil {
		version = scalar("1", "!!int")
		root.Content = append(root.Content, scalar("schema_version", "!!str"), version)
	} else {
		var number int
		if version.Kind != yaml.ScalarNode || version.Tag != "!!int" || version.Decode(&number) != nil || (number != 0 && number != 1) {
			return nil, errors.New("unsupported card schema version")
		}
	}
	cards := mappingValue(root, "id_cards")
	switch {
	case cards == nil:
		cards = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		root.Content = append(root.Content, scalar("id_cards", "!!str"), cards)
	case cards.Kind == yaml.ScalarNode && cards.Tag == "!!null":
		cards.Kind, cards.Tag, cards.Value = yaml.SequenceNode, "!!seq", ""
	case cards.Kind != yaml.SequenceNode:
		return nil, errors.New("id_cards must be a sequence")
	}
	seen := make(map[string]bool)
	for _, card := range cards.Content {
		if card.Kind != yaml.MappingNode {
			return nil, errors.New("card must be a mapping")
		}
		id := mappingValue(card, "id")
		if id == nil || id.Tag != "!!str" || strings.TrimSpace(id.Value) == "" || seen[id.Value] {
			return nil, errors.New("missing or duplicate card ID")
		}
		if mappingValue(card, "dashboard_id") != nil || mappingValue(card, "shared") != nil {
			return nil, fmt.Errorf("card %s already has membership metadata", id.Value)
		}
		seen[id.Value] = true
		card.Content = append(
			card.Content,
			scalar("dashboard_id", "!!str"),
			scalar("default", "!!str"),
			scalar("shared", "!!str"),
			scalar("false", "!!bool"),
		)
	}
	version.Value = "2"
	return yaml.Marshal(&document)
}

// TransformSession assigns a legacy session to the default dashboard while
// retaining all other JSON fields. It does not write or publish the result.
func TransformSession(raw []byte) ([]byte, error) {
	if err := validateJSON(raw); err != nil {
		return nil, err
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errors.New("session must be an object")
	}
	var id string
	if err := json.Unmarshal(record["id"], &id); err != nil || id == "" {
		return nil, errors.New("session ID is required")
	}
	if _, exists := record["dashboard_id"]; exists {
		return nil, errors.New("session already has dashboard membership")
	}
	if _, exists := record["schema_version"]; exists {
		return nil, errors.New("session already has a schema version")
	}
	var managed bool
	if rawManaged, exists := record["instructions_managed"]; exists {
		if err := json.Unmarshal(rawManaged, &managed); err != nil {
			return nil, fmt.Errorf("invalid instruction ownership flag: %w", err)
		}
	}
	var ownership map[string]bool
	if rawOwnership, exists := record["instruction_ownership"]; exists {
		if err := json.Unmarshal(rawOwnership, &ownership); err != nil {
			return nil, fmt.Errorf("invalid instruction ownership map: %w", err)
		}
	}
	if managed && len(ownership) == 0 {
		record["instruction_ownership"] = json.RawMessage(`{"AGENTS.md":true}`)
	}
	record["dashboard_id"] = json.RawMessage(`"default"`)
	record["schema_version"] = json.RawMessage(`2`)
	return json.MarshalIndent(record, "", "  ")
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func scalar(value, tag string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value}
}

func hasAliases(node *yaml.Node) bool {
	if node.Kind == yaml.AliasNode || node.Tag == "!!merge" {
		return true
	}
	return slices.ContainsFunc(node.Content, hasAliases)
}
