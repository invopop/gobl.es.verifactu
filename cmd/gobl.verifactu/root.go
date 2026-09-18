package main

import (
	"io"
	"os"

	verifactu "github.com/invopop/gobl.es.verifactu"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type rootOpts struct {
	cert                   string
	password               string
	swNombreRazon          string
	swNIF                  string
	swName                 string
	swIDSistemaInformatico string
	swVersion              string
	swNumeroInstalacion    string
	production             bool
	sign                   bool
}

func root() *rootOpts {
	return &rootOpts{}
}

func (o *rootOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(versionCmd())
	cmd.AddCommand(send(o).cmd())
	cmd.AddCommand(convert(o).cmd())
	cmd.AddCommand(cancel(o).cmd())

	return cmd
}

func (o *rootOpts) prepareFlags(f *pflag.FlagSet) {
	f.StringVar(&o.cert, "cert", os.Getenv("CERTIFICATE_PATH"), "Certificate for authentication")
	f.StringVar(&o.password, "password", os.Getenv("CERTIFICATE_PASSWORD"), "Password of the certificate")
	f.StringVar(&o.swNIF, "sw-nif", os.Getenv("SOFTWARE_COMPANY_NIF"), "NIF of the software company")
	f.StringVar(&o.swName, "sw-name", os.Getenv("SOFTWARE_NAME"), "Name of the software")
	f.StringVar(&o.swNombreRazon, "sw-company", os.Getenv("SOFTWARE_COMPANY_NAME"), "Name of the software company")
	f.StringVar(&o.swVersion, "sw-version", os.Getenv("SOFTWARE_VERSION"), "Version of the software")
	f.StringVar(&o.swIDSistemaInformatico, "sw-id", os.Getenv("SOFTWARE_ID_SISTEMA_INFORMATICO"), "ID of the software system")
	f.StringVar(&o.swNumeroInstalacion, "sw-inst", os.Getenv("SOFTWARE_NUMERO_INSTALACION"), "Number of the software installation")
	f.BoolVarP(&o.production, "production", "p", false, "Production environment")
	f.BoolVar(&o.sign, "sign", false, "Enable XML digital signatures on records")
}

func (o *rootOpts) software() verifactu.Software {
	return verifactu.Software{
		NIF:                         o.swNIF,
		NombreRazon:                 o.swNombreRazon,
		Version:                     o.swVersion,
		IdSistemaInformatico:        o.swIDSistemaInformatico,
		NumeroInstalacion:           o.swNumeroInstalacion,
		NombreSistemaInformatico:    o.swName,
		TipoUsoPosibleSoloVerifactu: "S",
		TipoUsoPosibleMultiOT:       "S",
		IndicadorMultiplesOT:        "N",
	}
}

func (o *rootOpts) outputFilename(args []string) string {
	if len(args) >= 2 && args[1] != "-" {
		return args[1]
	}
	return ""
}

func openInput(cmd *cobra.Command, args []string) (io.ReadCloser, error) {
	if inFile := inputFilename(args); inFile != "" {
		return os.Open(inFile)
	}
	return io.NopCloser(cmd.InOrStdin()), nil
}

func (o *rootOpts) openOutput(cmd *cobra.Command, args []string) (io.WriteCloser, error) {
	if outFile := o.outputFilename(args); outFile != "" {
		flags := os.O_CREATE | os.O_WRONLY
		return os.OpenFile(outFile, flags, os.ModePerm)
	}
	return writeCloser{cmd.OutOrStdout()}, nil
}

type writeCloser struct {
	io.Writer
}

func (writeCloser) Close() error { return nil }
