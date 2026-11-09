package executor

import runtimekernel "riders.agency/pulse/api/internal/runtime/kernel"

func (e *Engine) registerExecutorRuntimeTriggers() {
	runtimekernel.RegisterDatabaseTriggers(e.Registry, e.P, e)
}
