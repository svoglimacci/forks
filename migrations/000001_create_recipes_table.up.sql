CREATE TABLE IF NOT EXISTS recipes (
id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  title text NOT NULL,
  description text,
  prep_time integer NOT NULL,
  cooking_time integer NOT NULL,
  categories text[] NOT NULL,
  servings integer NOT NULL,
  version integer NOT NULL DEFAULT 1,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

