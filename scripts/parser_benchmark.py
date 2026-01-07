#!/usr/bin/env python3
"""
Benchmark script to compare python-gedcom parser with ligneous-gedcom Go parser.

This script:
1. Tests both parsers on the same GEDCOM files
2. Measures parsing time for each
3. Compares performance metrics
4. Outputs detailed results
"""

import os
import sys
import time
import subprocess
import json
import statistics
from pathlib import Path

# Add python-gedcom to path
sys.path.insert(0, '/apps/python-gedcom')

try:
    from gedcom.parser import Parser as PythonParser
except ImportError:
    print("ERROR: Could not import python-gedcom. Make sure it's installed or cloned.")
    sys.exit(1)

# Test data directory
TESTDATA_DIR = Path('/apps/gedcom-go/testdata')
GEDCOM_FILES = [
    'xavier.ged',
    'tree1.ged',
    'gracis.ged',
    'pres2020.ged',
    'royal92.ged',
]

# Go parser binary (we'll build it)
GO_PARSER_BINARY = '/tmp/go_parser_benchmark'


def build_go_parser():
    """Build a Go benchmark binary that parses files and reports timing."""
    go_code = '''package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintf(os.Stderr, "Usage: %s <gedcom_file>\\n", os.Args[0])
        os.Exit(1)
    }
    
    filePath := os.Args[1]
    
    start := time.Now()
    p := parser.NewParser()
    tree, err := p.Parse(filePath)
    duration := time.Since(start)
    
    result := map[string]interface{}{
        "success": err == nil,
        "duration_ms": duration.Seconds() * 1000,
        "duration_ns": duration.Nanoseconds(),
        "error": "",
    }
    
    if err != nil {
        result["error"] = err.Error()
    } else {
        individuals := tree.GetAllIndividuals()
        families := tree.GetAllFamilies()
        result["individuals"] = len(individuals)
        result["families"] = len(families)
        result["errors"] = len(p.GetErrors())
    }
    
    jsonData, _ := json.Marshal(result)
    fmt.Println(string(jsonData))
}
'''
    
    go_file = '/tmp/go_parser_benchmark.go'
    with open(go_file, 'w') as f:
        f.write(go_code)
    
    # Build the binary
    cmd = ['go', 'build', '-o', GO_PARSER_BINARY, go_file]
    result = subprocess.run(cmd, capture_output=True, text=True, cwd='/apps/gedcom-go')
    
    if result.returncode != 0:
        print(f"ERROR: Failed to build Go parser: {result.stderr}")
        return False
    
    return True


def benchmark_python_parser(file_path):
    """Benchmark the Python parser."""
    try:
        parser = PythonParser()
        
        start = time.perf_counter()
        parser.parse_file(str(file_path), strict=True)
        duration = (time.perf_counter() - start) * 1000  # Convert to milliseconds
        
        # Get statistics
        from gedcom.element.individual import IndividualElement
        from gedcom.element.family import FamilyElement
        
        element_list = parser.get_element_list()
        element_dict = parser.get_element_dictionary()
        
        individuals = [e for e in element_list if isinstance(e, IndividualElement)]
        families = [e for e in element_list if isinstance(e, FamilyElement)]
        
        return {
            'success': True,
            'duration_ms': duration,
            'duration_ns': int(duration * 1_000_000),
            'individuals': len(individuals),
            'families': len(families),
            'total_elements': len(element_list),
            'error': None
        }
    except Exception as e:
        return {
            'success': False,
            'duration_ms': 0,
            'duration_ns': 0,
            'error': str(e)
        }


def benchmark_go_parser(file_path):
    """Benchmark the Go parser."""
    if not os.path.exists(GO_PARSER_BINARY):
        return {
            'success': False,
            'duration_ms': 0,
            'duration_ns': 0,
            'error': 'Go parser binary not found'
        }
    
    try:
        result = subprocess.run(
            [GO_PARSER_BINARY, str(file_path)],
            capture_output=True,
            text=True,
            timeout=300  # 5 minute timeout
        )
        
        if result.returncode != 0:
            return {
                'success': False,
                'duration_ms': 0,
                'duration_ns': 0,
                'error': result.stderr or 'Unknown error'
            }
        
        # Parse JSON output
        data = json.loads(result.stdout)
        return data
    except subprocess.TimeoutExpired:
        return {
            'success': False,
            'duration_ms': 0,
            'duration_ns': 0,
            'error': 'Timeout (exceeded 5 minutes)'
        }
    except json.JSONDecodeError as e:
        return {
            'success': False,
            'duration_ms': 0,
            'duration_ns': 0,
            'error': f'JSON decode error: {e}'
        }
    except Exception as e:
        return {
            'success': False,
            'duration_ms': 0,
            'duration_ns': 0,
            'error': str(e)
        }


def format_duration(ms):
    """Format duration in a human-readable way."""
    if ms < 1:
        return f"{ms*1000:.2f} μs"
    elif ms < 1000:
        return f"{ms:.2f} ms"
    else:
        return f"{ms/1000:.2f} s"


def format_speedup(go_time, python_time):
    """Calculate and format speedup ratio."""
    if go_time == 0:
        return "N/A"
    speedup = python_time / go_time
    return f"{speedup:.2f}x"


def get_file_size(file_path):
    """Get file size in a human-readable format."""
    size = os.path.getsize(file_path)
    if size < 1024:
        return f"{size} B"
    elif size < 1024 * 1024:
        return f"{size/1024:.1f} KB"
    else:
        return f"{size/(1024*1024):.1f} MB"


def main():
    print("=" * 80)
    print("GEDCOM Parser Performance Comparison")
    print("=" * 80)
    print()
    
    # Build Go parser
    print("Building Go parser benchmark binary...")
    if not build_go_parser():
        print("ERROR: Failed to build Go parser. Exiting.")
        sys.exit(1)
    print("✓ Go parser built successfully")
    print()
    
    # Find available test files
    available_files = []
    for filename in GEDCOM_FILES:
        file_path = TESTDATA_DIR / filename
        if file_path.exists():
            available_files.append((filename, file_path))
        else:
            print(f"⚠ Warning: {filename} not found, skipping")
    
    if not available_files:
        print("ERROR: No test files found!")
        sys.exit(1)
    
    print(f"Found {len(available_files)} test file(s)")
    print()
    
    # Run benchmarks
    results = []
    
    for filename, file_path in available_files:
        print(f"Testing: {filename} ({get_file_size(file_path)})")
        print("-" * 80)
        
        # Python parser
        print("  Running Python parser...", end=" ", flush=True)
        python_result = benchmark_python_parser(file_path)
        if python_result['success']:
            print(f"✓ {format_duration(python_result['duration_ms'])}")
        else:
            print(f"✗ ERROR: {python_result['error']}")
        
        # Go parser
        print("  Running Go parser...", end=" ", flush=True)
        go_result = benchmark_go_parser(file_path)
        if go_result['success']:
            print(f"✓ {format_duration(go_result['duration_ms'])}")
        else:
            print(f"✗ ERROR: {go_result['error']}")
        
        # Calculate speedup
        if python_result['success'] and go_result['success']:
            speedup = format_speedup(go_result['duration_ms'], python_result['duration_ms'])
            print(f"  Speedup: Go is {speedup} faster")
        
        # Store results
        results.append({
            'filename': filename,
            'file_size': get_file_size(file_path),
            'python': python_result,
            'go': go_result,
        })
        
        print()
    
    # Print summary table
    print("=" * 80)
    print("SUMMARY")
    print("=" * 80)
    print()
    print(f"{'File':<20} {'Size':<10} {'Python (ms)':<15} {'Go (ms)':<15} {'Speedup':<15}")
    print("-" * 80)
    
    python_times = []
    go_times = []
    
    for result in results:
        filename = result['filename']
        file_size = result['file_size']
        
        if result['python']['success']:
            python_time = result['python']['duration_ms']
            python_times.append(python_time)
            python_str = f"{python_time:.2f}"
        else:
            python_str = "ERROR"
        
        if result['go']['success']:
            go_time = result['go']['duration_ms']
            go_times.append(go_time)
            go_str = f"{go_time:.2f}"
        else:
            go_str = "ERROR"
        
        if result['python']['success'] and result['go']['success']:
            speedup = format_speedup(go_time, python_time)
        else:
            speedup = "N/A"
        
        print(f"{filename:<20} {file_size:<10} {python_str:<15} {go_str:<15} {speedup:<15}")
    
    print("-" * 80)
    
    # Calculate averages
    if python_times and go_times:
        avg_python = statistics.mean(python_times)
        avg_go = statistics.mean(go_times)
        avg_speedup = format_speedup(avg_go, avg_python)
        
        print(f"{'AVERAGE':<20} {'':<10} {avg_python:<15.2f} {avg_go:<15.2f} {avg_speedup:<15}")
        print()
        
        # Overall statistics
        print("Overall Statistics:")
        print(f"  Average Python time: {format_duration(avg_python)}")
        print(f"  Average Go time: {format_duration(avg_go)}")
        print(f"  Average speedup: {avg_speedup}")
        
        if len(python_times) > 1 and len(go_times) > 1:
            median_python = statistics.median(python_times)
            median_go = statistics.median(go_times)
            median_speedup = format_speedup(median_go, median_python)
            print(f"  Median Python time: {format_duration(median_python)}")
            print(f"  Median Go time: {format_duration(median_go)}")
            print(f"  Median speedup: {median_speedup}")
    
    print()
    print("=" * 80)
    
    # Detailed results
    print("\nDETAILED RESULTS")
    print("=" * 80)
    for result in results:
        print(f"\n{result['filename']} ({result['file_size']}):")
        print(f"  Python:")
        if result['python']['success']:
            print(f"    Time: {format_duration(result['python']['duration_ms'])}")
            if 'individuals' in result['python']:
                print(f"    Individuals: {result['python']['individuals']}")
                print(f"    Families: {result['python']['families']}")
                print(f"    Total Elements: {result['python']['total_elements']}")
        else:
            print(f"    ERROR: {result['python']['error']}")
        
        print(f"  Go:")
        if result['go']['success']:
            print(f"    Time: {format_duration(result['go']['duration_ms'])}")
            if 'individuals' in result['go']:
                print(f"    Individuals: {result['go']['individuals']}")
                print(f"    Families: {result['go']['families']}")
                if 'errors' in result['go']:
                    print(f"    Errors: {result['go']['errors']}")
        else:
            print(f"    ERROR: {result['go']['error']}")
    
    # Cleanup
    if os.path.exists(GO_PARSER_BINARY):
        os.remove(GO_PARSER_BINARY)
    if os.path.exists('/tmp/go_parser_benchmark.go'):
        os.remove('/tmp/go_parser_benchmark.go')


if __name__ == '__main__':
    main()

