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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/penny-vault/import-fidelity/fidelity"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/ratelimit"
)

func init() {
	rootCmd.AddCommand(tickersCmd)

	// Local flags
	tickersCmd.Flags().Int("rate-limit", 5, "rate limit (items per second)")
	viper.BindPFlag("rate_limit", tickersCmd.Flags().Lookup("rate-limit"))

	tickersCmd.Flags().String("parquet-file", "", "save results to parquet")
	viper.BindPFlag("parquet_file", tickersCmd.Flags().Lookup("parquet-file"))
}

var tickersCmd = &cobra.Command{
	Use:   "tickers [symbols...]",
	Short: "Download information about assets traded on Fidelity",
	Long: `Downloads Stock type, currency, exchange, symbol, name,
CUSIP, and CIK for each symbol listed in arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		assets := []*fidelity.Asset{}

		log.Info().Int("RateLimit", viper.GetInt("rate_limit")).Msg("setting rate limit")
		limit := ratelimit.New(viper.GetInt("rate_limit"))

		// check if arguments should be read from file
		if len(args) == 1 {
			arg0 := args[0]
			if arg0[0] == '@' {
				raw, err := os.ReadFile(arg0[1:])
				if err != nil {
					log.Error().Err(err).Str("FileName", arg0).Msg("cannot read argument file")
				}
				tmpArgs := strings.Split(string(raw), "\n")
				args = []string{}
				for _, arg := range tmpArgs {
					if arg != "" {
						args = append(args, arg)
					}
				}
			}
		}

		// get bearer token
		page, context, browser, pw := fidelity.StartPlaywright(!viper.GetBool("show_browser"))
		fidelity.Login(page)

		// load the default homepage
		log.Info().Msg("waiting for market data request")
		symbol := "MSFT"
		quoteUrl := fmt.Sprintf(fidelity.QUOTE_URL, symbol)
		url := fmt.Sprintf(fidelity.MARKET_DATA_URL, symbol)
		req, err := page.ExpectRequest(url, func() error {
			_, err := page.Goto(quoteUrl)
			return err
		})
		if err != nil {
			log.Error().Err(err).Msg("error waiting for market data request")
			return
		}

		headers, err := req.AllHeaders()
		if err != nil {
			log.Error().Err(err).Msg("error fetching request headers")
			return
		}

		var bearerToken string
		var ok bool
		if bearerToken, ok = headers["authorization"]; !ok {
			log.Error().Err(err).Msg("error fetching bearer token")
			return
		}

		log.Debug().Str("BearerToken", bearerToken).Msg("authorization")

		assetChan := make(chan *fidelity.Asset, len(args))

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)

		t.AppendHeader(table.Row{"Name", "Ticker", "Exchange", "AssetType", "Currency", "CIK", "CUSIP"})
		bar := progressbar.Default(int64(len(args)))
		for _, symbol := range args {
			limit.Take()
			bar.Add(1)
			go func(mySymbol string) {
				asset := fidelity.FetchStockTickerData(mySymbol, bearerToken)
				if asset.Ticker != "" {
					assetChan <- asset
				}
			}(symbol)
		}

		for xx := 0; xx < 5; xx++ {
			if len(assetChan) < 1 {
				log.Info().Msg("waiting for results ...")
				time.Sleep(time.Second * 1)
			}
		}

		for xx := 0; xx < len(args); xx++ {
			if len(assetChan) < 1 {
				break
			}
			asset := <-assetChan
			assets = append(assets, asset)
			t.AppendRow(table.Row{asset.Name, asset.Ticker, asset.Exchange, asset.AssetType, asset.Currency, asset.CIK, asset.CUSIP})
		}

		t.AppendFooter(table.Row{"", "", "", "", "", "Total", len(assets)})
		t.Render()

		if viper.GetString("parquet_file") != "" {
			fidelity.SaveToParquet(assets, viper.GetString("parquet_file"))
		}

		fidelity.StopPlaywright(page, context, browser, pw)
	},
}
