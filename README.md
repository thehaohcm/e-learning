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
   Lưu vào thẻ SD để học offline.
4. **Gửi kết quả** (có thể gửi dồn sau khi mất mạng):
   `POST /api/device/reviews` với
   `{"reviews":[{"client_review_id":"<uuid>","card_id":12,"rating":3,"reviewed_at":"2026-10-22T08:00:00Z","duration_ms":4000}]}`.
   `rating`: 1 = quên, 2 = khó, 3 = nhớ, 4 = dễ. Gửi lại cùng `client_review_id` sẽ không bị tính hai lần.
   Thiết bị nên lấy giờ qua NTP. Nếu giờ sai (ở tương lai), server dùng giờ của server.

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

## Luyện từ đã quên

Trong **Review** (thẻ bị chấm Again: chọn sai, "I don't remember", hoặc Again) và **Quiz** (chọn sai lần đầu),
app gom các từ tiếng Anh đã quên. Màn hình cuối (và thanh trên cùng trong Review) có nút **Practise them**
→ trang `/practice`:
1. **1 request AI** viết 3 câu ví dụ cho mọi từ (tối đa 10 từ/lần), lưu vào `card_examples` để lần sau mở ngay.
   Mỗi thẻ giữ 6 câu; lần quên sau sẽ gặp 3 câu khác (`practice_rounds`).
2. Mỗi câu: gõ bản dịch tiếng Việt → xem bản dịch mẫu → đọc to (🎤, cần ≥ 80%, sau 3 lần vẫn cho qua;
   công tắc "Can't speak right now" để bỏ bước đọc). Câu 1 của mỗi từ là bắt buộc, câu 2–3 tùy chọn.
3. Cuối buổi **1 request AI** chấm mọi bản dịch (đúng/sai, điểm, góp ý tiếng Việt).

Không có AI: dùng câu ví dụ có sẵn của thẻ và bỏ bước chấm. Lịch ôn FSRS không bị thay đổi bởi phần luyện này.

## Còn thiếu / bước tiếp theo

- Firmware ESP32 (LVGL, WiFiManager captive portal, lưu cache trên SD).
- Audio MP3 cho thiết bị (server sinh bằng edge-tts, trường `audio_url` đã có sẵn).
- Chấm phát âm (luyện nói) qua dịch vụ speech-to-text.
