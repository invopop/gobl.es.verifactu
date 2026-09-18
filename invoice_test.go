package verifactu_test

import (
	"testing"
	"time"

	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/invopop/gobl.es.verifactu/addon"
	"github.com/invopop/gobl.es.verifactu/test"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistroAlta(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2022-02-01T04:00:00Z")
	require.NoError(t, err)
	vc, err := verifactu.New(
		verifactu.Software{},
		verifactu.WithCurrentTime(ts),
	)
	require.NoError(t, err)

	t.Run("should contain basic document info", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "1.0", ra.IDVersion)
		assert.Equal(t, "B85905495", ra.IDFactura.IDEmisorFactura)
		assert.Equal(t, "SAMPLE-004", ra.IDFactura.NumSerieFactura)
		assert.Equal(t, "13-11-2024", ra.IDFactura.FechaExpedicionFactura)
		assert.Equal(t, "Invopop S.L.", ra.NombreRazonEmisor)
		assert.Equal(t, "F1", ra.TipoFactura)
		assert.Equal(t, "This is a sample invoice with a standard tax", ra.DescripcionOperacion)
		assert.Equal(t, "378.00", ra.CuotaTotal.String())
		assert.Equal(t, "2178.00", ra.ImporteTotal.String())
		assert.Empty(t, ra.FechaOperacion)

		require.Len(t, ra.Destinatarios, 1)
		dest := ra.Destinatarios[0].IDDestinatario
		assert.Equal(t, "Sample Consumer", dest.NombreRazon)
		assert.Equal(t, "B63272603", dest.NIF)

		require.Len(t, ra.Desglose.DetalleDesglose, 1)
		desg := ra.Desglose.DetalleDesglose[0]
		assert.Equal(t, "01", desg.Impuesto)
		assert.Equal(t, "01", desg.ClaveRegimen)
		assert.Equal(t, "S1", desg.CalificacionOperacion)
		assert.Equal(t, "21.0", desg.TipoImpositivo)
		assert.Equal(t, "1800.00", desg.BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "378.00", desg.CuotaRepercutida)
	})
	t.Run("should handle simplified invoices", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.SetTags(tax.TagSimplified)
		inv.Customer = nil
		require.NoError(t, inv.Calculate())

		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "S", ra.FacturaSinIdentifDestinatarioArt61d)
	})

	t.Run("should handle rectificative invoices", func(t *testing.T) {
		env := test.LoadEnvelope("cred-note-base.json")

		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "R1", ra.TipoFactura)
		assert.Equal(t, "I", ra.TipoRectificativa)
		require.Len(t, ra.FacturasRectificadas.Items, 1)

		rectified := ra.FacturasRectificadas.Items[0]
		assert.Equal(t, "B85905495", rectified.IDEmisorFactura)
		assert.Equal(t, "SAMPLE-085", rectified.NumSerieFactura)
		assert.Equal(t, "10-01-2022", rectified.FechaExpedicionFactura)
		assert.Equal(t, "-1620.00", ra.Desglose.DetalleDesglose[0].BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "-340.20", ra.Desglose.DetalleDesglose[0].CuotaRepercutida)
		assert.Equal(t, "-340.20", ra.CuotaTotal.String())
		assert.Equal(t, "-1960.20", ra.ImporteTotal.String())
	})

	t.Run("should handle substitution invoices", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.Preceding = []*org.DocumentRef{
			{
				Series:    "SAMPLE",
				Code:      "002",
				IssueDate: cal.NewDate(2024, 1, 15),
			},
		}
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyDocType, "F3")

		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		require.Len(t, ra.FacturasSustituidas.Items, 1)
		substituted := ra.FacturasSustituidas.Items[0]
		assert.Equal(t, "B85905495", substituted.IDEmisorFactura)
		assert.Equal(t, "SAMPLE-002", substituted.NumSerieFactura)
		assert.Equal(t, "15-01-2024", substituted.FechaExpedicionFactura)
	})

	t.Run("should handle an empty note", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.Notes = nil
		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "1.0", ra.IDVersion)
		assert.Equal(t, "B85905495", ra.IDFactura.IDEmisorFactura)
		assert.Equal(t, "SAMPLE-004", ra.IDFactura.NumSerieFactura)
		assert.Equal(t, "13-11-2024", ra.IDFactura.FechaExpedicionFactura)
		assert.Equal(t, "Invopop S.L.", ra.NombreRazonEmisor)
		assert.Equal(t, "Development services.", ra.DescripcionOperacion)
	})

	t.Run("should handle an empty note with multiple items", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.Notes = nil
		inv.Lines[0].Item.Name = "a"
		for range int(100) {
			inv.Lines = append(inv.Lines, inv.Lines[0])
		}
		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "1.0", ra.IDVersion)
		assert.Equal(t, "B85905495", ra.IDFactura.IDEmisorFactura)
		assert.Equal(t, "SAMPLE-004", ra.IDFactura.NumSerieFactura)
		assert.Equal(t, "13-11-2024", ra.IDFactura.FechaExpedicionFactura)
		assert.Equal(t, "Invopop S.L.", ra.NombreRazonEmisor)
		assert.True(t, len(ra.DescripcionOperacion) <= 500)
	})

	t.Run("should handle an operation date", func(t *testing.T) {
		env := test.LoadEnvelope("inv-op-date.json")
		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "10-11-2024", ra.FechaOperacion)
	})
}

func TestPartialBreakdownRegime(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2022-02-01T04:00:00Z")
	require.NoError(t, err)
	vc, err := verifactu.New(
		verifactu.Software{},
		verifactu.WithCurrentTime(ts),
	)
	require.NoError(t, err)

	t.Run("REBU keeps untaxed charge in total", func(t *testing.T) {
		env := test.LoadEnvelope("inv-rebu.json")
		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "10.00", ra.ImporteTotal.String())
		assert.Equal(t, "0.17", ra.CuotaTotal.String())
		require.Len(t, ra.Desglose.DetalleDesglose, 1)
		assert.Equal(t, "03", ra.Desglose.DetalleDesglose[0].ClaveRegimen)
		assert.Equal(t, "0.83", ra.Desglose.DetalleDesglose[0].BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "0.17", ra.Desglose.DetalleDesglose[0].CuotaRepercutida)
	})

	t.Run("REBU with outlay subtracts only the outlay", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-rebu.json")
		inv.Charges = append(inv.Charges, &bill.Charge{
			Key:    bill.ChargeKeyOutlay,
			Reason: "Notary fees",
			Amount: num.MakeAmount(5000, 2),
		})
		require.NoError(t, inv.Calculate())

		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		// 10.00 (total_with_tax) + 50.00 (outlay) - 50.00 (outlay subtracted) = 10.00
		assert.Equal(t, "10.00", ra.ImporteTotal.String())
	})

	t.Run("REBU with multiple untaxed charges keeps all non-outlay", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-rebu.json")
		inv.Charges = append(inv.Charges, &bill.Charge{
			Reason: "Additional cost component",
			Amount: num.MakeAmount(2000, 2),
		})
		require.NoError(t, inv.Calculate())

		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		// Both untaxed charges (9.00 + 20.00) stay in total
		assert.Equal(t, "30.00", ra.ImporteTotal.String())
	})

	t.Run("non-partial regime subtracts all untaxed charges", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base-outlay.json")
		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		// Regime 01: outlay (100.00) subtracted from total_with_tax (2278.00)
		assert.Equal(t, "2178.00", ra.ImporteTotal.String())
		assert.Equal(t, "01", ra.Desglose.DetalleDesglose[0].ClaveRegimen)
	})

	t.Run("non-partial regime subtracts untaxed charge without key", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.Charges = []*bill.Charge{
			{
				Reason: "Some non-taxable fee",
				Amount: num.MakeAmount(10000, 2),
			},
		}
		require.NoError(t, inv.Calculate())

		ra, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		// Regime 01: untaxed charge (100.00) subtracted from total_with_tax (2278.00)
		assert.Equal(t, "2178.00", ra.ImporteTotal.String())
	})
}
