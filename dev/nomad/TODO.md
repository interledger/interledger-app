
* [ ] add reschedule on group/job level 
```hcl
reschedule {
 delay          = "30s"
 delay_function = "exponential"
 max_delay      = "1h"
 unlimited      = true
}
```
This helps quicker startup as well as potential issues with the sidecar not getting secrets in time.