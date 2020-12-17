package skynet

import (
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ed25519"
)

// encodeNumber converts the given number into a byte array.
func encodeNumber(number uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, number)
	return b
}

// encodeString converts the given string into a byte array.
func encodeString(toEncode string) []byte {
	encodedNumber := encodeNumber(uint64(len(toEncode)))
	return append(encodedNumber, []byte(toEncode)...)
}

// hashDataKey hashes the given data key.
func hashDataKey(dataKey string) []byte {
	encodedDataKey := encodeString(dataKey)
	hash := blake2b.Sum256(encodedDataKey)
	return hash[:]
}

// hashAll takes all given arguments and hashes them.
func hashAll(args ...[]byte) []byte {
	var bytes []byte
	for _, arg := range args {
		bytes = append(bytes, arg...)
	}
	hash := blake2b.Sum256(bytes)
	return hash[:]
}

// hashRegistryEntry hashes the given registry entry.
func hashRegistryEntry(s RegistryEntry) []byte {
	return hashAll(
		hashDataKey(s.DataKey),
		encodeString(s.Data),
		encodeNumber(s.Revision),
	)
}

// publicKeyFromPrivateKey return publicKey from privateKey.
func publicKeyFromPrivateKey(key ed25519.PrivateKey) (ed25519.PublicKey, error) {
	if len(key) != 64 {
		return nil, errors.New("invalid privateKey")
	}

	publicKey := make([]byte, 32)
	copy(publicKey, key[32:])

	return publicKey, nil
}
