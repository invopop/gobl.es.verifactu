package verifactu

import (
	"fmt"
	"net/url"
)

// Base URLs of the AEAT services used to validate invoice QR codes. As defined in the
// AEAT specification for generating the invoice QR code ("Detalle de las especificaciones
// técnicas para generación del código QR de la factura"), verifiable (VERI*FACTU) and
// non-verifiable (NO VERI*FACTU) invoicing systems must each point to their own service.
const (
	testURL            = "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR"
	prodURL            = "https://www2.agenciatributaria.gob.es/wlpl/TIKE-CONT/ValidarQR"
	testNoVerifactuURL = "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQRNoVerifactu"
	prodNoVerifactuURL = "https://www2.agenciatributaria.gob.es/wlpl/TIKE-CONT/ValidarQRNoVerifactu"
)

// generateURL generates the encoded URL code with parameters for the given environment.
func (r *InvoiceRegistration) generateURL(production, noVerifactu bool) string {
	nif := url.QueryEscape(r.IDFactura.IDEmisorFactura)
	numSerie := url.QueryEscape(r.IDFactura.NumSerieFactura)
	fecha := url.QueryEscape(r.IDFactura.FechaExpedicionFactura)
	importe := url.QueryEscape(r.ImporteTotal.String())

	base := qrBaseURL(production, noVerifactu)
	return fmt.Sprintf("%s?nif=%s&numserie=%s&fecha=%s&importe=%s", base, nif, numSerie, fecha, importe)
}

// qrBaseURL returns the AEAT QR validation service URL for the environment and type of
// invoicing system.
func qrBaseURL(production, noVerifactu bool) string {
	switch {
	case production && noVerifactu:
		return prodNoVerifactuURL
	case production && !noVerifactu:
		return prodURL
	case !production && noVerifactu:
		return testNoVerifactuURL
	default: // !production && !noVerifactu
		return testURL
	}
}
