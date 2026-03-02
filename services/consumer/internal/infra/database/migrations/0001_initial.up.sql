CREATE TABLE processed_items (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  processed_at timestamptz not null,
  processed_data varchar(255) not null
);

CREATE INDEX ON processed_items(processed_at);
