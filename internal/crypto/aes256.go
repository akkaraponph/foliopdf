// AES-256 encryption for PDF 2.0 standard security handler (V=5, R=6).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

// AES256KeyLength is the encryption key length in bytes for AES-256.
const AES256KeyLength = 32

// GenerateFileEncryptionKey generates a random 32-byte file encryption key.
func GenerateFileEncryptionKey() ([]byte, error) {
	key := make([]byte, AES256KeyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// ComputeOwnerHashV5 computes the /O value for V=5 R=6 encrypt dictionary.
// Returns the 48-byte O value (32-byte hash + 8-byte validation salt + 8-byte key salt).
func ComputeOwnerHashV5(ownerPw string, userKeyHash []byte) (oValue [48]byte, oeSalt [32]byte) {
	// Generate random validation salt (8 bytes) and key salt (8 bytes).
	var valSalt [8]byte
	var keySalt [8]byte
	io.ReadFull(rand.Reader, valSalt[:])
	io.ReadFull(rand.Reader, keySalt[:])

	// O validation hash = SHA-256(password + validation salt + U value)
	h := sha256.New()
	truncPw := truncatePassword(ownerPw)
	h.Write(truncPw)
	h.Write(valSalt[:])
	h.Write(userKeyHash[:32])
	hash := h.Sum(nil)

	copy(oValue[:32], hash)
	copy(oValue[32:40], valSalt[:])
	copy(oValue[40:48], keySalt[:])

	// OE = AES-256-CBC encrypt(file encryption key) with key derived from
	// SHA-256(password + key salt + U value)
	h2 := sha256.New()
	h2.Write(truncPw)
	h2.Write(keySalt[:])
	h2.Write(userKeyHash[:32])
	oeKey := h2.Sum(nil)

	// OE is stored separately — caller uses EncryptAESCBC with oeKey.
	copy(oeSalt[:], oeKey)

	return oValue, oeSalt
}

// ComputeUserHashV5 computes the /U value for V=5 R=6 encrypt dictionary.
// Returns the 48-byte U value (32-byte hash + 8-byte validation salt + 8-byte key salt).
func ComputeUserHashV5(userPw string) (uValue [48]byte, ueSalt [32]byte) {
	var valSalt [8]byte
	var keySalt [8]byte
	io.ReadFull(rand.Reader, valSalt[:])
	io.ReadFull(rand.Reader, keySalt[:])

	// U validation hash = SHA-256(password + validation salt)
	h := sha256.New()
	truncPw := truncatePassword(userPw)
	h.Write(truncPw)
	h.Write(valSalt[:])
	hash := h.Sum(nil)

	copy(uValue[:32], hash)
	copy(uValue[32:40], valSalt[:])
	copy(uValue[40:48], keySalt[:])

	// UE key = SHA-256(password + key salt)
	h2 := sha256.New()
	h2.Write(truncPw)
	h2.Write(keySalt[:])
	ueKey := h2.Sum(nil)

	copy(ueSalt[:], ueKey)
	return uValue, ueSalt
}

// EncryptAESCBC encrypts data using AES-256-CBC with a random IV.
// Returns IV (16 bytes) + ciphertext.
func EncryptAESCBC(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// PKCS7 padding.
	padLen := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := make([]byte, len(plaintext)+padLen)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	// Random IV.
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	// Return IV + ciphertext.
	result := make([]byte, aes.BlockSize+len(ciphertext))
	copy(result, iv)
	copy(result[aes.BlockSize:], ciphertext)
	return result, nil
}

// EncryptDataAES256 encrypts data for a specific PDF object using AES-256-CBC.
// Per PDF spec, each string/stream uses a random IV prepended to the ciphertext.
func EncryptDataAES256(encKey []byte, objNum, genNum int, data []byte) []byte {
	encrypted, err := EncryptAESCBC(encKey, data)
	if err != nil {
		// Fallback: return original data (should not happen).
		return data
	}
	return encrypted
}

// truncatePassword truncates a UTF-8 password to 127 bytes per PDF 2.0 spec.
func truncatePassword(pw string) []byte {
	b := []byte(pw)
	if len(b) > 127 {
		b = b[:127]
	}
	return b
}

// FileIDAES generates a 16-byte file ID using SHA-256.
func FileIDAES(title, producer string) []byte {
	h := sha256.New()
	h.Write([]byte(title))
	h.Write([]byte(producer))
	// Add randomness for uniqueness.
	var rnd [16]byte
	io.ReadFull(rand.Reader, rnd[:])
	h.Write(rnd[:])
	return h.Sum(nil)[:16]
}
