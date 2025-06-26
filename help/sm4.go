package help

import (
	"bytes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"github.com/emmansun/gmsm/sm4"
)

func SM4Encrypt(hexkey string, plaintext string) (ciphertext string, err error) {
	var key []byte
	if key, err = hex.DecodeString(hexkey); err != nil {
		return
	}

	var block cipher.Block
	if block, err = sm4.NewCipher(key); err != nil {
		return
	}

	b := []byte(plaintext)
	b = pkcs5Padding(b, block.BlockSize())

	content := make([]byte, len(b))
	mode := newECBEncrypter(block)
	mode.CryptBlocks(content, b)

	return hex.EncodeToString(content), nil
}

func SM4Decrypt(hexkey string, ciphertext string) (plaintext string, err error) {
	var key []byte
	if key, err = hex.DecodeString(hexkey); err != nil {
		return
	}

	var block cipher.Block
	if block, err = sm4.NewCipher(key); err != nil {
		return
	}

	var content []byte
	if content, err = hex.DecodeString(ciphertext); err != nil {
		return
	}

	b := make([]byte, len(content))
	mode := newECBDecrypter(block)
	mode.CryptBlocks(b, content)

	var unpadding []byte
	if unpadding, err = pkcs5UnPadding(b); err != nil {
		return
	}

	return string(unpadding), nil
}

func SM4Verify(key string, ciphertext string, plaintext string) (r bool, err error) {
	var decryptText string
	if decryptText, err = SM4Decrypt(key, ciphertext); err != nil {
		return
	}
	r = decryptText == plaintext
	return
}

type ecb struct {
	b         cipher.Block
	blockSize int
}

func NewECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypter ecb

func newECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(NewECB(b))
}

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("the input data length must be a multiple of the block size")
	}
	if len(dst) < len(src) {
		panic("the output buffer length is insufficient")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecrypter ecb

func newECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(NewECB(b))
}

func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("the input data length must be a multiple of the block size")
	}
	if len(dst) < len(src) {
		panic("the output buffer length is insufficient")
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

func pkcs5Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func pkcs5UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("the data is empty")
	}
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, errors.New("filling format error")
	}
	return data[:(length - unpadding)], nil
}
