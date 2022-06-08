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

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/penny-vault/import-fidelity/backblaze"
	"github.com/penny-vault/import-fidelity/common"
	"github.com/penny-vault/import-fidelity/errorcode"
	"github.com/penny-vault/import-fidelity/fidelity"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/ratelimit"
)

var downloadFromBackblaze bool
var uploadToBackblaze bool

func init() {
	rootCmd.AddCommand(cusipCmd)

	cusipCmd.Flags().BoolVarP(&downloadFromBackblaze, "download-from-backblaze", "d", false, "Download ticker database from backblaze")
	cusipCmd.Flags().BoolVarP(&uploadToBackblaze, "upload-to-backblaze", "s", false, "Upload ticker database to backblaze")
}

func getBearerToken(page playwright.Page) string {
	log.Info().Msg("waiting for market data request")
	symbol := "MSFT"
	quoteURL := fmt.Sprintf(fidelity.QuoteURL, symbol)
	url := fmt.Sprintf(fidelity.MarketDataURL, symbol)
	req, err := page.ExpectRequest(url, func() error {
		_, err := page.Goto(quoteURL)
		return err
	})
	if err != nil {
		log.Error().Err(err).Msg("error waiting for market data request")
		return ""
	}

	headers, err := req.AllHeaders()
	if err != nil {
		log.Error().Err(err).Msg("error fetching request headers")
		return ""
	}

	var bearerToken string
	var ok bool
	if bearerToken, ok = headers["authorization"]; !ok {
		log.Error().Err(err).Msg("error fetching bearer token")
		return ""
	}

	log.Debug().Str("BearerToken", bearerToken).Msg("authorization")
	return bearerToken
}

var cusipCmd = &cobra.Command{
	Use:   "cusip [symbols...]",
	Short: "Download CUSIP from Fidelity using the quotes webpage",
	Long: `Downloads CUSIP for each symbol listed in arguments. If no arguments
provided load tickers from backblaze and use assets that have no CUSIP.
To search for mutual funds use the :MF suffix, e.g. to find data for VFIAX use VFIAX:MF`,
	Run: func(cmd *cobra.Command, args []string) {
		assets := []*common.Asset{}

		limit := ratelimit.New(1)

		log.Info().Bool("Download", downloadFromBackblaze).Bool("Upload", uploadToBackblaze).Msg("backblaze flags")

		// check if arguments should be read from file
		switch len(args) {
		case 0:
			if downloadFromBackblaze {
				if err := backblaze.Download(viper.GetString("parquet_file"), viper.GetString("backblaze.bucket")); err != nil {
					os.Exit(errorcode.Backblaze)
				}
			}
			assets = common.ReadFromParquet(viper.GetString("parquet_file"))
		case 1:
			arg0 := args[0]
			if arg0[0] == '@' {
				raw, err := os.ReadFile(arg0[1:])
				if err != nil {
					log.Error().Err(err).Str("FileName", arg0).Msg("cannot read argument file")
				}
				tmpArgs := strings.Split(string(raw), "\n")
				for _, arg := range tmpArgs {
					if arg != "" {
						asset := &common.Asset{
							Ticker: arg,
						}
						assets = append(assets, asset)
					}
				}
			}
		default:
			for _, arg := range args {
				if arg != "" {
					asset := &common.Asset{
						Ticker: arg,
					}
					assets = append(assets, asset)
				}
			}
		}

		noCusip := make([]*common.Asset, 0, len(assets))
		for _, asset := range assets {
			if asset.CUSIP == "" && !asset.FidelityCusip {
				noCusip = append(noCusip, asset)
			}
		}

		// start playwright
		page, context, browser, pw := fidelity.StartPlaywright(!viper.GetBool("show_browser"))
		if err := fidelity.Login(page); err != nil {
			fidelity.StopPlaywright(page, context, browser, pw)
			os.Exit(errorcode.Login)
		}

		bearerToken := getBearerToken(page)

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)

		t.AppendHeader(table.Row{"Name", "Ticker", "Asset Type", "CUSIP"})
		bar := progressbar.Default(int64(len(noCusip)))
		for _, asset := range noCusip {
			limit.Take()
			if err := bar.Add(1); err != nil {
				log.Warn().Err(err).Msg("could not add to progress bar")
			}
			asset.FidelityCusip = true
			var err error
			if asset.AssetType == common.MutualFund {
				err = fidelity.FetchTickerData(asset, page)
			} else {
				err = fidelity.FetchStockTickerData(asset, bearerToken)
			}
			if err != nil {
				t.AppendRow(table.Row{asset.Name, asset.Ticker, asset.AssetType, asset.CUSIP})
			}
		}

		t.AppendFooter(table.Row{"", "", "Total", len(assets)})
		t.Render()

		if viper.GetString("parquet_file") != "" {
			if err := common.SaveToParquet(assets, viper.GetString("parquet_file")); err != nil {
				os.Exit(errorcode.WriteParquet)
			}
		}

		fidelity.StopPlaywright(page, context, browser, pw)

		if uploadToBackblaze {
			if err := backblaze.Upload(viper.GetString("parquet_file"), viper.GetString("backblaze.bucket"), "."); err != nil {
				os.Exit(errorcode.Backblaze)
			}
		}
	},
}
