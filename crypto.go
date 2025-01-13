package tapo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

func sha256HashUpperCase(payload []byte) string {
	hash := sha256.Sum256(payload)
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func sha256Hash(payload []byte) []byte {
	hash := sha256.Sum256(payload)
	return hash[:]
}

// AES provides methods for encrypting and decrypting using AES in CBC mode with PKCS7 padding.
type AES struct {
	key []byte
	iv  []byte
}

// NewAES creates a new AES instance.
func NewAES(key, iv []byte) (*AES, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("key must be 16, 24, or 32 bytes")
	}
	if len(iv) != aes.BlockSize {
		return nil, errors.New("iv must be 16 bytes")
	}
	return &AES{key: key, iv: iv}, nil
}

// Encrypt encrypts the input data using AES CBC with PKCS7 padding.
func (a *AES) Encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}

	// Add PKCS7 padding
	paddedData := pkcs7Pad(data, aes.BlockSize)
	ciphertext := make([]byte, len(paddedData))

	mode := cipher.NewCBCEncrypter(block, a.iv)
	mode.CryptBlocks(ciphertext, paddedData)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the base64-encoded input data using AES CBC with PKCS7 unpadding.
func (a *AES) Decrypt(data string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return []byte(""), err
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return []byte(""), err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return []byte(""), errors.New("ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, a.iv)
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	plaintext, err = pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return []byte(""), err
	}

	return plaintext, nil
}

// pkcs7Pad adds PKCS7 padding to the data.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: Data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: Data is not block-aligned")
	}
	padLen := int(data[length-1])
	ref := bytes.Repeat([]byte{byte(padLen)}, padLen)
	if padLen > blockSize || padLen == 0 || !bytes.HasSuffix(data, ref) {
		return nil, errors.New("pkcs7: Invalid padding")
	}
	return data[:length-padLen], nil
}
