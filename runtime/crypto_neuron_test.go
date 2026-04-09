package main

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

// ---------------------------------------------------------------------------
// Functional Tests
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	dek, err := GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK failed: %v", err)
	}

	original := []byte("brainstem/canon/절대_폴백_금지: counter=103")
	ct, nonce, err := EncryptNeuron(original, dek)
	if err != nil {
		t.Fatalf("EncryptNeuron failed: %v", err)
	}

	if bytes.Equal(ct, original) {
		t.Fatal("ciphertext must differ from plaintext")
	}

	pt, err := DecryptNeuron(ct, dek, nonce)
	if err != nil {
		t.Fatalf("DecryptNeuron failed: %v", err)
	}

	if !bytes.Equal(pt, original) {
		t.Fatalf("round-trip failed: got %q, want %q", pt, original)
	}
	t.Logf("OK: round-trip encrypt/decrypt verified (%d→%d→%d bytes)", len(original), len(ct), len(pt))
}

func TestEncrypt_EmptyPlaintext(t *testing.T) {
	dek, _ := GenerateDEK()
	_, _, err := EncryptNeuron(nil, dek)
	if err != errEmptyData {
		t.Fatalf("expected errEmptyData, got: %v", err)
	}
	t.Log("OK: empty plaintext correctly rejected")
}

func TestEncrypt_WrongKeyLength(t *testing.T) {
	_, _, err := EncryptNeuron([]byte("data"), []byte("short"))
	if err != errDEKLength {
		t.Fatalf("expected errDEKLength, got: %v", err)
	}
	t.Log("OK: invalid DEK length correctly rejected")
}

func TestDecrypt_WrongKey(t *testing.T) {
	dek1, _ := GenerateDEK()
	dek2, _ := GenerateDEK()

	ct, nonce, _ := EncryptNeuron([]byte("secret neuron data"), dek1)
	_, err := DecryptNeuron(ct, dek2, nonce)
	if err != errDecrypt {
		t.Fatalf("expected errDecrypt with wrong key, got: %v", err)
	}
	t.Log("OK: wrong DEK correctly rejected (tamper detection)")
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	dek, _ := GenerateDEK()
	ct, nonce, _ := EncryptNeuron([]byte("neuron integrity data"), dek)

	// Flip a byte in ciphertext
	tampered := make([]byte, len(ct))
	copy(tampered, ct)
	tampered[len(tampered)/2] ^= 0xFF

	_, err := DecryptNeuron(tampered, dek, nonce)
	if err != errDecrypt {
		t.Fatalf("expected errDecrypt for tampered ciphertext, got: %v", err)
	}
	t.Log("OK: tampered ciphertext correctly detected")
}

func TestDecrypt_WrongNonceLength(t *testing.T) {
	dek, _ := GenerateDEK()
	ct, _, _ := EncryptNeuron([]byte("data"), dek)
	_, err := DecryptNeuron(ct, dek, []byte("short-nonce"))
	if err != errNonceLength {
		t.Fatalf("expected errNonceLength, got: %v", err)
	}
	t.Log("OK: invalid nonce length correctly rejected")
}

func TestEncrypt_UniqueNonces(t *testing.T) {
	dek, _ := GenerateDEK()
	data := []byte("same data for nonce uniqueness test")

	nonces := make(map[string]bool)
	for i := 0; i < 100; i++ {
		_, nonce, err := EncryptNeuron(data, dek)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		key := string(nonce)
		if nonces[key] {
			t.Fatalf("nonce collision at iteration %d", i)
		}
		nonces[key] = true
	}
	t.Log("OK: 100 nonces all unique (no collision)")
}

func TestGenerateDEK_Length(t *testing.T) {
	dek, err := GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK: %v", err)
	}
	if len(dek) != chacha20poly1305.KeySize {
		t.Fatalf("DEK length: got %d, want %d", len(dek), chacha20poly1305.KeySize)
	}
	t.Logf("OK: DEK generated (%d bytes)", len(dek))
}

// ---------------------------------------------------------------------------
// Benchmarks: 1KB, 100KB, 1MB neuron files
// ---------------------------------------------------------------------------

func benchmarkEncrypt(b *testing.B, size int) {
	dek, _ := GenerateDEK()
	data := make([]byte, size)
	rand.Read(data)

	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = EncryptNeuron(data, dek)
	}
}

func benchmarkDecrypt(b *testing.B, size int) {
	dek, _ := GenerateDEK()
	data := make([]byte, size)
	rand.Read(data)
	ct, nonce, _ := EncryptNeuron(data, dek)

	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecryptNeuron(ct, dek, nonce)
	}
}

func BenchmarkEncrypt_1KB(b *testing.B)   { benchmarkEncrypt(b, 1024) }
func BenchmarkEncrypt_100KB(b *testing.B) { benchmarkEncrypt(b, 100*1024) }
func BenchmarkEncrypt_1MB(b *testing.B)   { benchmarkEncrypt(b, 1024*1024) }

func BenchmarkDecrypt_1KB(b *testing.B)   { benchmarkDecrypt(b, 1024) }
func BenchmarkDecrypt_100KB(b *testing.B) { benchmarkDecrypt(b, 100*1024) }
func BenchmarkDecrypt_1MB(b *testing.B)   { benchmarkDecrypt(b, 1024*1024) }

// ---------------------------------------------------------------------------
// Integrated benchmark report (runs as test, not go bench)
// ---------------------------------------------------------------------------

func TestCryptoBenchmarkReport(t *testing.T) {
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
	}

	dek, _ := GenerateDEK()
	iterations := 1000

	t.Log("")
	t.Log("╔══════════════════════════════════════════════════════════╗")
	t.Log("║    NeuronFS XChaCha20-Poly1305 Benchmark Report        ║")
	t.Log("╠══════════════════════════════════════════════════════════╣")
	t.Logf("║  %-10s │ %-12s │ %-12s │ %-8s ║", "Size", "Encrypt", "Decrypt", "Iters")
	t.Log("╠══════════════════════════════════════════════════════════╣")

	for _, s := range sizes {
		data := make([]byte, s.size)
		rand.Read(data)

		// Encrypt benchmark
		start := time.Now()
		var ct []byte
		var nonce []byte
		for i := 0; i < iterations; i++ {
			ct, nonce, _ = EncryptNeuron(data, dek)
		}
		encDur := time.Since(start) / time.Duration(iterations)

		// Decrypt benchmark
		start = time.Now()
		for i := 0; i < iterations; i++ {
			_, _ = DecryptNeuron(ct, dek, nonce)
		}
		decDur := time.Since(start) / time.Duration(iterations)

		t.Logf("║  %-10s │ %12s │ %12s │ %8d ║", s.name, encDur, decDur, iterations)
	}

	t.Log("╚══════════════════════════════════════════════════════════╝")
}
