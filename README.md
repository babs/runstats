# runstats

A lightweight command wrapper that measures and reports resource usage metrics for any command.

## Overview

`runstats` executes a command and collects detailed resource usage statistics, including memory consumption, CPU time, I/O operations, and context switches. It outputs metrics as JSON to stderr while transparently passing through the command's stdin, stdout, and stderr.

## Features

- **Memory tracking**: Peak RSS (Resident Set Size) in kilobytes
- **CPU time**: User, system, and total CPU time in milliseconds
- **Wall clock time**: Actual elapsed time in milliseconds
- **Page faults**: Minor and major page faults
- **I/O operations**: Block input and output operations
- **Context switches**: Voluntary and involuntary context switches
- **Child process tracking**: All metrics include the main process and all its child processes
- **Environment capture**: Optional environment variable dumping with `-dump-env` flag
- **Exit code preservation**: Returns the wrapped command's exit code
- **Transparent I/O**: Command's stdin/stdout/stderr work normally
- **JSON output**: Machine-readable metrics on stderr

## Building

Simple build:
```bash
go build
```

Build with version info (recommended):
```bash
./build.sh
```

This produces a `runstats` binary in the current directory. The build script injects version information from git tags.

### Docker Build

Build the Docker image:
```bash
docker build -t runstats:latest .
```

Build with version tag:
```bash
docker build --build-arg VERSION=v1.0.0 -t runstats:v1.0.0 .
```

## Usage

```bash
runstats [options] <command> [args...]
```

### Options

- `-dump-env`: Include environment variables in the JSON output

### Examples

Simple command:
```bash
./runstats sleep 1
```

Command with arguments:
```bash
./runstats ls -lah /tmp
```

Pipeline usage:
```bash
echo "test" | ./runstats grep "test"
```

Capturing metrics:
```bash
./runstats go build 2> metrics.json
```

With environment dump:
```bash
./runstats -dump-env ./myscript.sh 2> metrics.json
```

Child process tracking (shell spawning subprocesses):
```bash
./runstats sh -c 'for i in $(seq 1 1000000); do : ; done'
```
This example shows metrics for both the shell and the `seq` subprocess it spawns.

### Docker Usage

Run commands in a container:
```bash
docker run --rm runstats:latest sleep 1
```

With arguments:
```bash
docker run --rm runstats:latest ls -lah /app
```

Save metrics to file:
```bash
docker run --rm runstats:latest sleep 1 2> metrics.json
```

## Output Format

Metrics are output as JSON to stderr after the command completes:

```json
{
  "app_name": "runstats",
  "app_version": "v1.0.0",
  "command": "go",
  "args": ["build"],
  "wall_clock_time_ms": 1000,
  "exit_code": 0,
  "main_process": {
    "peak_rss_kb": 2048,
    "cpu_user_time_ms": 150,
    "cpu_system_time_ms": 50,
    "cpu_total_time_ms": 200,
    "minor_page_faults": 256,
    "major_page_faults": 0,
    "block_input_ops": 8,
    "block_output_ops": 16,
    "voluntary_context_switches": 10,
    "involuntary_context_switches": 2
  },
  "children": {
    "peak_rss_kb": 512,
    "cpu_user_time_ms": 25,
    "cpu_system_time_ms": 10,
    "cpu_total_time_ms": 35,
    "minor_page_faults": 64,
    "major_page_faults": 0,
    "block_input_ops": 2,
    "block_output_ops": 4,
    "voluntary_context_switches": 3,
    "involuntary_context_switches": 1
  },
  "total": {
    "peak_rss_kb": 2560,
    "cpu_user_time_ms": 175,
    "cpu_system_time_ms": 60,
    "cpu_total_time_ms": 235,
    "minor_page_faults": 320,
    "major_page_faults": 0,
    "block_input_ops": 10,
    "block_output_ops": 20,
    "voluntary_context_switches": 13,
    "involuntary_context_switches": 3
  },
  "environment": {
    "PATH": "/usr/bin:/bin",
    "HOME": "/home/user"
  }
}
```

Note: The `environment` field is only included when the `-dump-env` flag is used.

### Metrics Explained

#### Top-level fields

- **app_name**: Name of the runstats application
- **app_version**: Version of runstats (from git tags or "dev")
- **command**: The command that was executed
- **args**: Array of arguments passed to the command
- **wall_clock_time_ms**: Actual elapsed time
- **exit_code**: Command's exit code (0 = success)
- **environment**: Map of environment variables (only with `-dump-env` flag)

#### Resource usage fields

The following metrics are provided for three different scopes:

- **main_process**: Metrics for the main command process only
- **children**: Aggregated metrics for all child processes spawned by the main process
- **total**: Combined metrics (main process + all children)

Each resource usage object contains:

- **peak_rss_kb**: Maximum resident set size (memory) in kilobytes
- **cpu_user_time_ms**: CPU time spent in user mode
- **cpu_system_time_ms**: CPU time spent in kernel mode
- **cpu_total_time_ms**: Total CPU time (user + system)
- **minor_page_faults**: Page faults serviced without I/O
- **major_page_faults**: Page faults requiring disk I/O
- **block_input_ops**: Block device read operations
- **block_output_ops**: Block device write operations
- **voluntary_context_switches**: Process yielded CPU voluntarily
- **involuntary_context_switches**: Process was preempted

## Use Cases

- **Performance monitoring**: Track resource usage of build processes, tests, or deployments
- **Benchmarking**: Compare resource consumption across different implementations
- **CI/CD integration**: Monitor resource trends over time
- **Debugging**: Identify memory leaks or excessive I/O
- **Profiling**: Lightweight alternative to full profilers for quick metrics

## Platform Support

Linux-only. Uses `syscall.Wait4()` and `rusage` which are Unix/Linux specific.

## License

See LICENSE file for details.
