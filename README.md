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
   Lưu vào thẻ SD để học offline.
4. **Gửi kết quả** (có thể gửi dồn sau khi mất mạng):
   `POST /api/device/reviews` với
   `{"reviews":[{"client_review_id":"<uuid>","card_id":12,"rating":3,"reviewed_at":"2026-10-22T08:00:00Z","duration_ms":4000}]}`.
   `rating`: 1 = quên, 2 = khó, 3 = nhớ, 4 = dễ. Gửi lại cùng `client_review_id` sẽ không bị tính hai lần.
   Thiết bị nên lấy giờ qua NTP. Nếu giờ sai (ở tương lai), server dùng giờ của server.

## Còn thiếu / bước tiếp theo

- Firmware ESP32 (LVGL, WiFiManager captive portal, lưu cache trên SD).
- Audio MP3 cho thiết bị (server sinh bằng edge-tts, trường `audio_url` đã có sẵn).
- Chấm phát âm (luyện nói) qua dịch vụ speech-to-text.
