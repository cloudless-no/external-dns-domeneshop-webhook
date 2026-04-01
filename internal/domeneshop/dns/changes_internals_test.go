/*
 * Changes Internals - unit tests
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
	"testing"

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type changeType interface {
	GetLogFields() log.Fields
}

var defaultTTL = -1

func Test_GetLogFields(t *testing.T) {
	type testCase struct {
		name     string
		object   changeType
		expected log.Fields
	}

	run := func(t *testing.T, tc testCase) {
		actual := tc.object.GetLogFields()
		assert.Equal(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name: "domeneshopChangeCreate",
			object: &domeneshopChangeCreate{
				DomainID: "testDomainID",
				Options: &dsdns.RecordCreateOpts{
					Host:  "testName",
					Ttl:   &defaultTTL,
					Data: "testValue",
					Type:  "CNAME",
					Domain: &dsdns.Domain{
						ID:   "testDomainID",
						Domain: "testDomainName",
					},
				},
			},
			expected: log.Fields{
				"domain":     "testDomainName",
				"domainID":     "testDomainID",
				"dnsName":    "testName",
				"recordType": "CNAME",
				"value":      "testValue",
				"ttl":        defaultTTL,
			},
		},
		{
			name: "domeneshopChangeUpdate",
			object: &domeneshopChangeUpdate{
				DomainID: "testDomainID",
				Record: dsdns.Record{
					ID: "recordID",
					Domain: &dsdns.Domain{
						ID:   "testDomainID",
						Domain: "testDomainName",
					},
					Host:  "recordOldName",
					Data: "recordOldValue",
				},
				Options: &dsdns.RecordUpdateOpts{
					Host:  "testNewName",
					Ttl:   &defaultTTL,
					Data: "testNewValue",
					Type:  "CNAME",
					Domain: &dsdns.Domain{
						ID:   "testDomainID",
						Domain: "testDomainName",
					},
				},
			},
			expected: log.Fields{
				"domain":      "testDomainName",
				"domainID":      "testDomainID",
				"recordID":    "recordID",
				"*dnsName":    "testNewName",
				"*recordType": "CNAME",
				"*value":      "testNewValue",
				"*ttl":        defaultTTL,
			},
		},
		{
			name: "domeneshopChangeDelete",
			object: &domeneshopChangeDelete{
				DomainID: "testDomainID",
				Record: dsdns.Record{
					ID: "recordID",
					Domain: &dsdns.Domain{
						ID:   "testDomainID",
						Domain: "testDomainName",
					},
					Type:  "CNAME",
					Host:  "recordName",
					Data: "recordValue",
				},
			},
			expected: log.Fields{
				"domain":     "testDomainName",
				"domainID":     "testDomainID",
				"dnsName":    "recordName",
				"recordType": "CNAME",
				"value":      "recordValue",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
