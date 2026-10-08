// Package main provides the command line interface to the VeriFactu package.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/invopop/gobl"
	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/xmldsig"
	"github.com/spf13/cobra"
)

type cancelOpts struct {
	*rootOpts
	previous string
}

func cancel(o *rootOpts) *cancelOpts {
	return &cancelOpts{rootOpts: o}
}

func (c *cancelOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [infile]",
		Short: "Cancels the GOBL invoice to the VeriFactu service",
		RunE:  c.runE,
	}

	f := cmd.Flags()
	c.prepareFlags(f)

	f.StringVar(&c.previous, "prev", "", "Previous document fingerprint to chain with")

	return cmd
}

func (c *cancelOpts) runE(cmd *cobra.Command, args []string) error {
	input, err := openInput(cmd, args)
	if err != nil {
		return err
	}
	defer input.Close() // nolint:errcheck

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(input); err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	env := new(gobl.Envelope)
	if err := json.Unmarshal(buf.Bytes(), env); err != nil {
		return fmt.Errorf("unmarshaling gobl envelope: %w", err)
	}

	cert, err := xmldsig.LoadCertificate(c.cert, c.password)
	if err != nil {
		return fmt.Errorf("loading certificate: %w", err)
	}

	opts := []verifactu.Option{
		verifactu.WithCertificate(cert),
	}

	if c.sign {
		opts = append(opts, verifactu.WithSigning())
	}

	if c.production {
		opts = append(opts, verifactu.InProduction())
	} else {
		opts = append(opts, verifactu.InSandbox())
	}

	vc, err := verifactu.New(c.software(), opts...)
	if err != nil {
		return err
	}

	var prev *verifactu.ChainData
	if c.previous != "" {
		prev = new(verifactu.ChainData)
		if err := json.Unmarshal([]byte(c.previous), prev); err != nil {
			return err
		}
	}

	req, err := vc.CancelInvoice(env, prev)
	if err != nil {
		return fmt.Errorf("generating invoice cancellation: %w", err)
	}

	inv := env.Extract().(*bill.Invoice)
	ir, err := vc.NewInvoiceRequest(inv.Supplier)
	if err != nil {
		return fmt.Errorf("preparing invoice request: %w", err)
	}
	ir.AddCancellation(req)

	res, err := vc.SendInvoiceRequest(cmd.Context(), ir)
	if err != nil {
		return fmt.Errorf("sending cancellation: %w", err)
	}

	data, err := json.Marshal(ir.Lines[len(ir.Lines)-1].ChainData())
	if err != nil {
		return err
	}
	fmt.Printf("Generated document with fingerprint: \n%s\n", string(data))

	// Response
	rd, err := res.Bytes()
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	fmt.Print(string(rd))

	return nil
}
