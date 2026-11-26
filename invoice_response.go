package verifactu

import "github.com/nbio/xml"

// Default response status
const (
	StatusCorrect            string = "Correcto"
	StatusAcceptedWithErrors string = "AceptadoConErrores"
	StatusCancelled          string = "Anulada"
	StatusIncorrect          string = "Incorrecto"
)

// OpType defines the type of operation that was performed.
type OpType string

// Operation types.
const (
	OpTypeRegistration OpType = "Alta"
	OpTypeCancellation OpType = "Anulacion"
)

// InvoiceResponse defines the response fields from the VeriFactu gateway.
type InvoiceResponse struct {
	XMLName xml.Name `xml:"RespuestaRegFactuSistemaFacturacion"`
	Header  struct {
		Issuer             InvoiceResponseIssuer               `xml:"ObligadoEmision"`
		Representative     *InvoiceResponseIssuer              `xml:"Representante,omitempty"`
		RemisionVoluntaria *InvoiceResponseVoluntarySubmission `xml:"sum1:RemisionVoluntaria,omitempty"`
	} `xml:"Cabecera"`
	Wait   int                    `xml:"TiempoEsperaEnvio"`
	Status string                 `xml:"EstadoEnvio"`
	Lines  []*InvoiceResponseLine `xml:"RespuestaLinea"`
}

// InvoiceResponseIssuer maps the response from the invoice request
// for the issuer.
type InvoiceResponseIssuer struct {
	Name string `xml:"NombreRazon"`
	NIF  string `xml:"NIF"`
}

// InvoiceResponseVoluntarySubmission for additional header data.
type InvoiceResponseVoluntarySubmission struct {
	Date     string `xml:"FechaFinVerifactu,omitempty"`
	Incident string `xml:"Incidencia,omitempty"`
}

// InvoiceResponseLine defines the contents of a single line of the request.
type InvoiceResponseLine struct {
	XMLName xml.Name `xml:"RespuestaLinea"`
	ID      struct {
		Issuer string `xml:"IDEmisorFactura"`
		Code   string `xml:"NumSerieFactura"`
		Date   string `xml:"FechaExpedicionFactura"`
	} `xml:"IDFactura"`
	Operation struct {
		Type             OpType `xml:"TipoOperacion"`
		Correction       string `xml:"Subsanacion,omitempty"`
		RejectedPrevious string `xml:"RechazoPrevio,omitempty"`
		NoPrevious       string `xml:"SinRegistroPrevio,omitempty"`
	} `xml:"Operacion"`
	Ref         string `xml:"RefExterna,omitempty"`
	Status      string `xml:"EstadoRegistro"`
	Code        string `xml:"CodigoErrorRegistro,omitempty"`
	Description string `xml:"DescripcionErrorRegistro,omitempty"`

	// Duplicated contains information about possible duplicated requests
	// that may need to be handled differently.
	Duplicated *InvoiceResponseLineDuplicated `xml:"RegistroDuplicado,omitempty"`
}

// InvoiceResponseLineDuplicated describes details about the detected duplicate.
type InvoiceResponseLineDuplicated struct {
	ID          string `xml:"IdPeticionRegistroDuplicado"`
	Status      string `xml:"EstadoRegistroDuplicado"`
	Code        string `xml:"CodigoErrorRegistro,omitempty"`
	Description string `xml:"DescripcionErrorRegistro,omitempty"`
}

// Error provides a specific error repsonse for the individual line so that it
// can be handled correctly by the consumer.
func (r *InvoiceResponseLine) Error() error {
	switch r.Status {
	case StatusCorrect, StatusCancelled:
		// This is the response with want
		return nil
	case StatusAcceptedWithErrors:
		return ErrWarning.WithCode(r.Code).WithMessage(r.Message())
	default:
		if r.Duplicated != nil {
			return ErrDuplicate
		}
		return ErrValidation.WithCode(r.Code).WithMessage(r.Message())
	}
}

// Message provides the description of the issue reported by the gateway, or
// the status of the line when no description was provided. The status is not
// included alongside the description as it is already available separately.
func (r *InvoiceResponseLine) Message() string {
	if r.Description != "" {
		return r.Description
	}
	return r.Status
}

// Bytes returns the XML document bytes.
func (ir *InvoiceResponse) Bytes() ([]byte, error) {
	return toBytes(ir)
}

// BytesIndent returns the indented XML document bytes.
func (ir *InvoiceResponse) BytesIndent() ([]byte, error) {
	return toBytesIndent(ir)
}
