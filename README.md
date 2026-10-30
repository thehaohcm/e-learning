# E-Learning: English + PMP (web + thiết bị ESP32)

Hệ thống học theo kiểu **lặp lại ngắt quãng (FSRS)** dùng chung cho nhiều môn:
tiếng Anh (TOEIC/IELTS, phrasal verbs, writing) và PMI PMP (thuật ngữ, công thức,
câu hỏi luyện tập). Có website và API cho thiết bị học để bàn ESP32-S3 (ES3C28P).

## Chạy nhanh

```bash
cp .env.example .env        # điền DB_*, JWT_SECRET (openssl rand -hex 32), ADMIN_EMAILS
docker compose up --build -d
# mở http://localhost:8085 → Create account
```

- PostgreSQL chạy luôn trong docker-compose (service `db`, dữ liệu ở volume `db_data`).
  Lần đầu khởi động, container tự tạo user/database theo `DB_USER`, `DB_PASSWORD`, `DB_NAME`.
- API chờ DB sẵn sàng rồi tự chạy migration (tạo bảng + dữ liệu mẫu). Không cần `migrate_db.py` nữa.
- Không bao giờ chạy `docker compose down -v` (xoá volume = mất dữ liệu). Backup: `make backup`.
- Chạy dev không Docker: `make api-dev` (cổng 8000) và `make web-dev` (cổng 8080, tự proxy `/api`).
- Test: `make test`.

## Domain + Cloudflare (cổng 80/443)

Container `proxy` (nginx) nhận cổng **80** và **443** rồi chuyển vào web; web chuyển `/api` vào API.
Cổng 8085 không còn mở ra ngoài.

1. Cloudflare → DNS: thêm bản ghi **A** `learn` → IP server, bật **Proxied** (đám mây cam).
2. Cloudflare → SSL/TLS → **Origin Server** → *Create Certificate* (RSA, 15 năm). Lưu 2 file vào
   `proxy/certs/origin.pem` và `proxy/certs/origin.key` (trên server hoặc trên máy rồi deploy).
3. Cloudflare → SSL/TLS → Overview: chọn **Full (strict)**. Bật *Always Use HTTPS*.
4. Mở firewall cổng 80, 443 trên VPS. Deploy lại.

Chưa có chứng chỉ thì proxy chỉ chạy HTTP cổng 80 (dùng được với chế độ *Flexible* của Cloudflare,
nhưng đoạn Cloudflare → server không mã hoá). Proxy lấy IP thật của người dùng từ `CF-Connecting-IP`
(chỉ tin các dải IP của Cloudflare), nên giới hạn đăng nhập theo IP hoạt động đúng.

## Nâng cấp từ bản cũ

1. Deploy bản mới. Migration `0005_import_legacy` tự **copy** từ vựng, phrasal verbs
   và tiến độ học từ các bảng cũ sang bảng mới. Bảng cũ được giữ nguyên.
2. Tài khoản cũ chưa có mật khẩu: vào **Create account** với đúng email cũ để đặt mật khẩu,
   tiến độ học vẫn còn.
3. Kiểm tra xong thì có thể xoá bảng cũ: `psql ... -f docs/drop_legacy_tables.sql` (nhớ backup trước).
4. Giá trị FSRS cũ bị sai công thức, nên thẻ đã học được chuyển thành trạng thái "review"
   với độ khó trung tính. Ngày ôn vẫn giữ nguyên, FSRS tự hiệu chỉnh sau 1–2 lần ôn.

## Cấu trúc

```
api/                Go (Gin + GORM + PostgreSQL)
  cmd/server        main: kết nối DB → chạy migration → HTTP server
  migrations/       *.sql, nhúng vào binary, chạy theo thứ tự tên file
  internal/srs      bọc thư viện FSRS chính thức (go-fsrs v3)
  internal/study    hàng đợi ôn, chấm điểm, quiz, thống kê (dùng chung cho web và thiết bị)
  internal/content  định dạng payload của từng loại thẻ
  internal/auth     bcrypt, JWT, token thiết bị
  internal/ai       DeepSeek: sinh thẻ, chấm writing
  internal/http     router + handlers
web/                Vue 3 + Vite, phục vụ bằng nginx (proxy /api → api)
docs/               SQL xoá bảng cũ
```

### Mô hình dữ liệu

`subjects` → `decks` → `cards` (`kind` + `payload` JSONB) và `user_cards` (trạng thái FSRS
theo từng user) + `review_logs` (lịch sử mỗi lần trả lời).

| kind      | front            | payload                                              |
|-----------|------------------|------------------------------------------------------|
| `vocab`   | từ               | `ipa, pos, senses[{level, meaning_vi, meaning_en, example_en, example_vi, usage_note}]` |
| `phrasal` | "give up"        | `verb, particle, separable, senses[...]`             |
| `concept` | thuật ngữ        | `back, note_vi, example`                             |
| `mcq`     | câu hỏi          | `choices[], answer (index), explanation`             |

Thêm một môn mới (ví dụ AWS, JLPT) chỉ cần tạo subject + deck (API admin hoặc migration SQL),
không phải sửa code. Đề PMP: chỉ dùng câu hỏi tự soạn hoặc do AI sinh. Đề thật của PMI
được bảo mật và có bản quyền.

### Quyền

- Đăng nhập bằng email + mật khẩu (bcrypt), phiên là JWT 30 ngày.
- Admin (email trong `ADMIN_EMAILS`, hoặc tài khoản đầu tiên nếu chưa có admin) được thêm,
  sửa, xoá thẻ và sinh thẻ bằng AI.
- Sau khi tạo xong tài khoản, đặt `OPEN_SIGNUP=false` để người lạ không đăng ký được.

## API cho thiết bị ESP32

Thiết bị gọi qua cùng cổng với web (nginx proxy `/api`), ví dụ `http://<server>:8085/api`.

1. **Lần đầu bật máy**: `POST /api/device/register` với `{"name":"Desk box","firmware":"0.1.0"}`
   → nhận `{token, pair_code, pair_expires_at}`. Lưu `token` vào NVS, hiện `pair_code` lên màn hình.
2. **Chờ ghép**: gọi `GET /api/device/status` (header `Authorization: Bearer <token>`) mỗi 5 giây.
   Khi chưa ghép, kết quả là `{paired:false, pair_code}` (mã tự đổi sau 15 phút).
   User vào web → **Device & settings** → nhập mã. Sau đó kết quả là `{paired:true, user_name}`.
3. **Đồng bộ**: `GET /api/device/sync?limit=30` trả về `settings` (số từ mới/ngày, khoảng nhắc,
   giờ hoạt động), `utc_offset_sec`, các bộ đếm và `cards[]` ở dạng gọn
   (`id, kind, front, ipa, meaning_vi, meaning_en, example_en, example_vi, back, choices, answer, is_new`).
   Thẻ đến hạn ôn có thêm `options[]`: 4 nghĩa tiếng Việt để hỏi trắc nghiệm, đáp án đúng là option có `id` bằng `id` của thẻ.
   Chọn sai → gửi rating 1; chọn đúng → cho user chọn 2/3/4.
   Mỗi thẻ có `mode` (theo cài đặt "Device pop-up shows" = `settings.device_mode`):
   `flashcard` (hiện từ → chạm xem nghĩa → tự chấm), `quiz` (4 `options`, hoặc `choices` với thẻ mcq),
   `listening` (phát `audio_url` rồi chọn nghĩa). Thẻ mới luôn là `flashcard`; thẻ không có audio thì
   `listening` thành `quiz`. Thiết bị chỉ cần làm theo `mode`.
   Thiết bị **không** có các tính năng cần gõ/viết (Write it, Writing, điền từ trong More examples /
   Round check, dịch câu ở Forgotten words) và Reading/Shadowing/Role-play: những phần này chỉ có trên web.
   Lưu vào thẻ SD để học offline.
4. **Gửi kết quả** (có thể gửi dồn sau khi mất mạng):
   `POST /api/device/reviews` với
   `{"reviews":[{"client_review_id":"<uuid>","card_id":12,"rating":3,"reviewed_at":"2026-10-22T08:00:00Z","duration_ms":4000}]}`.
   `rating`: 1 = quên, 2 = khó, 3 = nhớ, 4 = dễ. Gửi lại cùng `client_review_id` sẽ không bị tính hai lần.
   Thiết bị nên lấy giờ qua NTP. Nếu giờ sai (ở tương lai), server dùng giờ của server.

### Quiz trên thiết bị

`/api/device/sync` trả thêm `quiz`: tối đa 40 thẻ đã học (không phải thẻ mới) thuộc chủ đề của thiết bị, ngẫu nhiên,
mỗi thẻ có `mode` quiz / cloze / listening và 4 lựa chọn. Câu trả lời gửi qua `/api/device/reviews` với `"quiz": true`:
chỉ thẻ đến hạn mới được xếp lịch lại (như Quiz trên web); trả lời sai thẻ chưa đến hạn thì thẻ đó đến hạn ngay
(trí nhớ FSRS không đổi).

### Câu hỏi AI cho từ hay quên (thiết bị)

`POST /api/device/practice {"card_ids":[...]}` (bấm *Practise these now*): mỗi từ 1 câu trắc nghiệm (usage /
synonym / collocation). Từ chưa có câu hỏi → một lần gọi DeepSeek cho tất cả, lưu ở bảng `word_questions` (dùng
lại về sau, không gọi AI nữa). Câu trả lời sai gửi như quiz review `Again` → từ đến hạn ngay.
Trên web: cuối Study / Quiz, ô "words to practise" có nút **Check them (A–D)** → trang `/practice/check?ids=…`
dùng chung các câu hỏi này (`POST /api/practice/word-questions`, trả lời sai → `.../answer` làm từ đến hạn ngay).

### Cài đặt của từng thiết bị

Trang Devices → nút **⚙ Settings** của từng thiết bị: **chủ đề** (vd. chỉ English), kiểu câu khi pop-up, số thẻ mỗi
pop-up (5/10/20), pop-up mỗi N phút và khung giờ. Đây là chỗ duy nhất chỉnh pop-up (Study settings bên phải chỉ còn
các mục dùng chung cho cả tài khoản: số thẻ mới mỗi ngày, deck cấp thẻ mới, múi giờ, nhắc học trên điện thoại).
API: `PUT /api/devices/:id/settings` (lưu ở `devices.settings`); `/api/device/sync` chỉ trả thẻ của các chủ đề đã
chọn và gửi các giá trị này trong `settings`. Thiết bị nhận thay đổi trong vòng 3 phút.

### Chế độ "cloze" trên thiết bị (điền từ, chỉ chọn)

`settings.device_mode = "cloze"` (hoặc nằm trong vòng `mix`): thẻ đã học có `mode: "cloze"` và
`cloze: {sentence: "Vegetable prices ____ wildly…", translation_vi, options: [{id, text} ×4]}`.
Đáp án đúng có `id == card id`. Thiết bị **không gõ chữ**, chỉ chọn 1 trong 4 từ; gửi kết quả như quiz
(đúng → rating 3, sai → 1) qua `POST /api/device/reviews`. Câu lấy từ câu ví dụ của thẻ (có xoay vòng câu AI đã lưu),
từ sai lấy từ các nghĩa nhiễu của thẻ. Thẻ không có câu phù hợp → quiz.

`cloze.kind`: `"word"` (4 từ khác nhau) hoặc `"form"` (4 dạng cùng họ từ kiểu TOEIC Part 5, đáp án sai có id âm),
dạng "form" chỉ có khi câu đó đã được chuẩn bị trên web.

**Round check trên web** có 3 cách làm:
- **Type**: gõ từ (đúng dạng trong câu hoặc dạng gốc).
- **Say**: đọc to **cả câu**, tự nói từ còn thiếu; chỉ kiểm tra từ đó có trong câu nghe được (có ngữ cảnh nên nhận
  diện tốt hơn đọc 1 từ). Chỉ để luyện, không đổi lịch ôn.
- **A B C D** (mặc định trên điện thoại): xen kẽ **Which word?** (4 từ cùng loại từ: danh/động/tính/trạng, lấy trong
  vòng học rồi trong deck, nên không đoán được bằng ngữ pháp) và **Which form?** (TOEIC Part 5:
  `fluctuated / fluctuation / fluctuating / fluctuant`). Dạng từ do AI viết **1 request cho cả vòng** khi bạn dùng
  A B C D, lưu ở `card_examples.form_options` (dùng chung, cả cho thiết bị). Phím 1–4 hoặc A–D.

## Reading (bài đọc / hội thoại do AI viết)

Trang **Reading** tạo một bài đọc kiểu TOEIC/IELTS hoặc một đoạn hội thoại hằng ngày, dùng
4–6 từ bạn đã học (ưu tiên từ hay quên, sắp đến hạn) và 2–3 từ sắp học. Có đọc to từng câu,
hội thoại 2 giọng, bật/tắt bản dịch tiếng Việt, 3–4 câu hỏi kèm giải thích và câu trích dẫn.
Câu hỏi về nghĩa của một từ đã học cũng cập nhật lịch ôn của từ đó (đúng = Good, sai = Again).
Server kiểm tra bài AI trả về (đủ câu hỏi, đáp án hợp lệ, ≥70% từ yêu cầu có trong bài) và yêu cầu
viết lại 1 lần nếu không đạt. Cần cấu hình AI (`DEEPSEEK_*`).

Luyện nói trong Reading (cần Chrome/Edge + HTTPS, không tốn lượt AI):
- **Shadowing** (bài đọc và hội thoại): app đọc từng câu → tự bật mic để bạn nhại lại → chấm %.
  Đạt 80% tự sang câu; chưa đạt thì "Listen slowly & try again" hoặc Skip (sau 3 lần vẫn cho qua).
  Câu dài hơn 20 từ được tách theo dấu phẩy. Cuối bài có tổng kết và "Practise the weak ones again".
- **Role-play** (chỉ hội thoại): chọn vai, app đọc lời người kia, đến lượt bạn thì bạn nói.
  *Easy*: câu của bạn hiện chữ (cần 80%). *Hard*: chỉ hiện nghĩa tiếng Việt, câu gốc bị ẩn,
  bạn tự nói bằng tiếng Anh (cần 60%); "Hear the line" để nghe gợi ý, sau 3 lần thì hiện câu gốc.

## Luyện nói (speech-to-text)

Nút 🎤 **Read it aloud** cạnh câu ví dụ (Review, phím `R`; Quiz sau khi trả lời): đọc to câu, trình duyệt
chuyển giọng nói thành chữ và app tô xanh/đỏ từng từ, chấm % khớp. Dùng Web Speech API của trình duyệt,
không cần server:
- Chỉ chạy trên **HTTPS** (hoặc localhost) và **Chrome / Edge / Safari**. Firefox không hỗ trợ nên nút bị ẩn.
- Chrome gửi âm thanh lên dịch vụ nhận giọng nói của Google để chuyển thành chữ.
- Lần đầu trình duyệt sẽ hỏi quyền micro.

## Menu

Home · **Study** (ôn thẻ đến hạn + từ mới) · **Practice** (Overview, Quiz, Reading & listening, Writing,
Forgotten words) · **Browse** (duyệt bộ từ; chọn "All decks" để tìm trong mọi bộ, thay cho trang Search cũ) · **Settings**.
Practice chỉ có trên web; thiết bị ESP32 chỉ làm phần Study (flashcard / quiz / listening).

## More examples (câu ví dụ theo ngữ cảnh)

Trong phần nghĩa của thẻ (Study khi lật thẻ, Browse) có **More examples**: tối đa 6 câu AI viết cho mỗi từ,
mỗi câu một ngữ cảnh (work, daily life, news, study…) và cùng thể hiện các cách kết hợp từ thường gặp.
Câu được tạo khi bấm "Get examples with AI" (1 request, lưu trong `card_examples`, dùng chung cho mọi người và
cho Forgotten words / Write it). Mặc định **Fill in the blanks**: từ đang học bị che, gõ đúng dạng trong câu
(gõ từ gốc cũng được, app chỉ ra dạng đúng) → app đọc câu → 🎤 đọc theo và chấm %. Có 🔊/🐢 (nghe để điền),
bấm vào câu để xem nghĩa tiếng Việt, công tắc tắt chế độ điền từ.
Phím (More examples và Round check): **Enter** chấm ô trống (sai thì chữ được bôi đen để gõ lại; đúng thì đọc
câu tiếng Anh và hiện nghĩa tiếng Việt) · **S** nói câu vừa điền và chấm % (≥ 80% tự sang ô tiếp theo) ·
**N** ô tiếp theo · **P** câu trước. S/N/P không chạy khi đang gõ trong ô.

## Phát âm (cho người Việt)

Dưới IPA của từ (Study, Browse) có **âm tiết + trọng âm**: `hy·POTH·e·sis` (âm nhấn viết hoa, tô màu),
nút 🐢 đọc chậm từng âm tiết rồi đọc cả từ, và 1–3 **mẹo bằng tiếng Việt** cho âm khó của chính từ đó
(th /θ/, âm cuối -s/-ts, nguyên âm yếu /ə/...). Không phiên âm bằng chữ tiếng Việt (dễ học sai âm và trọng âm).
Hiện thẻ chỉ **đọc phần đã lưu** (`POST /api/cards/pronunciations`, gom nhiều thẻ, không gọi AI); từ chưa có
thì hiện nút **How to say it** → `POST /api/cards/:id/pronunciation` = 1 request AI cho đúng từ đó.
Kết quả được kiểm tra (âm tiết ghép lại đúng chính tả) và lưu trong `card_pronunciations`, dùng chung.

## Round check (điền từ cuối round)

Cuối mỗi lượt **Study** (khi hết một lô thẻ, nếu có ≥ 3 từ tiếng Anh) và ở màn hình điểm của **Quiz**
có thẻ **Round check: fill in the blanks**: bấm Start → `/api/practice/cloze` lấy mỗi từ một câu mới
(câu AI trong `card_examples`, thiếu thì AI viết bù trong 1 request), trộn thứ tự. Có danh sách từ
(gạch đi khi điền đúng); gõ từ vào ô trống, Enter chấm ngay (đúng dạng hoặc từ gốc; sai 2 lần thì hiện đáp án);
điền xong thì nghe 🔊 / đọc theo 🎤, bấm "Vietnamese" để xem nghĩa câu. Có nút Skip.

## Quiz: Write it

Trong **Quiz** chọn **Write it**: mỗi lượt 10 câu tiếng Việt (lấy từ `card_examples`, câu tiếng Anh là đáp án mẫu;
từ nào chưa có câu thì AI viết bù trong 1 request). Bạn viết câu tiếng Anh dùng từ có nghĩa được cho;
nút **?** hiện từ cần dùng (câu đúng sau khi bấm **?** không được tính là ôn). Nộp từng câu thì app kiểm tra
ngay câu có dùng từ đó không (kể cả dạng chia), hiện câu mẫu và **AI chấm ngay câu đó** (đúng/sai, điểm,
góp ý, câu đã sửa; 1 request mỗi câu, chạy nền trong lúc bạn đọc câu mẫu). Câu sai có nút 🎤 **Read the corrected sentence**
(app đọc câu đã sửa, bạn đọc theo, chấm %; không bắt buộc). Cuối lượt chỉ chấm bù các câu bị lỗi mạng. Câu đúng có dùng từ (không bấm ?) = ôn "Good" nếu thẻ đang đến hạn; câu sai không bị phạt.
Từ sai/bỏ qua được gợi ý luyện ở Forgotten words.

## Writing: "Write about this" từ Reading

Cuối mỗi bài Reading có thẻ **Write about this** → `/writing?reading_id=…`. Lần đầu mở, **1 request AI**
soạn đề theo bài (trả lời email/thông báo, tin nhắn sau hội thoại, ý kiến/tóm tắt cho bài báo, hoặc
**bình luận casual** kiểu forum/Reddit — `comment_reply`, văn phong thân mật) kèm
3 từ bắt buộc lấy từ bài, số từ theo trình độ, **dàn ý** (tiếng Việt), **cụm từ hữu ích** và **bài mẫu**;
đề được lưu trong bài (`content.writing_task`), lần sau mở ngay. Không có AI: dùng đề mẫu + dàn ý/cụm từ chung.
Trang Writing hiện đề, đếm từ, từ bắt buộc chuyển xanh khi đã dùng; gợi ý mở dần: Outline → Useful phrases
(bấm để chèn vào bài) → Sample answer. Chấm điểm như Writing thường.

Điểm: thang **0–100** cho Overall / Grammar / Vocabulary / Coherence, kèm nhãn Good (80+), Fair (60–79),
Needs work (40–59), Redo (<40). AI được đưa đúng các mốc này; thiếu số từ hoặc thiếu từ bắt buộc thì
Overall tối đa 60 (thiếu số từ: server tự giới hạn). **Chấm theo register của đề** (email công sở /
bình luận casual / essay academic), không cứng IELTS — bài casual không bị trừ điểm vì viết thân mật;
bản viết lại "nâng cao" cũng giữ đúng văn phong của đề.

## Cheat sheet (ghi chú tự do theo từng khóa)

Trang **Cheat sheet** (menu, theo môn đang chọn) — ghi chú **cá nhân của từng user**, tách khỏi Notes-theo-từ:
- Thêm/sửa/xóa/di chuyển **từng item**; nội dung viết bằng **Markdown** (# tiêu đề, danh sách, **đậm**, `code`,
  khối ```code```), hiện bản render kèm xem trước khi gõ.
- Nút **Make a cheat sheet (AI)**: AI đọc toàn bộ item của khóa → sinh một bản tóm tắt gọn (Markdown), có nút
  **Regenerate** và **Copy**. Bản AI được lưu lại (cache per user+subject); ghi chú gốc là nguồn chính, AI không đè lên.
- Markdown render bằng bộ tự viết (`web/src/utils/markdown.ts`), escape HTML trước nên không chạy được script — an toàn.
- Nút **"Use this as my notes"**: thay toàn bộ ghi chú (lộn xộn) bằng bản AI đã sắp xếp, **tách theo tiêu đề**
  thành từng item để vẫn sửa được. Thao tác này xóa ghi chú cũ nên có hỏi xác nhận.
- API: `GET/POST/PATCH/DELETE /course-notes`, `POST /course-notes/summarize`, `POST /course-notes/adopt-summary`. Bảng `course_notes`,
  `course_note_summaries` (migration 0020).
- **Widget ghi chú nổi** (`CourseNotesWidget.vue`, góc dưới-phải, mọi trang khi đã đăng nhập): gõ/dán note nhanh cho
  **khóa đang chọn** mà không rời trang (hợp để chép câu lệnh khi đang Browse), lưu chung store với trang Cheat sheet;
  có danh sách note gần đây + link mở Cheat sheet. Ở môn English nó tự đẩy lên trên để không đè widget "My notes" theo-từ.

## Hỏi AI theo khóa (chat trong trang Cheat sheet)

Có nút **Ask AI** nổi (góc dưới-trái, mọi trang khi đã đăng nhập) + khung Ask AI trong trang Cheat sheet —
dùng chung component `CourseChat`. Hỏi tự do về khái niệm/lệnh/ví dụ của **môn đang chọn**; bám theo nút chuyển môn.
- AI trả lời bằng **Markdown** (có code block), và được đưa kèm **ghi chú của bạn** trong khóa làm ngữ cảnh.
- Giữ ngữ cảnh hội thoại (vài lượt gần nhất); chat là tạm thời trong phiên (chưa lưu lại).
- **Chọn model để hỏi**: ô chọn model lấy từ `/v1/models` của gateway (hoặc danh sách ghim trong `CHAT_MODELS`).
  "Default" = model mặc định của server. Server chỉ chấp nhận model nằm trong danh sách đó.
- API: `POST /course-notes/ask` ({subject_id, question, model?, history}), `GET /chat-models`.

## Quản lý khóa học (trang Courses, chỉ admin)

Trang **Courses** (menu chỉ hiện với tài khoản admin — email trong `ADMIN_EMAILS`) để tạo và quản lý khóa học
ngay trên web, không cần viết migration:
- Tạo **khóa học mới** (subject): tên, slug (tự sinh từ tên), mô tả, thứ tự. Sửa lại tên/mô tả/thứ tự sau.
- Trong mỗi khóa: thêm/sửa **deck** (tên, slug, mô tả, loại thẻ mặc định: vocab/phrasal/concept/mcq), xem số thẻ.
- **Sinh thẻ bằng AI** cho từng deck (nút AI, chọn số lượng ≤20) — nối thêm vào deck, không đụng thẻ cũ và tiến độ.
- **AI: gen toàn bộ course** (nút cho admin, hợp với môn scope hữu hạn như PMP/Pentest): AI liệt kê toàn bộ danh mục
  khái niệm của từng deck (`POST /decks/:id/outline`), rồi tự sinh theo lô cho đến khi phủ hết scope hoặc đạt **trần/deck**
  (mặc định 150). Có thanh tiến độ, chọn model, nút Dừng. **Resumable** (chạy lại chỉ sinh phần còn thiếu). Khi AI không
  còn khái niệm mới (dưới trần) → báo *đã phủ hết scope*; khi chạm trần → báo *đạt trần, tăng trần nếu muốn thêm*.
- Link sang **Browse** để thêm/sửa thẻ bằng tay (mỗi thẻ có nút ✏️ sửa trực tiếp).
- API đã có sẵn (POST/PUT `/subjects`, `/decks`, `/decks/:id/generate` [thêm `terms`/`model`], `/decks/:id/outline`, card CRUD).
- Chưa có **xóa** khóa học/deck trên UI (thao tác này xóa cả tiến độ người học nên để sau).

Khóa mới tạo xong tự xuất hiện ở nút chuyển môn English | PMP | Pentest | … như một môn bình thường.

## Quiz cho môn không phải tiếng Anh

Ở các môn không phải English (PMP, Pentest, khóa mới), trang Quiz **ẩn** các phần chỉ dành cho học từ tiếng Anh:
nút Play (đọc từ), "listen to the example", "Show sentence", phần phát âm/shadowing. Chỉ còn câu hỏi + đáp án + giải thích.

## Pentesting (môn thứ 3)

Môn **Pentesting** (migration 0019) để học nền tảng cho chứng chỉ HTB CPTS và THM Penetration Tester:
6 deck, 57 thẻ khái niệm + 12 câu hỏi trắc nghiệm.

- **Methodology & scoping** (có uỷ quyền/văn bản cho phép, scope, RoE, các giai đoạn, báo cáo, CVE/CVSS).
- **Linux & shell** (quyền file, SUID, sudo -l, cron, /etc/passwd vs shadow, các lệnh tìm/đọc).
- **Networking & enumeration** (bắt tay TCP, cổng phổ biến, trạng thái cổng, DNS/subdomain, OSINT).
- **Web vulnerabilities (OWASP)**: mỗi lớp lỗi kèm **cách phòng chống** (SQLi → parameterized query, XSS → output
  encoding + CSP, CSRF → token + SameSite, IDOR, SSRF, command injection, traversal, session/cookie…).
- **Tools of the trade**: mỗi công cụ nói rõ **để làm gì**; các công cụ tấn công nêu kèm giới hạn "chỉ test khi
  được phép" và biện pháp phòng thủ (ví dụ cracking hash → bcrypt/argon2 + MFA).
- **Practice questions**: trắc nghiệm kiểu đề thi.

Nội dung là tài liệu học: nhận biết, phương pháp và cách phòng thủ — không phải hướng dẫn tấn công vào một mục tiêu
cụ thể. Giống PMP, môn này chỉ tự lấy thẻ mới nếu bạn **tick deck Pentesting trong Settings** (hoặc thêm thẻ tay);
đặt số thẻ mới/ngày riêng cho nó trong Settings như các môn khác.

## Auto-pace (tự điều chỉnh số thẻ mới/ngày — tùy chọn, mặc định tắt)

Bật trong Settings. Mỗi ngày (lần đầu mở app trong ngày) server xét từng môn:
- Giữ mức ghi nhớ **Strong** đủ lâu (khoảng 7 ngày liên tục) → **tăng** số thẻ mới/ngày của môn đó **+2** (trần 20).
- Rơi về **Needs focus** → **giảm −2** (sàn 3).
- **Okay/mới học** → giữ nguyên, và chuỗi "strong" reset.

Nguyên tắc: dựa trên tín hiệu bền (nhiều ngày, không phải một lần), có trần/sàn, chạy 1 lần/ngày, nạp thẻ
**theo thứ tự** (không random), và **minh bạch** — mỗi lần đổi ghi một dòng trong Settings
("2026-10-27 · English 10 → 12/day (memory strong)"). Bạn vẫn chỉnh tay số thẻ/ngày đè lên bất cứ lúc nào.
Tắt đi thì không bao giờ tự đổi. Ngưỡng Strong/Needs focus dùng chung với khung thống kê Home.

## Lời khuyên học (trên khung thống kê Home)

Khung "How well you know your cards" (mỗi môn) đưa ra:
- **Mức ghi nhớ tổng thể** (badge cạnh tiêu đề): **Strong / Okay / Needs focus** (hoặc "Just started"),
  tính từ tỉ lệ nhớ tốt (mastered+familiar), tỉ lệ hay quên (struggling) và độ chính xác 30 ngày.
- **Nên học thêm từ mới hay nên ôn lại**:
  - Mới học (<10 thẻ): học ít thôi (~5–10/ngày) rồi xem phản ứng.
  - Căng (due ≥ 30, hoặc ≥25% thẻ hay quên, hoặc accuracy < 65%): **ngưng từ mới, tập trung ôn**.
  - Hơi căng (due ≥ 15, ≥12% hay quên, accuracy < 80%): giữ từ mới ở mức thấp (~5/ngày).
  - Khỏe mạnh: có thể tăng số thẻ mới/ngày (10–15).
  Kèm nút ôn ngay và mở Settings để chỉnh số thẻ mới/ngày. Tính riêng từng môn.

## English và PMP tách riêng

- Trên thanh menu có nút **English | PMP** (nhớ theo trình duyệt). Mọi trang chỉ hiện môn đang chọn: Home (số thẻ,
  chuỗi ngày, biểu đồ, thống kê tiến độ, deck), Study, Practice (PMP: Quiz + các deck PMP; English: thêm Reading,
  Writing, Forgotten words), Quiz, Browse. Nút ghi từ (Notes) chỉ hiện ở English.
- Mở link của một deck/môn (ví dụ `/review?deck_id=5`) thì tự chuyển sang môn của nó.
- **Thẻ mới mỗi ngày tính riêng từng môn** (Settings: English 10, PMP 5…); "Learn 5 more today" / "+5 new" cũng
  theo môn. Mỗi môn chỉ lấy thẻ mới từ các deck được tick của môn đó (không tick deck nào cả = English Core Vocabulary).
- Thông báo đẩy: "8 English · 3 PMP to review". Thiết bị dùng các môn đã chọn cho nó.

## Nhắc học trên web

- Thông báo đẩy (giờ cố định trong Settings) **bỏ qua** nếu đã học (web hoặc thiết bị) trong 60 phút trước đó.
- Đang mở web (không ở Home/Study/Quiz): mỗi chu kỳ *pop-up every N min* của thiết bị, trong khung giờ của nó,
  nếu có thẻ đến hạn và chưa học trong chu kỳ đó → dải nhắc nhỏ phía trên trang (**Study now** / × để nhắc sau).
  Không kêu, không pop-up.

## Phrase bank (khung câu để viết)

Bộ **Writing phrases** (migration 0018, 81 khung, 14 nhóm: mở bài, quan điểm, lý do, ví dụ, thêm ý, đối lập,
nguyên nhân – kết quả, giải pháp, ưu – nhược điểm, kết luận, và email: mở đầu / đề nghị / báo vấn đề / kết thư).
Mỗi thẻ: mặt trước là **điều muốn nói** (tiếng Việt), mặt sau là **khung tiếng Anh** + ghi chú + câu ví dụ.

- Trang Writing → nút **Phrase bank**: tìm (không dấu cũng được), lọc theo nhóm, bấm khung để chèn vào chỗ con trỏ,
  ⓘ xem ví dụ. **Learn these in Study** thêm cả bộ vào Study (vào dần theo số thẻ mới mỗi ngày).
- Study (web và thiết bị): thấy câu tiếng Việt → tự nhớ câu tiếng Anh → chọn trong 4 khung **cùng nhóm**
  (ví dụ 4 cách kết thư), nên không đoán được nhờ chức năng.
- Khi chấm bài, AI chỉ ra tối đa 3 câu nên viết lại bằng một khung trong bộ (**Use a frame**).

## My notes (ghi chú từ lạ)

Nút 📝 ở góc dưới bên phải (mọi trang, **Alt+N**) mở khung ghi chú: gõ từ/cụm từ + ghi chú riêng.
**Bôi đen một từ** ở bất kỳ đâu (bài Reading, câu ví dụ, câu hỏi...) → nút **Note** → lưu từ kèm **câu chứa nó**
và trang đang học. Chỉ nhận từ tiếng Anh (chữ có dấu/khác hệ chữ bị từ chối); khi giải nghĩa, từ AI nói
không phải tiếng Anh (gõ sai, từ bịa) bị xóa và được báo lại. Từ đã có trong kho thì hiện nghĩa ngay; **Explain** = 1 request AI giải nghĩa mọi từ mới
(dạng gốc, IPA, nghĩa theo đúng ngữ cảnh, ví dụ = câu bạn gặp); **+ Study / Add to Study** = vào lịch ôn:
dùng thẻ có sẵn (kể cả khi bạn ghi dạng chia: "negotiated" → "negotiate"), hoặc tạo thẻ trong bộ riêng
**My words** (chỉ bạn thấy; dùng được với Study, Quiz, More examples, Write it, thiết bị ESP32).
Bộ từ có `owner_id` là bộ cá nhân; mọi truy vấn thẻ/bộ đều lọc theo `study.VisibleDecksSQL`.

## Luyện từ đã quên

Trong **Review** (thẻ bị chấm Again: chọn sai, "I don't remember", hoặc Again) và **Quiz** (chọn sai lần đầu),
app gom các từ tiếng Anh đã quên. Màn hình cuối (và thanh trên cùng trong Review) có nút **Practise them**
→ trang `/practice/forgotten` (mở trực tiếp thì tự lấy các từ bị chấm Again hôm nay chưa luyện):
1. **1 request AI** viết 3 câu ví dụ cho mọi từ (tối đa 10 từ/lần), lưu vào `card_examples` để lần sau mở ngay.
   Mỗi thẻ giữ 6 câu; lần quên sau sẽ gặp 3 câu khác (`practice_rounds`).
2. Mỗi câu: gõ bản dịch tiếng Việt → xem bản dịch mẫu → đọc to (🎤, cần ≥ 80%, sau 3 lần vẫn cho qua;
   công tắc "Can't speak right now" để bỏ bước đọc). Câu 1 của mỗi từ là bắt buộc, câu 2–3 tùy chọn.
3. Cuối buổi **1 request AI** chấm mọi bản dịch (đúng/sai, điểm, góp ý tiếng Việt).

Không có AI: dùng câu ví dụ có sẵn của thẻ và bỏ bước chấm. Lịch ôn FSRS không bị thay đổi bởi phần luyện này.

## Phát âm có lịch ôn riêng (trọng âm)

Nghĩa và cách đọc được ôn **tách nhau**: `user_cards` là lịch ôn nghĩa, `user_pron` là lịch ôn phát âm
(FSRS riêng, log trong `pron_logs`). Quên cách đọc không làm reset lịch ôn nghĩa.

- **Câu hỏi trọng âm trong Quiz** (tối đa 3/10 câu): "Which syllable is stressed?", chạm âm tiết (hoặc phím 1–9).
  Chấm chính xác theo dữ liệu đã lưu (`card_pronunciations`), không dùng micro. Sai → hỏi lại sớm; đúng → giãn ra.
  Ưu tiên từ có lịch phát âm đến hạn, rồi đến từ chưa từng hỏi trọng âm. Chỉ từ đơn có ≥ 2 âm tiết.
- **Câu hỏi "Which sound?"** (chung suất với câu trọng âm, cùng lịch ôn phát âm): *How do you say the “u” in FLUC?*
  → chọn `/ʌ/ as in “cut”` (nút **?** hiện âm gần giống tiếng Việt: "gần “â”"). Sau khi trả lời: quy tắc chính tả
  bằng tiếng Việt + vài từ cùng âm (bấm để nghe). Dữ liệu (`card_pronunciations.sound`) do AI ghi cùng lúc với âm tiết;
  từ đã có âm tiết từ trước được bổ sung khi bấm **Prepare** (NULL = chưa chuẩn bị, `{}` = từ không có âm dễ đọc sai).
- **Say it first** (bật/tắt ở Quiz): trước khi chọn nghĩa, đọc to từ. 🎤 chỉ để luyện (máy nghe chỉ là tham khảo,
  không đổi lịch ôn). **I forgot how to say it (F)** → nghe mẫu chậm + trọng âm ngay, và từ vào lịch ôn phát âm
  (`POST /api/pron/forgot`). Trình duyệt không có micro: nút "I said it".
- Sau mỗi câu trả lời, từ luôn hiện kèm trọng âm (hy·POTH·e·sis).
- Từ chưa có dữ liệu âm tiết: Quiz hiện nút **Prepare 25 more words (AI)** → 1 request AI cho 25 từ của deck,
  lưu dùng chung (`POST /api/quiz/pron-generate`, `GET /api/quiz/pron-status`).

## Cài lên điện thoại (PWA) + nhắc giờ học

Web là PWA: `web/public/manifest.webmanifest`, icon trong `web/public/icons/`, service worker `web/public/sw.js`
(viết tay, không qua plugin). Service worker cache khung app để mở nhanh; `/api` luôn lấy từ mạng.

- **Cài app:** Settings → *Phone app & study reminders* → **Install app** (Chrome/Edge/Android).
  iPhone/iPad: mở bằng Safari → Share → **Add to Home Screen**.
- **Nhắc giờ học:** bật *Study reminders on this device*, chọn giờ (tối đa 6 mốc, theo time zone của tài khoản).
  Đến giờ server gửi Web Push: "12 cards to review · 5 new words left today". Không có gì để học thì không gửi.
  Bấm thông báo → mở trang Study. *Send a test now* để thử ngay.
- iOS chỉ nhận thông báo khi app đã được thêm vào Home Screen (iOS 16.4+). Cần HTTPS (đã có qua Cloudflare).
- Khóa VAPID tự sinh lần đầu và lưu trong bảng `app_secrets` (không cần cấu hình). `VAPID_SUBJECT` trong `.env`
  là tùy chọn (mặc định `https://APP_DOMAIN` hoặc `mailto:` admin đầu tiên).
- Server chỉ gửi tới push service của trình duyệt (FCM, Mozilla, Apple, Windows), không gửi tới URL tùy ý.
- API: `GET /api/push/key`, `POST /api/push/subscribe`, `POST /api/push/unsubscribe`, `POST /api/push/test`;
  giờ nhắc lưu trong `settings.push_times`. Worker kiểm tra mỗi phút (`internal/worker/reminders.go`),
  bảng `push_reminders_sent` chống gửi trùng.

## Ask about this (hỏi thêm AI sau khi chấm)

Dưới mỗi kết quả AI chấm (Write it, Writing, luyện từ đã quên) có **Ask about this**: gõ câu hỏi
("analyze và analyst khác gì?"), AI trả lời ngắn bằng tiếng Việt, biết sẵn đề, câu bạn viết, câu sửa và nhận xét.
Hỏi tiếp được (gửi kèm 4 lượt gần nhất). **1 request AI mỗi câu hỏi**, không lưu lại.
Các từ được nhắc tới hiện thành nút **+ từ** để lưu vào My notes. API: `POST /api/practice/ask`.

## Lưu từ khi lướt web (extension) và từ điện thoại (Share)

- **Extension** (thư mục `extension/`, xem `extension/README.md`): bôi đen từ → chuột phải → *Add to E-Learn notes*
  hoặc **Alt+Shift+N** → lưu vào My notes kèm câu chứa từ. Đăng nhập bằng **khoá riêng** tạo ở Settings →
  Browser extension (`ext_…`, chỉ lưu hash; chỉ được thêm/đọc notes; xoá là hết hiệu lực). API `/api/ext/*`.
- **Android**: app đã cài (PWA) xuất hiện trong menu **Share**: bôi đen từ ở bất kỳ app nào → Share → E-Learn →
  trang `/share` (văn bản dài thì chạm chọn từ, câu chứa từ được giữ làm ngữ cảnh). iPhone không hỗ trợ share target.

## Firmware thiết bị ES3C28P

Thư mục `firmware/` (ESP-IDF 5.4 + LVGL 9). Cách nạp, cài Wi-Fi, ghép nối và chỉnh màn hình: xem `firmware/README.md`.
File nạp sẵn: `firmware/prebuilt/elearn-device-1.3.4.bin` (một file, địa chỉ 0x0, giữ Wi-Fi và liên kết).
Cập nhật qua Wi-Fi: `deploy.sh` đưa `firmware/prebuilt/*-app.bin` lên server (mount vào API ở `/firmware`);
`/api/device/sync` trả `firmware.version`, thiết bị tải `/api/device/firmware/<file>` khi bấm Update.

### Đọc từ vựng trên thiết bị (container `tts`)

- Container `tts` (Python + edge-tts + ffmpeg, chỉ nội bộ) đọc từ bằng giọng Microsoft Edge, trả WAV 16 kHz mono.
- Khi thiết bị đồng bộ, API tạo audio cho các thẻ vocab/phrasal còn thiếu (chờ tối đa ~6 giây, phần còn lại
  làm tiếp ở nền) và lưu vào bảng `tts_audio` (~20–40 KB/từ). Thiết bị tải ở `/api/audio/<key>.wav`.
- **Dọn dẹp**: từ không được phát trong `TTS_KEEP_DAYS` ngày (mặc định 60) tự xoá mỗi ngày; xoá rồi vẫn được
  tạo lại khi học tới. Admin xem dung lượng và **Delete all audio** ở trang Devices (`GET/DELETE /api/admin/audio`).
- Edge TTS là dịch vụ không chính thức. Nếu log `tts:` báo lỗi liên tục: `bash deploy.sh --clean` để cài edge-tts mới nhất.
  Tắt hẳn: `TTS_URL=` (để trống) trong `.env`.

## Còn thiếu / bước tiếp theo

- Dùng mic trên thiết bị (cần speech-to-text phía server).
- Chấm phát âm (luyện nói) qua dịch vụ speech-to-text.
