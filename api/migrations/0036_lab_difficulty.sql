-- 0036_lab_difficulty: tag every existing lab/assignment with a difficulty
-- level (co_ban/trung_binh/nang_cao) so learners can see progression,
-- rendered as a badge in the list and runner header.

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('co_ban'::text), true)
WHERE kind = 'lab' AND front = 'Lab 1 — Làm quen lệnh Linux cơ bản';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('co_ban'::text), true)
WHERE kind = 'lab' AND front = 'Lab 2 — Host discovery & quét cổng (nmap)';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('co_ban'::text), true)
WHERE kind = 'lab' AND front = 'Lab 3 — Web enumeration cơ bản';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Lab 4 — SQL Injection (manual + sqlmap)';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Lab 5 — LFI & Command Injection';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Lab 6 — Reverse shell & nâng cấp TTY';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Lab 7 — Linux Privilege Escalation';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('nang_cao'::text), true)
WHERE kind = 'lab' AND front = 'Lab 8 — Active Directory enumeration';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Lab 9 — Password attacks (hydra / hashcat / john)';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('co_ban'::text), true)
WHERE kind = 'lab' AND front = 'Lab 10 — tmux: thao tác window/pane nhanh';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Lab 11 — Windows Privilege Escalation cơ bản';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('nang_cao'::text), true)
WHERE kind = 'lab' AND front = 'Lab 12 — File transfer & pivoting/tunneling';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Lab 13 — Web: XSS, SSRF, XXE cơ bản';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('co_ban'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 1 — External recon đến foothold qua FTP backdoor';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 2 — Web enum đến SQLi dump credentials';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 3 — File upload vulnerability đến reverse shell';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 4 — LFI + log poisoning đến RCE';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 5 — Linux privesc qua sudo misconfiguration';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('nang_cao'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 6 — AD enum đến AS-REP Roasting';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('nang_cao'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 7 — Password spraying đến shell qua WinRM';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('nang_cao'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 8 — BloodHound đến Domain Admin';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('trung_binh'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 9 — IDOR + JWT tampering trên API';

UPDATE cards SET payload = jsonb_set(payload, '{difficulty}', to_jsonb('nang_cao'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 10 — Capstone: full chain từ recon đến root';
