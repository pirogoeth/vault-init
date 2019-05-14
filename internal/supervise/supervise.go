package supervise

import (
	"context"

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
func (s *Supervisor) Start(ctx context.Context, updateCh chan map[string]string) error {
	log.Info("Starting supervisor")
	return nil
}
