package executor

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	httpclientruntime "riders.agency/pulse/api/internal/modules/httpclient/runtime"
	personruntime "riders.agency/pulse/api/internal/modules/person/runtime"
	"riders.agency/pulse/api/internal/platform/httptransport"
	runtimeactions "riders.agency/pulse/api/internal/runtime/actions"
	runtimekernel "riders.agency/pulse/api/internal/runtime/kernel"
	runtimeregistry "riders.agency/pulse/api/internal/runtime/registry"
)

// Engine adapts workflow definitions and module-owned runtime handlers to the runtime kernel.
type Engine struct {
	P          *pgxpool.Pool
	MasterHEX  string
	HTTPClient *http.Client
	Registry   *runtimeregistry.Registry
	ch         chan struct{}
	notify     func()
}

func New(p *pgxpool.Pool) *Engine {
	httpClient := httptransport.NewDefaultClient(0)
	reg := runtimeregistry.New()
	for _, handler := range runtimeactions.CoreHandlers(coreTemplateRenderer) {
		reg.MustRegisterAction(handler)
	}
	if err := personruntime.RegisterHandlers(reg, coreTemplateRenderer); err != nil {
		panic(err)
	}
	reg.MustRegisterAction(httpclientruntime.Action{
		HTTPClient: httpClient,
		Renderer:   coreTemplateRenderer,
	})
	engine := &Engine{
		P:          p,
		MasterHEX:  os.Getenv("RIDERS_PULSE_MASTER_KEY"),
		HTTPClient: httpClient,
		Registry:   reg,
		ch:         make(chan struct{}, 1),
	}
	engine.registerExecutorRuntimeActions()
	engine.registerExecutorRuntimeTriggers()
	return engine
}

func coreTemplateRenderer(template string, input json.RawMessage, steps map[string]any, workflowRunID string) string {
	return renderTemplate(template, input, steps, templateSystemForRun(workflowRunID))
}

func (e *Engine) RuntimeHTTPClient() *http.Client {
	if e == nil {
		return nil
	}
	return e.HTTPClient
}

func (e *Engine) Notify() {
	if e == nil {
		return
	}
	if e.notify != nil {
		e.notify()
		return
	}
	select {
	case e.ch <- struct{}{}:
	default:
	}
}

func (e *Engine) Run(ctx context.Context) {
	k := runtimekernel.New(e.P, e, runtimekernel.WithTriggerRegistry(e.Registry))
	e.SetNotifier(k.Notify)
	k.Run(ctx)
}
