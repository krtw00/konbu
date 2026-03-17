CREATE TABLE public_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('memo', 'todo', 'calendar', 'tool')),
    resource_id UUID NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (resource_type, resource_id)
);

CREATE INDEX idx_public_shares_token ON public_shares(token);
