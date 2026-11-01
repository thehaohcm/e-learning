-- 0029: Lab 7 step 3 ("find . -exec /bin/sh -p \;" GTFOBins SUID exploit)
-- required an un-escaped ";" right after "-p", but the step's own hint shows
-- the semicolon escaped with a backslash (\;), which is the syntax find
-- actually needs. So the exact command from the hint was being rejected.
-- Make the backslash before ";" optional so both "\;" and ";" are accepted.
UPDATE cards
SET payload = jsonb_set(
  payload,
  '{steps,2,patterns}',
  '["find\\s+\\S+\\s+-exec\\s+/bin/sh\\s+-p\\s*\\\\?;", "find\\s+.*-exec\\s+/bin/bash\\s+-p\\s*\\\\?;"]'::jsonb,
  false
)
WHERE kind = 'lab' AND front = 'Lab 7 — Linux Privilege Escalation';
