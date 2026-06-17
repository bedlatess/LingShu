package redeem

import "testing"

func TestHashNormalizesCode(t *testing.T) {
	if Hash(" ls-abcd ") != Hash("LS-ABCD") {
		t.Fatal("hash should normalize spaces and case")
	}
}
