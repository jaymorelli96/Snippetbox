package main

import (
	"testing"
	"time"

	"snippetbox.jmorelli.dev/internal/assert"
)

func TestHumanDate(t *testing.T) {

	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2023, 4, 16, 11, 11, 0, 0, time.UTC),
			want: "16 Apr 2023 at 11:11",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2023, 4, 16, 11, 11, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "16 Apr 2023 at 10:11",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)

			assert.Equal(t, hd, tt.want)
		})
	}
}
