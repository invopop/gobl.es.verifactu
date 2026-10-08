package verifactu_test

import (
	"encoding/xml"
	"testing"

	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvelopeResponseFault(t *testing.T) {
	t.Run("header error", func(t *testing.T) {
		data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
			<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/">
				<env:Body>
					<env:Fault>
						<faultcode>env:Client</faultcode>
						<faultstring>Codigo[4104].Error en la cabecera: el valor del campo NIF del bloque ObligadoEmision no est&#225; identificado..
						NIF:54387763P. NOMBRE_RAZON:Sample Consumer
						</faultstring>
					</env:Fault>
				</env:Body>
			</env:Envelope>`)

		res := new(verifactu.EnvelopeResponse)
		require.NoError(t, xml.Unmarshal(data, res))
		require.NotNil(t, res.Body.Fault)

		err := res.Body.Fault.Err()
		assert.ErrorIs(t, err, verifactu.ErrValidation)
		assert.Equal(t, "4104", err.Code())
		assert.Equal(t, "Error en la cabecera: el valor del campo NIF del bloque ObligadoEmision no está identificado.. NIF:54387763P. NOMBRE_RAZON:Sample Consumer", err.Message())
	})

	t.Run("server side fault", func(t *testing.T) {
		f := &verifactu.Fault{
			Code:    "env:Server",
			Message: "Codigo[401].No hay conexion a @firma afirma6.aeat",
		}
		err := f.Err()
		assert.ErrorIs(t, err, verifactu.ErrServer)
		assert.Equal(t, "401", err.Code())
		assert.Equal(t, "No hay conexion a @firma afirma6.aeat", err.Message())
	})

	t.Run("alternative fault code prefix", func(t *testing.T) {
		f := &verifactu.Fault{
			Code:    "soapenv:Server",
			Message: "Internal error",
		}
		err := f.Err()
		assert.ErrorIs(t, err, verifactu.ErrServer)
		assert.Empty(t, err.Code())
		assert.Equal(t, "Internal error", err.Message())
	})

	t.Run("without code", func(t *testing.T) {
		f := &verifactu.Fault{
			Code:    "env:Client",
			Message: "Error en la cabecera",
		}
		err := f.Err()
		assert.ErrorIs(t, err, verifactu.ErrValidation)
		assert.Empty(t, err.Code())
		assert.Equal(t, "Error en la cabecera", err.Message())
	})

	t.Run("code without trailing period", func(t *testing.T) {
		f := &verifactu.Fault{
			Code:    "env:Client",
			Message: "Codigo[4112] Error certificado",
		}
		err := f.Err()
		assert.Equal(t, "4112", err.Code())
		assert.Equal(t, "Error certificado", err.Message())
	})

	t.Run("code not at start of message", func(t *testing.T) {
		f := &verifactu.Fault{
			Code:    "env:Client",
			Message: "Error en la cabecera Codigo[4104]",
		}
		err := f.Err()
		assert.Empty(t, err.Code())
		assert.Equal(t, "Error en la cabecera Codigo[4104]", err.Message())
	})
}
