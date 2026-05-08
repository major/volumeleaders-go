package volumeleaders

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAspNetDateUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantValid bool
		wantTime  time.Time
		wantErr   bool
	}{
		{
			name:      "valid",
			data:      `"/Date(1767225600000)/"`,
			wantValid: true,
			wantTime:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{name: "null", data: `null`},
		{name: "empty string", data: `""`},
		{name: "dotnet minimum", data: `"/Date(-62135596800000)/"`},
		{name: "1900 sentinel", data: `"/Date(-2208988800000)/"`},
		{name: "invalid format", data: `"2026-01-01"`, wantErr: true},
		{name: "invalid type", data: `42`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AspNetDate{Time: time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC), Valid: true}

			err := json.Unmarshal([]byte(tt.data), &got)

			if tt.wantErr {
				require.Error(t, err, "Unmarshal(%s)", tt.data)
				return
			}
			require.NoError(t, err, "Unmarshal(%s)", tt.data)
			assert.Equal(t, tt.wantValid, got.Valid, "Unmarshal(%s).Valid", tt.data)
			if tt.wantValid {
				assert.True(
					t,
					got.Time.Equal(tt.wantTime),
					"Unmarshal(%s).Time = %s, want %s",
					tt.data,
					got.Time,
					tt.wantTime,
				)
			}
		})
	}
}

func TestAspNetDateMarshalJSON(t *testing.T) {
	valid := AspNetDate{Time: time.Date(2026, time.May, 1, 14, 30, 0, 0, time.FixedZone("EDT", -4*60*60)), Valid: true}

	data, err := json.Marshal(valid)
	require.NoError(t, err, "Marshal(valid AspNetDate)")
	assert.JSONEq(t, `"2026-05-01T18:30:00Z"`, string(data), "Marshal(valid AspNetDate)")

	data, err = json.Marshal(AspNetDate{})
	require.NoError(t, err, "Marshal(invalid AspNetDate)")
	assert.Equal(t, "null", string(data), "Marshal(invalid AspNetDate)")
}

func TestFlexBoolUnmarshalJSON(t *testing.T) {
	tests := []struct {
		data    string
		want    bool
		wantErr bool
	}{
		{data: `true`, want: true},
		{data: `1`, want: true},
		{data: `false`},
		{data: `0`},
		{data: `null`},
		{data: `"true"`, wantErr: true},
		{data: `2`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.data, func(t *testing.T) {
			var got FlexBool

			err := json.Unmarshal([]byte(tt.data), &got)

			if tt.wantErr {
				require.Error(t, err, "Unmarshal(%s)", tt.data)
				return
			}
			require.NoError(t, err, "Unmarshal(%s)", tt.data)
			assert.Equal(t, tt.want, bool(got), "Unmarshal(%s)", tt.data)
		})
	}
}

func TestFlexBoolMarshalJSON(t *testing.T) {
	data, err := json.Marshal(FlexBool(true))
	require.NoError(t, err, "Marshal(FlexBool(true))")
	assert.Equal(t, "true", string(data), "Marshal(FlexBool(true))")

	data, err = json.Marshal(FlexBool(false))
	require.NoError(t, err, "Marshal(FlexBool(false))")
	assert.Equal(t, "false", string(data), "Marshal(FlexBool(false))")
}

func TestTradeSemanticAccessors(t *testing.T) {
	trade := Trade{
		FrequencyLast30TD: 3,
		FrequencyLast90TD: 7,
		FrequencyLast1CY:  12,
		PhantomPrint:      true,
		OPEX:              true,
		VOLEX:             true,
		Cancelled:         true,
	}

	assert.Equal(t, 3, trade.SimilarTradeCountLast30Days())
	assert.Equal(t, 7, trade.SimilarTradeCountLast90Days())
	assert.Equal(t, 12, trade.SimilarTradeCountLastYear())
	assert.True(t, trade.IsPhantomPrint())
	assert.True(t, trade.IsOptionsExpiration())
	assert.True(t, trade.IsVolatilityExpiration())
	assert.True(t, trade.IsCancelled())
}

// TestTradeSemanticAccessors_ZeroValue verifies that all accessors return
// their zero/false equivalents when a Trade is constructed with no fields set.
// This is a regression guard: zero-value structs must not accidentally return
// truthy or non-zero results.
func TestTradeSemanticAccessors_ZeroValue(t *testing.T) {
	var trade Trade

	assert.Equal(t, 0, trade.SimilarTradeCountLast30Days())
	assert.Equal(t, 0, trade.SimilarTradeCountLast90Days())
	assert.Equal(t, 0, trade.SimilarTradeCountLastYear())
	assert.False(t, trade.IsPhantomPrint())
	assert.False(t, trade.IsOptionsExpiration())
	assert.False(t, trade.IsVolatilityExpiration())
	assert.False(t, trade.IsCancelled())
}

// TestTradeSemanticAccessors_FrequencyIsolation verifies that each frequency
// accessor reads only its own backing field and does not alias another.
// Setting only one field at a time ensures there is no accidental
// cross-wiring between FrequencyLast30TD, FrequencyLast90TD, and
// FrequencyLast1CY.
func TestTradeSemanticAccessors_FrequencyIsolation(t *testing.T) {
	t.Run("only FrequencyLast30TD set", func(t *testing.T) {
		trade := Trade{FrequencyLast30TD: 5}
		assert.Equal(t, 5, trade.SimilarTradeCountLast30Days())
		assert.Equal(t, 0, trade.SimilarTradeCountLast90Days())
		assert.Equal(t, 0, trade.SimilarTradeCountLastYear())
	})

	t.Run("only FrequencyLast90TD set", func(t *testing.T) {
		trade := Trade{FrequencyLast90TD: 10}
		assert.Equal(t, 0, trade.SimilarTradeCountLast30Days())
		assert.Equal(t, 10, trade.SimilarTradeCountLast90Days())
		assert.Equal(t, 0, trade.SimilarTradeCountLastYear())
	})

	t.Run("only FrequencyLast1CY set", func(t *testing.T) {
		trade := Trade{FrequencyLast1CY: 20}
		assert.Equal(t, 0, trade.SimilarTradeCountLast30Days())
		assert.Equal(t, 0, trade.SimilarTradeCountLast90Days())
		assert.Equal(t, 20, trade.SimilarTradeCountLastYear())
	})
}

// TestTradeSemanticAccessors_FrequencyBoundary checks frequency accessors at
// boundary values: 1 (minimum meaningful count) and a large value.
func TestTradeSemanticAccessors_FrequencyBoundary(t *testing.T) {
	t.Run("frequency of 1", func(t *testing.T) {
		trade := Trade{FrequencyLast30TD: 1, FrequencyLast90TD: 1, FrequencyLast1CY: 1}
		assert.Equal(t, 1, trade.SimilarTradeCountLast30Days())
		assert.Equal(t, 1, trade.SimilarTradeCountLast90Days())
		assert.Equal(t, 1, trade.SimilarTradeCountLastYear())
	})

	t.Run("large frequency", func(t *testing.T) {
		trade := Trade{FrequencyLast30TD: 9999, FrequencyLast90TD: 9999, FrequencyLast1CY: 9999}
		assert.Equal(t, 9999, trade.SimilarTradeCountLast30Days())
		assert.Equal(t, 9999, trade.SimilarTradeCountLast90Days())
		assert.Equal(t, 9999, trade.SimilarTradeCountLastYear())
	})
}

// TestTradeSemanticAccessors_BoolIsolation verifies that each boolean accessor
// reads only its own backing field. Each sub-test sets exactly one flag and
// confirms only that accessor returns true while the others remain false.
func TestTradeSemanticAccessors_BoolIsolation(t *testing.T) {
	t.Run("only PhantomPrint set", func(t *testing.T) {
		trade := Trade{PhantomPrint: true}
		assert.True(t, trade.IsPhantomPrint())
		assert.False(t, trade.IsOptionsExpiration())
		assert.False(t, trade.IsVolatilityExpiration())
		assert.False(t, trade.IsCancelled())
	})

	t.Run("only OPEX set", func(t *testing.T) {
		trade := Trade{OPEX: true}
		assert.False(t, trade.IsPhantomPrint())
		assert.True(t, trade.IsOptionsExpiration())
		assert.False(t, trade.IsVolatilityExpiration())
		assert.False(t, trade.IsCancelled())
	})

	t.Run("only VOLEX set", func(t *testing.T) {
		trade := Trade{VOLEX: true}
		assert.False(t, trade.IsPhantomPrint())
		assert.False(t, trade.IsOptionsExpiration())
		assert.True(t, trade.IsVolatilityExpiration())
		assert.False(t, trade.IsCancelled())
	})

	t.Run("only Cancelled set", func(t *testing.T) {
		trade := Trade{Cancelled: true}
		assert.False(t, trade.IsPhantomPrint())
		assert.False(t, trade.IsOptionsExpiration())
		assert.False(t, trade.IsVolatilityExpiration())
		assert.True(t, trade.IsCancelled())
	})
}

// TestTradeSemanticAccessors_MixedValues confirms accessors behave correctly
// when some bool fields are true and others are false, and frequency fields
// carry distinct non-zero values.
func TestTradeSemanticAccessors_MixedValues(t *testing.T) {
	trade := Trade{
		FrequencyLast30TD: 1,
		FrequencyLast90TD: 0,
		FrequencyLast1CY:  50,
		PhantomPrint:      true,
		OPEX:              false,
		VOLEX:             true,
		Cancelled:         false,
	}

	assert.Equal(t, 1, trade.SimilarTradeCountLast30Days())
	assert.Equal(t, 0, trade.SimilarTradeCountLast90Days())
	assert.Equal(t, 50, trade.SimilarTradeCountLastYear())
	assert.True(t, trade.IsPhantomPrint())
	assert.False(t, trade.IsOptionsExpiration())
	assert.True(t, trade.IsVolatilityExpiration())
	assert.False(t, trade.IsCancelled())
}
