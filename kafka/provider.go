package kafka

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"bootstrap_servers": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "A list of kafka brokers",
			},
			"ca_cert_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_CA_CERT", ""),
				Description: "Path to a CA certificate file to validate the server's certificate.",
				Deprecated:  "This parameter is now deprecated and will be removed in a later release, please use `ca_cert` instead.",
			},
			"client_cert_file": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_CLIENT_CERT", ""),
				Description: "Path to a file containing the client certificate.",
				Deprecated:  "This parameter is now deprecated and will be removed in a later release, please use `client_cert` instead.",
			},
			"client_key_file": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_CLIENT_KEY", ""),
				Description: "Path to a file containing the private key that the certificate was issued for.",
				Deprecated:  "This parameter is now deprecated and will be removed in a later release, please use `client_key` instead.",
			},
			"ca_cert": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_CA_CERT", ""),
				Description: "CA certificate file to validate the server's certificate.",
			},
			"client_cert": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_CLIENT_CERT", ""),
				Description: "The client certificate.",
			},
			"client_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_CLIENT_KEY", ""),
				Description: "The private key that the certificate was issued for.",
			},
			"sasl_username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_SASL_USERNAME", ""),
				Description: "Username for SASL authentication.",
			},
			"sasl_password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_SASL_PASSWORD", ""),
				Description: "Password for SASL authentication.",
			},
			"sasl_mechanism": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_SASL_MECHANISM", "plain"),
				Description: "SASL mechanism, can be plain, scram-sha512, scram-sha256",
			},
			"skip_tls_verify": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_SKIP_VERIFY", ""),
				Description: "Set this to true only if the target Kafka server is an insecure development instance.",
			},
			"tls_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_ENABLE_TLS", ""),
				Description: "Enable communication with the Kafka Cluster over TLS.",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     120,
				Description: "Timeout in seconds",
			},
		},

		ConfigureFunc: providerConfigure,
		ResourcesMap: map[string]*schema.Resource{
			"kafka_topic": kafkaTopicResource(),
			"kafka_acl":   kafkaACLResource(),
		},
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	brokers := dTos("bootstrap_servers", d)

	log.Printf("[DEBUG] configuring provider with Brokers @ %v", brokers)
	if brokers == nil {
		return nil, fmt.Errorf("bootstrap_servers was not set")
	}

	saslMechanism := d.Get("sasl_mechanism").(string)
	switch saslMechanism {
	case "scram-sha512", "scram-sha256", "plain":
	default:
		return nil, fmt.Errorf("[ERROR] Invalid sasl mechanism \"%s\": can only be \"scram-sha256\", \"scram-sha512\" or \"plain\"", saslMechanism)
	}

	config := &Config{
		BootstrapServers: brokers,
		CACert:           d.Get("ca_cert").(string),
		ClientCert:       d.Get("client_cert").(string),
		ClientCertKey:    d.Get("client_key").(string),
		SkipTLSVerify:    d.Get("skip_tls_verify").(bool),
		SASLUsername:     d.Get("sasl_username").(string),
		SASLPassword:     d.Get("sasl_password").(string),
		SASLMechanism:    saslMechanism,
		TLSEnabled:       d.Get("tls_enabled").(bool),
		Timeout:          d.Get("timeout").(int),
	}

	if config.CACert == "" {
		config.CACert = d.Get("ca_cert_file").(string)
	}
	if config.ClientCert == "" {
		config.ClientCert = d.Get("client_cert_file").(string)
	}
	if config.ClientCertKey == "" {
		config.ClientCertKey = d.Get("client_key_file").(string)
	}

	log.Printf("[DEBUG] Config @ %v", config)

	return NewClient(config)
}

func dTos(key string, d *schema.ResourceData) *[]string {
	var r *[]string

	if v, ok := d.GetOk(key); ok {
		if v == nil {
			return r
		}
		vI := v.([]interface{})
		b := make([]string, len(vI))

		for i, vv := range vI {
			if vv == nil {
				log.Printf("[DEBUG] %d %v was nil", i, vv)
				continue
			}
			log.Printf("[DEBUG] %d:Converting %v to string", i, vv)
			b[i] = vv.(string)
		}
		r = &b
	}

	return r
}
