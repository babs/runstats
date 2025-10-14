package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var (
	Version        = "v0.0.0"
	CommitHash     = "0000000"
	BuildTimestamp = "1970-01-01T00:00:00"
	Builder        = "unknown"
	ProjectURL     = "https://github.com/babs/runstats"
)

type cliOpts struct {
	dumpEnv bool
}

type ResourceUsage struct {
	PeakRSSKB        int64 `json:"peak_rss_kb"`
	CPUUserTimeMS    int64 `json:"cpu_user_time_ms"`
	CPUSystemTimeMS  int64 `json:"cpu_system_time_ms"`
	CPUTotalTimeMS   int64 `json:"cpu_total_time_ms"`
	MinorPageFaults  int64 `json:"minor_page_faults"`
	MajorPageFaults  int64 `json:"major_page_faults"`
	BlockInputOps    int64 `json:"block_input_ops"`
	BlockOutputOps   int64 `json:"block_output_ops"`
	VoluntaryCtxSw   int64 `json:"voluntary_context_switches"`
	InvoluntaryCtxSw int64 `json:"involuntary_context_switches"`
}

func rusageToResourceUsage(ru *syscall.Rusage) ResourceUsage {
	userTimeMS := ru.Utime.Sec*1000 + int64(ru.Utime.Usec/1000)
	systemTimeMS := ru.Stime.Sec*1000 + int64(ru.Stime.Usec/1000)
	return ResourceUsage{
		PeakRSSKB:        ru.Maxrss,
		CPUUserTimeMS:    userTimeMS,
		CPUSystemTimeMS:  systemTimeMS,
		CPUTotalTimeMS:   userTimeMS + systemTimeMS,
		MinorPageFaults:  ru.Minflt,
		MajorPageFaults:  ru.Majflt,
		BlockInputOps:    ru.Inblock,
		BlockOutputOps:   ru.Oublock,
		VoluntaryCtxSw:   ru.Nvcsw,
		InvoluntaryCtxSw: ru.Nivcsw,
	}
}

type Metrics struct {
	AppName         string            `json:"app_name"`
	AppVersion      string            `json:"app_version"`
	Command         string            `json:"command"`
	Args            []string          `json:"args"`
	WallClockTimeMS int64             `json:"wall_clock_time_ms"`
	ExitCode        int               `json:"exit_code"`
	MainProcess     ResourceUsage     `json:"main_process"`
	Children        ResourceUsage     `json:"children"`
	Total           ResourceUsage     `json:"total"`
	Environment     map[string]string `json:"environment,omitempty"`
}

func main() {
	opts := cliOpts{}
	flag.BoolVar(&opts.dumpEnv, "dump-env", false, "Include environment variables in output")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startTime := time.Now()

	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start command: %v\n", err)
		os.Exit(1)
	}

	var mainUsage syscall.Rusage
	var wstatus syscall.WaitStatus
	pid := cmd.Process.Pid
	_, waitErr := syscall.Wait4(pid, &wstatus, 0, &mainUsage)
	if waitErr != nil {
		_ = cmd.Wait()
	}

	wallClockTimeMS := time.Since(startTime).Milliseconds()

	var allChildrenUsage syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_CHILDREN, &allChildrenUsage)

	mainProc := rusageToResourceUsage(&mainUsage)
	allChildren := rusageToResourceUsage(&allChildrenUsage)

	// RUSAGE_CHILDREN includes the main process we just reaped, so subtract it
	children := ResourceUsage{
		PeakRSSKB:        allChildren.PeakRSSKB - mainProc.PeakRSSKB,
		CPUUserTimeMS:    allChildren.CPUUserTimeMS - mainProc.CPUUserTimeMS,
		CPUSystemTimeMS:  allChildren.CPUSystemTimeMS - mainProc.CPUSystemTimeMS,
		CPUTotalTimeMS:   allChildren.CPUTotalTimeMS - mainProc.CPUTotalTimeMS,
		MinorPageFaults:  allChildren.MinorPageFaults - mainProc.MinorPageFaults,
		MajorPageFaults:  allChildren.MajorPageFaults - mainProc.MajorPageFaults,
		BlockInputOps:    allChildren.BlockInputOps - mainProc.BlockInputOps,
		BlockOutputOps:   allChildren.BlockOutputOps - mainProc.BlockOutputOps,
		VoluntaryCtxSw:   allChildren.VoluntaryCtxSw - mainProc.VoluntaryCtxSw,
		InvoluntaryCtxSw: allChildren.InvoluntaryCtxSw - mainProc.InvoluntaryCtxSw,
	}

	total := allChildren

	exitCode := 0
	if waitErr == nil {
		if wstatus.Exited() {
			exitCode = wstatus.ExitStatus()
		} else if wstatus.Signaled() {
			exitCode = 128 + int(wstatus.Signal())
		}
	}

	metrics := Metrics{
		AppName:         "runstats",
		AppVersion:      Version,
		Command:         cmdName,
		Args:            cmdArgs,
		WallClockTimeMS: wallClockTimeMS,
		ExitCode:        exitCode,
		MainProcess:     mainProc,
		Children:        children,
		Total:           total,
	}

	if opts.dumpEnv {
		metrics.Environment = make(map[string]string)
		for _, env := range os.Environ() {
			var key, value string
			for i := 0; i < len(env); i++ {
				if env[i] == '=' {
					key = env[:i]
					value = env[i+1:]
					break
				}
			}
			if key != "" {
				metrics.Environment[key] = value
			}
		}
	}

	jsonOutput, err := json.Marshal(metrics)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal metrics to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "%s\n", string(jsonOutput))

	os.Exit(exitCode)
}
