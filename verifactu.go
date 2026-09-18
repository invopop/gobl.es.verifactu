// Package verifactu provides the VeriFactu client
package verifactu

import (
	"context"
	"fmt"
	"time"

	"github.com/invopop/gobl"
	noverifactu "github.com/invopop/gobl.es.verifactu/pkg/noverifactu"
	"github.com/invopop/gobl/addons/es/verifactu"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/head"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/org"
	"github.com/invopop/xmldsig"
)

const (
	// StampKeyHash defines the key used to store the Hash in the envelope
	// stamps.
	StampKeyHash cbc.Key = "verifactu-hash"
)

// Client provides the main interface to the VeriFactu package.
type Client struct {
	software Software
	env      Environment
	rep      *Issuer
	curTime  time.Time
	cert     *xmldsig.Certificate
	conn     *connection
	withSeal bool
	signing  bool
	signOpts []xmldsig.Option
}

// Option is used to configure the client.
type Option func(*Client)

// WithCertificate defines the signing certificate to use when producing the
// VeriFactu document.
func WithCertificate(cert *xmldsig.Certificate) Option {
	return func(c *Client) {
		c.cert = cert
	}
}

// WithCurrentTime defines the current time to use when generating the VeriFactu
// document. Only useful for testing.
func WithCurrentTime(curTime time.Time) Option {
	return func(c *Client) {
		c.curTime = curTime
	}
}

// WithRepresentative can be used to define a legal representative
// of the entity legally obliged to issue the invoice, usually the
// supplier. This may be required when the supplier's digital
// certificate is not available.
func WithRepresentative(name, nif string) Option {
	return func(c *Client) {
		c.rep = &Issuer{
			NombreRazon: name,
			NIF:         nif,
		}
	}
}

// InProduction defines the connection to use the production environment.
func InProduction() Option {
	return func(c *Client) {
		c.env = EnvironmentProduction
	}
}

// InSandbox defines the connection to use the testing environment.
func InSandbox() Option {
	return func(c *Client) {
		c.env = EnvironmentSandbox
	}
}

// WithCorporateSeal uses the specific VERI*FACTU endpoints for communicating with
// a legal entity's corporate seal instead of a regular certificate issued to
// individuals. In Spanish these are called "Sello de Entidad".
var WithCorporateSeal Option = func(c *Client) {
	c.withSeal = true
}

// WithSigning sets additional xmldsig options that will be passed
// through to SignDocument for each record. This is useful for controlling
// the document ID and signing time in tests.
func WithSigning(opts ...xmldsig.Option) Option {
	return func(c *Client) {
		c.signing = true
		c.signOpts = opts
	}
}

// New creates a new VeriFactu client with shared software and configuration
// options for creating and sending new documents.
func New(software Software, opts ...Option) (*Client, error) {
	c := new(Client)
	c.software = software

	// Set default values that can be overwritten by the options
	c.env = EnvironmentSandbox

	for _, opt := range opts {
		opt(c)
	}

	if c.cert == nil {
		return c, nil
	}

	if c.conn == nil {
		var err error
		c.conn, err = newConnection(c.env, c.cert, c.withSeal)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// CurrentTime returns the current time to use when generating
// the VeriFactu document.
func (c *Client) CurrentTime() time.Time {
	if !c.curTime.IsZero() {
		return c.curTime
	}
	return time.Now()
}

// Sandbox returns true if the client is using the sandbox environment.
func (c *Client) Sandbox() bool {
	return c.env == EnvironmentSandbox
}

// GenerateOption defines what is expected from an option used in generate
// methods.
type GenerateOption func(*generateOptions)

type generateOptions struct {
	amendment          string
	previouslyRejected string
	noPriorRecord      string
	installNumber      string
}

// Amended indicates that the incoming document is an amendment of a previous
// document, known in Spanish as "Subsanación". If incorrect data was sent, this
// will allow the new document to replace the old.
func Amended() GenerateOption {
	return func(o *generateOptions) {
		o.amendment = "S"
	}
}

// PreviouslyRejected indicates that an earlier attempt to process the document
// was rejected and thus never received by the verifactu system. This option
// automatically implies the Amended option when supported by the document.
func PreviouslyRejected() GenerateOption {
	return func(o *generateOptions) {
		o.amendment = "S"
		o.previouslyRejected = "X"
	}
}

// NoPriorRecord is used with cancellation documents when no previous attempt was
// made to issue an invoice via Verifactu, or if the document to cancel was rejected.
func NoPriorRecord() GenerateOption {
	return func(o *generateOptions) {
		o.previouslyRejected = ""
		o.noPriorRecord = "S"
	}
}

// WithInstallationNumber is used to override the installation number defined in the connection
// configuration while generating the invoice.
func WithInstallationNumber(code string) GenerateOption {
	return func(o *generateOptions) {
		o.installNumber = code
	}
}

// RegisterInvoice prepares a new registration document from the provided invoice
// inside the GOBL envelope. It will fingerprint and update the registration with
// the chaining hash and QR code. The resulting document can be persisted for
// sending later.
func (c *Client) RegisterInvoice(env *gobl.Envelope, prev *ChainData, opts ...GenerateOption) (*InvoiceRegistration, error) {
	o := new(generateOptions)
	for _, cb := range opts {
		cb(o)
	}

	inv, ok := env.Extract().(*bill.Invoice)
	if !ok {
		return nil, ErrOnlyInvoices
	}
	if inv.GetRegime() != l10n.ES.Tax() {
		return nil, ErrNotSpanish
	}

	if inv.Type == bill.InvoiceTypeCreditNote {
		// In VeriFactu credit notes become "facturas rectificativas por diferencias",
		// which require negative totals.
		if err := inv.Invert(); err != nil {
			return nil, err
		}
	}
	software := c.software // clone
	if o.installNumber != "" {
		software.NumeroInstalacion = o.installNumber
	}

	reg, err := newInvoiceRegistration(inv, c.CurrentTime(), &software)
	if err != nil {
		return nil, err
	}
	reg.Subsanacion = o.amendment
	reg.RechazoPrevio = o.previouslyRejected
	reg.fingerprint(prev)
	if c.signing && c.cert != nil {
		sig, err := SignDocument(reg, c.cert, c.signOpts...)
		if err != nil {
			return nil, fmt.Errorf("signing registration: %w", err)
		}
		reg.Signature = sig
	}
	c.addRegistrationStamps(env, reg)

	return reg, nil
}

// CancelInvoice builds a cancellation message from the provided document and previous
// chain data. Note that the cancellation does not require Hash information of the last,
// invoice, and instead only requires the previous chain entry.
func (c *Client) CancelInvoice(env *gobl.Envelope, prev *ChainData, opts ...GenerateOption) (*InvoiceCancellation, error) {
	inv, ok := env.Extract().(*bill.Invoice)
	if !ok {
		return nil, ErrOnlyInvoices
	}

	o := new(generateOptions)
	for _, cb := range opts {
		cb(o)
	}
	software := c.software // clone
	if o.installNumber != "" {
		software.NumeroInstalacion = o.installNumber
	}

	can := newInvoiceCancellation(inv, c.CurrentTime(), &software)
	can.RechazoPrevio = o.previouslyRejected
	can.SinRegistroPrevio = o.noPriorRecord
	can.fingerprint(prev)
	if c.signing && c.cert != nil {
		sig, err := SignDocument(can, c.cert, c.signOpts...)
		if err != nil {
			return nil, fmt.Errorf("signing cancellation: %w", err)
		}
		can.Signature = sig
	}

	return can, nil
}

// NewInvoiceRequest prepares a new invoice request for the provided supplier.
func (c *Client) NewInvoiceRequest(supplier *org.Party) (*InvoiceRequest, error) {
	if supplier == nil || supplier.TaxID == nil {
		return nil, ErrValidation.WithMessage("missing supplier or tax id")
	}
	ir := new(InvoiceRequest)
	ir.Header = &InvoiceRequestHeader{
		Obligado: Issuer{
			NombreRazon: supplier.Name,
			NIF:         supplier.TaxID.Code.String(),
		},
		Representante: c.rep,
	}
	return ir, nil
}

// NewEnvelopeInvoiceRequest is a convenience method that prepares a new InvoiceRequest
// from the GOBL envelope in a single method call.
func (c *Client) NewEnvelopeInvoiceRequest(env *gobl.Envelope, prev *ChainData, opts ...GenerateOption) (*InvoiceRequest, error) {
	req, err := c.RegisterInvoice(env, prev, opts...)
	if err != nil {
		return nil, err
	}
	inv := env.Extract().(*bill.Invoice)
	ir, err := c.NewInvoiceRequest(inv.Supplier)
	if err != nil {
		return nil, err
	}

	ir.AddRegistration(req)
	return ir, nil
}

// SendInvoiceRequest will prepare the final SOAP envelope with the invoice request
// data and send it the agency API.
func (c *Client) SendInvoiceRequest(ctx context.Context, ir *InvoiceRequest) (*InvoiceResponse, error) {
	if len(ir.Lines) == 0 {
		return nil, ErrValidation.WithMessage("no invoice request lines")
	}

	data, err := ir.Envelop().Bytes()
	if err != nil {
		return nil, err
	}

	out, err := c.conn.post(ctx, data)
	if err != nil {
		return nil, err
	}

	if res := out.Body.InvoiceResponse; res == nil {
		return nil, ErrConnection.WithMessage("missing response body")
	}
	return out.Body.InvoiceResponse, nil
}

// RegisterEvent prepares a new event registration document from the provided bill status
// inside the GOBL envelope. It will fingerprint and optionally sign the event. The
// resulting document can be persisted locally.
func (c *Client) RegisterEvent(env *gobl.Envelope, prev *EventChainData, opts ...GenerateOption) (*EventRegistration, error) {
	o := new(generateOptions)
	for _, cb := range opts {
		cb(o)
	}

	status, ok := env.Extract().(*bill.Status)
	if !ok {
		return nil, ErrOnlyStatuses
	}
	if status.GetRegime() != l10n.ES.Tax() {
		return nil, ErrNotSpanish
	}

	if err := noverifactu.Validate(status); err != nil {
		return nil, err
	}

	software := c.software // clone
	if o.installNumber != "" {
		software.NumeroInstalacion = o.installNumber
	}

	reg, err := newEventRegistration(status, c.CurrentTime(), &software)
	if err != nil {
		return nil, fmt.Errorf("creating event registration: %w", err)
	}
	reg.Event.fingerprint(prev)

	if c.signing && c.cert != nil {
		sig, err := SignDocument(reg, c.cert, c.signOpts...)
		if err != nil {
			return nil, fmt.Errorf("signing event registration: %w", err)
		}
		reg.Event.Signature = sig
	}

	c.addEventStamps(env, reg)

	return reg, nil
}

// addEventStamps adds the Hash stamp to the envelope for event registrations.
func (c *Client) addEventStamps(env *gobl.Envelope, reg *EventRegistration) {
	if reg.Event != nil {
		env.Head.AddStamp(&head.Stamp{
			Provider: StampKeyHash,
			Value:    reg.Event.Fingerprint,
		})
	}
}

// addRegistrationStamps adds the QR code stamp and Hash to the envelope.
func (c *Client) addRegistrationStamps(env *gobl.Envelope, reg *InvoiceRegistration) {
	// now generate the QR codes and add them to the envelope
	code := reg.generateURL(c.env == EnvironmentProduction)
	env.Head.AddStamp(&head.Stamp{
		Provider: verifactu.StampQR,
		Value:    code,
	})
	env.Head.AddStamp(&head.Stamp{
		Provider: StampKeyHash,
		Value:    reg.Huella,
	})
}
