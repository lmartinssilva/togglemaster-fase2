package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

// TestGenerateAPIKey_Format garante que a chave gerada tem o prefixo esperado
// e o comprimento correto (32 bytes em hex = 64 caracteres, + prefixo).
func TestGenerateAPIKey_Format(t *testing.T) {
	key, err := generateAPIKey()
	if err != nil {
		t.Fatalf("generateAPIKey retornou erro inesperado: %v", err)
	}

	if !strings.HasPrefix(key, "tm_key_") {
		t.Errorf("esperava prefixo 'tm_key_', obteve: %s", key)
	}

	hexPart := strings.TrimPrefix(key, "tm_key_")
	if len(hexPart) != 64 {
		t.Errorf("esperava 64 caracteres hex (32 bytes), obteve %d: %s", len(hexPart), hexPart)
	}
}

// TestGenerateAPIKey_Uniqueness garante que duas chamadas geram chaves diferentes
// (propriedade essencial de um gerador de chave seguro).
func TestGenerateAPIKey_Uniqueness(t *testing.T) {
	key1, err1 := generateAPIKey()
	key2, err2 := generateAPIKey()

	if err1 != nil || err2 != nil {
		t.Fatalf("generateAPIKey retornou erro: err1=%v err2=%v", err1, err2)
	}

	if key1 == key2 {
		t.Errorf("duas chamadas geraram a mesma chave, isso não deveria acontecer: %s", key1)
	}
}

// TestHashAPIKey_Deterministic garante que o hash da mesma chave é sempre igual
// (propriedade necessária para validar chaves salvas no banco).
func TestHashAPIKey_Deterministic(t *testing.T) {
	key := "tm_key_exemplo_fixo_para_teste"

	hash1 := hashAPIKey(key)
	hash2 := hashAPIKey(key)

	if hash1 != hash2 {
		t.Errorf("hash da mesma chave deveria ser igual, obteve %s e %s", hash1, hash2)
	}
}

// TestHashAPIKey_MatchesSHA256 garante que o hash gerado bate exatamente
// com o cálculo manual de SHA-256, confirmando que não há transformação
// adicional (como truncamento) sendo aplicada por engano.
func TestHashAPIKey_MatchesSHA256(t *testing.T) {
	key := "chave-de-teste-123"

	expected := sha256.Sum256([]byte(key))
	expectedHex := hex.EncodeToString(expected[:])

	got := hashAPIKey(key)

	if got != expectedHex {
		t.Errorf("hash incorreto.\nesperado: %s\nobtido:   %s", expectedHex, got)
	}

	if len(got) != 64 {
		t.Errorf("esperava hash SHA-256 com 64 caracteres hex, obteve %d", len(got))
	}
}
