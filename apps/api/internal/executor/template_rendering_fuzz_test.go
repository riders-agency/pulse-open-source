package executor

import "testing"

func FuzzRenderTemplate(f *testing.F) {
	for _, seed := range []struct {
		template string
		input    string
	}{
		{"Hello {{ input.name }}", `{"name":"Ada"}`},
		{"{{ json input }}", `{"webhook":{"amount":1200,"paid":false}}`},
		{"{{ steps.previous.status }} {{ system.run_id }}", `{"payload":{"id":"evt_1"}}`},
		{"{{ kss-bitrix-form-to-max.phone }}", `{"_pulse":{"webhook_slug":"kss-bitrix-form-to-max"},"webhook":{"phone":"+79857254855"}}`},
		{"{{", `not-json`},
	} {
		f.Add(seed.template, seed.input)
	}

	f.Fuzz(func(t *testing.T, template, input string) {
		if len(template) > 4096 || len(input) > 1<<14 {
			t.Skip("template rendering fuzz input is intentionally capped")
		}
		steps := map[string]any{
			"previous": map[string]any{
				"status": "ok",
				"count":  2,
			},
		}
		system := templateSystem{
			RunID:  "00000000-0000-0000-0000-000000000001",
			RunURL: "https://pulse.riders.agency/runs/00000000-0000-0000-0000-000000000001",
		}
		_ = renderTemplate(template, []byte(input), steps, system)
	})
}
