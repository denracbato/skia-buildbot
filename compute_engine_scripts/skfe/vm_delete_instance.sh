#!/bin/bash
#
# Deletes the compute instance for skia-docs.
#
set -x

source vm_config.sh

for NUM in $(seq 1 $NUM_INSTANCES); do

  gcloud compute instances delete \
    --project=$PROJECT_ID \
    --delete-disks "all" \
    --zone=$ZONE \
    $INSTANCE_NAME-$NUM

done
