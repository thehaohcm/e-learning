-- 0044_add_cert_tags_to_labs_and_assignments.sql
-- Assign explicit cert tags ("THM PT1", "THM WEB1", "HTB CPTS") to all Labs and Assignments
-- so learners immediately see which certification track each lab directly targets.

-- ============================================================================
-- LABS (1 - 19)
-- ============================================================================

-- Lab 1: Linux Basics -> PT1 & CPTS baseline
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 1 — Làm quen lệnh Linux cơ bản';

-- Lab 2: Nmap scanning -> PT1 & CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 2 — Host discovery & quét cổng (nmap)';

-- Lab 3: Web enum -> WEB1 & PT1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1", "THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 3 — Web enumeration cơ bản';

-- Lab 4: SQLi -> WEB1 & PT1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1", "THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 4 — SQL Injection (manual + sqlmap)';

-- Lab 5: LFI & Command Injection -> WEB1 & PT1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1", "THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 5 — LFI & Command Injection';

-- Lab 6: Reverse shell & TTY -> PT1 & CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 6 — Reverse shell & nâng cấp TTY';

-- Lab 7: Linux PrivEsc -> PT1 & CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 7 — Linux Privilege Escalation';

-- Lab 8: AD Enumeration -> HTB CPTS (và PT1 nâng cao)
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS", "THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 8 — Active Directory enumeration';

-- Lab 9: Password attacks (Hydra/Hashcat/John) -> PT1 & CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 9 — Password attacks (hydra / hashcat / john)';

-- Lab 10: Tmux -> PT1 & CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 10 — tmux: thao tác window/pane nhanh';

-- Lab 11: Windows PrivEsc -> PT1 & CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 11 — Windows Privilege Escalation cơ bản';

-- Lab 12: File transfer & Chisel tunneling -> PT1 & CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS", "THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 12 — File transfer & pivoting/tunneling';

-- Lab 13: XSS, SSRF, XXE -> THM WEB1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 13 — Web: XSS, SSRF, XXE cơ bản';

-- Lab 14: NFS, SNMP, DNS Zone Transfer -> PT1 & CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 14 — Network enum nâng cao: NFS, SNMP, DNS Zone Transfer';

-- Lab 15: AD Creds (Responder, GPP, Mimikatz) -> HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 15 — AD Credential Attacks: LLMNR/Responder, GPP, Mimikatz';

-- Lab 16: Web advanced (CORS, Smuggling, Deserialization) -> THM WEB1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 16 — Web nâng cao: CORS, HTTP Smuggling, Insecure Deserialization';

-- Lab 17: SSTI, NoSQLi & searchsploit -> THM WEB1 & THM PT1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1", "THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 17 — Web nâng cao 2: SSTI, NoSQLi & Tra cứu Exploit (searchsploit)';

-- Lab 18: Ligolo-ng pivoting & SMB -> HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 18 — Pivoting hiện đại với Ligolo-ng & Chuyển file SMB';

-- Lab 19: ADCS ESC1, Kerberoasting & PTH -> HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Lab 19 — AD nâng cao (CPTS): Kerberoasting, ADCS (Certipy) & Pass-the-Hash';

-- ============================================================================
-- ASSIGNMENTS (1 - 14)
-- ============================================================================

-- Assignment 1: Recon to FTP Backdoor -> THM PT1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 1 — External recon đến foothold qua FTP backdoor';

-- Assignment 2: Web enum to SQLi -> THM WEB1 & THM PT1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1", "THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 2 — Web enum đến SQLi dump credentials';

-- Assignment 3: File upload to reverse shell -> THM WEB1 & THM PT1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1", "THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 3 — File upload vulnerability đến reverse shell';

-- Assignment 4: LFI + log poisoning -> THM WEB1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 4 — LFI + log poisoning đến RCE';

-- Assignment 5: Linux sudo privesc -> THM PT1 & HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 5 — Linux privesc qua sudo misconfiguration';

-- Assignment 6: AD enum to AS-REP Roasting -> HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 6 — AD enum đến AS-REP Roasting';

-- Assignment 7: Password spraying to WinRM -> HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 7 — Password spraying đến shell qua WinRM';

-- Assignment 8: BloodHound to Domain Admin -> HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 8 — BloodHound đến Domain Admin';

-- Assignment 9: IDOR + JWT tampering -> THM WEB1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 9 — IDOR + JWT tampering trên API';

-- Assignment 10: Capstone recon to root -> THM PT1 & HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM PT1", "HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 10 — Capstone: full chain từ recon đến root';

-- Assignment 11: AD Credential Chain (Responder to DA) -> HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 11 — AD Credential Chain: Responder đến Domain Admin';

-- Assignment 12: Web Advanced Chain (CORS to Deserialization) -> THM WEB1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 12 — Web Advanced Chain: CORS đến RCE qua Deserialization';

-- Assignment 13: Web SSTI + CVE PrivEsc -> THM WEB1 & THM PT1
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["THM WEB1", "THM PT1"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 13 — Web SSTI & CVE Exploit (searchsploit) đến Root';

-- Assignment 14: CPTS Enterprise Chain (Kerberoasting -> Ligolo -> ADCS) -> HTB CPTS
UPDATE cards SET payload = jsonb_set(payload, '{cert_tags}', '["HTB CPTS"]'::jsonb, true)
WHERE kind = 'lab' AND front = 'Assignment 14 — CPTS Enterprise Chain: Kerberoasting → Ligolo-ng → ADCS ESC1 → Domain Admin';

