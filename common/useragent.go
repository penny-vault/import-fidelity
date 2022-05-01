package common

import (
	"strings"

	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func BuildUserAgent(browser *playwright.Browser) string {
	context, err := (*browser).NewContext()
	if err != nil {
		log.Error().Err(err).Msg("could not create context for building user agent")
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		log.Error().Err(err).Msg("could not create page BuildUserAgent")
	}

	resp, err := page.Goto("https://playwright.dev", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		log.Error().Err(err).Str("Url", "https://playwright.dev").Msg("could not load page")
	}

	headers, err := resp.Request().AllHeaders()
	if err != nil {
		log.Error().Err(err).Msg("could not load request headers")
	}

	userAgent := headers["user-agent"]
	userAgent = strings.Replace(userAgent, "Headless", "", -1)
	return userAgent
}
