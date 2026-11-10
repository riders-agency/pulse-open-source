package person

import (
	personruntime "riders.agency/pulse/api/internal/modules/person/runtime"
	"riders.agency/pulse/api/internal/runtime/registry"
)

type TemplateRenderer = personruntime.TemplateRenderer

func RegisterRuntimeHandlers(reg *registry.Registry, renderer TemplateRenderer) error {
	return personruntime.RegisterHandlers(reg, renderer)
}
