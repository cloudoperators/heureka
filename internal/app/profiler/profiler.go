// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package profiler

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
)

type Profiler struct {
	filepath string
	file     *os.File
}

func NewProfiler(filepath string) *Profiler {
	p := Profiler{filepath: filepath}
	return &p
}

func (p *Profiler) Start() {
	if p.filepath == "" {
		return
	}

	log.Printf("[Profiler]: Starting profiler. Results will be collected in '%s'. Stop app/container and copy results to browse perf data.", p.filepath)
	var err error
	p.file, err = os.Create(p.filepath)
	if err != nil {
		p.file = nil
		log.Print(fmt.Errorf("[Profiler]: Could not create profiler file '%s': %w", p.filepath, err))
	} else if err := pprof.StartCPUProfile(p.file); err != nil {
		p.cleanup()
		log.Print(fmt.Errorf("[Profiler]: Could not start profiler: %w", err))
	}
}

func (p *Profiler) Stop() {
	if p.file != nil {
		pprof.StopCPUProfile()
		log.Print("[Profiler]: Profiler stopped")
		p.cleanup()
	}
}

func (p *Profiler) cleanup() {
	p.file.Close()
	p.file = nil
}
