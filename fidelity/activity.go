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
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/penny-vault/pvlib"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

type TransactionDetails struct {
	AccountNumber string `json:"acctNum"`
	AccountName   string `json:"acctName"`
	AccountType   string `json:"acctType"`

	IsChecks     bool `json:"isChecks"`
	IsCitAccount bool `json:"isCITAccount"`
	IsDeposit    bool `json:"isDeposit"`
	IsMsla       bool `json:"isMsla"`
	IsMutualFund bool `json:"isMutualFund"`

	PostedDate  int    `json:"postedDate"`
	Date        string `json:"date"`
	OrderNumber string `json:"orderNumber"`
	Description string `json:"autoTxnDesc"`
	MtTitle     string `json:"mtTitle"`

	AmountDetail struct {
		Price      string  `json:"price"`
		Shares     float64 `json:"shares"`
		Fee        float64 `json:"fee"`
		Commission float64 `json:"commission"`
		Net        float64 `json:"net"`
		Interest   float64 `json:"interest"`
	} `json:"amtDetail"`

	BrokerageDetails struct {
		AccountType          string `json:"brokerageAccountType"`
		TransactionAttribute struct {
			TransactionID string `json:"txnTypNum"`
		} `json:"txnAttribute"`
		DateDetail struct {
			IsValidSettlementDate bool   `json:"isValidSettlementDate"`
			SettlementDate        string `json:"settlementDate"`
		} `json:"dateDetail"`
		SecurityDetail struct {
			SecurityDescription           string `json:"securityDesc"`
			MobileSecurityDescription     string `json:"mobileSecurityDesc"`
			QuotableSecurity              bool   `json:"quotableSecurityInd"`
			SecurityType                  string `json:"secType"`
			AssetClass                    string `json:"assetClass"`
			OSISymbolType                 string `json:"OSISymbolType"`
			CUSIP                         string `json:"67066G104"`
			CollateralIndicator           bool   `json:"collateralIndicator"`
			MMRIndication                 bool   `json:"mmrIndicator"`
			Symbol                        string `json:"symbol"`
			FloorTradingSymbol            string `json:"floorTradingSymbol"`
			FloorTradingSymbolDescription string `json:"floorTradingSymbolDesc"`
			QuoteText                     string `json:"quoteText"`
			SecurityId                    string `json:"securityId"`
		} `json:"securityDetail"`
	} `json:"brokerageDetail"`

	IntradayIndicator            bool   `json:"intradayInd"`
	MultiCurrencyTransactionType string `json:"multiCurrencyTransactionType"`
	HasChecks                    bool   `json:"hasChecks"`
	HasImages                    bool   `json:"hasImages"`
	Amount                       string `json:"amount"`
	TransactionDescription       string `json:"txnDescription"`
}

func AccountActivity(page playwright.Page) (map[string][]*pvlib.Transaction, error) {
	subLog := log.With().Str("Url", ACTIVITY_URL).Logger()
	// load the activity page
	req, err := page.ExpectRequest(ACTIVITY_API_URL, func() error {
		_, err := page.Goto(ACTIVITY_URL)
		return err
	})
	if err != nil {
		subLog.Error().Err(err).Msg("could not load activity page")
		return nil, err
	}

	resp, err := req.Response()
	if err != nil {
		subLog.Error().Err(err).Msg("error while waiting for response to activity api")
		return nil, err
	}

	body, err := resp.Body()
	if err != nil {
		subLog.Error().Err(err).Msg("error while fetching body")
		return nil, err
	}
	bodyStr := string(body)
	nyc, _ := time.LoadLocation("America/New_York")

	numTransactions := gjson.Get(bodyStr, "transaction.txnDetails.txnDetail.#").Int()
	log.Debug().Int64("NumTransactions", numTransactions).Str("json", bodyStr).Msg("downloaded transactions")
	trxMap := map[string][]*pvlib.Transaction{}
	result := gjson.Get(bodyStr, "transaction.txnDetails.txnDetail")
	result.ForEach(func(key, value gjson.Result) bool {
		// skip intraday activity
		if value.Get("intradayInd").Bool() {
			return true
		}

		id := uuid.New()
		idBinary, err := id.MarshalBinary()
		if err != nil {
			subLog.Error().Err(err).Msg("could not marshal UUID to binary")
		}

		date, err := time.Parse("01/02/2006", value.Get("date").String())
		if err != nil {
			log.Error().Err(err).Str("DateValue", value.Get("date").String()).Msg("could not parse transaction date")
			return true
		}

		date = time.Date(date.Year(), date.Month(), date.Day(), 16, 0, 0, 0, nyc)

		pricePerShare, err := strconv.ParseFloat(value.Get("amtDetail.price").String(), 64)
		if err != nil {
			log.Error().Err(err).Str("PricePerShare", value.Get("amtDetail.price").String()).Msg("could not parse transaction price to float")
			return true
		}

		trx := pvlib.Transaction{
			ID:            idBinary,
			Commission:    value.Get("amtDetail.commission").Float() + value.Get("amtDetail.fee").Float(),
			Date:          date,
			Memo:          value.Get("txnDescription").String(),
			PricePerShare: pricePerShare,
			Shares:        value.Get("amtDetail.shares").Float(),
			Source:        "fidelity.com",
			SourceID:      value.Get("orderNumber").String(),
			Ticker:        value.Get("brokerageDetail.securityDetail.symbol").String(),
			TotalValue:    value.Get("amtDetail.net").Float(),
		}

		acctNum := value.Get("acctNum").String()

		// determine kind
		trx.Kind = determineKind(trx.Shares, trx.TotalValue, trx.Ticker, value.Get("brokerageDetail.brokerageAccountType").String())

		if trx.Kind == pvlib.WithdrawTransaction && trx.Ticker == "FCASH" {
			// This is an investment in the FCASH holding which is effectively a cash investment.
			// ignore the transaction
			return true
		}

		if trx.Kind == pvlib.DepositTransaction || trx.Kind == pvlib.WithdrawTransaction {
			trx.Ticker = "CASH"
			trx.PricePerShare = 1.0
			trx.Shares = trx.TotalValue
		}

		if trxList, ok := trxMap[acctNum]; !ok {
			trxList := make([]*pvlib.Transaction, 0, numTransactions)
			trxList = append(trxList, &trx)
			trxMap[acctNum] = trxList
		} else {
			trxList = append(trxList, &trx)
			trxMap[acctNum] = trxList
		}

		return true
	})

	return trxMap, nil
}

func determineKind(shares float64, amount float64, ticker string, brokerageAccountType string) string {
	if brokerageAccountType == "Cash" {
		if amount >= 0 {
			return pvlib.DepositTransaction
		} else {
			return pvlib.WithdrawTransaction
		}
	}

	if shares > 0 {
		return pvlib.BuyTransaction
	}

	if shares < 0 {
		return pvlib.SellTransaction
	}

	if shares == 0 && ticker != "" {
		return pvlib.DividendTransaction
	}

	log.Error().Float64("Shares", shares).Float64("Amount", amount).Str("Ticker", ticker).Str("BrokerageAccountType", brokerageAccountType).Msg("could not determine transaction type")
	return ""
}
