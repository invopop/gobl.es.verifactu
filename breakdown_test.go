package verifactu_test

import (
	"testing"
	"time"

	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/invopop/gobl.es.verifactu/test"

	addon "github.com/invopop/gobl/addons/es/verifactu"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/regimes/es"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultBreakdownClient(t *testing.T) *verifactu.Client {
	t.Helper()
	vc, err := verifactu.New(
		verifactu.Software{},
		verifactu.WithCurrentTime(time.Now()),
	)
	require.NoError(t, err)
	return vc
}

func TestBreakdownConversion(t *testing.T) {
	t.Run("basic-invoice", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		require.NoError(t, env.Calculate())

		vc := defaultBreakdownClient(t)
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		dd0 := req.Desglose.DetalleDesglose[0]
		assert.Equal(t, "1800.00", dd0.BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "378.00", dd0.CuotaRepercutida)
		assert.Equal(t, "01", dd0.Impuesto)
		assert.Equal(t, "01", dd0.ClaveRegimen)
		assert.Equal(t, "S1", dd0.CalificacionOperacion)
	})

	t.Run("exempt-taxes", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		inv := env.Extract().(*bill.Invoice)
		inv.Lines = []*bill.Line{
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Price: num.NewAmount(100, 0),
				},
				Taxes: tax.Set{
					&tax.Combo{
						Category: "VAT",
						Ext: tax.MakeExtensions().
							Set(addon.ExtKeyExempt, "E1").
							Set(addon.ExtKeyRegime, "01"),
					},
				},
			},
		}
		require.NoError(t, env.Calculate())
		vc := defaultBreakdownClient(t)
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)
		dd0 := req.Desglose.DetalleDesglose[0]
		assert.Equal(t, "100.00", dd0.BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "01", dd0.Impuesto)
		assert.Equal(t, "E1", dd0.OperacionExenta)
		assert.Empty(t, dd0.CuotaRepercutida)
	})

	t.Run("multiple-tax-rates", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		inv := env.Extract().(*bill.Invoice)
		inv.Lines = []*bill.Line{
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Price: num.NewAmount(100, 0),
				},
				Taxes: tax.Set{
					&tax.Combo{
						Category: "VAT",
						Rate:     "standard",
						Ext: tax.MakeExtensions().
							Set(addon.ExtKeyOpClass, "S1").
							Set(addon.ExtKeyRegime, "01"),
					},
				},
			},
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Price: num.NewAmount(50, 0),
				},
				Taxes: tax.Set{
					&tax.Combo{
						Category: "VAT",
						Rate:     "reduced",
						Ext: tax.MakeExtensions().
							Set(addon.ExtKeyOpClass, "S1"),
					},
				},
			},
		}
		require.NoError(t, env.Calculate())
		vc := defaultBreakdownClient(t)
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		dd0 := req.Desglose.DetalleDesglose[0]
		assert.Equal(t, "100.00", dd0.BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "21.00", dd0.CuotaRepercutida)
		assert.Equal(t, "01", dd0.Impuesto)
		assert.Equal(t, "01", dd0.ClaveRegimen)

		dd1 := req.Desglose.DetalleDesglose[1]
		assert.Equal(t, "50.00", dd1.BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "5.00", dd1.CuotaRepercutida)
		assert.Equal(t, "01", dd1.Impuesto)
		assert.Equal(t, "01", dd1.ClaveRegimen)
	})

	t.Run("not-subject-taxes", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		inv := env.Extract().(*bill.Invoice)
		inv.Lines = []*bill.Line{
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Price: num.NewAmount(100, 0),
				},
				Taxes: tax.Set{
					&tax.Combo{
						Category: "VAT",
						Rate:     "exempt",
						Ext: tax.MakeExtensions().
							Set(addon.ExtKeyExempt, "E1").
							Set(addon.ExtKeyRegime, "01"),
					},
				},
			},
		}
		require.NoError(t, env.Calculate())
		vc := defaultBreakdownClient(t)
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)
		dd0 := req.Desglose.DetalleDesglose[0]
		assert.Equal(t, "100.00", dd0.BaseImponibleOImporteNoSujeto)
		assert.Empty(t, dd0.CuotaRepercutida)
		assert.Equal(t, "01", dd0.Impuesto)
		assert.Equal(t, "01", dd0.ClaveRegimen)
		assert.Equal(t, "E1", dd0.OperacionExenta)
	})

	t.Run("equivalence-surcharge", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		inv := env.Extract().(*bill.Invoice)
		inv.Lines = []*bill.Line{
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Price: num.NewAmount(100, 0),
				},
				Taxes: tax.Set{
					&tax.Combo{
						Category: "VAT",
						Rate:     "standard+eqs",
						Ext: tax.MakeExtensions().
							Set(addon.ExtKeyOpClass, "S1").
							Set(addon.ExtKeyRegime, "01"),
					},
				},
			},
		}
		require.NoError(t, env.Calculate())
		vc := defaultBreakdownClient(t)
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)
		dd := req.Desglose.DetalleDesglose[0]
		assert.Equal(t, "100.00", dd.BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "21.00", dd.CuotaRepercutida)
		assert.Equal(t, "5.20", dd.CuotaRecargoEquivalencia)
		assert.Equal(t, "01", dd.Impuesto)
		assert.Equal(t, "01", dd.ClaveRegimen)
		assert.Equal(t, "S1", dd.CalificacionOperacion)
	})

	t.Run("ipsi-tax", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		inv := env.Extract().(*bill.Invoice)
		p := num.MakePercentage(10, 2)
		inv.Lines = []*bill.Line{
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Price: num.NewAmount(100, 0),
				},
				Taxes: tax.Set{
					&tax.Combo{
						Category: es.TaxCategoryIPSI,
						Percent:  &p,
						Ext: tax.MakeExtensions().
							Set(addon.ExtKeyOpClass, "S1"),
					},
				},
			},
		}
		require.NoError(t, env.Calculate())
		vc := defaultBreakdownClient(t)
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)
		dd := req.Desglose.DetalleDesglose[0]
		assert.Equal(t, "100.00", dd.BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "10.00", dd.CuotaRepercutida)
		assert.Equal(t, "02", dd.Impuesto)
		assert.Empty(t, dd.ClaveRegimen)
		assert.Equal(t, "S1", dd.CalificacionOperacion)
	})

	t.Run("antiques", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		inv := env.Extract().(*bill.Invoice)
		inv.Lines = []*bill.Line{
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Price: num.NewAmount(1000, 0),
				},
				Taxes: tax.Set{
					&tax.Combo{
						Category: "VAT",
						Rate:     "reduced",
						Ext: tax.MakeExtensions().
							Set(addon.ExtKeyOpClass, "S1").
							Set(addon.ExtKeyRegime, "04"),
					},
				},
			},
		}
		require.NoError(t, env.Calculate())
		vc := defaultBreakdownClient(t)
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)
		dd := req.Desglose.DetalleDesglose[0]
		assert.Equal(t, "1000.00", dd.BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "100.00", dd.CuotaRepercutida)
		assert.Equal(t, "01", dd.Impuesto)
		assert.Equal(t, "04", dd.ClaveRegimen)
		assert.Equal(t, "S1", dd.CalificacionOperacion)
	})

	t.Run("reverse-charge", func(t *testing.T) {
		env := test.LoadEnvelope("inv-rev-charge.json")
		require.NoError(t, env.Calculate())
		vc := defaultBreakdownClient(t)
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)
		dd := req.Desglose.DetalleDesglose[0]
		assert.Equal(t, "1800.00", dd.BaseImponibleOImporteNoSujeto)
		assert.Equal(t, "0.00", dd.CuotaRepercutida)
		assert.Equal(t, "01", dd.Impuesto)
		assert.Equal(t, "01", dd.ClaveRegimen)
		assert.Equal(t, "S2", dd.CalificacionOperacion)
		assert.Equal(t, "0", dd.TipoImpositivo)
	})

	t.Run("foreign VAT rates (OSS)", func(t *testing.T) {
		env := test.LoadEnvelope("inv-eu-b2c.json")
		require.NoError(t, env.Calculate())
		vc := defaultBreakdownClient(t)
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)
		dd := req.Desglose.DetalleDesglose[0]
		assert.Equal(t, "1800.00", dd.BaseImponibleOImporteNoSujeto)
		assert.Empty(t, dd.CuotaRepercutida)
		assert.Equal(t, "01", dd.Impuesto)
		assert.Equal(t, "17", dd.ClaveRegimen)
		assert.Equal(t, "N2", dd.CalificacionOperacion)
		assert.Empty(t, dd.TipoImpositivo)
	})
}
