#!/bin/bash
# @meta
# name: example
# description: A starter script showing the metadata convention
# category: templates
# args:
#   - name: message
#     required: false
#     help: Text to echo back
# @end

echo "Hello from the example script. You said: ${1:-nothing}"
