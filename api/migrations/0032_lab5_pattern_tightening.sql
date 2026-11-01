-- 0032: systematic audit (user asked "have the other questions been
-- checked too?") found 2 more steps in Lab 5 with the same bug class as
-- the earlier Lab 2 report: the pattern accepted the hint's own
-- placeholder text verbatim (e.g. typing "<file>" or "<ip>" literally),
-- without requiring the learner to substitute a real value.
--
-- Step 2 (PHP filter wrapper): required only "...resource=", so typing the
-- hint's "resource=<file>" literally passed. Now requires a real
-- (non "<"-starting) value after resource=.
UPDATE cards
SET payload = jsonb_set(
  payload,
  '{steps,1,patterns}',
  '["php://filter.*convert\\.base64-encode.*resource=(?!<)\\S+"]'::jsonb,
  false
)
WHERE kind = 'lab' AND front = 'Lab 5 — LFI & Command Injection';

-- Step 4 (&& command injection): required only "&&\s*whoami", so typing
-- the hint's "<ip> && whoami" literally passed with no real IP at all.
-- Now requires at least one digit (a real address) before "&&".
UPDATE cards
SET payload = jsonb_set(
  payload,
  '{steps,3,patterns}',
  '["\\d.*&&\\s*whoami"]'::jsonb,
  false
)
WHERE kind = 'lab' AND front = 'Lab 5 — LFI & Command Injection';
