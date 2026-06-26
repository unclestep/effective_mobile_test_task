package http

import (
	"encoding/json"
	"fmt"
	"time"
)

const monthYearLayout = "01-2006"

type MonthYear time.Time

func (m MonthYear) Time() time.Time {
	return time.Time(m)
}

func (m MonthYear) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(m).Format(monthYearLayout))
}

func (m *MonthYear) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(monthYearLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q, expected MM-YYYY: %w", s, err)
	}
	*m = MonthYear(t)
	return nil
}
