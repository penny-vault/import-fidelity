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

	if err = page.AddInitScript(playwright.Script{
		Content: playwright.String(stealth.JS),
	}); err != nil {
		log.Error().Err(err).Msg("could not load stealth mode")
	}

	return page
}
