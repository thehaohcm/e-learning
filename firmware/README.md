# Firmware thiết bị học ES3C28P (ESP32-S3, 2.8", LVGL)

Thiết bị để bàn: ôn thẻ (flashcard / quiz / listening / cloze / mcq) bằng cách **chạm chọn**, không gõ chữ.
Dữ liệu lấy từ server qua `/api/device/*`, lưu cache để học được khi mất mạng; câu trả lời gửi lại khi có mạng.

- Phần cứng: ESP32-S3 16MB flash + 8MB PSRAM, LCD ILI9341 240×320 (SPI), cảm ứng FT6336 (I2C),
  codec ES8311 + loa (I2S), đèn nền IO45, pin đo qua ADC.
- Phần mềm: ESP-IDF **5.4**, LVGL **9.3**, esp_lvgl_port, font DejaVu (có tiếng Việt và IPA).

## Nạp nhanh (không cần cài ESP-IDF)

1. Mở https://espressif.github.io/esptool-js/ bằng Chrome/Edge, cắm board qua USB → **Connect**.
2. Chọn file `prebuilt/elearn-device-1.0.0.bin`, địa chỉ **0x0** → **Program**.
3. Nhấn RST.

File prebuilt dùng cấu hình mặc định (ngang, màu IPS). Nếu màu/cảm ứng bị sai, xem mục *Chỉnh màn hình* rồi tự build.

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

Đổi Wi-Fi: **Settings → Change Wi-Fi** trên máy, hoặc **giữ nút BOOT khi cắm điện**.
Huỷ liên kết: **Settings → Unlink** (hoặc xoá thiết bị trên website; máy sẽ tự hiện mã ghép nối mới).

## Học

- **Study**: thẻ đến hạn trước, sau đó thẻ mới (số lượng theo cài đặt trên website).
  - Flashcard: chạm *Show answer* → chọn Again / Hard / Good / Easy.
  - Quiz / listening / cloze / mcq: chọn 1 trong 4. Đúng = Good, sai = Again, có âm báo.
- Câu trả lời lưu trong máy (tối đa 400) và gửi lên server mỗi phút; thẻ mới tải về mỗi 10 phút
  hoặc khi bấm **Sync**. Mất mạng vẫn học được trên bộ thẻ đã tải.
- **Nhắc học**: trong khung giờ đặt trên website, nếu có thẻ chưa làm và máy đã rảnh đủ lâu
  (*reminder every N min*), máy tự sáng lên, kêu "ting" và đưa ra 1 thẻ.

## Tắt màn hình khi không dùng

| Không chạm | Máy làm gì |
|---|---|
| 45 giây | Giảm đèn nền |
| 90 giây | Tắt đèn nền, cho LCD ngủ, dừng vẽ LVGL, hạ CPU xuống 80 MHz |

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

## Giới hạn hiện tại

- **Chưa dùng mic** (server chưa có nhận giọng nói).
- **Listening** chỉ phát được khi thẻ có `audio_url` dạng **WAV 16-bit**; thẻ không có audio thì hiện như quiz.
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
partitions.csv app 6MB + SPIFFS 2MB (/data)
```
