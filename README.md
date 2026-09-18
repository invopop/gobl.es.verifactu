# GOBL to VeriFactu

Go library to convert [GOBL](https://github.com/invopop/gobl) invoices into VeriFactu declarations and send them to the AEAT (Agencia Estatal de Administración Tributaria) web services. This library assumes that clients will handle a local database of previous invoices in order to comply with the local requirements of chaining all invoices together.

Released under the Apache 2.0 [LICENSE](https://github.com/invopop/gobl/blob/main/LICENSE), Copyright 2021-2025 [Invopop S.L.](https://invopop.com).

[![Lint](https://github.com/invopop/gobl.es.verifactu/actions/workflows/lint.yaml/badge.svg)](https://github.com/invopop/gobl.es.verifactu/actions/workflows/lint.yaml)
[![Test Go](https://github.com/invopop/gobl.es.verifactu/actions/workflows/test.yaml/badge.svg)](https://github.com/invopop/gobl.es.verifactu/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/invopop/gobl.es.verifactu)](https://goreportcard.com/report/github.com/invopop/gobl.es.verifactu)
[![codecov](https://codecov.io/gh/invopop/gobl.es.verifactu/graph/badge.svg)](https://codecov.io/gh/invopop/gobl.es.verifactu)
[![GoDoc](https://godoc.org/github.com/invopop/gobl.es.verifactu?status.svg)](https://godoc.org/github.com/invopop/gobl.es.verifactu)
![Latest Tag](https://img.shields.io/github/v/tag/invopop/gobl.es.verifactu)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/invopop/gobl.es.verifactu)

## Addon

This module provides the `es-verifactu-v1` GOBL addon, in [`addon/`](./addon). It
defines the extensions, normalizers, scenarios and validation rules that adapt a
GOBL document to VeriFactu, for both the VERI*FACTU and NO VERI*FACTU modalities.

The addon lived in GOBL core until it moved here, alongside the converter that is
its main consumer. It registers itself on import, so a consumer that processes
documents declaring `es-verifactu-v1` needs a blank import:

```go
import _ "github.com/invopop/gobl.es.verifactu/addon"
```

This is required, not optional: GOBL refuses to `Calculate` or `Validate` a
document whose `$addons` names an addon that is not loaded, so that a document is
never processed without the rules its `$addons` promise.

### NO VERI*FACTU events

Event registrations are built from a `bill.Status` document. Core GOBL status line
keys describe the situation of a business document, so a NO VERI*FACTU event uses
the `other` key and carries its event type in the `es-verifactu-event-type`
extension, holding a code from the L2E list:

```yaml
lines:
  - key: "other"
    ext:
      es-verifactu-event-type: "01" # invoicing system startup
```

Until this changed, each event type had its own status line key
(`system-startup`, `invoice-anomaly`, and so on). Those keys are migrated onto
the extension automatically during `Calculate`, so previously issued documents
keep working. The one exception is the legacy `other` key, which meant code
`90`: it cannot be told apart from the `other` key every line now carries, so it
is reported rather than guessed at, and must set the extension explicitly.

### Fault codes

All rules report under a single `GOBL-ES-VERIFACTU` namespace, with the set
component saying which rules rejected the document:

| Example | Rules |
| --- | --- |
| `GOBL-ES-VERIFACTU-BILL-INVOICE-02` | Invoices |
| `GOBL-ES-VERIFACTU-BILL-STATUS-01` | Status events |
| `GOBL-ES-VERIFACTU-INVOICEANOMALY-01` | Event complements |

Invoice processing is identical under VERI*FACTU and NO VERI*FACTU, so there is
no namespace per modality: it would imply a split the invoice rules do not have.
Status documents are the only modality-specific surface, and `BILL-STATUS`
already identifies them.

The status and complement rules previously reported under a bare `NOVERIFACTU`
namespace, from a rule set held outside the addon that only this library
invoked. The set and assertion parts are unchanged, so
`NOVERIFACTU-BILL-STATUS-01` is now `GOBL-ES-VERIFACTU-BILL-STATUS-01`.
Anything matching on the old codes needs updating.

## Source

English is used for the principal actions inside this package, however a large amount of code is still in Spanish. We're working on translating, but this project may still be combersome to use if you do not understand Spanish.

The main resources used in this module include:

- [VeriFactu documentation](https://www.agenciatributaria.es/AEAT.desarrolladores/Desarrolladores/_menu_/Documentacion/Sistemas_Informaticos_de_Facturacion_y_Sistemas_VERI_FACTU/Sistemas_Informaticos_de_Facturacion_y_Sistemas_VERI_FACTU.html)
- [VeriFactu Ministerial Order](https://www.boe.es/diario_boe/txt.php?id=BOE-A-2024-22138)
- [Spanish VAT Law](https://www.boe.es/buscar/act.php?id=BOE-A-1992-28740).

## Usage

### Go Package

You must have first created a GOBL Envelope containing an Invoice that you'd like to send to the AEAT. For the document to be converted, the supplier contained in the invoice should have a "Tax ID" with the country set to `ES`.

The following is an example of how the GOBL VeriFactu package could be used:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/gobl"
	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/invopop/xmldsig"
)

func main() {
	ctx := context.Background()

	// Load sample envelope:
	data, _ := os.ReadFile("./test/data/sample-invoice.json")
	env := new(gobl.Envelope)
	if err := json.Unmarshal(data, env); err != nil {
		panic(err)
	}

	// VeriFactu requires a software definition to be provided. This is an example
	software := &verifactu.Software{
		NombreRazon:              "Company S.L.",    // Company Name
		NIF:                      "B123456789",     // Software company's tax code
		NombreSistemaInformatico: "Software Name",  // Name of application
		IdSistemaInformatico:     "A1",             // Software ID
		Version:                  "1.0",            // Software version
		NumeroInstalacion:        "00001",          // Software installation number
	}

	// Load the certificate
	cert, err := xmldsig.LoadCertificate(
		"./path/to/certificate.p12",
		"password",
	)
	if err != nil {
		panic(err)
	}

	// Create the client with the software and certificate
	opts := []verifactu.Option{
		verifactu.WithCertificate(cert),
		verifactu.InTesting(),          // Use the testing environment
	}
	vc, err := verifactu.New(software, opts...)
	if err != nil {
		panic(err)
	}

	// Prepare previous chain data. Ideally you want to do this from
	// your own database.
	prevData, err := os.ReadFile("./path/to/previous_invoice.json")
	if err != nil {
		panic(err)
	}
	prev := new(verifactu.ChainData)
	if err := json.Unmarshal([]byte(prevData), prev); err != nil {
		panic(err)
	}

	// Generate a registration document and update the envelope.
	reg, err := vc.RegisterInvoice(env)
	if err != nil {
		panic(err)
	}

	inv := env.(*bill.Invoice)
	ir, err := vc.InvoiceRequest(inv.Supplier)
	if err != nil {
		panic(err)
	}
	inv.AddRegistration(reg)

	// Send the document to the tax agency
	out, err = vc.SendInvoiceRequest(ctx, ir)
	if err != nil {
		panic(err)
	}
	line := out.Lines[0]

	// Print the data to be used as previous document chain for the next invoice
	// Persist the data somewhere to be used by the next invoice
	cd, err := json.Marshal(line.ChainData())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Generated document with fingerprint: \n%s\n", string(cd))

}
```

### Command Line

The GOBL VeriFactu package tool also includes a command line helper. You can install manually in your Go environment with:

```bash
go install github.com/invopop/gobl.es.verifactu/cmd/gobl.verifactu@latest
```

We recommend using a `.env` file to prepare configuration settings, although all parameters can be set using command line flags. Heres an example:

```
SOFTWARE_COMPANY_NIF=B85905495
SOFTWARE_COMPANY_NAME=Invopop S.L.
SOFTWARE_NAME=gobl.verifactu
SOFTWARE_VERSION=1.0
SOFTWARE_ID_SISTEMA_INFORMATICO=A1
SOFTWARE_NUMERO_INSTALACION=00001
DEBUG=true
CERTIFICATE_PATH=xxxxxxxxx
CERTIFICATE_PASSWORD=xxxxxxxxx
```

To convert a document to XML, run:

```bash
gobl.verifactu convert ./test/data/sample-invoice.json
```

This function will output the XML to the terminal, or to a file if a second argument is provided. The output file will not include a fingerprint, and therefore will not be able to be submitted to the tax agency.

To submit to the tax agency testing environment:

```bash
gobl.verifactu send ./test/data/sample-invoice.json ./test/data/previous-invoice-info.json
```

Now, the output file will include a fingerprint, linked to the previous document, and will be submitted to the tax agency. An example for a previous file would look like this:

```json
{
  "emisor": "B12345678",
  "serie": "SAMPLE-001",
  "fecha": "2024-11-28",
  "huella": "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
}
```

## Tax Extensions

The following extensions must be included in the document. Note that the GOBL addon will automatically add these extensions when processing invoices:

- `es-verifactu-doc-type` – defines the type of invoice being sent. In most cases this will be set automatically by GOBL, but it must be present. It is taken from list L2 of the VeriFactu Ministerial Order. These are the valid values:

  - `F1` - Standard invoice.
  - `F2` - Simplified invoice.
  - `F3` - Invoice in substitution of simplified invoices.
  - `R1` - Rectified invoice based on law and Article 80.1, 80.2 and 80.6 in the Spanish VAT Law.
  - `R2` - Rectified invoice based on law and Article 80.3.
  - `R3` - Rectified invoice based on law and Article 80.4.
  - `R4` - Rectified invoice based on law and other reasons.
  - `R5` - Rectified invoice based on simplified invoices.

- `es-verifactu-op-class` - Operation classification code used to identify if taxes should be applied to the line. It is taken from list L9 of the VeriFactu Ministerial Order. These are the valid values:

  - `S1` - Subject and not exempt - Without reverse charge
  - `S2` - Subject and not exempt - With reverse charge. Known as `Inversión del Sujeto Pasivo` in Spanish VAT Law
  - `N1` - Not subject - Articles 7, 14, others
  - `N2` - Not subject - Due to location rules

- `es-verifactu-exempt` - Exemption code used to identify if the line item is exempt from taxes. It is taken from list L10 of the VeriFactu Ministerial Order. These are the valid values:

  - `E1` - Exempt pursuant to Article 20 of the VAT Law
  - `E2` - Exempt pursuant to Article 21 of the VAT Law
  - `E3` - Exempt pursuant to Article 22 of the VAT Law
  - `E4` - Exempt pursuant to Articles 23 and 24 of the VAT Law
  - `E5` - Exempt pursuant to Article 25 of the VAT Law
  - `E6` - Exempt for other reasons

- `es-verifactu-correction-type` - Differentiates between the correction method. Corrective invoices in VeriFactu can be _Facturas Rectificativas por Diferencias_ or _Facturas Rectificativas por Sustitución_. It is taken from list L3 of the VeriFactu Ministerial Order. These are the valid values:

  - `I` - Differences. Used for credit and debit notes. In case of credit notes the values of the invoice are inverted to reflect the amount being a credit instead of a debit.
  - `S` - Substitution. Used for corrective invoices.

- `es-verifactu-regime` - Regime code used to identify the type of VAT/IGIC regime to be applied to the invoice. It combines the values of lists L8A and L8B of the VeriFactu Ministerial Order. These are the valid values:
  - `01` - General regime operation
  - `02` - Export
  - `03` - Special regime for used goods, art objects, antiques and collectibles
  - `04` - Special regime for investment gold
  - `05` - Special regime for travel agencies
  - `06` - Special regime for VAT/IGIC groups (Advanced Level)
  - `07` - Special cash accounting regime
  - `08` - Operations subject to a different regime
  - `09` - Billing of travel agency services acting as mediators in name and on behalf of others
  - `10` - Collection of professional fees or rights on behalf of third parties
  - `11` - Business premises rental operations
  - `14` - Invoice with pending VAT/IGIC accrual in work certifications for Public Administration
  - `15` - Invoice with pending VAT/IGIC accrual in successive tract operations
  - `17` - Operation under OSS and IOSS regimes (VAT) / Special regime for retail traders. (IGIC)
  - `18` - Equivalence surcharge (VAT) / Special regime for small traders or retailers (IGIC)
  - `19` - Operations included in the Special Regime for Agriculture, Livestock and Fisheries
  - `20` - Simplified regime (VAT only)

## Limitations

- VeriFactu allows more than one customer per invoice, but GOBL only has one possible customer.
- Invoices must have a note of type general that will be used as a general description of the invoice. If an invoice is missing this info, it will be rejected with an error.
- VeriFactu supports sending more than one invoice at a time (up to 1000). However, this module only currently supports 1 invoice at a time.
- VeriFactu requires a valid certificate to be provided, even when using the testing environment. It is the same certificate needed to access the AEAT's portal.
- When cancelling invoices, this module assumes the party issuing the cancellation is the same as the party that issued the original invoice. In the context of the app this would always be true, but VeriFactu does allow for a different issuer.

## Testing

This library includes a set of tests that can be used to validate the conversion and submission process. To run the tests, use the following command:

```bash
go test ./...
```

Some sample test data is available in the `./test` directory. To update the JSON documents and regenerate the XML files for testing, use the following command:

```bash
go test --update
```
