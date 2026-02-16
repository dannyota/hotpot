package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/digitalocean/godo"
)

// DomainData holds converted Domain data ready for Ent insertion.
type DomainData struct {
	ResourceID  string
	TTL         int
	ZoneFile    string
	CollectedAt time.Time
}

// ConvertDomain converts a godo Domain to DomainData.
func ConvertDomain(v godo.Domain, collectedAt time.Time) *DomainData {
	return &DomainData{
		ResourceID:  v.Name,
		TTL:         v.TTL,
		ZoneFile:    v.ZoneFile,
		CollectedAt: collectedAt,
	}
}

// DomainRecordData holds converted Domain Record data ready for Ent insertion.
type DomainRecordData struct {
	ResourceID  string
	DomainName  string
	RecordID    int
	Type        string
	Name        string
	Data        string
	Priority    int
	Port        int
	TTL         int
	Weight      int
	Flags       int
	Tag         string
	CollectedAt time.Time
}

// ConvertDomainRecord converts a godo DomainRecord to DomainRecordData.
func ConvertDomainRecord(v godo.DomainRecord, domainName string, collectedAt time.Time) *DomainRecordData {
	return &DomainRecordData{
		ResourceID:  fmt.Sprintf("%s:%s", domainName, strconv.Itoa(v.ID)),
		DomainName:  domainName,
		RecordID:    v.ID,
		Type:        v.Type,
		Name:        v.Name,
		Data:        v.Data,
		Priority:    v.Priority,
		Port:        v.Port,
		TTL:         v.TTL,
		Weight:      v.Weight,
		Flags:       v.Flags,
		Tag:         v.Tag,
		CollectedAt: collectedAt,
	}
}
