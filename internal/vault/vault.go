// Package vault stores connection credentials in the pass(1) format:
// GPG-encrypted files under ~/.password-store/mash/.
//
// Each file is encrypted to the user's default GPG key. The first line
// holds the password and the second line holds "username: <user>".
// Files are compatible with the standard pass CLI so users can browse,
// edit and share credentials with pass if they choose.
package vault

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const storeSubdir = "mash"

// Credential holds a stored username / password pair for a connection.
type Credential struct {
	Username string
	Password string
}

// Store encrypts and persists a credential for the given connection
// name. An existing entry for the same name is overwritten.
func Store(name string, c Credential) error {
	dir, err := storeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("vault: mkdir: %w", err)
	}

	key, err := recipientKey()
	if err != nil {
		return fmt.Errorf("vault: recipient key: %w", err)
	}

	content := []byte(c.Password + "\nusername: " + c.Username + "\n")
	dst := filepath.Join(dir, sanitise(name)+".gpg")

	cmd := exec.Command("gpg",
		"--encrypt",
		"--batch",
		"--yes",
		"--trust-model", "always",
		"--recipient", key,
		"--output", dst,
	)
	cmd.Stdin = bytes.NewReader(content)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("vault: gpg encrypt: %w\n%s", err, string(out))
	}
	return nil
}

// Load decrypts and returns the credential for the given connection
// name. It returns an error if no credential exists.
func Load(name string) (Credential, error) {
	return load(sanitise(name))
}

func load(safeName string) (Credential, error) {
	dir, err := storeDir()
	if err != nil {
		return Credential{}, err
	}

	src := filepath.Join(dir, safeName+".gpg")
	if _, err := os.Stat(src); err != nil {
		return Credential{}, fmt.Errorf("vault: no credential for %q", safeName)
	}

	cmd := exec.Command("gpg", "--decrypt", "--batch", "--yes", src)
	out, err := cmd.Output()
	if err != nil {
		return Credential{}, fmt.Errorf("vault: gpg decrypt: %w", err)
	}

	return parseCredential(out)
}

// Has reports whether credentials are stored for name.
func Has(name string) bool {
	dir, err := storeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, sanitise(name)+".gpg"))
	return err == nil
}

// List returns all connection names that have stored credentials.
func List() ([]string, error) {
	dir, err := storeDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".gpg")
		if name != e.Name() {
			names = append(names, name)
		}
	}
	return names, nil
}

// Delete removes stored credentials for name. It is a no-op when the
// entry does not exist.
func Delete(name string) error {
	dir, err := storeDir()
	if err != nil {
		return err
	}
	dst := filepath.Join(dir, sanitise(name)+".gpg")
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ---- helpers --------------------------------------------------------------

// storeDir returns ~/.password-store/mash/, creating parent directories
// as needed.
func storeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("vault: home dir: %w", err)
	}
	return filepath.Join(home, ".password-store", storeSubdir), nil
}

// recipientKey returns the first available GPG secret key fingerprint.
// Users who want a specific key should set it as their default key in
// ~/.gnupg/gpg.conf (default-key <fingerprint>).
func recipientKey() (string, error) {
	cmd := exec.Command("gpg",
		"--list-secret-keys",
		"--with-colons",
		"--batch",
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gpg list-secret-keys: %w", err)
	}

	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "sec:") || strings.HasPrefix(line, "sec#") {
			// Field 4 is the key ID.
			fields := strings.Split(line, ":")
			if len(fields) > 4 && fields[4] != "" {
				return fields[4], nil
			}
		}
	}
	return "", fmt.Errorf("no GPG secret key found — create one with: gpg --gen-key")
}

// sanitise replaces characters that are unsafe in a file name.
func sanitise(name string) string {
	r := strings.NewReplacer(
		"/", "_",
		"..", "_",
		"\x00", "",
	)
	return r.Replace(name)
}

// parseCredential extracts the password (first line) and username from
// the decrypted pass-compatible content.
func parseCredential(raw []byte) (Credential, error) {
	var c Credential
	sc := bufio.NewScanner(bytes.NewReader(raw))

	// First line = password.
	if !sc.Scan() {
		return c, fmt.Errorf("vault: empty credential file")
	}
	c.Password = sc.Text()

	// Subsequent lines are key: value pairs. We look for "username:".
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "username:") {
			c.Username = strings.TrimSpace(strings.TrimPrefix(line, "username:"))
			break
		}
	}

	if c.Username == "" {
		return c, fmt.Errorf("vault: no username found in credential")
	}

	return c, nil
}
