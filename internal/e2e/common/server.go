// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_common

import (
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
)

func NewRunningServer(cfg util.Config) *server.Server {
	s := server.NewServer(cfg)
	s.NonBlockingStart()
	s.GetApp().WaitPostMigrations()

	return s
}

func ServerTeardown(s *server.Server) {
	s.BlockingStop()
}
