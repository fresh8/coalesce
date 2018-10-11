CREATE TABLE IF NOT EXISTS circleci (
  repo_full_name varchar PRIMARY KEY,
  slack_channel varchar,
  slack_webhook_url varchar
);
