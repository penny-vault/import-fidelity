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

package fidelity

import (
	"encoding/json"
	"os"

	"github.com/penny-vault/import-fidelity/common"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// check if the current session is expired; if it is login
func Login(page playwright.Page) error {
	subLog := log.With().Str("Url", SummaryURL).Logger()

	if _, err := page.Goto(SummaryURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		subLog.Error().Err(err).Msg("could not load activity page")
		return err
	}

	locator, err := page.Locator("#userId-input")
	if err != nil {
		subLog.Error().Err(err).Msg("error acquiring locator for userId")
	}

	cnt, err := locator.Count()
	if err != nil {
		subLog.Error().Err(err).Msg("error evaluating locator count")
		return err
	}
	if cnt > 0 {
		// session expired login
		log.Info().Msg("session expired; login required")
		if err = page.Type("#userId-input", common.Username()); err != nil {
			log.Error().Err(err).Msg("could not type in username input box")
			return err
		}

		if err = page.Type("#password", common.Password()); err != nil {
			log.Error().Err(err).Msg("could not type in password input box")
			return err
		}
		if err = page.Click("#fs-login-button"); err != nil {
			log.Error().Err(err).Msg("could not click login button")
			return err
		}
		log.Debug().Msg("waiting for 10 seconds")
		page.WaitForTimeout(10000)
	} else {
		log.Info().Msg("session is active; no login necessary")
	}

	return nil
}

// StartPlaywright starts the playwright server and browser, it then creates a new context and page with the stealth extensions loaded
func StartPlaywright(headless bool) (page playwright.Page, context playwright.BrowserContext, browser playwright.Browser, pw *playwright.Playwright) {
	pw, err := playwright.Run()
	if err != nil {
		log.Error().Err(err).Msg("could not launch playwright")
	}

	browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		log.Error().Err(err).Msg("could not launch Chromium")
	}

	log.Info().Bool("Headless", headless).Str("ExecutablePath", pw.Chromium.ExecutablePath()).Str("BrowserVersion", browser.Version()).Msg("starting playwright")

	// load browser state
	stateFileName := viper.GetString("state_file")
	log.Info().Str("StateFile", stateFileName).Msg("state location")
	var storageState playwright.BrowserNewContextOptionsStorageState
	data, err := os.ReadFile(stateFileName)
	if err != nil {
		log.Error().Err(err)
	}
	err = json.Unmarshal(data, &storageState)
	if err != nil {
		log.Error().Err(err)
	}

	// calculate user-agent
	userAgent := viper.GetString("user_agent")
	if userAgent == "" {
		userAgent = common.BuildUserAgent(&browser)
	}
	log.Info().Str("UserAgent", userAgent).Msg("using user-agent")

	// create context
	context, err = browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent:    playwright.String(userAgent),
		StorageState: &storageState,
	})
	if err != nil {
		log.Printf("could not create browser context")
	}

	// get a page
	page = common.StealthPage(&context)

	return
}

func StopPlaywright(page playwright.Page, context playwright.BrowserContext, browser playwright.Browser, pw *playwright.Playwright) {
	// save session state
	log.Info().Msg("saving state")
	stateFileName := viper.GetString("state_file")
	storage, err := context.StorageState(stateFileName)
	if err != nil {
		log.Error().Err(err).Msg("could not get storage state")
	}
	log.Info().Int("NumCookies", len(storage.Cookies)).Msg("session state")

	log.Info().Msg("closing context")
	if err = context.Close(); err != nil {
		log.Error().Err(err).Msg("error encountered when closing context")
	}

	log.Info().Msg("closing browser")
	if err = browser.Close(); err != nil {
		log.Error().Err(err).Msg("error encountered when closing browser")
	}

	log.Info().Msg("stopping playwright")
	if err = pw.Stop(); err != nil {
		log.Error().Err(err).Msg("error encountered when stopping playwright")
	}
}
