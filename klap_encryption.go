package tapo

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"log"
)

type KlapEncryptionSession struct {
	localSeed  []byte
	remoteSeed []byte
	userHash   []byte
	key        []byte
	iv         []byte
	seq        int32
	sig        []byte
}

func NewKlapEncryptionSession(localSeed, remoteSeed, userHash string) *KlapEncryptionSession {
	session := &KlapEncryptionSession{
		localSeed:  []byte(localSeed),
		remoteSeed: []byte(remoteSeed),
		userHash:   []byte(userHash),
	}

	session.key = session.keyDerive()
	session.iv, session.seq = session.ivDerive()
	session.sig = session.sigDerive()

	return session
}

func (s *KlapEncryptionSession) keyDerive() []byte {
	payload := append([]byte("lsk"), append(append(s.localSeed, s.remoteSeed...), s.userHash...)...)
	hash := sha256.Sum256(payload)
	return hash[:16]
}

func (s *KlapEncryptionSession) ivDerive() ([]byte, int32) {
	payload := append([]byte("iv"), append(append(s.localSeed, s.remoteSeed...), s.userHash...)...)
	fullIV := sha256.Sum256(payload)
	seq := int32(binary.BigEndian.Uint32(fullIV[12:]))
	return fullIV[:12], seq
}

func (s *KlapEncryptionSession) sigDerive() []byte {
	payload := append([]byte("ldk"), append(append(s.localSeed, s.remoteSeed...), s.userHash...)...)
	hash := sha256.Sum256(payload)
	return hash[:28]
}

func (s *KlapEncryptionSession) ivSeq() []byte {
	seqBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(seqBytes, uint32(s.seq))
	return append(s.iv, seqBytes...)
}

func (s *KlapEncryptionSession) encrypt(msg string) ([]byte, int32) {
	s.seq++
	msgBytes := []byte(msg)

	block, err := aes.NewCipher(s.key)
	if err != nil {
		log.Printf("Error creating AES cipher: %s", err)
		return nil, 0
	}

	cbc := cipher.NewCBCEncrypter(block, s.ivSeq())
	paddedData := pkcs7Pad(msgBytes, aes.BlockSize)
	ciphertext := make([]byte, len(paddedData))
	cbc.CryptBlocks(ciphertext, paddedData)

	hash := sha256.New()
	hash.Write(append(append(s.sig, seqToBytes(s.seq)...), ciphertext...))
	signature := hash.Sum(nil)

	return append(signature, ciphertext...), s.seq
}

func (s *KlapEncryptionSession) decrypt(msg []byte) []byte {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		log.Println("Error creating AES cipher:", err)
		return []byte("")
	}

	cbc := cipher.NewCBCDecrypter(block, s.ivSeq())
	plaintext := make([]byte, len(msg)-32)
	cbc.CryptBlocks(plaintext, msg[32:])

	unpaddedData, err := pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		log.Println("Error unpadding PKCS7:", err)
		return []byte("")
	}

	return unpaddedData
}

func seqToBytes(seq int32) []byte {
	seqBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(seqBytes, uint32(seq))
	return seqBytes
}
