#!/bin/bash
# Generate coverage report markdown using Go's statement-based coverage
# This ensures CI reports match local `go test -cover` output

set -e

COVERAGE_FILE="${1:-coverage.out}"
OUTPUT_FILE="${2:-code-coverage-results.md}"
MIN_THRESHOLD="${3:-90}"
HEALTH_THRESHOLD="${4:-96}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "Error: Coverage file $COVERAGE_FILE not found"
    exit 1
fi

# Get total coverage (weighted by statements)
get_total_coverage() {
    go tool cover -func="$COVERAGE_FILE" | \
    grep "total:" | \
    awk '{print $NF}' | \
    sed 's/%//'
}

# Calculate per-package coverage weighted by statements
# Format: package covered_statements total_statements percentage
calc_package_stats() {
    # Parse coverage.out directly to get statement counts
    # Format: mode: set/count/atomic
    # Then: filename:startline.startcol,endline.endcol numstatements count
    tail -n +2 "$COVERAGE_FILE" | \
    awk -F'[: ,]' '
    {
        # $1 is the file path
        file = $1
        # Get package from file path
        n = split(file, parts, "/")
        if (n >= 2) {
            # Find github.com/SmrutAI/pedantigo in path
            for (i = 1; i <= n-1; i++) {
                if (parts[i] == "github.com" || parts[i] == "pedantigo") {
                    pkg = ""
                    for (j = i; j <= n-1; j++) {
                        if (pkg != "") pkg = pkg "/"
                        pkg = pkg parts[j]
                    }
                    break
                }
            }
        }

        # numstatements is field after the line:col,line:col pattern
        # The format is: file:line.col,line.col numstmts count
        # After splitting by [: ,], we need to find the right fields
        numstmts = $(NF-1)
        count = $NF

        if (pkg != "" && numstmts > 0) {
            total[pkg] += numstmts
            if (count > 0) {
                covered[pkg] += numstmts
            }
        }
    }
    END {
        for (pkg in total) {
            pct = (total[pkg] > 0) ? int(100 * covered[pkg] / total[pkg] + 0.5) : 0
            print pkg, covered[pkg], total[pkg], pct
        }
    }' | sort
}

# Determine health indicator
get_health() {
    local pct="$1"
    # Handle empty or non-numeric values
    if [ -z "$pct" ] || ! [[ "$pct" =~ ^[0-9]+$ ]]; then
        echo "➖"
        return
    fi
    if [ "$pct" -ge "$HEALTH_THRESHOLD" ]; then
        echo "✔"
    else
        echo "➖"
    fi
}

# Generate the report
echo "Generating coverage report..."

TOTAL_PCT=$(get_total_coverage)
TOTAL_INT=${TOTAL_PCT%.*}

# Create badge
if [ "$TOTAL_INT" -ge "$HEALTH_THRESHOLD" ]; then
    BADGE_COLOR="brightgreen"
elif [ "$TOTAL_INT" -ge "$MIN_THRESHOLD" ]; then
    BADGE_COLOR="yellow"
else
    BADGE_COLOR="red"
fi

# Calculate totals
TOTAL_COVERED=0
TOTAL_STMTS=0

# Start markdown output
{
    echo "![Code Coverage](https://img.shields.io/badge/Code%20Coverage-${TOTAL_INT}%25-${BADGE_COLOR}?style=flat)"
    echo ""
    echo "Package | Line Rate | Complexity | Health"
    echo "-------- | --------- | ---------- | ------"

    # Get package stats
    calc_package_stats | while read -r pkg covered total pct; do
        [ -z "$pkg" ] && continue
        # Default to 0 if pct is empty
        [ -z "$pct" ] && pct="0"
        HEALTH=$(get_health "$pct")
        echo "${pkg} | ${pct}% | 0 | ${HEALTH}"
    done

    echo "**Summary** | **${TOTAL_INT}%** | **0** | $(get_health "$TOTAL_INT")"
    echo ""
    echo "_Minimum allowed line rate is \`${MIN_THRESHOLD}%\`_"
    echo ""
    echo "<!-- Sticky Pull Request Comment -->"
} > "$OUTPUT_FILE"

echo "Coverage report written to $OUTPUT_FILE"
echo "Total coverage: ${TOTAL_PCT}%"

# Check threshold
if [ "$TOTAL_INT" -lt "$MIN_THRESHOLD" ]; then
    echo "Error: Coverage ${TOTAL_INT}% is below minimum threshold ${MIN_THRESHOLD}%"
    exit 1
fi

# Write coverage percentage for badge update
echo "${TOTAL_INT}" > coverage.txt
