-- 0069_pmp_flow_diagrams: embed 10 hand-drawn flow/process diagrams (SVG,
-- served from web/public/diagrams/pmp/) directly into the matching PMP
-- cheat-sheet sections, using the markdown image syntax now supported by
-- the frontend renderer. Appends (never overwrites); idempotent via a
-- NOT LIKE guard on each diagram's own path (can't already be present).

-- 5. Delivery approaches: Predictive, Iterative, Incremental, Agile, Hybrid
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$![So sánh Predictive (chuỗi tuần tự) và Agile (vòng lặp sprint)](/diagrams/pmp/agile-vs-predictive.svg)
*Predictive đi hết 1 chuỗi Initiate→Plan→Execute→Close; Agile lặp lại chu kỳ sprint ngắn, mỗi vòng đều có review & điều chỉnh.*$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pmp'
  AND cn.title = '5. Delivery approaches: Predictive, Iterative, Incremental, Agile, Hybrid'
  AND cn.body NOT LIKE '%/diagrams/pmp/agile-vs-predictive.svg%';

-- 9. People — Xây dựng team & giải quyết xung đột
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$![5 giai đoạn Tuckman và 5 kiểu xử lý xung đột theo quyết đoán/hợp tác](/diagrams/pmp/team-conflict.svg)
*Forming→Storming→Norming→Performing→Adjourning là vòng đời team; Collaborate (thắng-thắng) là kiểu PMI khuyến khích nhất khi xử lý xung đột.*$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pmp'
  AND cn.title = '9. People — Xây dựng team & giải quyết xung đột'
  AND cn.body NOT LIKE '%/diagrams/pmp/team-conflict.svg%';

-- 10. People — Stakeholder & Communication (2 diagrams)
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$![Mô hình giao tiếp Sender–Receiver với nhiễu (noise) và phản hồi (feedback)](/diagrams/pmp/communication-model.svg)
*Thông điệp luôn đi qua "nhiễu" (ngôn ngữ, văn hoá, cảm xúc) — đó là lý do PM luôn cần vòng feedback để xác nhận đối phương hiểu đúng.*

![Ma trận Power/Interest 4 ô để phân loại stakeholder](/diagrams/pmp/stakeholder-power-interest.svg)
*Power cao + Interest cao → Manage Closely (ưu tiên giao tiếp hàng đầu); Power thấp + Interest thấp → Monitor (tốn công sức tối thiểu).*$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pmp'
  AND cn.title = '10. People — Stakeholder & Communication'
  AND cn.body NOT LIKE '%/diagrams/pmp/communication-model.svg%';

-- 11. Process — Integration, Charter & Change control (2 diagrams)
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$![Vòng đời dự án qua 5 nhóm quy trình PMBOK và dải Monitoring & Controlling bao trùm](/diagrams/pmp/lifecycle-overview.svg)
*Initiating→Planning→Executing→Closing chạy tuần tự, nhưng Monitoring & Controlling bao trùm suốt từ đầu tới cuối, không phải 1 bước riêng.*

![Quy trình Change Control chi tiết: Change request → Log → Impact analysis → CCB → Approved/Rejected](/diagrams/pmp/change-control-detail.svg)
*Mọi thay đổi (dù nhỏ) đều phải đi qua CCB (Change Control Board) trước khi được áp dụng — đừng tự ý "linh động" thay đổi scope/schedule/cost.*$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pmp'
  AND cn.title = '11. Process — Integration, Charter & Change control'
  AND cn.body NOT LIKE '%/diagrams/pmp/lifecycle-overview.svg%';

-- 15. Process — Risk Management
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$![Quy trình Risk Management: Identify → Qualitative → Quantitative → Plan Responses → Implement → Monitor (vòng lặp)](/diagrams/pmp/risk-management-process.svg)
*Risk management không làm 1 lần rồi xong — đường nét đứt quay lại từ Monitor về Identify thể hiện đây là vòng lặp suốt dự án.*$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pmp'
  AND cn.title = '15. Process — Risk Management'
  AND cn.body NOT LIKE '%/diagrams/pmp/risk-management-process.svg%';

-- 16. Process — Quality Management
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$![Vòng Plan/Manage(QA)/Control(QC) Quality và cây Cost of Quality (Conformance vs Nonconformance)](/diagrams/pmp/quality-cost-flow.svg)
*PMI luôn ưu tiên phòng ngừa (Prevention/Appraisal) hơn là để lỗi xảy ra rồi sửa (Internal/External failure) — đầu tư Cost of Conformance càng nhiều thì Cost of Nonconformance càng giảm.*$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pmp'
  AND cn.title = '16. Process — Quality Management'
  AND cn.body NOT LIKE '%/diagrams/pmp/quality-cost-flow.svg%';

-- 17. Process — Resource Management, cấu trúc tổ chức & PMO
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$![Phổ cấu trúc tổ chức từ Functional đến Projectized, quyền lực PM tăng dần](/diagrams/pmp/org-structure-spectrum.svg)
*Đề thi thường hỏi "PM gặp khó vì không đủ quyền với team" → suy ra ngay đây là Functional hoặc Weak Matrix, không cần đọc hết câu hỏi.*$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pmp'
  AND cn.title = '17. Process — Resource Management, cấu trúc tổ chức & PMO'
  AND cn.body NOT LIKE '%/diagrams/pmp/org-structure-spectrum.svg%';

-- 18. Process — Procurement & Contracts
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$![Vòng đời Procurement (Plan→Conduct→Control→Close) và phổ rủi ro theo loại hợp đồng](/diagrams/pmp/procurement-lifecycle.svg)
*FFP dồn hết rủi ro cho seller (giá cố định); T&M chia sẻ rủi ro; CPFF dồn phần lớn rủi ro cho buyer (chi phí thực tế + phí cố định).*$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pmp'
  AND cn.title = '18. Process — Procurement & Contracts'
  AND cn.body NOT LIKE '%/diagrams/pmp/procurement-lifecycle.svg%';
