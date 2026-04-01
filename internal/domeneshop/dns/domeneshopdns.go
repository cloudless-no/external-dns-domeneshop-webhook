/*
 * DomeneshopDNS - This handles API calls towards Domeneshop DNS.
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

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
)

// domeneshopDNS is the DNS client API.
type domeneshopDNS struct {
	client *dsdns.Client
}

// NewDomeneshopDNS returns a new client.
func NewDomeneshopDNS(token string, secret string) (*domeneshopDNS, error) {
	if token == "" {
		return nil, errors.New("nil Token provided")
	}
	if secret == "" {
		return nil, errors.New("nil Secret provided")
	}
	return &domeneshopDNS{
		client: dsdns.NewClient(dsdns.WithCredentials(token, secret)),
	}, nil
}

// GetDomains returns the available domains.
func (h domeneshopDNS) GetDomains(ctx context.Context) ([]*dsdns.Domain, *dsdns.Response, error) {
	domainClient := h.client.Domain
	return domainClient.List(ctx)
}

// GetRecords returns the records for a given domain.
func (h domeneshopDNS) GetRecords(ctx context.Context, opts dsdns.RecordListOpts,
) ([]*dsdns.Record, *dsdns.Response, error) {
	recordClient := h.client.Record
	return recordClient.List(ctx, opts)
}

// CreateRecord creates a record.
func (h domeneshopDNS) CreateRecord(ctx context.Context, opts dsdns.RecordCreateOpts) (*dsdns.Record, *dsdns.Response, error) {
	recordClient := h.client.Record
	return recordClient.Create(ctx, opts)
}

// UpdateRecord updates a single record.
func (h domeneshopDNS) UpdateRecord(ctx context.Context, record *dsdns.Record, opts dsdns.RecordUpdateOpts) (*dsdns.Record, *dsdns.Response, error) {
	recordClient := h.client.Record
	return recordClient.Update(ctx, record, opts)
}

// DeleteRecord deletes a single record.
func (h domeneshopDNS) DeleteRecord(ctx context.Context, record *dsdns.Record, opts dsdns.RecordUpdateOpts) (*dsdns.Response, error) {
	recordClient := h.client.Record
	return recordClient.Delete(ctx, record, opts)
}
