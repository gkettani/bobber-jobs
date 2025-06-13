CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,  -- The careers page URL
    active BOOLEAN DEFAULT TRUE,  -- Determines if we scrape this source
    metadata JSONB DEFAULT '{}',  -- Store additional data in a flexible format
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    external_id TEXT UNIQUE NOT NULL,
    company_name TEXT NOT NULL,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    location TEXT NOT NULL,
    description TEXT NOT NULL,
    hash TEXT,
    first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expired_at TIMESTAMP NULL,
    search_vector TSVECTOR GENERATED ALWAYS AS (
        to_tsvector('english', title || ' ' || description)
    ) STORED
);

-- Create GIN index for efficient text search
CREATE INDEX jobs_search_idx ON jobs USING GIN (search_vector);

-- Function to update search vector
CREATE FUNCTION jobs_search_update() RETURNS TRIGGER AS $$
BEGIN
  NEW.search_vector := 
      to_tsvector('english', coalesce(NEW.title, '') || ' ' || 
                              coalesce(NEW.description, '') || ' ' || 
                              coalesce(NEW.location, ''));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update search vector on insert/update
CREATE TRIGGER update_jobs_search
BEFORE INSERT OR UPDATE ON jobs
FOR EACH ROW EXECUTE FUNCTION jobs_search_update();

-- Fast lookup on external_id
CREATE INDEX jobs_external_id_idx ON jobs (external_id);

-- Quick filtering on scraped_at & expired_at
CREATE INDEX jobs_scraped_at_idx ON jobs (scraped_at);
CREATE INDEX jobs_expired_at_idx ON jobs (expired_at);


-- Example
SELECT id, title, job_url, location
FROM jobs 
WHERE search_vector @@ to_tsquery('(software | engineer) & paris')
-- ORDER BY scraped_at DESC
LIMIT 20;
