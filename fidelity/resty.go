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

package fidelity

import (
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func RestyFromBrowser(context playwright.BrowserContext) (*resty.Client, error) {
	log.Info().Msg("getting resty client from playwright browser")

	client := resty.New()
	if cookies, err := context.Cookies(); err == nil {
		for _, cookie := range cookies {
			if strings.Contains(cookie.Value, `"`) {
				// skip cookies with invalid characters
				log.Warn().Str("Name", cookie.Name).Str("Domain", cookie.Domain).Str("Path", cookie.Path).Msg("skipping cookie with invalid characters")
				continue
			}
			client.SetCookie(&http.Cookie{
				Name:   cookie.Name,
				Value:  cookie.Value,
				Path:   cookie.Path,
				Domain: cookie.Domain,
			})
		}
	} else {
		log.Error().Err(err).Msg("could not get cookies")
		return nil, err
	}

	if viper.GetBool("log.debug") {
		client.SetDebug(true)
	}

	return client, nil
}
