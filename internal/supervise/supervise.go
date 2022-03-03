package supervise

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	reaper "github.com/ramr/go-reaper"
	"github.com/sirupsen/logrus"

	"glow.dev.maio.me/seanj/vault-init/pkg/io/forwarder"
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
func (s *Supervisor) Start(parentCtx context.Context, updateCh chan []string) error {
	log.Info("Starting supervisor")

	var err error
	var stop bool = false

	supState := newState(parentCtx)

	for {
		select {
		case envUpdate := <-updateCh:
			stop, supState, err = s.handleEnvironmentUpdate(supState, envUpdate)
			if err != nil {
				log.WithError(err).Errorf("Error handling environment update")
				return errors.Wrapf(err, "error while handling environment update")
			}
		case childState := <-s.stateCh:
			stop, supState, err = s.handleChildStateUpdate(supState, childState)
			if err != nil {
				log.WithError(err).Errorf("Error handling state update")
				return errors.Wrapf(err, "error while handling state update")
			}
		case <-parentCtx.Done():
			log.Infof("Supervisor received shutdown signal")
			stop = true
		}

		if stop {
			break
		}
	}

	// Always cancel the child context and attempt to wait it
	if supState != nil {
		supState.childCancel()
		return s.haltAndWaitChild(supState.child)
	}

	return nil
}

// handleEnvironmentUpdate accepts the `*state` structure and returns (stop bool, err error)
// When an environment update occurs, try to gracefully terminate the previous child, if any,
// and spawn a new child in its place.
func (s *Supervisor) handleEnvironmentUpdate(supState *state, envUpdate []string) (bool, *state, error) {
	s.lastEnv = envUpdate

	if supState.child == nil {
		log.Infof("Received initial environment update, spawning the child!")
		log.Debugf("envUpdate: %#v", envUpdate)

		// Perform the initial child spawn
		if err := s.spawnChild(supState, envUpdate); err != nil {
			log.WithError(err).Errorf("Could not spawn child")
			return true, supState, errors.Wrapf(err, "error spawning child")
		}

		return false, nil, nil
	}

	log.Debugf("Got an environment update, restarting the child! %#v\n", envUpdate)
	newChildState, err := s.restartChild(supState, envUpdate)
	if err != nil {
		log.WithError(err).Errorf("Could not restart child")
		return true, nil, errors.Wrapf(err, "error restarting child")
	}

	return false, newChildState, nil
}

func (s *Supervisor) handleChildStateUpdate(supState *state, childState *os.ProcessState) (bool, *state, error) {
	if s.config.OneShot {
		log.Debugf("Child process died; one-shot mode prevents restart!")
		return true, nil, nil
	}

	log.WithFields(logrus.Fields{
		"pid":        childState.Pid(),
		"success":    childState.Success(),
		"exitCode":   childState.ExitCode(),
		"systemTime": childState.SystemTime().String(),
		"userTime":   childState.UserTime(),
	}).Debugf("Child process died; restarting")

	newChildState, err := s.restartChild(supState, s.lastEnv)
	if err != nil {
		log.WithError(err).Errorf("Could not restart child")
		return true, nil, errors.Wrapf(err, "error restarting child")
	}

	return false, newChildState, nil
}

// spawnChild spawns the child process, writing the new exec.Cmd instance into the `*event` structure
// ctx context.Context -> evt.childCtx
func (s *Supervisor) spawnChild(supState *state, environ []string) error {
	program, err := s.config.Program()
	if err != nil {
		return errors.Wrap(err, "could not determine path to child executable")
	}

	child := exec.CommandContext(supState.childCtx, program, s.config.Args()...)
	child.Env = environ

	stdoutRPipe, stdoutWPipe := io.Pipe()
	child.Stdout = stdoutWPipe

	stderrRPipe, stderrWPipe := io.Pipe()
	child.Stderr = stderrWPipe

	stdoutFwd := forwarder.New(stdoutRPipe)
	stdoutFwd.Tee(log.WithField("stream", "stdout").WriterLevel(logrus.InfoLevel))
	stdoutFwd.Tee(s.config.ForwarderStdoutWriters...)
	go stdoutFwd.WaitClose()

	stderrFwd := forwarder.New(stderrRPipe)
	stderrFwd.Tee(log.WithField("stream", "stderr").WriterLevel(logrus.InfoLevel))
	stderrFwd.Tee(s.config.ForwarderStderrWriters...)
	go stderrFwd.WaitClose()

	stdoutFwd.Start(supState.childCtx)
	stderrFwd.Start(supState.childCtx)

	log.WithFields(logrus.Fields{
		"program": program,
		"args":    s.config.Args(),
	}).Debugf("Starting child")
	if err = child.Start(); err != nil {
		return errors.Wrap(err, "could not spawn child process")
	}

	log.WithField("pid", child.Process.Pid).Debugf("Child started!")
	supState.child = child

	go s.waitChild(supState)

	return nil
}

func (s *Supervisor) restartChild(supState *state, environ []string) (*state, error) {
	// When restarting the child, the previous child context needs to be
	// cancelled and a new one needs to be created
	// supState.replaceChildContext()
	newChildState := newState(supState.parentCtx)

	if err := s.spawnChild(newChildState, environ); err != nil {
		// If the child could not be restarted, cancel the above context
		newChildState.childCancel()
		return nil, errors.Wrap(err, "could not restart child")
	}

	return newChildState, nil
}

func (s *Supervisor) waitChild(supState *state) {
	err := supState.child.Wait()
	if err != nil {
		log.WithError(err).Errorf("Could not wait on child")
	}

	// Close the stdout/stderr pipes on the child. Before moving forward.
	supState.closeChildOutputs()

	// This is a gross hack to let childCtx short circuit
	// responding on stateCh with a process state.
	//
	// When an environment update is received, the child is
	// restarted, which would cause this waiter to finish and
	// submit a process state, which would trigger another restart.
	nextState := make(chan *os.ProcessState, 1)
	nextState <- supState.child.ProcessState

	select {
	case <-supState.childCtx.Done():
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
		if err, ok := err.(*exec.ExitError); ok {
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
