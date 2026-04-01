/*
 * Changes Internals - Internal structures for processing changes.
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
	dsdns "github.com/cloudless-no/domeneshop-dns-go/dns"
	log "github.com/sirupsen/logrus"
)

// domeneshopChangeCreate stores the information for a create request.
type domeneshopChangeCreate struct {
	DomainID  string
	Options *dsdns.RecordCreateOpts
}

// GetLogFields returns the log fields for this object.
func (cc domeneshopChangeCreate) GetLogFields() log.Fields {
	return log.Fields{
		"domain":     cc.Options.Domain.Domain,
		"domainID":     cc.DomainID,
		"dnsName":    cc.Options.Host,
		"recordType": string(cc.Options.Type),
		"value":      cc.Options.Data,
		"ttl":        *cc.Options.Ttl,
	}
}

// domeneshopChangeUpdate stores the information for an update request.
type domeneshopChangeUpdate struct {
	DomainID  string
	Record  dsdns.Record
	Options *dsdns.RecordUpdateOpts
}

// GetLogFields returns the log fields for this object. An asterisk indicate
// that the new value is shown.
func (cu domeneshopChangeUpdate) GetLogFields() log.Fields {
	return log.Fields{
		"domain":      cu.Record.Domain.Domain,
		"domainID":      cu.DomainID,
		"recordID":    cu.Record.ID,
		"*dnsName":    cu.Options.Host,
		"*recordType": string(cu.Options.Type),
		"*value":      cu.Options.Data,
		"*ttl":        *cu.Options.Ttl,
	}
}

// domeneshopChangeDelete stores the information for a delete request.
type domeneshopChangeDelete struct {
	DomainID string
	Record dsdns.Record
}

// GetLogFields returns the log fields for this object.
func (cd domeneshopChangeDelete) GetLogFields() log.Fields {
	return log.Fields{
		"domain":     cd.Record.Domain.Domain,
		"domainID":     cd.DomainID,
		"dnsName":    cd.Record.Host,
		"recordType": string(cd.Record.Type),
		"value":      cd.Record.Data,
	}
}
