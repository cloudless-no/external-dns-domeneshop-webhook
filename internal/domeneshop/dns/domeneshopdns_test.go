/*
 * DomeneshopDNS - Common test routines for the domeneshop package.
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
	"testing"

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
	"github.com/stretchr/testify/assert"
)

// testTTL is a test ttl.
var testTTL = 7200

// domainsResponse simulates a response that returns a list of domains.
type domainsResponse struct {
	domains []*dsdns.Domain
	resp  *dsdns.Response
	err   error
}

// recordsResponse simulates a response that returns a list of records.
type recordsResponse struct {
	records []*dsdns.Record
	resp    *dsdns.Response
	err     error
}

// recordResponse simulates a response that returns a single record.
type recordResponse struct {
	record *dsdns.Record
	resp   *dsdns.Response
	err    error
}

// deleteResponse simulates a response to a record deletion request.
type deleteResponse struct {
	resp *dsdns.Response
	err  error
}

// mockClientState keeps track of which methods were called.
type mockClientState struct {
	GetDomainsCalled     bool
	GetRecordsCalled   bool
	CreateRecordCalled bool
	UpdateRecordCalled bool
	DeleteRecordCalled bool
}

// mockClient represents the mock client used to simulate calls to the DNS API.
type mockClient struct {
	getDomains            domainsResponse
	getRecords          recordsResponse
	createRecord        recordResponse
	updateRecord        recordResponse
	deleteRecord        deleteResponse
	filterRecordsByDomain bool
	state               mockClientState
}

// GetState returns the internal state
func (m mockClient) GetState() mockClientState {
	return m.state
}

// GetDomains simulates a request to get a list of domains.
func (m *mockClient) GetDomains(ctx context.Context) ([]*dsdns.Domain, *dsdns.Response, error) {
	r := m.getDomains
	m.state.GetDomainsCalled = true
	return r.domains, r.resp, r.err
}

// filterRecordsByDomain filters the records, returning only those for the selected domain.
func filterRecordsByDomain(r recordsResponse, opts dsdns.RecordListOpts) []*dsdns.Record {
	records := make([]*dsdns.Record, 0)
	for _, rec := range r.records {
		if rec != nil && rec.Domain.ID == opts.DomainID {
			records = append(records, rec)
		}
	}
	return records
}

// GetRecords simulates a request to get a list of records for a given domain.
func (m *mockClient) GetRecords(ctx context.Context, opts dsdns.RecordListOpts) ([]*dsdns.Record, *dsdns.Response, error) {
	r := m.getRecords
	m.state.GetRecordsCalled = true
	if r.err != nil {
		return nil, r.resp, r.err
	}
	var records []*dsdns.Record
	if m.filterRecordsByDomain {
		records = filterRecordsByDomain(r, opts)
	} else {
		records = r.records
	}
	return records, r.resp, r.err
}

// CreateRecord simulates a request to create a DNS record.
func (m *mockClient) CreateRecord(ctx context.Context, opts dsdns.RecordCreateOpts) (*dsdns.Record, *dsdns.Response, error) {
	r := m.createRecord
	m.state.CreateRecordCalled = true
	return r.record, r.resp, r.err
}

// UpdateRecord simulates a request to update a DNS record.
func (m *mockClient) UpdateRecord(ctx context.Context, record *dsdns.Record, opts dsdns.RecordUpdateOpts) (*dsdns.Record, *dsdns.Response, error) {
	r := m.updateRecord
	m.state.UpdateRecordCalled = true
	return r.record, r.resp, r.err
}

// DeleteRecord simulates a request to delete a DNS record.
func (m *mockClient) DeleteRecord(ctx context.Context, record *dsdns.Record, opts dsdns.RecordUpdateOpts) (*dsdns.Response, error) {
	r := m.deleteRecord
	m.state.DeleteRecordCalled = true
	return r.resp, r.err
}

// assertError checks if an error is thrown when expected. It also returns
// true if an error was expected and false it wasn't.
func assertError(t *testing.T, expected, actual error) bool {
	var expError bool
	if expected == nil {
		assert.Nil(t, actual)
		expError = false
	} else {
		assert.EqualError(t, actual, expected.Error())
		expError = true
	}
	return expError
}

func Test_NewDomeneshopDNS(t *testing.T) {
	type testCase struct {
		name     string
		token    string
		secret    string
		expected struct {
			clientPresent bool
			err           error
		}
	}

	run := func(t *testing.T, tc testCase) {
		exp := tc.expected
		client, err := NewDomeneshopDNS(tc.token, tc.secret)
		if !assertError(t, exp.err, err) {
			if exp.clientPresent {
				assert.NotNil(t, client)
				assert.NotNil(t, client.client)
			} else {
				assert.Nil(t, client)
			}
		}
	}

	testCases := []testCase{
		{
			name:  "empty api key",
			token: "",
			secret: "",
			expected: struct {
				clientPresent bool
				err           error
			}{
				err: errors.New("nil Token provided"),
			},
		},
		{
			name:  "some api key",
			token: "TEST_TOKEN",
			secret: "TEST_SECRET",
			expected: struct {
				clientPresent bool
				err           error
			}{
				clientPresent: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
