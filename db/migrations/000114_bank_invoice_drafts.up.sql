CREATE TABLE public.bank_invoice_drafts (
    payment_action_id uuid PRIMARY KEY REFERENCES public.payment_actions(payment_action_id) ON DELETE CASCADE,
    revision integer NOT NULL CHECK (revision > 0),
    kind text NOT NULL CHECK (kind IN ('bank_invoice', 'payment_link')),
    payer_name text NOT NULL CHECK (length(payer_name) BETWEEN 1 AND 240),
    tax_id text NOT NULL DEFAULT '' CHECK (length(tax_id) <= 32),
    purpose text NOT NULL CHECK (length(purpose) BETWEEN 1 AND 1000),
    due_date date,
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency text NOT NULL CHECK (currency = 'RUB'),
    status text NOT NULL DEFAULT 'draft' CHECK (status = 'draft'),
    request_key uuid NOT NULL,
    request_hash text NOT NULL CHECK (length(request_hash) = 64),
    updated_at timestamptz NOT NULL DEFAULT now()
);
COMMENT ON TABLE public.bank_invoice_drafts IS 'Bank-neutral preparation only; tenant and project derive from payment_actions. No provider side effects.';
