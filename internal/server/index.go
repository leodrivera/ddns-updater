package server

import (
	"net/http"
	"net/netip"
	"time"

	"github.com/qdm12/ddns-updater/internal/constants"
	"github.com/qdm12/ddns-updater/internal/models"
	"github.com/qdm12/ddns-updater/internal/records"
)

func (h *handlers) index(w http.ResponseWriter, _ *http.Request) {
	// Prevent caching to ensure status updates are always fresh
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	var htmlData models.HTMLData
	successCount := 0
	errorCount := 0
	updatingCount := 0

	records := h.db.SelectAll()
	for _, record := range records {
		row := record.HTML(h.timeNow())
		htmlData.Rows = append(htmlData.Rows, row)

		// Count statuses for summary statistics
		switch record.Status {
		case constants.SUCCESS, constants.UPTODATE:
			successCount++
		case constants.FAIL:
			errorCount++
		case constants.UPDATING:
			updatingCount++
		}
	}

	// Set summary statistics
	htmlData.TotalDomains = len(records)
	htmlData.SuccessCount = successCount
	htmlData.ErrorCount = errorCount
	htmlData.UpdatingCount = updatingCount
	htmlData.LastUpdate = h.timeNow().Format("15:04:05")

	// Public IP addresses: the most recent current IP of each family among the
	// records, i.e. what the updater last saw and set. Not fetched here: the page
	// auto-refreshes, and querying public IP services on every load would eat into
	// their rate limits, which the updater itself depends on.
	ipv4, ipv6 := latestCurrentIPs(records)
	if ipv4.IsValid() {
		htmlData.PublicIPv4 = ipv4.String()
	}
	if ipv6.IsValid() {
		htmlData.PublicIPv6 = ipv6.String()
	}

	err := h.indexTemplate.ExecuteTemplate(w, "index.html", htmlData)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "failed generating webpage: "+err.Error())
	}
}

// latestCurrentIPs returns, for IPv4 and IPv6, the current IP of the record whose
// last successful update is the most recent. An IP is invalid if no record has one.
func latestCurrentIPs(records []records.Record) (ipv4, ipv6 netip.Addr) {
	var ipv4Time, ipv6Time time.Time
	for _, record := range records {
		ip := record.History.GetCurrentIP()
		if !ip.IsValid() {
			continue
		}
		successTime := record.History.GetSuccessTime()
		ip = ip.Unmap()
		switch {
		case ip.Is4() && (!ipv4.IsValid() || successTime.After(ipv4Time)):
			ipv4, ipv4Time = ip, successTime
		case ip.Is6() && (!ipv6.IsValid() || successTime.After(ipv6Time)):
			ipv6, ipv6Time = ip, successTime
		}
	}
	return ipv4, ipv6
}
