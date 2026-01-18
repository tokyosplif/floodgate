CREATE TABLE IF NOT EXISTS clicks (
  id UUID,
  campaign_id String,
  source String,
  ip String,
  user_agent String,
  is_bot UInt8,
  processed_at DateTime64(3, 'UTC')
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(processed_at)
ORDER BY (campaign_id, processed_at, id);