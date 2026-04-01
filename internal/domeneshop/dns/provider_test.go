/*
 * Provider - unit tests.
 *
 * Copyright 2026 Marco Confalonieri & Kim Engebretsen.
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
	"fmt"
	"net/http"
	"testing"
	"time"

	"external-dns-domeneshop-webhook/internal/domeneshop"

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
	"github.com/stretchr/testify/assert"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

// Test_NewDomeneshopProvider tests NewDomeneshopProvider().
func Test_NewDomeneshopProvider(t *testing.T) {
	type testCase struct {
		name     string
		input    *domeneshop.Configuration
		expected struct {
			provider DomeneshopProvider
			err      error
		}
	}

	run := func(t *testing.T, tc testCase) {
		exp := tc.expected
		p, err := NewDomeneshopProvider(tc.input)
		if !assertError(t, exp.err, err) {
			assert.NotNil(t, p.client)
			assert.Equal(t, exp.provider.dryRun, p.dryRun)
			assert.Equal(t, exp.provider.debug, p.debug)
			actualJSON, _ := p.domainFilter.MarshalJSON()
			expectedJSON, _ := exp.provider.domainFilter.MarshalJSON()
			assert.Equal(t, actualJSON, expectedJSON)
		}
	}

	testCases := []testCase{
		{
			name: "empty api key",
			input: &domeneshop.Configuration{
				Token:        "",
				Secret:       "",
				DryRun:       true,
				Debug:        true,
				DomainFilter: []string{"alpha.com, beta.com"},
			},
			expected: struct {
				provider DomeneshopProvider
				err      error
			}{
				err: errors.New("cannot instantiate DNS provider: nil Token provided"),
			},
		},
		{
			name: "some api key",
			input: &domeneshop.Configuration{
				Token:        "TEST_TOKEN",
				Secret:       "TEST_API_KEY",
				DryRun:       true,
				Debug:        true,
				DomainFilter: []string{"alpha.com, beta.com"},
			},
			expected: struct {
				provider DomeneshopProvider
				err      error
			}{
				provider: DomeneshopProvider{
					client:       nil, // This will be ignored
					debug:        true,
					dryRun:       true,
					domainFilter: endpoint.NewDomainFilter([]string{"alpha.com, beta.com"}),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_Domains tests DomeneshopProvider.Domains().
func Test_Domains(t *testing.T) {
	type testCase struct {
		name     string
		provider DomeneshopProvider
		expected struct {
			domains []dsdns.Domain
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		obj := tc.provider
		exp := tc.expected
		actual, err := obj.Domains(context.Background())
		if !assertError(t, exp.err, err) {
			assert.ElementsMatch(t, exp.domains, actual)
		}
	}

	testCases := []testCase{
		{
			name: "all domains returned",
			provider: DomeneshopProvider{
				client: &mockClient{
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
						resp: &dsdns.Response{},
					},
				},
				debug:             true,
				dryRun:            false,
				domainFilter:      &endpoint.DomainFilter{},
				domainCacheDuration: time.Duration(int64(3600) * int64(time.Second)),
				domainCacheUpdate:   time.Now(),
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
			name: "filtered domains returned",
			provider: DomeneshopProvider{
				client: &mockClient{
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
							{
								ID:   "domainIDGamma",
								Domain: "gamma.com",
							},
						},
						resp: &dsdns.Response{},
					},
				},
				debug:        true,
				dryRun:       false,
				domainFilter: endpoint.NewDomainFilter([]string{"alpha.com", "gamma.com"}),
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
						ID:   "domainIDGamma",
						Domain: "gamma.com",
					},
				},
			},
		},
		{
			name: "cached domains returned",
			provider: DomeneshopProvider{
				client: &mockClient{
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
							{
								ID:   "domainIDGamma",
								Domain: "gamma.com",
							},
						},
						resp: &dsdns.Response{},
					},
				},
				debug:             true,
				dryRun:            false,
				domainFilter:      &endpoint.DomainFilter{},
				domainCacheDuration: time.Duration(int64(3600) * int64(time.Second)),
				domainCacheUpdate:   time.Now().Add(time.Duration(int64(3600) * int64(time.Second))),
				domainCache: []dsdns.Domain{
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
			name: "error returned",
			provider: DomeneshopProvider{
				client: &mockClient{
					getDomains: domainsResponse{
						err: errors.New("test domains error"),
					},
				},
				debug:        true,
				dryRun:       false,
				domainFilter: &endpoint.DomainFilter{},
			},
			expected: struct {
				domains []dsdns.Domain
				err   error
			}{
				err: errors.New("test domains error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_AdjustEndpoints tests DomeneshopProvider.AdjustEndpoints().
func Test_AdjustEndpoints(t *testing.T) {
	type testCase struct {
		name     string
		provider DomeneshopProvider
		input    []*endpoint.Endpoint
		expected []*endpoint.Endpoint
	}

	run := func(t *testing.T, tc testCase) {
		obj := tc.provider
		inp := tc.input
		exp := tc.expected
		actual, err := obj.AdjustEndpoints(inp)
		assert.Nil(t, err) // This implementation shouldn't throw errors
		assert.EqualValues(t, exp, actual)
	}

	testCases := []testCase{
		{
			name: "empty list",
			provider: DomeneshopProvider{
				domainIDNameMapper: provider.ZoneIDName{
					"domainIDAlpha": "alpha.com",
					"domainIDBeta":  "beta.com",
				},
			},
			input:    []*endpoint.Endpoint{},
			expected: []*endpoint.Endpoint{},
		},
		{
			name: "adjusted elements",
			provider: DomeneshopProvider{
				domainIDNameMapper: provider.ZoneIDName{
					"domainIDAlpha": "alpha.com",
					"domainIDBeta":  "beta.com",
				},
			},
			input: []*endpoint.Endpoint{
				{
					DNSName:    "www.alpha.com",
					RecordType: "A",
					Targets:    endpoint.Targets{"1.1.1.1"},
				},
				{
					DNSName:    "alpha.com",
					RecordType: "CNAME",
					Targets:    endpoint.Targets{"www.alpha.com."},
				},
				{
					DNSName:    "www.beta.com",
					RecordType: "A",
					Targets:    endpoint.Targets{"2.2.2.2"},
				},
				{
					DNSName:    "ftp.beta.com",
					RecordType: "CNAME",
					Targets:    endpoint.Targets{"www.alpha.com."},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "www.alpha.com",
					RecordType: "A",
					Targets:    endpoint.Targets{"1.1.1.1"},
				},
				{
					DNSName:    "alpha.com",
					RecordType: "CNAME",
					Targets:    endpoint.Targets{"www.alpha.com"},
				},
				{
					DNSName:    "www.beta.com",
					RecordType: "A",
					Targets:    endpoint.Targets{"2.2.2.2"},
				},
				{
					DNSName:    "ftp.beta.com",
					RecordType: "CNAME",
					Targets:    endpoint.Targets{"www.alpha.com"},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_Records tests DomeneshopProvider.Records().
func Test_Records(t *testing.T) {
	type testCase struct {
		name     string
		provider DomeneshopProvider
		expected struct {
			endpoints []*endpoint.Endpoint
			err       error
		}
	}

	run := func(t *testing.T, tc testCase) {
		obj := tc.provider
		exp := tc.expected
		actual, err := obj.Records(context.Background())
		if !assertError(t, exp.err, err) {
			assert.EqualValues(t, exp.endpoints, actual)
		}
	}

	testCases := []testCase{
		{
			name: "empty list",
			provider: DomeneshopProvider{
				client: &mockClient{
					getDomains: domainsResponse{
						domains: []*dsdns.Domain{
							{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
						},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
					getRecords: recordsResponse{
						records: []*dsdns.Record{},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
					filterRecordsByDomain: true, // we want the records by domain
				},
				debug:        true,
				dryRun:       false,
				domainFilter: &endpoint.DomainFilter{},
			},
			expected: struct {
				endpoints []*endpoint.Endpoint
				err       error
			}{
				endpoints: []*endpoint.Endpoint{},
			},
		},
		{
			name: "records returned",
			provider: DomeneshopProvider{
				client: &mockClient{
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
								Type: dsdns.RecordTypeCNAME,
								Domain: &dsdns.Domain{
									ID:   "domainIDAlpha",
									Domain: "alpha.com",
								},
								Data: "www",
								Ttl:   -1,
							},
							{
								ID:   "id_3",
								Host: "www",
								Type: dsdns.RecordTypeA,
								Domain: &dsdns.Domain{
									ID:   "domainIDBeta",
									Domain: "beta.com",
								},
								Data: "2.2.2.2",
								Ttl:   -1,
							},
							{
								ID:   "id_4",
								Host: "ftp",
								Type: dsdns.RecordTypeA,
								Domain: &dsdns.Domain{
									ID:   "domainIDBeta",
									Domain: "beta.com",
								},
								Data: "3.3.3.3",
								Ttl:   -1,
							},
						},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
					filterRecordsByDomain: true,
				},
				debug:        true,
				dryRun:       false,
				domainFilter: &endpoint.DomainFilter{},
			},
			expected: struct {
				endpoints []*endpoint.Endpoint
				err       error
			}{
				endpoints: []*endpoint.Endpoint{
					{
						DNSName:    "www.alpha.com",
						RecordType: "A",
						Targets:    endpoint.Targets{"1.1.1.1"},
						Labels:     endpoint.Labels{},
						RecordTTL:  -1,
					},
					{
						DNSName:    "ftp.alpha.com",
						RecordType: "CNAME",
						Targets:    endpoint.Targets{"www.alpha.com"},
						Labels:     endpoint.Labels{},
						RecordTTL:  -1,
					},
					{
						DNSName:    "www.beta.com",
						RecordType: "A",
						Targets:    endpoint.Targets{"2.2.2.2"},
						Labels:     endpoint.Labels{},
						RecordTTL:  -1,
					},
					{
						DNSName:    "ftp.beta.com",
						RecordType: "A",
						Targets:    endpoint.Targets{"3.3.3.3"},
						Labels:     endpoint.Labels{},
						RecordTTL:  -1,
					},
				},
			},
		},
		{
			name: "error getting domains",
			provider: DomeneshopProvider{
				client: &mockClient{
					getDomains: domainsResponse{
						err: errors.New("test domains error"),
					},
				},
				debug:        true,
				dryRun:       false,
				domainFilter: &endpoint.DomainFilter{},
			},
			expected: struct {
				endpoints []*endpoint.Endpoint
				err       error
			}{
				err: errors.New("test domains error"),
			},
		},
		{
			name: "error getting records",
			provider: DomeneshopProvider{
				client: &mockClient{
					getDomains: domainsResponse{
						domains: []*dsdns.Domain{
							{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
						},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
					getRecords: recordsResponse{
						err: errors.New("test records error"),
					},
				},
				debug:        true,
				dryRun:       false,
				domainFilter: &endpoint.DomainFilter{},
			},
			expected: struct {
				endpoints []*endpoint.Endpoint
				err       error
			}{
				err: errors.New("test records error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_ensureDomainIDMappingPresent tests
// DomeneshopProvider.ensureDomainIDMappingPresent().
func Test_ensureDomainIDMappingPresent(t *testing.T) {
	type testCase struct {
		name     string
		provider DomeneshopProvider
		input    []dsdns.Domain
		expected map[string]string
	}

	run := func(t *testing.T, tc testCase) {
		tc.provider.ensureDomainIDMappingPresent(tc.input)
		actual := tc.provider.domainIDNameMapper
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name:     "empty list",
			provider: DomeneshopProvider{},
			input:    []dsdns.Domain{},
			expected: map[string]string{},
		},
		{
			name:     "domains present",
			provider: DomeneshopProvider{},
			input: []dsdns.Domain{
				{
					ID:   "domainIDAlpha",
					Domain: "alpha.com",
				},
				{
					ID:   "domainIDBeta",
					Domain: "beta.com",
				},
			},
			expected: map[string]string{
				"domainIDAlpha": "alpha.com",
				"domainIDBeta":  "beta.com",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_getRecordsByDomainID tests DomeneshopProvider.getRecordsByDomainID()
func Test_getRecordsByDomainID(t *testing.T) {
	type testCase struct {
		name     string
		provider DomeneshopProvider
		expected struct {
			recordsByDomainID map[string][]dsdns.Record
			err             error
		}
	}

	run := func(t *testing.T, tc testCase) {
		obj := tc.provider
		exp := tc.expected
		actual, err := obj.getRecordsByDomainID(context.Background())
		if assertError(t, exp.err, err) {
			assert.ElementsMatch(t, exp.recordsByDomainID, actual)
		}
	}

	testCases := []testCase{
		{
			name: "empty list",
			provider: DomeneshopProvider{
				client: &mockClient{
					getDomains: domainsResponse{
						domains: []*dsdns.Domain{
							{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
						},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
					getRecords: recordsResponse{
						records: []*dsdns.Record{},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
					filterRecordsByDomain: true, // we want the records by domain
				},
				debug:        true,
				dryRun:       false,
				domainFilter: &endpoint.DomainFilter{},
			},
			expected: struct {
				recordsByDomainID map[string][]dsdns.Record
				err             error
			}{
				recordsByDomainID: map[string][]dsdns.Record{},
			},
		},
		{
			name: "records returned",
			provider: DomeneshopProvider{
				client: &mockClient{
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
								Type: dsdns.RecordTypeCNAME,
								Domain: &dsdns.Domain{
									ID:   "domainIDAlpha",
									Domain: "alpha.com",
								},
								Data: "www",
								Ttl:   -1,
							},
							{
								ID:   "id_3",
								Host: "www",
								Type: dsdns.RecordTypeA,
								Domain: &dsdns.Domain{
									ID:   "domainIDBeta",
									Domain: "beta.com",
								},
								Data: "2.2.2.2",
								Ttl:   -1,
							},
							{
								ID:   "id_4",
								Host: "ftp",
								Type: dsdns.RecordTypeA,
								Domain: &dsdns.Domain{
									ID:   "domainIDBeta",
									Domain: "beta.com",
								},
								Data: "3.3.3.3",
								Ttl:   -1,
							},
						},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
					filterRecordsByDomain: true,
				},
				debug:        true,
				dryRun:       false,
				domainFilter: &endpoint.DomainFilter{},
			},
			expected: struct {
				recordsByDomainID map[string][]dsdns.Record
				err             error
			}{
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
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
							Type: dsdns.RecordTypeCNAME,
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Data: "www",
							Ttl:   -1,
						},
					},
					"domainIDBeta": {
						{
							ID:   "id_3",
							Host: "www",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "2.2.2.2",
							Ttl:   -1,
						},
						{
							ID:   "id_4",
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "3.3.3.3",
							Ttl:   -1,
						},
					},
				},
			},
		},
		{
			name: "error getting domains",
			provider: DomeneshopProvider{
				client: &mockClient{
					getDomains: domainsResponse{
						err: errors.New("test domains error"),
					},
				},
				debug:        true,
				dryRun:       false,
				domainFilter: &endpoint.DomainFilter{},
			},
			expected: struct {
				recordsByDomainID map[string][]dsdns.Record
				err             error
			}{
				err: errors.New("test domains error"),
			},
		},
		{
			name: "error getting records",
			provider: DomeneshopProvider{
				client: &mockClient{
					getDomains: domainsResponse{
						domains: []*dsdns.Domain{
							{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
						},
						resp: &dsdns.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						},
					},
					getRecords: recordsResponse{
						err: errors.New("test records error"),
					},
				},
				debug:        true,
				dryRun:       false,
				domainFilter: &endpoint.DomainFilter{},
			},
			expected: struct {
				recordsByDomainID map[string][]dsdns.Record
				err             error
			}{
				err: errors.New("test records error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_incFailCount tests incFailCount().
func Test_incFailCount(t *testing.T) {
	type testCase struct {
		name     string
		object   *DomeneshopProvider
		expected struct {
			failCount       int
			logFatalfCalled bool
			logFatalfMsg    string
		}
	}

	run := func(t *testing.T, tc testCase) {
		origLogFatalf := logFatalf
		obj := tc.object
		exp := tc.expected
		logFatalfCalled := false
		logFatalfMsg := ""

		// Mock logFatalf
		logFatalf = func(format string, a ...interface{}) {
			logFatalfCalled = true
			logFatalfMsg = fmt.Sprintf(format, a...)
		}
		// Do the call
		obj.incFailCount()
		// Restore logFatalf
		logFatalf = origLogFatalf

		assert.Equal(t, exp.failCount, obj.failCount)
		assert.Equal(t, exp.logFatalfCalled, logFatalfCalled)
		assert.Equal(t, exp.logFatalfMsg, logFatalfMsg)
	}

	testCases := []testCase{
		{
			name: "failCount is disabled",
			object: &DomeneshopProvider{
				maxFailCount: -1,
				failCount:    -1, // impossible value, but will not be reset if disabled
			},
			expected: struct {
				failCount       int
				logFatalfCalled bool
				logFatalfMsg    string
			}{
				failCount: -1,
			},
		},
		{
			name: "failCount is enabled and zero",
			object: &DomeneshopProvider{
				maxFailCount: 3,
				failCount:    0,
			},
			expected: struct {
				failCount       int
				logFatalfCalled bool
				logFatalfMsg    string
			}{
				failCount: 1,
			},
		},
		{
			name: "failCount is enabled and low",
			object: &DomeneshopProvider{
				maxFailCount: 3,
				failCount:    1,
			},
			expected: struct {
				failCount       int
				logFatalfCalled bool
				logFatalfMsg    string
			}{
				failCount: 2,
			},
		},
		{
			name: "failCount is enabled and high",
			object: &DomeneshopProvider{
				maxFailCount: 3,
				failCount:    2,
			},
			expected: struct {
				failCount       int
				logFatalfCalled bool
				logFatalfMsg    string
			}{
				failCount:       3,
				logFatalfCalled: true,
				logFatalfMsg:    "Failure count reached 3. Shutting down container.",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_resetFailCount tests resetFailCount().
func Test_resetFailCount(t *testing.T) {
	type testCase struct {
		name     string
		object   *DomeneshopProvider
		expected int
	}

	run := func(t *testing.T, tc testCase) {
		obj := tc.object
		exp := tc.expected
		obj.resetFailCount()
		assert.Equal(t, exp, obj.failCount)
	}

	testCases := []testCase{
		{
			name: "failCount is disabled",
			object: &DomeneshopProvider{
				maxFailCount: -1,
				failCount:    -1, // impossible value, but will not be reset if disabled
			},
			expected: -1,
		},
		{
			name: "failCount is enabled and zero",
			object: &DomeneshopProvider{
				maxFailCount: 3,
				failCount:    0,
			},
			expected: 0,
		},
		{
			name: "failCount is enabled and not zero",
			object: &DomeneshopProvider{
				maxFailCount: 3,
				failCount:    2,
			},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
