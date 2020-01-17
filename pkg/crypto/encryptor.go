package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

//NewAes Encoder returns a AesEncryptor
func NewAesEncryptor() *AesEncryptor {
	return &AesEncryptor{}
}

// NewRsaEncryptor uses the key pair to build a rsa encoder
func NewRsaEncryptor() *RsaEncryptor {
	return &RsaEncryptor{}
}

type AesEncryptor struct{}

func (*AesEncryptor) Encrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	cryptedData := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(cryptedData, origData)
	return cryptedData, nil
}

func (*AesEncryptor) Decrypt(cryptedData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(cryptedData))
	// origData := crypted
	blockMode.CryptBlocks(origData, cryptedData)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}

func ZeroPadding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{0}, padding)
	return append(cipherText, padText...)
}

func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unPadding 次
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

type RsaEncryptor struct{}

// 加密
func (re *RsaEncryptor) Encrypt(origData, publicKey []byte) ([]byte, error) {
	//解密pem格式的公钥
	pubBlock, _ := pem.Decode(publicKey)
	if pubBlock == nil {
		return nil, errors.New("public key error")
	}
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, err
	}
	// 类型断言
	pubKey := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pubKey, origData)
}

// 解密
func (re *RsaEncryptor) Decrypt(cryptedData, privateKey []byte) ([]byte, error) {
	//解密
	priBlock, _ := pem.Decode(privateKey)
	if priBlock == nil {
		return nil, errors.New("private key error!")
	}
	//解析PKCS1格式的私钥
	priKey, err := x509.ParsePKCS1PrivateKey(priBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priKey, cryptedData)
}
