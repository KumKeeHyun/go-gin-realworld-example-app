package crypto

import "testing"

func TestHashPassword(t *testing.T) {
	expected := "test password"
	actural, err := HashPassword(expected)
	if err != nil {
		t.Errorf("generate hash error = %v", err)
	}
	if !CheckHashAndPassword(actural, expected) {
		t.Error("check password expect true, got false")
	}
}
