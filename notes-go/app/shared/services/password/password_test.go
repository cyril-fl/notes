package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	hash := Hash("correct horse")
	if !Verify("correct horse", hash) {
		t.Error("the right password is refused")
	}
	if Verify("wrong horse", hash) || Verify("correct horse", "nope") {
		t.Error("a wrong password or hash is accepted")
	}
	if Hash("correct horse") == hash {
		t.Error("two hashes of a password are equal: no salt")
	}
	if Token() == Token() || TokenHash("a") == TokenHash("b") {
		t.Error("tokens repeat")
	}
}
