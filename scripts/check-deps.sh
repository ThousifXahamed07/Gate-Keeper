#!/usr/bin/env bash

set -e

# Default allowed list of dependencies (space-separated)
ALLOWED_DEPS=("gopkg.in/yaml.v3" "gopkg.in/check.v1")

# Extract all required modules from go.mod
# Handles both single line 'require module' and block 'require ( module )' 
# Removes version string and any trailing // indirect
extracted_deps=$(awk '
/^require \(/ { in_block=1; next }
/^\)/ { if(in_block) in_block=0; next }
in_block && $1 != "" { print $1 }
/^require / && !in_block { print $2 }
' go.mod)

failed=0

for dep in $extracted_deps; do
    is_allowed=0
    for allowed in "${ALLOWED_DEPS[@]}"; do
        if [ "$dep" = "$allowed" ]; then
            is_allowed=1
            break
        fi
    done

    if [ $is_allowed -eq 0 ]; then
        echo "❌ Unauthorized dependency found: $dep"
        failed=1
    fi
done

if [ "$failed" -eq 1 ]; then
    exit 1
fi

echo "✅ Dependency check passed"
exit 0
