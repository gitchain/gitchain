package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeECDSAPrivateKey(t *testing.T) {
	key, err := GenerateECDSA()
	if err != nil {
		t.Errorf("error while generating ECDSA key: %v", err)
	}
	bytes, err := EncodeECDSAPrivateKey(key)
	if err != nil {
		t.Errorf("error while encoding ECDSA key: %v", err)
	}
	key1, err := DecodeECDSAPrivateKey(bytes)
	if err != nil {
		t.Errorf("error while decoding ECDSA key: %v", err)
	}

	assert.Equal(t, key, key1)

}
func TestEncodeDecodeECDSAPublicKey(t *testing.T) {
	key, err := GenerateECDSA()
	if err != nil {
		t.Errorf("error while generating ECDSA key: %v", err)
	}
	bytes, err := EncodeECDSAPublicKey(&key.PublicKey)
	if err != nil {
		t.Errorf("error while encoding ECDSA key: %v", err)
	}
	key1, err := DecodeECDSAPublicKey(bytes)
	if err != nil {
		t.Errorf("error while decoding ECDSA key: %v", err)
	}

	assert.Equal(t, &key.PublicKey, key1)

}

func TestECDSAPrivateKeyEquality(t *testing.T) {
	key, err := GenerateECDSA()
	if err != nil {
		t.Errorf("error while generating ECDSA key: %v", err)
	}
	key1, err := GenerateECDSA()
	if err != nil {
		t.Errorf("error while generating ECDSA key: %v", err)
	}

	v, err := EqualECDSAPrivateKeys(key, key)
	if err != nil {
		t.Errorf("error while checking ECDSA keys equality: %v", err)
	}
	assert.True(t, v)
	v1, err := EqualECDSAPrivateKeys(key, key1)
	if err != nil {
		t.Errorf("error while checking ECDSA keys equality: %v", err)
	}
	assert.False(t, v1)
}

func TestEncodeECDSAPublicKey(t *testing.T) {
	key, err := GenerateECDSA()
	if err != nil {
		t.Errorf("error while generating ECDSA key: %v", err)
	}
	// FIXME: That's not a very comprehensive test
	assert.NotEmpty(t, ECDSAPublicKeyToString(key.PublicKey))
}
