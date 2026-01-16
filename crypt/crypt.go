// Copyright (c) 2026 Reiner Pröls
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// SPDX-License-Identifier: MIT
//
// Author: Reiner Pröls

package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/scrypt"
)

// Parameter
const (
	saltSize       = 16
	nonceSize      = 12
	keySize        = 32
	InternPassword = "ti6xGckIznYZwihc6vp5I8gI"
)

// EncryptPassword verschlüsselt ein Passwort mit einem Masterpasswort
func Encrypt(masterPassword, password string) (string, error) {
	// Salt generieren
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// Key ableiten
	key, err := scrypt.Key([]byte(masterPassword), salt, 1<<15, 8, 1, keySize)
	if err != nil {
		return "", err
	}

	// AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Nonce generieren
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Verschlüsseln
	ciphertext := gcm.Seal(nil, nonce, []byte(password), nil)

	// Alles zusammenfügen: Salt + Nonce + Ciphertext
	data := append(salt, nonce...)
	data = append(data, ciphertext...)

	// Base64 kodieren
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

// DecryptPassword entschlüsselt den Base64-String wieder
func Decrypt(masterPassword, encoded string) (string, error) {
	if len(encoded) == 0 {
		return "", errors.New("length 0")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	// Salt, Nonce, Ciphertext extrahieren
	if (len(data)) < saltSize+nonceSize {
		return "", errors.New("data length too short")
	}
	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	// Key ableiten
	key, err := scrypt.Key([]byte(masterPassword), salt, 1<<15, 8, 1, keySize)
	if err != nil {
		return "", err
	}

	// AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Entschlüsseln
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
