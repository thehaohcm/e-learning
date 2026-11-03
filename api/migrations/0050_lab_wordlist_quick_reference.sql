-- 0050_lab_wordlist_quick_reference.sql: append a "wordlist nen nho" quick
-- reference (absolute paths, when/where to use each) to Lab 3 (web enum) and
-- Lab 9 (password attacks) reference fields, per explicit user request after
-- they said there are too many wordlists with long unmemorable names spread
-- across different sources. Full original reference text (from
-- 0043_comprehensive_lab_and_assignment_reference_notes.sql, the latest
-- migration that set these fields) is preserved verbatim with the new
-- section appended, since jsonb_set on '{reference}' REPLACES the whole
-- field rather than appending.

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Web Enumeration: Dò tìm tài nguyên ẩn (WEB1 / PT1)

### 1. Web Enumeration là gì? Tại sao bắt buộc phải làm trước khi exploit?
Trang web hiển thị cho người dùng thông thường chỉ là "bề nổi của tảng băng chìm". Phía sau đó thường có các trang quản trị (`/admin`), các tệp sao lưu do lập trình viên để quên (`config.php.bak`, `backup.zip`), các endpoint API (`/api/v1`), hoặc thư mục mã nguồn lộ lọt (`/.git/`).
Nếu không dùng công cụ tự động hóa dò tìm danh mục (Directory Brute-force), bạn sẽ không thể tìm thấy các bề mặt tấn công này.

### 2. Các công cụ dò quét & giải thích tham số

#### A. `ffuf` (Fuzz Faster U Fool)
* Viết bằng ngôn ngữ Go, có tốc độ cực nhanh, là công cụ fuzzer chuẩn hiện nay.
* **Cú pháp:** `ffuf -w common.txt -u http://10.10.10.5/FUZZ -e .php,.txt,.bak`
* **`-w <file>`:** Đường dẫn đến tệp danh sách từ khóa (wordlist).
* **`-u <url>`:** URL mục tiêu, trong đó từ khóa `FUZZ` sẽ được thay thế tuần tự bằng từng dòng trong wordlist.
* **`-e .php,.txt`:** Mở rộng phần đuôi file. Với từ `admin`, ffuf sẽ thử cả `admin`, `admin.php`, `admin.txt`.

#### B. `gobuster`
* Chuyên dụng cho brute-force đường dẫn web và tên miền phụ (subdomain).
* **Cú pháp:** `gobuster dir -u http://10.10.10.5 -w common.txt -x php,txt,zip`
* `dir`: Chế độ quét thư mục/tệp tin.
* `-x`: Danh sách extension mở rộng.

#### C. `curl` & `nikto`
* `curl -I http://10.10.10.5`: Cờ `-I` (HEAD request) chỉ lấy Header mà không tải toàn bộ nội dung HTML. Rất nhanh để kiểm tra Web server đang chạy (Nginx/Apache), công nghệ backend qua header `X-Powered-By` (PHP/NodeJS/ASP.NET).
* `nikto -h 10.10.10.5`: Công cụ quét lỗ hổng cấu hình web tự động (cảnh báo header thiếu bảo mật, file test mặc định của server).

### 3. Wordlist nên nhớ (đường dẫn tuyệt đối, dùng với -w)
* `/usr/share/seclists/Discovery/Web-Content/common.txt` — nhỏ (~4.7k dòng), chạy trước tiên để có kết quả nhanh.
* `/usr/share/seclists/Discovery/Web-Content/raft-medium-directories.txt` — khi common.txt không ra gì, dò sâu hơn.
* `/usr/share/seclists/Discovery/Web-Content/directory-list-2.3-medium.txt` — list kinh điển của DirBuster, hay được writeup/box cũ trích dẫn.
* `/usr/share/seclists/Discovery/DNS/subdomains-top1million-5000.txt` — mặc định khi dò subdomain/vhost.
* `/usr/share/seclists/Discovery/Web-Content/burp-parameter-names.txt` — khi nghi web có tham số ẩn (debug=true, admin=1...).
* Đuôi file hay thử kèm `-x`/`-e`: `php,txt,html,bak,zip,old,config`.
* Không nhớ tên chính xác? Gõ `ls /usr/share/seclists/Discovery/Web-Content/ | grep -i <từ khóa>` hoặc `locate seclists | grep -i <từ khóa>`.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 3 — Web enumeration cơ bản';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Password Attacks: Online Brute-force & Offline Hash Cracking (PT1 / CPTS)

### 1. Phân biệt Tấn công trực tuyến (Online) và Ngoại tuyến (Offline)
* **Online Attack (Trực tuyến):** Gửi từng cặp username/mật khẩu qua mạng tới dịch vụ thật (SSH, FTP, Web form) bằng công cụ như `hydra`. Nhược điểm: Chậm (giới hạn bởi độ trễ mạng) và dễ bị phát hiện bởi hệ thống ghi log hoặc bị chặn bởi chính sách khóa tài khoản (Account Lockout).
* **Offline Attack (Ngoại tuyến):** Sau khi đã đánh cắp được file băm mật khẩu (file hash từ database hoặc SAM/NTDS.dit), kẻ tấn công đem file này về máy cá nhân của mình và dùng sức mạnh của GPU/CPU chạy `hashcat` hoặc `john` để thử hàng tỷ mật khẩu mỗi giây. Ưu điểm: Tốc độ cực nhanh và máy chủ nạn nhân **hoàn toàn không hề hay biết**.

### 2. Công cụ `hydra`
* `hydra -l admin -P passwords.txt ssh://10.10.10.5`
* **Quy ước cờ hoa/thường:**
  * `-l`: Chỉ định 1 username cố định.
  * `-L`: Nạp danh sách nhiều username từ file.
  * `-p`: Chỉ định 1 mật khẩu cố định.
  * `-P`: Nạp danh sách wordlist mật khẩu (như `rockyou.txt`).

### 3. Công cụ `hashcat`
* Hashcat là bộ bẻ khóa mật khẩu tận dụng sức mạnh card đồ họa (GPU) nhanh nhất thế giới.
* **Cú pháp:** `hashcat -m <mode> <hashfile> <wordlist>`
* **Bảng mode Hashcat quan trọng nhất:**
  * `-m 0`: MD5
  * `-m 1000`: NTLM (SAM, NTDS.dit) — **dùng được cho Pass-the-Hash**
  * `-m 5600`: NetNTLMv2 (bắt qua mạng bằng Responder) — **không Pass-the-Hash được**
  * `-m 13100`: Kerberoasting (vé TGS-REP)
  * `-m 18200`: AS-REP Roasting

### 4. Wordlist nên nhớ (đường dẫn tuyệt đối)
* `/usr/share/seclists/Usernames/top-usernames-shortlist.txt` — username phổ biến (admin, administrator, root...) khi chưa biết username thật.
* `/usr/share/wordlists/rockyou.txt` trên Kali (cần `gunzip /usr/share/wordlists/rockyou.txt.gz` trước nếu chưa giải nén) — wordlist password kinh điển nhất, dùng chung với hydra/hashcat/john trong gần như mọi bài THM/HTB dạng "crack password yếu".'::text), true)
WHERE kind = 'lab' AND front = 'Lab 9 — Password attacks (hydra / hashcat / john)';
