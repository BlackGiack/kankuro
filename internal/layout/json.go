package layout

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// validateJSON rejects duplicate object keys, including in extension fields;
// ordinary [json.Unmarshal] would silently discard earlier values.
func validateJSON(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := jsonValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

//nolint:gocognit // The recursive object and array grammar is kept together.
func jsonValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]bool)
		for decoder.More() {
			keyToken, errAt40 := decoder.Token()
			if errAt40 != nil {
				return errAt40
			}
			key, okAt44 := keyToken.(string)
			if !okAt44 || seen[key] {
				return fmt.Errorf("invalid or duplicate JSON object key %q", key)
			}
			seen[key] = true
			if errAt49 := jsonValue(decoder); errAt49 != nil {
				return errAt49
			}
		}
		closing, errAt53 := decoder.Token()
		if errAt53 != nil || closing != json.Delim('}') {
			return fmt.Errorf("invalid JSON object: %w", errAt53)
		}
	case '[':
		for decoder.More() {
			if errAt59 := jsonValue(decoder); errAt59 != nil {
				return errAt59
			}
		}
		closing, errAt63 := decoder.Token()
		if errAt63 != nil || closing != json.Delim(']') {
			return fmt.Errorf("invalid JSON array: %w", errAt63)
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	return nil
}
