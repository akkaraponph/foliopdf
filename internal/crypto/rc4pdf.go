// Package crypto implements PDF standard security handler encryption
// (V=1, R=2, 40-bit RC4) per PDF spec 1.4.
package crypto

import (
	"crypto/md5"
	"crypto/rc4"
)

// PDF encryption padding string (32 bytes) from the PDF spec.
var padding = [32]byte{
	0x28, 0xBF, 0x4E, 0x5E, 0x4E, 0x75, 0x8A, 0x41,
	0x64, 0x00, 0x4E, 0x56, 0xFF, 0xFA, 0x01, 0x08,
	0x2E, 0x2E, 0x00, 0xB6, 0xD0, 0x68, 0x3E, 0x80,
	0x2F, 0x0C, 0xA9, 0xFE, 0x64, 0x53, 0x69, 0x7A,
}

// KeyLength is the encryption key length in bytes for V=1 (40-bit).
const KeyLength = 5

// ComputeOwnerHash computes the /O value for the encrypt dictionary.
// ownerPw and userPw are the owner and user passwords.
func ComputeOwnerHash(ownerPw, userPw string) [32]byte {
	// Step 1: Pad owner password.
	ownerPadded := padPassword(ownerPw)

	// Step 2: MD5 hash of padded owner password.
	h := md5.Sum(ownerPadded[:])

	// Step 3: Use first KeyLength bytes as RC4 key.
	key := h[:KeyLength]

	// Step 4: Pad user password.
	userPadded := padPassword(userPw)

	// Step 5: RC4 encrypt padded user password.
	c, _ := rc4.NewCipher(key)
	var result [32]byte
	c.XORKeyStream(result[:], userPadded[:])
	return result
}

// ComputeEncryptionKey computes the file encryption key.
func ComputeEncryptionKey(userPw string, ownerHash [32]byte, permissions int32, fileID []byte) []byte {
	userPadded := padPassword(userPw)

	h := md5.New()
	h.Write(userPadded[:])
	h.Write(ownerHash[:])

	// Permission bytes (little-endian int32).
	p := uint32(permissions)
	h.Write([]byte{byte(p), byte(p >> 8), byte(p >> 16), byte(p >> 24)})

	h.Write(fileID)

	digest := h.Sum(nil)
	return digest[:KeyLength]
}

// ComputeUserHash computes the /U value for the encrypt dictionary.
func ComputeUserHash(encKey []byte) [32]byte {
	c, _ := rc4.NewCipher(encKey)
	var result [32]byte
	c.XORKeyStream(result[:], padding[:])
	return result
}

// EncryptData encrypts data for a specific PDF object using RC4.
// The per-object key is derived from the file key + object number + generation.
func EncryptData(encKey []byte, objNum, genNum int, data []byte) []byte {
	objKey := ObjectKey(encKey, objNum, genNum)
	c, _ := rc4.NewCipher(objKey)
	out := make([]byte, len(data))
	c.XORKeyStream(out, data)
	return out
}

// ObjectKey derives the per-object RC4 key.
func ObjectKey(encKey []byte, objNum, genNum int) []byte {
	// Key = MD5(encKey + objNum(3 bytes LE) + genNum(2 bytes LE))[:keyLen+5]
	h := md5.New()
	h.Write(encKey)
	h.Write([]byte{
		byte(objNum), byte(objNum >> 8), byte(objNum >> 16),
		byte(genNum), byte(genNum >> 8),
	})
	digest := h.Sum(nil)
	// Key length = min(encKeyLen + 5, 16)
	kl := len(encKey) + 5
	if kl > 16 {
		kl = 16
	}
	return digest[:kl]
}

// FileID generates a document ID from the creation metadata.
func FileID(title, producer string) []byte {
	h := md5.New()
	h.Write([]byte(title))
	h.Write([]byte(producer))
	// Add some variability.
	h.Write([]byte("folio-pdf"))
	return h.Sum(nil)
}

func padPassword(pw string) [32]byte {
	var padded [32]byte
	n := copy(padded[:], []byte(pw))
	copy(padded[n:], padding[:32-n])
	return padded
}

// Permission flag constants for PDF standard security handler.
const (
	PermPrint      = 1 << 2  // Allow printing
	PermModify     = 1 << 3  // Allow modifying contents
	PermCopy       = 1 << 4  // Allow copying/extracting text
	PermAnnotate   = 1 << 5  // Allow adding annotations
	PermAll        = PermPrint | PermModify | PermCopy | PermAnnotate
)
