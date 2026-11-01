-- 0025: Lab 1 step 4 ("grep password in *.txt") required both -r and -i
-- flags (grep -ri / grep -i -r), but the step's own wording only asks to
-- search files in the CURRENT directory, not recursively -- so a correct
-- answer like `grep -i 'password' *.txt` (no -r) was being rejected.
-- Widen the accepted patterns to also allow -i alone (quoted or
-- unquoted 'password'), while still accepting the original recursive
-- -ri / -i -r / -ir forms.
UPDATE cards
SET payload = jsonb_set(
  payload,
  '{steps,3,patterns}',
  '["grep\\s+-ri\\s+password\\s+.*\\.txt", "grep\\s+-i\\s+-r\\s+password\\s+.*\\.txt", "grep\\s+-ir\\s+password\\s+.*\\.txt", "grep\\s+-i\\s+.?password.?\\s+\\*\\.txt"]'::jsonb,
  false
)
WHERE kind = 'lab' AND front = 'Lab 1 — Làm quen lệnh Linux cơ bản';
