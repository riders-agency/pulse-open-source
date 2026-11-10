package kernel

const upsertActionAttemptSQL = `
INSERT INTO action_attempts (
  action_run_id, attempt_no, status, input_json, output_json,
  provider_response_json, error, error_type, started_at, finished_at, duration_ms)
VALUES ($1,$2,$3,$4::jsonb,$5::jsonb,$6::jsonb,$7,$8, now(), now(), 1)
ON CONFLICT (action_run_id, attempt_no) DO UPDATE SET
  status = EXCLUDED.status,
  input_json = EXCLUDED.input_json,
  output_json = EXCLUDED.output_json,
  provider_response_json = EXCLUDED.provider_response_json,
  finished_at = now(),
  duration_ms = GREATEST(
    1,
    ROUND(EXTRACT(EPOCH FROM (NOW() - action_attempts.started_at)) * 1000)::integer
  ),
  error = EXCLUDED.error,
  error_type = EXCLUDED.error_type`
