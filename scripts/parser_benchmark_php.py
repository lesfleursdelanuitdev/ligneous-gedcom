#!/usr/bin/env python3
"""
Benchmark script to compare ligneous-gedcom (Go) parser with php-gedcom parser.

This script:
1. Tests both parsers on the same GEDCOM files
2. Measures parsing time for each
3. Compares performance metrics
4. Outputs detailed results

Note: php-gedcom requires PHP 8.4+. The script will check and report if PHP 8.4 is not available.
"""

import os
import sys
import time
import subprocess
import json
import statistics
from pathlib import Path

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
PHP_BENCHMARK_SCRIPT = '/tmp/php_parser_benchmark.php'


def check_php_version():
    """Check if PHP 8.3+ is available (8.3 for v2.2.0, 8.4+ for v4.0+)."""
    try:
        result = subprocess.run(['php', '--version'], capture_output=True, text=True)
        version_line = result.stdout.split('\n')[0]
        # Extract version number
        version_str = version_line.split()[1]
        major, minor = map(int, version_str.split('.')[:2])
        
        # Check for PHP 8.3+ (v2.2.0 supports 8.3+, v4.0+ requires 8.4+)
        if major > 8 or (major == 8 and minor >= 3):
            return True, version_str, minor >= 4
        return False, version_str, False
    except Exception as e:
        return False, f"Error: {e}", False


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


def create_php_benchmark_script():
    """Create PHP benchmark script."""
    php_code = '''<?php
/**
 * PHP GEDCOM Parser Benchmark
 * 
 * This script benchmarks the php-gedcom parser and outputs JSON results.
 */

// Check PHP version (8.3+ for v2.2.0, 8.4+ for v4.0+)
if (version_compare(PHP_VERSION, '8.3.0', '<')) {
    echo json_encode([
        'success' => false,
        'error' => 'PHP 8.3+ required. Current version: ' . PHP_VERSION,
        'duration_ms' => 0,
        'duration_ns' => 0
    ]);
    exit(1);
}

// Set up autoloader for php-gedcom
$autoloader = '/apps/php-gedcom/vendor/autoload.php';
if (!file_exists($autoloader)) {
    echo json_encode([
        'success' => false,
        'error' => 'php-gedcom not installed. Run: cd /apps/php-gedcom && composer install',
        'duration_ms' => 0,
        'duration_ns' => 0
    ]);
    exit(1);
}

require_once $autoloader;

use Gedcom\Parser;

if ($argc < 2) {
    echo json_encode([
        'success' => false,
        'error' => 'Usage: php benchmark.php <gedcom_file>',
        'duration_ms' => 0,
        'duration_ns' => 0
    ]);
    exit(1);
}

$filePath = $argv[1];

try {
    $startTime = microtime(true);
    $startNano = hrtime(true);
    
    $parser = new Parser();
    $gedcom = $parser->parse($filePath);
    
    $endTime = microtime(true);
    $endNano = hrtime(true);
    
    $durationMs = ($endTime - $startTime) * 1000;
    $durationNs = $endNano - $startNano;
    
    if ($gedcom === null) {
        echo json_encode([
            'success' => false,
            'error' => 'Parser returned null',
            'duration_ms' => $durationMs,
            'duration_ns' => $durationNs
        ]);
        exit(1);
    }
    
    $individuals = $gedcom->getIndi();
    $families = $gedcom->getFam();
    
    echo json_encode([
        'success' => true,
        'duration_ms' => $durationMs,
        'duration_ns' => $durationNs,
        'individuals' => count($individuals),
        'families' => count($families),
        'errors' => 0  // php-gedcom doesn't expose errors easily
    ]);
    
} catch (Exception $e) {
    echo json_encode([
        'success' => false,
        'error' => $e->getMessage(),
        'duration_ms' => 0,
        'duration_ns' => 0
    ]);
    exit(1);
}
'''
    
    with open(PHP_BENCHMARK_SCRIPT, 'w') as f:
        f.write(php_code)
    
    os.chmod(PHP_BENCHMARK_SCRIPT, 0o755)
    return True


def benchmark_php_parser(file_path):
    """Benchmark the PHP parser."""
    if not os.path.exists(PHP_BENCHMARK_SCRIPT):
        return {
            'success': False,
            'duration_ms': 0,
            'duration_ns': 0,
            'error': 'PHP benchmark script not found'
        }
    
    try:
        result = subprocess.run(
            ['php', PHP_BENCHMARK_SCRIPT, str(file_path)],
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
            'error': f'JSON decode error: {e}. Output: {result.stdout[:200]}'
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


def format_speedup(go_time, php_time):
    """Calculate and format speedup ratio."""
    if go_time == 0:
        return "N/A"
    speedup = php_time / go_time
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
    print("GEDCOM Parser Performance Comparison: Go vs PHP")
    print("=" * 80)
    print()
    
    # Check PHP version
    php_ok, php_version, php_84_plus = check_php_version()
    if not php_ok:
        print(f"⚠ WARNING: PHP 8.3+ required for php-gedcom, but found PHP {php_version}")
        print("  The PHP parser benchmarks will be skipped.")
        print("  Install PHP 8.3+ to run comparison.")
        print()
        php_available = False
    else:
        if php_84_plus:
            print(f"✓ PHP {php_version} detected (compatible with php-gedcom v4.0+)")
        else:
            print(f"✓ PHP {php_version} detected (will use php-gedcom v2.2.0)")
        php_available = True
        print()
    
    # Build Go parser
    print("Building Go parser benchmark binary...")
    if not build_go_parser():
        print("ERROR: Failed to build Go parser. Exiting.")
        sys.exit(1)
    print("✓ Go parser built successfully")
    print()
    
    # Create PHP benchmark script
    if php_available:
        print("Creating PHP benchmark script...")
        create_php_benchmark_script()
        print("✓ PHP benchmark script created")
        
        # Check if php-gedcom is installed
        if not os.path.exists('/apps/php-gedcom/vendor/autoload.php'):
            print("⚠ WARNING: php-gedcom not installed. Installing...")
            result = subprocess.run(
                ['composer', 'install', '--no-dev'],
                cwd='/apps/php-gedcom',
                capture_output=True,
                text=True
            )
            if result.returncode != 0:
                print(f"ERROR: Failed to install php-gedcom: {result.stderr}")
                php_available = False
            else:
                print("✓ php-gedcom installed")
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
        
        # Go parser
        print("  Running Go parser...", end=" ", flush=True)
        go_result = benchmark_go_parser(file_path)
        if go_result['success']:
            print(f"✓ {format_duration(go_result['duration_ms'])}")
        else:
            print(f"✗ ERROR: {go_result['error']}")
        
        # PHP parser
        if php_available:
            print("  Running PHP parser...", end=" ", flush=True)
            php_result = benchmark_php_parser(file_path)
            if php_result['success']:
                print(f"✓ {format_duration(php_result['duration_ms'])}")
            else:
                print(f"✗ ERROR: {php_result['error']}")
        else:
            php_result = {'success': False, 'error': 'PHP 8.4+ not available'}
        
        # Calculate speedup
        if go_result['success'] and php_result['success']:
            speedup = format_speedup(go_result['duration_ms'], php_result['duration_ms'])
            print(f"  Speedup: Go is {speedup} faster")
        
        # Store results
        results.append({
            'filename': filename,
            'file_size': get_file_size(file_path),
            'go': go_result,
            'php': php_result,
        })
        
        print()
    
    # Print summary table
    print("=" * 80)
    print("SUMMARY")
    print("=" * 80)
    print()
    print(f"{'File':<20} {'Size':<10} {'Go (ms)':<15} {'PHP (ms)':<15} {'Speedup':<15}")
    print("-" * 80)
    
    go_times = []
    php_times = []
    
    for result in results:
        filename = result['filename']
        file_size = result['file_size']
        
        if result['go']['success']:
            go_time = result['go']['duration_ms']
            go_times.append(go_time)
            go_str = f"{go_time:.2f}"
        else:
            go_str = "ERROR"
        
        if result['php']['success']:
            php_time = result['php']['duration_ms']
            php_times.append(php_time)
            php_str = f"{php_time:.2f}"
        else:
            php_str = "ERROR" if not php_available else "ERROR"
        
        if result['go']['success'] and result['php']['success']:
            speedup = format_speedup(go_time, php_time)
        else:
            speedup = "N/A"
        
        print(f"{filename:<20} {file_size:<10} {go_str:<15} {php_str:<15} {speedup:<15}")
    
    print("-" * 80)
    
    # Calculate averages
    if go_times and php_times:
        avg_go = statistics.mean(go_times)
        avg_php = statistics.mean(php_times)
        avg_speedup = format_speedup(avg_go, avg_php)
        
        print(f"{'AVERAGE':<20} {'':<10} {avg_go:<15.2f} {avg_php:<15.2f} {avg_speedup:<15}")
        print()
        
        # Overall statistics
        print("Overall Statistics:")
        print(f"  Average Go time: {format_duration(avg_go)}")
        print(f"  Average PHP time: {format_duration(avg_php)}")
        print(f"  Average speedup: {avg_speedup}")
        
        if len(go_times) > 1 and len(php_times) > 1:
            median_go = statistics.median(go_times)
            median_php = statistics.median(php_times)
            median_speedup = format_speedup(median_go, median_php)
            print(f"  Median Go time: {format_duration(median_go)}")
            print(f"  Median PHP time: {format_duration(median_php)}")
            print(f"  Median speedup: {median_speedup}")
    elif go_times:
        avg_go = statistics.mean(go_times)
        print(f"{'AVERAGE (Go only)':<20} {'':<10} {avg_go:<15.2f} {'N/A':<15} {'N/A':<15}")
    
    print()
    print("=" * 80)
    
    # Detailed results
    print("\nDETAILED RESULTS")
    print("=" * 80)
    for result in results:
        print(f"\n{result['filename']} ({result['file_size']}):")
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
        
        print(f"  PHP:")
        if result['php']['success']:
            print(f"    Time: {format_duration(result['php']['duration_ms'])}")
            if 'individuals' in result['php']:
                print(f"    Individuals: {result['php']['individuals']}")
                print(f"    Families: {result['php']['families']}")
        else:
            print(f"    ERROR: {result['php']['error']}")
    
    # Cleanup
    if os.path.exists(GO_PARSER_BINARY):
        os.remove(GO_PARSER_BINARY)
    if os.path.exists('/tmp/go_parser_benchmark.go'):
        os.remove('/tmp/go_parser_benchmark.go')
    if os.path.exists(PHP_BENCHMARK_SCRIPT):
        os.remove(PHP_BENCHMARK_SCRIPT)


if __name__ == '__main__':
    main()

