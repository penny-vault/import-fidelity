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

package fidelity_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/penny-vault/import-fidelity/fidelity"
	"github.com/penny-vault/pvlib"
)

var _ = Describe("Activity", func() {
	Describe("parse account activity", func() {
		var err error
		var trxMap map[string][]*pvlib.Transaction

		When("JSON fails to parse", func() {
			BeforeEach(func() {
				trxMap, err = fidelity.ParseAccountActivity("")
			})

			It("returns an empty transaction map", func() {
				Expect(len(trxMap)).To(Equal(0))
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("JSON has multiple accounts", func() {
			BeforeEach(func() {
				var fidelityActivityJSON []byte
				fidelityActivityJSON, err = os.ReadFile("../test/transactions-06022022.json")
				Expect(err).NotTo(HaveOccurred())
				trxMap, err = fidelity.ParseAccountActivity(string(fidelityActivityJSON))
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns 3 accounts", func() {
				// there are 4 accounts in the file but 1 of
				// them only has intraDay transactions
				Expect(len(trxMap)).Should(Equal(3))
			})

			It("has test that failes", func() {
				Expect(false).To(BeTrue())
			})

			It("has correct account numbers", func() {
				_, ok := trxMap["238000000"]
				Expect(ok).To(BeFalse(), "238000000")

				_, ok = trxMap["Z00000000"]
				Expect(ok).To(BeTrue(), "Z00000000")

				_, ok = trxMap["Z00000001"]
				Expect(ok).To(BeTrue(), "Z00000001")

				_, ok = trxMap["244000000"]
				Expect(ok).To(BeTrue(), "244000000")
			})

			It("has correct number of transactions", func() {
				acct := trxMap["Z00000000"]
				Expect(acct).To(HaveLen(3), "Z00000000")

				acct = trxMap["Z00000001"]
				Expect(acct).To(HaveLen(5), "Z00000001")

				acct = trxMap["244000000"]
				Expect(acct).To(HaveLen(1), "244000000")
			})

			It("should have a transaction for sale of whole shares of STIP", func() {
				acct := trxMap["Z00000001"]
				cnt := 0
				for _, trx := range acct {
					if trx.Kind == pvlib.SellTransaction && trx.Ticker == "STIP" && trx.Shares < -1 {
						cnt++
						Expect(trx.Kind).To(Equal(pvlib.SellTransaction))
						Expect(trx.Shares).To(Equal(-32.0))
					}
				}
				Expect(cnt).To(Equal(1))
			})

			It("should have a transaction for sale of fractional shares of STIP", func() {
				acct := trxMap["Z00000001"]
				cnt := 0
				for _, trx := range acct {
					if trx.Kind == pvlib.SellTransaction && trx.Ticker == "STIP" && trx.Shares < 0 && trx.Shares > -1 {
						cnt++
						Expect(trx.Kind).To(Equal(pvlib.SellTransaction))
						Expect(trx.Shares).To(Equal(-0.598))
					}
				}
				Expect(cnt).To(Equal(1))
			})
		})
	})
})
