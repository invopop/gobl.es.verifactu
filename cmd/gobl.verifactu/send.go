// Package main provides the command line interface to the VeriFactu package.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/invopop/gobl"
	verifactu "github.com/invopop/gobl.es.verifactu"
	"github.com/invopop/xmldsig"
	"github.com/spf13/cobra"
)

type sendOpts struct {
	*rootOpts
	previous string
}

func send(o *rootOpts) *sendOpts {
	return &sendOpts{rootOpts: o}
}

func (c *sendOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [infile]",
		Short: "Sends the GOBL invoice to the VeriFactu service",
		RunE:  c.runE,
	}

	f := cmd.Flags()
	c.prepareFlags(f)

	f.StringVar(&c.previous, "prev", "", "Previous document fingerprint to chain with")

	return cmd
}

func (c *sendOpts) runE(cmd *cobra.Command, args []string) error {
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
		return err
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

	tc, err := verifactu.New(c.software(), opts...)
	if err != nil {
		return fmt.Errorf("preparing software: %w", err)
	}

	var prev *verifactu.ChainData
	if c.previous != "" {
		prev = new(verifactu.ChainData)
		if err := json.Unmarshal([]byte(c.previous), prev); err != nil {
			return fmt.Errorf("chain data: %w", err)
		}
	}

	ir, err := tc.NewEnvelopeInvoiceRequest(env, prev)
	if err != nil {
		return fmt.Errorf("envelope invoice request: %w", err)
	}

	res, err := tc.SendInvoiceRequest(cmd.Context(), ir)
	if err != nil {
		return fmt.Errorf("send invoice request: %w", err)
	}

	// get last line
	l := ir.Lines[len(ir.Lines)-1]
	data, err := json.Marshal(l.ChainData())
	if err != nil {
		return err
	}
	fmt.Printf("Document sent: \n%s\n", string(data))

	// Response
	rd, err := res.Bytes()
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	fmt.Print(string(rd))

	return nil
}
