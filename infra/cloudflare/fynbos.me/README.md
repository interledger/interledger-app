## Fynbos.dev Cloudflare

The following stack is used to provision all CF resources for `ilp.link`. In order to run the stack you need to 
get keys from the cloudflare UI. The two keys required are:
* `CLOUDFLARE_API_TOKEN`
* `CLOUDFLARE_API_USER_SERVICE_KEY` 

### Run
Run the following command for the stack
`CLOUDFLARE_API_TOKEN=XXX CLOUDFLARE_API_USER_SERVICE_KEY=XXX aws-vault exec XXX -- pulumi up -s fynbos/main`