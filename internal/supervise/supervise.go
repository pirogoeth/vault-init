package supervise

import (
	reaper "github.com/ramr/go-reaper"
	log "github.com/sirupsen/logrus"
)

// NewSupervisor creates a supervisor instance
func NewSupervisor(config *Config) *Supervisor {
	if !config.DisableReaper {
		log.Info("Starting process reaper")
		go reaper.Reap()
	}

	return &Supervisor{config}
}

// Start spawns the specified child process and runs a goroutine with
// the subprocess reaper
func (s *Supervisor) Start() error {
	log.WithField("command", s.config.CommandString()).Info("Starting child subprocess")
	return nil
}
