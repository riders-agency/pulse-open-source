package executor

func pulseSteps(pulse map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{"steps", "steps_overrides"} {
		raw, _ := pulse[key].(map[string]any)
		for stepKey, stepValue := range raw {
			out[stepKey] = stepValue
		}
	}
	return out
}
