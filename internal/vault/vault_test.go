package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCredential(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantUser string
		wantPass string
		wantErr  bool
	}{
		{
			name:     "standard",
			raw:      "s3cret\nusername: alice\n",
			wantPass: "s3cret",
			wantUser: "alice",
		},
		{
			name:     "extra lines ignored",
			raw:      "mypass\nusername: bob\nurl: example.com\nnotes: foo\n",
			wantPass: "mypass",
			wantUser: "bob",
		},
		{
			name:    "no username line",
			raw:     "onlypass\n",
			wantErr: true,
		},
		{
			name:    "empty",
			raw:     "",
			wantErr: true,
		},
		{
			name:     "password with spaces",
			raw:      "my pass word\nusername: charlie\n",
			wantPass: "my pass word",
			wantUser: "charlie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := parseCredential([]byte(tt.raw))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.Password != tt.wantPass {
				t.Errorf("password = %q, want %q", c.Password, tt.wantPass)
			}
			if c.Username != tt.wantUser {
				t.Errorf("username = %q, want %q", c.Username, tt.wantUser)
			}
		})
	}
}

func TestSanitise(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"prod-web-01", "prod-web-01"},
		{"path/to/host", "path_to_host"},
		{"a..b", "a_b"},
		{"normal", "normal"},
	}
	for _, tt := range tests {
		got := sanitise(tt.in)
		if got != tt.want {
			t.Errorf("sanitise(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestStoreDir(t *testing.T) {
	dir, err := storeDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir == "" {
		t.Fatal("empty store dir")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("store dir must be absolute, got %s", dir)
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		t.Skip("HOME not set")
	}
	expected := filepath.Join(home, ".password-store", "mash")
	if dir != expected {
		t.Errorf("storeDir = %s, want %s", dir, expected)
	}
}

func TestRecipientKey_NoKey(t *testing.T) {
	_, err := recipientKey()
	if err != nil {
		t.Logf("No GPG key found (expected if none set up): %v", err)
		return
	}
	// If we get here, a key exists — good.
}
