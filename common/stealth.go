package common

import (
	"github.com/go-rod/stealth"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

// StealthPage creates a new playwright page with stealth js loaded to prevent bot detection
func StealthPage(context *playwright.BrowserContext) playwright.Page {
	page, err := (*context).NewPage()
	if err != nil {
		log.Error().Err(err).Msg("could not create page")
	}

	if err = page.AddInitScript(playwright.PageAddInitScriptOptions{
		Script: playwright.String(stealth.JS),
	}); err != nil {
		log.Error().Err(err).Msg("could not load stealth mode")
	}

	return page
}
