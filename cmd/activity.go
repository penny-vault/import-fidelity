// Copyright 2021
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
	"github.com/penny-vault/import-fidelity/fidelity"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(activityCmd)

	activityCmd.Flags().BoolP("show-browser", "d", true, "don't run the browser in headless mode")
	viper.BindPFlag("show_browser", activityCmd.Flags().Lookup("show-browser"))
}

var activityCmd = &cobra.Command{
	Use:   "activity",
	Short: "Download account activity",
	Long:  `Retrieves the account activity for the last 10 days`,
	Run: func(cmd *cobra.Command, args []string) {

		page, context, browser, pw := fidelity.StartPlaywright(false)

		// load the activity page
		if _, err := page.Goto(fidelity.ACTIVITY_URL, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		}); err != nil {
			log.Error().Err(err).Msg("could not load activity page")
		}

		req := page.WaitForRequest(fidelity.ACTIVITY_API_URL)

		fidelity.StopPlaywright(page, context, browser, pw)
	},
}
