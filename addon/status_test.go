package addon_test

import (
	"testing"

	"github.com/invopop/gobl.es.verifactu/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/rules"
	"github.com/invopop/gobl/schema"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// eventLine builds a status line for the given L2E event type code.
func eventLine(code cbc.Code) *bill.StatusLine {
	return &bill.StatusLine{
		Key: bill.StatusLineOther,
		Ext: tax.ExtensionsOf(cbc.CodeMap{
			addon.ExtKeyEventType: code,
		}),
	}
}

// validStatus builds a status document declaring the addon, without which the
// addon-guarded status rules would not be applied at all.
func validStatus() *bill.Status {
	st := &bill.Status{
		Type: bill.StatusTypeSystem,
		Code: "EV-001",
		Supplier: &org.Party{
			Name: "Test Supplier",
			TaxID: &tax.Identity{
				Country: "ES",
				Code:    "B85905495",
			},
		},
		Lines: []*bill.StatusLine{eventLine(addon.CodeSystemStartup)},
	}
	st.SetAddons(addon.V1)
	return st
}

func TestStatusValidation(t *testing.T) {
	t.Run("valid simple event", func(t *testing.T) {
		st := validStatus()
		require.Nil(t, rules.Validate(st))
	})

	t.Run("valid event with complement", func(t *testing.T) {
		st := validStatus()
		st.Lines[0] = eventLine(addon.CodeInvoiceAnomalyLaunch)
		obj, err := schema.NewObject(&addon.InvoiceAnomalyLaunch{})
		require.NoError(t, err)
		st.Lines[0].Complements = []*schema.Object{obj}
		require.Nil(t, rules.Validate(st))
	})

	t.Run("no lines", func(t *testing.T) {
		st := validStatus()
		st.Lines = nil
		faults := rules.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "exactly one line")
	})

	t.Run("empty lines", func(t *testing.T) {
		st := validStatus()
		st.Lines = []*bill.StatusLine{}
		faults := rules.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "exactly one line")
	})

	t.Run("multiple lines", func(t *testing.T) {
		st := validStatus()
		st.Lines = append(st.Lines, eventLine(addon.CodeSystemShutdown))
		faults := rules.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "exactly one line")
	})

	t.Run("wrong line key", func(t *testing.T) {
		st := validStatus()
		st.Lines[0].Key = bill.StatusLineIssued
		faults := rules.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "line key must be 'other'")
	})

	t.Run("missing event type", func(t *testing.T) {
		st := validStatus()
		st.Lines[0].Ext = tax.MakeExtensions()
		faults := rules.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "a supported event type is required")
	})

	t.Run("unsupported event type", func(t *testing.T) {
		st := validStatus()
		st.Lines[0] = eventLine("99")
		faults := rules.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "a supported event type is required")
	})
}

func TestStatusComplementValidation(t *testing.T) {
	// Event types that require a complement
	complementCodes := []struct {
		code       cbc.Code
		complement any
	}{
		{addon.CodeInvoiceAnomalyLaunch, &addon.InvoiceAnomalyLaunch{}},
		{addon.CodeInvoiceAnomaly, &addon.InvoiceAnomaly{Type: "01"}},
		{addon.CodeEventAnomalyLaunch, &addon.EventAnomalyLaunch{}},
		{addon.CodeEventAnomaly, &addon.EventAnomaly{Type: "01"}},
		{addon.CodeInvoiceExport, &addon.InvoiceExport{Start: "x", End: "x", Discarded: "N", FirstRecord: &addon.InvoiceRecord{}, LastRecord: &addon.InvoiceRecord{}}},
		{addon.CodeEventExport, &addon.EventExport{Start: "x", End: "x", Discarded: "N", FirstRecord: &addon.EventRecord{}, LastRecord: &addon.EventRecord{}}},
		{addon.CodeEventSummary, &addon.EventSummary{Events: []*addon.EventTypeCount{{Type: "01", Count: 1}}}},
	}

	for _, tc := range complementCodes {
		t.Run(tc.code.String()+" missing complement", func(t *testing.T) {
			st := validStatus()
			st.Lines[0] = eventLine(tc.code)
			// no complements
			faults := rules.Validate(st)
			require.NotNil(t, faults)
			assert.Contains(t, faults.Error(), "complement is required")
		})

		t.Run(tc.code.String()+" correct complement", func(t *testing.T) {
			st := validStatus()
			st.Lines[0] = eventLine(tc.code)
			obj, err := schema.NewObject(tc.complement)
			require.NoError(t, err)
			st.Lines[0].Complements = []*schema.Object{obj}
			require.Nil(t, rules.Validate(st))
		})

		t.Run(tc.code.String()+" wrong complement", func(t *testing.T) {
			st := validStatus()
			st.Lines[0] = eventLine(tc.code)
			// use a different complement type — pick one that won't match
			wrong := pickWrongComplement(tc.complement)
			obj, err := schema.NewObject(wrong)
			require.NoError(t, err)
			st.Lines[0].Complements = []*schema.Object{obj}
			faults := rules.Validate(st)
			require.NotNil(t, faults)
			assert.Contains(t, faults.Error(), "complement must correspond")
		})
	}

	// Event types that do NOT require a complement — should pass without one
	simpleCodes := []cbc.Code{
		addon.CodeSystemStartup,
		addon.CodeSystemShutdown,
		addon.CodeBackupRestoration,
		addon.CodeOther,
	}

	for _, code := range simpleCodes {
		t.Run(code.String()+" no complement", func(t *testing.T) {
			st := validStatus()
			st.Lines[0] = eventLine(code)
			require.Nil(t, rules.Validate(st))
		})
	}

	// Description length validation for anomaly detection events
	anomalyCodes := []struct {
		code       cbc.Code
		complement any
	}{
		{addon.CodeInvoiceAnomaly, &addon.InvoiceAnomaly{Type: "01"}},
		{addon.CodeEventAnomaly, &addon.EventAnomaly{Type: "01"}},
	}
	for _, tc := range anomalyCodes {
		t.Run(tc.code.String()+" description ok", func(t *testing.T) {
			st := validStatus()
			st.Lines[0] = eventLine(tc.code)
			st.Lines[0].Description = "Short description"
			obj, err := schema.NewObject(tc.complement)
			require.NoError(t, err)
			st.Lines[0].Complements = []*schema.Object{obj}
			require.Nil(t, rules.Validate(st))
		})

		t.Run(tc.code.String()+" description too long", func(t *testing.T) {
			st := validStatus()
			st.Lines[0] = eventLine(tc.code)
			st.Lines[0].Description = string(make([]byte, 101))
			obj, err := schema.NewObject(tc.complement)
			require.NoError(t, err)
			st.Lines[0].Complements = []*schema.Object{obj}
			faults := rules.Validate(st)
			require.NotNil(t, faults)
			assert.Contains(t, faults.Error(), "description must be 100 characters or less")
		})
	}

	// Other note text length validation
	t.Run("other note text ok", func(t *testing.T) {
		st := validStatus()
		st.Notes = []*org.Note{{Key: org.NoteKeyOther, Text: "Short note"}}
		require.Nil(t, rules.Validate(st))
	})

	t.Run("other note text too long", func(t *testing.T) {
		st := validStatus()
		st.Notes = []*org.Note{{Key: org.NoteKeyOther, Text: string(make([]byte, 101))}}
		faults := rules.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "other note text must be 100 characters or less")
	})
}

// pickWrongComplement returns a complement type that is different from the given one.
func pickWrongComplement(correct any) any {
	// Always return EventSummary unless the correct one is already EventSummary,
	// in which case return InvoiceAnomalyLaunch.
	switch correct.(type) {
	case *addon.EventSummary:
		return &addon.InvoiceAnomalyLaunch{}
	default:
		return &addon.EventSummary{
			Events: []*addon.EventTypeCount{{Type: "01", Count: 1}},
		}
	}
}
