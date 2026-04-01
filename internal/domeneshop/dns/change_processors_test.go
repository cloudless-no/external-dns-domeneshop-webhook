/*
 * Change processors - unit tests.
 *
 * Copyright 2026 Marco Confalonieri & Kim Engebretsen.
 * Copyright 2017 The Kubernetes Authors.
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
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

var testDomainIDMapper = provider.ZoneIDName{
	"domainIDAlpha": "alpha.com",
	"domainIDBeta":  "beta.com",
}

func assertEqualChanges(t *testing.T, expected, actual domeneshopChanges) {
	assert.Equal(t, expected.dryRun, actual.dryRun)
	assert.ElementsMatch(t, expected.creates, actual.creates)
	assert.ElementsMatch(t, expected.updates, actual.updates)
	assert.ElementsMatch(t, expected.deletes, actual.deletes)
}

// Test_adjustCNAMETarget tests adjustCNAMETarget()
func Test_adjustCNAMETarget(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domain string
			target string
		}
		expected string
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := adjustCNAMETarget(inp.domain, inp.target)
		assert.Equal(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name: "target matches domain",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "www.alpha.com",
			},
			expected: "www",
		},
		{
			name: "target matches domain with dot",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "www.alpha.com.",
			},
			expected: "www",
		},
		{
			name: "target without dot does not match domain",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "www.beta.com",
			},
			expected: "www.beta.com.",
		},
		{
			name: "target with dot does not match domain",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "www.beta.com.",
			},
			expected: "www.beta.com.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_adjustMXTarget tests adjustMXTarget()
func Test_adjustMXTarget(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domain string
			target string
		}
		expected string
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := adjustMXTarget(inp.domain, inp.target)
		assert.Equal(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name: "MX target with local hostname",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "10 mail.alpha.com",
			},
			expected: "10 mail",
		},
		{
			name: "MX target with local hostname and trailing dot",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "10 mail.alpha.com.",
			},
			expected: "10 mail",
		},
		{
			name: "MX target with external hostname",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "10 mail.beta.com",
			},
			expected: "10 mail.beta.com.",
		},
		{
			name: "MX target with external hostname and trailing dot",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "10 mail.beta.com.",
			},
			expected: "10 mail.beta.com.",
		},
		{
			name: "MX target with apex record",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "10 alpha.com",
			},
			expected: "10 @",
		},
		{
			name: "MX target with apex record and trailing dot",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "10 alpha.com.",
			},
			expected: "10 @",
		},
		{
			name: "MX target with deep local subdomain",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "10 smtp.mail.alpha.com",
			},
			expected: "10 smtp.mail",
		},
		{
			name: "MX target with invalid format (no space)",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "mail.alpha.com",
			},
			expected: "mail.alpha.com", // returned unchanged
		},
		{
			name: "MX target with non-numeric priority",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "high mail.alpha.com",
			},
			expected: "high mail.alpha.com", // returned unchanged
		},
		{
			name: "MX target with priority leading zeros",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "010 mail.alpha.com",
			},
			expected: "010 mail", // strconv.Atoi accepts leading zeros
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func Test_adjustTarget(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domain     string
			recordType string
			target     string
		}
		expected string
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := adjustTarget(inp.domain, inp.recordType, inp.target)
		assert.Equal(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name: "cname target",
			input: struct {
				domain     string
				recordType string
				target     string
			}{
				domain:     "alpha.com",
				recordType: "CNAME",
				target:     "www.alpha.com",
			},
			expected: "www",
		},
		{
			name: "mx target",
			input: struct {
				domain     string
				recordType string
				target     string
			}{
				domain:     "alpha.com",
				recordType: "MX",
				target:     "10 mail.alpha.com.",
			},
			expected: "10 mail",
		},
		{
			name: "other target",
			input: struct {
				domain     string
				recordType string
				target     string
			}{
				domain:     "alpha.com",
				recordType: "A",
				target:     "10.0.0.1",
			},
			expected: "10.0.0.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processCreateActionsByDomain tests processCreateActionsByDomain().
func Test_processCreateActionsByDomain(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainID    string
			domainName  string
			records   []dsdns.Record
			endpoints []*endpoint.Endpoint
		}
		expectedChanges domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		changes := domeneshopChanges{}
		processCreateActionsByDomain(inp.domainID, inp.domainName, inp.records,
			inp.endpoints, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "record already created",
			input: struct {
				domainID    string
				domainName  string
				records   []dsdns.Record
				endpoints []*endpoint.Endpoint
			}{
				domainID:   "domainIDAlpha",
				domainName: "alpha.com",
				records: []dsdns.Record{
					{
						Type:  "A",
						Host:  "www",
						Data: "127.0.0.1",
						Ttl:   7200,
					},
				},
				endpoints: []*endpoint.Endpoint{
					{
						DNSName:    "www.alpha.com",
						Targets:    endpoint.Targets{"127.0.0.1"},
						RecordType: "A",
						RecordTTL:  7200,
					},
				},
			},
			expectedChanges: domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host:  "www",
							Ttl:   &testTTL,
							Type:  "A",
							Data: "127.0.0.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
						},
					},
				},
			},
		},
		{
			name: "new record created",
			input: struct {
				domainID    string
				domainName  string
				records   []dsdns.Record
				endpoints []*endpoint.Endpoint
			}{
				domainID:   "domainIDAlpha",
				domainName: "alpha.com",
				records: []dsdns.Record{
					{
						Type:  "A",
						Host:  "ftp",
						Data: "127.0.0.1",
						Ttl:   7200,
					},
				},
				endpoints: []*endpoint.Endpoint{
					{
						DNSName:    "www.alpha.com",
						Targets:    endpoint.Targets{"127.0.0.1"},
						RecordType: "A",
						RecordTTL:  7200,
					},
				},
			},
			expectedChanges: domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host:  "www",
							Ttl:   &testTTL,
							Type:  "A",
							Data: "127.0.0.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
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

// Test_processCreateActions tests processCreateActions().
func Test_processCreateActions(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainIDNameMapper provider.ZoneIDName
			recordsByDomainID  map[string][]dsdns.Record
			createsByDomainID  map[string][]*endpoint.Endpoint
		}
		expectedChanges domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		changes := domeneshopChanges{}
		processCreateActions(inp.domainIDNameMapper, inp.recordsByDomainID,
			inp.createsByDomainID, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "empty changeset",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				recordsByDomainID  map[string][]dsdns.Record
				createsByDomainID  map[string][]*endpoint.Endpoint
			}{
				domainIDNameMapper: testDomainIDMapper,
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
						dsdns.Record{
							Type:  "A",
							Host:  "www",
							Data: "127.0.0.1",
						},
					},
				},
				createsByDomainID: map[string][]*endpoint.Endpoint{},
			},
		},
		{
			name: "empty changeset with key present",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				recordsByDomainID  map[string][]dsdns.Record
				createsByDomainID  map[string][]*endpoint.Endpoint
			}{
				domainIDNameMapper: testDomainIDMapper,
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
						dsdns.Record{
							Type:  "A",
							Host:  "www",
							Data: "127.0.0.1",
						},
					},
				},
				createsByDomainID: map[string][]*endpoint.Endpoint{
					"domainIDAlpha": {},
				},
			},
		},
		{
			name: "record already created",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				recordsByDomainID  map[string][]dsdns.Record
				createsByDomainID  map[string][]*endpoint.Endpoint
			}{
				domainIDNameMapper: testDomainIDMapper,
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
						dsdns.Record{
							Type:  "A",
							Host:  "www",
							Data: "127.0.0.1",
							Ttl:   7200,
						},
					},
				},
				createsByDomainID: map[string][]*endpoint.Endpoint{
					"domainIDAlpha": {
						&endpoint.Endpoint{
							DNSName:    "www.alpha.com",
							Targets:    endpoint.Targets{"127.0.0.1"},
							RecordType: "A",
							RecordTTL:  7200,
						},
					},
				},
			},
			expectedChanges: domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host:  "www",
							Ttl:   &testTTL,
							Type:  "A",
							Data: "127.0.0.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
						},
					},
				},
			},
		},
		{
			name: "new record created",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				recordsByDomainID  map[string][]dsdns.Record
				createsByDomainID  map[string][]*endpoint.Endpoint
			}{
				domainIDNameMapper: testDomainIDMapper,
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
						dsdns.Record{
							Type:  "A",
							Host:  "ftp",
							Data: "127.0.0.1",
							Ttl:   7200,
						},
					},
				},
				createsByDomainID: map[string][]*endpoint.Endpoint{
					"domainIDAlpha": {
						&endpoint.Endpoint{
							DNSName:    "www.alpha.com",
							Targets:    endpoint.Targets{"127.0.0.1"},
							RecordType: "A",
							RecordTTL:  7200,
						},
					},
				},
			},
			expectedChanges: domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host:  "www",
							Ttl:   &testTTL,
							Type:  "A",
							Data: "127.0.0.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
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

// Test_processUpdateEndpoint tests processUpdateEndpoint().
func Test_processUpdateEndpoint(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainID                  string
			domainName                string
			matchingRecordsByTarget map[string]dsdns.Record
			ep                      *endpoint.Endpoint
		}
		expectedChanges domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := domeneshopChanges{}
		inp := tc.input
		processUpdateEndpoint(inp.domainID, inp.domainName, inp.matchingRecordsByTarget,
			inp.ep, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "name changed",
			input: struct {
				domainID                  string
				domainName                string
				matchingRecordsByTarget map[string]dsdns.Record
				ep                      *endpoint.Endpoint
			}{
				domainID:   "domainIDAlpha",
				domainName: "alpha.com",
				matchingRecordsByTarget: map[string]dsdns.Record{
					"1.1.1.1": {
						ID:   "id_1",
						Type: dsdns.RecordTypeA,
						Host: "www",
						Domain: &dsdns.Domain{
							ID:   "domainIDAlpha",
							Domain: "alpha.com",
						},
						Data: "1.1.1.1",
						Ttl:   -1,
					},
				},
				ep: &endpoint.Endpoint{
					DNSName:    "ftp.alpha.com",
					RecordType: "A",
					Targets:    []string{"1.1.1.1"},
					RecordTTL:  -1,
				},
			},
			expectedChanges: domeneshopChanges{
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:   "id_1",
							Type: dsdns.RecordTypeA,
							Host: "www",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Data: "1.1.1.1",
							Ttl:   -1,
						},
						Options: &dsdns.RecordUpdateOpts{
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Ttl:   nil,
							Data: "1.1.1.1",
						},
					},
				},
			},
		},
		{
			name: "TTL changed",
			input: struct {
				domainID                  string
				domainName                string
				matchingRecordsByTarget map[string]dsdns.Record
				ep                      *endpoint.Endpoint
			}{
				domainID:   "domainIDAlpha",
				domainName: "alpha.com",
				matchingRecordsByTarget: map[string]dsdns.Record{
					"1.1.1.1": {
						ID:   "id_1",
						Type: dsdns.RecordTypeA,
						Host: "www",
						Domain: &dsdns.Domain{
							ID:   "domainIDAlpha",
							Domain: "alpha.com",
						},
						Data: "1.1.1.1",
						Ttl:   -1,
					},
				},
				ep: &endpoint.Endpoint{
					DNSName:    "ftp.alpha.com",
					RecordType: "A",
					Targets:    []string{"1.1.1.1"},
					RecordTTL:  7200,
				},
			},
			expectedChanges: domeneshopChanges{
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:   "id_1",
							Type: dsdns.RecordTypeA,
							Host: "www",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Data: "1.1.1.1",
							Ttl:   -1,
						},
						Options: &dsdns.RecordUpdateOpts{
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Ttl:   &testTTL,
							Data: "1.1.1.1",
						},
					},
				},
			},
		},
		{
			name: "target changed",
			input: struct {
				domainID                  string
				domainName                string
				matchingRecordsByTarget map[string]dsdns.Record
				ep                      *endpoint.Endpoint
			}{
				domainID:   "domainIDAlpha",
				domainName: "alpha.com",
				matchingRecordsByTarget: map[string]dsdns.Record{
					"1.1.1.1": {
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
				},
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					RecordType: "A",
					Targets:    []string{"2.2.2.2"},
					RecordTTL:  -1,
				},
			},
			expectedChanges: domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host:  "www",
							Ttl:   nil,
							Type:  dsdns.RecordTypeA,
							Data: "2.2.2.2",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
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

// Test_cleanupRemainingTargets tests cleanupRemainingTargets().
func Test_cleanupRemainingTargets(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainID                  string
			matchingRecordsByTarget map[string]dsdns.Record
		}
		expectedChanges domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := domeneshopChanges{}
		inp := tc.input
		cleanupRemainingTargets(inp.domainID, inp.matchingRecordsByTarget,
			&changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "no deletes",
			input: struct {
				domainID                  string
				matchingRecordsByTarget map[string]dsdns.Record
			}{
				domainID:                  "domainIDAlpha",
				matchingRecordsByTarget: map[string]dsdns.Record{},
			},
		},
		{
			name: "delete",
			input: struct {
				domainID                  string
				matchingRecordsByTarget map[string]dsdns.Record
			}{
				domainID: "domainIDAlpha",
				matchingRecordsByTarget: map[string]dsdns.Record{
					"1.1.1.1": {
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
				},
			},
			expectedChanges: domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
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

// Test_getMatchingRecordsByTarget tests getMatchingRecordsByTarget().
func Test_getMatchingRecordsByTarget(t *testing.T) {
	type testCase struct {
		name     string
		input    []dsdns.Record
		expected map[string]dsdns.Record
	}

	run := func(t *testing.T, tc testCase) {
		actual := getMatchingRecordsByTarget(tc.input)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name:     "empty array",
			input:    []dsdns.Record{},
			expected: map[string]dsdns.Record{},
		},
		{
			name: "some values",
			input: []dsdns.Record{
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
			},
			expected: map[string]dsdns.Record{
				"1.1.1.1": {
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
				"2.2.2.2": {
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
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processUpdateActionsByDomain tests processUpdateActionsByDomain().
func Test_processUpdateActionsByDomain(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainID    string
			domainName  string
			records   []dsdns.Record
			endpoints []*endpoint.Endpoint
		}
		expectedChanges domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := domeneshopChanges{}
		inp := tc.input
		processUpdateActionsByDomain(inp.domainID, inp.domainName, inp.records,
			inp.endpoints, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "empty changeset",
			input: struct {
				domainID    string
				domainName  string
				records   []dsdns.Record
				endpoints []*endpoint.Endpoint
			}{
				domainID:   "domainIDAlpha",
				domainName: "alpha.com",
				records: []dsdns.Record{
					{
						ID:   "id_1",
						Host: "www",
						Domain: &dsdns.Domain{
							ID:   "domainIDAlpha",
							Domain: "alpha.com",
						},
						Data: "1.1.1.1",
						Ttl:   -1,
					},
				},
				endpoints: []*endpoint.Endpoint{},
			},
			expectedChanges: domeneshopChanges{},
		},
		{
			name: "mixed changeset",
			input: struct {
				domainID    string
				domainName  string
				records   []dsdns.Record
				endpoints []*endpoint.Endpoint
			}{
				domainID:   "domainIDAlpha",
				domainName: "alpha.com",
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
				},
				endpoints: []*endpoint.Endpoint{
					{
						DNSName:    "www.alpha.com",
						RecordType: "A",
						Targets:    []string{"3.3.3.3"},
						RecordTTL:  -1,
					},
					{
						DNSName:    "ftp.alpha.com",
						RecordType: "A",
						Targets:    []string{"2.2.2.2"},
						RecordTTL:  7200,
					},
				},
			},
			expectedChanges: domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host:  "www",
							Ttl:   nil,
							Type:  dsdns.RecordTypeA,
							Data: "3.3.3.3",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
						},
					},
				},
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
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
					},
				},
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
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
						Options: &dsdns.RecordUpdateOpts{
							Host: "ftp",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Type:  dsdns.RecordTypeA,
							Data: "2.2.2.2",
							Ttl:   &testTTL,
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

// Test_processUpdateActions tests processUpdateActions().
func Test_processUpdateActions(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainIDNameMapper provider.ZoneIDName
			recordsByDomainID  map[string][]dsdns.Record
			updatesByDomainID  map[string][]*endpoint.Endpoint
		}
		expectedChanges domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := domeneshopChanges{}
		inp := tc.input
		processUpdateActions(inp.domainIDNameMapper, inp.recordsByDomainID,
			inp.updatesByDomainID, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "empty changeset",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				recordsByDomainID  map[string][]dsdns.Record
				updatesByDomainID  map[string][]*endpoint.Endpoint
			}{
				domainIDNameMapper: provider.ZoneIDName{
					"domainIDAlpha": "alpha.com",
					"domainIDBeta":  "beta.com",
				},
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
						dsdns.Record{
							ID:   "id_1",
							Host: "www",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Data: "1.1.1.1",
							Ttl:   -1,
						},
					},
					"domainIDBeta": {
						dsdns.Record{
							ID:   "id_2",
							Host: "ftp",
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "2.2.2.2",
							Ttl:   -1,
						},
					},
				},
				updatesByDomainID: map[string][]*endpoint.Endpoint{},
			},
			expectedChanges: domeneshopChanges{},
		},
		{
			name: "empty changeset with key present",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				recordsByDomainID  map[string][]dsdns.Record
				updatesByDomainID  map[string][]*endpoint.Endpoint
			}{
				domainIDNameMapper: provider.ZoneIDName{
					"domainIDAlpha": "alpha.com",
					"domainIDBeta":  "beta.com",
				},
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
						dsdns.Record{
							ID:   "id_1",
							Host: "www",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Data: "1.1.1.1",
							Ttl:   -1,
						},
					},
					"domainIDBeta": {
						dsdns.Record{
							ID:   "id_2",
							Host: "ftp",
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "2.2.2.2",
							Ttl:   -1,
						},
					},
				},
				updatesByDomainID: map[string][]*endpoint.Endpoint{
					"domainIDAlpha": {},
					"domainIDBeta":  {},
				},
			},
		},
		{
			name: "mixed changeset",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				recordsByDomainID  map[string][]dsdns.Record
				updatesByDomainID  map[string][]*endpoint.Endpoint
			}{
				domainIDNameMapper: provider.ZoneIDName{
					"domainIDAlpha": "alpha.com",
					"domainIDBeta":  "beta.com",
				},
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
						dsdns.Record{
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
					},
					"domainIDBeta": {
						dsdns.Record{
							ID:   "id_2",
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "2.2.2.2",
							Ttl:   -1,
						},
					},
				},
				updatesByDomainID: map[string][]*endpoint.Endpoint{
					"domainIDAlpha": {
						&endpoint.Endpoint{
							DNSName:    "www.alpha.com",
							RecordType: "A",
							Targets:    []string{"3.3.3.3"},
							RecordTTL:  -1,
						},
					},
					"domainIDBeta": {
						&endpoint.Endpoint{
							DNSName:    "ftp.beta.com",
							RecordType: "A",
							Targets:    []string{"2.2.2.2"},
							RecordTTL:  7200,
						},
					},
				},
			},
			expectedChanges: domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host:  "www",
							Type:  dsdns.RecordTypeA,
							Data: "3.3.3.3",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Ttl: nil,
						},
					},
				},
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
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
					},
				},
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDBeta",
						Record: dsdns.Record{
							ID:   "id_2",
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "2.2.2.2",
							Ttl:   -1,
						},
						Options: &dsdns.RecordUpdateOpts{
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "2.2.2.2",
							Ttl:   &testTTL,
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

func Test_targetsMatch(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			record dsdns.Record
			ep     *endpoint.Endpoint
		}
		expected bool
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := targetsMatch(inp.record, inp.ep)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name: "record does not matches",
			input: struct {
				record dsdns.Record
				ep     *endpoint.Endpoint
			}{
				record: dsdns.Record{
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
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					Targets:    endpoint.Targets{"7.7.7.7"},
					RecordType: "A",
					RecordTTL:  -1,
				},
			},
			expected: false,
		},
		{
			name: "record matches",
			input: struct {
				record dsdns.Record
				ep     *endpoint.Endpoint
			}{
				record: dsdns.Record{
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
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					Targets:    endpoint.Targets{"1.1.1.1"},
					RecordType: "A",
					RecordTTL:  -1,
				},
			},
			expected: true,
		},
		{
			name: "cname special matching",
			input: struct {
				record dsdns.Record
				ep     *endpoint.Endpoint
			}{
				record: dsdns.Record{
					ID:   "id_2",
					Host: "ftp",
					Type: dsdns.RecordTypeCNAME,
					Domain: &dsdns.Domain{
						ID:   "domainIDAlpha",
						Domain: "alpha.com",
					},
					Data: "www.beta.com.",
					Ttl:   -1,
				},
				ep: &endpoint.Endpoint{
					DNSName:    "ftp.alpha.com",
					Targets:    endpoint.Targets{"www.beta.com"},
					RecordType: "CNAME",
					RecordTTL:  -1,
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processDeleteActionsByEndpoint tests processDeleteActionsByEndpoint().
func Test_processDeleteActionsByEndpoint(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainID          string
			matchingRecords []dsdns.Record
			ep              *endpoint.Endpoint
		}
		expectedChanges domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := domeneshopChanges{}
		inp := tc.input
		processDeleteActionsByEndpoint(inp.domainID, inp.matchingRecords,
			inp.ep, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "no matching records",
			input: struct {
				domainID          string
				matchingRecords []dsdns.Record
				ep              *endpoint.Endpoint
			}{
				domainID:          "domainIDAlpha",
				matchingRecords: []dsdns.Record{},
				ep: &endpoint.Endpoint{
					DNSName:    "ccx.alpha.com",
					Targets:    endpoint.Targets{"7.7.7.7"},
					RecordType: "A",
					RecordTTL:  7200,
				},
			},
			expectedChanges: domeneshopChanges{},
		},
		{
			name: "one matching record",
			input: struct {
				domainID          string
				matchingRecords []dsdns.Record
				ep              *endpoint.Endpoint
			}{
				domainID: "domainIDAlpha",
				matchingRecords: []dsdns.Record{
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
						Host: "www",
						Type: dsdns.RecordTypeA,
						Domain: &dsdns.Domain{
							ID:   "domainIDAlpha",
							Domain: "alpha.com",
						},
						Data: "2.2.2.2",
						Ttl:   -1,
					},
				},
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					Targets:    endpoint.Targets{"1.1.1.1"},
					RecordType: "A",
					RecordTTL:  -1,
				},
			},
			expectedChanges: domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
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
					},
				},
			},
		},
		{
			name: "cname special matching",
			input: struct {
				domainID          string
				matchingRecords []dsdns.Record
				ep              *endpoint.Endpoint
			}{
				domainID: "domainIDAlpha",
				matchingRecords: []dsdns.Record{
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
						Data: "www.beta.com.",
						Ttl:   -1,
					},
				},
				ep: &endpoint.Endpoint{
					DNSName:    "ftp.alpha.com",
					Targets:    endpoint.Targets{"www.beta.com"},
					RecordType: "CNAME",
					RecordTTL:  -1,
				},
			},
			expectedChanges: domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:   "id_2",
							Host: "ftp",
							Type: dsdns.RecordTypeCNAME,
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Data: "www.beta.com.",
							Ttl:   -1,
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

// Test_processDeleteActions tests processDeleteActions().
func Test_processDeleteActions(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domainIDNameMapper provider.ZoneIDName
			recordsByDomainID  map[string][]dsdns.Record
			deletesByDomainID  map[string][]*endpoint.Endpoint
		}
		expectedChanges domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := domeneshopChanges{}
		inp := tc.input
		processDeleteActions(inp.domainIDNameMapper, inp.recordsByDomainID,
			inp.deletesByDomainID, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "No deletes created",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				recordsByDomainID  map[string][]dsdns.Record
				deletesByDomainID  map[string][]*endpoint.Endpoint
			}{
				domainIDNameMapper: provider.ZoneIDName{
					"domainIDAlpha": "alpha.com",
					"domainIDBeta":  "beta.com",
				},
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
						dsdns.Record{
							ID:   "id_1",
							Host: "www",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Data: "1.1.1.1",
							Ttl:   -1,
						},
					},
					"domainIDBeta": {
						dsdns.Record{
							ID:   "id_2",
							Host: "ftp",
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "2.2.2.2",
							Ttl:   -1,
						},
					},
				},
				deletesByDomainID: map[string][]*endpoint.Endpoint{
					"domainIDAlpha": {
						&endpoint.Endpoint{
							DNSName:    "ccx.alpha.com",
							Targets:    endpoint.Targets{"7.7.7.7"},
							RecordType: "A",
							RecordTTL:  7200,
						},
					},
				},
			},
			expectedChanges: domeneshopChanges{},
		},
		{
			name: "deletes performed",
			input: struct {
				domainIDNameMapper provider.ZoneIDName
				recordsByDomainID  map[string][]dsdns.Record
				deletesByDomainID  map[string][]*endpoint.Endpoint
			}{
				domainIDNameMapper: provider.ZoneIDName{
					"domainIDAlpha": "alpha.com",
					"domainIDBeta":  "beta.com",
				},
				recordsByDomainID: map[string][]dsdns.Record{
					"domainIDAlpha": {
						dsdns.Record{
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
					},
					"domainIDBeta": {
						dsdns.Record{
							ID:   "id_2",
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "2.2.2.2",
							Ttl:   -1,
						},
						dsdns.Record{
							ID:   "id_3",
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "4.4.4.4",
							Ttl:   -1,
						},
					},
				},
				deletesByDomainID: map[string][]*endpoint.Endpoint{
					"domainIDAlpha": {
						&endpoint.Endpoint{
							DNSName:    "www.alpha.com",
							Targets:    endpoint.Targets{"1.1.1.1"},
							RecordType: "A",
							RecordTTL:  -1,
						},
					},
					"domainIDBeta": {
						&endpoint.Endpoint{
							DNSName:    "ftp.beta.com",
							Targets:    endpoint.Targets{"2.2.2.2", "4.4.4.4"},
							RecordType: "A",
							RecordTTL:  -1,
						},
					},
				},
			},
			expectedChanges: domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
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
					},
					{
						DomainID: "domainIDBeta",
						Record: dsdns.Record{
							ID:   "id_2",
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "2.2.2.2",
							Ttl:   -1,
						},
					},
					{
						DomainID: "domainIDBeta",
						Record: dsdns.Record{
							ID:   "id_3",
							Host: "ftp",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDBeta",
								Domain: "beta.com",
							},
							Data: "4.4.4.4",
							Ttl:   -1,
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
