package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonthYearUnmarshalJSONValid(t *testing.T) {
	var m MonthYear
	err := m.UnmarshalJSON([]byte(`"07-2025"`))

	require.NoError(t, err)

	tm := m.Time()
	assert.Equal(t, 2025, tm.Year())
	assert.Equal(t, time.July, tm.Month())
	assert.Equal(t, 1, tm.Day())
}

func TestMonthYearUnmarshalJSONWrongFormat(t *testing.T) {
	var m MonthYear
	err := m.UnmarshalJSON([]byte(`"2025-07"`))

	require.Error(t, err)
}

func TestMonthYearUnmarshalJSONInvalidString(t *testing.T) {
	var m MonthYear
	err := m.UnmarshalJSON([]byte(`"string"`))

	require.Error(t, err)
}

func TestMonthYearUnmarshalJSONNotAString(t *testing.T) {
	var m MonthYear
	err := m.UnmarshalJSON([]byte(`123`))

	require.Error(t, err)
}

func TestMonthYearMarshalJSON(t *testing.T) {
	m := MonthYear(time.Date(2025, time.July, 15, 10, 30, 0, 0, time.UTC))

	data, err := m.MarshalJSON()

	require.NoError(t, err)
	assert.Equal(t, `"07-2025"`, string(data))
}

func TestMonthYearRoundTrip(t *testing.T) {
	original := MonthYear(time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC))

	data, err := original.MarshalJSON()
	require.NoError(t, err)

	var restored MonthYear
	require.NoError(t, restored.UnmarshalJSON(data))

	assert.True(t, time.Time(original).Equal(time.Time(restored)))
}
