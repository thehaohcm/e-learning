-- 0065_pmp_cheatsheet_seed_part2: PMP cheatsheet sections 14-28 (cost/EVM, risk,
-- quality, resources, procurement, closing, business environment, PMBOK map,
-- formulas, confusable terms, scenario patterns, glossary). See 0064 for context.
-- Seeded once for this account, skipping titles that already exist.

INSERT INTO course_notes (user_id, subject_id, title, body, sort_order, links, created_at, updated_at)
SELECT u.id, s.id, v.title, v.body, v.ord, v.links, NOW(), NOW()
FROM (VALUES
  ($cheat$14. Process — Cost Management & Earned Value (EVM)$cheat$, $cheat$**Ước lượng & ngân sách** **[Process]**:
- Kỹ thuật: analogous, parametric, bottom-up, three-point, reserve analysis, funding limit reconciliation
- **Contingency reserve**: cho **known-unknowns**, nằm **trong cost baseline**, PM quản lý
- **Management reserve**: cho **unknown-unknowns**, **ngoài cost baseline** (nhưng thuộc ngân sách dự án), cần phê duyệt để dùng
- **Cost baseline** = tổng work package + contingency reserve; **Budget** = cost baseline + management reserve
- **Sunk cost** (đã chi, không thu hồi) → **không** dùng để quyết định tiếp tục hay dừng
- **Cost of Quality**: conformance (prevention, appraisal) vs nonconformance (internal failure, external failure)

**Biến số EVM** **[Process]**: **PV** (planned value), **EV** (earned value = % hoàn thành × BAC), **AC** (actual cost), **BAC** (budget at completion).

**Công thức EVM**:
```
CV  = EV − AC          CPI = EV / AC     (CPI < 1: vượt chi phí)
SV  = EV − PV          SPI = EV / PV     (SPI < 1: chậm tiến độ)
% hoàn thành = EV / BAC
VAC = BAC − EAC
ETC = EAC − AC
TCPI (để đạt BAC) = (BAC − EV) / (BAC − AC)
TCPI (để đạt EAC) = (BAC − EV) / (EAC − AC)
```

**4 cách tính EAC** (chọn theo giả định):
```
EAC = BAC / CPI                          # hiệu suất chi phí hiện tại tiếp tục (typical)
EAC = AC + (BAC − EV)                    # sai lệch là bất thường, phần còn lại đúng kế hoạch (atypical)
EAC = AC + (BAC − EV) / (CPI × SPI)      # cả CPI và SPI cùng ảnh hưởng
EAC = AC + ETC từ bottom-up              # ước lượng lại hoàn toàn
```

**Diễn giải**: CV/SV **âm** = xấu; CPI/SPI **<1** = xấu. TCPI > 1 nghĩa là phần còn lại phải làm hiệu quả hơn mới kịp ngân sách.

**Ví dụ**: BAC 200k, PV 100k, EV 80k, AC 90k → CV −10k; SV −20k; CPI 0.889; SPI 0.8; EAC (typical) 225k; ETC 135k; VAC −25k; TCPI 1.09.

**Agile**: ngân sách theo timebox & team ổn định; đo bằng velocity, burn rate, giá trị giao được.$cheat$, 1014, '[]'::jsonb),

  ($cheat$15. Process — Risk Management$cheat$, $cheat$**Định nghĩa** **[Process]**: risk = sự kiện/điều kiện bất định, nếu xảy ra có tác động (xấu: **threat**, tốt: **opportunity**) lên mục tiêu. **Issue** = risk đã xảy ra (hoặc vấn đề đang diễn ra) → vào issue log.

**Quy trình**: Plan Risk Management → Identify → Qualitative → Quantitative → Plan Responses → Implement Responses → Monitor Risks.
- **Risk management plan**: phương pháp, vai trò, RBS (risk breakdown structure), định nghĩa xác suất/tác động, ngưỡng
- **Risk register**: danh sách risk, owner, trigger, response, trạng thái

**Nhận diện**: brainstorming, Delphi, phỏng vấn, **root cause analysis**, **SWOT**, checklist, phân tích assumption/constraint, prompt list (PESTLE, TECOP).

**Định tính**: ma trận **xác suất × tác động**, risk score = P × I, phân loại, mức khẩn cấp.
**Định lượng**: **EMV = xác suất × giá trị tác động**, **decision tree**, **sensitivity (tornado diagram)**, **Monte Carlo simulation**.

**Chiến lược ứng phó** **[Process]**:
- **Threat**: **Escalate** (vượt thẩm quyền → program/portfolio), **Avoid** (loại bỏ nguyên nhân/đổi plan), **Transfer** (chuyển tác động cho bên thứ ba: bảo hiểm, hợp đồng, **không xóa risk**), **Mitigate** (giảm xác suất/tác động), **Accept** (chủ động: dự phòng; bị động: chỉ ghi nhận)
- **Opportunity**: **Escalate, Exploit** (đảm bảo xảy ra), **Share** (chia với bên có năng lực tốt hơn), **Enhance** (tăng xác suất/tác động), **Accept**

**Phân biệt hay nhầm** **[Process]**:
- **Contingency plan**: kế hoạch chạy khi **trigger** xảy ra; **Fallback plan**: dự phòng khi contingency không hiệu quả
- **Workaround**: đối phó **không được lên kế hoạch trước** với risk đã xảy ra (hoặc chưa nhận diện)
- **Residual risk**: còn lại sau khi ứng phó; **Secondary risk**: phát sinh do chính biện pháp ứng phó
- **Risk appetite** (mức sẵn sàng chấp nhận), **tolerance** (mức dao động chấp nhận được), **threshold** (ngưỡng phải hành động)

**Ví dụ EMV**: risk 30% × (−50k) = −15k; cơ hội 20% × (+40k) = +8k.

**Agile**: giảm rủi ro bằng iteration ngắn, **spike**, ưu tiên item rủi ro cao sớm, risk là một phần của backlog.
**Đánh giá lại** risk liên tục (risk reassessment, risk audit).$cheat$, 1015, '[]'::jsonb),

  ($cheat$16. Process — Quality Management$cheat$, $cheat$**Khái niệm** **[Process]**:
- **Quality** (đáp ứng yêu cầu) khác **Grade** (đặc tính/hạng): grade thấp có thể chấp nhận, **quality thấp thì không**
- Chất lượng **được lập kế hoạch vào sản phẩm, không "kiểm tra vào"**; **phòng ngừa rẻ hơn kiểm tra**; chi phí lỗi cao nhất khi lỗi đến tay khách hàng
- **Plan Quality** (chuẩn, metric) → **Manage Quality** (**QA**: audit, phân tích quy trình) → **Control Quality** (**QC**: kiểm tra, đo, test sản phẩm)

**7 công cụ chất lượng cơ bản** **[Process]**: Cause-and-effect (**fishbone / Ishikawa**), Flowchart, **Check sheet**, **Pareto** (80/20, vital few), **Histogram**, **Control chart**, **Scatter diagram**.

**Control chart**:
- **UCL / LCL** (giới hạn kiểm soát, từ dữ liệu quy trình) khác **USL / LSL** (giới hạn đặc tả của khách hàng)
- **Rule of seven**: 7 điểm liên tiếp cùng một phía đường trung bình → có nguyên nhân đặc biệt, cần điều tra
- **Common cause** (ngẫu nhiên, vốn có) vs **special cause** (có nguyên nhân gán được)
- Phân phối chuẩn: ±1σ = 68.27%, ±2σ = 95.45%, ±3σ = 99.73%; **Six Sigma** ≈ 3.4 lỗi/triệu cơ hội

**Khác**:
- **Accuracy** (gần giá trị thật) vs **Precision** (kết quả nhất quán)
- Cải tiến liên tục: **PDCA** (Plan-Do-Check-Act), **Kaizen**, **Six Sigma (DMAIC)**, **Lean**, TQM, JIT
- Kỹ thuật khác: design of experiments, benchmarking, statistical sampling, testing, inspection/audit
- Deliverable lỗi → **defect repair** qua change request; deliverable đạt → vào Validate Scope
- **Agile**: Definition of Done, TDD, continuous integration, retrospective, demo mỗi sprint

**Chi phí chất lượng**: conformance = **prevention** + **appraisal**; nonconformance = **internal failure** + **external failure**.$cheat$, 1016, '[]'::jsonb),

  ($cheat$17. Process — Resource Management, cấu trúc tổ chức & PMO$cheat$, $cheat$**Cấu trúc tổ chức** **[Process]** (quyền của PM giảm dần từ projectized về functional):
- **Functional**: PM ít/không có quyền; nhân sự thuộc trưởng bộ phận chức năng
- **Weak matrix**: PM như điều phối viên, quyền thấp
- **Balanced matrix**: quyền ở mức vừa; PM và functional manager chia sẻ
- **Strong matrix**: PM có quyền khá cao, thường full-time
- **Projectized**: PM có quyền cao nhất, team thuộc dự án
- **Composite**: kết hợp nhiều dạng

**PMO** **[Business]**: **Supportive** (tư vấn, template, mức kiểm soát thấp), **Controlling** (yêu cầu tuân thủ, mức vừa), **Directive** (trực tiếp quản lý dự án, mức cao).

**Quy trình Resource Management** **[Process]**:
- **Plan Resource Management**: vai trò, trách nhiệm, org chart (hierarchical, matrix **RAM/RACI**, text), team charter
- **Estimate Activity Resources**: người, thiết bị, vật tư; **resource breakdown structure**
- **Acquire Resources**: đàm phán, **pre-assignment**, virtual team, tuyển/thuê
- **Develop Team**: training, team building, co-location, recognition, đánh giá hiệu suất team
- **Manage Team**: quản lý xung đột, đánh giá hiệu suất, phản hồi, issue log
- **Control Resources**: kiểm soát nguồn lực vật lý, so sánh sử dụng thực với kế hoạch

**RACI**: **R**esponsible (làm), **A**ccountable (chịu trách nhiệm cuối, chỉ một), **C**onsulted (hỏi ý kiến hai chiều), **I**nformed (được thông báo một chiều).
**Resource histogram**: biểu đồ cột tải nguồn lực theo thời gian (phát hiện quá tải).
**Resource calendar**: thời gian có sẵn của từng nguồn lực.

**Vai trò**: sponsor (tài trợ, charter), PM (dẫn dắt), functional manager (người tài nguyên), CCB (duyệt thay đổi), Product Owner (agile: giá trị & backlog).$cheat$, 1017, '[]'::jsonb),

  ($cheat$18. Process — Procurement & Contracts$cheat$, $cheat$**Quy trình** **[Process]**: Plan Procurement (make-or-buy, strategy) → Conduct Procurement (RFP/bid, chọn nhà cung cấp) → Control Procurement (quản lý hợp đồng, thay đổi, thanh toán) → Close Procurement.

**Make-or-buy**: so sánh chi phí tự làm vs mua/thuê, năng lực, rủi ro, thời gian.

**Tài liệu mời thầu**: **RFI** (xin thông tin), **RFQ** (báo giá, hàng hóa chuẩn), **RFP** (đề xuất giải pháp, đánh giá nhiều tiêu chí), **IFB / ITB** (đấu thầu theo giá). **SOW** mô tả công việc của hợp đồng; **bidder conference** để mọi nhà thầu nghe cùng thông tin; **weighted scoring** để chấm.

**Loại hợp đồng và rủi ro** **[Process]**:
- **Fixed price** — **FFP** (giá cố định, **seller** chịu rủi ro vượt chi phí; cần scope rõ; **ít rủi ro nhất cho buyer**), **FPIF** (có incentive chia sẻ tiết kiệm, có **ceiling price**), **FP-EPA** (điều chỉnh theo lạm phát/hợp đồng dài hạn)
- **Cost-reimbursable** — **CPFF** (cost + fee cố định), **CPIF** (cost + incentive fee theo mục tiêu), **CPAF** (cost + award fee theo đánh giá chủ quan của buyer); **buyer** chịu nhiều rủi ro chi phí hơn
- **Time & Material (T&M)**: tính theo giờ/đơn giá + vật tư; hợp khi scope chưa rõ hoặc bổ sung nhân sự ngắn hạn; nên có ceiling/thời hạn; rủi ro chia sẻ

**Point of Total Assumption (PTA)** (FPIF): mức chi phí mà từ đó seller chịu 100% chi phí vượt:
```
PTA = (Ceiling price − Target price) / Buyer's share ratio + Target cost
Ví dụ: ceiling 130, target price 120, target cost 100, buyer share 80%  → PTA = 10/0.8 + 100 = 112.5
```

**Quản lý hợp đồng**: mọi thay đổi theo **procurement change control**; hợp đồng có hiệu lực pháp lý; PM thường **không** có quyền ký trừ khi được ủy quyền (thuộc procurement/legal).
- **Tranh chấp/claims**: đàm phán → hòa giải (mediation) → trọng tài (arbitration) → kiện tụng (litigation)
- Chọn nhà cung cấp công bằng, minh bạch; tránh xung đột lợi ích; **single/sole source** cần lý do chính đáng

**Đóng hợp đồng**: xác nhận mọi bên hoàn thành nghĩa vụ, lưu hồ sơ, đánh giá hiệu suất seller, lessons learned.$cheat$, 1018, '[]'::jsonb),

  ($cheat$19. Process — Closing$cheat$, $cheat$**Close Project or Phase** **[Process]**:
- Xác nhận deliverable đã **được khách hàng/sponsor chấp nhận** (từ Validate Scope)
- Hoàn tất mọi hợp đồng (Close Procurement), thanh toán, ký xác nhận
- **Chuyển giao (transition)** sản phẩm và tài liệu cho vận hành; huấn luyện người dùng
- **Lessons learned**: tổng kết, lưu vào kho tri thức của tổ chức (OPA); phải làm kể cả khi dự án bị hủy
- **Final report**, lưu trữ hồ sơ dự án (archive)
- **Giải phóng nguồn lực**, giải tán team, ghi nhận công lao
- Bàn giao **benefits management plan** cho đơn vị chịu trách nhiệm hiện thực hóa lợi ích sau dự án

**Đóng sớm / hủy dự án** **[Process]**: lý do hay gặp — business case không còn hợp lý, chi phí vượt ngưỡng, thay đổi chiến lược, công nghệ lỗi thời, rủi ro không chấp nhận được. Vẫn phải: ghi lại lý do, lessons learned, đóng hợp đồng, giải phóng nguồn lực, thông báo stakeholder.

**Phase gate**: điểm quyết định tiếp tục/ sửa/ dừng dựa trên business case và kết quả giai đoạn.

**Agile**: đóng theo release; mỗi sprint kết thúc bằng review & retro; dự án agile kết thúc khi mục tiêu (Product Goal) đạt hoặc không còn giá trị.

**Lưu ý**: hoàn thành deliverable ≠ hiện thực hóa lợi ích — benefits có thể xuất hiện sau khi dự án đóng.$cheat$, 1019, '[]'::jsonb),

  ($cheat$20. Business Environment — Governance & Compliance$cheat$, $cheat$**Governance** **[Business]**:
- **Organizational governance**: khung luật chơi của tổ chức; **project governance**: ai quyết định gì, escalation path, tiêu chí duyệt, **stage gates**, báo cáo định kỳ
- Thành phần: steering committee, sponsor, PMO, CCB, quyền quyết định (decision rights), chính sách và quy trình
- **Define and establish project governance**: thiết lập sớm, rõ ràng, phù hợp quy mô/độ phức tạp, nhất quán với governance của tổ chức/portfolio
- Governance tách khỏi management: governance **quyết định định hướng và giám sát**, management **thực hiện**

**Compliance** **[Business]**:
- Xác định luật, quy định, chuẩn mực, hợp đồng, chính sách áp dụng (vd bảo vệ dữ liệu cá nhân, an toàn, môi trường, tiêu chuẩn ngành ISO)
- Đưa yêu cầu compliance vào **scope, quality, risk, procurement**; phân công người chịu trách nhiệm
- Theo dõi, kiểm tra (**compliance audit**), lưu bằng chứng; vi phạm → báo cáo theo quy định, khắc phục, không che giấu
- Thay đổi luật/quy định trong dự án → đánh giá tác động → change request

**Môi trường dự án** **[Business]**:
- **EEF** (Enterprise Environmental Factors): yếu tố **bên trong/ngoài** tổ chức **không kiểm soát được**: văn hóa, cấu trúc, luật, điều kiện thị trường, hạ tầng, công cụ sẵn có
- **OPA** (Organizational Process Assets): **quy trình, chính sách, template, kho tri thức, lessons learned, dữ liệu lịch sử** — tổ chức **có thể cập nhật**
- Văn hóa tổ chức ảnh hưởng cách giao tiếp, ra quyết định, chấp nhận rủi ro

**Hierarchy**: **Portfolio** (chiến lược, chọn dự án) → **Program** (nhóm dự án liên quan, quản lý chung để đạt lợi ích) → **Project** (kết quả duy nhất, tạm thời); **Operations** (liên tục, lặp lại).$cheat$, 1020, '[]'::jsonb),

  ($cheat$21. Business Environment — Value, Benefits & chọn dự án$cheat$, $cheat$**Giá trị & lợi ích** **[Business]**:
- **Output / deliverable** (sản phẩm giao) → **Outcome** (thay đổi tạo ra) → **Benefit** (giá trị thu được cho tổ chức)
- **Business case**: lý do, phân tích nhu cầu, phương án, đánh giá tài chính, khuyến nghị
- **Benefits management plan**: lợi ích mục tiêu, **owner**, metric/KPI, thời điểm, rủi ro; theo dõi **cả sau khi dự án đóng**
- **Help ensure value-based delivery**: ưu tiên theo giá trị, giao sớm (MVP), liên tục đánh giá lại business case
- Dự án có thể bị dừng nếu giá trị không còn; **không** tiếp tục chỉ vì đã chi nhiều (**sunk cost**)

**Tiêu chí chọn dự án** **[Business]**:
- **NPV** (giá trị hiện tại ròng): dương thì chấp nhận; **cao hơn tốt hơn**
- **IRR**: tỷ suất sinh lợi nội bộ; **cao hơn tốt hơn**
- **Payback period**: thời gian hoàn vốn; **ngắn hơn tốt hơn** (bỏ qua giá trị sau hoàn vốn)
- **BCR** (benefit-cost ratio) **> 1** là có lợi
- **ROI** = lợi ích ròng / chi phí
- **Opportunity cost** = giá trị của phương án tốt nhất bị bỏ qua
- **Present Value**: PV = FV / (1 + r)^n (ví dụ 11,000 sau 1 năm, r = 10% → PV = 10,000)
- Quy luật lợi suất giảm dần (law of diminishing returns), kinh tế quy mô (economies of scale)

**Phương pháp chọn**: benefit measurement (scoring, so sánh, cost-benefit) và mathematical (linear programming). Dự án lớn/chiến lược do **portfolio governance** quyết định.

**Project success vs product success**: success criteria nên thống nhất từ charter (không chỉ time/cost/scope mà còn giá trị, hài lòng stakeholder).$cheat$, 1021, '[]'::jsonb),

  ($cheat$22. Business Environment — Thay đổi tổ chức & môi trường bên ngoài$cheat$, $cheat$**Organizational change management** **[Business]** (khác với change control của dự án):
- **Kotter 8 bước**: tạo cảm giác cấp bách → lập liên minh dẫn dắt → xây tầm nhìn → huy động "đội quân tình nguyện" → loại bỏ rào cản → tạo thắng lợi ngắn hạn → duy trì đà → neo vào văn hóa
- **ADKAR**: Awareness, Desire, Knowledge, Ability, Reinforcement
- **Lewin**: Unfreeze → Change → Refreeze
- **Support organizational change**: đánh giá mức sẵn sàng, xác định người bị ảnh hưởng, truyền thông "vì sao", huấn luyện, **change champions**, xử lý kháng cự, củng cố sau triển khai
- Kháng cự thường do sợ mất quyền lợi, thiếu thông tin, không tin tưởng — lắng nghe và hỗ trợ, không áp đặt

**Môi trường bên ngoài** **[Business]**:
- **PESTLE**: Political, Economic, Social, Technological, Legal, Environmental
- Theo dõi thị trường, đối thủ, quy định, chuỗi cung ứng, công nghệ mới, kỳ vọng xã hội
- Khi môi trường đổi → **đánh giá tác động** lên business case, scope, risk, benefits → cập nhật qua change control hoặc báo governance
- **SWOT** (Strengths, Weaknesses, Opportunities, Threats) dùng cho chiến lược & nhận diện risk

**Dự án toàn cầu/virtual**: múi giờ, ngôn ngữ, văn hóa; cần working agreement và công cụ phù hợp.

**Liên kết chiến lược**: dự án phải phù hợp mục tiêu tổ chức; PM theo dõi sự thay đổi chiến lược để điều chỉnh hoặc đề xuất dừng.$cheat$, 1022, '[]'::jsonb),

  ($cheat$23. AI & Sustainability trong dự án (xu hướng ECO 2026)$cheat$, $cheat$**Ghi chú**: ECO 2026 đưa **AI** và **sustainability** vào các tình huống (scenario), không thành năng lực riêng. Nội dung dưới đây là kiến thức nền để xử lý các tình huống đó.

**AI trong quản lý dự án** **[Business]**:
- Dùng để soạn thảo tài liệu, tóm tắt họp, hỗ trợ ước lượng, phân tích risk, dự báo xu hướng
- PM vẫn **chịu trách nhiệm cuối cùng** — phải **kiểm chứng** kết quả AI (có thể sai/bịa) trước khi dùng
- **Bảo mật dữ liệu**: không đưa thông tin mật của khách hàng/công ty vào công cụ AI công cộng; tuân theo chính sách tổ chức
- Rủi ro: thiên lệch (bias), thiếu minh bạch, vi phạm bản quyền/quyền riêng tư; cần **human in the loop**
- Governance: có chính sách sử dụng AI, ghi nhận việc dùng khi cần, đào tạo team

**Sustainability** **[Business]**:
- **Triple bottom line**: People – Planet – Profit; **ESG**: Environmental, Social, Governance
- Tích hợp vào business case, yêu cầu, risk register, tiêu chí chọn nhà cung cấp, chỉ số thành công
- Xem xét tác động theo vòng đời sản phẩm: năng lượng, vật liệu, rác thải, tác động xã hội, tuân thủ quy định môi trường
- PMBOK Guide 8th có nguyên tắc **"Integrate sustainability"**

**Cách xử lý đáp án**: ưu tiên minh bạch với stakeholder, tuân thủ chính sách/luật, đánh giá tác động trước khi áp dụng công nghệ mới, thử nghiệm nhỏ (pilot), đào tạo người dùng, không bỏ qua yếu tố đạo đức.$cheat$, 1023, '[]'::jsonb),

  ($cheat$24. PMBOK Guide — 6th, 7th, 8th$cheat$, $cheat$**PMBOK 6** (nền cho nhiều thuật ngữ trong đề) **[Chung]**:
- **5 Process Groups**: Initiating, Planning, Executing, Monitoring & Controlling, Closing
- **10 Knowledge Areas**: Integration, Scope, Schedule, Cost, Quality, Resource, Communications, Risk, Procurement, Stakeholder
- 49 processes

**PMBOK 7 (2021)** **[Chung]**:
- **12 principles**: Stewardship, Team, Stakeholders, Value, Systems thinking, Leadership, Tailoring, Quality, Complexity, Risk, Adaptability & resiliency, Change
- **8 performance domains**: Stakeholder, Team, Development Approach & Life Cycle, Planning, Project Work, Delivery, Measurement, Uncertainty
- Chuyển từ "process" sang "principle-based"

**PMBOK 8 (phát hành cuối 2025)** **[Chung]** — theo tài liệu bên thứ ba, nên đối chiếu pmi.org:
- **6 principles**: Adopt a holistic view, Focus on value, Embed quality, Lead accountably, Integrate sustainability, Build empowered teams
- **7 performance domains**: Governance, Scope, Schedule, Finance, Stakeholders, Resources, Risk
- **5 focus areas**: Initiating, Planning, Executing, Monitoring & Controlling, Closing, gắn với khoảng 40 quy trình không áp đặt (non-prescriptive)
- Nhấn mạnh value, AI, vai trò PMO hướng giá trị, procurement được bổ sung hướng dẫn

**Lưu ý**: ECO 2026 được cho là aligned với PMBOK 8 nhưng **vẫn có khác biệt**; thuật ngữ và quy trình trong đề dựa trên ECO, không chỉ riêng một phiên bản PMBOK.

**Tài liệu liên quan**: Agile Practice Guide, Standard for Risk/Program/Portfolio Management, PMI Code of Ethics.$cheat$, 1024, '[{"url": "https://www.pmi.org/standards/pmbok", "label": "PMI – PMBOK Guide (trang chính thức)"}]'::jsonb),

  ($cheat$25. Công thức tổng hợp & ví dụ tính nhanh$cheat$, $cheat$**Công thức cần thuộc** **[Chung]**:
```
Communication channels = n(n − 1) / 2
PERT (beta)      = (O + 4M + P) / 6        Độ lệch chuẩn = (P − O) / 6     Variance = ((P − O)/6)^2
Triangular       = (O + M + P) / 3
EMV              = Probability × Impact
Float            = LS − ES = LF − EF       EF = ES + Duration       LS = LF − Duration
PV (tiền)        = FV / (1 + r)^n
```

**EVM**:
```
CV = EV − AC     SV = EV − PV     CPI = EV / AC     SPI = EV / PV
EAC = BAC/CPI | AC + (BAC − EV) | AC + (BAC − EV)/(CPI × SPI) | AC + bottom-up ETC
ETC = EAC − AC   VAC = BAC − EAC  TCPI(BAC) = (BAC − EV)/(BAC − AC)
```

**Agile / Kanban**:
```
Số sprint còn lại = Backlog còn lại / Velocity
Cycle time (Little's Law) = WIP / Throughput
```

**Hợp đồng**: PTA = (Ceiling − Target price) / Buyer share + Target cost.

**Ví dụ tính**:
- 8 người → 8×7/2 = **28** kênh; thêm 2 người (10) → **45** kênh
- PERT: O 4, M 6, P 14 → (4 + 24 + 14)/6 = **7**; σ = 10/6 ≈ **1.67**; triangular = 24/3 = **8**
- Backlog 240 points, velocity 30 → **8** sprint
- WIP 12, throughput 3/tuần → cycle time **4 tuần**
- EVM: BAC 200k, PV 100k, EV 80k, AC 90k → CPI 0.889, SPI 0.8, EAC (typical) 225k, EAC (atypical) 210k, EAC (CPI×SPI) 258.75k, TCPI 1.09
- Float: ES 5, D 4 → EF 9; LF 12 → float **3**
- EMV: 30% × (−50k) = **−15k**
- PTA: ceiling 130, target price 120, target cost 100, buyer 80% → **112.5**

**Mẹo**: ghi biến vào giấy nháp ngay đầu bài; SPI/CPI dùng **EV** ở tử số; "variance" dùng **EV − thứ gì**; ước lượng EAC chọn công thức đúng với giả định đề nêu.$cheat$, 1025, '[]'::jsonb),

  ($cheat$26. Các cặp thuật ngữ dễ nhầm$cheat$, $cheat$**Tổng hợp cặp dễ nhầm** **[Chung]**:
- **Verify (Control Quality)** vs **Validate (Validate Scope)**: kiểm tra đúng yêu cầu (nội bộ) vs khách hàng chính thức chấp nhận
- **QA** (Manage Quality: kiểm soát quy trình) vs **QC** (Control Quality: kiểm tra sản phẩm)
- **Contingency reserve** (known-unknowns, trong baseline) vs **Management reserve** (unknown-unknowns, ngoài baseline)
- **Crashing** (thêm nguồn lực, tăng cost) vs **Fast tracking** (làm song song, tăng risk)
- **Resource leveling** (có thể đổi critical path) vs **Resource smoothing** (dùng float, giữ critical path)
- **Risk** (chưa xảy ra) vs **Issue** (đã xảy ra)
- **Workaround** (không kế hoạch) vs **Contingency plan** (có kế hoạch, theo trigger) vs **Fallback plan**
- **Residual risk** (còn lại) vs **Secondary risk** (do biện pháp ứng phó gây ra)
- **Assumption** (cho là đúng) vs **Constraint** (giới hạn bắt buộc)
- **Scope creep** (thêm không kiểm soát) vs **Gold plating** (team tự thêm tính năng)
- **Corrective** (đưa hiệu suất về kế hoạch) vs **Preventive** (tránh sai lệch trong tương lai) vs **Defect repair**
- **Deliverable** vs **Outcome** vs **Benefit**
- **Project** vs **Program** vs **Portfolio** vs **Operations**
- **Lead** (gối đầu) vs **Lag** (chờ thêm)
- **Mitigate** (giảm) vs **Avoid** (loại bỏ) vs **Transfer** (chuyển tác động, không xóa)
- **Product backlog** (toàn bộ việc, PO sắp xếp) vs **Sprint backlog** (việc của sprint, Developers sở hữu)
- **Lead time** vs **Cycle time**; **Velocity** (points/sprint) vs **Throughput** (số item/đơn vị thời gian)
- **Burndown** (còn lại) vs **Burnup** (đã xong + scope)
- **RFQ** (giá) vs **RFP** (giải pháp) vs **IFB** (bid theo giá)
- **Governance** (định hướng, giám sát) vs **Management** (thực thi)
- **Leader** vs **Manager**; **Servant leadership** vs **Transformational**
- **Organizational change management** (con người thích nghi) vs **Project change control** (quản lý thay đổi baseline)
- **Efficiency vs effectiveness**; **Accuracy vs precision**
- **Quality vs grade**
- **Sponsor** vs **Customer** vs **Product Owner**$cheat$, 1026, '[]'::jsonb),

  ($cheat$27. Mẫu tình huống: chọn đáp án theo tư duy PMI$cheat$, $cheat$**Mẫu 1 — "What should the PM do FIRST?"** **[Chung]**: chọn bước **phân tích/đánh giá tác động/xác nhận thông tin**, không phải hành động quyết liệt (sa thải, escalate, đổi baseline).

**Mẫu 2 — Thay đổi từ khách hàng/sponsor** **[Process]**: đánh giá tác động → tạo change request → CCB duyệt → mới triển khai. Agile: đưa vào product backlog để Product Owner sắp thứ tự.

**Mẫu 3 — Xung đột trong team** **[People]**: gặp riêng các bên, ưu tiên **Collaborate/Problem-solve**; chỉ Force khi khẩn cấp; PM không bênh phe nào.

**Mẫu 4 — Team thiếu kỹ năng / chậm** **[People]**: tìm nguyên nhân, coaching/training, điều chỉnh phân công; không đổ lỗi, không vội sa thải.

**Mẫu 5 — Risk xảy ra** **[Process]**: xem risk register → thực hiện response đã lên kế hoạch → nếu chưa có thì phân tích, workaround → cập nhật risk register, thông báo stakeholder.

**Mẫu 6 — Hết ngân sách / chậm tiến độ** **[Process]**: phân tích variance (EVM), tìm nguyên nhân, đề xuất nén tiến độ/cắt scope qua change request; không tự ý hy sinh chất lượng.

**Mẫu 7 — Agile**: team **tự tổ chức**; PM/Scrum Master gỡ vướng mắc, coach; không giao việc; yêu cầu mới → backlog; review với khách hàng mỗi sprint; chọn cách giúp team nâng năng lực.

**Mẫu 8 — Đạo đức** **[Business]**: minh bạch, công khai xung đột lợi ích, từ chối quà, báo cáo vi phạm đúng kênh.

**Mẫu 9 — Stakeholder khó** **[People]**: tìm hiểu mối quan tâm, tăng engagement, điều chỉnh communication, không phớt lờ.

**Mẫu 10 — Governance/compliance** **[Business]**: theo chính sách, xác định luật áp dụng, đưa vào kế hoạch, escalate đúng quyền quyết định.

**Mẹo loại đáp án**: bỏ đáp án "bỏ qua quy trình", "làm ngay không phân tích", "đổ lỗi", "che giấu"; bỏ đáp án quá tuyệt đối (always/never); khi nhiều đáp án đúng, chọn cái **chủ động, hợp tác, theo kế hoạch/giá trị**.$cheat$, 1027, '[]'::jsonb),

  ($cheat$28. Viết tắt & thuật ngữ thường gặp$cheat$, $cheat$**Viết tắt PMP** **[Chung]**:
- **EEF** Enterprise Environmental Factors · **OPA** Organizational Process Assets
- **WBS** Work Breakdown Structure · **RBS** Risk (hoặc Resource) Breakdown Structure · **OBS** Organizational Breakdown Structure
- **RAM / RACI** Responsibility Assignment Matrix · **RTM** Requirements Traceability Matrix
- **CCB** Change Control Board · **PMO** Project Management Office
- **PDM** Precedence Diagramming Method · **CPM** Critical Path Method · **PERT** Program Evaluation and Review Technique
- **EVM** Earned Value Management: **PV, EV, AC, BAC, EAC, ETC, VAC, CPI, SPI, TCPI, CV, SV**
- **EMV** Expected Monetary Value · **NPV, IRR, ROI, BCR**
- **SOW** Statement of Work · **RFP/RFQ/RFI/IFB** · **PTA** Point of Total Assumption · **T&M, FFP, FPIF, FP-EPA, CPFF, CPIF, CPAF**
- **DoD** Definition of Done · **MVP** Minimum Viable Product · **WIP** Work In Progress · **TDD** Test-Driven Development · **CI/CD**
- **KPI** Key Performance Indicator · **SLA** Service Level Agreement · **NDA** Non-Disclosure Agreement
- **UCL/LCL** Upper/Lower Control Limit · **USL/LSL** Upper/Lower Specification Limit
- **ECO** Exam Content Outline · **PDU** Professional Development Unit · **PESTLE, SWOT, ESG**

**Thuật ngữ cần nhớ**:
- **Progressive elaboration**: làm rõ dần khi có thêm thông tin
- **Tailoring**: điều chỉnh quy trình cho phù hợp dự án
- **Rolling wave planning**: kế hoạch chi tiết cho phần gần, thô cho phần xa
- **Timebox**: khung thời gian cố định
- **Baseline**: bản kế hoạch đã duyệt, dùng để so sánh
- **Stage gate / phase gate**: điểm duyệt giữa các giai đoạn
- **Tacit vs explicit knowledge**: kiến thức ngầm vs hiển ngôn
- **Single point of failure**, **bus factor**: rủi ro phụ thuộc một người$cheat$, 1028, '[]'::jsonb)
) AS v(title, body, ord, links)
JOIN users u ON u.email = 'thehaohcm@tradesign.io.vn'
JOIN subjects s ON s.slug = 'pmp'
WHERE NOT EXISTS (
    SELECT 1 FROM course_notes cn
    WHERE cn.user_id = u.id AND cn.subject_id = s.id AND cn.title = v.title
);
