package manage_start

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
)

const (
	secretsSubDir = "secrets"
	adminPassword = "admin"
)

// createSecrets creates random values used as secrets in Docker Compose file
// if the secrets don't already exist.
func createSecrets(d string) error {
	p := path.Join(d, secretsSubDir)
	if err := os.MkdirAll(p, fs.ModePerm); err != nil {
		return fmt.Errorf("creating directory `%s`: %w", p, err)
	}

	randomSecrets := []string{
		"auth_token_key",
		"auth_cookie_key",
	}
	for _, key := range randomSecrets {
		err := func() error {
			s := path.Join(p, key)
			if fileExists(s) {
				fmt.Printf("File %s does already exist. Skip this step.\n", s)
				return nil
			}

			f, err := os.Create(s)
			if err != nil {
				return fmt.Errorf("creating file `%s`: %w", path.Join(p, key), err)
			}
			defer f.Close()

			// This creates cryptographically secure random bytes. 32 bytes means
			// 256bit. The output can contain zero bytes.
			b64e := base64.NewEncoder(base64.StdEncoding, f)
			defer b64e.Close()
			if _, err := io.Copy(b64e, io.LimitReader(rand.Reader, 32)); err != nil {
				return fmt.Errorf("writing cryptographically secure random base64 encoded bytes: %w", err)
			}
			fmt.Printf("Successfully created file %s.\n", s)

			return nil
		}()
		if err != nil {
			return err
		}
	}

	a := path.Join(p, "admin")
	if fileExists(a) {
		fmt.Printf("File %s does already exist. Skip this step.\n", a)
		return nil
	}
	if err := os.WriteFile(a, []byte(adminPassword), 0666); err != nil {
		return fmt.Errorf("writing admin password to secret file: %w", err)
	}
	fmt.Printf("Successfully created file %s.\n", a)

	return nil
}

// populateSecrets is a small helper function that populates secret paths
// to the given template data.
func populateSecrets(td *tplData, p string) {
	secrets := map[string]string{
		"AuthTokenKey":  "auth_token_key",
		"AuthCookieKey": "auth_cookie_key",
		"Admin":         "admin",
	}
	td.Secret = make(map[string]string, len(secrets))
	for k, v := range secrets {
		td.Secret[k] = path.Join(p, secretsSubDir, v)
	}
}
