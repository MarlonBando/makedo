#!/bin/bash

# Build the latest binary
echo "Building makedo..."
make build || exit 1
echo ""

FAILED=0

FILES=("${@}")
if [ ${#FILES[@]} -eq 0 ]; then
    FILES=(tests/*.md README.md skills/SKILL.md)
fi

# Iterate over every markdown file
for file in "${FILES[@]}"; do
    # Skip if file doesn't exist (e.g. glob didn't match)
    [ -f "$file" ] || continue

    echo "========================================"
    echo "Running: $file"
    echo "========================================"
    
    if ! ./bin/makedo test "$file"; then
        echo "❌ FAILED: $file"
        FAILED=1
    else
        echo "✅ PASSED: $file"
    fi
    echo ""
done

if [ $FAILED -ne 0 ]; then
    echo "Some tests failed! ❌"
    exit 1
else
    echo "All tests passed successfully! 🎉"
    exit 0
fi
