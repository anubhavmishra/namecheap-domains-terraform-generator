package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/namecheap/go-namecheap-sdk/v2/namecheap"
)

type Domain struct {
	Name         string
	Nameservers  []string
	ResourceName string
}

var usageText = `Usage: namecheap-domains-terraform-generator {TERRAFORM_OUTPUT_FILE_NAME}
eg: ./namecheap-domains-terraform-generator test.tf
`

var domainRecordsResource = `resource "namecheap_domain_records" "{{.ResourceName}}" {
	domain = "{{.Name}}"
	mode = "OVERWRITE"
  
	nameservers = [
		{{range $i, $v := .Nameservers}}{{if $i}}{{printf ",\n		" }}{{end}}"{{.}}"{{end}}
	]
}

`

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 1 {
		fmt.Printf("error: output file name should be supplied\n\n")
		fmt.Println(usageText)
		os.Exit(1)
	}
	// Set terraform output file name
	fileName := argsWithoutProg[0]

	// Get namecheap environment variables
	namecheapUsername := os.Getenv("NAMECHEAP_USER_NAME")
	if namecheapUsername == "" {
		fmt.Printf("error: namecheap username needs to be set using the 'NAMECHEAP_USER_NAME' environment variable\n\n")
		os.Exit(1)
	}
	namecheapAPIUser := os.Getenv("NAMECHEAP_API_USER")
	if namecheapAPIUser == "" {
		fmt.Printf("error: namecheap api user needs to be set using the 'NAMECHEAP_API_USER' environment variable\n\n")
		os.Exit(1)
	}
	namecheapAPIKey := os.Getenv("NAMECHEAP_API_KEY")
	if namecheapAPIKey == "" {
		fmt.Printf("error: namecheap api key needs to be set using the 'NAMECHEAP_API_KEY' environment variable\n\n")
		os.Exit(1)
	}
	namecheapClientIP := os.Getenv("NAMECHEAP_CLIENT_IP")
	if namecheapClientIP == "" {
		fmt.Printf("error: namecheap client ip needs to be set using the 'NAMECHEAP_CLIENT_IP' environment variable\n\n")
		os.Exit(1)
	}
	namecheapUseSandbox := os.Getenv("NAMECHEAP_USE_SANDBOX")
	if namecheapUseSandbox == "" {
		fmt.Printf("error: namecheap use sandbox api needs to be set using the 'NAMECHEAP_USE_SANDBOX' environment variable\n\n")
		os.Exit(1)
	}
	namecheapUseSandboxBool, err := strconv.ParseBool(namecheapUseSandbox)
	if err != nil {
		fmt.Printf("error: couldn't convert 'NAMECHEAP_USE_SANDBOX' environment variable value to boolean: %v\n\n", err)
		os.Exit(1)
	}

	// Create namecheap client
	client := namecheap.NewClient(&namecheap.ClientOptions{
		UserName:   namecheapUsername,
		ApiUser:    namecheapAPIUser,
		ApiKey:     namecheapAPIKey,
		ClientIp:   namecheapClientIP,
		UseSandbox: namecheapUseSandboxBool,
	})

	domains, err := client.Domains.GetList(nil)
	if err != nil {
		log.Fatalf("error getting a domain list from namecheap: %v\n", err)
	}

	// Validate file doesn't exist
	if _, err := os.Stat(fileName); !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("error: file %q already exist\n\n", fileName)
		os.Exit(1)
	}

	// Create a new file
	file, _ := os.Create(fileName)
	defer file.Close()

	terraformImportCommands := []string{}
	for i, d := range *domains.Domains {

		// Check if domain is using namecheap dns for dns management
		namecheapDNS, err := usingNamecheapDNSManagement(client, *d.Name)
		if err != nil {
			log.Fatalf("error getting domain name %q info from namecheap: %v\n", *d.Name, err)
		}
		// Generate terraform resources for domains that don't use namecheap dns management
		// TODO: extend support for domains that use namecheap dns for dns management
		if !namecheapDNS {
			response, err := client.DomainsDNS.GetList(*d.Name)
			if err != nil {
				log.Fatalf("error listing domain name %q from namecheap: %v\n", *d.Name, err)
			}
			domain := Domain{
				Name:         *d.Name,
				Nameservers:  *response.DomainDNSGetListResult.Nameservers,
				ResourceName: fmt.Sprintf("domain_%s_%d", strings.Replace(*d.Name, ".", "_", -2), i+1),
			}
			err = renderTemplate(&domain, file)
			if err != nil {
				log.Fatalf("error rendering domain resource: %v\n", err)
			}
			// Append terraform import command
			terraformImportCommands = append(terraformImportCommands,
				fmt.Sprintf("terraform import namecheap_domain_records.%s %s",
					domain.ResourceName,
					domain.Name,
				),
			)
		}
	}

	log.Println("-> Successfully wrote all domain resources")
	fmt.Printf("\nFile saved: %q\n", file.Name())

	// Show terraform import command if any
	if len(terraformImportCommands) > 0 {
		fmt.Printf("Terraform import command for the resources are as follows:\n\n")
		for _, c := range terraformImportCommands {
			fmt.Println(c)
		}
	}
}

func renderTemplate(domain *Domain, f *os.File) error {
	tmpl := template.New("domainRecordsResource")
	tmp, err := tmpl.Parse(domainRecordsResource)
	if err != nil {
		return err
	}
	err = tmp.Execute(f, domain)
	if err != nil {
		return err
	}

	log.Printf("-> Wrote terraform resource for %q domain\n", domain.Name)

	return nil
}

func usingNamecheapDNSManagement(client *namecheap.Client, name string) (bool, error) {
	responseInfo, err := client.Domains.GetInfo(name)
	if err != nil {
		return false, err
	}

	if responseInfo != nil {
		if responseInfo.DomainDNSGetListResult != nil {
			return *responseInfo.DomainDNSGetListResult.DnsDetails.IsUsingOurDNS, nil
		}
	}

	return false, nil
}
