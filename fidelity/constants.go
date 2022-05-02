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

const (
	HOMEPAGE_URL     = `https://fidelity.com/`
	LOGIN_URL        = `https://digital.fidelity.com/prgw/digital/login/full-page`
	LOGOUT_URL       = `https://login.fidelity.com/ftgw/Fidelity/RtlCust/Logout/Init?AuthRedUrl=https://www.fidelity.com/customer-service/customer-logout`
	SUMMARY_URL      = `https://oltx.fidelity.com/ftgw/fbc/oftop/portfolio#summary`
	ACTIVITY_URL     = `https://oltx.fidelity.com/ftgw/fbc/oftop/portfolio#activity`
	ACTIVITY_API_URL = `https://digital.fidelity.com/ftgw/digital/acct-activity/activity-tab-history/api`
	QUOTE_URL        = `https://digital.fidelity.com/prgw/digital/research/quote/dashboard/summary?symbol=%s`
	MARKET_DATA_URL  = `https://api.markitdigital.com/xref/v1/symbols/%s`
)
