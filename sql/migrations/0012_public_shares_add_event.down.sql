ALTER TABLE public_shares
DROP CONSTRAINT public_shares_resource_type_check;

ALTER TABLE public_shares
ADD CONSTRAINT public_shares_resource_type_check
CHECK (resource_type IN ('memo', 'todo', 'calendar', 'tool'));
