package verifactu

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/invopop/gobl.es.verifactu/test"
	"github.com/invopop/gobl/num"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSigningClient(t *testing.T) *Client {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, "2024-11-26T04:00:00Z")
	require.NoError(t, err)

	cert := test.Certificate(t)
	c, err := New(
		Software{
			NombreRazon:              "My Software",
			NIF:                      "12345678A",
			NombreSistemaInformatico: "My Software",
			IdSistemaInformatico:     "A1",
			Version:                  "1.0",
			NumeroInstalacion:        "12345678A",
		},
		WithCurrentTime(ts),
		WithCertificate(cert),
		WithSigning(),
	)
	require.NoError(t, err)
	return c
}

func TestSignDocument(t *testing.T) {
	cert := test.Certificate(t)

	reg := &InvoiceRegistration{
		SUM1:      SUM1,
		IDVersion: CurrentVersion,
		IDFactura: &IDFactura{
			IDEmisorFactura:        "A28083806",
			NumSerieFactura:        "SAMPLE-001",
			FechaExpedicionFactura: "11-11-2024",
		},
		NombreRazonEmisor:        "Test Company",
		TipoFactura:              "F1",
		DescripcionOperacion:     "Test operation",
		Desglose:                 &Desglose{},
		CuotaTotal:               num.MakeAmount(2100, 2),
		ImporteTotal:             num.MakeAmount(12100, 2),
		Encadenamiento:           &Encadenamiento{PrimerRegistro: "S"},
		SistemaInformatico:       &Software{},
		FechaHoraHusoGenRegistro: "2024-11-20T19:00:55+01:00",
		TipoHuella:               FingerprintType,
		Huella:                   "0000000000000000000000000000000000000000000000000000000000000000",
	}

	sig, err := SignDocument(reg, cert)
	require.NoError(t, err)
	require.NotNil(t, sig)

	require.NotNil(t, sig.SignedInfo)
	assert.Equal(t,
		"http://www.w3.org/TR/2001/REC-xml-c14n-20010315",
		sig.SignedInfo.CanonicalizationMethod.Algorithm,
	)
	assert.Equal(t,
		"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
		sig.SignedInfo.SignatureMethod.Algorithm,
	)

	require.Len(t, sig.SignedInfo.Reference, 2)

	docRef := sig.SignedInfo.Reference[0]
	assert.Equal(t, "", docRef.URI)
	require.NotNil(t, docRef.Transforms)
	var transformAlgs []string
	for _, tr := range docRef.Transforms.Transform {
		transformAlgs = append(transformAlgs, tr.Algorithm)
	}
	assert.Contains(t, transformAlgs, "http://www.w3.org/2000/09/xmldsig#enveloped-signature")
	assert.Equal(t, "http://www.w3.org/2001/04/xmlenc#sha256", docRef.DigestMethod.Algorithm)

	propsRef := sig.SignedInfo.Reference[1]
	assert.Equal(t, "http://uri.etsi.org/01903#SignedProperties", propsRef.Type)
	assert.Equal(t, "http://www.w3.org/2001/04/xmlenc#sha256", propsRef.DigestMethod.Algorithm)

	require.NotNil(t, sig.KeyInfo)
	require.NotNil(t, sig.KeyInfo.X509Data)
	assert.NotEmpty(t, sig.KeyInfo.X509Data.X509Certificate)
	require.NotNil(t, sig.KeyInfo.KeyValue)
	require.NotNil(t, sig.KeyInfo.KeyValue.RSA)
	assert.NotEmpty(t, sig.KeyInfo.KeyValue.RSA.Modulus)
	assert.NotEmpty(t, sig.KeyInfo.KeyValue.RSA.Exponent)

	require.NotNil(t, sig.Object)
	require.NotNil(t, sig.Object.QualifyingProperties)
	require.NotNil(t, sig.Object.QualifyingProperties.SignedProperties)
}

func TestSignDocumentPolicy(t *testing.T) {
	cert := test.Certificate(t)

	reg := &InvoiceRegistration{
		SUM1:      SUM1,
		IDVersion: CurrentVersion,
		IDFactura: &IDFactura{
			IDEmisorFactura:        "A28083806",
			NumSerieFactura:        "SAMPLE-001",
			FechaExpedicionFactura: "11-11-2024",
		},
		NombreRazonEmisor:        "Test Company",
		TipoFactura:              "F1",
		DescripcionOperacion:     "Test operation",
		Desglose:                 &Desglose{},
		CuotaTotal:               num.MakeAmount(2100, 2),
		ImporteTotal:             num.MakeAmount(12100, 2),
		Encadenamiento:           &Encadenamiento{PrimerRegistro: "S"},
		SistemaInformatico:       &Software{},
		FechaHoraHusoGenRegistro: "2024-11-20T19:00:55+01:00",
		TipoHuella:               FingerprintType,
		Huella:                   "0000000000000000000000000000000000000000000000000000000000000000",
	}

	sig, err := SignDocument(reg, cert)
	require.NoError(t, err)

	data, err := xml.Marshal(sig)
	require.NoError(t, err)
	xmlStr := string(data)

	assert.Contains(t, xmlStr, "urn:oid:2.16.724.1.3.1.1.2.1.9")
	assert.Contains(t, xmlStr, "http://www.w3.org/2000/09/xmldsig#sha1")
	assert.Contains(t, xmlStr, "G7roucf600+f03r/o0bAOQ6WAs0=")
	assert.Contains(t, xmlStr, "https://sede.administracion.gob.es/politica_de_firma_anexo_1.pdf")

	assert.Contains(t, xmlStr, "urn:oid:1.2.840.10003.5.109.10")
	assert.Contains(t, xmlStr, "text/xml")
	assert.Contains(t, xmlStr, "UTF-8")
}

func TestRegisterInvoiceWithSigning(t *testing.T) {
	c := testSigningClient(t)
	env := test.LoadEnvelope("inv-base.json")

	reg, err := c.RegisterInvoice(env, nil)
	require.NoError(t, err)
	require.NotNil(t, reg.Signature)

	data, err := reg.Bytes()
	require.NoError(t, err)
	assert.Contains(t, string(data), "<ds:Signature")
}

func TestRegisterInvoiceWithoutSigning(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2024-11-26T04:00:00Z")
	require.NoError(t, err)

	c, err := New(Software{}, WithCurrentTime(ts))
	require.NoError(t, err)

	env := test.LoadEnvelope("inv-base.json")

	reg, err := c.RegisterInvoice(env, nil)
	require.NoError(t, err)
	assert.Nil(t, reg.Signature)

	data, err := reg.Bytes()
	require.NoError(t, err)
	assert.NotContains(t, string(data), "<ds:Signature")
}

func TestCancelInvoiceWithSigning(t *testing.T) {
	c := testSigningClient(t)
	env := test.LoadEnvelope("cred-note-base.json")

	can, err := c.CancelInvoice(env, nil)
	require.NoError(t, err)
	require.NotNil(t, can.Signature)

	data, err := can.Bytes()
	require.NoError(t, err)
	assert.Contains(t, string(data), "<ds:Signature")
}
