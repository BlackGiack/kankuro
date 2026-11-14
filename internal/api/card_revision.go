package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
)

// cardRevision v1 frames every persisted field by name and type. Map keys are
// sorted; slices retain order and nil differs from empty. Unsupported types fail
// closed so a new storage field cannot silently escape concurrency detection.
func cardRevision(card config.IDCard) (string, error) {
	var b bytes.Buffer
	if err := encodeRevision(&b, reflect.ValueOf(card)); err != nil {
		return "", err
	}
	sum := sha256.Sum256(b.Bytes())
	return "v1:" + hex.EncodeToString(sum[:]), nil
}

func revisionFrame(b *bytes.Buffer, s string) {
	b.WriteString(strconv.Itoa(len(s)))
	b.WriteByte(':')
	b.WriteString(s)
}

//nolint:gocognit // One exhaustive reflection boundary rejects unsupported persisted types.
func encodeRevision(b *bytes.Buffer, v reflect.Value) error {
	revisionFrame(b, v.Type().String())
	if v.Type() == reflect.TypeFor[time.Time]() {
		timestamp, ok := reflect.TypeAssert[time.Time](v)
		if !ok {
			return errors.New("invalid timestamp revision value")
		}
		revisionFrame(b, timestamp.UTC().Format(time.RFC3339Nano))
		return nil
	}
	//nolint:exhaustive // The default deliberately rejects unsupported kinds.
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			field := v.Type().Field(i)
			if field.Tag.Get("yaml") == "-" {
				continue
			}
			if !field.IsExported() {
				return fmt.Errorf("unsupported revision field %s", field.Name)
			}
			revisionFrame(b, field.Name)
			if err := encodeRevision(b, v.Field(i)); err != nil {
				return err
			}
		}
	case reflect.String:
		revisionFrame(b, v.String())
	case reflect.Bool:
		if v.Bool() {
			b.WriteByte(1)
		} else {
			b.WriteByte(0)
		}
	case reflect.Slice:
		if v.IsNil() {
			b.WriteByte(0)
			return nil
		}
		b.WriteByte(1)
		revisionFrame(b, strconv.Itoa(v.Len()))
		for i := range v.Len() {
			if err := encodeRevision(b, v.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("unsupported revision map %s", v.Type())
		}
		if v.IsNil() {
			b.WriteByte(0)
			return nil
		}
		b.WriteByte(1)
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		revisionFrame(b, strconv.Itoa(len(keys)))
		for _, key := range keys {
			if err := encodeRevision(b, key); err != nil {
				return err
			}
			if err := encodeRevision(b, v.MapIndex(key)); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported revision type %s", v.Type())
	}
	return nil
}
