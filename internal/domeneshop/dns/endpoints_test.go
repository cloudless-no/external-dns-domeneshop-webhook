/*
 * Endpoints - unit tests.
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
	"testing"

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

// Test_fromDomeneshopHostname tests fromDomeneshopHostname().
func Test_fromDomeneshopHostname(t *testing.T) {
	type testCase struct {
		name     string
		domain   string
		host     string
		expected string
	}

	testCases := []testCase{
		// Apex record
		{
			name:     "apex record @",
			domain:   "alpha.com",
			host:     "@",
			expected: "alpha.com",
		},
		// Local subdomains (no dots)
		{
			name:     "local subdomain",
			domain:   "alpha.com",
			host:     "mail",
			expected: "mail.alpha.com",
		},
		{
			name:     "local subdomain www",
			domain:   "alpha.com",
			host:     "www",
			expected: "www.alpha.com",
		},
		// External hostnames with trailing dot (Domeneshop returns these for external)
		{
			name:     "external hostname with trailing dot",
			domain:   "alpha.com",
			host:     "mail.beta.com.",
			expected: "mail.beta.com",
		},
		{
			name:     "external hostname deeper with trailing dot",
			domain:   "alpha.com",
			host:     "a.b.c.beta.com.",
			expected: "a.b.c.beta.com",
		},
		// Deep local subdomains (dots but no trailing dot)
		{
			name:     "deep local subdomain",
			domain:   "alpha.com",
			host:     "a.b.c",
			expected: "a.b.c.alpha.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := fromDomeneshopHostname(tc.domain, tc.host)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

// Test_makeEndpointName tests makeEndpointName().
func Test_makeEndpointName(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domain    string
			entryName string
			epType    string
		}
		expected string
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := makeEndpointName(inp.domain, inp.entryName)
		assert.Equal(t, actual, tc.expected)
	}

	testCases := []testCase{
		{
			name: "no adjustment required",
			input: struct {
				domain    string
				entryName string
				epType    string
			}{
				domain:    "alpha.com",
				entryName: "test",
				epType:    "A",
			},
			expected: "test",
		},
		{
			name: "stripping domain from name",
			input: struct {
				domain    string
				entryName string
				epType    string
			}{
				domain:    "alpha.com",
				entryName: "test.alpha.com",
				epType:    "A",
			},
			expected: "test",
		},
		{
			name: "top entry adjustment",
			input: struct {
				domain    string
				entryName string
				epType    string
			}{
				domain:    "alpha.com",
				entryName: "alpha.com",
				epType:    "A",
			},
			expected: "@",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_makeEndpointTarget tests makeEndpointTarget().
func Test_makeEndpointTarget(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domain      string
			entryTarget string
			epType      string
		}
		expected string
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		actual := makeEndpointTarget(inp.domain, inp.entryTarget, inp.epType)
		assert.Equal(t, exp, actual)
	}

	testCases := []testCase{
		{
			name: "IP without domain provided",
			input: struct {
				domain      string
				entryTarget string
				epType      string
			}{
				domain:      "",
				entryTarget: "0.0.0.0",
				epType:      "A",
			},
			expected: "0.0.0.0",
		},
		{
			name: "IP with domain provided",
			input: struct {
				domain      string
				entryTarget string
				epType      string
			}{
				domain:      "alpha.com",
				entryTarget: "0.0.0.0",
				epType:      "A",
			},
			expected: "0.0.0.0",
		},
		{
			name: "No domain provided",
			input: struct {
				domain      string
				entryTarget string
				epType      string
			}{
				domain:      "",
				entryTarget: "www.alpha.com",
				epType:      "CNAME",
			},
			expected: "www.alpha.com",
		},
		{
			name: "Other domain without trailing dot provided",
			input: struct {
				domain      string
				entryTarget string
				epType      string
			}{
				domain:      "alpha.com",
				entryTarget: "www.beta.com",
				epType:      "CNAME",
			},
			expected: "www.beta.com",
		},
		{
			name: "Other domain with trailing dot provided",
			input: struct {
				domain      string
				entryTarget string
				epType      string
			}{
				domain:      "alpha.com",
				entryTarget: "www.beta.com.",
				epType:      "CNAME",
			},
			expected: "www.beta.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_mergeEndpointsByNameType tests mergeEndpointsByNameType().
func Test_mergeEndpointsByNameType(t *testing.T) {
	mkEndpoint := func(params [3]string) *endpoint.Endpoint {
		return &endpoint.Endpoint{
			RecordType: params[0],
			DNSName:    params[1],
			Targets:    []string{params[2]},
		}
	}

	type testCase struct {
		name        string
		input       [][3]string
		expectedLen int
	}

	run := func(t *testing.T, tc testCase) {
		input := make([]*endpoint.Endpoint, 0, len(tc.input))
		for _, r := range tc.input {
			input = append(input, mkEndpoint(r))
		}
		actual := mergeEndpointsByNameType(input)
		assert.Equal(t, len(actual), tc.expectedLen)
	}

	testCases := []testCase{
		{
			name: "1:1 endpoint",
			input: [][3]string{
				{"A", "www.alfa.com", "8.8.8.8"},
				{"A", "www.beta.com", "9.9.9.9"},
				{"A", "www.gamma.com", "1.1.1.1"},
			},
			expectedLen: 3,
		},
		{
			name: "6:4 endpoint",
			input: [][3]string{
				{"A", "www.alfa.com", "1.1.1.1"},
				{"A", "www.beta.com", "2.2.2.2"},
				{"A", "www.beta.com", "3.3.3.3"},
				{"A", "www.gamma.com", "4.4.4.4"},
				{"A", "www.gamma.com", "5.5.5.5"},
				{"A", "www.delta.com", "6.6.6.6"},
			},
			expectedLen: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_createEndpointFromRecord tests createEndpointFromRecord().
func Test_createEndpointFromRecord(t *testing.T) {
	type testCase struct {
		name     string
		input    dsdns.Record
		expected *endpoint.Endpoint
	}

	run := func(t *testing.T, tc testCase) {
		actual := createEndpointFromRecord(tc.input)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name: "top domain",
			input: dsdns.Record{
				ID:   "id_0",
				Host: "@",
				Type: dsdns.RecordTypeCNAME,
				Data: "www.alpha.com.",
				Domain: &dsdns.Domain{
					ID:     "domainIDBeta",
					Domain: "beta.com",
				},
				Ttl: 7200,
			},
			expected: &endpoint.Endpoint{
				DNSName:    "beta.com",
				RecordType: "CNAME",
				Targets:    endpoint.Targets{"www.alpha.com"},
				RecordTTL:  7200,
				Labels:     endpoint.Labels{},
			},
		},
		{
			name: "a record",
			input: dsdns.Record{
				ID:   "id_1",
				Host: "ftp",
				Type: dsdns.RecordTypeA,
				Data: "10.0.0.1",
				Domain: &dsdns.Domain{
					ID:     "domainIDBeta",
					Domain: "beta.com",
				},
				Ttl: 7200,
			},
			expected: &endpoint.Endpoint{
				DNSName:    "ftp.beta.com",
				RecordType: "A",
				Targets:    endpoint.Targets{"10.0.0.1"},
				RecordTTL:  7200,
				Labels:     endpoint.Labels{},
			},
		},
		{
			name: "cname record",
			input: dsdns.Record{
				ID:   "id_1",
				Host: "ftp",
				Type: dsdns.RecordTypeCNAME,
				Data: "www.alpha.com.",
				Domain: &dsdns.Domain{
					ID:     "domainIDBeta",
					Domain: "beta.com",
				},
				Ttl: 7200,
			},
			expected: &endpoint.Endpoint{
				DNSName:    "ftp.beta.com",
				RecordType: "CNAME",
				Targets:    endpoint.Targets{"www.alpha.com"},
				RecordTTL:  7200,
				Labels:     endpoint.Labels{},
			},
		},
		{
			name: "mx record",
			input: dsdns.Record{
				ID:   "id_1",
				Host: "@",
				Type: dsdns.RecordTypeMX,
				Data: "10 mail.alpha.com.",
				Domain: &dsdns.Domain{
					ID:     "domainIDBeta",
					Domain: "beta.com",
				},
				Ttl: 7200,
			},
			expected: &endpoint.Endpoint{
				DNSName:    "beta.com",
				RecordType: "MX",
				Targets:    endpoint.Targets{"10 mail.alpha.com"},
				RecordTTL:  7200,
				Labels:     endpoint.Labels{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_endpointsByDomainID tests endpointsByDomainID().
func Test_endpointsByDomainID(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainIDNameMapper provider.ZoneIDName
			endpoints          []*endpoint.Endpoint
		}
		expected map[string][]*endpoint.Endpoint
	}

	run := func(t *testing.T, tc testCase) {
		actual := endpointsByDomainID(tc.input.domainIDNameMapper, tc.input.endpoints)
		assert.Equal(t, actual, tc.expected)
	}

	testCases := []testCase{
		{
			name: "empty input",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				endpoints          []*endpoint.Endpoint
			}{
				domainIDNameMapper: provider.ZoneIDName{
					"domainIDAlpha": "alpha.com",
					"domainIDBeta":  "beta.com",
				},
				endpoints: []*endpoint.Endpoint{},
			},
			expected: map[string][]*endpoint.Endpoint{},
		},
		{
			name: "some input",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				endpoints          []*endpoint.Endpoint
			}{
				domainIDNameMapper: provider.ZoneIDName{
					"domainIDAlpha": "alpha.com",
					"domainIDBeta":  "beta.com",
				},
				endpoints: []*endpoint.Endpoint{
					{
						DNSName:    "www.alpha.com",
						RecordType: "A",
						Targets: endpoint.Targets{
							"127.0.0.1",
						},
					},
					{
						DNSName:    "www.beta.com",
						RecordType: "A",
						Targets: endpoint.Targets{
							"127.0.0.1",
						},
					},
				},
			},
			expected: map[string][]*endpoint.Endpoint{
				"domainIDAlpha": {
					&endpoint.Endpoint{
						DNSName:    "www.alpha.com",
						RecordType: "A",
						Targets: endpoint.Targets{
							"127.0.0.1",
						},
					},
				},
				"domainIDBeta": {
					&endpoint.Endpoint{
						DNSName:    "www.beta.com",
						RecordType: "A",
						Targets: endpoint.Targets{
							"127.0.0.1",
						},
					},
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

// Test_getMatchingDomainRecords tests getMatchingDomainRecords().
func Test_getMatchingDomainRecords(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			records    []dsdns.Record
			domainName string
			ep         *endpoint.Endpoint
		}
		expected []dsdns.Record
	}

	testCases := []testCase{
		{
			name: "no matches",
			input: struct {
				records    []dsdns.Record
				domainName string
				ep         *endpoint.Endpoint
			}{
				records: []dsdns.Record{
					{
						ID: "id1",
						Domain: &dsdns.Domain{
							ID:     "domainIDAlpha",
							Domain: "alpha.com",
						},
						Host: "www",
						Type: dsdns.RecordTypeA,
						Data: "1.1.1.1",
					},
				},
				domainName: "alpha.com",
				ep: &endpoint.Endpoint{
					DNSName:    "ftp.alpha.com",
					RecordType: endpoint.RecordTypeA,
					Targets:    endpoint.Targets{"1.1.1.1"},
				},
			},
			expected: []dsdns.Record{},
		},
		{
			name: "matches",
			input: struct {
				records    []dsdns.Record
				domainName string
				ep         *endpoint.Endpoint
			}{
				records: []dsdns.Record{
					{
						ID: "id1",
						Domain: &dsdns.Domain{
							ID:     "domainIDAlpha",
							Domain: "alpha.com",
						},
						Host: "www",
						Type: dsdns.RecordTypeA,
						Data: "1.1.1.1",
					},
				},
				domainName: "alpha.com",
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					RecordType: endpoint.RecordTypeA,
					Targets:    endpoint.Targets{"1.1.1.1"},
				},
			},
			expected: []dsdns.Record{
				{
					ID: "id1",
					Domain: &dsdns.Domain{
						ID:     "domainIDAlpha",
						Domain: "alpha.com",
					},
					Host: "www",
					Type: dsdns.RecordTypeA,
					Data: "1.1.1.1",
				},
			},
		},
		{
			name: "matches with warning",
			input: struct {
				records    []dsdns.Record
				domainName string
				ep         *endpoint.Endpoint
			}{
				records: []dsdns.Record{
					{
						ID: "id1",
						Domain: &dsdns.Domain{
							ID:     "domainIDAlpha",
							Domain: "alpha.com",
						},
						Host: "www",
						Type: dsdns.RecordTypeA,
						Data: "1.1.1.1",
					},
				},
				domainName: "alpha.com",
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					RecordType: endpoint.RecordTypeA,
					Targets:    endpoint.Targets{"1.1.1.1"},
					ProviderSpecific: endpoint.ProviderSpecific{
						endpoint.ProviderSpecificProperty{
							Name:  "webhook/domeneshop-label-environment",
							Value: "test",
						},
					},
				},
			},
			expected: []dsdns.Record{
				{
					ID: "id1",
					Domain: &dsdns.Domain{
						ID:     "domainIDAlpha",
						Domain: "alpha.com",
					},
					Host: "www",
					Type: dsdns.RecordTypeA,
					Data: "1.1.1.1",
				},
			},
		},
	}

	run := func(t *testing.T, tc testCase) {
		actual := getMatchingDomainRecords(tc.input.records, tc.input.domainName, tc.input.ep)
		assert.ElementsMatch(t, actual, tc.expected)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_getEndpointTTL tests getEndpointTTL().
func Test_getEndpointTTL(t *testing.T) {
	type testCase struct {
		name     string
		input    *endpoint.Endpoint
		expected *int
	}
	configuredTTL := 7200

	run := func(t *testing.T, tc testCase) {
		actualTTL := getEndpointTTL(tc.input)
		assert.EqualValues(t, tc.expected, actualTTL)
	}

	testCases := []testCase{
		{
			name: "TTL configured",
			input: &endpoint.Endpoint{
				RecordTTL: 7200,
			},
			expected: &configuredTTL,
		},
		{
			name: "TTL not configured",
			input: &endpoint.Endpoint{
				RecordTTL: -1,
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func Test_getEndpointLogFields(t *testing.T) {
	type testCase struct {
		name     string
		input    *endpoint.Endpoint
		expected log.Fields
	}

	run := func(t *testing.T, tc testCase) {
		actual := getEndpointLogFields(tc.input)
		assert.Equal(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name: "single target endpoint",
			input: &endpoint.Endpoint{
				DNSName:    "www.alpha.com",
				RecordType: "A",
				Targets:    endpoint.Targets{"1.1.1.1"},
				RecordTTL:  7200,
			},
			expected: log.Fields{
				"DNSName":    "www.alpha.com",
				"RecordType": "A",
				"Targets":    "1.1.1.1",
				"TTL":        7200,
			},
		},
		{
			name: "multiple target endpoint",
			input: &endpoint.Endpoint{
				DNSName:    "www.alpha.com",
				RecordType: "A",
				Targets:    endpoint.Targets{"1.1.1.1", "2.2.2.2"},
				RecordTTL:  7200,
			},
			expected: log.Fields{
				"DNSName":    "www.alpha.com",
				"RecordType": "A",
				"Targets":    "1.1.1.1;2.2.2.2",
				"TTL":        7200,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
