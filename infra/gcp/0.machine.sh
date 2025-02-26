#!/bin/bash

# abort on nonzero exitstatus
set -o errexit

# abort on unbound variable
set -o nounset

# don't hide errors within pipes
set -o pipefail

INSTANCE_NAME=interledger-app-dev
ZONE=us-central1-c
PROJECT=wallet-451208
INSTANCE_SERVICE_ACCOUNT=interledger-dev-server@wallet-451208.iam.gserviceaccount.com
SSH_USERNAME=stephan
SSH_PUB_KEY=$(cat ./authorized_keys)

# Check if there is not already an instance with the same name, if so then error out
if gcloud compute instances list --project=$PROJECT --filter="name=$INSTANCE_NAME" | grep -q $INSTANCE_NAME; then
    echo "Instance with name $INSTANCE_NAME already exists"
    exit 1
fi


gcloud compute instances create $INSTANCE_NAME \
    --project=$PROJECT \
    --zone=$ZONE \
    --machine-type=e2-custom-2-8192 \
    --network-interface=network-tier=PREMIUM,stack-type=IPV4_ONLY,subnet=default \
    --maintenance-policy=MIGRATE \
    --provisioning-model=STANDARD \
    --service-account=$INSTANCE_SERVICE_ACCOUNT \
    --scopes=https://www.googleapis.com/auth/cloud-platform \
    --tags=http-server,https-server \
    --create-disk=auto-delete=yes,boot=yes,device-name=${INSTANCE_NAME},image=projects/debian-cloud/global/images/debian-12-bookworm-v20250212,mode=rw,size=30,type=pd-balanced \
    --no-shielded-secure-boot \
    --shielded-vtpm \
    --shielded-integrity-monitoring \
    --labels=goog-ec-src=vm_add-gcloud \
    --reservation-affinity=any \
    --metadata=ssh-keys="$SSH_USERNAME:$SSH_PUB_KEY"

# Get the external IP of the instance
EXTERNAL_IP=$(gcloud compute instances describe $INSTANCE_NAME --zone=$ZONE --project=$PROJECT --format='get(networkInterfaces[0].accessConfigs[0].natIP)')
INTERNAL_IP=$(gcloud compute instances describe $INSTANCE_NAME --zone=$ZONE --project=$PROJECT --format='get(networkInterfaces[0].networkIP)')


echo "Instance created with external IP: $EXTERNAL_IP"

echo "export EXTERNAL_IP=$EXTERNAL_IP" > environment
echo "export INTERNAL_IP=$INTERNAL_IP" >> environment
echo "export HOST_IP=$INTERNAL_IP" >> environment
echo "export NOMAD_ADDR=http://$INTERNAL_IP:4646" >> environment

echo "Waiting 45s for machine to boot and services to start"
sleep 45

# Copy contents of this folder to the instance
scp -r * $SSH_USERNAME@$EXTERNAL_IP:/home/$SSH_USERNAME

echo "Instance created and files copied to instance"
echo "Run the rest of the steps directly on the instance"
