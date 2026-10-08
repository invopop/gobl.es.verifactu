package verifactu

import (
	"github.com/invopop/gobl.es.verifactu/addon"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/org"
)

var idTypeCodeMap = map[cbc.Key]cbc.Code{
	org.IdentityKeyPassport: "03",
	org.IdentityKeyForeign:  "04",
	org.IdentityKeyResident: "05",
	org.IdentityKeyOther:    "06",
}

// newParty builds a new party, but only if there are enough tax identification details
func newParty(p *org.Party, date cal.Date) *Party {
	if p == nil {
		return nil
	}
	pty := &Party{
		NombreRazon: p.Name,
	}
	if p.TaxID != nil && !p.TaxID.Code.IsEmpty() && p.TaxID.Country.In("ES") {
		pty.NIF = p.TaxID.Code.String()
	} else {
		pty.IDOtro = otherIdentity(p, date)
	}
	if pty.NIF == "" && pty.IDOtro == nil {
		return nil
	}
	return pty
}

func otherIdentity(p *org.Party, date cal.Date) *IDOtro {
	oid := new(IDOtro)
	if p.TaxID != nil {
		oid.CodigoPais = isoCountryCode(p.TaxID.Country).String()
		if p.TaxID.InEU(date) {
			oid.IDType = "02"         // NIF-VAT
			oid.ID = p.TaxID.String() // with country prefix
		} else {
			oid.IDType = "04"              // foreign
			oid.ID = p.TaxID.Code.String() // just code
		}
		return oid
	}

	for _, id := range p.Identities {
		code := id.Ext.Get(addon.ExtKeyIdentityType)

		// Fallback to matching the key, this should no longer be needed
		if code.IsEmpty() {
			if it, ok := idTypeCodeMap[id.Key]; ok {
				code = it
			}
		}
		if code.IsEmpty() {
			// Nothing to do here as there is no useful identity type
			continue
		}

		oid.CodigoPais = id.Country.String()
		oid.IDType = code.String()
		oid.ID = id.Code.String()

		return oid
	}
	return nil
}

// isoCountryCode turns the tax country code into an ISO version, this is
// required in order to map troublesome countries like Greece from EL to GR.
func isoCountryCode(code l10n.TaxCountryCode) l10n.ISOCountryCode {
	if code == "" {
		return ""
	}
	if def := l10n.Countries().Code(code.Code()); def != nil {
		if !def.ISO {
			return def.AltCode.ISO()
		}
		return def.Code.ISO()
	}
	return ""
}
