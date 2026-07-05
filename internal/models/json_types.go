// json_types.go — shared custom column types.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringSlice maps a []string to a JSON column (pasos, alérgenos, fotos).
// Same Valuer/Scanner pattern as Roles in user.go.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(value any) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("stringslice: unsupported column type %T", value)
	}
	if len(bytes) == 0 {
		*s = StringSlice{}
		return nil
	}
	return json.Unmarshal(bytes, s)
}
