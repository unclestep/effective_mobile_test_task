package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDurationUnmarshalYAMLValid(t *testing.T) {
	var d Duration
	err := yaml.Unmarshal([]byte(`30s`), &d)

	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, d.Duration())
}

func TestDurationUnmarshalYAMLComplex(t *testing.T) {
	var d Duration
	err := yaml.Unmarshal([]byte(`1h30m`), &d)

	require.NoError(t, err)
	assert.Equal(t, time.Hour+30*time.Minute, d.Duration())
}

func TestDurationUnmarshalYAMLInvalid(t *testing.T) {
	var d Duration
	err := yaml.Unmarshal([]byte(`invalid`), &d)

	require.Error(t, err)
}

func TestDurationUnmarshalTextValid(t *testing.T) {
	var d Duration
	err := d.UnmarshalText([]byte("5m"))

	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, d.Duration())
}

func TestDurationUnmarshalTextInvalid(t *testing.T) {
	var d Duration
	err := d.UnmarshalText([]byte("bad"))

	require.Error(t, err)
}

func TestDurationDuration(t *testing.T) {
	d := Duration(45 * time.Second)

	assert.Equal(t, 45*time.Second, d.Duration())
}
