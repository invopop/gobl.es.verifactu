package noverifactu_test

import (
	"testing"

	noverifactu "github.com/invopop/gobl.es.verifactu/pkg/noverifactu"
	"github.com/invopop/gobl/cal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvoiceAnomalyLaunchValidation(t *testing.T) {
	t.Run("valid with all checks enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		require.Nil(t, noverifactu.Validate(c))
	})

	t.Run("valid with no checks enabled", func(t *testing.T) {
		c := &noverifactu.InvoiceAnomalyLaunch{}
		require.Nil(t, noverifactu.Validate(c))
	})

	t.Run("missing fingerprint count when check enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		c.FingerprintCount = nil
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "fingerprint count is required when check is enabled")
	})

	t.Run("missing signature count when check enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		c.SignatureCount = nil
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "signature count is required when check is enabled")
	})

	t.Run("missing chain count when check enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		c.ChainCount = nil
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "chain count is required when check is enabled")
	})

	t.Run("missing date count when check enabled", func(t *testing.T) {
		c := validInvoiceAnomalyLaunch()
		c.DateCount = nil
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "date count is required when check is enabled")
	})
}

func TestInvoiceAnomalyValidation(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		c := &noverifactu.InvoiceAnomaly{}
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "anomaly type is required")
	})

	t.Run("valid with invoice", func(t *testing.T) {
		c := &noverifactu.InvoiceAnomaly{
			Type: "01",
			Invoice: &noverifactu.AnomalousInvoice{
				IssuerTaxCode: "B85905495",
				Code:          "SAMPLE-001",
				IssueDate:     cal.MakeDate(2024, 11, 15),
			},
		}
		require.Nil(t, noverifactu.Validate(c))
	})

	t.Run("invoice missing required fields", func(t *testing.T) {
		c := &noverifactu.InvoiceAnomaly{
			Type:    "01",
			Invoice: &noverifactu.AnomalousInvoice{},
		}
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "issuer tax code is required")
		assert.Contains(t, faults.Error(), "invoice code is required")
	})
}

func TestEventAnomalyLaunchValidation(t *testing.T) {
	t.Run("valid with no checks", func(t *testing.T) {
		c := &noverifactu.EventAnomalyLaunch{}
		require.Nil(t, noverifactu.Validate(c))
	})

	t.Run("missing count when check enabled", func(t *testing.T) {
		c := &noverifactu.EventAnomalyLaunch{
			FingerprintCheck: true,
			SignatureCheck:   true,
		}
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "fingerprint count is required when check is enabled")
		assert.Contains(t, faults.Error(), "signature count is required when check is enabled")
	})
}

func TestEventAnomalyValidation(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		c := &noverifactu.EventAnomaly{}
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "anomaly type is required")
	})

	t.Run("event missing required fields", func(t *testing.T) {
		c := &noverifactu.EventAnomaly{
			Type:  "07",
			Event: &noverifactu.AnomalousEvent{},
		}
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "event type is required")
		assert.Contains(t, faults.Error(), "timestamp is required")
		assert.Contains(t, faults.Error(), "fingerprint is required")
	})
}

func TestInvoiceExportValidation(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		c := &noverifactu.InvoiceExport{}
		faults := noverifactu.Validate(c)
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
		c := &noverifactu.EventExport{}
		faults := noverifactu.Validate(c)
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
		c := &noverifactu.EventSummary{}
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "event type counts are required")
	})

	t.Run("event type entry missing type", func(t *testing.T) {
		c := &noverifactu.EventSummary{
			Events: []*noverifactu.EventTypeCount{
				{Count: 5},
			},
		}
		faults := noverifactu.Validate(c)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "event type is required")
	})

	t.Run("valid summary", func(t *testing.T) {
		c := &noverifactu.EventSummary{
			Events: []*noverifactu.EventTypeCount{
				{Type: "01", Count: 2},
				{Type: "10", Count: 4},
			},
			TaxTotal:    "3780.00",
			AmountTotal: "21780.00",
		}
		require.Nil(t, noverifactu.Validate(c))
	})
}

func validInvoiceAnomalyLaunch() *noverifactu.InvoiceAnomalyLaunch {
	count := 150
	return &noverifactu.InvoiceAnomalyLaunch{
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
