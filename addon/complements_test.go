package addon_test

import (
	"testing"

	"github.com/invopop/gobl.es.verifactu/addon"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvoiceAnomalyLaunchValidation(t *testing.T) {
	t.Run("valid with all checks enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		require.Nil(t, rules.Validate(c))
	})

	t.Run("valid with no checks enabled", func(t *testing.T) {
		c := &addon.InvoiceAnomalyLaunch{}
		require.Nil(t, rules.Validate(c))
	})

	t.Run("missing fingerprint count when check enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		c.FingerprintCount = nil
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "fingerprint count is required when check is enabled")
	})

	t.Run("missing signature count when check enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		c.SignatureCount = nil
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "signature count is required when check is enabled")
	})

	t.Run("missing chain count when check enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		c.ChainCount = nil
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "chain count is required when check is enabled")
	})

	t.Run("missing date count when check enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		c.DateCount = nil
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "date count is required when check is enabled")
	})
}

func TestInvoiceAnomalyValidation(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		c := &addon.InvoiceAnomaly{}
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "anomaly type is required")
	})

	t.Run("valid with invoice", func(t *testing.T) {
		c := &addon.InvoiceAnomaly{
			Type: "01",
			Invoice: &addon.AnomalousInvoice{
				IssuerTaxCode: "B85905495",
				Code:          "SAMPLE-001",
				IssueDate:     cal.MakeDate(2024, 11, 15),
			},
		}
		require.Nil(t, rules.Validate(c))
	})

	t.Run("invoice missing required fields", func(t *testing.T) {
		c := &addon.InvoiceAnomaly{
			Type:    "01",
			Invoice: &addon.AnomalousInvoice{},
		}
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "issuer tax code is required")
		assert.Contains(t, faults.Error(), "invoice code is required")
	})
}

func TestEventAnomalyLaunchValidation(t *testing.T) {
	t.Run("valid with no checks", func(t *testing.T) {
		c := &addon.EventAnomalyLaunch{}
		require.Nil(t, rules.Validate(c))
	})

	t.Run("missing count when check enabled", func(t *testing.T) {
		c := &addon.EventAnomalyLaunch{
			FingerprintCheck: true,
			SignatureCheck:   true,
		}
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "fingerprint count is required when check is enabled")
		assert.Contains(t, faults.Error(), "signature count is required when check is enabled")
	})
}

func TestEventAnomalyValidation(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		c := &addon.EventAnomaly{}
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "anomaly type is required")
	})

	t.Run("event missing required fields", func(t *testing.T) {
		c := &addon.EventAnomaly{
			Type:  "07",
			Event: &addon.AnomalousEvent{},
		}
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "event type is required")
		assert.Contains(t, faults.Error(), "timestamp is required")
		assert.Contains(t, faults.Error(), "fingerprint is required")
	})
}

func TestInvoiceExportValidation(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		c := &addon.InvoiceExport{}
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "period start is required")
		assert.Contains(t, faults.Error(), "period end is required")
		assert.Contains(t, faults.Error(), "first invoice record is required")
		assert.Contains(t, faults.Error(), "last invoice record is required")
		assert.Contains(t, faults.Error(), "discarded flag is required")
	})
}

func TestEventExportValidation(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		c := &addon.EventExport{}
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "period start is required")
		assert.Contains(t, faults.Error(), "period end is required")
		assert.Contains(t, faults.Error(), "first event record is required")
		assert.Contains(t, faults.Error(), "last event record is required")
		assert.Contains(t, faults.Error(), "discarded flag is required")
	})
}

func TestEventSummaryValidation(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		c := &addon.EventSummary{}
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "event type counts are required")
	})

	t.Run("event type entry missing type", func(t *testing.T) {
		c := &addon.EventSummary{
			Events: []*addon.EventTypeCount{
				{Count: 5},
			},
		}
		faults := rules.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "event type is required")
	})

	t.Run("valid summary", func(t *testing.T) {
		c := &addon.EventSummary{
			Events: []*addon.EventTypeCount{
				{Type: "01", Count: 2},
				{Type: "10", Count: 4},
			},
			TaxTotal:    "3780.00",
			AmountTotal: "21780.00",
		}
		require.Nil(t, rules.Validate(c))
	})
}

func validInvoiceAnomalyLaunch() *addon.InvoiceAnomalyLaunch {
	count := 150
	return &addon.InvoiceAnomalyLaunch{
		FingerprintCheck: true,
		FingerprintCount: &count,
		SignatureCheck:   true,
		SignatureCount:   &count,
		ChainCheck:       true,
		ChainCount:       &count,
		DateCheck:        true,
		DateCount:        &count,
	}
}
