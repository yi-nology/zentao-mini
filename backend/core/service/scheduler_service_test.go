package service

import (
	"strings"
	"testing"
)

func TestCronExprFieldCount(t *testing.T) {
	tests := []struct {
		name     string
		cronExpr string
		want     string // expected expression after prepending
	}{
		{"5-field weekday", "0 9 * * 1-5", "0 0 9 * * 1-5"},
		{"5-field daily", "0 9 * * *", "0 0 9 * * *"},
		{"5-field every 8h", "0 */8 * * *", "0 0 */8 * * *"},
		{"5-field every 30min", "*/30 * * * *", "0 */30 * * * *"},
		{"6-field already", "0 0 9 * * 1-5", "0 0 9 * * 1-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cronExpr := tt.cronExpr
			if len(strings.Fields(cronExpr)) == 5 {
				cronExpr = "0 " + cronExpr
			}
			if cronExpr != tt.want {
				t.Errorf("got %q, want %q", cronExpr, tt.want)
			}
		})
	}
}
