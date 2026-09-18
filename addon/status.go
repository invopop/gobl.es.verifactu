package addon

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/rules"
	"github.com/invopop/gobl/rules/is"
	"github.com/invopop/gobl/schema"
	"github.com/invopop/gobl/tax"
)

var statusRules = []*rules.Set{
	statusRuleSet(),
}

// codeToComplement maps each event type to the complement it must carry, or
// nil when the event type takes no complement.
var codeToComplement = map[cbc.Code]any{
	CodeSystemStartup:        nil,
	CodeSystemShutdown:       nil,
	CodeInvoiceAnomalyLaunch: InvoiceAnomalyLaunch{},
	CodeInvoiceAnomaly:       InvoiceAnomaly{},
	CodeEventAnomalyLaunch:   EventAnomalyLaunch{},
	CodeEventAnomaly:         EventAnomaly{},
	CodeBackupRestoration:    nil,
	CodeInvoiceExport:        InvoiceExport{},
	CodeEventExport:          EventExport{},
	CodeEventSummary:         EventSummary{},
	CodeOther:                nil,
}

// eventTypeCodes lists the codes accepted by the event type extension. Taken
// from the definition rather than the registry so that the rule set does not
// depend on registration having already happened.
func eventTypeCodes() []cbc.Code {
	return cbc.DefinitionCodes(extEventType.Values)
}

// EventType returns the L2E event type code carried by the given status line,
// or an empty code when none is present.
func EventType(line *bill.StatusLine) cbc.Code {
	if line == nil {
		return cbc.CodeEmpty
	}
	return line.Ext.Get(ExtKeyEventType)
}

var isAnomalyDetection = is.Func("anomaly detection event", func(v any) bool {
	line, _ := v.(*bill.StatusLine)
	code := EventType(line)
	return code == CodeInvoiceAnomaly || code == CodeEventAnomaly
})

var isOtherNote = is.Func("other note", func(v any) bool {
	note, _ := v.(*org.Note)
	return note != nil && note.Key == org.NoteKeyOther
})

var requiresComplement = is.Func("requires complement", func(v any) bool {
	line, _ := v.(*bill.StatusLine)
	return codeToComplement[EventType(line)] != nil
})

var hasCorrectComplement = is.Func("correct complement schema", func(v any) bool {
	line, _ := v.(*bill.StatusLine)
	if line == nil || len(line.Complements) == 0 {
		return true // other rules handle missing info
	}

	s := schema.Lookup(codeToComplement[EventType(line)])
	if s == schema.UnknownID {
		return true // other rules handle unsupported event types
	}

	return line.Complements[0].Schema == s
})

func statusRuleSet() *rules.Set {
	return rules.For(new(bill.Status),
		rules.Field("lines",
			rules.Assert("01", "status must have exactly one line", is.Present, is.Length(1, 1)),
			rules.Each(
				rules.Field("key",
					rules.Assert("02", "line key must be 'other'", is.In(bill.StatusLineOther)),
				),
				rules.Field("ext",
					rules.Assert("03", "a supported event type is required",
						tax.ExtensionsRequire(ExtKeyEventType),
						tax.ExtensionsHasCodes(ExtKeyEventType, eventTypeCodes()...),
					),
				),
				rules.When(requiresComplement,
					rules.Field("complements",
						rules.Assert("04", "complement is required for this event type", is.Present, is.Length(1, 1)),
					),
					rules.Assert("05", "complement must correspond to the event type", hasCorrectComplement),
				),
				rules.When(isAnomalyDetection,
					rules.Field("description",
						rules.AssertIfPresent("06", "description must be 100 characters or less", is.Length(0, 100)),
					),
				),
			),
		),
		rules.Field("notes",
			rules.Each(
				rules.When(isOtherNote,
					rules.Field("text",
						rules.AssertIfPresent("07", "other note text must be 100 characters or less", is.Length(0, 100)),
					),
				),
			),
		),
	)
}
