package records

import (
	"testing"
	"time"

	"github.com/qdm12/ddns-updater/internal/constants"
	"github.com/qdm12/ddns-updater/internal/models"
	"github.com/qdm12/ddns-updater/internal/provider"
	"github.com/stretchr/testify/assert"
)

type htmlStubProvider struct {
	provider.Provider // only HTML is called
}

func (htmlStubProvider) HTML() models.HTMLRow {
	return models.HTMLRow{Domain: "example.com", Owner: "@", Provider: "stub", IPVersion: "ipv4"}
}

// The status message often carries the DNS provider's HTTP response body, which the
// web UI renders through text/template (no automatic escaping): it must stay escaped.
func Test_Record_HTML_escapesMessage(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	record := Record{
		Provider: htmlStubProvider{},
		Status:   constants.FAIL,
		Message:  `HTTP status is not valid: 500: "><img src=x onerror=alert(1)><script>alert(2)</script>`,
		Time:     now.Add(-time.Minute),
	}

	row := record.HTML(now)

	assert.NotContains(t, row.Status, "<img")
	assert.NotContains(t, row.Status, "<script>")
	assert.Contains(t, row.Status,
		`data-tooltip="HTTP status is not valid: 500: &#34;&gt;&lt;img src=x onerror=alert(1)&gt;`+
			`&lt;script&gt;alert(2)&lt;/script&gt;"`)
}
