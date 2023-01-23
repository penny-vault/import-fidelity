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
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

type Account struct {
	AccountNumber             string  `json:"acctNum"`
	AccountType               string  `json:"acctType"`
	AccountSubType            string  `json:"acctSubType"`
	AccountSubTypeDescription string  `json:"acctSubTypeDesc"`
	Name                      string  `json:"name"`
	BorrowFullyPaidCode       string  `json:"borrowFullyPaidCode"`
	RegTypeDescription        string  `json:"regTypeDesc"`
	IsMultiCurrencyAllowed    bool    `json:"isMultiCurrencyAllowed"`
	RelationshipRoleTypeCode  string  `json:"relRoleTypeCode"`
	CostBasisCode             *string `json:"costBasisCode"`
	IsTradable                bool    `json:"isTradable"`
	SystemOfRecord            *string `json:"sysOfRcd"`
	BillPayEnrolled           bool    `json:"billPayEnrolled"`
}

func GetAccounts(client *resty.Client) ([]*Account, error) {
	gqlQuery := GraphQLQuery{
		OperationName: "GetContext",
		Variables:     map[string]any{},
		Query:         GQLGetContext,
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

	// create account array
	numAccounts := gjson.Get(bodyStr, "data.getContext.person.assets.#").Int()
	accounts := make([]*Account, 0, numAccounts)
	result := gjson.Get(bodyStr, "data.getContext.person.assets")
	result.ForEach(func(key, value gjson.Result) bool {
		accounts = append(accounts, &Account{
			AccountNumber:             value.Get("acctNum").String(),
			AccountType:               value.Get("acctType").String(),
			AccountSubType:            value.Get("acctSubType").String(),
			AccountSubTypeDescription: value.Get("acctSubTypeDesc").String(),
			Name:                      value.Get("preferenceDetail.name").String(),
			RegTypeDescription:        value.Get("acctAttrDetail.regTypeDesc").String(),
			BorrowFullyPaidCode:       value.Get("acctTradeAttrDetail.borrowFullyPaidCode").String(),
			IsMultiCurrencyAllowed:    value.Get("acctIndDetail.isMultiCurrencyAllowed").Bool(),
			RelationshipRoleTypeCode:  value.Get("acctRelAttrDetail.relRoleTypeCode").String(),
			IsTradable:                value.Get("acctTradeAttrDetail.isTradable").Bool(),
		})

		fmt.Println(value.Get("acctAttrDetail.costBasisCode").String())
		return true
	})

	return accounts, nil
}
