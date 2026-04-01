/*
 * Endpoints - functions for handling and transforming endpoints.
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
	"fmt"
	"strings"

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

// fromDomeneshopHostname converts Domeneshop DNS hostname format back to FQDN for ExternalDNS.
// This is the inverse of adjustCNAMETarget() and adjustMXTarget() in change_processors.go.
// Domeneshop uses domain-relative hostnames: "mail" for local, "external.com." for external.
// ExternalDNS works WITHOUT trailing dot internally, so we return names without it.
//
// Key insight: Domeneshop convention is that EXTERNAL hostnames have trailing dot,
// while LOCAL hostnames (within the domain) do NOT have trailing dot.
//
// References:
//   - Domeneshop DNS docs: "When there is no period at the end, the domain itself is appended automatically"
//     https://docs.domeneshop.com/dns-console/dns/record-types/mx-record/
//   - DNS trailing dot convention: https://docs.dnscontrol.org/language-reference/why-the-dot
//
// Examples (domain = "alpha.com"):
//
//	"@"              → "alpha.com"       (apex record)
//	"mail"           → "mail.alpha.com"  (local subdomain)
//	"a.b"            → "a.b.alpha.com"   (deep local subdomain)
//	"external.com."  → "external.com"    (external, has trailing dot → strip it)
//	"mail.beta.com." → "mail.beta.com"   (external, has trailing dot → strip it)
func fromDomeneshopHostname(domain string, host string) string {
	// Handle apex record
	if host == "@" {
		return domain
	}

	// Domeneshop convention: trailing dot means EXTERNAL hostname (outside of domain)
	// No trailing dot means LOCAL hostname (within domain)
	if strings.HasSuffix(host, ".") {
		// External hostname - just strip the trailing dot
		return strings.TrimSuffix(host, ".")
	}

	// Local hostname (no trailing dot) - append domain
	return host + "." + domain
}

// makeEndpointName makes a endpoint name that conforms to Domeneshop DNS
// requirements. It converts an FQDN to a domain-relative name.
func makeEndpointName(domain, entryName string) string {
	// Trim the domain off the name if present.
	adjustedName := strings.TrimSuffix(entryName, "."+domain)

	// Record at the root should be defined as @ instead of the full domain name.
	if adjustedName == domain {
		adjustedName = "@"
	}

	return adjustedName
}

// makeEndpointTarget makes a endpoint target that conforms to Domeneshop DNS
// requirements:
//   - A-Records should respect ignored networks and should only contain IPv4
//     entries.
func makeEndpointTarget(domain, entryTarget string, _ string) string {
	if domain == "" {
		return entryTarget
	}

	// Trim the trailing dot
	adjustedTarget := strings.TrimSuffix(entryTarget, ".")

	return adjustedTarget
}

// mergeEndpointsByNameType merges Endpoints with the same Name and Type into a
// single endpoint with multiple Targets.
func mergeEndpointsByNameType(endpoints []*endpoint.Endpoint) []*endpoint.Endpoint {
	endpointsByNameType := map[string][]*endpoint.Endpoint{}

	for _, e := range endpoints {
		key := fmt.Sprintf("%s-%s", e.DNSName, e.RecordType)
		endpointsByNameType[key] = append(endpointsByNameType[key], e)
	}

	// If no merge occurred, just return the existing endpoints.
	if len(endpointsByNameType) == len(endpoints) {
		return endpoints
	}

	// Otherwise, construct a new list of endpoints with the endpoints merged.
	var result []*endpoint.Endpoint
	for _, endpoints := range endpointsByNameType {
		dnsName := endpoints[0].DNSName
		recordType := endpoints[0].RecordType

		targets := make([]string, len(endpoints))
		for i, e := range endpoints {
			targets[i] = e.Targets[0]
		}

		e := endpoint.NewEndpoint(dnsName, recordType, targets...)
		e.RecordTTL = endpoints[0].RecordTTL
		result = append(result, e)
	}

	return result
}

// createEndpointFromRecord creates an endpoint from a record.
func createEndpointFromRecord(r dsdns.Record) *endpoint.Endpoint {
	name := fmt.Sprintf("%s.%s", r.Host, r.Domain.Domain)

	// root name is identified by @ and should be
	// translated to domain name for the endpoint entry.
	if r.Host == "@" {
		name = r.Domain.Domain
	}

	// Handle local CNAMEs
	target := r.Data
	domainName := r.Domain.Domain
	switch r.Type {
	case dsdns.RecordTypeCNAME:
		target = fromDomeneshopHostname(domainName, target)
	case dsdns.RecordTypeMX:
		// MX records in Domeneshop: "10 mail" (local) or "10 mail.beta.com." (external)
		// Convert to ExternalDNS format: "10 mail.domain.com" (FQDN without trailing dot)
		parts := strings.SplitN(target, " ", 2)
		if len(parts) == 2 {
			priority := parts[0]
			host := fromDomeneshopHostname(domainName, parts[1])
			target = priority + " " + host
		} else {
			log.WithFields(log.Fields{
				"domain": domainName,
				"target": target,
			}).Warn("MX record from Domeneshop API has unexpected format (expected 'priority hostname')")
		}
	}
	ep := endpoint.NewEndpoint(name, string(r.Type), target)
	ep.RecordTTL = endpoint.TTL(r.Ttl)
	return ep
}

// endpointsByDomainID arranges the endpoints in a map by domain ID.
func endpointsByDomainID(domainIDNameMapper provider.ZoneIDName, endpoints []*endpoint.Endpoint) map[string][]*endpoint.Endpoint {
	endpointsByDomainID := make(map[string][]*endpoint.Endpoint)

	for idx, ep := range endpoints {
		domainID, _ := domainIDNameMapper.FindZone(ep.DNSName)
		if domainID == "" {
			log.Debugf("Skipping record %d (%s) because no hosted domain matching record DNS Name was detected", idx, ep.DNSName)
			continue
		} else {
			log.WithFields(getEndpointLogFields(ep)).Debugf("Reading endpoint %d for dividing by domain", idx)
		}
		endpointsByDomainID[domainID] = append(endpointsByDomainID[domainID], ep)
	}

	return endpointsByDomainID
}

// getMatchingDomainRecords returns the records that match an endpoint.
func getMatchingDomainRecords(records []dsdns.Record, domainName string, ep *endpoint.Endpoint) []dsdns.Record {
	var name string
	if len(ep.ProviderSpecific) > 0 {
		log.Warnf("Ignoring provider-specific directives in endpoint [%s] of type [%s].", ep.DNSName, ep.RecordType)
	}
	if ep.DNSName != domainName {
		name = strings.TrimSuffix(ep.DNSName, "."+domainName)
	} else {
		name = "@"
	}

	var result []dsdns.Record
	for _, r := range records {
		if r.Host == name && string(r.Type) == ep.RecordType {
			result = append(result, r)
		}
	}
	return result
}

// getEndpointTTL returns a pointer to a value representing the endpoint TTL or
// nil if it is not configured.
func getEndpointTTL(ep *endpoint.Endpoint) *int {
	if !ep.RecordTTL.IsConfigured() {
		return nil
	}
	ttl := int(ep.RecordTTL)
	return &ttl
}

// getEndpointLogFields returns a loggable field map.
func getEndpointLogFields(ep *endpoint.Endpoint) log.Fields {
	return log.Fields{
		"DNSName":    ep.DNSName,
		"RecordType": ep.RecordType,
		"Targets":    ep.Targets.String(),
		"TTL":        int(ep.RecordTTL),
	}
}
