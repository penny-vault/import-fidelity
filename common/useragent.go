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
	userAgent = strings.ReplaceAll(userAgent, "Headless", "")
	return userAgent
}
