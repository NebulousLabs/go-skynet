package skynet

import (
	"encoding/binary"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ed25519"
)

func encodeNumber(number uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, number)
	return b
}

func encodeString(toEncode string) []byte {
	encodedNumber := encodeNumber(uint64(len(toEncode)))
	return append(encodedNumber, []byte(toEncode)...)
}

func hashDataKey(dataKey string) []byte {
	encodedDataKey := encodeString(dataKey)
	hash := blake2b.Sum256(encodedDataKey)
	return hash[:]
}

func hashAll(args ...[]byte) []byte {
	var bytes []byte
	for _, arg := range args {
		bytes = append(bytes, arg...)
	}
	hash := blake2b.Sum256(bytes)
	return hash[:]
}

func hashRegistryEntry(s RegistryEntry) []byte {
	return hashAll(
		hashDataKey(s.DataKey),
		encodeString(s.Data),
		encodeNumber(s.Revision),
	)
}

func publicKeyFromPrivateKey(key ed25519.PrivateKey) ed25519.PublicKey {
	publicKey := make([]byte, 32)
	copy(publicKey, key[32:])
	return publicKey
}
