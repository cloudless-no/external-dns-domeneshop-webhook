/*
 * Provider - class and functions that handle the connection to Domeneshop DNS.
 *
 * This file was MODIFIED from the original provider to be used as a standalone
 * webhook server.
 *
 * Copyright 2026 Marco Confalonieri.
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
	"context"
	"fmt"
	"time"

	"external-dns-domeneshop-webhook/internal/domeneshop"
	"external-dns-domeneshop-webhook/internal/metrics"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
	log "github.com/sirupsen/logrus"
)

// logFatalf is a mockable call to log.Fatalf
var logFatalf = log.Fatalf

// DomeneshopProvider implements ExternalDNS' provider.Provider interface for
// Domeneshop.
type DomeneshopProvider struct {
	provider.BaseProvider
	client            apiClient
	batchSize         int
	debug             bool
	dryRun            bool
	domainIDNameMapper  provider.ZoneIDName
	domainFilter      *endpoint.DomainFilter
	maxFailCount      int
	failCount         int
	domainCacheDuration time.Duration
	domainCacheUpdate   time.Time
	domainCache         []dsdns.Domain
}

// NewDomeneshopProvider creates a new DomeneshopProvider instance.
func NewDomeneshopProvider(config *domeneshop.Configuration) (*DomeneshopProvider, error) {
	var logLevel log.Level
	if config.Debug {
		logLevel = log.DebugLevel
	} else {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	client, err := NewDomeneshopDNS(config.Token, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate DNS provider: %w", err)
	}

	var msg string
	if config.MaxFailCount > 0 {
		msg = fmt.Sprintf("Configuring DNS provider with maximum fail count of %d", config.MaxFailCount)
	} else {
		msg = "Configuring DNS provider without maximum fail count"
	}
	log.Info(msg)

	zcTTL := time.Duration(int64(config.DomainCacheTTL) * int64(time.Second))
	zcUpdate := time.Now()

	if zcTTL > 0 {
		log.Infof("Domain cache enabled. TTL=%ds.", config.DomainCacheTTL)
	} else {
		log.Info("Domain cache disabled in configuration.")
	}

	return &DomeneshopProvider{
		client:            client,
		batchSize:         config.BatchSize,
		debug:             config.Debug,
		dryRun:            config.DryRun,
		domainFilter:      domeneshop.GetDomainFilter(*config),
		maxFailCount:      config.MaxFailCount,
		domainCacheDuration: zcTTL,
		domainCacheUpdate:   zcUpdate,
	}, nil
}

// incFailCount increments the fail count and exit if necessary.
func (p *DomeneshopProvider) incFailCount() {
	if p.maxFailCount <= 0 {
		return
	}
	p.failCount++
	if p.failCount >= p.maxFailCount {
		logFatalf("Failure count reached %d. Shutting down container.", p.failCount)
	}
}

// resetFailCount resets the fail count.
func (p *DomeneshopProvider) resetFailCount() {
	if p.maxFailCount <= 0 {
		return
	}
	p.failCount = 0
}

// Domains returns the list of the hosted DNS domains.
// If a domain filter is set, it only returns the domains that match it.
func (p *DomeneshopProvider) Domains(ctx context.Context) ([]dsdns.Domain, error) {
	now := time.Now()
	if now.Before(p.domainCacheUpdate) && p.domainCache != nil {
		nextUpdate := int(p.domainCacheUpdate.Sub(now).Seconds())
		log.Debugf("Using cached domains. The cache expires in %d seconds.", nextUpdate)
		return p.domainCache, nil
	}
	metrics := metrics.GetOpenMetricsInstance()
	result := []dsdns.Domain{}

	domains, err := fetchDomains(ctx, p.client)
	if err != nil {
		return nil, err
	}

	filteredOutDomains := 0
	for _, domain := range domains {
		if p.domainFilter.Match(domain.Domain) {
			result = append(result, domain)
		} else {
			filteredOutDomains++
		}
	}
	metrics.SetFilteredOutDomains(filteredOutDomains)

	log.Debugf("Got %d domains, filtered out %d domains.", len(domains), filteredOutDomains)
	p.ensureDomainIDMappingPresent(domains)
	p.domainCache = result
	p.domainCacheUpdate = now.Add(p.domainCacheDuration)

	return result, nil
}

// AdjustEndpoints adjusts the endpoints according to the provider
// requirements.
func (p DomeneshopProvider) AdjustEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error) {
	adjustedEndpoints := []*endpoint.Endpoint{}

	for _, ep := range endpoints {
		_, domainName := p.domainIDNameMapper.FindZone(ep.DNSName)
		adjustedTargets := endpoint.Targets{}
		for _, t := range ep.Targets {
			adjustedTarget := makeEndpointTarget(domainName, t, ep.RecordType)
			adjustedTargets = append(adjustedTargets, adjustedTarget)
		}

		ep.Targets = adjustedTargets
		adjustedEndpoints = append(adjustedEndpoints, ep)
	}

	return adjustedEndpoints, nil
}

// logDebugEndpoints logs every endpoint as a a line.
func logDebugEndpoints(endpoints []*endpoint.Endpoint) {
	for idx, ep := range endpoints {
		log.WithFields(getEndpointLogFields(ep)).Debugf("Endpoint %d", idx)
	}
}

// Records returns the list of records in all domains as a slice of endpoints.
func (p *DomeneshopProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	domains, err := p.Domains(ctx)
	if err != nil {
		p.incFailCount()
		return nil, err
	}
	p.resetFailCount()

	endpoints := []*endpoint.Endpoint{}
	for _, domain := range domains {
		records, err := fetchRecords(ctx, domain.ID, p.client)
		if err != nil {
			return nil, err
		}

		skippedRecords := 0
		// Add only endpoints from supported types.
		for _, r := range records {
			// Ensure the record has all the required domain information
			r.Domain = &domain
			// Use our own IsSupportedRecordType instead of provider.SupportedRecordType
			// because the SDK function doesn't include MX in its hardcoded list.
			if domeneshop.IsSupportedRecordType(string(r.Type)) {
				ep := createEndpointFromRecord(r)
				endpoints = append(endpoints, ep)
			} else {
				skippedRecords++
			}
		}
		m := metrics.GetOpenMetricsInstance()
		m.SetSkippedRecords(domain.Domain, skippedRecords)
	}

	// Merge endpoints with the same name and type (e.g., multiple A records for a single
	// DNS name) into one endpoint with multiple targets.
	endpoints = mergeEndpointsByNameType(endpoints)

	// Log the endpoints that were found.
	if p.debug {
		log.Debugf("Returning %d endpoints.", len(endpoints))
		logDebugEndpoints(endpoints)
	}

	return endpoints, nil
}

// ensureDomainIDMappingPresent prepares the domainIDNameMapper, that associates
// each DomainID woth the domain name.
func (p *DomeneshopProvider) ensureDomainIDMappingPresent(domains []dsdns.Domain) {
	domainIDNameMapper := provider.ZoneIDName{}
	for _, z := range domains {
		domainIDNameMapper.Add(z.ID, z.Domain)
	}
	p.domainIDNameMapper = domainIDNameMapper
}

// getRecordsByDomainID returns a map that associates each DomainID with the
// records contained in that domain.
func (p *DomeneshopProvider) getRecordsByDomainID(ctx context.Context) (map[string][]dsdns.Record, error) {
	recordsByDomainID := map[string][]dsdns.Record{}

	domains, err := p.Domains(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch records for each domain
	for _, domain := range domains {
		records, err := fetchRecords(ctx, domain.ID, p.client)
		if err != nil {
			return nil, err
		}
		// Add full domain information
		domaindRecords := []dsdns.Record{}
		for _, r := range records {
			r.Domain = &domain
			domaindRecords = append(domaindRecords, r)
		}
		recordsByDomainID[domain.ID] = append(recordsByDomainID[domain.ID], domaindRecords...)
	}

	return recordsByDomainID, nil
}

// ApplyChanges applies the given set of generic changes to the provider.
func (p DomeneshopProvider) ApplyChanges(ctx context.Context, planChanges *plan.Changes) error {
	if !planChanges.HasChanges() {
		return nil
	}

	recordsByDomainID, err := p.getRecordsByDomainID(ctx)
	if err != nil {
		return err
	}

	log.Debug("Preparing creates")
	createsByDomainID := endpointsByDomainID(p.domainIDNameMapper, planChanges.Create)
	log.Debug("Preparing updates")
	updatesByDomainID := endpointsByDomainID(p.domainIDNameMapper, planChanges.UpdateNew)
	log.Debug("Preparing deletes")
	deletesByDomainID := endpointsByDomainID(p.domainIDNameMapper, planChanges.Delete)

	changes := domeneshopChanges{
		dryRun: p.dryRun,
	}

	processCreateActions(p.domainIDNameMapper, recordsByDomainID, createsByDomainID, &changes)
	processUpdateActions(p.domainIDNameMapper, recordsByDomainID, updatesByDomainID, &changes)
	processDeleteActions(p.domainIDNameMapper, recordsByDomainID, deletesByDomainID, &changes)

	return changes.ApplyChanges(ctx, p.client)
}

// GetDomainFilter returns the domain filter
func (p DomeneshopProvider) GetDomainFilter() endpoint.DomainFilterInterface {
	return p.domainFilter
}
