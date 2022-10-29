#!/bin/bash

# Script is inspired from: https://github.com/rabbitmq/cluster-operator/blob/main/hack/remove-override-descriptions.sh
# Licensed under the Mozilla Public license, Version 2.0 (the "License"). Copyright 2020 VMware, Inc. All Rights Reserved.

tmp=$(mktemp)
yj -yj < config/crd/bases/temporal.io_temporalclusters.yaml | \
    jq 'delpaths([.. | paths(scalars)|select(contains(["spec","versions",0,"schema","openAPIV3Schema","properties","spec","properties","services","properties","frontend","properties","overrides","description"]))])' | \
    jq 'delpaths([.. | paths(scalars)|select(contains(["spec","versions",0,"schema","openAPIV3Schema","properties","spec","properties","services","properties","history","properties","overrides","description"]))])' | \
    jq 'delpaths([.. | paths(scalars)|select(contains(["spec","versions",0,"schema","openAPIV3Schema","properties","spec","properties","services","properties","matching","properties","overrides","description"]))])' | \
    jq 'delpaths([.. | paths(scalars)|select(contains(["spec","versions",0,"schema","openAPIV3Schema","properties","spec","properties","services","properties","worker","properties","overrides","description"]))])' | \
    jq 'delpaths([.. | paths(scalars)|select(contains(["spec","versions",0,"schema","openAPIV3Schema","properties","spec","properties","services","properties","overrides","description"]))])' | \
    yj -jy > "$tmp"
mv "$tmp" config/crd/bases/temporal.io_temporalclusters.yaml