package supervise

import (
	"context"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	reaper "github.com/ramr/go-reaper"
	log "github.com/sirupsen/logrus"
)

// NewSupervisor creates a supervisor instance
func NewSupervisor(config *Config) *Supervisor {
	if !config.DisableReaper {
		log.Info("Starting process reaper")
		go reaper.Reap()
	}

	stateCh := make(chan *os.ProcessState, 1)

	return &Supervisor{
		config:  config,
		stateCh: stateCh,
		lastEnv: nil,
	}
}

// Start spawns the specified child process and runs a goroutine with
// the subprocess reaper
func (s *Supervisor) Start(ctx context.Context, updateCh chan []string) error {
	log.Info("Starting supervisor")

	var child *exec.Cmd
	var err error

	childCtx, childCancel := context.WithCancel(ctx)

	for {
		select {
		case envUpdate := <-updateCh:
			s.lastEnv = envUpdate

			if child == nil {
				log.Infof("Received initial environment update, spawning the child!")
				log.Debugf("<- envUpdate: %#v", envUpdate)

				// Perform the initial child spawn
				child, err = s.spawnChild(childCtx, envUpdate)
				if err != nil {
					log.WithError(err).Errorf("Could not spawn child")
				}

				continue
			}

			log.Debugf("Got an environment update, restarting the child! %#v\n", envUpdate)
			child, childCtx, childCancel, err = s.restartChild(ctx, childCancel, envUpdate)
			if err != nil {
				log.WithError(err).Errorf("Could not restart child")
			}
		case childState := <-s.stateCh:
			log.WithFields(log.Fields{
				"pid":        childState.Pid(),
				"success":    childState.Success(),
				"exitCode":   childState.ExitCode(),
				"systemTime": childState.SystemTime().String(),
				"userTime":   childState.UserTime(),
			}).Debugf("Child process died; restarting")

			child, childCtx, childCancel, err = s.restartChild(ctx, childCancel, s.lastEnv)
			if err != nil {
				log.WithError(err).Errorf("Could not restart child")
			}
		case <-ctx.Done():
			log.Infof("Supervisor received shutdown signal")

			childCancel()
			return s.haltAndWaitChild(child)
		}
	}
}

// spawnChild spawns the child process, returning a running exec.Cmd instance
func (s *Supervisor) spawnChild(ctx context.Context, environ []string) (*exec.Cmd, error) {
	program, err := s.config.Program()
	if err != nil {
		return nil, errors.Wrap(err, "could not determine path to child executable")
	}

	child := exec.CommandContext(ctx, program, s.config.Args()...)
	child.Env = environ

	stdoutPipe, err := child.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "could not get stdout pipe")
	}

	stderrPipe, err := child.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "could not get stderr pipe")
	}

	fwd := newForwarder(stdoutPipe, stderrPipe)
	fwd.Start(ctx)

	log.WithFields(log.Fields{
		"program": program,
		"args":    s.config.Args(),
	}).Debugf("Starting child")
	if err = child.Start(); err != nil {
		return nil, errors.Wrap(err, "could not spawn child process")
	}

	log.WithField("pid", child.Process.Pid).Debugf("Child started!")

	go s.waitChild(ctx, child)

	return child, nil
}

func (s *Supervisor) restartChild(parentCtx context.Context, prevCancel context.CancelFunc, environ []string) (*exec.Cmd, context.Context, context.CancelFunc, error) {
	// When restarting the child, the previous childCtx needs to be
	// cancelled and a new one needs to be created
	prevCancel()
	childCtx, childCancel := context.WithCancel(parentCtx)

	child, err := s.spawnChild(childCtx, environ)
	if err != nil {
		// If the child could not be restarted, cancel the above context
		childCancel()
		return nil, nil, nil, errors.Wrap(err, "could not restart child")
	}

	return child, childCtx, childCancel, nil
}

func (s *Supervisor) waitChild(ctx context.Context, child *exec.Cmd) {
	err := child.Wait()
	if err != nil {
		log.WithError(err).Errorf("Could not wait on child")
	}

	// This is a gross hack to let childCtx short circuit
	// responding on stateCh with a process state.
	//
	// When an environment update is received, the child is
	// restarted, which would cause this waiter to finish and
	// submit a process state, which would trigger another restart.
	nextState := make(chan *os.ProcessState, 1)
	nextState <- child.ProcessState

	select {
	case <-ctx.Done():
		return
	case state := <-nextState:
		s.stateCh <- state
	}
}

func (s *Supervisor) haltAndWaitChild(child *exec.Cmd) error {
	if child == nil {
		log.Debugf("Can not halt child; is nil")
		return nil
	}

	log.Infof("Waiting for child to halt")
	if err := child.Wait(); err != nil {
		if err, ok := err.(*exec.ExitError); ok == true {
			return errors.Wrapf(
				err,
				"while waiting for child to halt: [code %d] %s",
				err.ExitCode(),
				err.Stderr,
			)
		}

		return errors.Wrapf(err, "while waiting for child to halt")
	}

	return nil
}
