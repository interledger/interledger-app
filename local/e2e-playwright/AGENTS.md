The correct way to run specific tests from `features/*.feature` is to tag the specific test appropriately and then use the -args and -tags flags together.
```
# Run only the @signuponly tests
go test -v -timeout 5m -args -tags @signuponly 
# Run tests tagged with @withdrawal AND @fees
go test -v -timeout 10m -args -tags="@withdrawal&&@fees"
# Run tests without debug info
go test -v -timeout 10m -args -debug=false
```

DO NOT SUPPRESS TEST OUTPUT EVER

## Troubleshooting
- Remember during tests users are unique, so database cleanup has very limited value if at all.
- You can investigate temporal jobs by executin temporal cli commands within the temporal container
- We spent a long time chasing Kratos `format: "tel"` validation failures that appeared to reject valid phone numbers. After cleaning the environment with `make reset` in `local`, the issue disappeared. The root cause is still unknown, so keep this in mind if `tel` errors resurface after environment changes.
- Important details about phone number troubleshooting
  + Keep in mind that the tests aim to generate randomised phone numbers so they are not supposed tobe duplicate
  + We've confirmed that the correct format is +49987654321
  + Most issues relate to kratos validation, either the phone number already exist or format is wrong
  + When starting up the environment then use `make all-nowatch`
  + The `/withdraw` page loads a MockGatehub iframe widget similar to deposits

## Maintain
It is the job of the agent to add, update or remove relevant information to this file.