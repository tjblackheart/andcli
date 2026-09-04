package ente

import (
	"crypto/subtle"
	"encoding/binary"
	"errors"

	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/poly1305"
)

// This implementation is mostly taken from the ente CLI.
// See https://github.com/ente/ente/tree/main/cli/internal/crypto

const (
	tagMessage = 0x00
	tagFinal   = 0x03
	keySize    = 32
	headerSize = 24
	overhead   = 17 // 1 tag byte + 16 MAC bytes
)

var pad0 [16]byte

type decryptor struct {
	k     [keySize]byte
	nonce [12]byte
}

func newStreamDecryptor(key, header []byte) (*decryptor, error) {
	if len(key) != keySize || len(header) != headerSize {
		return nil, errors.New("invalid key or header size")
	}

	k, err := chacha20.HChaCha20(key, header[:16])
	if err != nil {
		return nil, err
	}

	d := &decryptor{}
	copy(d.k[:], k)

	d.nonce[0] = 1 // counter starts at 1
	copy(d.nonce[4:], header[16:])

	return d, nil
}

func (d *decryptor) pull(cipher []byte) ([]byte, error) {
	if len(cipher) < overhead {
		return nil, errors.New("ciphertext too short")
	}
	mlen := len(cipher) - overhead

	ch, err := chacha20.NewUnauthenticatedCipher(d.k[:], d.nonce[:])
	if err != nil {
		return nil, err
	}

	var block [64]byte
	ch.XORKeyStream(block[:], block[:]) // keystream block #0

	var polyKey [32]byte
	copy(polyKey[:], block[:32])
	mac := poly1305.New(&polyKey)

	block = [64]byte{cipher[0]}
	ch.XORKeyStream(block[:], block[:]) // keystream block #1

	tag := block[0]
	block[0] = cipher[0] // encrypted tag byte goes into the MAC

	mac.Write(block[:]) // 64 bytes
	mac.Write(cipher[1 : 1+mlen])
	mac.Write(pad0[:mlen&0xf])

	var slen [8]byte
	binary.LittleEndian.PutUint64(slen[:], 0)
	mac.Write(slen[:])

	binary.LittleEndian.PutUint64(slen[:], 64+uint64(mlen))
	mac.Write(slen[:])

	if subtle.ConstantTimeCompare(mac.Sum(nil), cipher[1+mlen:]) != 1 {
		return nil, errors.New("authentication failed")
	}

	if tag != tagFinal && tag != tagMessage {
		return nil, errors.New("invalid tag")
	}

	plain := make([]byte, mlen)
	ch.XORKeyStream(plain, cipher[1:1+mlen])

	return plain, nil
}
