package verifactu_test

import (
	"testing"
	"time"

	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/invopop/gobl.es.verifactu/test"
	"github.com/invopop/gobl/bill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistroEvento(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2024-11-26T04:00:00Z")
	require.NoError(t, err)
	vc, err := verifactu.New(
		verifactu.Software{
			NombreRazon:              "My Software",
			NIF:                      "12345678A",
			NombreSistemaInformatico: "My Software",
			IdSistemaInformatico:     "A1",
			Version:                  "1.0",
			NumeroInstalacion:        "12345678A",
		},
		verifactu.WithCurrentTime(ts),
	)
	require.NoError(t, err)

	t.Run("system startup", func(t *testing.T) {
		env := test.LoadEnvelope("status-system-startup.json")
		reg, err := vc.RegisterEvent(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "1.0", reg.Version)
		assert.Equal(t, "01", reg.Event.EventType)
		assert.Equal(t, "01", reg.Event.FingerprintType)
		assert.Equal(t, "Invopop S.L.", reg.Event.Issuer.Name)
		assert.Equal(t, "B85905495", reg.Event.Issuer.NIF)
		assert.Equal(t, "My Software", reg.Event.Software.Name)
		assert.Equal(t, "12345678A", reg.Event.Software.NIF)
		assert.Nil(t, reg.Event.EventData)
		assert.Empty(t, reg.Event.OtherEventData)
		assert.Equal(t, "S", reg.Event.Chaining.FirstEvent)
		assert.NotEmpty(t, reg.Event.Fingerprint)
	})

	t.Run("backup restoration with note", func(t *testing.T) {
		env := test.LoadEnvelope("status-backup-restoration.json")
		reg, err := vc.RegisterEvent(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "07", reg.Event.EventType)
		assert.Nil(t, reg.Event.EventData)
		assert.Equal(t, "Backup restored from 2024-11-18", reg.Event.OtherEventData)
	})

	t.Run("invoice anomaly detection launch", func(t *testing.T) {
		env := test.LoadEnvelope("status-anomaly-launch-invoices.json")
		reg, err := vc.RegisterEvent(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "03", reg.Event.EventType)
		require.NotNil(t, reg.Event.EventData)
		require.NotNil(t, reg.Event.EventData.InvoiceAnomalyDetectionLaunch)

		launch := reg.Event.EventData.InvoiceAnomalyDetectionLaunch
		assert.Equal(t, "S", launch.FingerprintCheck)
		assert.Equal(t, "150", launch.FingerprintCount)
		assert.Equal(t, "S", launch.SignatureCheck)
		assert.Equal(t, "150", launch.SignatureCount)
		assert.Equal(t, "S", launch.ChainCheck)
		assert.Equal(t, "150", launch.ChainCount)
		assert.Equal(t, "N", launch.DateCheck)
		assert.Empty(t, launch.DateCount)
	})

	t.Run("event anomaly detection launch", func(t *testing.T) {
		env := test.LoadEnvelope("status-anomaly-launch-events.json")
		reg, err := vc.RegisterEvent(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "05", reg.Event.EventType)
		require.NotNil(t, reg.Event.EventData)
		require.NotNil(t, reg.Event.EventData.EventAnomalyDetectionLaunch)

		launch := reg.Event.EventData.EventAnomalyDetectionLaunch
		assert.Equal(t, "S", launch.FingerprintCheck)
		assert.Equal(t, "50", launch.FingerprintCount)
		assert.Equal(t, "N", launch.SignatureCheck)
		assert.Empty(t, launch.SignatureCount)
		assert.Equal(t, "S", launch.ChainCheck)
		assert.Equal(t, "50", launch.ChainCount)
		assert.Equal(t, "S", launch.DateCheck)
		assert.Equal(t, "50", launch.DateCount)
	})

	t.Run("invoice anomaly detected", func(t *testing.T) {
		env := test.LoadEnvelope("status-anomaly-detected-invoices.json")
		reg, err := vc.RegisterEvent(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "04", reg.Event.EventType)
		require.NotNil(t, reg.Event.EventData)
		require.NotNil(t, reg.Event.EventData.InvoiceAnomalyDetection)

		det := reg.Event.EventData.InvoiceAnomalyDetection
		assert.Equal(t, "01", det.AnomalyType)
		assert.Equal(t, "Fingerprint mismatch in invoice record", det.OtherAnomalyData)
		require.NotNil(t, det.AnomalousInvoice)
		assert.Equal(t, "B85905495", det.AnomalousInvoice.IssuerNIF)
		assert.Equal(t, "SAMPLE-001", det.AnomalousInvoice.InvoiceNumber)
		assert.Equal(t, "15-11-2024", det.AnomalousInvoice.IssueDate)
	})

	t.Run("event anomaly detected", func(t *testing.T) {
		env := test.LoadEnvelope("status-anomaly-detected-events.json")
		reg, err := vc.RegisterEvent(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "06", reg.Event.EventType)
		require.NotNil(t, reg.Event.EventData)
		require.NotNil(t, reg.Event.EventData.EventAnomalyDetection)

		det := reg.Event.EventData.EventAnomalyDetection
		assert.Equal(t, "07", det.AnomalyType)
		assert.Equal(t, "Chain traceability issue detected", det.OtherAnomalyData)
		require.NotNil(t, det.AnomalousEvent)
		assert.Equal(t, "01", det.AnomalousEvent.EventType)
		assert.Equal(t, "2024-11-19T10:00:00+01:00", det.AnomalousEvent.EventTimestamp)
	})

	t.Run("invoice export period", func(t *testing.T) {
		env := test.LoadEnvelope("status-export-invoices.json")
		reg, err := vc.RegisterEvent(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "08", reg.Event.EventType)
		require.NotNil(t, reg.Event.EventData)
		require.NotNil(t, reg.Event.EventData.InvoiceExportPeriod)

		exp := reg.Event.EventData.InvoiceExportPeriod
		assert.Equal(t, "2024-11-01T00:00:00+01:00", exp.PeriodStart)
		assert.Equal(t, "2024-11-20T23:59:59+01:00", exp.PeriodEnd)
		assert.Equal(t, "45", exp.RegistrationRecordCount)
		assert.Equal(t, "9450.00", exp.TotalTaxSum)
		assert.Equal(t, "54450.00", exp.TotalAmountSum)
		assert.Equal(t, "5", exp.CancellationRecordCount)
		assert.Equal(t, "N", exp.ExportedRecordsDiscarded)
		require.NotNil(t, exp.FirstInvoiceRecord)
		assert.Equal(t, "B85905495", exp.FirstInvoiceRecord.IssuerNIF)
		assert.Equal(t, "SAMPLE-001", exp.FirstInvoiceRecord.InvoiceNumber)
		require.NotNil(t, exp.LastInvoiceRecord)
		assert.Equal(t, "SAMPLE-050", exp.LastInvoiceRecord.InvoiceNumber)
	})

	t.Run("event export period", func(t *testing.T) {
		env := test.LoadEnvelope("status-export-events.json")
		reg, err := vc.RegisterEvent(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "09", reg.Event.EventType)
		require.NotNil(t, reg.Event.EventData)
		require.NotNil(t, reg.Event.EventData.EventExportPeriod)

		exp := reg.Event.EventData.EventExportPeriod
		assert.Equal(t, "30", exp.EventRecordCount)
		assert.Equal(t, "N", exp.ExportedRecordsDiscarded)
		require.NotNil(t, exp.FirstEventRecord)
		assert.Equal(t, "01", exp.FirstEventRecord.EventType)
		require.NotNil(t, exp.LastEventRecord)
		assert.Equal(t, "10", exp.LastEventRecord.EventType)
	})

	t.Run("event summary", func(t *testing.T) {
		env := test.LoadEnvelope("status-event-summary.json")
		reg, err := vc.RegisterEvent(env, nil)
		require.NoError(t, err)

		assert.Equal(t, "10", reg.Event.EventType)
		require.NotNil(t, reg.Event.EventData)
		require.NotNil(t, reg.Event.EventData.EventSummary)

		sum := reg.Event.EventData.EventSummary
		require.Len(t, sum.EventTypes, 5)
		assert.Equal(t, "01", sum.EventTypes[0].EventType)
		assert.Equal(t, "2", sum.EventTypes[0].EventCount)
		assert.Equal(t, "10", sum.EventTypes[4].EventType)
		assert.Equal(t, "4", sum.EventTypes[4].EventCount)
		assert.Equal(t, "18", sum.RegistrationRecordCount)
		assert.Equal(t, "3780.00", sum.TotalTaxSum)
		assert.Equal(t, "21780.00", sum.TotalAmountSum)
		assert.Equal(t, "2", sum.CancellationRecordCount)
		require.NotNil(t, sum.FirstInvoiceRecord)
		assert.Equal(t, "SAMPLE-001", sum.FirstInvoiceRecord.InvoiceNumber)
		require.NotNil(t, sum.LastInvoiceRecord)
		assert.Equal(t, "SAMPLE-020", sum.LastInvoiceRecord.InvoiceNumber)
	})

	t.Run("invalid status", func(t *testing.T) {
		env := test.LoadEnvelope("status-system-startup.json")
		st := env.Extract().(*bill.Status)
		st.Lines = nil
		_, err := vc.RegisterEvent(env, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one line")
	})

	t.Run("with previous event chain", func(t *testing.T) {
		env := test.LoadEnvelope("status-system-startup.json")
		prev := &verifactu.EventChainData{
			EventType:           "02",
			GenerationTimestamp: "2024-11-25T20:00:00+01:00",
			Fingerprint:         "AABBCCDD00112233AABBCCDD00112233AABBCCDD00112233AABBCCDD00112233",
		}
		reg, err := vc.RegisterEvent(env, prev)
		require.NoError(t, err)

		assert.Empty(t, reg.Event.Chaining.FirstEvent)
		require.NotNil(t, reg.Event.Chaining.PreviousEvent)
		assert.Equal(t, "02", reg.Event.Chaining.PreviousEvent.EventType)
		assert.Equal(t, "2024-11-25T20:00:00+01:00", reg.Event.Chaining.PreviousEvent.GenerationTimestamp)
		assert.Equal(t, prev.Fingerprint, reg.Event.Chaining.PreviousEvent.Fingerprint)
	})
}
