## Fynbos.app Cloudflare

The following stack is used to provision all CF resources for `fynbos.app`. In order to run the stack you need to 
get keys from the cloudflare UI. The key required is:
* `CLOUDFLARE_API_TOKEN`

### Run
Run the following command for the stack
`CLOUDFLARE_API_TOKEN=XXX pulumi up -s fynbos/main`