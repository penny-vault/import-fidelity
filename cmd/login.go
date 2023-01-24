// Copyright 2022-2023
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
	"os"

	"github.com/penny-vault/import-fidelity/errorcode"
	"github.com/penny-vault/import-fidelity/fidelity"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Display a browser for login",
	Long: `The login command displays a browser for logging into Fidelity. Once logged in, the session
state is saved to the state-file (--state-file). Fidelity sessions do expire so you will need to login
again after the session has expired. To achieve fully automated control provide username and password
in the configuration and do not use the login command. If you have enabled multi-factor authentication
you will need to use this sub-command for logging in. Check the 'remember device' if you want to
use the automated login on future runs.`,
	Run: func(cmd *cobra.Command, args []string) {
		page, context, browser, pw := fidelity.StartPlaywright(false)

		// load the default homepage
		if _, err := page.Goto(fidelity.HomePageURL, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		}); err != nil {
			log.Error().Err(err).Msg("could not load login page")
		}

		// Wait for the user to press login button
		if _, err := page.WaitForNavigation(); err != nil {
			log.Error().Err(err).Msg("error waiting for navigation on login")
			os.Exit(errorcode.Login)
		}
		page.WaitForTimeout(30000)

		// load the activity page
		if _, err := page.Goto(fidelity.ActivityURL, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		}); err != nil {
			log.Error().Err(err).Msg("could not load activity page")
		}

		fidelity.StopPlaywright(page, context, browser, pw)
	},
}
