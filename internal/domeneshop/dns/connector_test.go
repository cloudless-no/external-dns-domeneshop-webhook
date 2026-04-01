/*
 * Connector - unit tests.
 *
 * Copyright 2026 Marco Confalonieri.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package domeneshopdns

import (
	"context"
	"errors"
	"net/http"
	"testing"

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
	"github.com/stretchr/testify/assert"
)

// Test_fetchRecords tests fetchRecords().
func Test_fetchRecords(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainID    string
			dnsClient apiClient
			batchSize int
		}
		expected struct {
			records []dsdns.Record
			err     error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		actual, err := fetchRecords(context.Background(), inp.domainID, inp.dnsClient, inp.batchSize)
		if !assertError(t, exp.err, err) {
			assert.ElementsMatch(t, exp.records, actual)
		}
	}

	testCases := []testCase{
		{
			name: "records fetched",
			input: struct {
				domainID    string
				dnsClient apiClient
				batchSize int
			}{
				domainID: "domainIDAlpha",
				dnsClient: &mockClient{
					getRecords: recordsResponse{
						records: []*dsdns.Record{
							{
								ID:   "id_1",
								Host: "www",
								Type: dsdns.RecordTypeA,
								Domain: &dsdns.Domain{
									ID:   "domainIDAlpha",
									Domain: "alpha.com",
								},
								Data: "1.1.1.1",
								Ttl:   -1,
							},
							{
								ID:   "id_2",
								Host: "ftp",
								Type: dsdns.RecordTypeA,
								Domain: &dsdns.Domain{
									ID:   "domainIDAlpha",
									Domain: "alpha.com",
								},
								Data: "2.2.2.2",
								Ttl:   -1,
							},
							{
								ID:   "id_3",
								Host: "mail",
								Type: dsdns.RecordTypeMX,
								Domain: &dsdns.Domain{
									ID:   "domainIDAlpha",
									Domain: "alpha.com",
								},
								Data: "3.3.3.3",
								Ttl:   -1,
							},
						},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
				},
				batchSize: 100,
			},
			expected: struct {
				records []dsdns.Record
				err     error
			}{
				records: []dsdns.Record{
					{
						ID:   "id_1",
						Host: "www",
						Type: dsdns.RecordTypeA,
						Domain: &dsdns.Domain{
							ID:   "domainIDAlpha",
							Domain: "alpha.com",
						},
						Data: "1.1.1.1",
						Ttl:   -1,
					},
					{
						ID:   "id_2",
						Host: "ftp",
						Type: dsdns.RecordTypeA,
						Domain: &dsdns.Domain{
							ID:   "domainIDAlpha",
							Domain: "alpha.com",
						},
						Data: "2.2.2.2",
						Ttl:   -1,
					},
					{
						ID:   "id_3",
						Host: "mail",
						Type: dsdns.RecordTypeMX,
						Domain: &dsdns.Domain{
							ID:   "domainIDAlpha",
							Domain: "alpha.com",
						},
						Data: "3.3.3.3",
						Ttl:   -1,
					},
				},
			},
		},
		{
			name: "error fetching records",
			input: struct {
				domainID    string
				dnsClient apiClient
				batchSize int
			}{
				domainID: "domainIDAlpha",
				dnsClient: &mockClient{
					getRecords: recordsResponse{
						err: errors.New("records test error"),
					},
				},
				batchSize: 100,
			},
			expected: struct {
				records []dsdns.Record
				err     error
			}{
				err: errors.New("records test error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_fetchDomains tests DomeneshopProvider.fetchDomains().
func Test_fetchDomains(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			dnsClient apiClient
			batchSize int
		}
		expected struct {
			domains []dsdns.Domain
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		actual, err := fetchDomains(context.Background(), inp.dnsClient, inp.batchSize)
		if !assertError(t, exp.err, err) {
			assert.ElementsMatch(t, actual, exp.domains)
		}
	}

	testCases := []testCase{
		{
			name: "domains fetched",
			input: struct {
				dnsClient apiClient
				batchSize int
			}{
				dnsClient: &mockClient{
					getDomains: domainsResponse{
						domains: []*dsdns.Domain{
							{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
						},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
				},
				batchSize: 100,
			},
			expected: struct {
				domains []dsdns.Domain
				err   error
			}{
				domains: []dsdns.Domain{
					{
						ID:   "domainIDAlpha",
						Domain: "alpha.com",
					},
					{
						ID:   "domainIDBeta",
						Domain: "beta.com",
					},
				},
			},
		},
		{
			name: "error fetching domains",
			input: struct {
				dnsClient apiClient
				batchSize int
			}{
				dnsClient: &mockClient{
					getDomains: domainsResponse{
						err: errors.New("domains test error"),
					},
				},
				batchSize: 100,
			},
			expected: struct {
				domains []dsdns.Domain
				err   error
			}{
				err: errors.New("domains test error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
