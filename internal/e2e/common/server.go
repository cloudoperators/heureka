package e2e_common

import (
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/gomega"
)

func NewRunningServer(cfg util.Config) *server.Server {
	s := server.NewServer(cfg)
	s.NonBlockingStart()
	err := s.GetApp().WaitPostMigrations()
	Expect(err).Should(BeNil())
	return s
}

func ServerTeardown(s *server.Server) {
	s.BlockingStop()
}
