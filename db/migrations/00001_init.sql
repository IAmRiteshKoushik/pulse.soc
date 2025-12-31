-- +goose Up

-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS "public";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS maintainers(
  id SERIAL NOT NULL,
  ghUsername TEXT NOT NULL UNIQUE,
  full_name TEXT NOT NULL,

  CONSTRAINT "maintainers_pkey" PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_account(
  id SERIAL NOT NULL,
  first_name TEXT NOT NULL,
  middle_name TEXT,
  last_name TEXT NOT NULL,
  email TEXT NOT NULL,
  ghUsername TEXT UNIQUE,
  password TEXT NOT NULL,
  status BOOLEAN DEFAULT true,
  bounty INT NOT NULL DEFAULT 0,
  refresh_token TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  CONSTRAINT "user_account_pkey" PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_onboarding(
  id SERIAL NOT NULL,
  first_name TEXT NOT NULL,
  middle_name TEXT,
  last_name TEXT NOT NULL,
  email TEXT NOT NULL,
  ghUsername TEXT UNIQUE,
  password TEXT NOT NULL,
  otp TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  expiry_at TIMESTAMP NOT NULL,

  CONSTRAINT "user_onboarding_pkey" PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS repository(
  id UUID DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  url TEXT NOT NULL UNIQUE,
  maintainers TEXT[],
  tags TEXT[],
  is_internal BOOLEAN DEFAULT false,
  installation_id BIGINT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  CONSTRAINT "repository_pkey" PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS issues(
  id UUID DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  repoUrl TEXT NOT NULL,
  url TEXT NOT NULL UNIQUE,
  tags TEXT[],
  difficulty TEXT DEFAULT 'Easy',
  resolved BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  CONSTRAINT "issues_pkey" PRIMARY KEY (id),

  CONSTRAINT "issues_repourl_fkey"
    FOREIGN KEY (repoUrl)
      REFERENCES repository(url)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS issue_claims(
  id UUID DEFAULT gen_random_uuid(),
  ghUsername TEXT NOT NULL,
  issue_url TEXT NOT NULL,
  claimed_on TIMESTAMP NOT NULL,
  elapsed_on TIMESTAMP NOT NULL,

  CONSTRAINT "issue_claims_pkey" PRIMARY KEY (id),

  CONSTRAINT "issue_claims_ghUsername_fkey"
    FOREIGN KEY (ghUsername)
      REFERENCES user_account(ghUsername)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS solutions(
  id SERIAL NOT NULL,
  url TEXT NOT NULL,
  repo_url TEXT NOT NULL,
  ghUsername TEXT NOT NULL,
  is_merged BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  CONSTRAINT "solutions_pkey" PRIMARY KEY (id),
  CONSTRAINT "solutions_repo_url_fkey" 
    FOREIGN KEY (repo_url)
      REFERENCES repository(url)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,
  CONSTRAINT "solutions_ghUsername_fkey"
    FOREIGN KEY (ghUsername)
      REFERENCES user_account(ghUsername)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bounty_log(
  id SERIAL NOT NULL,
  ghUsername TEXT NOT NULL,
  dispatched_by TEXT NOT NULL,
  proof_url TEXT NOT NULL,
  amount INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),

  CONSTRAINT "bounty_log_pkey" PRIMARY KEY (id),
  CONSTRAINT "bounty_log_dispatch_fkey"
    FOREIGN KEY (dispatched_by)
      REFERENCES maintainers(ghUsername)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,
  CONSTRAINT "bounty_log_ghUsername_fkey"
    FOREIGN KEY (ghUsername)
      REFERENCES user_account(ghUsername)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS badge_info(
  id SERIAL NOT NULL,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL,

  CONSTRAINT "badge_info_pkey" PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS badge_dispatch(
  id SERIAL NOT NULL,
  ghUsername TEXT NOT NULL,
  badge_name TEXT NOT NULL,
  awarded_at TIMESTAMP DEFAULT NOW(),

  CONSTRAINT "badge_dispatch_pkey" PRIMARY KEY (id),
  CONSTRAINT "badge_dispatch_ghUsername_fkey"
    FOREIGN KEY (ghUsername)
      REFERENCES user_account(ghUsername)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,
  CONSTRAINT "badge_dispatch_badge_name_fkey" 
    FOREIGN KEY (badge_name)
      REFERENCES badge_info(name)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA IF EXISTS public CASCADE;
-- +goose StatementEnd
