package seasnve

import (
	"testing"
	"os"
)

func TestFailedLogin(t * testing.T) {
	c := NewClient()

	err := c.Login("bad-username", "wrong-password")

	if err == nil {
		t.Errorf("Expected error, got nothing")
	}

	if err.Error() != "Not authorized. Wrong username or password" {
		t.Errorf("Expected error `Not authorized. Wrong username or password`, got `%s`", err.Error())
	}
}

func TestFull(t *testing.T) {
	email:= os.Getenv("EMAIL")
	password:= os.Getenv("PASSWORD")

	c := NewClient()

	err := c.Login(email, password)
	if (err != nil) {
		t.Errorf("Unexpected error logging in: %s", err)
	}
}
