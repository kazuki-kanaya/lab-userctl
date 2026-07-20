package register

import "github.com/kazuki-kanaya/lab-userctl/internal/terminal"

type Service struct {
	prompt *terminal.Prompter
}

func New(prompt *terminal.Prompter) *Service {
	return &Service{
		prompt: prompt,
	}
}

func (s *Service) Run() error {
	user, err := s.resolveUser()
	if err != nil {
		return err
	}

	if err := s.configureSudo(user); err != nil {
		return err
	}

	return s.registerSSHKey(user)
}
