package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	tTime := time.Time(t)
	if tTime.IsZero() {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "\"%s\"", tTime.Format("2006-01-02 15:04:05")), nil
}

func (t Time) ToDateString() string {
	if time.Time(t).IsZero() {
		return ""
	}
	return time.Time(t).Format("2006-01-02")
}

func (t Time) ToString() string {
	if time.Time(t).IsZero() {
		return ""
	}
	return time.Time(t).Format("2006-01-02 15:04:05")
}

func (t Time) Value() (driver.Value, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return time.Time(t), nil
}

func (t *Time) Scan(v any) error {
	if value, ok := v.(time.Time); ok {
		*t = Time(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
