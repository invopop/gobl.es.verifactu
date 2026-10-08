package verifactu

import (
	"testing"
	"time"

	"github.com/invopop/gobl.es.verifactu/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistroAnulacion(t *testing.T) {
	vc, err := New(
		Software{},
		WithCurrentTime(time.Now()),
	)
	require.NoError(t, err)

	t.Run("basic", func(t *testing.T) {
		env := test.LoadEnvelope("cred-note-base.json")

		ra, err := vc.CancelInvoice(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "B85905495", ra.IDFactura.IDEmisorFactura)
		assert.Equal(t, "FR-012", ra.IDFactura.NumSerieFactura)
		assert.Equal(t, "01-02-2022", ra.IDFactura.FechaExpedicionFactura)
		assert.Equal(t, "01", ra.TipoHuella)
	})
}

func TestInvoiceCancellationFingerprint(t *testing.T) {
	t.Run("Anulacion", func(t *testing.T) {
		tests := []struct {
			name      string
			anulacion *InvoiceCancellation
			prev      *ChainData
			expected  string
		}{
			{
				name: "Basic 1",
				anulacion: &InvoiceCancellation{
					IDFactura: &IDFacturaAnulada{
						IDEmisorFactura:        "A28083806",
						NumSerieFactura:        "SAMPLE-001",
						FechaExpedicionFactura: "11-11-2024",
					},
					FechaHoraHusoGenRegistro: "2024-11-21T10:00:55+01:00",
				},
				prev: &ChainData{
					IDIssuer:    "A28083806",
					NumSeries:   "SAMPLE-000",
					IssueDate:   "10-11-2024",
					Fingerprint: "4B0A5C1D3F28E6A79B8C2D1E0F3A4B5C6D7E8F9A0B1C2D3E4F5A6B7C8D9E0F1",
				},
				expected: "F5AB85A94450DF8752F4A7840C72456B753010E5EC1F26D8EAE0D4523E287948",
			},
			{
				name: "Basic 2",
				anulacion: &InvoiceCancellation{
					IDFactura: &IDFacturaAnulada{
						IDEmisorFactura:        "B08194359",
						NumSerieFactura:        "SAMPLE-002",
						FechaExpedicionFactura: "12-11-2024",
					},
					FechaHoraHusoGenRegistro: "2024-11-21T12:00:55+01:00",
				},
				prev: &ChainData{
					IDIssuer:    "A28083806",
					NumSeries:   "SAMPLE-001",
					IssueDate:   "11-11-2024",
					Fingerprint: "CBA051CBF59488B6978FA66E95ED4D0A84A97F5C0700EA952B923BD6E7C3FD7A",
				},
				expected: "E86A5172477A636958B2F98770FB796BEEDA43F3F1C6A1C601EC3EEDF9C033B1",
			},
			{
				name: "No Previous",
				anulacion: &InvoiceCancellation{
					IDFactura: &IDFacturaAnulada{
						IDEmisorFactura:        "A28083806",
						NumSerieFactura:        "SAMPLE-001",
						FechaExpedicionFactura: "11-11-2024",
					},
					FechaHoraHusoGenRegistro: "2024-11-21T10:00:55+01:00",
					Encadenamiento:           &Encadenamiento{PrimerRegistro: "S"},
				},
				prev:     nil,
				expected: "A166B0391BCE34DA3A5B022837D0C426F7A4E2F795EBB4581B7BD79E74BCAA95",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.anulacion.fingerprint(tt.prev)
				if got := tt.anulacion.Huella; got != tt.expected {
					t.Errorf("fingerprint = %v, want %v", got, tt.expected)
				}
			})
		}
	})
}
