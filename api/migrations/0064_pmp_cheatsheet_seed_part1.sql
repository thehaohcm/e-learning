-- 0064_pmp_cheatsheet_seed_part1: a Cheatsheet for the PMP subject, same mechanism
-- as Pentest (0052 + 0054): (1) a "cheatsheet" deck so AI-generated questions and
-- Study have somewhere to live, (2) the learner's own course_notes seeded with
-- PMP sections 1-13 (exam info, mindset, agile, people, process areas). Part 2
-- (0065) adds sections 14-28. Each section is Markdown, tagged **[People]**,
-- **[Process]**, **[Business]** or **[Chung]** (matching the ECO domains) so the
-- Cheat sheet page's filter chips work. Content is original study notes, NOT
-- PMI exam questions. Seeded once for this account, skipping titles that exist.

INSERT INTO decks (subject_id, slug, name, description, default_kind, sort_order)
SELECT s.id, 'cheatsheet', 'Cheatsheet',
       'PMP cheatsheet by section. Editable, studyable, and the source for AI-generated questions.',
       'concept', 0
FROM subjects s WHERE s.slug = 'pmp'
ON CONFLICT (subject_id, slug) DO NOTHING;

INSERT INTO course_notes (user_id, subject_id, title, body, sort_order, links, created_at, updated_at)
SELECT u.id, s.id, v.title, v.body, v.ord, v.links, NOW(), NOW()
FROM (VALUES
  ($cheat$1. Thông tin kỳ thi PMP (ECO từ 09/07/2026)$cheat$, $cheat$**Format kỳ thi** **[Chung]**:
- **180 câu** = 170 câu tính điểm + 10 câu **unscored** (pretest, không biết câu nào)
- **240 phút** (ECO cũ 2021: 230 phút); trung bình ≈ 80 giây/câu
- **2 break × 10 phút**: break 1 sau phần **case study**, break 2 khoảng giữa phần câu độc lập; đã vào break thì **không quay lại** phần trước
- Đề mở đầu bằng phần case study (nhóm câu hỏi dựa trên một tình huống chung)
- Kết quả theo từng domain: Above Target / Target / Below Target / Needs Improvement (PMI không công bố điểm đậu cụ thể)

**3 domain và tỷ trọng (ECO 2026)** **[Chung]**:
- **People 33%** — 8 tasks (ECO 2021: 42%)
- **Process 41%** — 10 tasks (ECO 2021: 50%)
- **Business Environment 26%** — 8 tasks (ECO 2021: 8%)
- Tổng **26 tasks** (ECO 2021: 35 tasks, 42/50/8, 180 câu / 230 phút)
- Tỷ lệ cách tiếp cận xuyên cả 3 domain: khoảng **40% predictive**, **60% agile + hybrid**
- **AI** và **sustainability** xuất hiện trong các tình huống (scenario), không phải domain riêng

**8 dạng câu hỏi**: Case/Scenario (mới), Graphic-based (chính thức hóa), Multiple-choice (1 đáp án), Multiple-response (nhiều đáp án), Matching, Enhanced matching, Point and click (hotspot), Pull-down list. Dạng fill-in-the-blank giới hạn đã bỏ. Một số dạng (matching, point and click, pull-down) theo nguồn bên thứ ba chỉ có ở test center.

**Task nhắc tên trong ECO 2026** (theo nguồn thứ ba): "Develop a common vision", "Help ensure value-based delivery", "Plan and manage risk", "Manage and control changes", "Define and establish project governance", "Remove impediments and manage issues", "Support organizational change".

**Lưu ý nguồn**: các con số trên tổng hợp từ bài viết bên thứ ba khi đề mới ra mắt; luôn đối chiếu **ECO PDF chính thức** và trang PMP trên pmi.org (trang PMI từng còn hiển thị số liệu cũ 230 phút / 175+5 câu / 42-50-8).$cheat$, 1001, '[{"url": "https://www.pmi.org/certifications/project-management-pmp", "label": "PMI – trang chính thức chứng chỉ PMP"}]'::jsonb),

  ($cheat$2. Điều kiện thi, cấp chứng chỉ & duy trì$cheat$, $cheat$**Điều kiện dự thi** (kiểm tra lại trên pmi.org trước khi đăng ký) **[Chung]**:
- **Bằng cử nhân (4 năm)**: ≥ **36 tháng** kinh nghiệm lãnh đạo/chỉ đạo dự án (không chồng lấn thời gian, trong **8 năm** gần nhất) + **35 contact hours** đào tạo quản lý dự án (hoặc có chứng chỉ CAPM)
- **Trung học / Associate (2 năm)**: ≥ **60 tháng** kinh nghiệm (trong 8 năm gần nhất) + **35 contact hours**
- Bằng từ chương trình được **GAC-accredited**: yêu cầu kinh nghiệm ngắn hơn (24 tháng) và không cần 35 giờ đào tạo
- Kinh nghiệm tính cả dự án predictive, agile, hybrid; PMI có thể **audit** hồ sơ

**Thi & thi lại**:
- Có **1 năm** kể từ khi đơn được duyệt, tối đa **3 lần** thi trong thời gian đó
- Thi tại test center hoặc online có giám thị (online proctored)

**Duy trì chứng chỉ** **[Chung]**:
- Hiệu lực **3 năm**; cần **60 PDUs** mỗi chu kỳ
- Talent Triangle: **Ways of Working**, **Power Skills**, **Business Acumen** — tối thiểu **8 PDUs** mỗi nhóm, phần còn lại tùy chọn
- PDU kiếm từ học tập, đóng góp cộng đồng (dạy, tình nguyện, viết)

**Đạo đức**: PMI có Code of Ethics — vi phạm (gian lận hồ sơ, lộ đề thi) có thể bị thu hồi chứng chỉ. Đề thi là tài liệu bảo mật, không chia sẻ câu hỏi thật.$cheat$, 1002, '[]'::jsonb),

  ($cheat$3. Mindset PMI & chiến lược làm bài$cheat$, $cheat$**Tư duy "PMI-ism"** **[Chung]**:
- PM **chủ động**, không chờ chỉ đạo; **phân tích trước khi hành động**
- Làm theo kế hoạch, hợp đồng, quy trình change control; không tự ý đổi scope/baseline
- PM tạo giá trị (value) cho tổ chức, không chỉ "làm đúng tiến độ - ngân sách"
- Predictive: tuân thủ plan + change control. Agile: tin tưởng team tự tổ chức, PM/Scrum Master là **servant leader** (gỡ vướng mắc, không giao việc)

**Thứ tự phản ứng điển hình khi có vấn đề** **[Chung]**:
1. Thu thập thông tin, **đánh giá tác động** (scope/time/cost/quality/risk)
2. Xem lại tài liệu: project plan, risk register, hợp đồng, charter
3. Trao đổi với team/stakeholder liên quan (trực tiếp, riêng tư)
4. Đề xuất giải pháp; nếu ảnh hưởng baseline → **change request** → CCB
5. Cập nhật tài liệu, communicate kết quả

**Các phản ứng thường SAI trong đề**: sa thải ngay, báo sponsor/escalate khi chưa phân tích, im lặng chịu đựng, tự thêm tính năng (**gold plating**), che giấu vấn đề, chấp nhận thay đổi không qua quy trình, tự ý cắt scope/chất lượng để kịp hạn.

**Escalate** chỉ khi: vượt thẩm quyền/tolerance của PM, hoặc đã thử giải quyết trong nhóm mà không được.

**Từ khóa trong câu hỏi**: **first** (việc đầu tiên: thường là phân tích/xác nhận thông tin), **best** (phương án tốt nhất trong các phương án đều khả thi), **next** (bước kế tiếp trong quy trình), **most likely**, **NOT / EXCEPT**.

**Chiến thuật thời gian** **[Chung]**:
- ≈ 80 giây/câu; câu khó **đánh dấu (mark for review)** rồi làm tiếp
- Không bỏ trống câu nào (không trừ điểm sai)
- Loại 2 đáp án sai rõ ràng trước, rồi chọn theo nguyên tắc PMI
- Case study: đọc kỹ tình huống một lần, các câu hỏi sau dùng chung ngữ cảnh
- Tận dụng 2 break để nghỉ, ăn nhẹ, không cần vội$cheat$, 1003, '[]'::jsonb),

  ($cheat$4. PMI Code of Ethics & Professional Conduct$cheat$, $cheat$**4 giá trị cốt lõi** **[Chung]**: **Responsibility** (trách nhiệm), **Respect** (tôn trọng), **Fairness** (công bằng), **Honesty** (trung thực).

Chuẩn mực gồm 2 loại: **aspirational** (hướng tới, nên làm) và **mandatory** (bắt buộc tuân thủ).

**Tình huống đạo đức & cách xử lý chuẩn** **[Business]**:
- **Xung đột lợi ích** (người thân dự thầu, lợi ích cá nhân) → **công khai (disclose)** với bên liên quan, rút khỏi quyết định
- **Quà tặng / hối lộ** → từ chối, báo cáo theo chính sách; không nhận lợi ích làm ảnh hưởng quyết định
- **Thông tin bảo mật** (của khách hàng, đối thủ, nội bộ) → bảo vệ, không dùng cho lợi ích cá nhân (insider)
- **Phát hiện vi phạm** của đồng nghiệp/nhà thầu → báo qua **kênh phù hợp** (ethics/compliance, quản lý), không tự xử
- **Sai sót của chính mình** → thừa nhận sớm, sửa, thông báo cho người bị ảnh hưởng
- **Báo cáo trạng thái** → trung thực, dù tin xấu; không tô hồng (misrepresent)
- **Bản quyền / sở hữu trí tuệ** → không dùng tài liệu có bản quyền khi chưa được phép
- **Khác biệt văn hóa** → tôn trọng, tìm hiểu quy tắc địa phương, không phân biệt đối xử
- **Áp lực từ cấp trên làm việc sai quy định** → từ chối làm sai, giải thích hậu quả, escalate theo chính sách

**Quy tắc chọn đáp án**: ưu tiên minh bạch, tuân thủ luật/chính sách và hợp đồng, bảo vệ lợi ích tổ chức/khách hàng/cộng đồng; không chọn "im lặng cho yên chuyện" hay "tự ý làm ngoài quy trình".$cheat$, 1004, '[]'::jsonb),

  ($cheat$5. Delivery approaches: Predictive, Iterative, Incremental, Agile, Hybrid$cheat$, $cheat$**Các cách tiếp cận (development approach)** **[Chung]**:
- **Predictive (plan-driven / waterfall)**: scope, time, cost xác định sớm; thay đổi qua change control; hợp với yêu cầu ổn định, rủi ro thấp (xây dựng, hạ tầng)
- **Iterative**: lặp nhiều vòng để tinh chỉnh sản phẩm theo feedback
- **Incremental**: giao từng phần chức năng hoàn chỉnh theo thời gian
- **Agile**: iterative + incremental, giao giá trị sớm, thích nghi thay đổi
- **Hybrid**: kết hợp (vd predictive cho phần cứng/hạ tầng + agile cho phần mềm)

**Tam giác ràng buộc**:
- Predictive: **scope cố định**, time/cost ước lượng và kiểm soát
- Agile: **time & cost cố định (timebox)**, scope linh hoạt (ưu tiên theo giá trị)

**Chọn approach (tailoring)** **[Chung]** — dựa vào: mức độ không chắc chắn của yêu cầu, độ phức tạp, quy mô, mức tiếp cận khách hàng, quy định/compliance, văn hóa tổ chức, kỹ năng team.
- Yêu cầu thay đổi nhiều, cần feedback nhanh → **agile**
- Yêu cầu rõ, hợp đồng fixed, quy định nghiêm → **predictive** hoặc **hybrid**

**Đo tiến độ**: predictive dùng EVM (CPI, SPI); agile dùng velocity, burndown/burnup, số sprint còn lại.

**Life cycle**: project life cycle (initiating → closing) khác product life cycle; phase gate là điểm duyệt cuối giai đoạn ở predictive.

**Chi phí thay đổi**: predictive càng về sau đổi càng đắt; agile giữ chi phí thay đổi thấp nhờ feedback ngắn.

**Hybrid khi nào**: một phần đã cố định bởi quy định/hợp đồng, phần còn lại chưa chắc; hoặc tổ chức đang chuyển đổi sang agile.$cheat$, 1005, '[]'::jsonb),

  ($cheat$6. Agile Manifesto & Scrum$cheat$, $cheat$**4 giá trị Agile Manifesto** **[Chung]**:
- **Cá nhân & tương tác** hơn quy trình & công cụ
- **Phần mềm chạy được** hơn tài liệu đầy đủ
- **Hợp tác với khách hàng** hơn đàm phán hợp đồng
- **Phản hồi với thay đổi** hơn bám theo kế hoạch

(Vế bên phải vẫn có giá trị, nhưng vế trái được đặt cao hơn.) Điểm nhấn 12 nguyên tắc: giao sớm và liên tục, đón nhận thay đổi, giao thường xuyên (tuần → tháng), business và dev làm việc cùng nhau hằng ngày, **working software là thước đo tiến độ chính**, nhịp độ bền vững, đơn giản (tối đa hóa lượng việc **không** phải làm), team tự tổ chức, định kỳ **reflect & adjust**.

**Scrum (Scrum Guide 2020)** **[Chung]**:
- **3 accountabilities**: **Product Owner** (tối đa hóa giá trị, sở hữu & sắp thứ tự Product Backlog; 1 người), **Scrum Master** (đảm bảo Scrum được hiểu, gỡ impediment, coach; servant-leader), **Developers** (tạo Increment; tự quản lý)
- Scrum Team nhỏ, thường ≤ **10** người
- **5 events**: **Sprint** (≤ 1 tháng, container của mọi event), **Sprint Planning** (≤ 8h với sprint 1 tháng; why/what/how), **Daily Scrum** (**15 phút**, cho Developers), **Sprint Review** (≤ 4h; inspect Increment, adapt backlog), **Sprint Retrospective** (≤ 3h; cải tiến quy trình/con người/công cụ)
- **3 artifacts + commitment**: Product Backlog → **Product Goal**; Sprint Backlog → **Sprint Goal**; Increment → **Definition of Done**
- **3 trụ cột**: Transparency, Inspection, Adaptation. **5 giá trị**: Commitment, Focus, Openness, Respect, Courage

**Điểm hay ra đề**:
- Sprint Goal không đổi giữa sprint; **chỉ Product Owner** được hủy sprint
- Thay đổi scope trong sprint được thương lượng với PO khi đã hiểu thêm
- **Backlog refinement** diễn ra liên tục, không phải event chính thức
- Increment phải **usable** và đạt DoD
- Stakeholder không chỉ đạo Developers; yêu cầu đi qua PO
- Scrum Master không giao việc, không là "project manager" của team$cheat$, 1006, '[{"url": "https://scrumguides.org/scrum-guide.html", "label": "Scrum Guide (bản chính thức)"}, {"url": "https://agilemanifesto.org/", "label": "Agile Manifesto & 12 principles"}]'::jsonb),

  ($cheat$7. Agile: Kanban, XP, ước lượng & metrics$cheat$, $cheat$**Kanban** **[Chung]**:
- Thực hành: **visualize** workflow, **giới hạn WIP**, quản lý flow, chính sách rõ ràng, vòng feedback, cải tiến cộng tác
- **Pull system**: làm xong mới kéo việc mới; WIP limit làm lộ **bottleneck**
- **Lead time** = từ lúc yêu cầu đến lúc xong; **Cycle time** = từ lúc bắt đầu làm đến xong; **Throughput** = số item xong / đơn vị thời gian
- **Little's Law**: Cycle time = WIP / Throughput
- **Cumulative Flow Diagram (CFD)**: băng màu phình rộng = nút thắt

**XP (Extreme Programming)**: pair programming, **TDD**, continuous integration, refactoring, collective code ownership, simple design, small releases, sustainable pace.

**Ước lượng** **[Chung]**:
- **Story points** (tương đối), **Planning Poker** (dãy Fibonacci), **T-shirt sizing**, affinity estimation, ideal days
- **Velocity** = tổng story points **hoàn thành** trong sprint (làm dở = 0)
- Dự báo: số sprint còn lại = backlog còn lại / velocity
- **Burndown**: công việc còn lại giảm dần; **Burnup**: công việc đã xong tăng + đường scope (thấy rõ scope thay đổi)

**User story & backlog**:
- Mẫu: "Là một <vai trò>, tôi muốn <việc> để <lợi ích>" + **acceptance criteria**
- **INVEST**: Independent, Negotiable, Valuable, Estimable, Small, Testable
- Phân cấp: Epic → Feature → Story → Task
- **Spike**: time-box nghiên cứu để giảm bất định; **MVP**: bản tối thiểu để học nhanh
- **Definition of Done** (áp dụng mọi item) khác **acceptance criteria** (riêng từng story) khác **Definition of Ready**

**Khác**: information radiator (bảng burndown, kanban board hiển thị công khai), retrospective, team working agreement, kỹ thuật nợ (technical debt), Lean (loại bỏ lãng phí), DevOps (CI/CD).$cheat$, 1007, '[]'::jsonb),

  ($cheat$8. People — Leadership, quyền lực & động lực$cheat$, $cheat$**Leader vs Manager** **[People]**: leader tập trung tầm nhìn, truyền cảm hứng, thay đổi; manager tập trung kiểm soát, quy trình, đạt mục tiêu. PM cần cả hai.

**Phong cách lãnh đạo** **[People]**: **Servant** (phục vụ team, gỡ vướng mắc — ưu tiên trong agile), **Transactional** (thưởng/phạt theo kết quả), **Transformational** (truyền cảm hứng), **Charismatic**, **Laissez-faire** (thả nổi), **Interactional** (kết hợp).
- **Situational leadership** (Hersey–Blanchard): Directing → Coaching → Supporting → Delegating, tùy mức độ trưởng thành của người làm.

**Emotional Intelligence** (Goleman): self-awareness, self-regulation, motivation, empathy, social skills.

**5 loại quyền lực của PM** **[People]**:
- **Positional / Formal** (chức vụ), **Reward**, **Penalty / Coercive**, **Expert**, **Referent** (được quý mến/ngưỡng mộ)
- Hiệu quả nhất: **Expert** và **Reward**; kém bền: **Penalty**

**Lý thuyết động lực** **[People]**:
- **Maslow**: Physiological → Safety → Social → Esteem → Self-actualization
- **Herzberg**: **hygiene** (lương, điều kiện, chính sách) chỉ ngăn bất mãn; **motivators** (thành tựu, ghi nhận, trách nhiệm) mới tạo động lực
- **McGregor**: Theory **X** (cần giám sát) vs Theory **Y** (tự giác, có động lực nội tại); **Ouchi Theory Z** (cam kết dài hạn, trung thành)
- **McClelland**: nhu cầu **Achievement**, **Affiliation**, **Power**
- **Vroom (Expectancy)**: động lực = tin rằng nỗ lực → kết quả → phần thưởng có giá trị

**Khen thưởng**: cá nhân hóa; thưởng kiểu "chỉ một người thắng" (win-lose) làm hại tinh thần đội.

**Coaching / mentoring**: coach nâng năng lực, mentor chia sẻ kinh nghiệm dài hạn.$cheat$, 1008, '[]'::jsonb),

  ($cheat$9. People — Xây dựng team & giải quyết xung đột$cheat$, $cheat$**Tuckman — 5 giai đoạn phát triển team** **[People]**: **Forming** (làm quen, cần định hướng) → **Storming** (xung đột, tranh quyền) → **Norming** (hình thành quy tắc, hợp tác) → **Performing** (hiệu suất cao, tự chủ) → **Adjourning** (kết thúc, giải tán).
- Thành viên mới/đổi người có thể đưa team quay lại Storming.

**Xây dựng team** **[People]**: team charter / ground rules / working agreements, mục tiêu chung, tin cậy và **psychological safety**, colocation hoặc công cụ cho **virtual team** (múi giờ, văn hóa), training, hoạt động team building, ghi nhận kịp thời.

**Agile team**: cross-functional, **self-organizing**, kỹ năng chữ T (T-shaped), swarming (cả team xử lý một item).

**5 kỹ thuật giải quyết xung đột** **[People]**:
- **Collaborate / Problem-solve** — cùng tìm giải pháp win-win, **ưu tiên nhất**
- **Compromise** — mỗi bên nhường một phần (thua-thua một phần)
- **Smooth / Accommodate** — nhấn mạnh điểm chung, nhường để giữ hòa khí (tạm thời)
- **Force / Direct** — áp đặt quyết định (win-lose), chỉ khi khẩn cấp/an toàn
- **Withdraw / Avoid** — né tránh, hoãn (không giải quyết gốc)

**Nguyên tắc**: xử lý **sớm**, riêng tư, giữa các bên liên quan trước; tập trung vào **vấn đề, không vào con người**; chỉ chuyển lên cấp trên khi các bên không tự giải quyết được.
- Nguyên nhân xung đột hay gặp: lịch trình, ưu tiên dự án, nguồn lực, ý kiến kỹ thuật, thủ tục hành chính, chi phí, cá tính.

**Ra quyết định nhóm**: unanimity (nhất trí), majority (>50%), plurality (nhiều phiếu nhất), dictatorship (một người quyết). Kỹ thuật thu ý kiến: brainstorming, **nominal group technique**, **Delphi** (ẩn danh nhiều vòng), multicriteria.

**Gỡ vướng mắc (impediments)**: PM/Scrum Master ghi vào **issue log**, xử lý hoặc escalate; không để team tự "chịu trận".$cheat$, 1009, '[]'::jsonb),

  ($cheat$10. People — Stakeholder & Communication$cheat$, $cheat$**Stakeholder** **[People]**:
- Nhận diện **sớm và liên tục**; ghi vào **stakeholder register**
- **Power/Interest grid**: quyền lực cao + quan tâm cao → **Manage closely**; quyền lực cao + quan tâm thấp → **Keep satisfied**; quyền lực thấp + quan tâm cao → **Keep informed**; thấp/thấp → **Monitor**
- **Salience model**: Power, Urgency, Legitimacy
- Mức engagement: **Unaware, Resistant, Neutral, Supportive, Leading**; so sánh hiện tại (C) và mong muốn (D) trong engagement assessment matrix
- Stakeholder engagement plan; quản lý kỳ vọng, đàm phán, giải quyết vấn đề
- Phân biệt: **sponsor** (tài trợ, ban hành charter), **customer/user** (dùng/ chấp nhận sản phẩm), **PMO**, **functional manager**

**Communication** **[People]**:
- Số kênh giao tiếp = **n(n−1)/2** (n gồm cả PM)
- Phương pháp: **Interactive** (họp, gọi điện), **Push** (email, báo cáo gửi đi), **Pull** (intranet, kho tài liệu)
- Mô hình: sender → encode → medium → decode → receiver, có **noise** và **feedback**
- Việc nhạy cảm, xung đột, phức tạp → **trực tiếp (face-to-face)**; quyết định chính thức → **văn bản**
- **Communications management plan**: ai cần gì, khi nào, bằng cách nào, ai gửi; cập nhật khi stakeholder thay đổi
- Kỹ năng: **active listening** (nhắc lại để xác nhận), tailor thông điệp theo đối tượng, tôn trọng khác biệt văn hóa/múi giờ
- Họp hiệu quả: agenda rõ, timebox, biên bản/action items
- PM thường dành phần lớn thời gian cho giao tiếp (hay trích dẫn ~90%)

**Agile**: information radiator, daily stand-up, sprint review để stakeholder xem sản phẩm thật.

**Khi stakeholder phản đối**: tìm hiểu nguyên nhân, lắng nghe, tác động bằng ảnh hưởng/đàm phán; không phớt lờ hay loại bỏ họ.$cheat$, 1010, '[]'::jsonb),

  ($cheat$11. Process — Integration, Charter & Change control$cheat$, $cheat$**Khởi tạo** **[Process]**:
- **Business case** (vì sao nên làm) có trước **Project Charter**
- **Charter**: ban hành bởi **sponsor / initiator** (ngoài ranh giới dự án); chính thức **ủy quyền** dự án và PM dùng nguồn lực; chứa mục tiêu, yêu cầu mức cao, milestone tổng quát, ngân sách sơ bộ, tiêu chí thành công, PM được chỉ định
- **Assumption log**, **benefits management plan**, kick-off meeting

**Project Management Plan** **[Process]**: gồm các plan phụ (scope, requirements, schedule, cost, quality, resource, communications, risk, procurement, stakeholder engagement), **change & configuration management plan**, **3 baseline** (scope, schedule, cost = performance measurement baseline), development approach và life cycle.

**Thực thi & giám sát**:
- Direct and Manage Project Work → deliverables, work performance data, issue log, change requests
- Monitor and Control → work performance information/reports
- **Manage Project Knowledge**: tacit (kinh nghiệm ngầm) vs explicit (văn bản); ghi **lessons learned register** liên tục

**Perform Integrated Change Control** **[Process]**:
1. **Đánh giá tác động** (scope, schedule, cost, quality, risk, resource) — việc đầu tiên
2. Tìm phương án thay thế
3. Trình **CCB / change authority** để duyệt / từ chối / hoãn
4. Triển khai (chỉ khi được duyệt; khẩn cấp theo quy trình đã định)
5. Cập nhật plan/baseline, **communicate** cho stakeholder
- Loại change request: **corrective**, **preventive**, **defect repair**, **update**
- **Change log** theo dõi trạng thái; **configuration management** quản lý phiên bản các item
- **Agile**: thay đổi = thêm/sắp xếp lại item trong backlog, **Product Owner** quyết định

**Phân biệt**: change **không** được duyệt thì không đổi baseline; PM không tự duyệt thay đổi vượt quyền hạn.$cheat$, 1011, '[]'::jsonb),

  ($cheat$12. Process — Scope Management$cheat$, $cheat$**Product scope** (tính năng/đặc điểm sản phẩm) vs **Project scope** (công việc để tạo ra sản phẩm) **[Process]**.

**Thu thập yêu cầu (Collect Requirements)**: phỏng vấn, focus group, khảo sát, quan sát, prototype, benchmarking, context diagram, affinity diagram, mind map, nominal group, Delphi.
- Loại yêu cầu: business, stakeholder, solution (functional / non-functional), transition, project, quality
- **Requirements traceability matrix (RTM)**: nối yêu cầu → deliverable → test; chống bỏ sót và scope creep
- Ưu tiên hóa: **MoSCoW** (Must, Should, Could, Won't), Kano, bỏ phiếu 100 điểm

**Define Scope → Project scope statement**: mô tả scope, deliverables, **acceptance criteria**, **exclusions**, assumptions, constraints.

**WBS (Work Breakdown Structure)** **[Process]**:
- Phân rã deliverable thành **work package** (mức thấp nhất, có thể ước lượng cost/time)
- **Quy tắc 100%**: WBS gồm 100% công việc của scope, không thêm không bớt
- **WBS dictionary**: mô tả chi tiết từng phần tử WBS
- **Scope baseline = scope statement + WBS + WBS dictionary**
- **Rolling wave planning**: lập kế hoạch chi tiết phần gần, tổng quát phần xa (progressive elaboration)

**Validate Scope vs Control Quality** **[Process]**:
- **Control Quality**: kiểm tra deliverable đúng yêu cầu (**verified**, nội bộ) — làm trước
- **Validate Scope**: khách hàng/sponsor **chính thức chấp nhận** deliverable (**accepted**)

**Control Scope**: so sánh thực tế với baseline (variance analysis), quản lý thay đổi.
- **Scope creep**: thêm scope không kiểm soát (không qua change control)
- **Gold plating**: team tự thêm tính năng không ai yêu cầu — không nên làm

**Agile**: scope = **product backlog**, yêu cầu nảy sinh theo thời gian; user story mapping; Definition of Done quyết định hoàn thành.$cheat$, 1012, '[]'::jsonb),

  ($cheat$13. Process — Schedule Management$cheat$, $cheat$**Trình tự** **[Process]**: Define Activities → Sequence Activities → Estimate Durations → Develop Schedule → Control Schedule.

**Quan hệ phụ thuộc (PDM)**: **FS** (Finish-to-Start, phổ biến nhất), **SS**, **FF**, **SF** (hiếm). Phụ thuộc: **mandatory** (hard logic), **discretionary** (soft logic), **external**, **internal**. **Lead** = cho phép gối đầu (âm), **Lag** = chờ thêm.

**Ước lượng thời gian**: analogous (top-down, dựa dự án cũ), parametric, **three-point (PERT)**, bottom-up (chính xác nhất nhưng tốn công), reserve analysis. **Ước lượng ban đầu có độ bất định cao, thu hẹp dần (cone of uncertainty).**

**Critical Path Method (CPM)** **[Process]**:
- **Critical path** = đường dài nhất, quyết định thời gian tối thiểu của dự án; float = 0
- Forward pass: **EF = ES + Duration** (một số tài liệu dùng ES + D − 1; theo nhất quán của đề)
- Backward pass: **LS = LF − Duration**
- **Total float = LS − ES = LF − EF**; **free float** = trễ được mà không ảnh hưởng ES của activity kế tiếp
- Float âm = đang chậm; có thể nhiều critical path; **near-critical path** cần theo dõi

**Nén tiến độ (Schedule compression)** **[Process]**:
- **Crashing**: thêm nguồn lực → **tăng chi phí**; chọn activity trên critical path có chi phí/ngày thấp nhất
- **Fast tracking**: làm song song các activity tuần tự → **tăng rủi ro** (và có thể rework)
- Chỉ nén activity **trên critical path**

**Tối ưu nguồn lực**: **resource leveling** (giải quyết xung đột nguồn lực, **có thể đổi critical path và kéo dài** tiến độ) vs **resource smoothing** (dùng float, **không đổi critical path**).

**Khác**: **Critical chain** (buffer: project buffer, feeding buffer), **Monte Carlo** (mô phỏng xác suất hoàn thành), what-if, **schedule baseline** (bản đã duyệt), Gantt/milestone chart/network diagram.

**Agile**: lập kế hoạch theo release & iteration, timebox, xác định cadence thay vì lịch chi tiết dài.$cheat$, 1013, '[]'::jsonb)
) AS v(title, body, ord, links)
JOIN users u ON u.email = 'thehaohcm@tradesign.io.vn'
JOIN subjects s ON s.slug = 'pmp'
WHERE NOT EXISTS (
    SELECT 1 FROM course_notes cn
    WHERE cn.user_id = u.id AND cn.subject_id = s.id AND cn.title = v.title
);
