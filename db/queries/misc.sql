-- name: FetchParticipantListQuery :many
SELECT
  CASE
    WHEN middle_name IS NULL THEN first_name || ' ' || last_name
    ELSE first_name || ' ' || middle_name || ' ' || last_name
  END as full_name,
  ghUsername AS github_username,
  0 as bounty,
  0 as solutions
FROM
  user_account
WHERE
  status = true;

-- name: FetchUsersWithNoContributionsQuery :many
SELECT
  ua.ghUsername AS github_username,
  ua.bounty,
  COUNT(s.id) as solutions
FROM
  user_account ua
  LEFT JOIN solutions s ON ua.ghUsername = s.ghUsername
  AND s.is_merged = true
WHERE
  ua.status = TRUE
  AND ua.ghUsername IS NOT NULL
GROUP BY
  ua.ghUsername,
  ua.bounty
HAVING
  ua.bounty = 0
  AND COUNT(s.id) = 0;
