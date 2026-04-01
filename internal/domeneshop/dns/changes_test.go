/*
 * Changes - unit tests.
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

// Test_domeneshopChanges_empty tests domeneshopChanges.empty().
func Test_domeneshopChanges_empty(t *testing.T) {
	type testCase struct {
		name     string
		changes  domeneshopChanges
		expected bool
	}

	run := func(t *testing.T, tc testCase) {
		actual := tc.changes.empty()
		assert.Equal(t, actual, tc.expected)
	}

	testCases := []testCase{
		{
			name:     "Empty",
			changes:  domeneshopChanges{},
			expected: true,
		},
		{
			name: "Creations",
			changes: domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID:  "alphaDomainID",
						Options: &dsdns.RecordCreateOpts{},
					},
				},
			},
		},
		{
			name: "Updates",
			changes: domeneshopChanges{
				updates: []*domeneshopChangeUpdate{
					{
						DomainID:  "alphaDomainID",
						Record:  dsdns.Record{},
						Options: &dsdns.RecordUpdateOpts{},
					},
				},
			},
		},
		{
			name: "Deletions",
			changes: domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "alphaDomainID",
						Record: dsdns.Record{},
					},
				},
			},
		},
		{
			name: "All",
			changes: domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID:  "alphaDomainID",
						Options: &dsdns.RecordCreateOpts{},
					},
				},
				updates: []*domeneshopChangeUpdate{
					{
						DomainID:  "alphaDomainID",
						Record:  dsdns.Record{},
						Options: &dsdns.RecordUpdateOpts{},
					},
				},
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "alphaDomainID",
						Record: dsdns.Record{},
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

// Test_domeneshopChanges_AddChangeCreate tests domeneshopChanges.AddChangeCreate().
func Test_domeneshopChanges_AddChangeCreate(t *testing.T) {
	type testCase struct {
		name     string
		instance domeneshopChanges
		input    struct {
			domainID  string
			options *dsdns.RecordCreateOpts
		}
		expected domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := tc.instance
		actual.AddChangeCreate(inp.domainID, inp.options)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name:     "add create",
			instance: domeneshopChanges{},
			input: struct {
				domainID  string
				options *dsdns.RecordCreateOpts
			}{
				domainID: "domainIDAlpha",
				options: &dsdns.RecordCreateOpts{
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
			expected: domeneshopChanges{
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

// Test_domeneshopChanges_AddChangeUpdate tests domeneshopChanges.AddChangeUpdate().
func Test_domeneshopChanges_AddChangeUpdate(t *testing.T) {
	type testCase struct {
		name     string
		instance domeneshopChanges
		input    struct {
			domainID  string
			record  dsdns.Record
			options *dsdns.RecordUpdateOpts
		}
		expected domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := tc.instance
		actual.AddChangeUpdate(inp.domainID, inp.record, inp.options)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name:     "add update",
			instance: domeneshopChanges{},
			input: struct {
				domainID  string
				record  dsdns.Record
				options *dsdns.RecordUpdateOpts
			}{
				domainID: "domainIDAlpha",
				record: dsdns.Record{
					ID:    "id_1",
					Host:  "www",
					Ttl:   -1,
					Type:  "A",
					Data: "127.0.0.1",
					Domain: &dsdns.Domain{
						ID:   "domainIDAlpha",
						Domain: "alpha.com",
					},
				},
				options: &dsdns.RecordUpdateOpts{
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
			expected: domeneshopChanges{
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:    "id_1",
							Host:  "www",
							Ttl:   -1,
							Type:  "A",
							Data: "127.0.0.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
						},
						Options: &dsdns.RecordUpdateOpts{
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

// addChangeDelete adds a new delete entry to the current object.
func Test_domeneshopChanges_AddChangeDelete(t *testing.T) {
	type testCase struct {
		name     string
		instance domeneshopChanges
		input    struct {
			domainID string
			record dsdns.Record
		}
		expected domeneshopChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := tc.instance
		actual.AddChangeDelete(inp.domainID, inp.record)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name:     "add update",
			instance: domeneshopChanges{},
			input: struct {
				domainID string
				record dsdns.Record
			}{
				domainID: "domainIDAlpha",
				record: dsdns.Record{
					ID:    "id_1",
					Host:  "www",
					Ttl:   -1,
					Type:  "A",
					Data: "127.0.0.1",
					Domain: &dsdns.Domain{
						ID:   "domainIDAlpha",
						Domain: "alpha.com",
					},
				},
			},
			expected: domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:    "id_1",
							Host:  "www",
							Ttl:   -1,
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

// applyDeletes processes the records to be deleted.
func Test_domeneshopChanges_applyDeletes(t *testing.T) {
	type testCase struct {
		name     string
		changes  *domeneshopChanges
		input    *mockClient
		expected struct {
			state mockClientState
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		err := tc.changes.applyDeletes(context.Background(), inp)
		assertError(t, exp.err, err)
		assert.Equal(t, exp.state, inp.GetState())
	}

	testCases := []testCase{
		{
			name: "deletion",
			changes: &domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:    "id1",
							Type:  dsdns.RecordTypeA,
							Host:  "www",
							Data: "1.1.1.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
			},
			input: &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{DeleteRecordCalled: true},
			},
		},
		{
			name: "deletion error",
			changes: &domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:    "id1",
							Type:  dsdns.RecordTypeA,
							Host:  "www",
							Data: "1.1.1.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
			},
			input: &mockClient{
				deleteRecord: deleteResponse{
					err: errors.New("test delete error"),
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{DeleteRecordCalled: true},
				err:   errors.New("test delete error"),
			},
		},
		{
			name: "deletion dry run",
			changes: &domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:    "id1",
							Type:  dsdns.RecordTypeA,
							Host:  "www",
							Data: "1.1.1.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
				dryRun: true,
			},
			input: &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// applyCreates processes the records to be created.
func Test_domeneshopChanges_applyCreates(t *testing.T) {
	type testCase struct {
		name     string
		changes  *domeneshopChanges
		input    *mockClient
		expected struct {
			state mockClientState
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		err := tc.changes.applyCreates(context.Background(), inp)
		assertError(t, exp.err, err)
		assert.Equal(t, exp.state, inp.GetState())
	}

	testCases := []testCase{
		{
			name: "creation",
			changes: &domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host: "www",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			input: &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{CreateRecordCalled: true},
			},
		},
		{
			name: "creation error",
			changes: &domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host: "www",
							Type: dsdns.RecordTypeA,
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			input: &mockClient{
				createRecord: recordResponse{
					err: errors.New("test creation error"),
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{CreateRecordCalled: true},
				err:   errors.New("test creation error"),
			},
		},
		{
			name: "creation dry run",
			input: &mockClient{
				createRecord: recordResponse{
					err: errors.New("test creation error"),
				},
			},
			changes: &domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host: "www",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
				dryRun: true,
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// applyUpdates processes the records to be updated.
func Test_domeneshopChanges_applyUpdates(t *testing.T) {
	type testCase struct {
		name     string
		changes  *domeneshopChanges
		input    *mockClient
		expected struct {
			state mockClientState
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		err := tc.changes.applyUpdates(context.Background(), inp)
		assertError(t, exp.err, err)
		assert.Equal(t, exp.state, inp.GetState())
	}

	testCases := []testCase{
		{
			name:  "update",
			input: &mockClient{},
			changes: &domeneshopChanges{
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "www",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &dsdns.RecordUpdateOpts{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "ftp",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{UpdateRecordCalled: true},
			},
		},
		{
			name: "update error",
			input: &mockClient{
				updateRecord: recordResponse{
					err: errors.New("test update error"),
				},
			},
			changes: &domeneshopChanges{
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "www",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &dsdns.RecordUpdateOpts{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "ftp",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{UpdateRecordCalled: true},
				err:   errors.New("test update error"),
			},
		},
		{
			name:  "update dry run",
			input: &mockClient{},
			changes: &domeneshopChanges{
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "www",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &dsdns.RecordUpdateOpts{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "ftp",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
				dryRun: true,
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_domeneshopChanges_ApplyChanges tests domeneshopChanges.ApplyChanges().
func Test_domeneshopChanges_ApplyChanges(t *testing.T) {
	type testCase struct {
		name     string
		changes  *domeneshopChanges
		input    *mockClient
		expected struct {
			state mockClientState
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		err := tc.changes.ApplyChanges(context.Background(), inp)
		assertError(t, exp.err, err)
		assert.Equal(t, exp.state, inp.GetState())
	}

	testCases := []testCase{
		{
			name:    "no changes",
			changes: &domeneshopChanges{},
			input:   &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{},
			},
		},
		{
			name: "all changes",
			changes: &domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:    "id1",
							Type:  dsdns.RecordTypeA,
							Host:  "www",
							Data: "1.1.1.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host: "www",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "www",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &dsdns.RecordUpdateOpts{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "ftp",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			input: &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{
					CreateRecordCalled: true,
					DeleteRecordCalled: true,
					UpdateRecordCalled: true,
				},
			},
		},
		{
			name: "deletion error",
			changes: &domeneshopChanges{
				deletes: []*domeneshopChangeDelete{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							ID:    "id1",
							Type:  dsdns.RecordTypeA,
							Host:  "www",
							Data: "1.1.1.1",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
			},
			input: &mockClient{
				deleteRecord: deleteResponse{
					err: errors.New("test delete error"),
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{
					DeleteRecordCalled: true,
				},
				err: errors.New("test delete error"),
			},
		},
		{
			name: "creation error",
			changes: &domeneshopChanges{
				creates: []*domeneshopChangeCreate{
					{
						DomainID: "domainIDAlpha",
						Options: &dsdns.RecordCreateOpts{
							Host: "www",
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			input: &mockClient{
				createRecord: recordResponse{
					err: errors.New("test creation error"),
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{
					CreateRecordCalled: true,
				},
				err: errors.New("test creation error"),
			},
		},
		{
			name: "update error",
			input: &mockClient{
				updateRecord: recordResponse{
					err: errors.New("test update error"),
				},
			},
			changes: &domeneshopChanges{
				updates: []*domeneshopChangeUpdate{
					{
						DomainID: "domainIDAlpha",
						Record: dsdns.Record{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "www",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &dsdns.RecordUpdateOpts{
							Domain: &dsdns.Domain{
								ID:   "domainIDAlpha",
								Domain: "alpha.com",
							},
							Host:  "ftp",
							Type:  "A",
							Data: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{
					UpdateRecordCalled: true,
				},
				err: errors.New("test update error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
