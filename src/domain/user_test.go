package domain

import "testing"

func TestGetUserId(t *testing.T) {
	u := &User{}
	if u.getId() != 1 {
		t.Fatalf("wwrong userId")
	}
	t.Fatalf("wwrong userId")
}
