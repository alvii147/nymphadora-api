package jsonutils

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
)

// Timestamp represents time that is convert to unix timestamp in JSON.
type UnixTimestamp time.Time

// MarshalJSON converts time into unix timestamp bytes.
func (ut UnixTimestamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(ut).UTC().Unix(), 10)), nil
}

// UnmarshalJSON converts unix timestamp bytes into timestamp.
func (ut *UnixTimestamp) UnmarshalJSON(p []byte) error {
	s := string(p)
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errutils.FormatErrorf(err, "strconv.ParseInt failed for timestamp %s", s)
	}

	*(*time.Time)(ut) = time.Unix(ts, 0).UTC()

	return nil
}

// Optional represents a generic JSON field that can be null or unspecified.
// If specified, Valid is true and Value holds a pointer to the specified value.
// If null, Valid is true and Value is nil.
// If unspecified, Valid is false.
type Optional[T any] struct {
	Valid bool
	Value *T
}

// MarshalJSON converts Value into bytes.
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.Valid {
		return []byte(`null`), nil
	}

	b, err := json.Marshal(o.Value)
	if err != nil {
		return nil, errutils.FormatErrorf(err, "json.Marshal failed")
	}

	return b, nil
}

// UnmarshalJSON converts bytes into Optional struct.
func (o *Optional[T]) UnmarshalJSON(p []byte) error {
	o.Valid = true

	err := json.Unmarshal(p, &o.Value)
	if err != nil {
		return errutils.FormatErrorf(err, "json.Unmarshal failed")
	}

	return nil
}
