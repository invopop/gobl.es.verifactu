package verifactu_test

import (
	"strings"
	"testing"

	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/invopop/gobl.es.verifactu/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRecord(t *testing.T) {
	c := loadClient(t)

	t.Run("invoice registrations in a chain", func(t *testing.T) {
		first, err := c.RegisterInvoice(test.LoadEnvelope("inv-base.json"), nil)
		require.NoError(t, err)
		second, err := c.RegisterInvoice(test.LoadEnvelope("inv-eu-b2c.json"), first.ChainData())
		require.NoError(t, err)

		data, err := first.Bytes()
		require.NoError(t, err)
		rec, err := verifactu.ParseRecord(data)
		require.NoError(t, err)
		assert.Equal(t, first.Huella, rec.Fingerprint)
		assert.Equal(t, rec.Fingerprint, rec.ComputedFingerprint)
		assert.Empty(t, rec.PreviousFingerprint)
		assert.Equal(t, first.FechaHoraHusoGenRegistro, rec.GenerationTimestamp)

		data, err = second.Bytes()
		require.NoError(t, err)
		rec, err = verifactu.ParseRecord(data)
		require.NoError(t, err)
		assert.Equal(t, second.Huella, rec.Fingerprint)
		assert.Equal(t, rec.Fingerprint, rec.ComputedFingerprint)
		assert.Equal(t, first.Huella, rec.PreviousFingerprint)
	})

	t.Run("invoice cancellation", func(t *testing.T) {
		reg, err := c.RegisterInvoice(test.LoadEnvelope("inv-base.json"), nil)
		require.NoError(t, err)
		can, err := c.CancelInvoice(test.LoadEnvelope("inv-base.json"), reg.ChainData())
		require.NoError(t, err)

		data, err := can.Bytes()
		require.NoError(t, err)
		rec, err := verifactu.ParseRecord(data)
		require.NoError(t, err)
		assert.Equal(t, can.Huella, rec.Fingerprint)
		assert.Equal(t, rec.Fingerprint, rec.ComputedFingerprint)
		assert.Equal(t, reg.Huella, rec.PreviousFingerprint)
		assert.Equal(t, can.FechaHoraHusoGenRegistro, rec.GenerationTimestamp)
	})

	t.Run("event records in a chain", func(t *testing.T) {
		first, err := c.RegisterEvent(test.LoadEnvelope("status-system-startup.json"), nil)
		require.NoError(t, err)
		second, err := c.RegisterEvent(test.LoadEnvelope("status-event-summary.json"), first.ChainData())
		require.NoError(t, err)

		data, err := first.Bytes()
		require.NoError(t, err)
		rec, err := verifactu.ParseRecord(data)
		require.NoError(t, err)
		assert.Equal(t, first.Event.Fingerprint, rec.Fingerprint)
		assert.Equal(t, rec.Fingerprint, rec.ComputedFingerprint)
		assert.Empty(t, rec.PreviousFingerprint)
		assert.Equal(t, first.Event.GenerationTimestamp, rec.GenerationTimestamp)

		data, err = second.Bytes()
		require.NoError(t, err)
		rec, err = verifactu.ParseRecord(data)
		require.NoError(t, err)
		assert.Equal(t, second.Event.Fingerprint, rec.Fingerprint)
		assert.Equal(t, rec.Fingerprint, rec.ComputedFingerprint)
		assert.Equal(t, first.Event.Fingerprint, rec.PreviousFingerprint)
	})

	t.Run("altered record", func(t *testing.T) {
		reg, err := c.RegisterInvoice(test.LoadEnvelope("inv-base.json"), nil)
		require.NoError(t, err)
		data, err := reg.Bytes()
		require.NoError(t, err)
		altered := strings.Replace(string(data), "<sum1:ImporteTotal>"+reg.ImporteTotal.String(), "<sum1:ImporteTotal>1.00", 1)
		require.NotEqual(t, string(data), altered)

		rec, err := verifactu.ParseRecord([]byte(altered))
		require.NoError(t, err)
		assert.Equal(t, reg.Huella, rec.Fingerprint)
		assert.NotEqual(t, rec.Fingerprint, rec.ComputedFingerprint)
	})

	t.Run("incomplete records", func(t *testing.T) {
		for _, data := range []string{
			`<sum1:RegistroAlta xmlns:sum1="x"><sum1:Huella>AB</sum1:Huella></sum1:RegistroAlta>`,
			`<sum1:RegistroAnulacion xmlns:sum1="x"><sum1:Huella>AB</sum1:Huella></sum1:RegistroAnulacion>`,
			`<sf:RegistroEvento xmlns:sf="x"><sf:Evento><sf:HuellaEvento>AB</sf:HuellaEvento></sf:Evento></sf:RegistroEvento>`,
		} {
			_, err := verifactu.ParseRecord([]byte(data))
			assert.Same(t, verifactu.ErrIncompleteRecord, err)
		}
		_, err := verifactu.ParseRecord([]byte(`<sf:RegistroEvento xmlns:sf="x"/>`))
		assert.Same(t, verifactu.ErrIncompleteRecord, err)
	})

	t.Run("not a record", func(t *testing.T) {
		_, err := verifactu.ParseRecord([]byte(`<Factura/>`))
		assert.Same(t, verifactu.ErrOnlyRecords, err)
		_, err = verifactu.ParseRecord([]byte(`not xml`))
		assert.ErrorIs(t, err, verifactu.ErrValidation)
	})
}
