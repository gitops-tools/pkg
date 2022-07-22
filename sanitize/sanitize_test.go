package sanitize

import (
	"testing"
)

func TestSanitizeDNSName(t *testing.T) {
	sanitizeTests := []struct {
		raw  string
		want string
	}{
		{
			raw:  "$edgeAgent",
			want: "edgeagent",
		},
		{
			raw:  "$edgeHub",
			want: "edgehub",
		},
		// all characters are forced lowercase.
		{
			raw:  "ABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABC",
			want: "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabc",
		},
		// allow '-'
		{
			raw:  "ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABC",
			want: "abcdefghi-abcdefghi-abcdefghi-abcdefghi-abcdefghi-abcdefghi-abc",
		},
		{
			want: "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijab----------c",
			raw:  "ABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJAB----------C",
		},
		// must start with alphabet and end with alphanumeric
		{
			raw:  "---a-0---",
			want: "a-0",
		},
		{
			raw:  "---a-0---",
			want: "a-0",
		},
		{
			raw:  "---z-9---",
			want: "z-9",
		},
		{
			raw:  "---A-0---",
			want: "a-0",
		},
		{
			raw:  "---Z-9---",
			want: "z-9",
		},
		{
			raw:  "---a-z---",
			want: "a-z",
		},
		{
			raw:  "---a-z-/--1",
			want: "a-z---1",
		},
	}

	for _, tt := range sanitizeTests {
		t.Run(tt.raw, func(t *testing.T) {
			v, err := SanitizeDNSName(tt.raw)
			if err != nil {
				t.Fatal(err)
			}
			if v != tt.want {
				t.Fatalf("SanitizeDNSName() got %s, want %s", v, tt.want)
			}
		})
	}
}

func TestSanitizeDNSName_errors(t *testing.T) {
	sanitizeTests := []struct {
		raw     string
		wantErr string
	}{
		{
			raw:     "",
			wantErr: ErrEmptyName.Error(),
		},
		{
			raw:     "ABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJ",
			wantErr: `DNS name "ABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJ" exceeded maximum length of 63`,
		},
		{
			raw:     "ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHIJ",
			wantErr: `DNS name "ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHI-ABCDEFGHIJ" exceeded maximum length of 63`,
		},
		{
			raw:     "ABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJAB-------J",
			wantErr: `DNS name "ABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJABCDEFGHIJAB-------J" exceeded maximum length of 63`,
		},
		{
			raw:     "$$$$$$",
			wantErr: `DNS name "$$$$$$" sanitized is empty`,
		},
	}

	for _, tt := range sanitizeTests {
		t.Run(tt.raw, func(t *testing.T) {
			if _, err := SanitizeDNSName(tt.raw); err.Error() != tt.wantErr {
				t.Fatalf("SanitizeDNSName() got %s, want %s", err, tt.wantErr)
			}
		})
	}

}
