package verifactu

import (
	"fmt"
	"strconv"
	"time"

	"github.com/invopop/gobl.es.verifactu/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/org"
	"github.com/invopop/xmldsig"
	"github.com/nbio/xml"
)

// XML namespaces
const (
	SF = "https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/EventosSIF.xsd"
)

// EventRegistration represents the root element of a RegistroEvento document.
type EventRegistration struct {
	XMLName xml.Name `xml:"sf:RegistroEvento"`
	SF      string   `xml:"xmlns:sf,attr,omitempty"`
	Version string   `xml:"sf:IDVersion"`
	Event   *Event   `xml:"sf:Evento"`
}

// Event contains the event data (EventoType).
type Event struct {
	Software                      *EventSoftware     `xml:"sf:SistemaInformatico"`
	Issuer                        *EventIssuer       `xml:"sf:ObligadoEmision"`
	IssuedByThirdPartyOrRecipient string             `xml:"sf:EmitidaPorTerceroODestinatario,omitempty"`
	ThirdPartyOrRecipient         *EventParty        `xml:"sf:TerceroODestinatario,omitempty"`
	GenerationTimestamp           string             `xml:"sf:FechaHoraHusoGenEvento"`
	EventType                     string             `xml:"sf:TipoEvento"`
	EventData                     *EventData         `xml:"sf:DatosPropiosEvento,omitempty"`
	OtherEventData                string             `xml:"sf:OtrosDatosEvento,omitempty"`
	Chaining                      *EventChaining     `xml:"sf:Encadenamiento"`
	FingerprintType               string             `xml:"sf:TipoHuella"`
	Fingerprint                   string             `xml:"sf:HuellaEvento"`
	Signature                     *xmldsig.Signature `xml:"ds:Signature,omitempty"`
}

// EventSoftware contains the details about the software system that generated
// the event (SistemaInformaticoType).
type EventSoftware struct {
	Name               string        `xml:"sf:NombreRazon"`
	NIF                string        `xml:"sf:NIF,omitempty"`
	IDOther            *EventOtherID `xml:"sf:IDOtro,omitempty"`
	SoftwareName       string        `xml:"sf:NombreSistemaInformatico,omitempty"`
	SoftwareID         string        `xml:"sf:IdSistemaInformatico"`
	Version            string        `xml:"sf:Version"`
	InstallationNumber string        `xml:"sf:NumeroInstalacion"`
	OnlyVerifactu      string        `xml:"sf:TipoUsoPosibleSoloVerifactu,omitempty"`
	MultiOT            string        `xml:"sf:TipoUsoPosibleMultiOT,omitempty"`
	MultipleOT         string        `xml:"sf:IndicadorMultiplesOT,omitempty"`
}

// EventIssuer represents a Spanish person or entity obligated to issue invoices
// (PersonaFisicaJuridicaESType) in the events namespace.
type EventIssuer struct {
	Name string `xml:"sf:NombreRazon"`
	NIF  string `xml:"sf:NIF"`
}

// EventParty represents a person or entity that may be Spanish or foreign
// (PersonaFisicaJuridicaType) in the events namespace.
type EventParty struct {
	Name    string        `xml:"sf:NombreRazon"`
	NIF     string        `xml:"sf:NIF,omitempty"`
	IDOther *EventOtherID `xml:"sf:IDOtro,omitempty"`
}

// EventOtherID contains alternative identification for non-NIF entities
// (IDOtroType) in the events namespace.
type EventOtherID struct {
	CountryCode string `xml:"sf:CodigoPais,omitempty"`
	IDType      string `xml:"sf:IDType"`
	ID          string `xml:"sf:ID"`
}

// EventData contains event-specific data (DatosPropiosEventoType).
// Only one of the fields should be set at a time (choice).
type EventData struct {
	InvoiceAnomalyDetectionLaunch *InvoiceAnomalyDetectionLaunch `xml:"sf:LanzamientoProcesoDeteccionAnomaliasRegFacturacion,omitempty"`
	InvoiceAnomalyDetection       *InvoiceAnomalyDetection       `xml:"sf:DeteccionAnomaliasRegFacturacion,omitempty"`
	EventAnomalyDetectionLaunch   *EventAnomalyDetectionLaunch   `xml:"sf:LanzamientoProcesoDeteccionAnomaliasRegEvento,omitempty"`
	EventAnomalyDetection         *EventAnomalyDetection         `xml:"sf:DeteccionAnomaliasRegEvento,omitempty"`
	InvoiceExportPeriod           *InvoiceExportPeriod           `xml:"sf:ExportacionRegFacturacionPeriodo,omitempty"`
	EventExportPeriod             *EventExportPeriod             `xml:"sf:ExportacionRegEventoPeriodo,omitempty"`
	EventSummary                  *EventSummary                  `xml:"sf:ResumenEventos,omitempty"`
}

// InvoiceAnomalyDetectionLaunch contains data for launching the anomaly
// detection process on invoice registration records.
type InvoiceAnomalyDetectionLaunch struct {
	FingerprintCheck string `xml:"sf:RealizadoProcesoSobreIntegridadHuellasRegFacturacion"`
	FingerprintCount string `xml:"sf:NumeroDeRegistrosFacturacionProcesadosSobreIntegridadHuellas,omitempty"`
	SignatureCheck   string `xml:"sf:RealizadoProcesoSobreIntegridadFirmasRegFacturacion"`
	SignatureCount   string `xml:"sf:NumeroDeRegistrosFacturacionProcesadosSobreIntegridadFirmas,omitempty"`
	ChainCheck       string `xml:"sf:RealizadoProcesoSobreTrazabilidadCadenaRegFacturacion"`
	ChainCount       string `xml:"sf:NumeroDeRegistrosFacturacionProcesadosSobreTrazabilidadCadena,omitempty"`
	DateCheck        string `xml:"sf:RealizadoProcesoSobreTrazabilidadFechasRegFacturacion"`
	DateCount        string `xml:"sf:NumeroDeRegistrosFacturacionProcesadosSobreTrazabilidadFechas,omitempty"`
}

// InvoiceAnomalyDetection contains data for a detected anomaly in invoice
// registration records.
type InvoiceAnomalyDetection struct {
	AnomalyType      string              `xml:"sf:TipoAnomalia"`
	OtherAnomalyData string              `xml:"sf:OtrosDatosAnomalia,omitempty"`
	AnomalousInvoice *EventInvoiceRecord `xml:"sf:RegistroFacturacionAnomalo,omitempty"`
}

// EventAnomalyDetectionLaunch contains data for launching the anomaly
// detection process on event registration records.
type EventAnomalyDetectionLaunch struct {
	FingerprintCheck string `xml:"sf:RealizadoProcesoSobreIntegridadHuellasRegEvento"`
	FingerprintCount string `xml:"sf:NumeroDeRegistrosEventoProcesadosSobreIntegridadHuellas,omitempty"`
	SignatureCheck   string `xml:"sf:RealizadoProcesoSobreIntegridadFirmasRegEvento"`
	SignatureCount   string `xml:"sf:NumeroDeRegistrosEventoProcesadosSobreIntegridadFirmas,omitempty"`
	ChainCheck       string `xml:"sf:RealizadoProcesoSobreTrazabilidadCadenaRegEvento"`
	ChainCount       string `xml:"sf:NumeroDeRegistrosEventoProcesadosSobreTrazabilidadCadena,omitempty"`
	DateCheck        string `xml:"sf:RealizadoProcesoSobreTrazabilidadFechasRegEvento"`
	DateCount        string `xml:"sf:NumeroDeRegistrosEventoProcesadosSobreTrazabilidadFechas,omitempty"`
}

// EventAnomalyDetection contains data for a detected anomaly in event
// registration records.
type EventAnomalyDetection struct {
	AnomalyType      string       `xml:"sf:TipoAnomalia"`
	OtherAnomalyData string       `xml:"sf:OtrosDatosAnomalia,omitempty"`
	AnomalousEvent   *EventRecord `xml:"sf:RegEventoAnomalo,omitempty"`
}

// InvoiceExportPeriod contains data about an export of invoice registration
// records for a given period.
type InvoiceExportPeriod struct {
	PeriodStart              string                             `xml:"sf:FechaHoraHusoInicioPeriodoExport"`
	PeriodEnd                string                             `xml:"sf:FechaHoraHusoFinPeriodoExport"`
	FirstInvoiceRecord       *EventInvoiceRecordWithFingerprint `xml:"sf:RegistroFacturacionInicialPeriodo"`
	LastInvoiceRecord        *EventInvoiceRecordWithFingerprint `xml:"sf:RegistroFacturacionFinalPeriodo"`
	RegistrationRecordCount  string                             `xml:"sf:NumeroDeRegistrosFacturacionAltaExportados"`
	TotalTaxSum              string                             `xml:"sf:SumaCuotaTotalAlta"`
	TotalAmountSum           string                             `xml:"sf:SumaImporteTotalAlta"`
	CancellationRecordCount  string                             `xml:"sf:NumeroDeRegistrosFacturacionAnulacionExportados"`
	ExportedRecordsDiscarded string                             `xml:"sf:RegistrosFacturacionExportadosDejanDeConservarse"`
}

// EventExportPeriod contains data about an export of event registration
// records for a given period.
type EventExportPeriod struct {
	PeriodStart              string       `xml:"sf:FechaHoraHusoInicioPeriodoExport"`
	PeriodEnd                string       `xml:"sf:FechaHoraHusoFinPeriodoExport"`
	FirstEventRecord         *EventRecord `xml:"sf:RegistroEventoInicialPeriodo"`
	LastEventRecord          *EventRecord `xml:"sf:RegistroEventoFinalPeriodo"`
	EventRecordCount         string       `xml:"sf:NumeroDeRegEventoExportados"`
	ExportedRecordsDiscarded string       `xml:"sf:RegEventoExportadosDejanDeConservarse"`
}

// EventSummary contains a summary of events (ResumenEventosType).
type EventSummary struct {
	EventTypes              []*AggregatedEventType             `xml:"sf:TipoEvento"`
	FirstInvoiceRecord      *EventInvoiceRecordWithFingerprint `xml:"sf:RegistroFacturacionInicialPeriodo,omitempty"`
	LastInvoiceRecord       *EventInvoiceRecordWithFingerprint `xml:"sf:RegistroFacturacionFinalPeriodo,omitempty"`
	RegistrationRecordCount string                             `xml:"sf:NumeroDeRegistrosFacturacionAltaGenerados"`
	TotalTaxSum             string                             `xml:"sf:SumaCuotaTotalAlta"`
	TotalAmountSum          string                             `xml:"sf:SumaImporteTotalAlta"`
	CancellationRecordCount string                             `xml:"sf:NumeroDeRegistrosFacturacionAnulacionGenerados"`
}

// AggregatedEventType represents an aggregated event type count (TipoEventoAgrType).
type AggregatedEventType struct {
	EventType  string `xml:"sf:TipoEvento"`
	EventCount string `xml:"sf:NumeroDeEventos"`
}

// EventRecord identifies an event registration record (RegEventoType).
type EventRecord struct {
	EventType      string `xml:"sf:TipoEvento"`
	EventTimestamp string `xml:"sf:FechaHoraHusoEvento"`
	Fingerprint    string `xml:"sf:HuellaEvento"`
}

// EventInvoiceRecord identifies an issued invoice (IDFacturaExpedidaType) in the
// events namespace.
type EventInvoiceRecord struct {
	IssuerNIF     string `xml:"sf:IDEmisorFactura"`
	InvoiceNumber string `xml:"sf:NumSerieFactura"`
	IssueDate     string `xml:"sf:FechaExpedicionFactura"`
}

// EventInvoiceRecordWithFingerprint identifies an issued invoice with its fingerprint
// (IDFacturaExpedidaHuellaType) in the events namespace.
type EventInvoiceRecordWithFingerprint struct {
	IssuerNIF     string `xml:"sf:IDEmisorFactura"`
	InvoiceNumber string `xml:"sf:NumSerieFactura"`
	IssueDate     string `xml:"sf:FechaExpedicionFactura"`
	Fingerprint   string `xml:"sf:Huella"`
}

// newEventRegistration builds a new VeriFactu event registration from bill.Status document.
func newEventRegistration(status *bill.Status, ts time.Time, s *Software) (*EventRegistration, error) {
	line := status.Lines[0]
	eventType := addon.EventType(line).String()

	ed, err := newEventData(line)
	if err != nil {
		return nil, err
	}

	return &EventRegistration{
		SF:      SF,
		Version: CurrentVersion,
		Event: &Event{
			Software: newEventSoftware(s),
			Issuer: &EventIssuer{
				Name: status.Supplier.Name,
				NIF:  status.Supplier.TaxID.Code.String(),
			},
			GenerationTimestamp: formatDateTimeZone(ts),
			FingerprintType:     FingerprintType,
			OtherEventData:      otherEventData(status),
			EventType:           eventType,
			EventData:           ed,
		},
	}, nil
}

// newEventSoftware builds a EventSoftware struct from a Software struct.
func newEventSoftware(s *Software) *EventSoftware {
	return &EventSoftware{
		Name:               s.NombreRazon,
		NIF:                s.NIF,
		SoftwareName:       s.NombreSistemaInformatico,
		SoftwareID:         s.IdSistemaInformatico,
		Version:            s.Version,
		InstallationNumber: s.NumeroInstalacion,
		OnlyVerifactu:      s.TipoUsoPosibleSoloVerifactu,
		MultiOT:            s.TipoUsoPosibleMultiOT,
		MultipleOT:         s.IndicadorMultiplesOT,
	}
}

// newEventData builds an EventData from the complements in the given status line.
// Simple events (startup, shutdown, etc.) have no complements and return nil.
func newEventData(line *bill.StatusLine) (*EventData, error) {
	if len(line.Complements) == 0 {
		return nil, nil // simple event, no event data
	}

	switch c := line.Complements[0].Instance().(type) {
	case *addon.InvoiceAnomalyLaunch:
		return &EventData{InvoiceAnomalyDetectionLaunch: newInvoiceAnomalyDetectionLaunch(c)}, nil
	case *addon.InvoiceAnomaly:
		return &EventData{InvoiceAnomalyDetection: newInvoiceAnomalyDetection(c, line)}, nil
	case *addon.EventAnomalyLaunch:
		return &EventData{EventAnomalyDetectionLaunch: newEventAnomalyDetectionLaunch(c)}, nil
	case *addon.EventAnomaly:
		return &EventData{EventAnomalyDetection: newEventAnomalyDetection(c, line)}, nil
	case *addon.InvoiceExport:
		return &EventData{InvoiceExportPeriod: newInvoiceExportPeriod(c)}, nil
	case *addon.EventExport:
		return &EventData{EventExportPeriod: newEventExportPeriod(c)}, nil
	case *addon.EventSummary:
		return &EventData{EventSummary: newEventSummary(c)}, nil
	default:
		return nil, fmt.Errorf("complement %s not supported", line.Complements[0].Schema)
	}
}

func newInvoiceAnomalyDetectionLaunch(c *addon.InvoiceAnomalyLaunch) *InvoiceAnomalyDetectionLaunch {
	return &InvoiceAnomalyDetectionLaunch{
		FingerprintCheck: checkStr(c.FingerprintCheck),
		FingerprintCount: countStr(c.FingerprintCount),
		SignatureCheck:   checkStr(c.SignatureCheck),
		SignatureCount:   countStr(c.SignatureCount),
		ChainCheck:       checkStr(c.ChainCheck),
		ChainCount:       countStr(c.ChainCount),
		DateCheck:        checkStr(c.DateCheck),
		DateCount:        countStr(c.DateCount),
	}
}

func newInvoiceAnomalyDetection(c *addon.InvoiceAnomaly, line *bill.StatusLine) *InvoiceAnomalyDetection {
	d := &InvoiceAnomalyDetection{
		AnomalyType:      c.Type.String(),
		OtherAnomalyData: line.Description,
	}
	if c.Invoice != nil {
		d.AnomalousInvoice = &EventInvoiceRecord{
			IssuerNIF:     c.Invoice.IssuerTaxCode,
			InvoiceNumber: c.Invoice.Code,
			IssueDate:     c.Invoice.IssueDate.Time().Format("02-01-2006"),
		}
	}
	return d
}

func newEventAnomalyDetectionLaunch(c *addon.EventAnomalyLaunch) *EventAnomalyDetectionLaunch {
	return &EventAnomalyDetectionLaunch{
		FingerprintCheck: checkStr(c.FingerprintCheck),
		FingerprintCount: countStr(c.FingerprintCount),
		SignatureCheck:   checkStr(c.SignatureCheck),
		SignatureCount:   countStr(c.SignatureCount),
		ChainCheck:       checkStr(c.ChainCheck),
		ChainCount:       countStr(c.ChainCount),
		DateCheck:        checkStr(c.DateCheck),
		DateCount:        countStr(c.DateCount),
	}
}

func newEventAnomalyDetection(c *addon.EventAnomaly, line *bill.StatusLine) *EventAnomalyDetection {
	d := &EventAnomalyDetection{
		AnomalyType:      c.Type.String(),
		OtherAnomalyData: line.Description,
	}
	if c.Event != nil {
		d.AnomalousEvent = &EventRecord{
			EventType:      c.Event.Type,
			EventTimestamp: c.Event.Timestamp,
			Fingerprint:    c.Event.Fingerprint,
		}
	}
	return d
}

func newInvoiceExportPeriod(c *addon.InvoiceExport) *InvoiceExportPeriod {
	return &InvoiceExportPeriod{
		PeriodStart:              c.Start,
		PeriodEnd:                c.End,
		RegistrationRecordCount:  strconv.Itoa(c.RegistrationCount),
		TotalTaxSum:              c.TaxTotal,
		TotalAmountSum:           c.AmountTotal,
		CancellationRecordCount:  strconv.Itoa(c.CancellationCount),
		ExportedRecordsDiscarded: c.Discarded,
		FirstInvoiceRecord:       newEventInvoiceRecordWithFingerprint(c.FirstRecord),
		LastInvoiceRecord:        newEventInvoiceRecordWithFingerprint(c.LastRecord),
	}
}

func newEventExportPeriod(c *addon.EventExport) *EventExportPeriod {
	return &EventExportPeriod{
		PeriodStart:              c.Start,
		PeriodEnd:                c.End,
		EventRecordCount:         countStr(&c.Count),
		ExportedRecordsDiscarded: c.Discarded,
		FirstEventRecord:         newEventRecord(c.FirstRecord),
		LastEventRecord:          newEventRecord(c.LastRecord),
	}
}

func newEventSummary(c *addon.EventSummary) *EventSummary {
	s := &EventSummary{
		RegistrationRecordCount: countStr(&c.RegistrationCount),
		TotalTaxSum:             c.TaxTotal,
		TotalAmountSum:          c.AmountTotal,
		CancellationRecordCount: countStr(&c.CancellationCount),
		FirstInvoiceRecord:      newEventInvoiceRecordWithFingerprint(c.FirstRecord),
		LastInvoiceRecord:       newEventInvoiceRecordWithFingerprint(c.LastRecord),
	}
	if len(c.Events) > 0 {
		s.EventTypes = make([]*AggregatedEventType, len(c.Events))
		for i, et := range c.Events {
			s.EventTypes[i] = &AggregatedEventType{
				EventType:  et.Type,
				EventCount: strconv.Itoa(et.Count),
			}
		}
	}
	return s
}

func newEventInvoiceRecordWithFingerprint(r *addon.InvoiceRecord) *EventInvoiceRecordWithFingerprint {
	if r == nil {
		return nil
	}
	return &EventInvoiceRecordWithFingerprint{
		IssuerNIF:     r.IssuerTaxCode,
		InvoiceNumber: r.Code,
		IssueDate:     r.IssueDate.Time().Format("02-01-2006"),
		Fingerprint:   r.Fingerprint,
	}
}

func newEventRecord(r *addon.EventRecord) *EventRecord {
	if r == nil {
		return nil
	}
	return &EventRecord{
		EventType:      r.Type,
		EventTimestamp: r.Timestamp,
		Fingerprint:    r.Fingerprint,
	}
}

// fingerprint sets the chaining information and computes the SHA-256 fingerprint for this
// event using the previous event's chain data. Per the AEAT spec (Art. 13).
func (e *Event) fingerprint(prev *EventChainData) {
	h := ""
	if prev == nil {
		e.Chaining = &EventChaining{
			FirstEvent: "S",
		}
	} else {
		e.Chaining = &EventChaining{
			PreviousEvent: &PreviousEvent{
				EventType:           prev.EventType,
				GenerationTimestamp: prev.GenerationTimestamp,
				Fingerprint:         prev.Fingerprint,
			},
		}
		h = prev.Fingerprint
	}

	var softwareNIF, softwareIDOtro string
	if e.Software != nil {
		softwareNIF = e.Software.NIF
		if e.Software.IDOther != nil {
			softwareIDOtro = e.Software.IDOther.ID
		}
	}

	var issuerNIF string
	if e.Issuer != nil {
		issuerNIF = e.Issuer.NIF
	}

	e.Fingerprint = computeFingerprint([]string{
		formatChainField("NIF", softwareNIF),
		formatChainField("ID", softwareIDOtro),
		formatChainField("IdSistemaInformatico", e.Software.SoftwareID),
		formatChainField("Version", e.Software.Version),
		formatChainField("NumeroInstalacion", e.Software.InstallationNumber),
		formatChainField("NIF", issuerNIF),
		formatChainField("TipoEvento", e.EventType),
		formatChainField("HuellaEvento", h),
		formatChainField("FechaHoraHusoGenEvento", e.GenerationTimestamp),
	})
}

// ChainData returns the chaining data from this event for use
// when fingerprinting the next event in the chain.
func (e *Event) ChainData() *EventChainData {
	return &EventChainData{
		EventType:           e.EventType,
		GenerationTimestamp: e.GenerationTimestamp,
		Fingerprint:         e.Fingerprint,
	}
}

// ChainData returns the chaining data from the inner event.
func (r *EventRegistration) ChainData() *EventChainData {
	if r.Event != nil {
		return r.Event.ChainData()
	}
	return nil
}

// Bytes prepares an XML document suitable for persistence. Signed documents
// use compact XML to preserve the enveloped signature.
func (r *EventRegistration) Bytes() ([]byte, error) {
	return toBytesIndent(r)
}

func otherEventData(status *bill.Status) string {
	for _, note := range status.Notes {
		if note.Key == org.NoteKeyOther {
			return note.Text
		}
	}
	return ""
}

func countStr(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}

func checkStr(v bool) string {
	if v {
		return "S"
	}
	return "N"
}
