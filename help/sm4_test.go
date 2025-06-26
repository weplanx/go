package help

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSM4(t *testing.T) {
	key := `f93c920868b4e5a88dfb27fd44b9f8db`
	plaintext := "hello world"

	t.Log(`明文`, plaintext)
	ciphertext, err := SM4Encrypt(key, plaintext)
	assert.NoError(t, err)
	t.Log(`密文`, ciphertext)

	// 解密
	decryptedText, err := SM4Decrypt(key, ciphertext)
	assert.NoError(t, err)
	t.Log(`解密结果`, decryptedText)

	// 验证
	valid, err := SM4Verify(key, ciphertext, plaintext)
	assert.NoError(t, err)
	if !valid {
		t.Fatal(`验证失败，解密结果与原文不一致`)
	}
	t.Log(`验证成功`)

	// 使用示例密钥进行测试
	testCiphertext := "056df5b3d1b15e2567d0dcd6e6cfbeff"
	testDecryptedText, err := SM4Decrypt(key, testCiphertext)
	assert.NoError(t, err)
	assert.Equal(t, testDecryptedText, plaintext)
	t.Log(`测试解密结果`, testDecryptedText)
}
