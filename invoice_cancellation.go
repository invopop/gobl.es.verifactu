package verifactu

import (
	"time"

	"github.com/invopop/gobl/bill"
	"github.com/invopop/xmldsig"
	"github.com/nbio/xml"
)

// InvoiceCancellation contains the details of an invoice cancellation
type InvoiceCancellation struct {
	XMLName                  xml.Name           `xml:"sum1:RegistroAnulacion"`
	SUM1                     string             `xml:"xmlns:sum1,attr,omitempty"`
	IDVersion                string             `xml:"sum1:IDVersion"`
	IDFactura                *IDFacturaAnulada  `xml:"sum1:IDFactura"`
	RefExterna               string             `xml:"sum1:RefExterna,omitempty"`
	SinRegistroPrevio        string             `xml:"sum1:SinRegistroPrevio,omitempty"`
	RechazoPrevio            string             `xml:"sum1:RechazoPrevio,omitempty"`
	GeneradoPor              string             `xml:"sum1:GeneradoPor,omitempty"`
	Generador                *Party             `xml:"sum1:Generador,omitempty"`
	Encadenamiento           *Encadenamiento    `xml:"sum1:Encadenamiento"`
	SistemaInformatico       *Software          `xml:"sum1:SistemaInformatico"`
	FechaHoraHusoGenRegistro string             `xml:"sum1:FechaHoraHusoGenRegistro"`
	TipoHuella               string             `xml:"sum1:TipoHuella"`
	Huella                   string             `xml:"sum1:Huella"`
	Signature                *xmldsig.Signature `xml:"ds:Signature,omitempty"`
}

// IDFacturaAnulada contains the identifying information for an invoice
type IDFacturaAnulada struct {
	IDEmisorFactura        string `xml:"sum1:IDEmisorFacturaAnulada"`
	NumSerieFactura        string `xml:"sum1:NumSerieFacturaAnulada"`
	FechaExpedicionFactura string `xml:"sum1:FechaExpedicionFacturaAnulada"`
}

// newInvoiceCancellation provides support for cancelling invoices
func newInvoiceCancellation(inv *bill.Invoice, ts time.Time, s *Software) *InvoiceCancellation {
	reg := &InvoiceCancellation{
		SUM1:      SUM1,
		IDVersion: CurrentVersion,
		IDFactura: &IDFacturaAnulada{
			IDEmisorFactura:        inv.Supplier.TaxID.Code.String(),
			NumSerieFactura:        invoiceNumber(inv.Series, inv.Code),
			FechaExpedicionFactura: inv.IssueDate.Time().Format("02-01-2006"),
		},
		SistemaInformatico:       s,
		FechaHoraHusoGenRegistro: formatDateTimeZone(ts),
		TipoHuella:               FingerprintType,
	}
	return reg
}

// fingerprint will add a fingerprint to the cancellation message using the
// ChainData from the last entry.
func (c *InvoiceCancellation) fingerprint(prev *ChainData) {
	h := ""
	if prev == nil {
		c.Encadenamiento = &Encadenamiento{
			PrimerRegistro: "S",
		}
	} else {
		c.Encadenamiento = &Encadenamiento{
			RegistroAnterior: &RegistroAnterior{
				IDEmisorFactura:        prev.IDIssuer,
				NumSerieFactura:        prev.NumSeries,
				FechaExpedicionFactura: prev.IssueDate,
				Huella:                 prev.Fingerprint,
			},
		}
		h = prev.Fingerprint
	}

	c.Huella = computeFingerprint([]string{
		formatChainField("IDEmisorFacturaAnulada", c.IDFactura.IDEmisorFactura),
		formatChainField("NumSerieFacturaAnulada", c.IDFactura.NumSerieFactura),
		formatChainField("FechaExpedicionFacturaAnulada", c.IDFactura.FechaExpedicionFactura),
		formatChainField("Huella", h),
		formatChainField("FechaHoraHusoGenRegistro", c.FechaHoraHusoGenRegistro),
	})
}

// ChainData provides the details for this cancellation entry.
func (c *InvoiceCancellation) ChainData() *ChainData {
	return &ChainData{
		IDIssuer:    c.IDFactura.IDEmisorFactura,
		NumSeries:   c.IDFactura.NumSerieFactura,
		IssueDate:   c.IDFactura.FechaExpedicionFactura,
		Fingerprint: c.Huella,
	}
}

// Bytes prepares an XML document suitable for persistence. Signed documents
// use compact XML to preserve the enveloped signature.
func (c *InvoiceCancellation) Bytes() ([]byte, error) {
	return toBytes(c)
}

// BytesIndent returns the indented XML document bytes, which breaks the signature of a signed document.
func (c *InvoiceCancellation) BytesIndent() ([]byte, error) {
	return toBytesIndent(c)
}
