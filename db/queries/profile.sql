-- name: FetchProfileQuery :one
SELECT
  ua.email,
  ua.ghUsername,
  ua.first_name,
  ua.middle_name,
  ua.last_name,
  ua.bounty,
  COUNT(DISTINCT CASE WHEN s.is_merged = true THEN s.id END) as solutions_accepted,
  COUNT(DISTINCT CASE WHEN s.is_merged = false THEN s.id END) as solutions_pending,
  COUNT(DISTINCT ic.id) as active_claims,
  ARRAY_AGG(DISTINCT bd.badge_name) as badges
FROM 
  user_account ua
LEFT JOIN solutions s 
  ON s.ghUsername = ua.ghUsername
LEFT JOIN issue_claims ic 
  ON ic.ghUsername = ua.ghUsername
LEFT JOIN issues i 
  ON i.url = ic.issue_url 
  AND i.resolved = false 
  AND ic.elapsed_on > NOW()
LEFT JOIN badge_dispatch bd
  ON bd.ghUsername = ua.ghUsername
WHERE
  ua.status = true
  AND ua.ghUsername = $1
GROUP BY
  ua.email,
  ua.ghUsername,
  ua.first_name,
  ua.middle_name,
  ua.last_name,
  ua.bounty;