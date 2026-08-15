package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/peterintech/nosleepp/internal/watch"

	"github.com/peterintech/nosleepp/internal/process"

	"github.com/peterintech/nosleepp/internal/power"

	"github.com/peterintech/nosleepp/internal/agent"

	"github.com/peterintech/nosleepp/internal/defaults"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type ExitError struct {
	Code    int
	Message string
}

func (e ExitError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

type options struct {
	interval     time.Duration
	sample       time.Duration
	cpuThreshold time.Duration
	quiet        time.Duration
	agentFlags   []string
	configPath   string
	jsonOutput   bool
	includeAll   bool
	once         bool
	output       io.Writer
	errorOutput  io.Writer
	processScan  watch.ProcessScanner
	powerManager watch.PowerManager
}

func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	return NewRootCommand(os.Stdout, os.Stderr).ExecuteContext(ctx)
}

func NewRootCommand(stdout, stderr io.Writer) *cobra.Command {
	opts := &options{
		interval:     defaults.Interval,
		sample:       defaults.Sample,
		cpuThreshold: defaults.CPUThreshold(),
		quiet:        defaults.Quiet,
		output:       stdout,
		errorOutput:  stderr,
	}
	return newRootCommand(opts)
}

func newRootCommand(opts *options) *cobra.Command {
	root := &cobra.Command{
		Use:           "nosleepp",
		Short:         "Keep your PC awake while AI agents are working",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringArrayVar(&opts.agentFlags, "agent", nil, "agent signature in name=process1,process2 format")
	root.PersistentFlags().StringVar(&opts.configPath, "config", "", "optional config path reserved for future use")

	root.AddCommand(newListCommand(opts))
	root.AddCommand(newWatchCommand(opts))
	root.AddCommand(newPowerTestCommand(opts))
	root.AddCommand(newVersionCommand(opts.output))

	return root
}

func newListCommand(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List working agent processes",
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := buildProfiles(opts.agentFlags)
			if err != nil {
				return ExitError{Code: 2, Message: err.Error()}
			}

			scanner := opts.processScan
			if scanner == nil {
				scanner = process.NewScanner()
			}

			if !opts.jsonOutput {
				fmt.Fprintf(opts.output, "Checking agents for activity over %s...\n", opts.sample)
			}

			before, err := scanner.Scan(cmd.Context())
			if err != nil {
				return err
			}

			if err := sleepContext(cmd.Context(), opts.sample); err != nil {
				return err
			}

			after, err := scanner.Scan(cmd.Context())
			if err != nil {
				return err
			}

			matches := agent.DetectActivity(profiles, before, after, agent.ActivityOptions{
				CPUThreshold: opts.cpuThreshold,
				IncludeIdle:  opts.includeAll,
			})
			return printMatches(opts.output, matches, opts.jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "print machine-readable JSON")
	cmd.Flags().DurationVar(&opts.sample, "sample", defaults.Sample, "activity sample window")
	cmd.Flags().DurationVar(&opts.cpuThreshold, "cpu-threshold", defaults.CPUThreshold(), "minimum CPU delta for working status")
	cmd.Flags().BoolVar(&opts.includeAll, "all", false, "include idle matching agent processes")
	return cmd
}

func newWatchCommand(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Prevent system sleep while agents are working",
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := buildProfiles(opts.agentFlags)
			if err != nil {
				return ExitError{Code: 2, Message: err.Error()}
			}

			scanner := opts.processScan
			if scanner == nil {
				scanner = process.NewScanner()
			}

			powerManager := opts.powerManager
			if powerManager == nil {
				powerManager = power.NewManager()
			}

			firstCheck := true
			watcher := watch.NewWatcher(scanner, powerManager, profiles, watch.Options{
				Interval:     opts.interval,
				Sample:       opts.sample,
				CPUThreshold: opts.cpuThreshold,
				Quiet:        opts.quiet,
				Once:         opts.once,
				OnCheck: func(sample time.Duration) {
					if firstCheck {
						firstCheck = false
						if !opts.jsonOutput {
							fmt.Fprintln(opts.output, "Checking agents for activity...")
						}
					}
				},
				OnChange: func(matches []agent.Match, state watch.State) {
					switch state {
					case watch.StateWorking:
						fmt.Fprintln(opts.output, "Agents running; preventing system sleep.")
					case watch.StateQuiet:
						fmt.Fprintf(opts.output, "No current agent activity; keeping PC awake for the %s quiet window.\n", opts.quiet)
					case watch.StateReleased:
						fmt.Fprintln(opts.output, "No agents running; normal sleep behavior restored.")
					}
					_ = printMatches(opts.output, matches, opts.jsonOutput)
				},
			})

			err = watcher.Run(cmd.Context())
			if errors.Is(err, context.Canceled) {
				return nil
			}
			if errors.Is(err, watch.ErrNoAgents) {
				if opts.once {
					return ExitError{Code: 1, Message: "no working agents found"}
				}
				return nil
			}
			if err != nil {
				return ExitError{Code: 3, Message: err.Error()}
			}
			return nil
		},
	}

	cmd.Flags().DurationVar(&opts.interval, "interval", defaults.Interval, "polling interval")
	cmd.Flags().DurationVar(&opts.sample, "sample", defaults.Sample, "activity sample window")
	cmd.Flags().DurationVar(&opts.cpuThreshold, "cpu-threshold", defaults.CPUThreshold(), "minimum CPU delta for working status")
	cmd.Flags().DurationVar(&opts.quiet, "quiet", defaults.Quiet, "quiet grace period before releasing sleep prevention")
	cmd.Flags().BoolVar(&opts.once, "once", false, "check once and exit")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "print machine-readable JSON")
	return cmd
}

func newPowerTestCommand(opts *options) *cobra.Command {
	var duration time.Duration
	cmd := &cobra.Command{
		Use:   "power-test",
		Short: "Hold the no-sleep lock for a fixed duration",
		RunE: func(cmd *cobra.Command, args []string) error {
			powerManager := opts.powerManager
			if powerManager == nil {
				powerManager = power.NewManager()
			}

			if duration <= 0 {
				return ExitError{Code: 2, Message: "--duration must be greater than 0"}
			}

			if err := powerManager.Acquire(); err != nil {
				return ExitError{Code: 3, Message: err.Error()}
			}
			fmt.Fprintf(opts.output, "No-sleep lock acquired for %s. Do not close this terminal during the test.\n", duration)
			defer func() {
				if err := powerManager.Release(); err != nil {
					fmt.Fprintf(opts.errorOutput, "Failed to release no-sleep lock: %v\n", err)
				} else {
					fmt.Fprintln(opts.output, "No-sleep lock released.")
				}
			}()

			err := sleepContext(cmd.Context(), duration)
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		},
	}

	cmd.Flags().DurationVar(&duration, "duration", defaults.PowerTestDuration, "how long to hold the no-sleep lock")
	return cmd
}

func newVersionCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(stdout, "nosleepp %s (commit %s, built %s)\n", version, commit, date)
			return err
		},
	}
}

func buildProfiles(agentFlags []string) ([]agent.Profile, error) {
	profiles := agent.DefaultProfiles()
	for _, flag := range agentFlags {
		profile, err := parseAgentFlag(flag)
		if err != nil {
			return nil, err
		}
		profiles = agent.UpsertProfile(profiles, profile)
	}
	return profiles, nil
}

func parseAgentFlag(value string) (agent.Profile, error) {
	name, processes, ok := strings.Cut(value, "=")
	name = strings.TrimSpace(name)
	if !ok || name == "" {
		return agent.Profile{}, fmt.Errorf("invalid --agent %q: expected name=process1,process2", value)
	}

	var names []string
	for _, processName := range strings.Split(processes, ",") {
		processName = strings.TrimSpace(processName)
		if processName != "" {
			names = append(names, processName)
		}
	}
	if len(names) == 0 {
		return agent.Profile{}, fmt.Errorf("invalid --agent %q: at least one process name is required", value)
	}

	return agent.Profile{Name: name, Processes: names}, nil
}

func printMatches(w io.Writer, matches []agent.Match, asJSON bool) error {
	if asJSON {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(matches)
	}

	if len(matches) == 0 {
		_, err := fmt.Fprintln(w, "No working agents found.")
		return err
	}

	_, err := fmt.Fprintln(w, "AGENT\tPID\tPROCESS\tSTATUS\tCPU_DELTA\tCHILDREN\tEVIDENCE")
	if err != nil {
		return err
	}
	for _, match := range matches {
		evidence := strings.Join(match.Evidence, ",")
		if evidence == "" {
			evidence = "-"
		}
		if _, err := fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%dms\t%d\t%s\n", match.Agent, match.PID, match.Executable, match.Status, match.CPUDeltaMS, match.Children, evidence); err != nil {
			return err
		}
	}
	return nil
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
