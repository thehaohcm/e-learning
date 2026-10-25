# Firmware thiết bị học ES3C28P (ESP32-S3, 2.8", LVGL)

Thiết bị để bàn: ôn thẻ (flashcard / quiz / listening / cloze / mcq) bằng cách **chạm chọn**, không gõ chữ.
Dữ liệu lấy từ server qua `/api/device/*`, lưu cache để học được khi mất mạng; câu trả lời gửi lại khi có mạng.

- Phần cứng: ESP32-S3 16MB flash + 8MB PSRAM, LCD ILI9341 240×320 (SPI), cảm ứng FT6336 (I2C),
  codec ES8311 + loa (I2S), đèn nền IO45, pin đo qua ADC.
- Phần mềm: ESP-IDF **5.4**, LVGL **9.3**, esp_lvgl_port, font DejaVu (có tiếng Việt và IPA).

## Nạp bằng cáp (không cần cài ESP-IDF)

1. Vào chế độ nạp: rút cáp, **giữ BOOT**, cắm cáp, thả BOOT (hoặc đang cắm: giữ BOOT, bấm RST, thả BOOT).
2. Mở https://espressif.github.io/esptool-js/ bằng Chrome/Edge → **Connect**.
3. Chọn **một file** `prebuilt/elearn-device-<bản mới nhất>.bin` (file **không** có chữ `-app`), địa chỉ **0x0** → **Program**. Không bấm Erase Flash.
4. Xong: rút cáp, cắm lại không giữ nút nào.

Từ bản **1.1.0**, Wi-Fi, liên kết và thẻ đã lưu nằm ở cuối bộ nhớ nên nạp file ở 0x0 **không làm mất** chúng.
Chuyển từ bản 1.0.x lên 1.1.0 (đổi cách chia bộ nhớ) thì phải cài Wi-Fi và ghép nối lại **một lần**.

File prebuilt dùng cấu hình mặc định (ngang, màu IPS). Nếu màu/cảm ứng bị sai, xem mục *Chỉnh màn hình* rồi tự build.

## Cập nhật qua Wi-Fi (từ bản 1.1.1)

1. Build bản mới (đổi `PROJECT_VER` trong `CMakeLists.txt`), đặt file app vào `prebuilt/elearn-device-X.Y.Z-app.bin`
   (= `build/elearn_device.bin`), rồi chạy `bash deploy.sh`: server giữ bản mới nhất.
2. Sau lần Sync tiếp theo (≤ 3 phút), màn hình chính báo *New firmware X.Y.Z*; vào **Settings → Update to X.Y.Z**.
   Máy tải về (~30 giây), cài, khởi động lại. Câu trả lời chưa gửi được gửi trước khi cập nhật.
3. Nếu bản mới không khởi động tới màn hình chính, lần khởi động lại sau máy **tự quay về bản cũ**.

Cáp USB vẫn dùng được bất cứ lúc nào (một file ở 0x0, xem trên).

## Build từ mã nguồn

```bash
# cài ESP-IDF 5.4 (hoặc VS Code + extension "ESP-IDF")
cd firmware
idf.py set-target esp32s3
idf.py menuconfig        # tuỳ chọn: "E-Learn device"
idf.py build flash monitor
```

Lần build đầu tự tải thư viện (LVGL, driver LCD/cảm ứng/codec) theo `main/idf_component.yml`.

## Lần đầu sử dụng

1. Bật máy → màn hình hiện mã QR và tên Wi-Fi **ELearn-XXXX**. Kết nối điện thoại vào Wi-Fi đó
   (trang cài đặt tự mở, hoặc vào http://192.168.4.1).
2. Chọn Wi-Fi nhà, nhập mật khẩu, kiểm tra địa chỉ server → **Save**. Máy khởi động lại.
3. Màn hình hiện **mã ghép nối**. Trên website: **Settings → Devices → Link a device** → nhập mã.
4. Máy tải thẻ và vào màn hình chính (số thẻ cần ôn, thẻ mới, chuỗi ngày học).

Đổi Wi-Fi: **Settings → Change Wi-Fi** trên máy, hoặc cắm điện rồi **bấm giữ BOOT ngay khi màn hình hiện "Starting…"** (giữ BOOT từ trước khi cắm thì chip vào chế độ nạp).
Huỷ liên kết: **Settings → Unlink** (hoặc xoá thiết bị trên website; máy sẽ tự hiện mã ghép nối mới).

## Học

- **Study**: thẻ đến hạn trước, sau đó thẻ mới (số lượng theo cài đặt trên website).
  - Flashcard: chạm *Show answer* → chọn Again / Hard / Good / Easy.
  - Quiz / listening / cloze / mcq: chọn 1 trong 4. Đúng = Good, sai = Again, có âm báo.
- Câu trả lời lưu trong máy (tối đa 400) và gửi lên server mỗi phút; thẻ mới và cài đặt tải về mỗi 3 phút
  hoặc khi bấm **Sync**. Mất mạng vẫn học được trên bộ thẻ đã tải.
- **Nhắc học**: trong khung giờ đặt trên website, nếu có thẻ chưa làm và máy đã rảnh đủ lâu
  (*pop-up every N min*), máy tự sáng lên, kêu "ting" và đưa ra 5/10/20 thẻ (*Cards per pop-up*, mặc định 10).
  Hết lượt: **Keep studying** để học tiếp phần còn lại, hoặc **Done**.

## Nhớ thật, không đoán

- **Recall first** (Settings trên máy, mặc định bật): câu trắc nghiệm chỉ hiện từ; tự nhớ nghĩa rồi bấm
  **Show options**, hoặc **I don't know** (tính là quên).
- Chọn đúng nhưng nhờ nhìn đáp án: bấm **Guessed** (tính là quên, từ đó quay lại sớm). Không bấm = nhớ.
- Cuối mỗi lượt: danh sách từ ✓ / ✗ / ?; các từ sai/đoán đã được đánh dấu, đánh dấu thêm từ nào chưa chắc
  rồi **Practise these now** (flashcard: tự nhớ → xem đáp án → Again / Got it; không tính vào lịch ôn).
- **Practise these now** (cuối lượt): mỗi từ đã tick = 1 flashcard + 1 **câu hỏi AI** (dùng đúng ngữ cảnh / gần nghĩa
  nhất / kết hợp từ). Câu hỏi được AI viết lần đầu (một lần gọi cho tất cả từ, ≤ ~40 giây, màn hình chờ
  "Preparing questions…") rồi lưu lại dùng mãi. Sai câu AI → từ đó đến hạn ôn ngay. Mất mạng/AI lỗi → chỉ flashcard.
- Đáp án gây nhiễu ưu tiên cùng loại từ (động từ với động từ) nên khó loại trừ hơn.

## Khi không còn gì để học

- **Pop-up** (đến giờ, máy rảnh) mà không còn thẻ ôn / thẻ mới: máy hỏi **Quiz 5 từ đã học** thay vì im lặng
  (không làm tăng lịch ôn; sai thì từ đó được ôn lại sớm).
- Màn hình chính và cuối Quiz có nút **+5 new words**: thêm 5 từ mới cho hôm nay (tối đa +20/ngày, giống
  "Learn 5 more today" trên web). Không tự động thêm, vì mỗi từ mới kéo theo nhiều lần ôn các tuần sau.

## Quiz

Nút **Quiz** trên màn hình chính: 10 câu ngẫu nhiên từ các từ **đã học** (chủ đề của máy), lúc nào cũng làm được,
kể cả khi Study đã hết. Trộn 3 kiểu: chọn nghĩa của từ (máy đọc từ), nghe rồi chọn nghĩa, điền từ vào câu.
Cuối lượt: điểm và các từ sai kèm nghĩa; **Quiz again** hoặc **Home**.
Câu **"Which form fits?"** (TOEIC Part 5: câu ví dụ AI có chỗ trống, chọn dạng đúng của từ): bấm Quiz là máy nhờ
server chuẩn bị sẵn (AI, chạy nền, tối đa 12 câu/lần, 10 phút một lần); các lượt Quiz sau sẽ có nhiều câu dạng này.
Lịch ôn: trả lời đúng một từ đang đến hạn = đã ôn; từ chưa đến hạn không đổi lịch; trả lời **sai** thì từ đó
đến hạn ngay (Study hỏi lại sớm). Máy tải sẵn 40 từ mỗi lần Sync nên mất mạng vẫn làm được.

## Tắt màn hình khi không dùng

| Không chạm | Máy làm gì |
|---|---|
| 45 giây | Giảm đèn nền |
| 90 giây | Tắt đèn nền (màn hình tối hẳn) |

Chạm màn hình hoặc nhấn BOOT để bật lại; cú chạm đánh thức **không** bấm nhầm nút bên dưới.
Wi-Fi vẫn chạy để gửi câu trả lời và nhận nhắc học. Đổi thời gian ở `DIM_AFTER_S` / `SLEEP_AFTER_S` trong `main/ui.c`.

## Chỉnh màn hình, cảm ứng, loa (menuconfig → E-Learn device)

| Hiện tượng | Tuỳ chọn |
|---|---|
| Màu bị âm bản | `Invert LCD colours` |
| Đỏ ↔ xanh dương bị đảo | `LCD uses BGR colour order` |
| Hình bị lật gương | `Mirror LCD X` |
| Hình ngược đầu | `Landscape the other way up` |
| Chạm bị lệch sang phía đối diện | `Mirror touch X` / `Mirror touch Y` |
| Không có tiếng | `Speaker amplifier is enabled by a LOW level on IO1` (đổi bật/tắt) |
| Muốn màn hình dọc | tắt `Landscape` |
| Đổi server mặc định | `Default server address` (cũng đổi được trong trang cài Wi-Fi) |

## Đọc từ vựng

- Thẻ hiện ra là máy đọc từ tiếng Anh một lần (giọng Microsoft Edge, server tạo và lưu sẵn). Nút
  **🔊** cạnh từ để nghe lại. Fill in the blank: đọc sau khi trả lời (để không lộ đáp án).
  Listening: nghe rồi chọn nghĩa.
- Tắt/bật: **Settings → Read words** trên máy. Âm lượng 0 thì không đọc.

## Giới hạn hiện tại

- **Chưa dùng mic** (server chưa có nhận giọng nói).
- Từ chưa có audio (server đang tạo, hoặc dịch vụ đọc lỗi) thì không đọc; lần đồng bộ sau sẽ có.
- Firmware đã build sạch nhưng **chưa chạy thử trên board thật**. Nếu treo/không lên hình, gửi log `idf.py monitor`.

## Cấu trúc

```
main/
  main.c       khởi động, vòng lặp: kết nối, ghép nối, đồng bộ, gửi câu trả lời, nhắc học
  board.c      chân GPIO, LCD + cảm ứng + LVGL, đèn nền, pin, ngủ màn hình
  audio.c      ES8311 + I2S: âm báo, phát WAV từ URL
  net.c        Wi-Fi STA / AP cài đặt, SNTP
  portal.c     trang cài Wi-Fi (captive portal + DNS)
  api.c        gọi /api/device/* (register, status, sync, reviews)
  store.c      NVS (cài đặt), cache bộ thẻ + câu trả lời chờ gửi trên SPIFFS
  ui*.c        giao diện LVGL: màn hình chính, cài đặt, học
  fonts/       DejaVu Sans (rút gọn: Latin, tiếng Việt, IPA) — giấy phép trong DEJAVU_LICENSE.txt
partitions.csv ota_0 / ota_1 4MB (firmware, cập nhật qua Wi-Fi), SPIFFS 2MB (/data), nvs ở cuối
```
