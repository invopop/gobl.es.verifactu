package verifactu

import (
	"bytes"
	"fmt"

	"github.com/nbio/xml"
)

// Record contains the chain details read back from a record, to check its fingerprint, which covers only some fields and not the signature.
type Record struct {
	// Fingerprint as stated in the record.
	Fingerprint string
	// PreviousFingerprint as stated in the record, empty for the first record of a chain.
	PreviousFingerprint string
	// ComputedFingerprint is recomputed from the fields the fingerprint covers, differing when any of them was altered.
	ComputedFingerprint string
	// GenerationTimestamp is when the record was generated, with its time zone.
	GenerationTimestamp string
}

// ParseRecord reads the chain details of an invoice registration, cancellation or event record.
func ParseRecord(data []byte) (*Record, error) {
	name, err := recordName(data)
	if err != nil {
		return nil, errParsingRecord(err)
	}
	var r interface{ record() (*Record, error) }
	switch name {
	case "RegistroAlta":
		r = new(InvoiceRegistration)
	case "RegistroAnulacion":
		r = new(InvoiceCancellation)
	case "RegistroEvento":
		r = new(EventRegistration)
	default:
		return nil, ErrOnlyRecords
	}
	if err := xml.Unmarshal(data, r); err != nil {
		return nil, errParsingRecord(err)
	}
	return r.record()
}

// recordName reads the name of the record's root element, without reading the rest of the record.
func recordName(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		if el, ok := tok.(xml.StartElement); ok {
			return el.Name.Local, nil
		}
	}
}

func errParsingRecord(err error) error {
	return ErrValidation.WithMessage(fmt.Errorf("parsing record: %w", err).Error())
}
