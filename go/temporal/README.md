# Build folder for Temporal

This folder exists for the exclusive purpose of building the customised Temporal Docker image used in deployments. It will be build in the CI/CD pipeline and pushed to the Docker registry.

It uses the official Temporal 'autosetup' image and we copy the bootstrap script into it during build.

We inherited this scheme from Fynbos. It doesn't seem like the autosetup.sh script has any real modifications inside of it so we could probably just use the official image directly. But we didn't want to change too much at once.