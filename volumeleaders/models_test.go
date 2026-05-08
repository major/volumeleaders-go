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

func TestTradeNullableStringFields(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		raw := `{"Industry":null,"FullDateTime":null,"FullTimeString24":null}`
		var trade Trade
		require.NoError(t, json.Unmarshal([]byte(raw), &trade), "Unmarshal(null fields)")
		assert.Nil(t, trade.Industry, "Trade.Industry")
		assert.Nil(t, trade.FullDateTime, "Trade.FullDateTime")
		assert.Nil(t, trade.FullTimeString24, "Trade.FullTimeString24")
	})

	t.Run("non-null", func(t *testing.T) {
		raw := `{"Industry":"Consumer Finance","FullDateTime":"2026-05-01T16:20:51","FullTimeString24":"16:20:51"}`
		var trade Trade
		require.NoError(t, json.Unmarshal([]byte(raw), &trade), "Unmarshal(non-null fields)")
		require.NotNil(t, trade.Industry, "Trade.Industry")
		assert.Equal(t, "Consumer Finance", *trade.Industry, "Trade.Industry value")
		require.NotNil(t, trade.FullDateTime, "Trade.FullDateTime")
		assert.Equal(t, "2026-05-01T16:20:51", *trade.FullDateTime, "Trade.FullDateTime value")
		require.NotNil(t, trade.FullTimeString24, "Trade.FullTimeString24")
		assert.Equal(t, "16:20:51", *trade.FullTimeString24, "Trade.FullTimeString24 value")
	})

	t.Run("marshal null emits null", func(t *testing.T) {
		trade := Trade{}
		data, err := json.Marshal(trade)
		require.NoError(t, err, "Marshal(zero-value Trade)")
		var raw map[string]any
		require.NoError(t, json.Unmarshal(data, &raw), "Unmarshal(marshaled Trade)")

		for _, field := range []string{"Industry", "FullDateTime", "FullTimeString24"} {
			val, ok := raw[field]
			assert.True(t, ok, "marshaled JSON should contain %s field", field)
			assert.Nil(t, val, "marshaled Trade.%s should be null", field)
		}
	})
}

func TestFlexBoolMarshalJSON(t *testing.T) {
	data, err := json.Marshal(FlexBool(true))
	require.NoError(t, err, "Marshal(FlexBool(true))")
	assert.Equal(t, "true", string(data), "Marshal(FlexBool(true))")

	data, err = json.Marshal(FlexBool(false))
	require.NoError(t, err, "Marshal(FlexBool(false))")
	assert.Equal(t, "false", string(data), "Marshal(FlexBool(false))")
}
