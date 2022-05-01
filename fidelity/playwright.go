package fidelity

import (
	"encoding/json"
	"os"

	"github.com/penny-vault/import-fidelity/common"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func StartPlaywright(headless bool) (page playwright.Page, context playwright.BrowserContext, browser playwright.Browser, pw *playwright.Playwright) {
	pw, err := playwright.Run()
	if err != nil {
		log.Error().Err(err).Msg("could not launch playwright")
	}

	log.Info().Str("ExecutablePath", pw.Chromium.ExecutablePath()).Msg("chromium install")

	browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		log.Error().Err(err).Msg("could not launch Chromium")
	}

	log.Info().Str("BrowserVersion", browser.Version()).Msg("browser info")

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
	stateFileName := viper.GetString("state_file")
	storage, err := context.StorageState(stateFileName)
	if err != nil {
		log.Error().Err(err).Msg("could not get storage state")
	}
	log.Info().Int("NumCookies", len(storage.Cookies)).Msg("session state")

	if err = context.Close(); err != nil {
		log.Error().Err(err).Msg("error encountered when closing context")
	}

	if err = browser.Close(); err != nil {
		log.Error().Err(err).Msg("error encountered when closing browser")
	}

	if err = pw.Stop(); err != nil {
		log.Error().Err(err).Msg("error encountered when stopping playwright")
	}
}
