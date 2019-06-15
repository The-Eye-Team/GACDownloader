package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

const (
	ImageKeyHex = "5b63db113b7af3e0b1435556c8f9530c"
	ImageIvHex  = "71e70405353a778bfa6fbc30321b9592"
)

var (
	ImageKey, _ = hex.DecodeString(ImageKeyHex)
	ImageIv, _  = hex.DecodeString(ImageIvHex)
	FileHeader  = []byte{0x0A, 0x0A, 0x0A, 0x0A}
)

func splitInThree(buffer []byte, begin, end int) ([]byte, []byte, []byte) {
	return buffer[:begin], buffer[begin:end], buffer[end:]
}

func decodeImage(encoded []byte) (decoded []byte, err error) {
	encryptionMarker, body, indexBytes := splitInThree(encoded, 4, len(encoded)-4)

	if !bytes.Equal(encryptionMarker, FileHeader) {
		return nil, errors.New("can't find Encryption marker")
	}

	index := int(binary.LittleEndian.Uint32(indexBytes))

	decoded, replaceCountBytes, rest := splitInThree(body, index, index+4)

	replaceCount := int(binary.LittleEndian.Uint32(replaceCountBytes))

	_, encrypted, clearSuffix := splitInThree(rest, 0, replaceCount)

	block, err := aes.NewCipher(ImageKey)
	if err != nil {
		return nil, err
	}

	// CBC mode always works in whole blocks.
	if len(encrypted)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, ImageIv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(encrypted, encrypted)

	decoded = append(decoded, encrypted...)
	decoded = append(decoded, clearSuffix...)

	return decoded, nil
}