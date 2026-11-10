package cliproxy

import (
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/schedule"
	log "github.com/sirupsen/logrus"
)

// initScheduleService creates the scheduled-request service. The directory comes
// from configuration at Start time, so a reload can move it without a restart.
func (s *Service) initScheduleService() *schedule.Service {
	s.cfgMu.Lock()
	defer s.cfgMu.Unlock()
	if s.scheduleService == nil {
		s.scheduleService = schedule.NewService(schedule.NewStore(""))
	}
	return s.scheduleService
}

func (s *Service) currentScheduleService() *schedule.Service {
	if s == nil {
		return nil
	}
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	return s.scheduleService
}

// scheduleDirectory resolves the configured directory. An unusable path disables
// the feature and is reported instead of being silently replaced by a default,
// which would otherwise leave the operator writing into a directory they never
// chose.
func (s *Service) scheduleDirectory(cfg *config.Config) string {
	raw := ""
	if cfg != nil {
		raw = cfg.Schedule.EffectiveDir()
	}
	dir, errResolve := schedule.ResolveDir(raw)
	if errResolve != nil {
		log.Errorf("schedule: %v; scheduled tasks stay disabled", errResolve)
		return ""
	}
	return dir
}

// applyScheduleConfig re-points the scheduler after a configuration change.
func (s *Service) applyScheduleConfig(cfg *config.Config) {
	service := s.currentScheduleService()
	if service == nil {
		return
	}
	service.ApplyDir(s.scheduleDirectory(cfg))
}
