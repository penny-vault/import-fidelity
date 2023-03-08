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
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/penny-vault/import-fidelity/common"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

var (
	ErrBadHTTPResponse = errors.New("bad HTTP response")
)

func FetchMutualFundTickerData(asset *common.Asset, page playwright.Page) error {
	assetType := "stock"
	if asset.AssetType == common.MutualFund {
		assetType = "fund"
	}
	url := fmt.Sprintf(CUSIPURL, assetType, asset.Ticker)
	if _, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Error().Err(err).Msg("could not load asset page")
	}

	// name
	selector := "body > table > tbody > tr > td:nth-child(2) > table:nth-child(4) > tbody > tr > td:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(1) > font"
	locator, err := page.Locator(selector)
	if err != nil {
		log.Error().Err(err).Msg("could not create locator for name")
		return err
	}

	cnt, err := locator.Count()
	if err != nil {
		log.Error().Err(err).Msg("could not fetch locator")
		return err
	}

	if cnt == 0 {
		return nil
	}

	// cusip
	selector = "body > table > tbody > tr > td:nth-child(2) > table:nth-child(4) > tbody > tr > td:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(3) > font"
	locator, err = page.Locator(selector)
	if err != nil {
		log.Error().Err(err).Msg("could not create locator")
		return err
	}

	asset.CUSIP, err = locator.InnerText()
	if err != nil {
		log.Error().Err(err).Msg("could not evalute locator for cusip")
		return err
	}
	asset.CUSIP = strings.TrimSpace(asset.CUSIP)
	return nil
}

func FetchStockTickerData(asset *common.Asset, bearerToken string) error {
	symbol := strings.ReplaceAll(asset.Ticker, ".", "%2F")
	symbol = strings.ReplaceAll(symbol, "/", "%2F")

	client := resty.New()
	url := fmt.Sprintf(MarketDataURL, symbol)
	resp, err := client.
		R().
		SetHeader("Accept", "application/json").
		SetHeader("Origin", "https://digital.fidelity.com").
		SetHeader("Referer", "https://digital.fidelity.com/").
		SetHeader("Authorization", bearerToken).
		Get(url)
	if err != nil {
		log.Error().
			Err(err).
			Str("Url", url).
			Msg("http request failed")
		return err
	}
	if resp.StatusCode() >= 400 {
		log.Error().
			Int("StatusCode", resp.StatusCode()).
			Str("Url", url).
			Bytes("Body", resp.Body()).
			Msg("invalid status code received")
		return ErrBadHTTPResponse
	}

	body := string(resp.Body())
	if asset.Name == "" {
		asset.Name = gjson.Get(body, `data.name`).String()
	}

	if asset.AssetType == "" {
		asset.AssetType = gjson.Get(body, `data.classification.name`).String()
	}

	asset.CUSIP = gjson.Get(body, `data.supplementalData.#(name=="cusip").value`).String()
	asset.CIK = gjson.Get(body, `data.supplementalData.#(name=="cik").value`).String()
	return nil
}
