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

package fidelity_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/penny-vault/import-fidelity/fidelity"
	"github.com/penny-vault/pvlib"
)

var _ = Describe("Account activity", func() {
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
			fidelityActivityJSON, err = os.ReadFile("../test/getTransactions.json")
			Expect(err).NotTo(HaveOccurred())
			trxMap, err = fidelity.ParseAccountActivity(string(fidelityActivityJSON))
		})

		It("does not error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns 4 accounts", func() {
			Expect(len(trxMap)).Should(Equal(4))
		})

		It("has correct account numbers", func() {
			for k := range trxMap {
				fmt.Println(k)
			}
			_, ok := trxMap["200000001"]
			Expect(ok).To(BeTrue(), "200000001")

			_, ok = trxMap["Z00000001"]
			Expect(ok).To(BeTrue(), "Z00000001")

			_, ok = trxMap["Z00000002"]
			Expect(ok).To(BeTrue(), "Z00000002")

			_, ok = trxMap["200000002"]
			Expect(ok).To(BeTrue(), "200000002")
		})

		It("has correct number of transactions", func() {
			acct := trxMap["Z00000001"]
			Expect(acct).To(HaveLen(11), "Z00000001")

			acct = trxMap["Z00000002"]
			Expect(acct).To(HaveLen(26), "Z00000002")

			acct = trxMap["200000001"]
			Expect(acct).To(HaveLen(12), "200000001")

			acct = trxMap["200000002"]
			Expect(acct).To(HaveLen(3), "200000002")
		})

		It("should have a transaction for sale of whole shares of STIP", func() {
			acct := trxMap["200000001"]
			cnt := 0
			for _, trx := range acct {
				if trx.Kind == pvlib.SellTransaction && trx.Ticker == "STIP" && trx.Shares == 1834 {
					cnt++
					Expect(trx.Kind).To(Equal(pvlib.SellTransaction))
					Expect(trx.Shares).To(Equal(1834.0))
				}
			}
			Expect(cnt).To(Equal(1))
		})
	})
})
