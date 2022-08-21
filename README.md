# namecheap-domains-terraform-generator
Fetches Namecheap domains and generates Terraform configuration

**Note: Currently, the project only supports domains that don't use Namecheap DNS for DNS management.**

## Usage

Set environment variables

```bash
export NAMECHEAP_USER_NAME="USER_NAME"
export NAMECHEAP_API_USER="API_USER" # Usually the same as 'NAMECHEAP_USER_NAME'
export NAMECHEAP_API_KEY="API_KEY"
export NAMECHEAP_CLIENT_IP="ALLOWED_IP_ADDRESS"
export NAMECHEAP_USE_SANDBOX="false" # Set this to 'true' when using sandbox

```

Build project

```bash
go build .
```

Run

```bash
./namecheap-domains-terraform-generator {TERRAFORM_OUTPUT_FILE_NAME}
```

Example

```bash
./namecheap-domains-terraform-generator main.tf
```

Expected output

```
2022/08/20 16:16:43 -> Wrote terraform resource for "example.com" domain
.....
2022/08/20 16:16:45 -> Successfully wrote all domain resources

File saved: "main.tf"
Terraform import command for the resources are as follows:

terraform import namecheap_domain_records.domain_example_com_1 example.com

```
