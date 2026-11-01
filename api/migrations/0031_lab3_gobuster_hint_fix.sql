-- 0031: Lab 3 step 3 (gobuster dir) hint only showed a generic "<wordlist>"
-- placeholder, relying on the learner remembering the exact filename
-- (common.txt) from the earlier ffuf step -- user reported not knowing
-- which file to type. The accepted pattern already required common.txt
-- (unchanged), just the hint was vague. Spell out the filename.
UPDATE cards
SET payload = jsonb_set(
  payload,
  '{steps,2,hint}',
  '"gobuster dir -u <url> -w common.txt  (cùng wordlist đã dùng ở bước ffuf trước đó)"'::jsonb,
  false
)
WHERE kind = 'lab' AND front = 'Lab 3 — Web enumeration cơ bản';
