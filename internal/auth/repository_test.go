package auth

import (
	"os"
	"testing"
)

func TestSQLiteFixture(t *testing.T) {
	r, err := NewSQLiteRepository("file:testdata/fixture.db")
	if err != nil {
		t.Fatal(err)
	}

	_, err = r.Auth("bob", "bad password")
	if want, have := ErrBadAuth, err; want != have {
		t.Errorf("Auth with bad creds: want %v, have %v", want, have)
	}
	token, err := r.Auth("bob", "qwerty")
	if want, have := error(nil), err; want != have {
		t.Fatalf("Auth failed: %v", err)
	}

	if want, have := ErrBadAuth, r.Validate("bob", "bad token"); want != have {
		t.Errorf("Validate with bad token: want %v, have %v", want, have)
	}
	if want, have := error(nil), r.Validate("bob", token); want != have {
		t.Errorf("Validate: want %v, have %v", want, have)
	}

	if want, have := ErrBadAuth, r.Deauth("bob", "bad token"); want != have {
		t.Errorf("Deauth with bad token: want %v, have %v", want, have)
	}
	if want, have := error(nil), r.Deauth("bob", token); want != have {
		t.Errorf("Deauth: want %v, have %v", want, have)
	}
}

func TestSQLiteIntegration(t *testing.T) {
	var (
		filevar  = "AUTH_INTEGRATION_TEST_FILE"
		filename = os.Getenv(filevar)
	)
	if filename == "" {
		t.Skipf("skipping; set %s to run this test", filevar)
	}

	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Fatalf("%s: %v", filename, err)
	}

	defer func() {
		if err := os.Remove(filename); err != nil {
			t.Errorf("rm %s: %v", filename, err)
		}
	}()

	r, err := NewSQLiteRepository("file:" + filename)
	if err != nil {
		t.Fatal(err)
	}

	const (
		user = "alpha"
		pass = "beta"
	)
	if want, have := error(nil), r.Create(user, pass); want != have {
		t.Fatalf("Create: want %v, have %v", want, have)
	}

	token, err := r.Auth(user, pass)
	if want, have := error(nil), err; want != have {
		t.Fatalf("Auth: want %v, have %v", want, have)
	}

	if want, have := error(nil), r.Validate(user, token); want != have {
		t.Errorf("Validate: want %v, have %v", want, have)
	}

	if want, have := error(nil), r.Deauth(user, token); want != have {
		t.Errorf("Deauth: want %v, have %v", want, have)
	}
}
