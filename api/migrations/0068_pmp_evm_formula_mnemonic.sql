-- 0068_pmp_evm_formula_mnemonic: add a mnemonic framing for the EVM/forecasting
-- formulas to "25. Công thức tổng hợp & ví dụ tính nhanh" — complements the
-- existing worked examples with a way to *derive* the formulas from 3 root
-- variables (BAC/AC/EV) instead of memorizing each one separately. Appends
-- (never overwrites); idempotent via a NOT LIKE guard on a unique marker.

UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$**Mẹo nhớ công thức EVM theo quy tắc, không học thuộc từng cái** **[Chung]**:

Chỉ cần nhớ **3 biến gốc**: **BAC** (tổng ngân sách ban đầu), **AC** (đã thực chi), **EV** (giá trị công việc đã hoàn thành, quy ra tiền theo kế hoạch) — toàn bộ công thức còn lại chỉ là so sánh 3 biến này theo 1 quy tắc duy nhất:
- **Variance = trừ** → kết quả dương là tốt (dư/sớm): `CV = EV − AC`, `SV = EV − PV`
- **Index = chia** → kết quả > 1 là tốt: `CPI = EV / AC`, `SPI = EV / PV`
- Cả Variance và Index đều **luôn đặt EV lên trước/tử số** — đây là mẹo nhớ chiều phép tính, khỏi nhầm `AC − EV` hay `AC / EV`.

**Nhóm dự báo (ETC/EAC) — nhớ theo "còn phải chi / sẽ chi hết bao nhiêu", 2 công thức là nghịch đảo của nhau**:
- `EAC = AC + ETC` (tổng sẽ chi = đã chi + còn phải chi)
- `ETC = EAC − AC` (còn phải chi = tổng dự kiến − đã chi) — chỉ cần nhớ 1 trong 2, suy ra cái kia bằng cách đổi vế.
- `EAC = BAC / CPI` — dùng khi giả định hiệu suất chi tiêu hiện tại (CPI) giữ nguyên tới hết dự án (trường hợp "typical" hay gặp nhất trong đề thi).
- `TCPI = (BAC − EV) / (BAC − AC)` — đọc là "phần việc CÒN LẠI" chia "tiền CÒN LẠI": hiệu suất CẦN có từ giờ tới cuối dự án để không vỡ ngân sách.

Khi gặp câu hỏi EVM trong đề thi: viết 3 biến gốc (BAC, AC, EV) ra giấy nháp trước, rồi áp đúng 1 trong 2 quy tắc (trừ = variance, chia = index) thay vì cố nhớ mặt chữ từng công thức.$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pmp'
  AND cn.title = '25. Công thức tổng hợp & ví dụ tính nhanh'
  AND cn.body NOT LIKE '%Variance = trừ%';
