package config

import "testing"

func TestResolve(t *testing.T) {
	tests := []struct {
		name         string
		flagEndpoint string
		flagToken    string
		envEndpoint  string
		envToken     string
		wantBaseURL  string
		wantToken    string
		wantErr      bool
	}{
		{
			name:        "from env",
			envEndpoint: "whatwedo.mocoapp.com",
			envToken:    "abc",
			wantBaseURL: "https://whatwedo.mocoapp.com/api/v1",
			wantToken:   "abc",
		},
		{
			name:         "flag beats env",
			flagEndpoint: "flag.mocoapp.com",
			flagToken:    "flagtoken",
			envEndpoint:  "env.mocoapp.com",
			envToken:     "envtoken",
			wantBaseURL:  "https://flag.mocoapp.com/api/v1",
			wantToken:    "flagtoken",
		},
		{
			name:         "https prefix is normalized",
			flagEndpoint: "https://whatwedo.mocoapp.com",
			flagToken:    "t",
			wantBaseURL:  "https://whatwedo.mocoapp.com/api/v1",
			wantToken:    "t",
		},
		{
			name:         "trailing slash and path are stripped",
			flagEndpoint: "https://whatwedo.mocoapp.com/api/v1/",
			flagToken:    "t",
			wantBaseURL:  "https://whatwedo.mocoapp.com/api/v1",
			wantToken:    "t",
		},
		{
			name:         "any scheme is normalized to https",
			flagEndpoint: "whatever://whatwedo.mocoapp.com",
			flagToken:    "t",
			wantBaseURL:  "https://whatwedo.mocoapp.com/api/v1",
			wantToken:    "t",
		},
		{
			name:     "missing endpoint",
			envToken: "t",
			wantErr:  true,
		},
		{
			name:         "missing token",
			flagEndpoint: "whatwedo.mocoapp.com",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(EnvEndpoint, tt.envEndpoint)
			t.Setenv(EnvToken, tt.envToken)

			cfg, err := Resolve(tt.flagEndpoint, tt.flagToken)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := cfg.BaseURL(); got != tt.wantBaseURL {
				t.Errorf("BaseURL = %q, want %q", got, tt.wantBaseURL)
			}
			if cfg.Token != tt.wantToken {
				t.Errorf("Token = %q, want %q", cfg.Token, tt.wantToken)
			}
		})
	}
}
