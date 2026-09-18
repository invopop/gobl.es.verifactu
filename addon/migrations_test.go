package addon_test

import (
	"testing"

	"github.com/invopop/gobl.es.verifactu/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/rules"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// legacyStatus builds a status document carrying a pre-extension status line
// key, as documents issued before the event type moved to an extension did.
func legacyStatus(key cbc.Key) *bill.Status {
	st := validStatus()
	st.Lines[0] = &bill.StatusLine{Key: key}
	return st
}

func TestStatusLineKeyMigration(t *testing.T) {
	// Every legacy key and the event type code it must migrate to.
	migrations := map[cbc.Key]cbc.Code{
		"system-startup":         addon.CodeSystemStartup,
		"system-shutdown":        addon.CodeSystemShutdown,
		"invoice-anomaly-launch": addon.CodeInvoiceAnomalyLaunch,
		"invoice-anomaly":        addon.CodeInvoiceAnomaly,
		"event-anomaly-launch":   addon.CodeEventAnomalyLaunch,
		"event-anomaly":          addon.CodeEventAnomaly,
		"backup-restoration":     addon.CodeBackupRestoration,
		"invoice-export":         addon.CodeInvoiceExport,
		"event-export":           addon.CodeEventExport,
		"event-summary":          addon.CodeEventSummary,
	}

	for key, code := range migrations {
		t.Run(key.String(), func(t *testing.T) {
			st := legacyStatus(key)
			require.NoError(t, st.Calculate())

			line := st.Lines[0]
			assert.Equal(t, bill.StatusLineOther, line.Key, "key should become 'other'")
			assert.Equal(t, code, addon.EventType(line), "event type should come from the legacy key")
		})
	}

	t.Run("leaves an existing event type alone", func(t *testing.T) {
		// An explicit event type wins over the one implied by the key, so a
		// document that has already been migrated is never rewritten.
		st := legacyStatus("system-startup")
		st.Lines[0].Ext = tax.ExtensionsOf(cbc.CodeMap{
			addon.ExtKeyEventType: addon.CodeEventSummary,
		})
		require.NoError(t, st.Calculate())

		assert.Equal(t, addon.CodeEventSummary, addon.EventType(st.Lines[0]))
	})

	t.Run("leaves an unknown key alone", func(t *testing.T) {
		st := legacyStatus("not-an-event-type")
		require.NoError(t, st.Calculate())

		assert.Equal(t, cbc.Key("not-an-event-type"), st.Lines[0].Key)
		assert.True(t, addon.EventType(st.Lines[0]).IsEmpty())
	})

	t.Run("does not migrate the legacy other key", func(t *testing.T) {
		// `other` is both the legacy key for code 90 and the key every line
		// now carries, so migrating it would silently turn a line with a
		// missing event type into code 90. It is reported instead.
		st := legacyStatus(bill.StatusLineOther)
		require.NoError(t, st.Calculate())

		assert.True(t, addon.EventType(st.Lines[0]).IsEmpty())
		faults := rules.Validate(st)
		require.NotNil(t, faults)
		assert.Contains(t, faults.Error(), "a supported event type is required")
	})

	t.Run("migrated document validates", func(t *testing.T) {
		st := legacyStatus("system-startup")
		require.NoError(t, st.Calculate())
		assert.Nil(t, rules.Validate(st))
	})

	t.Run("ignores a nil line", func(t *testing.T) {
		st := validStatus()
		st.Lines = []*bill.StatusLine{nil}
		assert.NotPanics(t, func() {
			_ = st.Calculate()
		})
	})
}

func TestEventTypeWithoutLine(t *testing.T) {
	assert.True(t, addon.EventType(nil).IsEmpty())
}
