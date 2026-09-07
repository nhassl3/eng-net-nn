package hash

import (
	"encoding/base64"
	"fmt"
	"testing"

	"golang.org/x/crypto/argon2"
)

func argonIDKey(password string, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte {
	return argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
}

func formatHash(memory, time uint32, threads uint8, salt, hash []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, time, threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))
}

func TestVerifyPassword_UsesParamsFromHash(t *testing.T) {
	const password = "correct horse battery staple"

	encoded, err := CreateHashPassword(password)
	if err != nil {
		t.Fatalf("CreateHashPassword() error = %v", err)
	}

	ok, err := VerifyPassword(password, encoded)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !ok {
		t.Fatal("VerifyPassword() = false, want true")
	}

	ok, err = VerifyPassword("wrong password", encoded)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if ok {
		t.Fatal("VerifyPassword() = true, want false for wrong password")
	}
}

func TestVerifyPassword_OldParamsStillVerify(t *testing.T) {
	// Simulates a hash created with weaker, "old" parameters than the
	// package's current constants, to guard against regressions where
	// VerifyPassword recomputes with the package constants instead of the
	// parameters embedded in the hash.
	const password = "legacy password"

	salt := make([]byte, saltLen)
	for i := range salt {
		salt[i] = byte(i)
	}

	const oldTime, oldMemory, oldThreads, oldKeyLen = 1, 8 * 1024, 1, 16

	oldHashBytes := argonIDKey(password, salt, oldTime, oldMemory, oldThreads, oldKeyLen)
	encoded := formatHash(oldMemory, oldTime, oldThreads, salt, oldHashBytes)

	ok, err := VerifyPassword(password, encoded)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !ok {
		t.Fatal("VerifyPassword() = false, want true for hash created with old params")
	}
}

func TestNeedsRehash(t *testing.T) {
	const password = "some password"

	current, err := CreateHashPassword(password)
	if err != nil {
		t.Fatalf("CreateHashPassword() error = %v", err)
	}

	needs, err := NeedsRehash(current)
	if err != nil {
		t.Fatalf("NeedsRehash() error = %v", err)
	}
	if needs {
		t.Fatal("NeedsRehash() = true for hash created with current params, want false")
	}

	salt := make([]byte, saltLen)
	weakHash := argonIDKey(password, salt, 1, 8*1024, 1, argonKeyLen)
	weakEncoded := formatHash(8*1024, 1, 1, salt, weakHash)

	needs, err = NeedsRehash(weakEncoded)
	if err != nil {
		t.Fatalf("NeedsRehash() error = %v", err)
	}
	if !needs {
		t.Fatal("NeedsRehash() = false for weak hash, want true")
	}
}
