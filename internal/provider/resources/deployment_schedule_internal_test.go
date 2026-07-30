package resources

import "testing"

func TestNormalizeRRuleForState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server string
		prior  string
		want   string
	}{
		{
			name:   "empty server value passes through",
			server: "",
			prior:  "",
			want:   "",
		},
		{
			name:   "no DTSTART on either side passes through",
			server: "FREQ=DAILY;BYHOUR=10;BYMINUTE=30",
			prior:  "FREQ=DAILY;BYHOUR=10;BYMINUTE=30",
			want:   "FREQ=DAILY;BYHOUR=10;BYMINUTE=30",
		},
		{
			name:   "server prepends DTSTART, prior had none, strip it",
			server: "DTSTART:20200101T000000\nFREQ=DAILY;BYHOUR=10;BYMINUTE=30",
			prior:  "FREQ=DAILY;BYHOUR=10;BYMINUTE=30",
			want:   "FREQ=DAILY;BYHOUR=10;BYMINUTE=30",
		},
		{
			name:   "server prepends DTSTART, prior empty (Create path), strip it",
			server: "DTSTART:20200101T000000\nFREQ=DAILY;BYHOUR=10;BYMINUTE=30",
			prior:  "",
			want:   "FREQ=DAILY;BYHOUR=10;BYMINUTE=30",
		},
		{
			name:   "user-supplied DTSTART preserved verbatim",
			server: "DTSTART:20240101T000000\nFREQ=DAILY;BYHOUR=10;BYMINUTE=30",
			prior:  "DTSTART:20240101T000000\nFREQ=DAILY;BYHOUR=10;BYMINUTE=30",
			want:   "DTSTART:20240101T000000\nFREQ=DAILY;BYHOUR=10;BYMINUTE=30",
		},
		{
			name:   "user-supplied DTSTART;TZID preserved verbatim",
			server: "DTSTART;TZID=America/New_York:20240101T000000\nFREQ=DAILY",
			prior:  "DTSTART;TZID=America/New_York:20240101T000000\nFREQ=DAILY",
			want:   "DTSTART;TZID=America/New_York:20240101T000000\nFREQ=DAILY",
		},
		{
			name:   "case-insensitive DTSTART recognition on prior",
			server: "DTSTART:20240101T000000\nFREQ=DAILY",
			prior:  "dtstart:20240101T000000\nFREQ=DAILY",
			want:   "DTSTART:20240101T000000\nFREQ=DAILY",
		},
		{
			name:   "DTSTART without a trailing newline is left alone",
			server: "DTSTART:20200101T000000",
			prior:  "",
			want:   "DTSTART:20200101T000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeRRuleForState(tt.server, tt.prior)
			if got != tt.want {
				t.Errorf("normalizeRRuleForState(%q, %q) = %q; want %q", tt.server, tt.prior, got, tt.want)
			}
		})
	}
}

func TestHasDTStartPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"FREQ=DAILY", false},
		{"DTSTART", false},
		{"DTSTART:20200101T000000", true},
		{"dtstart:20200101T000000", true},
		{"DTSTART;TZID=UTC:20200101T000000", true},
		{"DTSTARTX:foo", false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			got := hasDTStartPrefix(tt.in)
			if got != tt.want {
				t.Errorf("hasDTStartPrefix(%q) = %v; want %v", tt.in, got, tt.want)
			}
		})
	}
}
