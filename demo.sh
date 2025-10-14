#!/bin/bash
# Demo script for runstats - showcasing various use cases

set -e

echo "=== runstats Demo ==="
echo

# Build if needed
if [ ! -f ./runstats ]; then
    echo "Building runstats..."
    go build
    echo
fi

# Demo 1: Simple sleep command
echo "1. Simple command (sleep 1):"
echo "   Command: ./runstats sleep 1"
./runstats sleep 1
echo

# Demo 2: CPU-intensive workload
echo "2. CPU-intensive workload (for loop):"
echo "   Command: ./runstats sh -c 'for i in \$(seq 1 1000000); do : ; done'"
./runstats sh -c 'for i in $(seq 1 1000000); do : ; done'
echo

# Demo 3: Memory usage demo
echo "3. Build command (shows higher memory usage):"
echo "   Command: ./runstats go build"
./runstats go build
echo

# Demo 4: Command with output
echo "4. Command with output (ls):"
echo "   Command: ./runstats ls -lh"
./runstats ls -lh
echo

# Demo 5: Failed command
echo "5. Failed command (returns non-zero exit code):"
echo "   Command: ./runstats false"
./runstats false || echo "   (Note: Exit code preserved)"
echo

# Demo 6: Pipeline demonstration
echo "6. Pipeline usage:"
echo "   Command: echo 'hello world' | ./runstats grep 'hello'"
echo 'hello world' | ./runstats grep 'hello'
echo

# Demo 7: Environment dump
echo "7. Environment dump (with -dump-env flag):"
echo "   Command: ./runstats -dump-env printenv | head -1"
TEST_VAR=demo_value ./runstats -dump-env sh -c 'echo "Environment captured"' 2>&1 | tail -1 | jq '.environment | keys | length'
echo "   (Captured all environment variables in JSON output)"
echo

echo "=== Demo Complete ==="
echo
echo "Tips:"
echo "  - Metrics are printed to stderr"
echo "  - Command output goes to stdout normally"
echo "  - Exit codes are preserved"
echo "  - Use '2> metrics.json' to capture metrics to file"
