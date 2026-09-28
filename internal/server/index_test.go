package server

import (
	"net/netip"
	"testing"
	"time"

	"github.com/qdm12/ddns-updater/internal/models"
	"github.com/qdm12/ddns-updater/internal/records"
	"github.com/stretchr/testify/assert"
)

func Test_latestCurrentIPs(t *testing.T) {
	t.Parallel()

	at := func(hour int) time.Time { return time.Date(2026, 9, 28, hour, 0, 0, 0, time.UTC) }
	record := func(events ...models.HistoryEvent) records.Record {
		return records.Record{History: events}
	}
	event := func(ip string, hour int) models.HistoryEvent {
		return models.HistoryEvent{IP: netip.MustParseAddr(ip), Time: at(hour)}
	}

	ipv4, ipv6 := latestCurrentIPs(nil)
	assert.False(t, ipv4.IsValid())
	assert.False(t, ipv6.IsValid())

	ipv4, ipv6 = latestCurrentIPs([]records.Record{
		record(), // no history yet
		record(event("203.0.113.1", 1), event("203.0.113.2", 3)), // current: .2 at 3h
		record(event("203.0.113.9", 2)),                          // older current IPv4
		record(event("2001:db8::1", 1)),
		record(event("::ffff:198.51.100.7", 0)), // IPv4-mapped counts as IPv4, but older
	})
	assert.Equal(t, netip.MustParseAddr("203.0.113.2"), ipv4)
	assert.Equal(t, netip.MustParseAddr("2001:db8::1"), ipv6)
}
