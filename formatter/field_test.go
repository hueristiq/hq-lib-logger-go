package formatter

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldConstructors(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 28, 12, 0, 0, 123, time.FixedZone("x", 3600))

	tests := []struct {
		name  string
		field Field
		value any
		text  string
		json  string
	}{
		{"string", String("k", "v"), "v", "v", `"v"`},
		{"int", Int("k", -2), int64(-2), "-2", "-2"},
		{"uint", Uint("k", 3), uint64(3), "3", "3"},
		{"float", Float64("k", 1.5), 1.5, "1.5", "1.5"},
		{"bool", Bool("k", true), true, "true", "true"},
		{"duration", Duration("k", time.Second), time.Second, "1s", `"1s"`},
		{"time", Time("k", now), now.UTC(), now.UTC().Format(time.RFC3339Nano), `"` + now.UTC().Format(time.RFC3339Nano) + `"`},
		{"error", Err("k", errors.New("boom")), errors.New("boom"), "boom", `"boom"`},
		{"any", Any("k", []int{1}), []int{1}, "[1]", "[1]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "k", tt.field.Key())

			if tt.name == "error" {
				require.EqualError(t, tt.field.Value().(error), "boom")
			} else if tt.name != "time" {
				assert.Equal(t, tt.value, tt.field.Value())
			}

			assert.Equal(t, tt.text, string(tt.field.AppendText(nil)))

			data, err := tt.field.AppendJSON(nil)
			require.NoError(t, err)
			assert.JSONEq(t, tt.json, string(data))
		})
	}
}

func TestFieldNilAndInvalidJSON(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "<nil>", string(Err("k", nil).AppendText(nil)))

	data, err := Float64("k", math.Inf(1)).AppendJSON(nil)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

func TestFieldValueIsTyped(t *testing.T) {
	t.Parallel()

	value := String("k", "v").Value()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	assert.Equal(t, `"v"`, string(raw))
}
