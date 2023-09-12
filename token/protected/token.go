/*
 Copyright 2023 Qiniu Limited (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package protected

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"io/fs"
	"net/url"
	"os"
)

var (
	KeySalt    string
	EnvKeyName string
)

// -----------------------------------------------------------------------------------------

// Decode decodes a protected token.
func Decode(token string) (_ url.Values, err error) {
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fs.ErrPermission
	}
	orig, err := decodeData(b)
	if err != nil {
		return
	}
	return url.ParseQuery(string(orig))
}

// Encode encodes a protected token from params.
func Encode(params url.Values) (token string, err error) {
	b := []byte(params.Encode())
	crypted, err := encodeData(b)
	if err != nil {
		return
	}
	return base64.RawURLEncoding.EncodeToString(crypted), nil
}

func decodeData(crypted []byte) (_ []byte, err error) {
	key := os.Getenv(EnvKeyName)
	if key == "" {
		return nil, fs.ErrPermission
	}
	key2 := sha256.Sum256([]byte(KeySalt + key))
	block, err := aes.NewCipher(key2[:])
	if err != nil {
		return nil, fs.ErrPermission
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key2[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	return unpadding(origData), nil
}

func encodeData(origData []byte) (_ []byte, err error) {
	key := os.Getenv(EnvKeyName)
	if key == "" {
		return nil, fs.ErrPermission
	}
	key2 := sha256.Sum256([]byte(KeySalt + key))
	block, err := aes.NewCipher(key2[:])
	if err != nil {
		return nil, fs.ErrPermission
	}
	blockSize := block.BlockSize()
	origData = padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key2[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func unpadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// -----------------------------------------------------------------------------------------
