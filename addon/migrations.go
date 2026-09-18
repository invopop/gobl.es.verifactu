package addon

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
)

// Status Line Key Migration
//
// Until the event type moved to the `es-verifactu-event-type` extension, each
// NO VERI*FACTU event type had its own `bill.StatusLine` key. Core GOBL now
// restricts status line keys to a fixed set describing the situation of a
// business document, so those keys no longer validate. This file maps the old
// keys onto the extension so that previously issued documents keep working.

// Legacy status line keys, one per event type.
const (
	legacyKeySystemStartup        cbc.Key = "system-startup"
	legacyKeySystemShutdown       cbc.Key = "system-shutdown"
	legacyKeyInvoiceAnomalyLaunch cbc.Key = "invoice-anomaly-launch"
	legacyKeyInvoiceAnomaly       cbc.Key = "invoice-anomaly"
	legacyKeyEventAnomalyLaunch   cbc.Key = "event-anomaly-launch"
	legacyKeyEventAnomaly         cbc.Key = "event-anomaly"
	legacyKeyBackupRestoration    cbc.Key = "backup-restoration"
	legacyKeyInvoiceExport        cbc.Key = "invoice-export"
	legacyKeyEventExport          cbc.Key = "event-export"
	legacyKeyEventSummary         cbc.Key = "event-summary"
)

// legacyStatusLineKeys maps each legacy status line key to its event type
// code.
//
// The legacy `other` key, which meant code 90, is deliberately absent. It is
// indistinguishable from the `other` key that every line now carries, so
// migrating it would mean treating any line whose event type extension is
// missing as code 90 — silently turning an invalid document into a valid one
// with the wrong event type. Legacy `other` events are instead reported by
// rule 03, and must set the extension explicitly.
var legacyStatusLineKeys = map[cbc.Key]cbc.Code{
	legacyKeySystemStartup:        CodeSystemStartup,
	legacyKeySystemShutdown:       CodeSystemShutdown,
	legacyKeyInvoiceAnomalyLaunch: CodeInvoiceAnomalyLaunch,
	legacyKeyInvoiceAnomaly:       CodeInvoiceAnomaly,
	legacyKeyEventAnomalyLaunch:   CodeEventAnomalyLaunch,
	legacyKeyEventAnomaly:         CodeEventAnomaly,
	legacyKeyBackupRestoration:    CodeBackupRestoration,
	legacyKeyInvoiceExport:        CodeInvoiceExport,
	legacyKeyEventExport:          CodeEventExport,
	legacyKeyEventSummary:         CodeEventSummary,
}

// migrateStatusLine moves a legacy event type key into the event type
// extension, leaving the line with the `other` key that core GOBL expects.
//
// Lines that already carry the extension are left alone, so a document that
// has been migrated once is never rewritten, and an explicit event type always
// wins over the one implied by the key.
func migrateStatusLine(line *bill.StatusLine) {
	if line == nil || !line.Ext.Get(ExtKeyEventType).IsEmpty() {
		return
	}
	code, ok := legacyStatusLineKeys[line.Key]
	if !ok {
		return
	}
	line.Key = bill.StatusLineOther
	line.Ext = line.Ext.Set(ExtKeyEventType, code)
}
