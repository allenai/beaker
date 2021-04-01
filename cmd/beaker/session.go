package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/allenai/bytefmt"
	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/beaker/runtime"
	"github.com/beaker/runtime/docker"
	"github.com/spf13/cobra"
)

const (
	// Label containing the session ID on session containers.
	sessionContainerLabel = "beaker.org/session"

	// Label containing a list of the GPUs assigned to the container e.g. "1,2".
	sessionGPULabel = "beaker.org/gpus"
)

func newSessionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session <command>",
		Short: "Manage sessions",
	}
	cmd.AddCommand(newSessionAttachCommand())
	cmd.AddCommand(newSessionCreateCommand())
	cmd.AddCommand(newSessionExecCommand())
	cmd.AddCommand(newSessionGetCommand())
	cmd.AddCommand(newSessionListCommand())
	cmd.AddCommand(newSessionUpdateCommand())
	return cmd
}

func newSessionAttachCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "attach <session>",
		Short: "Attach to a running session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := findRunningContainer(args[0])
			if err != nil {
				return err
			}
			return handleAttachErr(container.(*docker.Container).Attach(ctx))
		},
	}
}

func newSessionCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <command...>",
		Short: "Create a new interactive session",
		Long: `Create a new interactive session backed by a Docker container.

Arguments are passed to the Docker container as a command.
To pass flags, use "--" e.g. "create -- ls -l"`,
		Args: cobra.ArbitraryArgs,
	}

	var gpus int
	var image string
	var name string
	var node string
	cmd.Flags().IntVar(&gpus, "gpus", 0, "Number of GPUs assigned to the session")
	cmd.Flags().StringVar(
		&image,
		"image",
		"allenai/base:cuda11.2-ubuntu20.04",
		"Docker image for the session.")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Assign a name to the session")
	cmd.Flags().StringVar(&node, "node", "", "Node that the session will run on. Defaults to current node.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if node == "" {
			var err error
			node, err = getCurrentNode()
			if err != nil {
				return fmt.Errorf("failed to detect node; use --node flag: %w", err)
			}
		}

		session, err := beaker.CreateSession(ctx, api.SessionSpec{
			Name: name,
			Node: node,
			Resources: &api.TaskResources{
				GPUCount: gpus,
			},
		})
		if err != nil {
			return err
		}

		// Pass nil instead of empty slice when there are no arguments.
		var command []string
		if len(args) > 0 {
			command = args
		}

		if err := startSession(session.ID, image, command); err != nil {
			// If we fail to start the session, cancel it so that the executor
			// can immediately reclaim the resources allocated to it.
			//
			// Use context.Background() since ctx may already be canceled.
			_, _ = beaker.Session(session.ID).Patch(context.Background(), api.SessionPatch{
				State: &api.ExecutionState{
					Canceled: now(),
				},
			})
			return err
		}
		return nil
	}
	return cmd
}

func newSessionExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec <session> <command> <args...>",
		Short: "Execute a command in a session",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := findRunningContainer(args[0])
			if err != nil {
				return err
			}

			// Pass nil instead of empty slice when there are no arguments.
			var command []string
			if len(args) > 1 {
				command = args[1:]
			}

			err = container.(*docker.Container).Exec(ctx, &docker.ExecOpts{
				Command: command,
			})
			return handleAttachErr(err)
		},
	}
}

func newSessionGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <session...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more sessions",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessions []api.Session
			for _, id := range args {
				info, err := beaker.Session(id).Get(ctx)
				if err != nil {
					return err
				}
				sessions = append(sessions, *info)
			}
			return printSessions(sessions)
		},
	}
}

func newSessionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sessions",
		Args:  cobra.NoArgs,
	}

	var all bool
	var cluster string
	var node string
	var finalized bool
	cmd.Flags().BoolVar(&all, "all", false, "List all sessions.")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster to list sessions.")
	cmd.Flags().StringVar(&node, "node", "", "Node to list sessions. Defaults to current node.")
	cmd.Flags().BoolVar(&finalized, "finalized", false, "Show only finalized sessions")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var opts client.ListSessionOpts
		if !all {
			opts.Finalized = &finalized

			if cluster != "" {
				opts.Cluster = &cluster
			}

			if !cmd.Flag("node").Changed && cluster == "" {
				var err error
				node, err = getCurrentNode()
				if err != nil {
					return fmt.Errorf("failed to detect node; use --node flag: %w", err)
				}
			}
			if node != "" {
				opts.Node = &node
			}
		}

		sessions, err := beaker.ListSessions(ctx, &opts)
		if err != nil {
			return err
		}
		return printSessions(sessions)
	}
	return cmd
}

func newSessionUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a session",
		Args:  cobra.ExactArgs(1),
	}

	var cancel bool
	cmd.Flags().BoolVar(&cancel, "cancel", false, "Cancel a session")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		patch := api.SessionPatch{
			State: &api.ExecutionState{},
		}
		if cancel {
			patch.State.Canceled = now()
		}

		session, err := beaker.Session(args[0]).Patch(ctx, patch)
		if err != nil {
			return err
		}
		return printSessions([]api.Session{*session})
	}
	return cmd
}

func awaitSessionSchedule(session string) (*api.Session, error) {
	if !quiet {
		fmt.Printf("Waiting for session to be scheduled")
	}
	delay := time.NewTimer(0) // No delay on first attempt.
	for attempt := 0; ; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-delay.C:
			info, err := beaker.Session(session).Get(ctx)
			if err != nil {
				return nil, err
			}

			if info.State.Scheduled != nil {
				if !quiet {
					fmt.Println()
				}
				return info, nil
			}
			if !quiet {
				fmt.Print(".")
			}
			delay.Reset(time.Second)
		}
	}
}

func startSession(
	sessionID string,
	image string,
	command []string,
) error {
	session, err := awaitSessionSchedule(sessionID)
	if err != nil {
		return err
	}

	if !quiet && session.Limits != nil {
		fmt.Printf(
			"Reserved %d GPU, %v CPU, %.1fGiB memory\n",
			len(session.Limits.GPUs),
			session.Limits.CPUCount,
			// TODO Use friendly formatting from bytefmt when available.
			float64(session.Limits.Memory.Int64())/float64(bytefmt.GiB))
	}

	labels := map[string]string{
		sessionContainerLabel: session.ID,
		sessionGPULabel:       strings.Join(session.Limits.GPUs, ","),
	}

	u, err := user.Current()
	if err != nil {
		return err
	}

	env := make(map[string]string)
	var mounts []runtime.Mount
	if u.HomeDir != "" {
		env["HOME"] = u.HomeDir
		mounts = append(mounts, runtime.Mount{
			HostPath:      u.HomeDir,
			ContainerPath: u.HomeDir,
		})
	}
	if _, err := os.Stat("/net"); !os.IsNotExist(err) {
		// Mount in /net for NFS.
		mounts = append(mounts, runtime.Mount{
			HostPath:      "/net",
			ContainerPath: "/net",
		})
	}

	opts := &runtime.ContainerOpts{
		Name: strings.ToLower("session-" + session.ID),
		Image: &runtime.DockerImage{
			Tag: image,
		},
		Command:     command,
		Labels:      labels,
		Env:         env,
		Mounts:      mounts,
		CPUCount:    session.Limits.CPUCount,
		GPUs:        session.Limits.GPUs,
		Memory:      session.Limits.Memory.Int64(),
		Interactive: true,
		User:        u.Uid + ":" + u.Gid,
		WorkingDir:  u.HomeDir,
	}

	rt, err := docker.NewRuntime()
	if err != nil {
		return err
	}

	if !quiet {
		fmt.Println("Pulling image...")
	}
	if err := rt.PullImage(ctx, opts.Image, quiet); err != nil {
		return err
	}

	container, err := rt.CreateContainer(ctx, opts)
	if err != nil {
		return err
	}

	if err := container.Start(ctx); err != nil {
		return err
	}
	return handleAttachErr(container.(*docker.Container).Attach(ctx))
}

func handleAttachErr(err error) error {
	if err != nil && strings.HasPrefix(err.Error(), "exited with code ") {
		// Ignore errors coming from the container.
		// If the user exits using Ctrl-C, attach will return an error like:
		// "exited with code 130".
		return nil
	}
	return err
}

func findRunningContainer(session string) (runtime.Container, error) {
	info, err := beaker.Session(session).Get(ctx)
	if err != nil {
		return nil, err
	}
	if info.State.Started == nil {
		return nil, fmt.Errorf("session not started")
	}
	if info.State.Ended != nil {
		return nil, fmt.Errorf("session already ended")
	}
	if info.State.Finalized != nil {
		return nil, fmt.Errorf("session already finalized")
	}

	rt, err := docker.NewRuntime()
	if err != nil {
		return nil, err
	}

	containers, err := rt.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	var container runtime.Container
	for _, c := range containers {
		info, err := c.Info(ctx)
		if err != nil {
			return nil, err
		}

		if session == info.Labels[sessionContainerLabel] {
			container = c
			break
		}
	}
	if container == nil {
		return nil, fmt.Errorf("container not found")
	}
	return container, nil
}

func now() *time.Time {
	now := time.Now()
	return &now
}
