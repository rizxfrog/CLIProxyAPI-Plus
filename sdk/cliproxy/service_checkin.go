package cliproxy

import "github.com/router-for-me/CLIProxyAPI/v8/internal/checkin"

func (s *Service) initCheckinService() *checkin.Service {
	s.cfgMu.Lock()
	defer s.cfgMu.Unlock()
	if s.checkinService == nil {
		var source checkin.AccountSource
		if s.coreManager != nil {
			source = s.coreManager
		}
		s.checkinService = checkin.NewService(source, checkin.BuiltinAdapters()...)
	}
	return s.checkinService
}

func (s *Service) currentCheckinService() *checkin.Service {
	if s == nil {
		return nil
	}
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	return s.checkinService
}

func (s *Service) notifyCheckinAccounts() {
	if service := s.currentCheckinService(); service != nil {
		service.Notify()
	}
}
