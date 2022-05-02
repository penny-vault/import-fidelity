/*
Copyright 2022

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package fidelity

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type Asset struct {
	Name      string `json:"name" parquet:"name=Name, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Ticker    string `json:"ticker" parquet:"name=Ticker, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Exchange  string `json:"exchange" parquet:"name=Exchange, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	AssetType string `json:"asset_type" parquet:"name=AssetType, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Currency  string `json:"currency_iso" parquet:"name=Currency, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	CIK       string `json:"cik" parquet:"name=CIK, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	CUSIP     string `json:"cusip" parquet:"name=CUSIP, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func SaveToParquet(records []*Asset, fn string) error {
	var err error

	fh, err := local.NewLocalFileWriter(fn)
	if err != nil {
		log.Error().Str("OriginalError", err.Error()).Str("FileName", fn).Msg("cannot create local file")
		return err
	}
	defer fh.Close()

	pw, err := writer.NewParquetWriter(fh, new(Asset), 4)
	if err != nil {
		log.Error().
			Str("OriginalError", err.Error()).
			Msg("Parquet write failed")
		return err
	}

	pw.RowGroupSize = 128 * 1024 * 1024 // 128M
	pw.PageSize = 8 * 1024              // 8k
	pw.CompressionType = parquet.CompressionCodec_GZIP

	for _, r := range records {
		if err = pw.Write(r); err != nil {
			log.Error().
				Str("OriginalError", err.Error()).
				Str("Ticker", r.Ticker).
				Msg("Parquet write failed for record")
		}
	}

	if err = pw.WriteStop(); err != nil {
		log.Error().Str("OriginalError", err.Error()).Msg("Parquet write failed")
		return err
	}

	log.Info().Int("NumRecords", len(records)).Msg("Parquet write finished")
	return nil
}

func FetchStockTickerData(symbol string, bearerToken string) *Asset {
	symbol = strings.Replace(symbol, ".", "%2F", -1)

	client := resty.New()
	url := fmt.Sprintf(MARKET_DATA_URL, symbol)
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
	}
	if resp.StatusCode() >= 400 {
		log.Error().
			Int("StatusCode", resp.StatusCode()).
			Str("Url", url).
			Bytes("Body", resp.Body()).
			Msg("invalid status code received")
	}

	body := string(resp.Body())
	asset := &Asset{
		Name:      gjson.Get(body, "data.name").String(),
		Ticker:    gjson.Get(body, "data.symbol").String(),
		Exchange:  gjson.Get(body, "data.exchange.name").String(),
		AssetType: gjson.Get(body, "data.classification.name").String(),
		Currency:  gjson.Get(body, "data.currencyIso").String(),
		CIK:       gjson.Get(body, `data.supplementalData.#(name=="cik").value`).String(),
		CUSIP:     gjson.Get(body, `data.supplementalData.#(name=="cusip").value`).String(),
	}

	return asset
}
