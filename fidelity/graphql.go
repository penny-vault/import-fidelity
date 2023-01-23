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

type GraphQLQuery struct {
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
	Query         string         `json:"query"`
}

var (
	GQLGetContext = `query GetContext {
  getContext {
    sysStatus {
      balance
      backend {
        account
        feature
        __typename
      }
      account {
        Brokerage
        StockPlans
        ExternalLinked
        ExternalManual
        WorkplaceContributions
        WorkplaceBenefits
        Annuity
        FidelityCreditCards
        Charitable
        BrokerageLending
        InternalDigital
        ExternalDigital
        __typename
      }
      __typename
    }
    person {
      id
      sysMsgs {
        message
        source
        code
        type
        __typename
      }
      relationships {
        type
        __typename
      }
      balances {
        balanceDetail {
          gainLossBalanceDetail {
            totalMarketVal
            todaysGainLoss
            todaysGainLossPct
            __typename
          }
          __typename
        }
        __typename
      }
      assets {
        acctNum
        acctType
        acctSubType
        acctSubTypeDesc
        acctCreationDate
        parentBrokAcctNum
        linkedAcctDetails {
          acctNum
          isLinked
          __typename
        }
        brokerageLendingAcctDetail {
          institutionName
          creditLineAmount
          lineAvailablityAmount
          endInterestRate
          baseInterestRate
          spreadToBaseRate
          baseIndexName
          nextPaymentDueDate
          lastPaymentDate
          paymentAmountDue
          loanStatus
          pledgedAccountNumber
          __typename
        }
        acctStateDetail {
          statusCode
          __typename
        }
        preferenceDetail {
          name
          isHidden
          isDefaultAcct
          acctGroupId
          __typename
        }
        gainLossBalanceDetail {
          totalMarketVal
          todaysGainLoss
          todaysGainLossPct
          asOfDateTime
          hasUnpricedPositions
          __typename
        }
        acctRelAttrDetail {
          relCategoryCode
          relRoleTypeCode
          __typename
        }
        annuityProductDetail {
          systemOfRecord
          planTypeCode
          planCode
          productCode
          productDesc
          __typename
        }
        workplacePlanDetail {
          planInTransitionInd
          planTypeName
          planTypeCode
          planId
          clientId
          clientTickerSymbol
          enrollmentStatusCode
          isCrossoverEnabled
          isEnrollmentEligible
          nonQualifiedInd
          isRollup
          planName
          navigationKey
          url
          __typename
        }
        acctTypesIndDetail {
          isRetirement
          isYouthAcct
          __typename
        }
        acctAttrDetail {
          regTypeDesc
          costBasisCode
          addlBrokAcctCode
          __typename
        }
        acctIndDetail {
          isAdvisorAcct
          isAuthorizedAcct
          isMultiCurrencyAllowed
          isFFOSAcct
          isPrimaryCustomer
          __typename
        }
        acctTrustIndDetail {
          isTrustAcct
          __typename
        }
        acctLegalAttrDetail {
          accountTypeCode
          legalConstructCode
          legalConstructModifierCode
          offeringCode
          serviceSegmentCode
          lineOfBusinessCode
          __typename
        }
        acctTradeAttrDetail {
          optionAgrmntCode
          optionLevelCode
          borrowFullyPaidCode
          portfolioMarginCode
          isTradable
          mrgnAgrmntCode
          isSpecificShrTradingEligible
          isSpreadsAllowed
          limitedMrgnCode
          __typename
        }
        annuityPolicyDetail {
          policyStatus
          isImmediateLiquidityEnabled
          regTypeCode
          __typename
        }
        externalAcctDetail {
          acctType
          acctSubType
          __typename
        }
        managedAcctDetail {
          productCode
          svcModelCode
          __typename
        }
        __typename
      }
      groups {
        id
        name
        items {
          acctNum
          acctType
          acctSubType
          acctSubTypeDesc
          acctCreationDate
          parentBrokAcctNum
          linkedAcctDetails {
            acctNum
            isLinked
            __typename
          }
          acctStateDetail {
            statusCode
            __typename
          }
          acctAttrDetail {
            addlBrokAcctCode
            regTypeDesc
            __typename
          }
          acctIndDetail {
            isAdvisorAcct
            isAuthorizedAcct
            isMultiCurrencyAllowed
            isFFOSAcct
            isPrimaryCustomer
            __typename
          }
          acctTrustIndDetail {
            isTrustAcct
            __typename
          }
          acctTypesIndDetail {
            isRetirement
            isYouthAcct
            hasSPSPlans
            __typename
          }
          acctRelAttrDetail {
            relCategoryCode
            relRoleTypeCode
            __typename
          }
          preferenceDetail {
            name
            isHidden
            isDefaultAcct
            acctGroupId
            __typename
          }
          acctLegalAttrDetail {
            legalConstructCode
            legalConstructModifierCode
            serviceSegmentCode
            accountTypeCode
            offeringCode
            lineOfBusinessCode
            __typename
          }
          workplacePlanDetail {
            planTypeName
            planTypeCode
            planId
            clientId
            clientTickerSymbol
            enrollmentStatusCode
            isCrossoverEnabled
            isEnrollmentEligible
            nonQualifiedInd
            isRollup
            planName
            navigationKey
            url
            __typename
          }
          gainLossBalanceDetail {
            totalMarketVal
            todaysGainLoss
            todaysGainLossPct
            asOfDateTime
            hasUnpricedPositions
            __typename
          }
          annuityProductDetail {
            systemOfRecord
            planTypeCode
            planCode
            productCode
            productDesc
            __typename
          }
          annuityPolicyDetail {
            policyStatus
            isImmediateLiquidityEnabled
            __typename
          }
          externalAcctDetail {
            acctType
            acctSubType
            __typename
          }
          managedAcctDetail {
            productCode
            svcModelCode
            __typename
          }
          digiAcctAttrDetail {
            currencyCode
            currencyType
            amount
            __typename
          }
          __typename
        }
        balanceDetail {
          gainLossBalanceDetail {
            totalMarketVal
            todaysGainLoss
            todaysGainLossPct
            __typename
          }
          __typename
        }
        __typename
      }
      customerAttrDetail {
        externalCustomerID
        isShowWorkplaceSavingAccts
        isShowExternalAccts
        pledgedAcctNums
        __typename
      }
      groupDetails {
        groupId
        groupName
        typeCode
        __typename
      }
      __typename
    }
    __typename
  }
}
`
	GQLGetAcctFeatureContext = `query GetAcctFeatureContext($acctList: [AcctFeatureParamters], $featureParamsDetail: FeatureParamsDetail) {
  getAcctFeatureContext(
    acctList: $acctList
    featureParamsDetail: $featureParamsDetail
  ) {
    acctFeatures {
      acctNum
      featureDetails {
        eligible {
          moneyMovementDetail {
            automaticInvestmentDetail {
              isEligible
              __typename
            }
            hasBillPay
            hasBankWire
            hasEFT
            hasAutomaticWithdrawal
            __typename
          }
          fundAccessDetail {
            checkWritingDetail {
              isEligible
              ineligibilityReason
              hasReorderedChecks
              __typename
            }
            hasDebitCard
            hasDepositSlips
            __typename
          }
          __typename
        }
        established {
          moneyMovementDetail {
            hasAutomaticInvestments
            hasBankWire
            hasBillPay
            hasEFT
            hasAutomaticWithdrawal
            __typename
          }
          fundAccessDetail {
            hasCheckWriting
            hasDebitCard
            hasDepositSlips
            __typename
          }
          __typename
        }
        __typename
      }
      __typename
    }
    __typename
  }
}
`
	GQLGetTransactions = `query getTransactions($acctIdList: String, $acctDetailList: [AcctDetailList], $searchCriteriaDetail: SearchCriteriaDetail, $isNewOrderApi: Boolean! = false, $isSupportCrypto: Boolean! = false) {
  getTransactions(
    acctIdList: $acctIdList
    acctDetailList: $acctDetailList
    searchCriteriaDetail: $searchCriteriaDetail
    isNewOrderApi: $isNewOrderApi
    isSupportCrypto: $isSupportCrypto
  ) {
    backendStatus {
      order
      history
      transfers
      billpay
      __typename
    }
    orders {
      acctNum
      description
      date
      amount
      confNumOrig
      actionCode
      status
      symbol
      secType
      briefSymbol
      cancelParameters
      replaceParameters
      orderDate
      detailItems {
        key
        value
        __typename
      }
      isOption
      isCrypto
      isMutualFund
      cusip
      totalPriceImprovement
      displayEditExpirationLink
      displayQuoteRequestId
      isSpecificShareOrder
      specificShareOrderURL
      displayExecutions {
        executions {
          execDate
          execTime
          price
          amt
          totalAmount
          __typename
        }
        strTotalExecShares
        __typename
      }
      qtyExec
      totalAmountForExecutions
      displayExchangeFund {
        exchFundPrice
        exchFundShares
        exchangeFundTotal
        __typename
      }
      isInternationalOrder
      fixedIncomePriceTypeCode
      dbCrEvenIndicator
      priceCurrencyCode
      displayLimitPriceStr2
      isExchange
      isCurrencyExchange
      isAutoCurrencyExchange
      fxExecutions
      fxExecDate
      fxExecTime
      exchangeRate
      fromQuantityCurrency
      toQuantityCurrency
      displayOrderDescription
      __typename
    }
    historys {
      acctNum
      txnTypNum
      orderNumber
      isAnnuity
      isExchange
      isCrypto
      cryptoType
      description
      date
      amount
      txnTypeCode
      txnCatCode
      txnSubCatCode
      investmentTypeCode
      status
      symbol
      cashBalance
      filiCsvData
      brokCsvData
      imageDetail {
        imageId
        checkImage
        checkImageURL
        __typename
      }
      detailItems {
        key
        value
        __typename
      }
      annuityDetail {
        accumulationPayoutFlag
        accumulationPayoutTypeCode
        fundDetails {
          accountCode
          divisionCode
          fundAmount
          fundAmountTypeCode
          fundCode
          fundCodeDisplayStyleFlag
          fundDirection
          latestUnitValueDetail {
            price
            __typename
          }
          longName
          quantity
          recType
          redemptionFeeFlag
          symbol
          redemptionFee
          transactionAmount
          __typename
        }
        disbursementDetailFlag
        disbursementDetail {
          federalWithholding
          stateWithholding
          disbursement
          __typename
        }
        __typename
      }
      __typename
    }
    footNote {
      showSettlements
      isMoreThanTwentyAccounts
      selectMultipleAccounts
      isRetirementIncome
      isLinkBrokerage
      asOfDate
      hasShadowTransaction
      showCurrency
      __typename
    }
    __typename
  }
}
`
)
