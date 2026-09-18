package noverifactu_test

import (
	"testing"

	noverifactu "github.com/invopop/gobl.es.verifactu/pkg/noverifactu"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validStatus() *bill.Status {
	return &bill.Status{
		Type: bill.StatusTypeSystem,
		Code: "EV-001",
		Lines: []*bill.StatusLine{
			{Key: noverifactu.KeySystemStartup},
		},
	}
}

func TestStatusValidation(t *testing.T) {
	t.Run("valid simple event", func(t *testing.T) {
		st := validStatus()
		require.Nil(t, noverifactu.Validate(st))
	})

	t.Run("valid event with complement", func(t *testing.T) {
		st := validStatus()
		st.Lines[0].Key = noverifactu.KeyInvoiceAnomalyLaunch
		obj, err := schema.NewObject(&noverifactu.InvoiceAnomalyLaunch{})
		require.NoError(t, err)
		st.Lines[0].Complements = []*schema.Object{obj}
		require.Nil(t, noverifactu.Validate(st))
	})

	t.Run("no lines", func(t *testing.T) {
		st := validStatus()
		st.Lines = nil
		faults := noverifactu.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "exactly one line")
	})

	t.Run("empty lines", func(t *testing.T) {
		st := validStatus()
		st.Lines = []*bill.StatusLine{}
		faults := noverifactu.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "exactly one line")
	})

	t.Run("multiple lines", func(t *testing.T) {
		st := validStatus()
		st.Lines = append(st.Lines, &bill.StatusLine{Key: noverifactu.KeySystemShutdown})
		faults := noverifactu.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "exactly one line")
	})

	t.Run("missing line key", func(t *testing.T) {
		st := validStatus()
		st.Lines[0].Key = ""
		faults := noverifactu.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "line key is required")
	})

	t.Run("unsupported line key", func(t *testing.T) {
		st := validStatus()
		st.Lines[0].Key = "unsupported-key"
		faults := noverifactu.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "supported event type")
	})
}

func TestStatusComplementValidation(t *testing.T) {
	// Keys that require a complement
	complementKeys := []struct {
		key        string
		complement any
	}{
		{"invoice-anomaly-launch", &noverifactu.InvoiceAnomalyLaunch{}},
		{"invoice-anomaly", &noverifactu.InvoiceAnomaly{Type: "01"}},
		{"event-anomaly-launch", &noverifactu.EventAnomalyLaunch{}},
		{"event-anomaly", &noverifactu.EventAnomaly{Type: "01"}},
		{"invoice-export", &noverifactu.InvoiceExport{Start: "x", End: "x", Discarded: "N", FirstRecord: &noverifactu.InvoiceRecord{}, LastRecord: &noverifactu.InvoiceRecord{}}},
		{"event-export", &noverifactu.EventExport{Start: "x", End: "x", Discarded: "N", FirstRecord: &noverifactu.EventRecord{}, LastRecord: &noverifactu.EventRecord{}}},
		{"event-summary", &noverifactu.EventSummary{Events: []*noverifactu.EventTypeCount{{Type: "01", Count: 1}}}},
	}

	for _, tc := range complementKeys {
		t.Run(tc.key+" missing complement", func(t *testing.T) {
			st := validStatus()
			st.Lines[0].Key = cbc.Key(tc.key)
			// no complements
			faults := noverifactu.Validate(st)
			require.NotNil(t, faults)
			assert.Contains(t, faults.Error(), "complement is required")
		})

		t.Run(tc.key+" correct complement", func(t *testing.T) {
			st := validStatus()
			st.Lines[0].Key = cbc.Key(tc.key)
			obj, err := schema.NewObject(tc.complement)
			require.NoError(t, err)
			st.Lines[0].Complements = []*schema.Object{obj}
			require.Nil(t, noverifactu.Validate(st))
		})

		t.Run(tc.key+" wrong complement", func(t *testing.T) {
			st := validStatus()
			st.Lines[0].Key = cbc.Key(tc.key)
			// use a different complement type — pick one that won't match
			wrong := pickWrongComplement(tc.complement)
			obj, err := schema.NewObject(wrong)
			require.NoError(t, err)
			st.Lines[0].Complements = []*schema.Object{obj}
			faults := noverifactu.Validate(st)
			require.NotNil(t, faults)
			assert.Contains(t, faults.Error(), "complement must correspond")
		})
	}

	// Keys that do NOT require a complement — should pass without one
	simpleKeys := []string{
		"system-startup",
		"system-shutdown",
		"backup-restoration",
		"other",
	}

	for _, key := range simpleKeys {
		t.Run(key+" no complement", func(t *testing.T) {
			st := validStatus()
			st.Lines[0].Key = cbc.Key(key)
			require.Nil(t, noverifactu.Validate(st))
		})
	}

	// Description length validation for anomaly detection events
	anomalyKeys := []struct {
		key        cbc.Key
		complement any
	}{
		{noverifactu.KeyInvoiceAnomaly, &noverifactu.InvoiceAnomaly{Type: "01"}},
		{noverifactu.KeyEventAnomaly, &noverifactu.EventAnomaly{Type: "01"}},
	}
	for _, tc := range anomalyKeys {
		t.Run(string(tc.key)+" description ok", func(t *testing.T) {
			st := validStatus()
			st.Lines[0].Key = tc.key
			st.Lines[0].Description = "Short description"
			obj, err := schema.NewObject(tc.complement)
			require.NoError(t, err)
			st.Lines[0].Complements = []*schema.Object{obj}
			require.Nil(t, noverifactu.Validate(st))
		})

		t.Run(string(tc.key)+" description too long", func(t *testing.T) {
			st := validStatus()
			st.Lines[0].Key = tc.key
			st.Lines[0].Description = string(make([]byte, 101))
			obj, err := schema.NewObject(tc.complement)
			require.NoError(t, err)
			st.Lines[0].Complements = []*schema.Object{obj}
			faults := noverifactu.Validate(st)
			require.NotNil(t, faults)
			assert.Contains(t, faults.Error(), "description must be 100 characters or less")
		})
	}

	// Other note text length validation
	t.Run("other note text ok", func(t *testing.T) {
		st := validStatus()
		st.Notes = []*org.Note{{Key: org.NoteKeyOther, Text: "Short note"}}
		require.Nil(t, noverifactu.Validate(st))
	})

	t.Run("other note text too long", func(t *testing.T) {
		st := validStatus()
		st.Notes = []*org.Note{{Key: org.NoteKeyOther, Text: string(make([]byte, 101))}}
		faults := noverifactu.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "other note text must be 100 characters or less")
	})
}

// pickWrongComplement returns a complement type that is different from the given one.
func pickWrongComplement(correct any) any {
	// Always return EventSummary unless the correct one is already EventSummary,
	// in which case return InvoiceAnomalyLaunch.
	switch correct.(type) {
	case *noverifactu.EventSummary:
		return &noverifactu.InvoiceAnomalyLaunch{}
	default:
		return &noverifactu.EventSummary{
			Events: []*noverifactu.EventTypeCount{{Type: "01", Count: 1}},
		}
	}
}
