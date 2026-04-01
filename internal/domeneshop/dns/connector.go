/*
 * Connector - functions for reading domains and records from Domeneshop DNS
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
	"time"

	"external-dns-domeneshop-webhook/internal/metrics"

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
)

const (
	actGetDomains   = "get_domains"
	actGetRecords   = "get_records"
	actCreateRecord = "create_record"
	actUpdateRecord = "update_record"
	actDeleteRecord = "delete_record"
)

// apiClient is an abstraction of the REST API client.
type apiClient interface {
	// GetDomains returns the available domains.
	GetDomains(ctx context.Context) ([]*dsdns.Domain, *dsdns.Response, error)
	// GetRecords returns the records for a given domain.
	GetRecords(ctx context.Context, opts dsdns.RecordListOpts) ([]*dsdns.Record, *dsdns.Response, error)
	// CreateRecord creates a record.
	CreateRecord(ctx context.Context, opts dsdns.RecordCreateOpts) (*dsdns.Record, *dsdns.Response, error)
	// UpdateRecord updates a single record.
	UpdateRecord(ctx context.Context, record *dsdns.Record, opts dsdns.RecordUpdateOpts) (*dsdns.Record, *dsdns.Response, error)
	// DeleteRecord deletes a single record.
	DeleteRecord(ctx context.Context, record *dsdns.Record, opts dsdns.RecordUpdateOpts) (*dsdns.Response, error)
}

// fetchRecords fetches all records for a given domainID.
func fetchRecords(ctx context.Context, domainID string, dnsClient apiClient) ([]dsdns.Record, error) {
	metrics := metrics.GetOpenMetricsInstance()
	listOptions := dsdns.RecordListOpts{DomainID: domainID}
	start := time.Now()
	pagedRecords, _, err := dnsClient.GetRecords(ctx, listOptions)
	if err != nil {
		metrics.IncFailedApiCallsTotal(actGetRecords)
		return nil, err
	}
	delay := time.Since(start)
	metrics.IncSuccessfulApiCallsTotal(actGetRecords)
	metrics.AddApiDelayHist(actGetRecords, delay.Milliseconds())
	records := make([]dsdns.Record, 0, len(pagedRecords))
	for _, r := range pagedRecords {
		records = append(records, *r)
	}
	return records, nil
}

// fetchDomains fetches all the domains from the DNS client.
func fetchDomains(ctx context.Context, dnsClient apiClient) ([]dsdns.Domain, error) {
	metrics := metrics.GetOpenMetricsInstance()
	start := time.Now()
	pagedDomains, _, err := dnsClient.GetDomains(ctx)
	if err != nil {
		metrics.IncFailedApiCallsTotal(actGetDomains)
		return nil, err
	}
	delay := time.Since(start)
	metrics.IncSuccessfulApiCallsTotal(actGetDomains)
	metrics.AddApiDelayHist(actGetDomains, delay.Milliseconds())
	domains := make([]dsdns.Domain, 0, len(pagedDomains))
	for _, z := range pagedDomains {
		domains = append(domains, *z)
	}
	return domains, nil
}
