/*
Copyright 2022

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
package common

import (
	"crypto/aes"
	"encoding/hex"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func EncryptAES(key []byte, plaintext string) string {
	// create cipher
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Error().Err(err).Msg("could not create AES cipher for encryption")
	}

	// allocate space for ciphered data
	out := make([]byte, len(plaintext))

	// encrypt
	c.Encrypt(out, []byte(plaintext))

	// return hex string
	return hex.EncodeToString(out)
}

func DecryptAES(key []byte, ct string) string {
	ciphertext, _ := hex.DecodeString(ct)

	c, err := aes.NewCipher(key)
	if err != nil {
		log.Error().Err(err).Msg("could not create AES cipher for decryption")
	}

	pt := make([]byte, len(ciphertext))
	c.Decrypt(pt, ciphertext)

	s := string(pt[:])
	return s
}

func Username() string {
	usernameCrypted := viper.GetString("username_crypted")
	username := viper.GetString("username")

	if usernameCrypted != "" {
		return DecryptAES(encryptionKey(), usernameCrypted)
	}

	return username
}

func Password() string {
	pinCrypted := viper.GetString("pin_crypted")
	pin := viper.GetString("pin")

	if pinCrypted != "" {
		return DecryptAES(encryptionKey(), pinCrypted)
	}

	return pin
}

func encryptionKey() []byte {
	encryptionKeyPath := viper.GetString("encryption_key")
	key, err := os.ReadFile(encryptionKeyPath)
	if err != nil {
		log.Error().Err(err).Msg("could not read encryption key path")
		return []byte{}
	}

	return key
}
