#!/bin/bash

# Performance Test Runner for GEDCOM Go
# Runs comprehensive performance tests with large datasets (100K, 500K individuals)

set -e

echo "=========================================="
echo "GEDCOM Go Performance Test Suite"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test sizes
SIZES=(100000 500000)

# Function to run tests
run_tests() {
    local size=$1
    local test_name=$2
    
    echo -e "${YELLOW}Running $test_name tests with $size individuals...${NC}"
    
    go test -v -run "TestPerformance.*${size}" ./$test_name/... 2>&1 | tee /tmp/performance_${test_name}_${size}.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        echo -e "${GREEN}✓ $test_name ($size) passed${NC}"
    else
        echo -e "${RED}✗ $test_name ($size) failed${NC}"
        return 1
    fi
}

# Function to run benchmarks
run_benchmarks() {
    local size=$1
    local test_name=$2
    
    echo -e "${YELLOW}Running $test_name benchmarks with $size individuals...${NC}"
    
    go test -bench="Benchmark.*${size}" -benchmem ./$test_name/... 2>&1 | tee /tmp/benchmark_${test_name}_${size}.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        echo -e "${GREEN}✓ $test_name benchmarks ($size) completed${NC}"
    else
        echo -e "${RED}✗ $test_name benchmarks ($size) failed${NC}"
        return 1
    fi
}

# Main execution
echo "Starting performance tests..."
echo ""

# Query Performance Tests
echo "=== Query Performance Tests ==="
for size in "${SIZES[@]}"; do
    run_tests "$size" "pkg/gedcom/query" || true
    echo ""
done

# Parser Performance Tests
echo "=== Parser Performance Tests ==="
for size in "${SIZES[@]}"; do
    run_tests "$size" "internal/parser" || true
    echo ""
done

# Duplicate Detection Performance Tests
echo "=== Duplicate Detection Performance Tests ==="
for size in "${SIZES[@]}"; do
    run_tests "$size" "pkg/gedcom/duplicate" || true
    echo ""
done

# Benchmarks
echo "=== Running Benchmarks ==="
echo ""
echo "Query Benchmarks:"
for size in "${SIZES[@]}"; do
    run_benchmarks "$size" "pkg/gedcom/query" || true
    echo ""
done

echo "Parser Benchmarks:"
for size in "${SIZES[@]}"; do
    run_benchmarks "$size" "internal/parser" || true
    echo ""
done

echo "Duplicate Detection Benchmarks:"
for size in "${SIZES[@]}"; do
    run_benchmarks "$size" "pkg/gedcom/duplicate" || true
    echo ""
done

echo "=========================================="
echo -e "${GREEN}Performance test suite completed!${NC}"
echo "=========================================="
echo ""
echo "Log files saved to /tmp/performance_*.log and /tmp/benchmark_*.log"
