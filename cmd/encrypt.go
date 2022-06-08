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

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/penny-vault/import-fidelity/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	rootCmd.AddCommand(encryptCmd)
}

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}

	password := string(bytePassword)
	fmt.Println()
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt username and password suitable for use in the configuration file.",
	Long:  `This command prompts for the fidelity username and password and ecrypts them for use in the configuration file`,
	Run: func(cmd *cobra.Command, args []string) {
		username, password, err := credentials()
		if err != nil {
			log.Error().Err(err).Msg("could not get credentials")
			return
		}

		cryptedUsername := common.EncryptAES(username)
		cryptedPassword := common.EncryptAES(password)
		fmt.Printf("Username: %s\n", cryptedUsername)
		fmt.Printf("Password: %s\n", cryptedPassword)
	},
}
