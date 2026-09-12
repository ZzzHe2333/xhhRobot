package xhh

import "testing"

func TestQRStateQuery(t *testing.T) {
	tests := []struct {
		name    string
		qrURL   string
		want    string
		wantErr bool
	}{
		{
			name:  "absolute url",
			qrURL: "https://api.xiaoheihe.cn/account/qr_login/?ticket=abc123&device=web",
			want:  "?ticket=abc123&device=web",
		},
		{
			name:  "relative url",
			qrURL: "/account/qr_login/?ticket=abc123",
			want:  "?ticket=abc123",
		},
		{
			name:    "missing query",
			qrURL:   "https://api.xiaoheihe.cn/account/qr_login/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := qrStateQuery(tt.qrURL)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("qrStateQuery(%q) error = nil, want error", tt.qrURL)
				}
				return
			}
			if err != nil {
				t.Fatalf("qrStateQuery(%q) unexpected error: %v", tt.qrURL, err)
			}
			if got != tt.want {
				t.Fatalf("qrStateQuery(%q) = %q, want %q", tt.qrURL, got, tt.want)
			}
		})
	}
}
