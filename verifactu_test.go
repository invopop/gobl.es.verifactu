package verifactu_test

import (
	"testing"
	"time"

	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/invopop/gobl.es.verifactu/addon"
	"github.com/invopop/gobl.es.verifactu/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvoiceConversion(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2022-02-01T04:00:00Z")
	require.NoError(t, err)
	vc, err := verifactu.New(
		verifactu.Software{},
		verifactu.WithCurrentTime(ts),
	)
	require.NoError(t, err)

	t.Run("should contain basic document info", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		irq, err := vc.NewEnvelopeInvoiceRequest(env, nil)
		require.NoError(t, err)

		head := irq.Header
		req := irq.Lines[0].Registration
		assert.Equal(t, "Invopop S.L.", head.Obligado.NombreRazon)
		assert.Equal(t, "B85905495", head.Obligado.NIF)
		assert.Equal(t, "1.0", req.IDVersion)
		assert.Equal(t, "B85905495", req.IDFactura.IDEmisorFactura)
		assert.Equal(t, "SAMPLE-004", req.IDFactura.NumSerieFactura)
		assert.Equal(t, "13-11-2024", req.IDFactura.FechaExpedicionFactura)
	})
}

func TestInvoiceConversionWithInstallationNumber(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2022-02-01T04:00:00Z")
	require.NoError(t, err)
	soft := verifactu.Software{}
	vc, err := verifactu.New(
		soft,
		verifactu.WithCurrentTime(ts),
	)
	require.NoError(t, err)

	t.Run("should contain basic document info", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		irq, err := vc.NewEnvelopeInvoiceRequest(env, nil, verifactu.WithInstallationNumber("TEST1"))
		require.NoError(t, err)

		head := irq.Header
		req := irq.Lines[0].Registration
		assert.Equal(t, "Invopop S.L.", head.Obligado.NombreRazon)
		assert.Equal(t, "B85905495", head.Obligado.NIF)
		assert.Equal(t, "1.0", req.IDVersion)
		assert.Equal(t, "B85905495", req.IDFactura.IDEmisorFactura)
		assert.Equal(t, "SAMPLE-004", req.IDFactura.NumSerieFactura)
		assert.Equal(t, "13-11-2024", req.IDFactura.FechaExpedicionFactura)
		assert.Equal(t, "TEST1", req.SistemaInformatico.NumeroInstalacion)
		assert.Empty(t, soft.NumeroInstalacion, "leave original untouched")
	})
}

func TestInvoiceConversionWithRep(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2022-02-01T04:00:00Z")
	require.NoError(t, err)

	vc, err := verifactu.New(
		verifactu.Software{},
		verifactu.WithCurrentTime(ts),
		verifactu.WithRepresentative("Sample Rep", "B63272603"),
	)
	require.NoError(t, err)

	t.Run("should contain basic document info", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		irq, err := vc.NewEnvelopeInvoiceRequest(env, nil)
		require.NoError(t, err)

		head := irq.Header
		req := irq.Lines[0].Registration
		assert.Equal(t, "Invopop S.L.", head.Obligado.NombreRazon)
		assert.Equal(t, "B85905495", head.Obligado.NIF)
		assert.Equal(t, "Sample Rep", head.Representante.NombreRazon)
		assert.Equal(t, "B63272603", head.Representante.NIF)
		assert.Equal(t, "1.0", req.IDVersion)
		assert.Equal(t, "B85905495", req.IDFactura.IDEmisorFactura)
		assert.Equal(t, "SAMPLE-004", req.IDFactura.NumSerieFactura)
		assert.Equal(t, "13-11-2024", req.IDFactura.FechaExpedicionFactura)
	})
}

func TestRegisterInvoiceQRStamp(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2022-02-01T04:00:00Z")
	require.NoError(t, err)

	tests := []struct {
		name     string
		opts     []verifactu.Option
		expected string
	}{
		{
			name:     "sandbox",
			opts:     []verifactu.Option{verifactu.InSandbox()},
			expected: "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR?nif=B85905495&numserie=SAMPLE-004&fecha=13-11-2024&importe=2178.00",
		},
		{
			name:     "production",
			opts:     []verifactu.Option{verifactu.InProduction()},
			expected: "https://www2.agenciatributaria.gob.es/wlpl/TIKE-CONT/ValidarQR?nif=B85905495&numserie=SAMPLE-004&fecha=13-11-2024&importe=2178.00",
		},
		{
			name:     "no verifactu sandbox",
			opts:     []verifactu.Option{verifactu.InSandbox(), verifactu.NoVerifactu()},
			expected: "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQRNoVerifactu?nif=B85905495&numserie=SAMPLE-004&fecha=13-11-2024&importe=2178.00",
		},
		{
			name:     "no verifactu production",
			opts:     []verifactu.Option{verifactu.InProduction(), verifactu.NoVerifactu()},
			expected: "https://www2.agenciatributaria.gob.es/wlpl/TIKE-CONT/ValidarQRNoVerifactu?nif=B85905495&numserie=SAMPLE-004&fecha=13-11-2024&importe=2178.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := append([]verifactu.Option{verifactu.WithCurrentTime(ts)}, tt.opts...)
			vc, err := verifactu.New(verifactu.Software{}, opts...)
			require.NoError(t, err)

			env := test.LoadEnvelope("inv-base.json")
			_, err = vc.RegisterInvoice(env, nil)
			require.NoError(t, err)

			stamp := env.Head.GetStamp(addon.StampQR)
			require.NotNil(t, stamp)
			assert.Equal(t, tt.expected, stamp.Value)
		})
	}
}
