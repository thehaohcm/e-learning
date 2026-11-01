-- 0027: two fixes to Lab 2 ("Host discovery & quét cổng (nmap)"), reported by
-- the user while testing:
--
-- (a) Step 4's hint showed "nmap ... -oA <tên file gốc>" using "..." as a
--     placeholder for "the flags from the previous step" -- but the check
--     pattern (nmap .* -oA scan1) is a loose regex where ".*" matches ANY
--     characters, including the literal three dots. So typing the hint text
--     verbatim ("nmap ... -oA scan1") was wrongly accepted as correct, with
--     no real nmap flags or target IP at all. Fix: require the real target
--     IP to be present too (in any flag order), and reword the hint so it
--     reads as an instruction, not literal syntax to copy.
--
-- (b) Add a UDP port scan step (-sU), requested by the user, covering a
--     technique the lab didn't drill yet.
UPDATE cards
SET payload = jsonb_set(
  jsonb_set(
    payload,
    '{steps,3,patterns}',
    '["(?=.*nmap)(?=.*-oA\\s+scan1)(?=.*10\\.10\\.10\\.5).*"]'::jsonb,
    false
  ),
  '{steps,3,hint}',
  '"Gõ lại lệnh quét ở bước trước (nmap -sC -sV -p 22,80 10.10.10.5), thêm -oA scan1 vào cuối -- dấu \"...\" trong gợi ý chỉ nghĩa là \"phần lệnh đã gõ trước đó\", không phải gõ 3 dấu chấm theo nghĩa đen."'::jsonb,
  false
)
WHERE kind = 'lab' AND front = 'Lab 2 — Host discovery & quét cổng (nmap)';

UPDATE cards
SET payload = jsonb_set(
  payload,
  '{steps}',
  (payload->'steps') || '[{"prompt": "Quét nhanh top 20 cổng UDP phổ biến của 10.10.10.5 (UDP scan rất chậm nên luôn giới hạn phạm vi thay vì quét toàn bộ 65535 cổng như TCP).", "hint": "nmap -sU --top-ports 20 <ip>", "patterns": ["nmap\\s+-sU\\s+--top-ports\\s*[= ]?\\s*20\\s+10\\.10\\.10\\.5", "nmap\\s+-sU\\s+-F\\s+10\\.10\\.10\\.5"], "explanation": "-sU chuyển sang quét UDP thay vì TCP mặc định. UDP không có bắt tay 3 bước nên phản hồi chậm/không chắc chắn (open|filtered) -- luôn giới hạn phạm vi bằng --top-ports hoặc -F (fast, top 100) thay vì -p- như TCP, nếu không có thể mất hàng giờ. Mẹo nhớ: U = UDP, phân biệt với -sT/-sV dùng cho TCP."}]'::jsonb,
  true
)
WHERE kind = 'lab' AND front = 'Lab 2 — Host discovery & quét cổng (nmap)';
