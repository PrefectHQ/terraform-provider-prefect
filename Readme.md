# terraform-provider-prefect
Terraform provider for Prefect 2.0.
This is on a draft/early state of what it should look like. But it works to deploy basic resources, like workspaces, work queues and blocks.

Features included in this project work with the current state of the Prefect Cloud 2.0 REST API (as of 20/02/2023).
For deployment instructions see the "Deployment" section below.

## Features:
* Data Sources:
    - Workspaces
    - Work Queues
    - Block Types
    - Block Schemas
    - Block Documents

* Resources:
    - Workspace (Create)
    - Work Queue (Create, Update, Delete)
    - Block (Create, Update, Delete)

## Deployment:
The "examples" folder makes use of this local provider.   
By following the instructions below you can get it deployed to a target Prefect Cloud 2.0 account.

### 1. Create Prefect 2.0 Service Account
The creation of a [service account](https://docs.prefect.io/ui/service-accounts/#create-a-service-account) generates a key that we will use to authenticate the terraform provider to Prefect Cloud. 
The advantage of a service account is that it is not tied to a user, but directly to the account.

### 2. Configure environment variables
In this step we export 3 environment variables:

PREFECT_API_URL = the Prefect Cloud API endpoint  
PREFECT_API_KEY = the authentication key (generated as part of the previous step)  
PREFECT_ACCOUNT_ID = the account id (by clicking on your organisation you'll see this in the URL)  

Run this on your terminal (replace placeholders):

```
export PREFECT_API_URL="https://api.prefect.cloud/api"
export PREFECT_API_KEY="<YOUR-API-KEY>"
export PREFECT_ACCOUNT_ID="<YOUR-ACCOUNT-ID>"
```
### 3. Build the provider
This builds the providers's binary and move it to the Terraform plugins directory (usually under `~/.terraform.d/plugins/`)

Before building the provider make sure you've set the correct CPU architecture of your machine.  
E.g: for MAC M1, use `darwin_arm64`, for MAC Intel use `darwin_amd64`.  
```
export CPU_ARCHITECTURE="darwin_arm64"
```
Now run the `make` command to build the provider:
```
make install
```

### 4. Deploy the example infrastructure
This is the final step, which deploys the infrastructure to Prefect Cloud
```
cd examples
terraform init
terraform apply
```

In case of a success output, go to Prefect Cloud and find the workspace named `terraform-workspace`   

### 5. Tear down the example infrastructure
While the capability to _destroy_ work queues and blocks is in place, you won't be able to completely destroy the stack because the API endpoint to destroy a workspace hasn't been made available. For now you'll need to go to the Prefect Cloud UI, click on the workspace `terraform-workspace` > workspace settings > (hamburger button on the top right) > delete.

Then manually remove the local terraform state files

```
rm -rf examples/.terraform
rm -rf examples/.terraform.lock.hcl
rm -rf examples/.terraform
```

## Notes & Improvements:
* This is far from complete.   
* There are no code _tests_ in place at present (adding them will very likely lead to changes to make the code more robust).  
* The provider's implementation of `blocks` has been done generically. For a more granular infrastructure state control, consider changing the BlockDocument's `data` field to the target block type e.g: `s3`, `kubernetes job`. This would generate more work and duplication though.   
* The GO Prefect 2.0 API (`prefect_api`) should be moved out of the provider, into its own project.  
* Should authentication be handled with temporary tokens?  
