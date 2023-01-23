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

const (
	HomePageURL   = `https://fidelity.com/`
	LoginURL      = `https://digital.fidelity.com/prgw/digital/login/full-page`
	LogoutURL     = `https://login.fidelity.com/ftgw/Fidelity/RtlCust/Logout/Init?AuthRedUrl=https://www.fidelity.com/customer-service/customer-logout`
	SummaryURL    = `https://digital.fidelity.com/ftgw/digital/portfolio/summary`
	ActivityURL   = `https://digital.fidelity.com/ftgw/digital/portfolio/activity`
	GraphQLURL    = `https://digital.fidelity.com/ftgw/digital/portfolio/api/graphql`
	QuoteURL      = `https://digital.fidelity.com/prgw/digital/research/quote/dashboard/summary?symbol=%s`
	MarketDataURL = `https://api.markitdigital.com/xref/v1/symbols/%s`
	CUSIPURL      = `https://quotes.fidelity.com/mmnet/SymLookup.phtml?reqforlookup=REQUESTFORLOOKUP&productid=mmnet&isLoggedIn=mmnet&rows=50&for=%s&by=symbol&criteria=%s&submit=Search`
)
