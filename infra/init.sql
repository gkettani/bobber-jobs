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
        setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(location, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(description, '')), 'C')
    ) STORED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create GIN index for efficient text search
CREATE INDEX CONCURRENTLY jobs_search_idx ON jobs USING GIN (search_vector);

-- Partial index for active jobs (most common filter) - CRITICAL for performance
CREATE INDEX CONCURRENTLY jobs_active_search_idx ON jobs USING GIN (search_vector) WHERE expired_at IS NULL;

-- Enable trigram extension for fuzzy matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Composite indexes for common query patterns
CREATE INDEX CONCURRENTLY jobs_company_search_idx ON jobs USING GIN (company_name gin_trgm_ops, search_vector) WHERE expired_at IS NULL;
CREATE INDEX CONCURRENTLY jobs_location_search_idx ON jobs USING GIN (location gin_trgm_ops, search_vector) WHERE expired_at IS NULL;

-- B-tree indexes for ordering and filtering
CREATE INDEX CONCURRENTLY jobs_last_seen_idx ON jobs (last_seen_at DESC) WHERE expired_at IS NULL;
CREATE INDEX CONCURRENTLY jobs_first_seen_idx ON jobs (first_seen_at DESC) WHERE expired_at IS NULL;
CREATE INDEX CONCURRENTLY jobs_company_last_seen_idx ON jobs (company_name, last_seen_at DESC) WHERE expired_at IS NULL;

-- Unique constraint index (already exists but optimized)
CREATE UNIQUE INDEX CONCURRENTLY jobs_external_id_unique_idx ON jobs (external_id);

-- Statistics optimization for text search columns
ALTER TABLE jobs ALTER COLUMN title SET STATISTICS 200;
ALTER TABLE jobs ALTER COLUMN description SET STATISTICS 150;
ALTER TABLE jobs ALTER COLUMN location SET STATISTICS 100;
ALTER TABLE jobs ALTER COLUMN company_name SET STATISTICS 100;

-- Analyze the table to update statistics
ANALYZE jobs;

-- Create a function for optimized search queries with better memory usage
CREATE OR REPLACE FUNCTION search_jobs_optimized(
    search_query TEXT,
    page_limit INTEGER DEFAULT 20,
    page_offset INTEGER DEFAULT 0
) RETURNS TABLE (
    id INTEGER,
    company_name TEXT,
    title TEXT,
    location TEXT,
    first_seen_at TIMESTAMP,
    rank REAL,
    total_count BIGINT
) AS $$
DECLARE
    ts_query tsquery;
    total_rows BIGINT;
BEGIN
    -- Pre-compute the tsquery to avoid multiple conversions
    ts_query := websearch_to_tsquery('english', search_query);
    
    -- Get total count first (more efficient for pagination)
    SELECT COUNT(*) INTO total_rows
    FROM jobs j
    WHERE j.search_vector @@ ts_query
      AND j.expired_at IS NULL;
    
    -- Return paginated results with pre-computed total
    RETURN QUERY
    SELECT 
        j.id,
        j.company_name,
        j.title,
        j.location,
        j.first_seen_at,
        ts_rank_cd(j.search_vector, ts_query) as rank,
        total_rows as total_count
    FROM jobs j
    WHERE j.search_vector @@ ts_query
      AND j.expired_at IS NULL  -- Only active jobs
    ORDER BY 
        ts_rank_cd(j.search_vector, ts_query) DESC,
        j.last_seen_at DESC
    LIMIT page_limit OFFSET page_offset;
END;
$$ LANGUAGE plpgsql;

-- Create a lightweight search function for simple queries (better for low memory)
CREATE OR REPLACE FUNCTION search_jobs_simple(
    search_query TEXT,
    page_limit INTEGER DEFAULT 20,
    page_offset INTEGER DEFAULT 0
) RETURNS TABLE (
    id INTEGER,
    company_name TEXT,
    title TEXT,
    location TEXT,
    first_seen_at TIMESTAMP,
    rank REAL
) AS $$
DECLARE
    ts_query tsquery;
BEGIN
    ts_query := websearch_to_tsquery('english', search_query);
    
    RETURN QUERY
    SELECT 
        j.id,
        j.company_name,
        j.title,
        j.location,
        j.first_seen_at,
        ts_rank_cd(j.search_vector, ts_query) as rank
    FROM jobs j
    WHERE j.search_vector @@ ts_query
      AND j.expired_at IS NULL
    ORDER BY 
        ts_rank_cd(j.search_vector, ts_query) DESC,
        j.last_seen_at DESC
    LIMIT page_limit OFFSET page_offset;
END;
$$ LANGUAGE plpgsql;

-- Performance monitoring views
CREATE OR REPLACE VIEW search_performance_stats AS
SELECT 
    schemaname,
    relname,
    indexrelname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    CASE 
        WHEN idx_scan > 0 THEN round((idx_tup_fetch::numeric / idx_scan), 2)
        ELSE 0
    END as avg_tuples_per_scan
FROM pg_stat_user_indexes 
WHERE relname = 'jobs' 
ORDER BY idx_scan DESC;

-- Memory usage monitoring view
CREATE OR REPLACE VIEW memory_usage_stats AS
SELECT 
    name,
    setting,
    unit,
    context,
    short_desc
FROM pg_settings 
WHERE name IN ('shared_buffers', 'work_mem', 'maintenance_work_mem', 'effective_cache_size')
ORDER BY name;

-- Example optimized queries for testing
-- Basic search with ranking (optimized for partial index)
SELECT id, title, company_name, location,
       ts_rank_cd(search_vector, websearch_to_tsquery('english', 'software engineer')) as rank
FROM jobs 
WHERE search_vector @@ websearch_to_tsquery('english', 'software engineer')
  AND expired_at IS NULL
ORDER BY rank DESC, last_seen_at DESC
LIMIT 20;

-- Search with company filter (uses partial composite index)
SELECT id, title, company_name, location
FROM jobs 
WHERE search_vector @@ websearch_to_tsquery('english', 'python developer')
  AND company_name ILIKE '%google%'
  AND expired_at IS NULL
ORDER BY ts_rank_cd(search_vector, websearch_to_tsquery('english', 'python developer')) DESC
LIMIT 20;

-- Search with location filter (uses partial composite index)
SELECT id, title, company_name, location
FROM jobs 
WHERE search_vector @@ websearch_to_tsquery('english', 'remote frontend')
  AND location ILIKE '%remote%'
  AND expired_at IS NULL
ORDER BY ts_rank_cd(search_vector, websearch_to_tsquery('english', 'remote frontend')) DESC
LIMIT 20;

-- Performance analysis query for monitoring
SELECT 
    'Index Usage' as metric,
    indexrelname as name,
    idx_scan as scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes 
WHERE relname = 'jobs' AND idx_scan > 0
ORDER BY idx_scan DESC; 
