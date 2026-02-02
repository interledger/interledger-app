For the InterledgerApp Wallet application we want to have a consistent logging strategy across all the components that are expected to be actually deployed to either sandbox or production clusters. If logging is normalised across components then we can have clean and simple tools to assist us in troubleshooting problems, but it will also make it a lot easier to set up monitoring and alerts.

# Application Logging Guide

## General:
- Understand that **Application logs** are not **Audit logs** since they will not be kept forever. Application logs can be kept anything between 7 to 14 days depending on configuration, but audit logs are to be kept forever.
- All persistent components to be deployed to the cluster should have a concept of a LOG_LEVEL which is to be configured as an environmental variable to the process. 
- All logs should be rendered as JSON objects, one object per line
- Do not render big objects since there can be a performance cost to serialisation
- Do not embed conditionals inside the logging statement since this has a tendency to make code flaky
- Take great care when de-referencing objects inside a logging statement since this can make code flaky
- Should the LOG_LEVEL configured not be one of the valid options then the application should exit immediately and describe the problem using `fatal`
## Log level:
- `fatal`
    - Where to?
        - Should direct to `stderr` only
    - When?
        - Must always cause the application to exit as safely as it can. Yes, the whole process.
    - Examples
        - Database credentials are incorrect so I can't function at all
        - My security has been compromised and I refuse to continue
        - There is an obvious configuration problem, I cannot continue
- `error`
    - Where to?
        - Should direct to `stderr` only
    - Purpose
        - To notify and inform administrators of an issue that needs immediate attention.
    - Examples
        - I've now tried to trigger a transaction against provider X for the 10th time, but it failed AGAIN. Somebody needs to check this out.
        - 
- `warning`
    - Where to?
        - Should direct to `stdout` only
    - Purpose
        - To notify and inform maintainers of a non-critical but problematic event that occurred
        - Support and AI will periodically look at the warnings to see if issues needs to be escalated. **They are not to be ignored**.
    - When?
        - "There was some issue, but the rest of the system is probably fine so it can be checked out later"
    - When NOT?
        - My user typed in the wrong password - This is normal application flow and does not need to be logged
        - There was not enough money in the account for the transaction to complete - Normal application flow which may or may not justify a `info` log.
        - I want my logs to stand out so I'm logging some information at warning level
    - Examples
        - I tried to trigger a transaction against provider X, but it timed out. This job will try again in 30 seconds. Might just be a network hiccup.

- `info`
    - Should direct to `stdout` only
    - Purpose
        - Helps system administrators understand the state of the application
        - Tracks events that might be useful in understanding the 
    - When?
        - When noteworthy but non-problematic application events occurred
    - When NOT?
        - Events that occur often
    - Examples
        - Time to print out useful statistics
        - Indicate that a report is now being generated which might consume more processing than normal operations
        - `{"level":"info","ts":1769499135.3677459,"caller":"ops/workflows.go:219","msg":"No rafiki payment pointer found for wallet","walletID":"2e4ea62d-3e5f-4266-8e03-374cce423584"}`

- `debug`
    - Where to?
        - Should direct to `stdout` only
    - When?
        - When troubleshooting a specific service in some environment. 
    - Examples?
        - What IP address is that call coming from?
        - Print out that whole object so that I can see why my test is failing
    - Special notes:
        - In general,  debug logs should be stripped out when merging into main, but in some cases developers need to get debug logs into a target environment. This is to prevent performance cost which can sometimes come with rendering log strings even if they are not actually printed


# Additional fields
- `ts`
    - timestamp - For golang logs it makes sense to follow the zapper approach of emitting the timestamp in the unix format.
- `caller`
    - In golang, the zapper library provides this field to assist in troubleshooting.
    - This is optional, other services and packages do not have to emit this
- `requestId`
    - Optional for attaching a requestID which makes troubleshooting easier
- `correlationId`
    - Optional for correlating async logs together