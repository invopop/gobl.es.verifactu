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

type convertOpts struct {
	*rootOpts
}

func convert(o *rootOpts) *convertOpts {
	return &convertOpts{rootOpts: o}
}

func (c *convertOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert [infile] [outfile]",
		Short: "Convert a GOBL JSON into a VeriFactu XML",
		RunE:  c.runE,
	}

	f := cmd.Flags()
	c.prepareFlags(f)

	return cmd
}

func (c *convertOpts) runE(cmd *cobra.Command, args []string) error {
	input, err := openInput(cmd, args)
	if err != nil {
		return err
	}
	defer input.Close() // nolint:errcheck

	out, err := c.openOutput(cmd, args)
	if err != nil {
		return err
	}
	defer out.Close() // nolint:errcheck

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(input); err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	env := new(gobl.Envelope)
	if err := json.Unmarshal(buf.Bytes(), env); err != nil {
		return fmt.Errorf("unmarshaling gobl envelope: %w", err)
	}

	var opts []verifactu.Option

	if c.cert != "" {
		cert, err := xmldsig.LoadCertificate(c.cert, c.password)
		if err != nil {
			return err
		}
		opts = append(opts, verifactu.WithCertificate(cert))

		if c.sign {
			opts = append(opts, verifactu.WithSigning())
		}
	}

	vc, err := verifactu.New(c.software(), opts...)
	if err != nil {
		return fmt.Errorf("creating verifactu client: %w", err)
	}

	var reg interface{ Bytes() ([]byte, error) }
	switch env.Extract().(type) {
	case *bill.Invoice:
		reg, err = vc.RegisterInvoice(env, nil)
	case *bill.Status:
		reg, err = vc.RegisterEvent(env, nil)
	default:
		return fmt.Errorf("unsupported document type")
	}

	if err != nil {
		return fmt.Errorf("generating registration: %w", err)
	}

	data, err := reg.Bytes()
	if err != nil {
		return fmt.Errorf("generating verifactu xml: %w", err)
	}

	if _, err = out.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("writing verifactu xml: %w", err)
	}

	return nil
}
