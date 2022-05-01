package fidelity

import (
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
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

func AccountActivity(page playwright.Page) {
	// load the activity page
	if _, err := page.Goto(ACTIVITY_URL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Error().Err(err).Msg("could not load activity page")
	}

	req := page.WaitForRequest(ACTIVITY_API_URL)
	resp, err := req.Response()
	if err != nil {
		log.Error().Err(err).Msg("error while waiting for response to activity api")
	}

	body, err := resp.Body()
	if err != nil {
		log.Error().Err(err).Msg("error while fetching body")
	}
}
