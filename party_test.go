package verifactu_test

import (
	"testing"
	"time"

	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/invopop/gobl.es.verifactu/test"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParty(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2022-02-01T04:00:00Z")
	require.NoError(t, err)
	vc, err := verifactu.New(
		verifactu.Software{}, // no software
		verifactu.WithCurrentTime(ts),
	)
	require.NoError(t, err)

	t.Run("with tax ID", func(t *testing.T) {
		env := test.LoadEnvelope("inv-base.json")
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		p := req.Destinatarios[0].IDDestinatario
		assert.Equal(t, "Sample Consumer", p.NombreRazon)
		assert.Equal(t, "B63272603", p.NIF)
		assert.Nil(t, p.IDOtro)
	})

	t.Run("with passport", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.Customer = &org.Party{
			Name: "Mr. Pass Port",
			Identities: []*org.Identity{
				{
					Key:  org.IdentityKeyPassport,
					Code: "12345",
				},
			},
		}
		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		p := req.Destinatarios[0].IDDestinatario
		assert.Equal(t, "Mr. Pass Port", p.NombreRazon)
		assert.Empty(t, p.NIF)
		assert.NotNil(t, p.IDOtro)
		assert.Equal(t, "03", p.IDOtro.IDType)
		assert.Equal(t, "12345", p.IDOtro.ID)
	})

	t.Run("with EU identity", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.Customer = &org.Party{
			Name: "Foreign Company",
			TaxID: &tax.Identity{
				Country: "DE",
				Code:    "111111125",
			},
		}

		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		p := req.Destinatarios[0].IDDestinatario

		assert.Equal(t, "Foreign Company", p.NombreRazon)
		assert.Empty(t, p.NIF)
		require.NotNil(t, p.IDOtro)
		assert.Equal(t, "02", p.IDOtro.IDType)
		assert.Equal(t, "DE111111125", p.IDOtro.ID)
	})

	t.Run("without EU identity", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.Customer = &org.Party{
			Name: "Foreign Company",
			TaxID: &tax.Identity{
				Country: "GB",
				Code:    "123456789",
			},
		}

		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		p := req.Destinatarios[0].IDDestinatario

		assert.Equal(t, "Foreign Company", p.NombreRazon)
		assert.Empty(t, p.NIF)
		require.NotNil(t, p.IDOtro)
		assert.Equal(t, "04", p.IDOtro.IDType)
		assert.Equal(t, "123456789", p.IDOtro.ID)
	})

	t.Run("with no identifiers", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.Customer = &org.Party{
			Name: "Simple Company",
		}

		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)
		assert.Empty(t, req.Destinatarios)
	})

	t.Run("with greek identity", func(t *testing.T) {
		env, inv := test.LoadInvoice("inv-base.json")
		inv.Customer = &org.Party{
			Name: "Foreign Company",
			TaxID: &tax.Identity{
				Country: "EL",
				Code:    "925667500",
			},
		}

		req, err := vc.RegisterInvoice(env, nil)
		require.NoError(t, err)

		p := req.Destinatarios[0].IDDestinatario

		assert.Equal(t, "Foreign Company", p.NombreRazon)
		assert.Empty(t, p.NIF)
		require.NotNil(t, p.IDOtro)
		assert.Equal(t, "GR", p.IDOtro.CodigoPais)
		assert.Equal(t, "EL925667500", p.IDOtro.ID)
	})
}
