#!/usr/bin/env bash

# TODO... make more robust.

# Check for the expected asset types, or otherwise fail.
rc={[ "$PREFLIGHT_ASSET_TYPE" = "container" ] || [ "$PREFLIGHT_ASSET_TYPE" = "operator" ]; echo $? ;}
[ $rc -ne 0 ] && { echo "ERR An incorrect asset type was provided. Expecting 'container' or 'operator'."; exit 1}

echo "Starting demo execution of preflight..."

# Tell preflight to log verbosely.
export PFLT_LOGLEVEL=trace

# Sanity check: ensure preflight exists and execute it.
which preflight &>/dev/null && preflight check "${PREFLIGHT_ASSET_TYPE}" "${PREFLIGHT_TEST_ASSET}"

# Write logs from the current working directory to the artifacts directory defined by CI to extract them.
# results.json is in the working directory, but everything else is in a local ./artifacts directory, so
# we move all of those to the CI pipeline's artifact location.
cp -a results.json artifacts "${ARTIFACT_DIR}"/

echo "Ending demo execution of preflight..."
exit 0