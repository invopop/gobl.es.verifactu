// Package addon provides the VeriFactu addon for GOBL, covering both the
// VERI*FACTU and NO VERI*FACTU modalities.
package addon

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/norm"
	"github.com/invopop/gobl/pkg/here"
	"github.com/invopop/gobl/rules"
	"github.com/invopop/gobl/rules/is"
	"github.com/invopop/gobl/tax"
)

const (
	// Key identifies the Verifactu addon family. Individual versions append a
	// suffix; the family key is used as the fault-code namespace so that
	// rules that carry across versions keep stable codes.
	Key cbc.Key = "es-verifactu"

	// V1 for Verifactu versions 1.x
	V1 cbc.Key = Key + "-v1"
)

// Official stamps or codes validated by government agencies
const (
	// StampQR contains the URL included in the QR code.
	StampQR cbc.Key = "verifactu-qr"
)

// Fault code namespaces. The two VeriFactu modalities get one each, so that a
// fault makes clear which set of rules rejected the document: invoices are
// reported under ES-VERIFACTU, and NO VERI*FACTU events under ES-NOVERIFACTU.
const (
	codeVeriFactu   rules.Code = "ES-VERIFACTU"
	codeNoVeriFactu rules.Code = "ES-NOVERIFACTU"
)

func init() {
	tax.RegisterAddonDef(newAddon())
	rules.RegisterWithGuard(
		Key.String(),
		rules.GOBL.Add(codeVeriFactu),
		is.InContext(tax.AddonIn(V1)),
		billInvoiceRules(),
		taxComboRules(),
	)
	rules.RegisterWithGuard(
		Key.String(),
		rules.GOBL.Add(codeNoVeriFactu),
		is.InContext(tax.AddonIn(V1)),
		statusRules...,
	)
	// Complement rules are registered without a guard: the complement schemas
	// are defined by this addon and mean nothing outside it, so there is no
	// other document they could wrongly apply to, and validating a complement
	// on its own carries no addon context to guard against.
	rules.Register(
		Key.String(),
		rules.GOBL.Add(codeNoVeriFactu),
		complementRules...,
	)
	norm.RegisterWithGuard(
		is.InContext(tax.AddonIn(V1)),
		norm.For(normalizeBillInvoice),
		norm.For(normalizeTaxCombo),
		norm.For(migrateStatusLine),
	)
}

func newAddon() *tax.AddonDef {
	return &tax.AddonDef{
		Key: V1,
		Name: i18n.String{
			i18n.EN: "Spain VERI*FACTU V1",
		},
		Sources: []*cbc.Source{
			{
				Title: i18n.NewString("VERI*FACTU error response code list"),
				URL:   "https://prewww2.aeat.es/static_files/common/internet/dep/aplicaciones/es/aeat/tikeV1.0/cont/ws/errores.properties",
			},
		},
		Tags: []*tax.TagSet{
			{
				Schema: bill.ShortSchemaInvoice,
				List: []*cbc.Definition{
					{
						Key: tax.TagReplacement,
						Name: i18n.String{
							i18n.EN: "Replacement Invoice",
							i18n.ES: "Factura de Sustitución",
						},
						Desc: i18n.NewString(here.Doc(`
							Used under special circumstances to indicate that this invoice replaces a previously
							issued simplified invoice. The previous document was correct, but the replacement is
							necessary to provide tax details of the customer.
						`)),
					},
				},
			},
		},
		Extensions:  extensions,
		Scenarios:   scenarios,
		Corrections: invoiceCorrectionDefinitions,
	}
}
