#!/usr/bin/bash

# Script to launch a standalone EC2 instance with parameters

set -ueo pipefail

if [ $# -lt 2 ]; then
    echo "Usage: $0 <stack-name> [Key=Value]..."
    exit 1
fi

STACK_NAME="$1"
shift

# Get script parameters (formated as Key=Value) and convert to parameters for AWS CLI (ParameterKey=Key,ParameterValue=Value)
PARAMETERS=()
for PARAM in "$@"; do
    # fail if parameter is not in the form Key=Value
    REPLACED=$(echo "$PARAM" | sed -e 's/\([[:upper:]][[:alpha:]]*\)=\(.*\)/ParameterKey=\1,ParameterValue=\2/g; t')
    PARAMETERS+=("$REPLACED")
done
# Launch the stack and wait for its creation
aws cloudformation create-stack --stack-name $STACK_NAME --template-body file://vpc.yaml --parameters "${PARAMETERS[@]}"
aws cloudformation wait stack-create-complete --stack-name $STACK_NAME
