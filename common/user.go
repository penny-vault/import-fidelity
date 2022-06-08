// Copyright 2022
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"lukechampine.com/blake3"
)

func EncryptAES(plaintext string) string {
	// create cipher
	key := encryptionKey()
	if len(key) != 32 {
		log.Error().Msg("encryption key invalid length")
		return ""
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Error().Err(err).Msg("could not create AES cipher for encryption")
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	// if any error generating new GCM
	// handle them
	if err != nil {
		log.Error().Err(err).Msg("error creating gcm")
		return ""
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Error().Err(err).Msg("error populating nonce")
		return ""
	}

	// encrypt
	out := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// return hex string
	return base64.StdEncoding.EncodeToString(out)
}

func DecryptAES(ct string) string {
	ciphertext, _ := base64.StdEncoding.DecodeString(ct)
	key := encryptionKey()
	if len(key) != 32 {
		log.Error().Msg("encryption key invalid length")
		return ""
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Error().Err(err).Msg("could not create AES cipher for decryption")
		return ""
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		log.Error().Err(err).Msg("could not create gcm")
		return ""
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		log.Error().Int("CipherTextSize", len(ciphertext)).Int("NonceSize", nonceSize).Msg("encrypted text is smaller than the nonce")
		return ""
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Error().Err(err).Msg("unable to decrypt text")
	}

	return string(plaintext)
}

func Username() string {
	username := viper.GetString("username")
	return DecryptAES(username)
}

func Password() string {
	pin := viper.GetString("pin")
	return DecryptAES(pin)
}

func encryptionKey() []byte {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error().Err(err).Msg("cannot get home directory")
		return []byte{}
	}
	encryptionKeyPath := fmt.Sprintf("%s/.ssh/id_rsa", userHomeDir)
	key, err := os.ReadFile(encryptionKeyPath)
	if err != nil {
		log.Error().Err(err).Str("EncryptionKey", encryptionKeyPath).Msg("could not read encryption key")
		return []byte{}
	}

	key32 := blake3.Sum256(key)
	return key32[:]
}
