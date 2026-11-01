-- 0028: Lab 10 (tmux shortcuts) hints said "Prefix rồi phím c" assuming the
-- learner already knows what literal string to type for "prefix" -- user
-- reported not knowing what to actually type. Reword every hint to spell out
-- the exact literal answer (e.g. "Gõ: C-b c"), matching the step's own
-- accepted patterns (C-b <key> / ctrl-b <key> / ctrl+b <key>). Patterns,
-- prompts and explanations are unchanged.
UPDATE cards
SET payload = jsonb_set(
  payload,
  '{steps}',
  '[
    {"prompt": "Tạo một window mới trong tmux.", "hint": "Gõ: C-b c  (hoặc: ctrl+b c)", "patterns": ["^c-b\\s*c$", "^ctrl-b\\s*c$", "^ctrl\\+b\\s*c$", "^prefix\\s*c$"], "explanation": "C-b c tạo window mới (tab mới) trong cùng session. Mẹo nhớ: c = create (tạo mới)."},
    {"prompt": "Chuyển sang window kế tiếp (bên phải).", "hint": "Gõ: C-b n  (hoặc: ctrl+b n)", "patterns": ["^c-b\\s*n$", "^ctrl-b\\s*n$", "^ctrl\\+b\\s*n$"], "explanation": "C-b n chuyển tới window tiếp theo theo thứ tự số. Mẹo nhớ: n = next (tiến tới)."},
    {"prompt": "Chuyển sang window trước đó (bên trái).", "hint": "Gõ: C-b p  (hoặc: ctrl+b p)", "patterns": ["^c-b\\s*p$", "^ctrl-b\\s*p$", "^ctrl\\+b\\s*p$"], "explanation": "C-b p chuyển tới window trước đó theo thứ tự số. Mẹo nhớ: p = previous (lùi lại) — n/p luôn đi cặp next/previous như nhiều app khác (vim, tmux, browser)."},
    {"prompt": "Đóng (kill) pane/window hiện tại.", "hint": "Gõ: C-b x  (hoặc: ctrl+b x, hoặc gõ lệnh: exit)", "patterns": ["^c-b\\s*x$", "^ctrl-b\\s*x$", "^ctrl\\+b\\s*x$", "^exit$"], "explanation": "C-b x hỏi xác nhận rồi kill pane hiện tại; gõ exit trong shell của pane cũng đóng pane đó (không cần xác nhận). Mẹo nhớ: x = kill/đóng (giống dấu x để đóng cửa sổ trong nhiều UI)."},
    {"prompt": "Chia pane hiện tại theo chiều NGANG (pane mới nằm bên dưới).", "hint": "Gõ: C-b \"  (hoặc: ctrl+b \", tức C-b rồi ký tự dấu ngoặc kép)", "patterns": ["^c-b\\s*\"$", "^ctrl-b\\s*\"$", "^ctrl\\+b\\s*\"$"], "explanation": "C-b \" split-window ngang — pane mới xuất hiện phía dưới pane hiện tại. Mẹo nhớ: ký tự \" có 2 nét nằm NGANG cạnh nhau phía trên → liên tưởng chia theo chiều NGANG (trên/dưới)."},
    {"prompt": "Chia pane hiện tại theo chiều DỌC (pane mới nằm bên phải).", "hint": "Gõ: C-b %  (hoặc: ctrl+b %)", "patterns": ["^c-b\\s*%$", "^ctrl-b\\s*%$", "^ctrl\\+b\\s*%$"], "explanation": "C-b % split-window dọc — pane mới xuất hiện bên phải pane hiện tại. Mẹo nhớ: ký tự % có 2 chấm nằm CHÉO 2 bên (trái/phải) → liên tưởng chia theo chiều DỌC (trái/phải)."},
    {"prompt": "Di chuyển qua lại giữa các pane trong cùng window.", "hint": "Gõ: C-b o  (hoặc: ctrl+b o)", "patterns": ["^c-b\\s*o$", "^ctrl-b\\s*o$", "^ctrl\\+b\\s*o$", "^ctrl\\+b\\s*arrow$", "^c-b\\s*arrow$"], "explanation": "C-b o nhảy vòng qua pane kế tiếp; C-b + mũi tên di chuyển theo đúng hướng không gian (lên/xuống/trái/phải). Mẹo nhớ: o = other (pane khác, không cần nhớ hướng)."},
    {"prompt": "Detach (thoát tạm) khỏi session tmux, giữ session vẫn chạy nền.", "hint": "Gõ: C-b d  (hoặc: ctrl+b d)", "patterns": ["^c-b\\s*d$", "^ctrl-b\\s*d$", "^ctrl\\+b\\s*d$"], "explanation": "C-b d detach khỏi session — mọi lệnh đang chạy trong session (vd reverse shell listener) vẫn tiếp tục chạy nền. Mẹo nhớ: d = detach (tách ra, không phải thoát/kill)."},
    {"prompt": "Gắn lại (attach) vào session tmux gần nhất, gõ trực tiếp trên terminal (không phải phím tắt trong tmux).", "hint": "Gõ: tmux attach   (hoặc: tmux a)", "patterns": ["^tmux\\s+attach(-session)?\\s*$", "^tmux\\s+a\\s*$"], "explanation": "tmux attach (viết tắt tmux a) gắn lại vào session cuối cùng còn chạy — cứu bạn khi mất kết nối SSH giữa lúc đang thực chiến."},
    {"prompt": "Đổi tên window hiện tại.", "hint": "Gõ: C-b ,  (hoặc: ctrl+b ,)", "patterns": ["^c-b\\s*,$", "^ctrl-b\\s*,$", "^ctrl\\+b\\s*,$"], "explanation": "C-b , mở ô nhập tên mới cho window — đặt tên theo mục tiêu (vd ''recon'', ''web'', ''shell'') giúp không bị rối giữa nhiều window. Mẹo nhớ: dấu , (comma) không có nghĩa trực quan — cứ nhớ 'dấu phẩy để ghi chú/đặt tên'."}
  ]'::jsonb,
  false
)
WHERE kind = 'lab' AND front = 'Lab 10 — tmux: thao tác window/pane nhanh';
