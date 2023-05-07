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

package cmd

import (
	"encoding/hex"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/penny-vault/import-fidelity/errorcode"
	"github.com/penny-vault/import-fidelity/fidelity"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

var printTransactions bool

type parquetTransaction struct {
	Account       string
	ID            string
	Commission    float64
	CompositeFIGI string
	Date          string
	Kind          string
	Memo          string
	PricePerShare float64
	Shares        float64
	Source        string
	SourceID      string
	Ticker        string
	TotalValue    float64
}

func init() {
	rootCmd.AddCommand(activityCmd)

	activityCmd.Flags().BoolVar(&printTransactions, "print", true, "print transactions to the screen")
}

var activityCmd = &cobra.Command{
	Use:   "activity",
	Short: "Download account activity",
	Long:  `Retrieves the account activity for the last 10 days`,
	Run: func(cmd *cobra.Command, args []string) {
		page, context, browser, pw := fidelity.StartPlaywright(!viper.GetBool("show_browser"))
		if err := fidelity.Login(page); err != nil {
			fidelity.StopPlaywright(context, browser, pw)
			os.Exit(errorcode.Login)
		}

		client, err := fidelity.RestyFromBrowser(context)
		if err != nil {
			log.Error().Msg("could not get Resty client - exiting.")
			os.Exit(-1)
		}

		accounts, err := fidelity.GetAccounts(client)
		if err != nil {
			log.Error().Msg("error fetching users accounts")
			os.Exit(-1)
		}

		transactions, err := fidelity.AccountActivity(client, accounts)
		if err != nil {
			fidelity.StopPlaywright(context, browser, pw)
			os.Exit(errorcode.Activity)
		}

		if printTransactions {
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Account Number", "Date", "Kind", "Ticker", "Price Per Share", "Shares", "Total", "Memo", "Source ID", "Transaction ID"})
			for acctNum, trxList := range transactions {
				for _, trx := range trxList {
					t.AppendRow(table.Row{
						acctNum,
						trx.Date.Format("2006-01-02"),
						trx.Kind,
						trx.Ticker,
						trx.PricePerShare,
						trx.Shares,
						trx.TotalValue,
						trx.Memo,
						trx.SourceID,
						hex.EncodeToString(trx.ID),
					})
				}
			}
			t.Render()
		}

		// write parquet file
		if viper.GetString("parquet_file") != "" {
			log.Info().Str("fn", viper.GetString("parquet_file")).Msg("save transactions to parquet")
			fh, err := local.NewLocalFileWriter(viper.GetString("parquet_file"))
			if err != nil {
				log.Error().Err(err).Msg("can't create parquet transaction file")
				return
			}

			// parquet schema
			schema := `
						{
						  "Tag": "name=parquet_go_root, repetitiontype=REQUIRED",
						  "Fields": [
							{"Tag": "name=account, inname=Account, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},
							{"Tag": "name=id, inname=ID, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},
							{"Tag": "name=commission, inname=Commission, type=DOUBLE, repetitiontype=REQUIRED"},
							{"Tag": "name=compositeFigi, inname=CompositeFIGI, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},
							{"Tag": "name=date, inname=Date, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},
							{"Tag": "name=kind, inname=Kind, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},
							{"Tag": "name=memo, inname=Memo, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},
							{"Tag": "name=pricePerShare, inname=PricePerShare, type=DOUBLE, repetitiontype=REQUIRED"},
							{"Tag": "name=shares, inname=Shares, type=DOUBLE, repetitiontype=REQUIRED"},
							{"Tag": "name=source, inname=Source, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},
							{"Tag": "name=sourceId, inname=SourceID, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},
							{"Tag": "name=ticker, inname=Ticker, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},
							{"Tag": "name=totalValue, inname=TotalValue, type=DOUBLE, repetitiontype=REQUIRED"}
							]
						}
						`

			parquetWriter, err := writer.NewParquetWriter(fh, schema, 4)
			if err != nil {
				log.Error().Err(err).Msg("can't create parquet writer")
				return
			}

			parquetWriter.RowGroupSize = 128 * 1024 * 1024 // 128M
			parquetWriter.CompressionType = parquet.CompressionCodec_GZIP

			for acctNum, trxList := range transactions {
				for _, trx := range trxList {
					if err = parquetWriter.Write(parquetTransaction{
						Account:       acctNum,
						ID:            hex.EncodeToString(trx.ID),
						Commission:    trx.Commission,
						CompositeFIGI: trx.CompositeFIGI,
						Date:          trx.Date.Format("2006-01-02"),
						Kind:          trx.Kind,
						Memo:          trx.Memo,
						PricePerShare: trx.PricePerShare,
						Shares:        trx.Shares,
						Source:        trx.Source,
						SourceID:      trx.SourceID,
						Ticker:        trx.Ticker,
						TotalValue:    trx.TotalValue,
					}); err != nil {
						log.Error().Err(err).Msg("error writing transaction to parquet")
					}
				}
			}

			if err = parquetWriter.WriteStop(); err != nil {
				log.Error().Err(err).Msg("WriteStop error")
			}

			fh.Close()
		}
	},
}
