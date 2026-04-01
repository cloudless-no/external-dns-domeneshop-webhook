/*
 * Change processors - this file contains the code for processing changes and
 * queue them.
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
	"strconv"
	"strings"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"

	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
	log "github.com/sirupsen/logrus"
)

// adjustCNAMETarget fixes local CNAME targets. It ensures that targets
// matching the domain are stripped of the domain parts and that "external"
// targets end with a dot.
//
// Domeneshop DNS convention: local hostnames have NO trailing dot, external DO.
// See: https://docs.domeneshop.com/dns-console/dns/record-types/mx-record/
func adjustCNAMETarget(domain string, target string) string {
	adjustedTarget := target
	if strings.HasSuffix(target, "."+domain) {
		adjustedTarget = strings.TrimSuffix(target, "."+domain)
	} else if strings.HasSuffix(target, "."+domain+".") {
		adjustedTarget = strings.TrimSuffix(target, "."+domain+".")
	} else if !strings.HasSuffix(target, ".") {
		adjustedTarget += "."
	}
	return adjustedTarget
}

// adjustMXTarget adjusts MX record target to Domeneshop DNS format.
// MX target format from ExternalDNS: "10 mail.example.com"
// Domeneshop expects: "10 mail" (local) or "10 mail.other.com." (external with dot)
func adjustMXTarget(domain string, target string) string {
	parts := strings.SplitN(target, " ", 2)
	if len(parts) != 2 {
		log.WithFields(log.Fields{
			"target": target,
		}).Warn("MX target has invalid format (expected 'priority hostname')")
		return target
	}
	priority := parts[0]
	host := parts[1]

	// Validate priority is numeric
	if _, err := strconv.Atoi(priority); err != nil {
		log.WithFields(log.Fields{
			"target":   target,
			"priority": priority,
		}).Warn("MX priority is not a valid integer")
		return target
	}

	// Handle apex record (host equals domain)
	hostNoDot := strings.TrimSuffix(host, ".")
	if hostNoDot == domain {
		return priority + " @"
	}

	// Use existing CNAME logic for hostname
	return priority + " " + adjustCNAMETarget(domain, host)
}

// adjustTarget adjusts the target depending on its type
func adjustTarget(domain, recordType, target string) string {
	switch recordType {
	case "CNAME":
		target = adjustCNAMETarget(domain, target)
	case "MX":
		target = adjustMXTarget(domain, target)
	}
	return target
}

// processCreateActionsByDomain processes the create actions for one domain.
func processCreateActionsByDomain(domainID, domainInvalid string, records []dsdns.Record, endpoints []*endpoint.Endpoint, changes *domeneshopChanges) {
	for _, ep := range endpoints {
		// Warn if there are existing records since we expect to create only new records.
		matchingRecords := getMatchingDomainRecords(records, domainInvalid, ep)
		if len(matchingRecords) > 0 {
			log.WithFields(log.Fields{
				"domainInvalid":   domainInvalid,
				"dnsName":    ep.DNSName,
				"recordType": ep.RecordType,
			}).Warn("Preexisting records exist which should not exist for creation actions.")
		}

		for _, target := range ep.Targets {
			target = adjustTarget(domainInvalid, ep.RecordType, target)
			opts := &dsdns.RecordCreateOpts{
				Host:  makeEndpointName(domainInvalid, ep.DNSName),
				Ttl:   getEndpointTTL(ep),
				Type:  dsdns.RecordType(ep.RecordType),
				Data: target,
				Domain: &dsdns.Domain{
					ID:   domainID,
					Domain: domainInvalid,
				},
			}
			changes.AddChangeCreate(domainID, opts)
		}
	}
}

// processCreateActions processes the create requests.
func processCreateActions(
	domainIDNameMapper provider.ZoneIDName,
	recordsByDomainID map[string][]dsdns.Record,
	createsByDomainID map[string][]*endpoint.Endpoint,
	changes *domeneshopChanges,
) {
	// Process endpoints that need to be created.
	for domainID, endpoints := range createsByDomainID {
		domainInvalid := domainIDNameMapper[domainID]
		if len(endpoints) == 0 {
			log.WithFields(log.Fields{
				"domainInvalid": domainInvalid,
			}).Debug("Skipping domain, no creates found.")
			continue
		}
		records := recordsByDomainID[domainID]
		processCreateActionsByDomain(domainID, domainInvalid, records, endpoints, changes)
	}
}

// processUpdateEndpoint processes the update requests for an endpoint.
func processUpdateEndpoint(domainID, domainInvalid string, matchingRecordsByTarget map[string]dsdns.Record, ep *endpoint.Endpoint, changes *domeneshopChanges) {
	// Generate create and delete actions based on existence of a record for each target.
	for _, target := range ep.Targets {
		target = adjustTarget(domainInvalid, ep.RecordType, target)
		if record, ok := matchingRecordsByTarget[target]; ok {
			opts := &dsdns.RecordUpdateOpts{
				Host:  makeEndpointName(domainInvalid, ep.DNSName),
				Ttl:   getEndpointTTL(ep),
				Type:  dsdns.RecordType(ep.RecordType),
				Data: target,
				Domain: &dsdns.Domain{
					ID:   domainID,
					Domain: domainInvalid,
				},
			}
			changes.AddChangeUpdate(domainID, record, opts)

			// Updates are removed from this map.
			delete(matchingRecordsByTarget, target)
		} else {
			// Record did not previously exist, create new 'target'
			opts := &dsdns.RecordCreateOpts{
				Host:  makeEndpointName(domainInvalid, ep.DNSName),
				Ttl:   getEndpointTTL(ep),
				Type:  dsdns.RecordType(ep.RecordType),
				Data: target,
				Domain: &dsdns.Domain{
					ID:   domainID,
					Domain: domainInvalid,
				},
			}
			changes.AddChangeCreate(domainID, opts)
		}
	}
}

// cleanupRemainingTargets deletes the entries for the updates that are queued for creation.
func cleanupRemainingTargets(domainID string, matchingRecordsByTarget map[string]dsdns.Record, changes *domeneshopChanges) {
	for _, record := range matchingRecordsByTarget {
		changes.AddChangeDelete(domainID, record)
	}
}

// getMatchingRecordsByTarget organizes a slice of targets in a map with the
// target as key.
func getMatchingRecordsByTarget(records []dsdns.Record) map[string]dsdns.Record {
	recordsMap := make(map[string]dsdns.Record, 0)
	for _, r := range records {
		recordsMap[r.Data] = r
	}
	return recordsMap
}

// processUpdateActionsByDomain processes update actions for a single domain.
func processUpdateActionsByDomain(domainID, domainInvalid string, records []dsdns.Record, endpoints []*endpoint.Endpoint, changes *domeneshopChanges) {
	for _, ep := range endpoints {
		matchingRecords := getMatchingDomainRecords(records, domainInvalid, ep)

		if len(matchingRecords) == 0 {
			log.WithFields(log.Fields{
				"domainInvalid":   domainInvalid,
				"dnsName":    ep.DNSName,
				"recordType": ep.RecordType,
			}).Warn("Planning an update but no existing records found.")
		}

		matchingRecordsByTarget := getMatchingRecordsByTarget(matchingRecords)

		processUpdateEndpoint(domainID, domainInvalid, matchingRecordsByTarget, ep, changes)

		// Any remaining records have been removed, delete them
		cleanupRemainingTargets(domainID, matchingRecordsByTarget, changes)
	}
}

// processUpdateActions processes the update requests.
func processUpdateActions(
	domainIDNameMapper provider.ZoneIDName,
	recordsByDomainID map[string][]dsdns.Record,
	updatesByDomainID map[string][]*endpoint.Endpoint,
	changes *domeneshopChanges,
) {
	// Generate creates and updates based on existing
	for domainID, endpoints := range updatesByDomainID {
		domainInvalid := domainIDNameMapper[domainID]
		if len(endpoints) == 0 {
			log.WithFields(log.Fields{
				"domainInvalid": domainInvalid,
			}).Debug("Skipping Domain, no updates found.")
			continue
		}

		records := recordsByDomainID[domainID]
		processUpdateActionsByDomain(domainID, domainInvalid, records, endpoints, changes)

	}
}

// targetsMatch determines if a record matches one of the endpoint's targets.
func targetsMatch(record dsdns.Record, ep *endpoint.Endpoint) bool {
	for _, t := range ep.Targets {
		recordTarget := record.Data
		domain := record.Domain.Domain
		endpointTarget := adjustTarget(domain, ep.RecordType, t)
		if endpointTarget == recordTarget {
			return true
		}
	}
	return false
}

// processDeleteActionsByEndpoint processes delete actions for an endpoint.
func processDeleteActionsByEndpoint(domainID string, matchingRecords []dsdns.Record, ep *endpoint.Endpoint, changes *domeneshopChanges) {
	for _, record := range matchingRecords {
		if targetsMatch(record, ep) {
			changes.AddChangeDelete(domainID, record)
		}
	}
}

// processDeleteActionsByDomain processes delete actions for a single domain.
func processDeleteActionsByDomain(domainID, domainInvalid string, records []dsdns.Record, endpoints []*endpoint.Endpoint, changes *domeneshopChanges) {
	for _, ep := range endpoints {
		matchingRecords := getMatchingDomainRecords(records, domainInvalid, ep)

		if len(matchingRecords) == 0 {
			log.WithFields(log.Fields{
				"domainInvalid":   domainInvalid,
				"dnsName":    ep.DNSName,
				"recordType": ep.RecordType,
			}).Warn("Records to delete not found.")
		}
		processDeleteActionsByEndpoint(domainID, matchingRecords, ep, changes)
	}
}

// processDeleteActions processes the delete requests.
func processDeleteActions(
	domainIDNameMapper provider.ZoneIDName,
	recordsByDomainID map[string][]dsdns.Record,
	deletesByDomainID map[string][]*endpoint.Endpoint,
	changes *domeneshopChanges,
) {
	for domainID, endpoints := range deletesByDomainID {
		domainInvalid := domainIDNameMapper[domainID]
		if len(endpoints) == 0 {
			log.WithFields(log.Fields{
				"domainInvalid": domainInvalid,
			}).Debug("Skipping Domain, no deletes found.")
			continue
		}

		records := recordsByDomainID[domainID]
		processDeleteActionsByDomain(domainID, domainInvalid, records, endpoints, changes)

	}
}
