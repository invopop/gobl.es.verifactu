package verifactu_test

import (
	"testing"

	// Register the addon so the example documents declaring es-verifactu-v1
	// normalize and validate.
	_ "github.com/invopop/gobl.es.verifactu/addon"
	"github.com/invopop/gobl.es.verifactu/test"

	"github.com/invopop/gobl/pkg/examples"
)

// TestAddonExamples converts every document under examples/ into a calculated,
// validated JSON envelope and compares it against its golden output, using the
// shared GOBL example helpers. This exercises the addon end to end, separately
// from the XML conversion covered by TestXMLGeneration.
//
// Run with -update to regenerate the goldens.
func TestAddonExamples(t *testing.T) {
	examples.Run(t, "examples", *test.UpdateOut)
}
