// Auto-generated from the 4 PMP case-study stories (pmp_story_1..4.docx).
// Narrative project-management case studies for the PMP "Case Study" page.
export interface PmpConcept { term: string; desc: string }
export interface PmpChapter { num: number; title: string; narrative: string; concepts: PmpConcept[] }
export interface PmpStoryExtra { heading: string; body: string[] }
export interface PmpStory {
  id: number
  name: string
  world: string
  intro: string
  chapters: PmpChapter[]
  extra: PmpStoryExtra[]
}

export const pmpStories: PmpStory[] = [
  {
    "id": 1,
    "name": "The Core Infrastructure",
    "world": "Vũ trụ Predictive / Waterfall",
    "intro": "Bối cảnh tổng quan: Bạn được bổ nhiệm làm Project Manager (PM) cho dự án \"Apex Data Center\" của ngân hàng FinTrust. Mục tiêu là xây dựng hạ tầng máy chủ vật lý mới (Server, Storage, Network, Hệ thống làm mát) và di dời toàn bộ dữ liệu Core Banking từ trung tâm cũ sang trung tâm mới trong vòng 12 tháng với ngân sách $2,000,000. Vì yêu cầu an toàn tài chính nghiêm ngặt và phụ thuộc lớn vào phần cứng vật lý, dự án bắt buộc chạy theo mô hình Predictive (Waterfall). Dưới đây là bản đồ 10 Chương bao trọn toàn bộ quy trình, công thức và bẫy thi của nhánh Predictive:",
    "chapters": [
      {
        "num": 1,
        "title": "Lệnh Khởi Động & Quyền Lực (Initiating)",
        "narrative": "Ban giám đốc FinTrust nhận thấy trung tâm dữ liệu cũ sắp quá tải và vi phạm chuẩn an toàn mới của Ngân hàng Nhà nước. Bộ phận chiến lược đã hoàn thành đánh giá tính khả thi tài chính và kế hoạch thu hồi vốn. Sponsor gọi bạn vào giao nhiệm vụ chính thức.",
        "concepts": [
          {
            "term": "Business Case & Benefits Management Plan",
            "desc": "Tài liệu tiền dự án (Pre-project) chứng minh dự án đáng để đầu tư (sử dụng các chỉ số NPV, IRR, Payback Period). Bẫy thi: PM không tự viết Business Case, nhưng có trách nhiệm đảm bảo dự án đi đúng hướng để đạt được lợi ích đó."
          },
          {
            "term": "Project Charter",
            "desc": "Văn bản khai sinh dự án do Sponsor ký, chính thức cấp thẩm quyền (Authority) cho bạn được sử dụng nguồn lực công ty."
          },
          {
            "term": "Assumption Log",
            "desc": "Sổ ghi chép các giả định (ví dụ: Giả định nhà cung cấp điện lực sẽ kéo xong đường dây cao thế trước tháng 3) và các ràng buộc (Constraints như ngân sách cứng $2M)."
          }
        ]
      },
      {
        "num": 2,
        "title": "Bản Đồ Quyền Lực (Stakeholder Identification)",
        "narrative": "Dự án động chạm đến nhiều bên: Giám đốc Khối CNTT (Sponsor), Đội Vận hành (sợ mất việc/ngại học công nghệ mới), Bộ phận Pháp chế (Compliance), và Nhà thầu cung cấp thiết bị.",
        "concepts": [
          {
            "term": "Stakeholder Register",
            "desc": "Danh sách toàn bộ các bên liên quan."
          },
          {
            "term": "Power/Interest Grid (Ma trận Quyền lực/Quan tâm)",
            "desc": "Phân loại để có chiến lược giao tiếp. Giám đốc CNTT (High Power, High Interest) -> Manage Closely. Đội Vận hành (Low Power, High Interest) -> Keep Informed và tổ chức đào tạo sớm để giảm kháng cự."
          },
          {
            "term": "Stakeholder Engagement Assessment Matrix",
            "desc": "Đánh giá mức độ ủng hộ hiện tại (Unaware, Resistant, Neutral, Supportive, Leading) và mức độ kỳ vọng."
          }
        ]
      },
      {
        "num": 3,
        "title": "Khóa Chặt Phạm Vi (Scope Planning)",
        "narrative": "Bạn tổ chức các buổi họp thu thập yêu cầu. Trưởng phòng Bảo mật đòi lắp thêm hệ thống nhận diện mống mắt ở cửa ra vào, dù ban đầu chỉ yêu cầu thẻ từ. Bạn cần chốt chính xác những gì sẽ làm và không làm.",
        "concepts": [
          {
            "term": "Requirements Traceability Matrix (RTM)",
            "desc": "Bảng truy xuất nguồn gốc từng yêu cầu từ lúc phát sinh đến khi kiểm thử xong (đảm bảo mọi tính năng đều phục vụ mục tiêu kinh doanh)."
          },
          {
            "term": "Project Scope Statement",
            "desc": "Mô tả chi tiết phạm vi và các loại trừ (Exclusions)."
          },
          {
            "term": "WBS (Work Breakdown Structure) & WBS Dictionary",
            "desc": "Chia nhỏ công việc theo cấu trúc cây cho đến cấp thấp nhất gọi là Work Package."
          },
          {
            "term": "Scope Baseline",
            "desc": "Bộ 3 quyền lực gồm (Scope Statement + WBS + WBS Dictionary). Một khi đã chốt, không ai được tự ý thêm bớt."
          }
        ]
      },
      {
        "num": 4,
        "title": "Đường Găng & Tiến Độ (Schedule Planning)",
        "narrative": "Từ các Work Package, team kỹ thuật bóc tách ra các đầu việc chi tiết (Activities), ước lượng thời gian và sắp xếp thứ tự trước sau (ví dụ: Phải lắp xong hệ thống làm mát rồi mới được bật nguồn Server).",
        "concepts": [
          {
            "term": "PERT / Three-Point Estimating",
            "desc": "Công thức ước lượng thời gian khi có độ bất định: ."
          },
          {
            "term": "Critical Path Method (CPM)",
            "desc": "Đường dài nhất xuyên suốt mạng lưới dự án, quyết định tổng thời gian hoàn thành. Các công việc trên đường găng có Total Float = 0 (trễ 1 ngày là cả dự án trễ 1 ngày)."
          },
          {
            "term": "Schedule Baseline",
            "desc": "Lịch trình chuẩn đã được phê duyệt."
          }
        ]
      },
      {
        "num": 5,
        "title": "Ngân Sách & Quỹ Dự Phòng (Cost Planning)",
        "narrative": "Bạn tổng hợp chi phí mua sắm Server, thuê nhân công, bản quyền phần mềm và trình duyệt ngân sách.",
        "concepts": [
          {
            "term": "Cost Baseline",
            "desc": "Bao gồm chi phí ước tính của các Work Package + Contingency Reserve (Quỹ dự phòng cho các rủi ro đã biết - Known Unknowns, do PM tự quyền sử dụng)."
          },
          {
            "term": "Project Budget",
            "desc": "Bằng Cost Baseline + Management Reserve (Quỹ dự phòng cho rủi ro chưa lường trước - Unknown Unknowns, muốn dùng phải xin phép Sponsor)."
          }
        ]
      },
      {
        "num": 6,
        "title": "Chất Lượng, Nguồn Lực & Mua Sắm (Quality, Resource, Procurement)",
        "narrative": "Bạn tuyển dụng đội ngũ kỹ sư, ký hợp đồng mua lô máy chủ trị giá $800,000 từ hãng Dell/Cisco, và đặt ra tiêu chuẩn vận hành (Uptime 99.99%).",
        "concepts": [
          {
            "term": "RACI Chart",
            "desc": "Ma trận phân quyền (Responsible, Accountable, Consulted, Informed). Quy tắc vàng: Mỗi đầu việc chỉ có duy nhất 1 người A (Accountable)."
          },
          {
            "term": "Quality Metrics & Cost of Quality (CoQ)",
            "desc": "Chi phí phòng ngừa lỗi (đào tạo, kiểm tra thiết kế) luôn rẻ hơn chi phí sửa lỗi khi đã bàn giao (Failure Cost)."
          },
          {
            "term": "",
            "desc": "Các loại hợp đồng (Contracts):"
          },
          {
            "term": "Fixed-Price (FFP)",
            "desc": "Phạm vi rõ ràng -> Rủi ro thuộc về nhà thầu (Seller)."
          },
          {
            "term": "Time & Material (T&M)",
            "desc": "Thuê chuyên gia theo giờ khi chưa rõ phạm vi."
          },
          {
            "term": "Cost-Reimbursable (CPFF/CPIF)",
            "desc": "Thanh toán chi phí thực tế + phí thưởng -> Rủi ro thuộc về bên mua (Buyer - chính là bạn)."
          }
        ]
      },
      {
        "num": 7,
        "title": "Phòng Thủ Rủi Ro (Risk Management)",
        "narrative": "Đội ngũ ngồi lại liệt kê các thảm họa có thể xảy ra: Đứt cáp quang biển làm chậm giao hàng Server, kỹ sư trưởng nghỉ việc, hoặc mất điện diện rộng khi đang chuyển dữ liệu.",
        "concepts": [
          {
            "term": "Risk Register",
            "desc": "Danh sách rủi ro, xác suất xảy ra và mức độ tác động (Probability & Impact Matrix)."
          },
          {
            "term": "",
            "desc": "Chiến lược ứng phó rủi ro tiêu cực (Threats):"
          },
          {
            "term": "Avoid",
            "desc": "Đổi phương án để triệt tiêu rủi ro (ví dụ: dùng đường truyền vệ tinh dự phòng thay vì chỉ phụ thuộc cáp quang)."
          },
          {
            "term": "Transfer",
            "desc": "Chuyển giao cho bên thứ 3 (mua bảo hiểm thiết bị vận chuyển)."
          },
          {
            "term": "Mitigate",
            "desc": "Giảm xác suất hoặc tác động (thuê thêm 1 kỹ sư phó để sẵn sàng thay thế kỹ sư trưởng)."
          },
          {
            "term": "Accept",
            "desc": "Chấp nhận và trích quỹ Contingency nếu xảy ra."
          }
        ]
      },
      {
        "num": 8,
        "title": "Biến Cố & Quy Trình Thay Đổi (Executing & Integrated Change Control)",
        "narrative": "Tháng thứ 6, dự án đang thi công. Giám đốc Bảo mật (ở Chương 3) lại yêu cầu nâng cấp chuẩn mã hóa phần cứng. Đồng thời, một lô Switch mạng giao về bị lỗi kỹ thuật.",
        "concepts": [
          {
            "term": "Issue Log",
            "desc": "Khi một rủi ro thực sự xảy ra (lô Switch bị lỗi), nó không còn là Risk mà biến thành Issue và phải ghi vào Issue Log để giải quyết ngay."
          },
          {
            "term": "Quality Assurance (QA) vs Quality Control (QC)",
            "desc": "QA là kiểm tra quy trình làm việc có đúng chuẩn không (Audit); QC là kiểm tra sản phẩm đầu ra có đạt thông số không (Inspection)."
          },
          {
            "term": "Integrated Change Control (CCB)",
            "desc": "Với yêu cầu nâng cấp mã hóa của Giám đốc Bảo mật:"
          },
          {
            "term": "",
            "desc": "Ghi nhận vào Change Log."
          },
          {
            "term": "",
            "desc": "Phân tích tác động toàn diện (lên Scope, Schedule, Cost, Risk)."
          },
          {
            "term": "",
            "desc": "Trình lên Change Control Board (CCB) phê duyệt."
          },
          {
            "term": "",
            "desc": "Nếu được duyệt -> Cập nhật lại các Baseline và thông báo cho các bên. (Tuyệt đối không tự ý làm hay từ chối thẳng thừng trước khi phân tích)."
          }
        ]
      },
      {
        "num": 9,
        "title": "Đo Lường Sức Khỏe Dự Án (Monitoring & Controlling - EVM)",
        "narrative": "Tháng thứ 8, bạn kiểm tra sức khỏe dự án bằng phương pháp Earned Value Management (EVM). Theo kế hoạch đến lúc này phải tiêu hết $1,000,000 (PV), thực tế bạn mới hoàn thành lượng công việc trị giá $800,000 (EV), nhưng đã chi thực tế tới $1,000,000 (AC). Công thức & Bẫy thi: Schedule Performance Index:  ( Trễ tiến độ). Cost Performance Index:  ( Vượt ngân sách). Giải pháp nén tiến độ (Schedule Compression): Crashing: Đổ thêm tiền/nhân lực vào các công việc trên đường găng -> Làm tăng chi phí (tệ hơn cho CPI đang 0.8). Fast Tracking: Làm song song các việc đáng lẽ phải làm tuần tự -> Không tốn thêm tiền ngay nhưng làm tăng rủi ro phải làm lại (Rework).",
        "concepts": []
      },
      {
        "num": 10,
        "title": "Bàn Giao & Khép Lại (Closing)",
        "narrative": "Tháng thứ 12, toàn bộ hệ thống vận hành trơn tru sau khi đã xử lý xong các điểm nghẽn. Bạn tiến hành bàn giao cho đội Vận hành của ngân hàng.",
        "concepts": [
          {
            "term": "Validate Scope vs Close Project",
            "desc": "Validate Scope là bước khách hàng/Sponsor nghiệm thu chính thức sản phẩm (Formal Acceptance). Sau khi có biên bản nghiệm thu mới sang bước Close Project."
          },
          {
            "term": "",
            "desc": "Trình tự đóng dự án chuẩn PMP (Bắt buộc nhớ đúng thứ tự):"
          },
          {
            "term": "",
            "desc": "Nghiệm thu sản phẩm (Validate Scope)."
          },
          {
            "term": "",
            "desc": "Thanh lý tất cả hợp đồng với nhà thầu (Close Procurements)."
          },
          {
            "term": "",
            "desc": "Hoàn thiện báo cáo tổng kết (Final Report) và cập nhật Lessons Learned Register vào kho tri thức chung của công ty (OPA - Organizational Process Assets)."
          },
          {
            "term": "",
            "desc": "Giải phóng đội ngũ nhân sự (Release Team)."
          },
          {
            "term": "",
            "desc": "Tổ chức tiệc ăn mừng (Celebration)."
          },
          {
            "term": "",
            "desc": "Bạn thấy luồng đi 10 chương của Story 1 này thế nào? Nếu đã nắm được nhịp điệu này, tôi có thể tiếp tục viết Story 2 (Vũ trụ Agile/Scrum) hoặc đặt ra 3 câu hỏi tình huống bẫy dựa trên chính dự án Apex Data Center này để bạn test thử phản xạ ngay."
          }
        ]
      }
    ],
    "extra": []
  },
  {
    "id": 2,
    "name": "The Cloud SaaS Product",
    "world": "Vũ trụ Agile / Scrum",
    "intro": "Bối cảnh tổng quan: Bạn là Scrum Master / Agile Project Manager của dự án \"SalonAI Pro\" tại công ty công nghệ BeautyTech. Ban giám đốc muốn tung ra một nền tảng SaaS giúp các chuỗi salon tóc tự động hóa việc đặt lịch, nhắc hẹn qua Zalo và quản lý doanh thu. Vì thị trường làm đẹp thay đổi liên tục và chưa ai chắc chắn tính năng nào sẽ giữ chân được chủ salon, công ty quyết định chạy dự án theo mô hình Agile (Scrum) với mỗi Sprint kéo dài 2 tuần. Dưới đây là câu chuyện 8 Chương diễn tả trọn vẹn nhịp đập từ lúc sản phẩm chỉ là một ý tưởng mơ hồ cho đến khi ra mắt thị trường và tự tiến hóa:",
    "chapters": [
      {
        "num": 1,
        "title": "Khởi Động Bằng Tầm Nhìn, Không Phải Bản Đặc Tả (Product Vision & MVP)",
        "narrative": "Khác với các dự án truyền thống phải mất 2 tháng viết tài liệu yêu cầu dày cộp, CEO của BeautyTech chỉ đưa cho bạn và Product Owner (Minh Anh) một bài toán kinh doanh: \"Làm sao giúp 1,000 chủ salon giảm 50% tỷ lệ khách bùng kèo (No-show) trong 6 tháng tới?\". Hai bạn cùng nhau chốt phiên bản MVP (Minimum Viable Product) tinh gọn nhất chỉ gồm 3 tính năng cốt lõi để ra mắt sớm, tạm gác lại ý tưởng \"AI thử màu tóc thực tế ảo\" hào nhoáng nhưng tốn thời gian.",
        "concepts": [
          {
            "term": "Product Vision",
            "desc": "Kim chỉ nam định hướng giá trị cho cả đội, thay thế cho Project Scope Statement cứng nhắc."
          },
          {
            "term": "MVP (Minimum Viable Product)",
            "desc": "Sản phẩm khả dụng tối thiểu giúp kiểm chứng giả thuyết thị trường nhanh nhất với chi phí thấp nhất (Fail Fast, Learn Faster)."
          },
          {
            "term": "Value-Driven Delivery",
            "desc": "Tập trung bàn giao giá trị kinh doanh sớm thay vì bám sát một kế hoạch cố định."
          }
        ]
      },
      {
        "num": 2,
        "title": "Xây Dựng Đội Ngũ Tự Tổ Chức & Luật Chơi Chung (Team Charter & DoD)",
        "narrative": "Bạn tập hợp một nhóm 6 người gồm đủ các kỹ năng: 3 Fullstack Dev, 1 AI Engineer, 1 QA và 1 UI/UX Designer. Thay vì phân công ai làm gì như một người sếp truyền thống, bạn tổ chức một buổi họp khởi động để cả nhóm tự thảo luận cách phối hợp, giờ họp cố định và tiêu chuẩn chất lượng chung cho mọi dòng code.",
        "concepts": [
          {
            "term": "Cross-functional & Self-organizing Team",
            "desc": "Nhóm đa chức năng có đủ năng lực tự hoàn thành sản phẩm từ đầu đến cuối và tự quyết định cách thức thực hiện công việc."
          },
          {
            "term": "Team Charter (Social Contract)",
            "desc": "Bản cam kết do chính các thành viên tự viết ra (ví dụ: giờ họp Daily, quy tắc review code), giúp tạo ra Psychological Safety (Sự an toàn tâm lý) để mọi người dám nói thẳng khi gặp lỗi."
          },
          {
            "term": "Definition of Done (DoD)",
            "desc": "Bộ tiêu chuẩn chất lượng bắt buộc mọi tính năng phải đạt được (ví dụ: đã pass Unit Test, không còn bug nghiêm trọng, đã deploy lên Staging) mới được coi là hoàn thành."
          }
        ]
      },
      {
        "num": 3,
        "title": "Xếp Hàng Giá Trị & Ước Lượng Tương Đối (Backlog Refinement & Planning Poker)",
        "narrative": "Product Owner (Minh Anh) viết các nhu cầu của chủ salon thành các User Stories và sắp xếp thứ tự ưu tiên từ cao xuống thấp. Trước khi bắt đầu Sprint, cả nhóm ngồi lại làm rõ từng yêu cầu và dùng bộ bài Planning Poker để đánh giá độ phức tạp của từng tính năng mà không bị ảnh hưởng bởi ý kiến của người làm lâu năm nhất.",
        "concepts": [
          {
            "term": "Product Backlog & Prioritization (MoSCoW)",
            "desc": "Danh sách công việc động luôn được cập nhật. PO là người duy nhất có quyền quyết định thứ tự ưu tiên dựa trên giá trị kinh doanh."
          },
          {
            "term": "Backlog Refinement (Grooming)",
            "desc": "Hoạt động diễn ra xuyên suốt để chia nhỏ các tính năng lớn (Epics) thành các User Story đủ nhỏ để làm gọn trong 1 Sprint."
          },
          {
            "term": "Story Points & Relative Estimation",
            "desc": "Ước lượng công việc dựa trên độ phức tạp và rủi ro tương đối (dãy Fibonacci: 1, 2, 3, 5, 8, 13...) thay vì đếm số giờ làm việc cứng nhắc."
          }
        ]
      },
      {
        "num": 4,
        "title": "Cam Kết Nhịp Đập 2 Tuần (Sprint Planning)",
        "narrative": "Sáng thứ Hai đầu tuần, cả nhóm bước vào buổi họp Sprint Planning. Dựa trên tốc độ trung bình của 2 Sprint trước (khoảng 24 điểm), cả nhóm tự chọn các User Story từ đỉnh Product Backlog kéo sang Sprint Backlog và cùng chốt một mục tiêu chung cho 2 tuần tới: \"Hoàn thiện luồng đặt lịch và nhắc hẹn tự động để kịp cho 20 salon chạy thử\".",
        "concepts": [
          {
            "term": "Timeboxing",
            "desc": "Mọi sự kiện trong Scrum đều bị giới hạn thời gian nghiêm ngặt để tránh họp hành lan man (Sprint Planning tối đa 4 tiếng cho Sprint 2 tuần)."
          },
          {
            "term": "Velocity",
            "desc": "Thước đo tốc độ hoàn thành công việc trung bình của nhóm qua các Sprint, dùng để dự báo năng lực cho các Sprint tiếp theo (không dùng để so sánh giữa các team với nhau)."
          },
          {
            "term": "Sprint Goal",
            "desc": "Mục tiêu cốt lõi giữ cho cả nhóm đi đúng hướng ngay cả khi các chi tiết kỹ thuật bên trong có sự điều chỉnh."
          }
        ]
      },
      {
        "num": 5,
        "title": "Cơn Bão Giữa Tuần & Vai Trò Người Bảo Vệ (Servant Leadership)",
        "narrative": "Ngày thứ 4 của Sprint 3, Giám đốc Marketing đi gặp khách hàng về và nhắn tin riêng cho bạn Senior Dev yêu cầu làm gấp tính năng \"Xuất báo cáo PDF gửi kế toán\" ngay trong tuần này. Bạn Dev nể nang định nhận làm thêm. Ngay lập tức, bạn (Scrum Master) phát hiện ra và đứng ra xử lý: bạn gặp riêng Giám đốc Marketing giải thích quy tắc vận hành, đồng thời hướng dẫn vị Giám đốc làm việc với PO Minh Anh để xếp tính năng đó vào Sprint sau.",
        "concepts": [
          {
            "term": "Servant Leadership (Lãnh đạo phục vụ)",
            "desc": "Scrum Master không ra lệnh cho team mà đóng vai trò như một tấm khiên bảo vệ nhóm khỏi các tác nhân gây nhiễu loạn từ bên ngoài (Protecting the Team)."
          },
          {
            "term": "Sprint Backlog Protection",
            "desc": "Một khi Sprint đã bắt đầu, không ai (kể cả sếp lớn) được tự ý nhét thêm việc làm phá vỡ Sprint Goal và gây phân mảnh sự tập trung (Context Switching)."
          }
        ]
      },
      {
        "num": 6,
        "title": "Dọn Dẹp Chướng Ngại Vật & Biểu Đồ Đi Ngang (Daily Scrum & Impediments)",
        "narrative": "Mỗi sáng 9h15, cả nhóm đứng quây quần trước bảng Kanban đúng 15 phút để đồng bộ công việc. Đến ngày thứ 6, bạn nhìn vào Burndown Chart và thấy đường biểu diễn điểm số bị đi ngang 2 ngày liên tiếp. Hỏi ra mới biết tài khoản Zalo OA thử nghiệm bị khóa hạn mức và server Staging liên tục sập mỗi khi chạy test tự động. Bạn lập tức ghi nhận vào sổ theo dõi, tự mình đi làm việc với đối tác Zalo và đội DevOps để mở khóa ngay trong buổi chiều.",
        "concepts": [
          {
            "term": "Daily Scrum (Stand-up)",
            "desc": "Buổi họp 15 phút để nhóm tự đồng bộ kế hoạch trong ngày và phát hiện sớm các điểm nghẽn (NOT một buổi họp báo cáo tiến độ cho Scrum Master)."
          },
          {
            "term": "Information Radiators (Burndown Chart)",
            "desc": "Bảng biểu trực quan đặt ở nơi dễ thấy nhất để bất kỳ ai nhìn vào cũng biết sức khỏe thực tế của Sprint."
          },
          {
            "term": "Impediment Removal",
            "desc": "Nhiệm vụ sống còn của Scrum Master là dọn sạch mọi rào cản hành chính, kỹ thuật để nhóm tập trung tối đa vào việc tạo ra sản phẩm."
          }
        ]
      },
      {
        "num": 7,
        "title": "Trình Diễn Sản Phẩm Thật & Đón Nhận Thay Đổi (Sprint Review)",
        "narrative": "Chiều thứ Sáu cuối Sprint 3, cả nhóm không làm slide PowerPoint mà mở thẳng phần mềm chạy thật trên điện thoại để demo cho CEO và 5 chủ salon mời tới. Các chủ salon rất thích tính năng đặt lịch nhưng góp ý: \"Khách bên chị hay đi theo nhóm 3-4 người, app hiện tại chỉ cho đặt từng người một thì bất tiện quá\". PO Minh Anh lập tức ghi nhận ý tưởng này thành một User Story mới và đẩy ngay lên đầu danh sách ưu tiên cho Sprint 4.",
        "concepts": [
          {
            "term": "Sprint Review (Demo)",
            "desc": "Buổi kiểm tra phần mềm chạy được (Working Software) cùng các bên liên quan để thu thập phản hồi thực tế từ thị trường."
          },
          {
            "term": "Embracing Change",
            "desc": "Trong Agile, việc khách hàng thay đổi yêu cầu ở giai đoạn muộn được hoan nghênh vì nó giúp sản phẩm sát với nhu cầu thực tế hơn, thay vì bị coi là \"Scope Creep\" tiêu cực như ở Predictive."
          }
        ]
      },
      {
        "num": 8,
        "title": "Nhìn Lại Nội Bộ & Cải Tiến Liên Tục (Sprint Retrospective)",
        "narrative": "Ngay sau buổi Demo, 6 thành viên trong nhóm đóng cửa họp riêng với nhau (không có CEO, không có khách hàng). Bạn dẫn dắt buổi họp theo mô hình Start - Stop - Continue. Cả nhóm thẳng thắn nhìn nhận việc để dồn toàn bộ việc kiểm thử (QA) vào 2 ngày cuối Sprint khiến mọi người quá tải, và thống nhất từ Sprint 4 sẽ áp dụng kỹ thuật Swarming (cả nhóm cùng xúm vào dứt điểm từng tính năng một thay vì ai làm việc nấy).",
        "concepts": [
          {
            "term": "Sprint Retrospective",
            "desc": "Cuộc họp quan trọng nhất để cải tiến quy trình làm việc và sự phối hợp giữa con người. Bắt buộc phải giữ không gian riêng tư cho team để đảm bảo tính trung thực."
          },
          {
            "term": "Continuous Improvement (Kaizen)",
            "desc": "Cam kết đưa ít nhất 1-2 hành động cải tiến cụ thể vào áp dụng ngay ở Sprint tiếp theo."
          },
          {
            "term": "Swarming / Pair Programming",
            "desc": "Kỹ thuật cộng tác giúp tối ưu hóa luồng công việc (Flow), giảm lượng công việc dang dở (WIP - Work In Progress) và chia sẻ kiến thức chéo trong nhóm."
          },
          {
            "term": "",
            "desc": "Với câu chuyện 8 chương này kết hợp cùng bộ 10 file hồ sơ Agile ở trên, bạn đã có trọn vẹn cả mạch truyện diễn biến lẫn tài liệu thực tế của Vũ trụ Agile. Khi nào sẵn sàng, chúng ta sẽ tiếp tục dựng câu chuyện cho Story 3 (Vũ trụ Hybrid & Quản lý Con người/Stakeholders đa quốc gia) nhé!"
          }
        ]
      }
    ],
    "extra": []
  },
  {
    "id": 3,
    "name": "The Enterprise Rollout",
    "world": "Vũ trụ Hybrid & Con Người / Đa Văn Hóa",
    "intro": "Bối cảnh tổng quan: Bạn là Project Manager dẫn dắt dự án \"GlobalPOS Nexus\" cho tập đoàn bán lẻ thời trang AuraRetail. Mục tiêu là nâng cấp toàn bộ hệ thống thanh toán tại quầy (POS) và đồng bộ kho vận cho 300 cửa hàng tại 4 quốc gia (Việt Nam, Singapore, Nhật Bản, Úc) trong vòng 9 tháng. Vì dự án vừa phải đặt hàng sản xuất hàng ngàn máy quét mã vạch, máy in hóa đơn theo hợp đồng cứng, vừa phải phát triển phần mềm POS cập nhật liên tục theo thói quen tiêu dùng từng nước, bạn quyết định áp dụng mô hình Hybrid — kết hợp sự kỷ luật của Predictive cho phần cứng và sự linh hoạt của Agile cho phần mềm. Đặc biệt, đây là bài kiểm tra nặng ký nhất về People Domain (quản lý đội ngũ phân tán, xung đột văn hóa và chuyển giao vận hành).",
    "chapters": [
      {
        "num": 1,
        "title": "Nghệ Thuật \"May Đo\" Quy Trình (Tailoring & Hybrid Design)",
        "narrative": "Khi mới nhận dự án, Giám đốc Tài chính muốn áp dụng 100% quy trình Waterfall truyền thống để kiểm soát chặt ngân sách, trong khi Giám đốc Sản phẩm lại đòi chạy 100% Scrum cho kịp xu hướng. Bạn đứng ra phân tích đặc thù dự án và thiết kế một mô hình lai (Hybrid): Giai đoạn mua sắm, vận chuyển, lắp đặt phần cứng tại 300 cửa hàng sẽ chạy theo Predictive (có WBS và lịch trình cố định từng đợt); còn đội phát triển ứng dụng POS trên nền tảng Cloud sẽ chạy theo Agile (Sprint 2 tuần).",
        "concepts": [
          {
            "term": "Tailoring (May đo)",
            "desc": "Không có một phương pháp nào hoàn hảo cho mọi dự án; PM phải tự thiết kế quy trình phù hợp nhất với bối cảnh tổ chức, độ phức tạp và rủi ro."
          },
          {
            "term": "Hybrid Lifecycle",
            "desc": "Sử dụng Predictive cho các hạng mục có độ chắc chắn cao/vật lý (Hardware rollout) và Agile cho các hạng mục cần thử nghiệm/thay đổi liên tục (Software UI/UX)."
          }
        ]
      },
      {
        "num": 2,
        "title": "Vượt Rào Cản Địa Lý & Văn Hóa (Virtual Teams & Communication)",
        "narrative": "Đội ngũ của bạn phân tán ở 4 múi giờ khác nhau: Team Dev ở Việt Nam, Đội Pháp chế/Thuế ở Nhật Bản và Úc, còn Sponsor ngồi ở trụ sở chính Singapore. Tuần đầu tiên, team Việt Nam gửi email hỏi ý kiến kỹ thuật cho phía Nhật Bản nhưng 3 ngày không thấy hồi âm, dẫn đến hiểu lầm rằng phía Nhật thiếu hợp tác. Bạn lập tức can thiệp, tìm hiểu văn hóa làm việc đồng thuận (Ringi) của Nhật và thiết lập lại quy tắc giao tiếp chung.",
        "concepts": [
          {
            "term": "Virtual Teams & Time-zone Management",
            "desc": "Thiết lập \"Giờ vàng giao tiếp\" (Core Overlapping Hours) từ 10:00 đến 12:00 sáng (giờ VN) để các bên họp trực tuyến."
          },
          {
            "term": "Cultural Awareness & Communication Management Plan",
            "desc": "Quy định rõ kênh giao tiếp nào dùng cho việc gì (Slack cho trao đổi nhanh, Email cho xác nhận chính thức, Video Call cho giải quyết mâu thuẫn). Bẫy thi: Luôn ưu tiên Video Call thay vì chỉ gửi văn bản khi làm việc với đội nhóm từ xa."
          }
        ]
      },
      {
        "num": 3,
        "title": "Dập Tắt Xung Đột Nội Bộ (Conflict Management & Tuckman Ladder)",
        "narrative": "Tháng thứ 2, đội ngũ rơi vào giai đoạn căng thẳng tột độ: Tech Lead phần mềm (ở VN) và Trưởng nhóm Hạ tầng (ở Singapore) cãi nhau nảy lửa trên nhóm chat vì lỗi đồng bộ dữ liệu khi mạng chập chờn — mỗi bên đều đổ lỗi cho bên kia. Thay vì chọn phe hay lờ đi, bạn triệu tập ngay một cuộc họp Video Call 3 bên, yêu cầu hai người tập trung vào giải pháp kỹ thuật thay vì công kích cá nhân.",
        "concepts": [
          {
            "term": "Tuckman Ladder",
            "desc": "Nhóm đang ở giai đoạn Storming (Bão tố — nhiều xung đột nhất). Nhiệm vụ của PM là dẫn dắt nhóm vượt qua giai đoạn này để tiến tới Norming (Ổn định) và Performing (Hiệu suất cao)."
          },
          {
            "term": "",
            "desc": "5 Kỹ thuật giải quyết xung đột (Conflict Resolution):"
          },
          {
            "term": "Collaborate / Problem-Solve (Win-Win)",
            "desc": "Kỹ thuật tốt nhất trong PMP — cùng ngồi lại phân tích dữ liệu để tìm ra giải pháp tối ưu cho cả hai."
          },
          {
            "term": "Tránh dùng",
            "desc": "Force/Direct (áp đặt quyền lực) hay Withdraw/Avoid (né tránh)."
          }
        ]
      },
      {
        "num": 4,
        "title": "Sức Mạnh Của Trí Tuệ Cảm Xúc & Huấn Luyện (EQ, Coaching vs. Mentoring)",
        "narrative": "Bạn nhận thấy một bạn Business Analyst (BA) trẻ người Úc rất giỏi nghiệp vụ bán lẻ nhưng lại lúng túng khi viết User Story theo chuẩn Agile, khiến team Dev thường xuyên hiểu sai. Cùng lúc đó, một trưởng nhóm triển khai tại cửa hàng có dấu hiệu kiệt sức (Burnout) vì áp lực lịch trình. Bạn dành thời gian kèm cặp bạn BA, đồng thời lắng nghe và điều chỉnh lại khối lượng công việc cho trưởng nhóm.",
        "concepts": [
          {
            "term": "Emotional Intelligence (EQ)",
            "desc": "Khả năng nhận diện và quản lý cảm xúc của bản thân cũng như thấu cảm với đội ngũ (Self-awareness, Self-management, Social awareness, Social skills)."
          },
          {
            "term": "",
            "desc": "Phân biệt Đào tạo (Bẫy thi kinh điển):"
          },
          {
            "term": "Training",
            "desc": "Dạy kỹ năng cứng mới cho cả nhóm (ví dụ: khóa học sử dụng công cụ Jira)."
          },
          {
            "term": "Mentoring",
            "desc": "Người đi trước có kinh nghiệm chia sẻ lời khuyên định hướng nghề nghiệp lâu dài."
          },
          {
            "term": "Coaching",
            "desc": "Đặt câu hỏi gợi mở để giúp cá nhân tự tìm ra cách cải thiện hiệu suất công việc hiện tại (bạn áp dụng cho bạn BA)."
          }
        ]
      },
      {
        "num": 5,
        "title": "Sóng Gió Tuân Thủ Quốc Tế (Compliance & Regulatory Changes)",
        "narrative": "Đang ở Sprint 6 (tháng thứ 4), Chính phủ Nhật Bản ban hành quy định mới về định dạng hóa đơn điện tử (Qualified Invoice System) có hiệu lực ngay trong quý tới. Nếu phần mềm POS không cập nhật kịp, 80 cửa hàng tại Nhật sẽ bị phạt nặng và không thể khai trương. Bạn lập tức đánh giá đây là rủi ro tuân thủ ở mức độ cao nhất (Compliance).",
        "concepts": [
          {
            "term": "Compliance Management (Business Environment Domain)",
            "desc": "Các yêu cầu về luật pháp, thuế, bảo mật dữ liệu (như GDPR/hóa đơn điện tử) luôn có độ ưu tiên cao nhất, bắt buộc phải đáp ứng để dự án hợp pháp."
          },
          {
            "term": "Phối hợp Hybrid",
            "desc": "Với nhánh Agile, Product Owner lập tức đẩy các tính năng hóa đơn Nhật Bản lên đỉnh Product Backlog cho Sprint 7; với nhánh Predictive, bạn cập nhật Risk Register và điều chỉnh lịch trình mở cửa hàng tại Nhật lùi lại 2 tuần (vẫn nằm trong khoảng dự phòng)."
          }
        ]
      },
      {
        "num": 6,
        "title": "Quản Lý Các Bên Liên Quan Khó Tính (Stakeholder Resistance)",
        "narrative": "Khi bắt đầu triển khai thử nghiệm (Pilot) tại 10 cửa hàng đầu tiên ở Việt Nam và Singapore, các Cửa hàng trưởng phản ứng dữ dội: họ chê giao diện POS mới rắc rối và lén quay lại dùng phần mềm cũ để tính tiền cho nhanh. Thay vì báo cáo lên cấp trên để kỷ luật họ, bạn trực tiếp xuống cửa hàng quan sát thao tác thực tế của thu ngân trong 2 ngày.",
        "concepts": [
          {
            "term": "Active Listening & Empathy",
            "desc": "Lắng nghe nguyên nhân gốc rễ của sự kháng cự (hóa ra nút \"Áp dụng mã giảm giá\" nằm quá sâu qua 4 bước bấm, gây ùn tắc giờ cao điểm)."
          },
          {
            "term": "Feedback Loop",
            "desc": "Mang ngay phản hồi thực tế đó về buổi Sprint Review để team UI/UX tối giản thao tác xuống còn 1 chạm, biến những Cửa hàng trưởng từ trạng thái Resistant thành Supportive (những người ủng hộ nhiệt tình nhất)."
          }
        ]
      },
      {
        "num": 7,
        "title": "Vận Hành & Tối Ưu Luồng Triển Khai (Kanban & Bottlenecks)",
        "narrative": "Sang tháng thứ 7, phần mềm đã ổn định, dự án bước vào giai đoạn cao điểm lắp đặt đồng loạt cho 290 cửa hàng còn lại ở 4 nước. Lúc này bạn không dùng Scrum nữa mà chuyển đội triển khai sang dùng bảng Kanban để theo dõi tiến độ từng cửa hàng qua các cột: Khảo sát  Giao thiết bị  Cài đặt phần mềm  Đào tạo nhân viên  Go-Live. Bạn phát hiện cột Cài đặt phần mềm bị ùn ứ tới 45 cửa hàng do thiếu kỹ sư cấu hình từ xa.",
        "concepts": [
          {
            "term": "Kanban & Flow-based Agile",
            "desc": "Phù hợp cho các công việc mang tính lặp lại, vận hành liên tục."
          },
          {
            "term": "WIP Limits (Work In Progress Limits)",
            "desc": "Bạn đặt giới hạn tối đa chỉ 15 cửa hàng được nằm ở cột Cài đặt phần mềm cùng lúc, điều động thêm nhân sự từ cột Khảo sát sang hỗ trợ để giải tỏa nút thắt cổ chai (Bottleneck), giúp rút ngắn Cycle Time toàn hệ thống."
          }
        ]
      },
      {
        "num": 8,
        "title": "Chuyển Giao Vận Hành & Đo Lường Giá Trị (Transition & Operational Readiness)",
        "narrative": "Cuối tháng thứ 9, toàn bộ 300 cửa hàng tại 4 quốc gia đã vận hành trơn tru trên hệ thống GlobalPOS Nexus. Trước khi giải tán đội dự án, bạn tổ chức bàn giao toàn bộ tài liệu vận hành, quy trình xử lý sự cố (SOP) cho bộ phận IT Helpdesk toàn cầu của AuraRetail và thiết lập hệ thống đo lường doanh thu/tốc độ thanh toán sau khi dự án khép lại.",
        "concepts": [
          {
            "term": "Operational Readiness & Knowledge Transfer",
            "desc": "Đảm bảo đội ngũ vận hành (Operations) đủ năng lực tiếp nhận sản phẩm mà không làm gián đoạn kinh doanh."
          },
          {
            "term": "Benefits Realization Tracking",
            "desc": "Dự án kết thúc nhưng việc đo lường giá trị kinh doanh (tốc độ thanh toán tăng 35%, giảm 20% chi phí tồn kho) vẫn tiếp tục được theo dõi bởi Business Owner trong 6-12 tháng tiếp theo."
          }
        ]
      }
    ],
    "extra": [
      {
        "heading": "Phần 2: Cấu Trúc Cây Thư Mục & Hồ Sơ Mẫu (GlobalPOS_Nexus_Hybrid/)",
        "body": [
          "Plaintext",
          "GlobalPOS_Nexus_Hybrid/├── 01_Governance_&_Tailoring/│   ├── 01_Hybrid_Project_Management_Plan_(Tailoring_Matrix).md│   └── 02_Team_Communication_&_Cultural_Charter.md├── 02_People_&_Stakeholder_Management/│   ├── 01_Global_Stakeholder_Engagement_Matrix.md│   └── 02_Conflict_&_Coaching_Log.md├── 03_Hybrid_Execution_&_Compliance/│   ├── 01_Compliance_Register_(Tax_&_Privacy_Laws).md│   └── 02_Store_Rollout_Kanban_Board_(WIP_Limits).md└── 04_Transition_&_Handover/    └── 01_Operational_Readiness_Checklist_&_Handover_Signoff.md",
          "Nội dung Trích Yếu Các Hồ Sơ Đặc Trưng Của Hybrid:",
          "1. File 01_Hybrid_Project_Management_Plan_(Tailoring_Matrix).md",
          "Hạng mục công việc",
          "Phương pháp áp dụng",
          "Lý do lựa chọn (Tailoring Rationale)",
          "Công cụ quản lý",
          "Sản xuất & Nhập khẩu phần cứng POS",
          "Predictive",
          "Thông số kỹ thuật cố định, phụ thuộc chuỗi cung ứng và thủ tục hải quan 4 nước.",
          "WBS, Gantt Chart, Hợp đồng FFP",
          "Phát triển phần mềm POS & Tích hợp thanh toán",
          "Agile (Scrum)",
          "Giao diện và phương thức thanh toán (QR, Apple Pay, PayPay Nhật) cần thử nghiệm và chỉnh sửa liên tục.",
          "Product Backlog, Sprint 2 tuần, Jira",
          "Triển khai lắp đặt tại 300 cửa hàng",
          "Kanban (Flow)",
          "Quy trình lặp đi lặp lại cho từng cửa hàng, cần kiểm soát luồng công việc và nút thắt cổ chai.",
          "Kanban Board, WIP Limits, Cycle Time",
          "2. File 02_Team_Communication_&_Cultural_Charter.md",
          "Core Overlapping Hours: 10:00 – 12:00 (GMT+7 / Việt Nam)  11:00 – 13:00 (Singapore)  12:00 – 14:00 (Nhật Bản)  13:00 – 15:00 (Úc). Mọi cuộc họp đa quốc gia chỉ được đặt trong khung giờ này.",
          "Quy tắc giao tiếp đa văn hóa:",
          "Với team Nhật Bản: Gửi tài liệu đọc trước tối thiểu 24h trước cuộc họp để họ có thời gian thảo luận nội bộ và đồng thuận.",
          "Khi phát sinh bất đồng kỹ thuật quá 2 lượt tin nhắn qua lại: Bắt buộc chuyển sang Video Call ngay lập tức (nguyên tắc Face-to-Face/Rich Communication).",
          "3. File 01_Compliance_Register_(Tax_&_Privacy_Laws).md",
          "ID",
          "Quốc gia",
          "Quy định pháp lý",
          "Tác động đến dự án",
          "Hành động xử lý",
          "CMP-01",
          "Nhật Bản",
          "Qualified Invoice System (Hóa đơn chuẩn thuế)",
          "Phải in thêm mã đăng ký thuế và tách riêng thuế suất 8% / 10% trên biên lai POS.",
          "Đưa vào Sprint 7 với độ ưu tiên cao nhất; lùi lịch Go-live tại Nhật 2 tuần.",
          "CMP-02",
          "Úc & Singapore",
          "Luật bảo vệ dữ liệu cá nhân (PDPA / Privacy Act)",
          "Mã hóa toàn bộ thông tin thẻ và số điện thoại khách hàng thành viên trên Cloud.",
          "Thực hiện Security Penetration Test bắt buộc trước khi nghiệm thu.",
          "4. File 02_Store_Rollout_Kanban_Board_(WIP_Limits).md",
          "Giai đoạn (Cột)",
          "1. Khảo sát mặt bằng",
          "2. Giao phần cứng",
          "3. Cài đặt & Cấu hình (WIP Limit: 15)",
          "4. Đào tạo thu ngân",
          "5. Go-Live (Done)",
          "Số cửa hàng hiện tại",
          "30",
          "25",
          "15 / 15 (ĐẠT TRẦN)",
          "12",
          "218",
          "Quy tắc vận hành (Pull System)",
          "Làm theo lịch",
          "Đẩy hàng khi xong cột 1",
          "Ngừng nhận thêm cửa hàng mới cho đến khi giải phóng bớt cửa hàng sang cột 4.",
          "Kéo từ cột 3 sang",
          "Hoàn tất bàn giao",
          "Vậy là bạn đã có trong tay 3 Vũ trụ lớn nhất của PMP:",
          "Predictive (Apex Data Center): Kỷ luật, quy trình, công thức EVM, hợp đồng, CCB.",
          "Agile (SalonAI Pro): Linh hoạt, giá trị, Sprint, Servant Leadership, DoD.",
          "Hybrid & People (GlobalPOS Nexus): May đo quy trình, đa văn hóa, giải quyết xung đột, Kanban và tuân thủ pháp lý.",
          "Chúng ta chỉ còn mảnh ghép cuối cùng là Story 4 (Giải cứu dự án khủng hoảng & Đạo đức nghề nghiệp PMI) — nơi tập trung những câu hỏi \"khó nhằn\" nhất về kinh doanh và xử lý tình huống khi mọi thứ đổ vỡ. Bạn có muốn hoàn thiện nốt Story 4 này không?"
        ]
      }
    ]
  },
  {
    "id": 4,
    "name": "The Rescue Mission",
    "world": "Dự Án Khủng Hoảng, Môi Trường Kinh Doanh & Đạo Đức PMI",
    "intro": "Bối cảnh tổng quan: Bạn được Ban giám đốc tập đoàn năng lượng GreenGrid điều động khẩn cấp để tiếp quản và giải cứu dự án \"SmartMeter Grid\" từ một PM cũ vừa xin nghỉ việc. Dự án đã chạy được 6 tháng trên tổng kế hoạch 10 tháng, tiêu tốn 80% ngân sách ($1,600,000 / $2,000,000) nhưng mới chỉ hoàn thành 40% khối lượng công việc. Tinh thần đội ngũ rã rời, nhà thầu chính dọa kiện vì chậm thanh toán, và tệ hơn: thị trường công nghệ vừa xuất hiện một chuẩn kết nối mới khiến giải pháp hiện tại có nguy cơ lỗi thời ngay khi ra mắt. Đây là vũ trụ tập trung các câu hỏi \"khó nhằn\" nhất của Business Environment Domain, Xử lý khủng hoảng, Đạo đức nghề nghiệp (PMI Code of Ethics) và Đóng dự án sớm (Early Termination).",
    "chapters": [
      {
        "num": 1,
        "title": "Tiếp Quản Hiện Trường Đổ Nát (Project Takeover & Assessment)",
        "narrative": "Ngày đầu tiên nhận bàn giao, Sponsor yêu cầu bạn lập tức sa thải 2 kỹ sư bị cho là làm việc chậm chạp và cắt bớt tính năng để kịp tiến độ. Thay vì hành động cảm tính ngay lập tức, bạn dành 1 tuần đầu tiên để rà soát toàn bộ hồ sơ dự án, kiểm tra Project Charter, Issue Log, Risk Register và nói chuyện riêng 1-1 với từng thành viên để tìm ra nguyên nhân gốc rễ.",
        "concepts": [
          {
            "term": "PMP Golden Rule #1 (Assess Before Acting)",
            "desc": "Khi tiếp quản dự án gặp sự cố hoặc đối mặt với biến cố mới, PM không bao giờ hành động bộc phát (sa thải người, đổi lịch, xin thêm tiền) trước khi thu thập dữ liệu và đánh giá toàn diện tác động."
          },
          {
            "term": "Root Cause Analysis (Fishbone Diagram / 5 Whys)",
            "desc": "Bạn phát hiện nguyên nhân thực sự không phải do 2 kỹ sư lười biếng, mà do PM cũ nhận quá nhiều yêu cầu ngoài luồng (Scope Creep) mà không qua phê duyệt, khiến team bị quá tải."
          }
        ]
      },
      {
        "num": 2,
        "title": "Lập Lại Trật Tự & Chặn Đứng Scope Creep (Re-baselining)",
        "narrative": "Bạn tổ chức cuộc họp khẩn với Sponsor và các Stakeholders chủ chốt. Bạn trình bày thẳng thắn bức tranh tài chính thực tế (, ) và chỉ ra rằng nếu giữ nguyên toàn bộ các tính năng phát sinh ngoài luồng, dự án sẽ cần tới $3.2 triệu và trễ hạn 6 tháng. Bạn đề xuất đóng băng các thay đổi không thiết yếu và thiết lập lại các đường cơ sở (Re-baselining).",
        "concepts": [
          {
            "term": "Transparency & Honesty (PMI Code of Ethics)",
            "desc": "Báo cáo trung thực tình trạng tồi tệ của dự án cho Sponsor, tuyệt đối không che giấu số liệu xấu."
          },
          {
            "term": "Re-baselining",
            "desc": "Khi các đường cơ sở cũ (Scope, Schedule, Cost Baseline) không còn thực tế do biến cố lớn, PM phải trình bày phương án điều chỉnh và xin phê duyệt chính thức từ Change Control Board (CCB) / Sponsor để tạo ra một bộ Baseline mới khả thi hơn."
          }
        ]
      },
      {
        "num": 3,
        "title": "Bài Toán Ra Quyết Định Dưới Áp Lực Rủi Ro (Decision Tree & EMV)",
        "narrative": "Hệ thống đo điện thông minh đang gặp lỗi truyền dữ liệu ở vùng sóng yếu. Đội kỹ thuật đưa ra 2 giải pháp: Phương án A: Tự viết lại giao thức truyền tin (Chi phí $100,000; 70% thành công, nhưng nếu thất bại sẽ mất thêm $200,000 để sửa lỗi). Phương án B: Mua bản quyền module truyền tin công nghiệp từ hãng thứ 3 (Chi phí $180,000; 95% thành công, nếu thất bại chỉ mất $40,000 phí tích hợp lại). Bạn dùng phương pháp định lượng để chọn phương án tối ưu về mặt tài chính.",
        "concepts": [
          {
            "term": "Quantitative Risk Analysis & EMV (Expected Monetary Value)",
            "desc": "Tính giá trị kỳ vọng bằng tiền ():"
          },
          {
            "term": "Chi phí kỳ vọng Phương án A",
            "desc": "$\\$100,000 + (0.30 \\times \\$200,000) = \\mathbf{\\$160,000}$."
          },
          {
            "term": "Chi phí kỳ vọng Phương án B",
            "desc": "$\\$180,000 + (0.05 \\times \\$40,000) = \\mathbf{\\$182,000}$ (nhưng độ rủi ro thất bại chỉ 5%)."
          },
          {
            "term": "",
            "desc": "Bạn trình bày bảng phân tích này để Sponsor quyết định dựa trên Risk Appetite (Khẩu vị rủi ro) của công ty (vì dự án đang khủng hoảng, Sponsor chọn Phương án B để đảm bảo an toàn 95% thay vì đánh cược 30% rủi ro)."
          }
        ]
      },
      {
        "num": 4,
        "title": "Giải Quyết Tranh Chấp Hợp Đồng Nhà Thầu (Procurement Claims & ADR)",
        "narrative": "Nhà thầu cung cấp thiết bị cảm biến gửi thông báo ngừng giao hàng và dọa kiện ra tòa vì cho rằng GreenGrid chậm thanh toán đợt 2, trong khi bộ phận Kỹ thuật của GreenGrid lại cho rằng nhà thầu giao thiếu tài liệu kiểm thử nên chưa nghiệm thu. Thay vì đưa nhau ra tòa án tốn kém, bạn chủ động mời Giám đốc nhà thầu ngồi lại đàm phán dựa trên điều khoản hợp đồng.",
        "concepts": [
          {
            "term": "Claims Administration & ADR (Alternative Dispute Resolution)",
            "desc": "Khi có tranh chấp hợp đồng, thứ tự ưu tiên giải quyết trong PMP luôn là: 1. Negotiation (Đàm phán trực tiếp)  2. Mediation (Hòa giải)  3. Arbitration (Trọng tài)  4. Litigation (Kiện tụng ra tòa - giải pháp cuối cùng)."
          },
          {
            "term": "",
            "desc": "Bạn phát hiện hợp đồng cũ viết mập mờ về tiêu chí nghiệm thu đợt 2, nên hai bên thống nhất ký phụ lục làm rõ và giải ngân ngay 70% để nhà thầu tiếp tục giao hàng."
          }
        ]
      },
      {
        "num": 5,
        "title": "Cạm Bẫy Đạo Đức Nghề Nghiệp (PMI Code of Ethics & Professional Conduct)",
        "narrative": "Trước ngày chạy thử nghiệm (Pilot) với khách hàng lớn, một kỹ sư QA báo cáo với bạn rằng phần mềm vẫn còn một lỗ hổng bảo mật có thể làm lộ dữ liệu tiêu thụ điện của người dùng. Cùng lúc đó, Phó Tổng Giám đốc Kinh doanh gọi riêng cho bạn, gợi ý bạn \"tạm thời lờ lỗ hổng đó đi để kịp ký hợp đồng, sau này vá sau\" và hứa sẽ thưởng nóng cho bạn một khoản hoa hồng riêng.",
        "concepts": [
          {
            "term": "PMI Code of Ethics (4 Giá trị cốt lõi",
            "desc": "Responsibility, Respect, Fairness, Honesty):"
          },
          {
            "term": "",
            "desc": "Tuyệt đối từ chối khoản thưởng riêng (xung đột lợi ích / hối lộ - vi phạm Fairness & Honesty)."
          },
          {
            "term": "",
            "desc": "Tuyệt đối không che giấu lỗi bảo mật nghiêm trọng (vi phạm Responsibility & Honesty)."
          },
          {
            "term": "",
            "desc": "Bạn lập tức báo cáo sự thật lên Sponsor và đề xuất phương án vá lỗi khẩn cấp trong 48 giờ trước khi chạy Pilot. (Bẫy thi: Trong mọi tình huống vi phạm đạo đức hoặc pháp luật, PM luôn chọn sự trung thực và tuân thủ, bất kể áp lực từ cấp trên)."
          }
        ]
      },
      {
        "num": 6,
        "title": "Cú Sốc Môi Trường Kinh Doanh (Business Environment & Organizational Change)",
        "narrative": "Tháng thứ 8, khi dự án vừa đi vào ổn định trở lại, tập đoàn GreenGrid bất ngờ công bố sáp nhập với một công ty năng lượng tái tạo lớn và chuyển hướng chiến lược toàn tập đoàn sang mảng điện mặt trời áp mái, đồng thời cắt giảm 40% ngân sách cho mảng lưới điện truyền thống (chính là dự án SmartMeter Grid của bạn).",
        "concepts": [
          {
            "term": "",
            "desc": "Enterprise Environmental Factors (EEFs) & Organizational Change: Sự thay đổi chiến lược công ty hoặc M&A tác động trực tiếp đến tính khả thi của dự án."
          },
          {
            "term": "Benefits Realization Assessment",
            "desc": "PM phải cùng Sponsor đánh giá lại xem Business Case ban đầu của dự án còn hợp lệ trong bối cảnh chiến lược mới hay không."
          }
        ]
      },
      {
        "num": 7,
        "title": "Quyết Định Dừng Dự Án Sớm (Early Project Termination)",
        "narrative": "Sau khi đánh giá lại cùng Sponsor và Hội đồng quản trị, tất cả nhận thấy công nghệ của dự án SmartMeter Grid không còn phù hợp với hạ tầng điện mặt trời mới sau sáp nhập; nếu cố làm tiếp sẽ lãng phí thêm $600,000 mà không mang lại giá trị kinh doanh. Sponsor ra quyết định hủy bỏ dự án ngay lập tức (Early Termination).",
        "concepts": [
          {
            "term": "Project Termination Criteria",
            "desc": "Một dự án có thể kết thúc vì hoàn thành mục tiêu, hoặc bị hủy giữa chừng khi mục tiêu không còn khả thi, hết nguồn lực, hoặc nhu cầu kinh doanh không còn tồn tại."
          },
          {
            "term": "Bẫy thi PMP cực quan trọng",
            "desc": "Khi dự án bị hủy giữa chừng, PM không được phép thu dọn đồ đạc bỏ về ngay lập tức! PM vẫn phải thực hiện đầy đủ quy trình Close Project or Phase."
          }
        ]
      },
      {
        "num": 8,
        "title": "Khép Lại Trong Danh Dự (Formal Closure of a Terminated Project)",
        "narrative": "Trong 2 tuần cuối cùng, bạn dẫn dắt đội ngũ thực hiện quy trình đóng dự án một cách bài bản: ghi nhận chính xác mức độ hoàn thành của từng hạng mục tại thời điểm dừng, thanh lý hợp đồng với các nhà thầu theo điều khoản chấm dứt sớm, tổng hợp toàn bộ bài học xương máu từ thất bại này và gửi lời cảm ơn chân thành đến đội ngũ trước khi điều chuyển họ sang dự án mới.",
        "concepts": [
          {
            "term": "",
            "desc": "Closing a Terminated Project:"
          },
          {
            "term": "",
            "desc": "Ghi nhận lý do dự án bị hủy và mức độ hoàn thành hiện tại vào Final Report."
          },
          {
            "term": "",
            "desc": "Bàn giao các thành phần đã làm xong (nếu có thể tái sử dụng cho dự án điện mặt trời mới)."
          },
          {
            "term": "",
            "desc": "Thanh lý tất cả hợp đồng mua sắm (Close Procurements)."
          },
          {
            "term": "",
            "desc": "Cập nhật Lessons Learned Register vào kho tri thức chung (OPA) để các dự án sau không lặp lại vết xe đổ."
          },
          {
            "term": "",
            "desc": "Giải phóng nhân sự (Release Team)."
          }
        ]
      }
    ],
    "extra": [
      {
        "heading": "Phần 2: Cấu Trúc Cây Thư Mục & Bộ Hồ Sơ Giải Cứu (SmartMeter_Grid_Rescue/)",
        "body": [
          "Plaintext",
          "SmartMeter_Grid_Rescue/├── 01_Takeover_&_Diagnostic/│   ├── 01_Project_Health_Audit_&_Root_Cause_Analysis.md│   └── 02_Re-baseline_Proposal_(Scope_Schedule_Cost).md├── 02_Risk_&_Contract_Recovery/│   ├── 01_Decision_Tree_EMV_Analysis.md│   └── 02_Vendor_Claim_&_Settlement_Agreement.md├── 03_Ethics_&_Business_Alignment/│   ├── 01_Ethics_&_Security_Incident_Escalation_Memo.md│   └── 02_Business_Case_Re-evaluation_(Post-Merger).md└── 04_Early_Termination_&_Closure/    └── 01_Early_Termination_Final_Report_&_Lessons_Learned.md",
          "Nội Dung Chi Tiết Từng File:",
          "1. File 01_Project_Health_Audit_&_Root_Cause_Analysis.md",
          "Thời điểm đánh giá: Tuần 25 (Tháng 6/10).",
          "Chỉ số sức khỏe tại thời điểm tiếp quản:",
          "$BAC = \\$2,000,000$ | $PV = \\$1,200,000$ | $EV = \\$800,000$ | $AC = \\$1,600,000$.",
          "| .",
          "Phân tích nguyên nhân gốc rễ (5 Whys):",
          "Tại sao dự án trễ và bội chi nặng? Vì đội kỹ thuật phải làm thêm 14 tính năng phụ ngoài hợp đồng gốc.",
          "Tại sao lại làm các tính năng phụ này? Vì các trưởng phòng kinh doanh yêu cầu trực tiếp cho Dev làm.",
          "Tại sao Dev lại tự ý làm? Vì PM cũ không thiết lập quy trình Integrated Change Control (CCB) và không bảo vệ nhóm khỏi Scope Creep.",
          "2. File 02_Re-baseline_Proposal_(Scope_Schedule_Cost).md",
          "Mục đích: Xin phê duyệt bộ Baseline mới thực tế hơn sau khi loại bỏ các tính năng ngoài luồng.",
          "Nội dung đề xuất trình CCB & Sponsor:",
          "Cắt bỏ 10 tính năng phụ chưa làm (giữ lại 4 tính năng cốt lõi phục vụ vận hành cơ bản).",
          "Xin bổ sung $400,000 từ quỹ dự phòng chiến lược của công ty và gia hạn thêm 2 tháng (tổng thời gian 12 tháng).",
          "Trạng thái: APPROVED bởi Sponsor và Hội đồng CCB.",
          "3. File 01_Decision_Tree_EMV_Analysis.md",
          "Bài toán: Chọn phương án khắc phục lỗi truyền dữ liệu vùng sóng yếu.",
          "Phương án",
          "Chi phí cố định",
          "Xác suất thành công",
          "Xác suất thất bại",
          "Chi phí phát sinh nếu thất bại",
          "Công thức tính EMV (Tổng chi phí kỳ vọng)",
          "Kết luận lựa chọn",
          "PA-A: Tự viết lại giao thức",
          "$100,000",
          "70%",
          "30%",
          "$200,000",
          "$\\$100k + (0.3 \\times \\$200k) = \\mathbf{\\$160,000}$",
          "Rẻ hơn về kỳ vọng toán học nhưng rủi ro trễ hạn cao (30%).",
          "PA-B: Mua module hãng thứ 3",
          "$180,000",
          "95%",
          "5%",
          "$40,000",
          "$\\$180k + (0.05 \\times \\$40k) = \\mathbf{\\$182,000}$",
          "ĐƯỢC CHỌN vì phù hợp với khẩu vị rủi ro thấp (Low Risk Appetite) của dự án đang khủng hoảng.",
          "4. File 02_Vendor_Claim_&_Settlement_Agreement.md",
          "Bên khiếu nại: Công ty Cảm biến VietSensor (yêu cầu bồi thường $50,000 do chậm thanh toán).",
          "Phương pháp giải quyết: Đàm phán trực tiếp (Negotiation - ADR).",
          "Thỏa thuận đạt được: Hai bên ký phụ lục làm rõ checklist nghiệm thu; GreenGrid giải ngân ngay đợt 2 ($250,000) trong vòng 3 ngày làm việc; VietSensor rút đơn khiếu nại và cam kết giao đủ 5,000 bộ cảm biến còn lại đúng hạn.",
          "5. File 01_Ethics_&_Security_Incident_Escalation_Memo.md",
          "Người gửi: Hao Nguyen (Project Manager) | Người nhận: Project Sponsor & Chief Information Security Officer (CISO).",
          "Nội dung báo cáo:",
          "Phát hiện lỗ hổng bảo mật truyền dữ liệu chưa mã hóa trên module đo điện (mức độ: Critical).",
          "Tuân thủ quy tắc ứng xử nghề nghiệp (PMI Code of Ethics - Honesty & Responsibility), PM từ chối đề nghị \"bỏ qua lỗi để ký hợp đồng sớm\" và từ chối mọi khoản thưởng cá nhân không minh bạch.",
          "Đề xuất lùi lịch chạy thử nghiệm (Pilot) 48 giờ để đội kỹ thuật vá triệt để lỗ hổng trước khi đưa dữ liệu khách hàng vào hệ thống.",
          "6. File 02_Business_Case_Re-evaluation_(Post-Merger).md",
          "Bối cảnh: Tập đoàn GreenGrid sáp nhập và chuyển trọng tâm sang điện mặt trời áp mái (tháng 8).",
          "Kết quả đánh giá lại tính khả thi:",
          "Công nghệ đo lưới điện một chiều của dự án SmartMeter Grid không hỗ trợ đo hai chiều (mua/bán điện mặt trời).",
          "Chi phí nâng cấp lại từ đầu ước tính thêm $900,000  NPV chuyển sang âm (-$350,000).",
          "Khuyến nghị chung của PM & Sponsor: Dừng dự án sớm (Early Termination) để bảo toàn nguồn vốn cho dự án lưới điện hai chiều thế hệ mới.",
          "7. File 01_Early_Termination_Final_Report_&_Lessons_Learned.md",
          "Tình trạng đóng dự án: Đóng sớm do thay đổi chiến lược tổ chức (Early Termination due to Organizational Restructuring / M&A).",
          "Tài sản bàn giao lại cho tổ chức (Reusable Assets):",
          "Toàn bộ mã nguồn giao diện Dashboard quản trị và module bảo mật đã hoàn thiện (đạt 65% khối lượng) được đóng gói chuyển giao cho đội dự án Điện mặt trời mới.",
          "Thanh lý hợp đồng: Đã ký biên bản chấm dứt hợp đồng trước hạn và thanh toán đầy đủ phần khối lượng đã nghiệm thu cho nhà thầu VietSensor.",
          "Bài học kinh nghiệm (Lessons Learned - Lưu vào OPA):",
          "Về Quản trị phạm vi: Phải thiết lập quy trình Change Control ngay từ ngày đầu tiên; mọi yêu cầu không qua CCB đều là Scope Creep.",
          "Về Đạo đức & Minh bạch: Sự trung thực khi báo cáo số liệu xấu () giúp Sponsor đưa ra quyết định cắt lỗ kịp thời, tiết kiệm cho tập đoàn hàng trăm ngàn đô la.",
          "Giải phóng nhân sự: Toàn bộ 14 thành viên được đánh giá hiệu suất công bằng và điều chuyển chính thức sang dự án mới.",
          "Tổng Kết Hệ Thống 4 Vũ Trụ PMP Của Bạn",
          "Bạn đã sở hữu trọn bộ 4 Câu chuyện + 36 File Hồ sơ Mẫu bao phủ 100% ma trận đề thi PMP:",
          "Story 1 (Apex Data Center): Predictive / Waterfall — Quy trình, Đường găng, EVM, CCB, Hợp đồng.",
          "Story 2 (SalonAI Pro): Agile / Scrum — Tầm nhìn, MVP, User Story, Velocity, Servant Leadership, DoD.",
          "Story 3 (GlobalPOS Nexus): Hybrid & People — May đo (Tailoring), Đa văn hóa, Giải quyết xung đột, Kanban, Tuân thủ.",
          "Story 4 (SmartMeter Grid): Rescue & Business — Đánh giá trước khi hành động, EMV, Đạo đức PMI, Hủy dự án sớm.",
          "Bước tiếp theo hiệu quả nhất để biến toàn bộ kiến thức này thành phản xạ thi cử là luyện giải các câu hỏi tình huống (Mock Questions) lồng ghép trực tiếp vào 4 dự án này. Bạn muốn tôi ra thử cho bạn một bộ 5 câu hỏi tình huống chuẩn đề thi PMP (có đủ 4 đáp án A, B, C, D gây nhiễu) lấy bối cảnh từ chính 4 câu chuyện trên để bạn thực hành ngay không?"
        ]
      }
    ]
  }
]
