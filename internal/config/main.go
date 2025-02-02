package config

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

type FortiExporterParameter struct {
	AuthFile      *string
	Listen        *string
	ScrapeTimeout *int
	TLSTimeout    *int
	TLSInsecure   *bool
	TlsExtraCAs   *string
}

type FortiExporterConfig struct {
	AuthKeys      AuthKeys
	Listen        string
	ScrapeTimeout int
	TLSTimeout    int
	TLSInsecure   bool
	TlsExtraCAs   []LocalCert
}

type AuthKeys map[Target]Token

type Target string
type Token string

type LocalCert struct {
	Path    string
	Content []byte
}

var (
	parameter = FortiExporterParameter{
		AuthFile:      flag.String("auth-file", "", "file containing the authentication map to use when connecting to a Fortigate device"),
		Listen:        flag.String("listen", ":9710", "address to listen on"),
		ScrapeTimeout: flag.Int("scrape-timeout", 30, "max seconds to allow a scrape to take"),
		TLSTimeout:    flag.Int("https-timeout", 10, "TLS Handshake timeout in seconds"),
		TLSInsecure:   flag.Bool("insecure", false, "Allow insecure certificates"),
		TlsExtraCAs:   flag.String("extra-ca-certs", "", "comma-separated files containing extra PEMs to trust for TLS connections in addition to the system trust store"),
	}

	savedConfig FortiExporterConfig
)

func Init() error {
	// check if already parsed
	if flag.Parsed() {
		return nil
	}

	flag.Parse()

	savedConfig = FortiExporterConfig{
		Listen:        *parameter.Listen,
		ScrapeTimeout: *parameter.ScrapeTimeout,
		TLSTimeout:    *parameter.TLSTimeout,
		TLSInsecure:   *parameter.TLSInsecure,
	}

	// parse AuthKeys
	af, err := ioutil.ReadFile(*parameter.AuthFile)
	if err != nil {
		log.Fatalf("Failed to read API authentication map file: %v", err)
		return err
	}

	if err := yaml.Unmarshal(af, &savedConfig.AuthKeys); err != nil {
		log.Fatalf("Failed to parse API authentication map file: %v", err)
		return err
	}

	log.Printf("Loaded %d API keys", len(savedConfig.AuthKeys))

	// parse ExtraCAs
	for _, eca := range strings.Split(*parameter.TlsExtraCAs, ",") {
		if eca == "" {
			continue
		}

		certs, err := ioutil.ReadFile(eca)
		if err != nil {
			log.Fatalf("Failed to read extra CA file %q: %v", eca, err)
			return err
		}

		certObject := LocalCert{
			Path:    eca,
			Content: certs,
		}
		savedConfig.TlsExtraCAs = append(savedConfig.TlsExtraCAs, certObject)
	}

	return nil
}

func GetConfig() FortiExporterConfig {
	return savedConfig
}
