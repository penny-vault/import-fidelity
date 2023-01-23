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

package fidelity

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/penny-vault/pvlib"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

var (
	ErrInvalidResponseCode = errors.New("invalid status code returned from activity graph QL")
)

func isCoreHolding(ticker string) bool {
	return ticker == "FCASH" || ticker == "SPAXX" || ticker == "FZFXX"
}

func determineTransactionKind(trx pvlib.Transaction, trxType, trxCategory, trxSubCategory string) string {
	switch trxType {
	case "CT":
		switch trxCategory {
		case "DV":
			switch trxSubCategory {
			case "VP":
				return pvlib.DividendTransaction
			}
		case "IA":
			switch trxSubCategory {
			case "OC":
				if trx.TotalValue <= 0 {
					return pvlib.WithdrawTransaction
				} else {
					return pvlib.DepositTransaction
				}
			}
		case "X2":
			switch trxSubCategory {
			case "DP":
				return pvlib.DepositTransaction
			}
		case "X1":
			switch trxSubCategory {
			case "OC":
				return pvlib.WithdrawTransaction
			}
		}
	case "IT":
		switch trxCategory {
		case "DV":
			switch trxSubCategory {
			case "VP":
				return pvlib.DividendTransaction
			case "IT":
				return pvlib.InterestTransaction
			}
		case "IA":
			switch trxSubCategory {
			case "VP":
				return pvlib.DividendTransaction
			}
		}
	case "ST":
		switch trxCategory {
		case "IA":
			switch trxSubCategory {
			case "BY":
				return pvlib.BuyTransaction
			case "SL":
				return pvlib.SellTransaction
			}
		case "DV":
			switch trxSubCategory {
			case "RN": // Re-invest
				return pvlib.BuyTransaction
			}
		case "ZZ":
			switch trxSubCategory {
			case "BY":
				return pvlib.BuyTransaction
			case "SL":
				return pvlib.SellTransaction
			}
		}
	}

	log.Warn().Str("txnTypeCode", trxType).Str("txnCategory", trxCategory).Str("txnSubCategory", trxSubCategory).Object("Transaction", &trx).Msg("could not determine transaction type")
	return ""
}

func AccountActivity(client *resty.Client, accounts []*Account) (map[string][]*pvlib.Transaction, error) {
	idList := make([]string, len(accounts))
	for idx, account := range accounts {
		idList[idx] = account.AccountNumber
	}
	toDate := time.Now()
	fromDate := toDate.Add(86400 * time.Second * -90)
	gqlQuery := GraphQLQuery{
		OperationName: "getTransactions",
		Variables: map[string]any{
			"isNewOrderApi":   false,
			"isSupportCrypto": false,
			"acctIdList":      strings.Join(idList, ","),
			"acctDetailList":  accounts,
			"searchCriteriaDetail": map[string]any{
				"txnFromDate":   fromDate.Format("01/02/2006"),
				"txnToDate":     toDate.Format("01/02/2006"),
				"timePeriod":    90,
				"txnCat":        nil,
				"viewType":      "NON_CORE",
				"acctHistDays":  "Past 90 Days",
				"histSortDir":   "D",
				"acctHistSort":  "DATE",
				"hasBasketName": true,
			},
		},
		Query: GQLGetTransactions,
	}

	bodyStr := ""
	if resp, err := client.R().
		SetBody(gqlQuery).
		Post(GraphQLURL); err == nil {
		if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
			// it worked!
			bodyStr = resp.String()
		} else {
			// invalid status code
			log.Error().Int("StatusCode", resp.StatusCode()).Str("Status", resp.Status()).Msg("invalid status code received")
			return nil, ErrInvalidResponseCode
		}
	} else {
		log.Error().Err(err).Msg("request failed")
		return nil, err
	}

	trxMap, err := ParseAccountActivity(bodyStr)
	if err != nil {
		return nil, err
	}

	return trxMap, nil
}

func getDetailItemNumber(value gjson.Result, key string) float64 {
	var err error
	retVal := 0.0
	detailValue := value.Get(fmt.Sprintf(`detailItems.#(key=="%s").value`, key))
	if detailValue.Exists() {
		strVal := detailValue.String()
		strVal = strings.ReplaceAll(strVal, " ", "")
		strVal = strings.ReplaceAll(strVal, "$", "")
		strVal = strings.ReplaceAll(strVal, ",", "")
		retVal, err = strconv.ParseFloat(strVal, 64)
		if err != nil {
			log.Error().Err(err).Str("key", key).Str(key, detailValue.String()).Msg("could not parse float value from detailItems")
			return 0.0
		}
	}
	return retVal
}

func getDollarValue(value gjson.Result, key string) float64 {
	val := value.Get(key).String()
	val = strings.ReplaceAll(val, "$", "")
	val = strings.ReplaceAll(val, ",", "")
	retVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Error().Err(err).Str("key", key).Str("val", val).Msg("could not convert dollar value to a float")
		return 0.0
	}
	return retVal
}

// ParseAccountActivity reads a json string with account activity downloaded from Fidelity
func ParseAccountActivity(fidelityActivityJSON string) (trxMap map[string][]*pvlib.Transaction, err error) {
	log.Info().Msg("loading account activity")
	nyc, _ := time.LoadLocation("America/New_York")
	trxMap = make(map[string][]*pvlib.Transaction, 1)
	numTransactions := gjson.Get(fidelityActivityJSON, "data.getTransactions.historys.#").Int()
	log.Debug().Int64("NumTransactions", numTransactions).Msg("downloaded transactions")
	result := gjson.Get(fidelityActivityJSON, "data.getTransactions.historys")
	result.ForEach(func(key, value gjson.Result) bool {
		id := uuid.New()
		idBinary, err := id.MarshalBinary()
		if err != nil {
			log.Error().Err(err).Msg("could not marshal UUID to binary")
			return false
		}

		date, err := time.Parse("02 Jan 2006", value.Get("date").String())
		if err != nil {
			log.Error().Err(err).Str("DateValue", value.Get("date").String()).Msg("could not parse transaction date")
			return true
		}

		date = time.Date(date.Year(), date.Month(), date.Day(), 16, 0, 0, 0, nyc)

		trx := pvlib.Transaction{
			ID:            idBinary,
			Commission:    math.Abs(getDetailItemNumber(value, "Commission")) + math.Abs(getDetailItemNumber(value, "Fees")),
			Date:          date,
			Memo:          value.Get("description").String(),
			PricePerShare: getDetailItemNumber(value, "Price"),
			Shares:        getDetailItemNumber(value, "Shares"),
			Source:        "fidelity.com",
			SourceID:      value.Get("orderNumber").String(),
			Ticker:        value.Get("symbol").String(),
			TotalValue:    getDollarValue(value, "amount"),
		}

		acctNum := value.Get("acctNum").String()

		// determine kind
		trx.Kind = determineTransactionKind(trx, value.Get("txnTypeCode").String(),
			value.Get("txnCatCode").String(),
			value.Get("txnSubCatCode").String())

		if trx.Kind == "" {
			// skip unknown transactions
			return true
		}

		// modify core holdings
		if isCoreHolding(trx.Ticker) {
			if trx.Kind == pvlib.BuyTransaction || trx.Kind == pvlib.SellTransaction {
				// its a buy/sell just ignore
				return true
			}
			if trx.Kind == pvlib.DividendTransaction {
				trx.Kind = pvlib.InterestTransaction
			}
		}

		if trx.Kind == pvlib.BuyTransaction && isCoreHolding(trx.Ticker) {
			// This is an investment in the core holding which is effectively a cash investment.
			// ignore the transaction
			log.Debug().Object("Transaction", &trx).Msg("skipping transaction moving money to core investment")
			return true
		}

		if trx.Kind == pvlib.DepositTransaction || trx.Kind == pvlib.WithdrawTransaction {
			trx.Ticker = "CASH"
		}

		if trx.Kind == pvlib.DepositTransaction || trx.Kind == pvlib.WithdrawTransaction || trx.Kind == pvlib.DividendTransaction || trx.Kind == pvlib.InterestTransaction {
			trx.PricePerShare = 1.0
			trx.Shares = trx.TotalValue
		}

		trx.Shares = math.Abs(trx.Shares)
		trx.PricePerShare = math.Abs(trx.PricePerShare)
		trx.TotalValue = math.Abs(trx.TotalValue)

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

	return
}
