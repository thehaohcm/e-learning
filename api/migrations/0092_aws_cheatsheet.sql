-- 0092_aws_cheatsheet: a starter Cheatsheet for the AWS subject — Well-
-- Architected, IAM, VPC/networking, EC2, S3, RDS, caching/streaming, observability,
-- security & PCI/SOC2, Cloudflare-in-front, CLI and cost. Same mechanism as the
-- PMP/Pentest cheatsheets. Seeded once for this account, skipping existing titles.

INSERT INTO decks (subject_id, slug, name, description, default_kind, sort_order)
SELECT s.id, 'cheatsheet', 'Cheatsheet', $cs$AWS cheatsheet by section (Well-Architected, IAM, VPC, EC2, S3, RDS, security/PCI, Cloudflare, CLI, cost). Editable, studyable, source for AI questions.$cs$, 'concept', 0
FROM subjects s WHERE s.slug = 'aws'
ON CONFLICT (subject_id, slug) DO NOTHING;

INSERT INTO course_notes (user_id, subject_id, title, body, sort_order, links, created_at, updated_at)
SELECT u.id, s.id, v.title, v.body, v.ord, '[]'::jsonb, NOW(), NOW()
FROM (VALUES
  ($cs$1. Tổng quan & Well-Architected$cs$, $cs$**AWS Well-Architected — 6 trụ cột**:
1. **Operational Excellence** — vận hành, tự động hóa, giám sát.
2. **Security** — least privilege, mã hóa, phòng thủ nhiều lớp.
3. **Reliability** — Multi-AZ, auto-recovery, backup.
4. **Performance Efficiency** — đúng loại tài nguyên, scale hợp lý.
5. **Cost Optimization** — right-sizing, reserved/savings plan, tắt tài nguyên không dùng.
6. **Sustainability** — tối ưu năng lượng.

**Region vs AZ**: Region = vùng địa lý; AZ = data center độc lập trong region (luôn deploy đa-AZ cho HA).$cs$, 1),
  ($cs$2. IAM — danh tính & quyền$cs$, $cs$- **User**: người/ứng dụng cụ thể; **Group**: gom user; **Role**: danh tính tạm (EC2/Lambda/cross-account assume).
- **Policy** (JSON): Effect (Allow/Deny), Action, Resource, Condition. **Deny luôn thắng Allow**.
- **Least privilege**: cấp đúng quyền cần thiết. Tránh dùng root; bật **MFA** cho root + user quan trọng.
- **Instance profile**: gắn Role vào EC2 để app lấy credential tạm — KHÔNG hardcode access key.
- **Access Analyzer** để soát quyền public/cross-account.$cs$, 2),
  ($cs$3. VPC & Networking$cs$, $cs$- **VPC**: mạng ảo riêng (CIDR, vd 10.0.0.0/16). **Subnet**: chia theo AZ. Public subnet có route ra **IGW**; private subnet ra internet qua **NAT Gateway**.
- **Route table**: định tuyến; **IGW** (Internet Gateway) cho public; **NAT** cho private ra ngoài.
- **Security Group (SG)**: stateful, gắn ở ENI/instance, chỉ có Allow. **NACL**: stateless, ở subnet, có Allow+Deny.
- **VPC Endpoint**: truy cập dịch vụ AWS (S3…) không qua internet.
- **Peering / Transit Gateway**: nối nhiều VPC.$cs$, 3),
  ($cs$4. EC2 & Compute$cs$, $cs$- **Instance**: máy ảo; chọn family theo nhu cầu (t = burstable, m = cân bằng, c = CPU, r = RAM).
- **AMI**: image để khởi tạo. **User data**: script chạy lúc boot (cài đặt tự động).
- **EBS**: ổ block bền (gp3/io2); snapshot để backup. **Instance store**: tạm, mất khi stop.
- **Auto Scaling Group + ELB**: co giãn + phân tải. **Launch Template** định nghĩa cấu hình.
- **Pricing**: On-Demand, Reserved/Savings Plan (cam kết 1-3 năm rẻ hơn), Spot (rẻ nhất, có thể bị thu hồi).$cs$, 4),
  ($cs$5. S3 — lưu trữ object$cs$, $cs$- **Bucket** (tên toàn cầu duy nhất) chứa **object**. Bền 11 số 9 (durability).
- **Storage class**: Standard, Intelligent-Tiering, Standard-IA, Glacier (lưu trữ lạnh, rẻ).
- **Bảo mật**: mặc định private; **Block Public Access** bật. Phân quyền qua **bucket policy** / IAM. Bật **versioning** + **MFA delete**.
- **Mã hóa**: SSE-S3, **SSE-KMS** (kiểm soát key), SSE-C.
- **Lifecycle**: tự chuyển class / xóa sau N ngày. **Presigned URL**: cấp quyền tạm.$cs$, 5),
  ($cs$6. RDS & Databases$cs$, $cs$- **RDS**: DB quản lý (PostgreSQL/MySQL…). **Multi-AZ**: bản standby đồng bộ ở AZ khác → tự failover (HA). **Read Replica**: bản đọc (scale đọc, async).
- **Subnet Group**: đặt RDS trong private subnet; chỉ cho SG của app truy cập.
- **Backup**: automated backup (PITR) + snapshot thủ công. Mã hóa at-rest bằng **KMS**.
- **Aurora**: tương thích MySQL/Postgres, hiệu năng cao, auto-scaling storage.
- **DynamoDB**: NoSQL key-value, serverless, scale lớn.$cs$, 6),
  ($cs$7. ElastiCache & Kafka (MSK)$cs$, $cs$- **ElastiCache**: cache in-memory trong VPC — **Redis** (có persistence, pub/sub, sorted set) hoặc **Memcached** (đơn giản, đa luồng). Giảm tải DB, tăng tốc đọc.
- Đặt trong private subnet; truy cập qua SG.
- **MSK (Managed Kafka)**: streaming/message bus quản lý. Dùng cho event-driven, log pipeline, tách service.
- Cân nhắc: cache-aside pattern (app đọc cache trước, miss thì đọc DB rồi set cache).$cs$, 7),
  ($cs$8. CloudWatch & CloudTrail$cs$, $cs$- **CloudWatch**: **Metrics** (CPU, latency…), **Logs** (log group/stream), **Alarms** (cảnh báo → SNS/auto scale), **Dashboards**, **Agent** (đẩy log/metric từ EC2).
- **CloudTrail**: ghi lại **API call** (ai làm gì, khi nào) — phục vụ audit/forensic. Bật **log file validation** để chống sửa log.
- **Log Integrity**: gửi log sang S3 (immutable, Object Lock) / tách account để không bị xóa.
- Phân biệt: CloudWatch = vận hành/hiệu năng; CloudTrail = audit hành vi.$cs$, 8),
  ($cs$9. Bảo mật & PCI-DSS / SOC 2$cs$, $cs$- **Mã hóa**: at-rest (KMS cho S3/RDS/EBS), in-transit (TLS). Quản lý key bằng **KMS**; bí mật bằng **Secrets Manager** (tự xoay vòng).
- **Phân tách mạng**: thẻ/PII trong private subnet, không public; WAF + SG chặt.
- **PCI-DSS**: giảm phạm vi (scope) — tách môi trường chứa dữ liệu thẻ; log đầy đủ (CloudTrail), review quyền định kỳ, không lưu CVV.
- **SOC 2**: kiểm soát access, change management, monitoring, encryption — AWS cung cấp báo cáo qua **Artifact**.
- Nguyên tắc: least privilege + audit + mã hóa + immutable log.$cs$, 9),
  ($cs$10. Cloudflare trước AWS$cs$, $cs$- **DNS**: Cloudflare quản lý bản ghi, proxy (cam cam) để ẩn IP origin.
- **SSL modes**: *Flexible* (CF↔client HTTPS, CF↔origin HTTP — KHÔNG an toàn), *Full* (HTTPS cả 2 nhưng cert origin không kiểm), **Full (strict)** (khuyến nghị — origin có cert hợp lệ, vd Cloudflare Origin Certificate).
- **WAF**: chặn tấn công tầng 7 (SQLi, XSS), rate limiting, bot management.
- **Bảo vệ origin**: chỉ cho IP Cloudflare vào origin (SG/allowlist) để tránh bypass.$cs$, 10),
  ($cs$11. AWS CLI — lệnh thường dùng$cs$, $cs$```
aws configure                      # cấu hình key/region/output
aws sts get-caller-identity        # đang là ai (account/role)
aws ec2 describe-instances         # liệt kê EC2
aws s3 ls                          # liệt kê bucket
aws s3 cp file s3://bucket/        # upload
aws s3 sync ./dir s3://bucket/     # đồng bộ thư mục
aws rds describe-db-instances      # thông tin RDS
aws logs tail /aws/group --follow  # xem log realtime
aws iam list-roles                 # liệt kê role
```
Dùng `--profile`, `--region`, `--query` (JMESPath) để lọc output.$cs$, 11),
  ($cs$12. Chi phí & best practices$cs$, $cs$- **Right-sizing**: chọn instance/RDS vừa đủ; dùng Compute Optimizer.
- **Savings Plan / Reserved** cho tải ổn định; **Spot** cho batch chịu gián đoạn.
- **Tắt tài nguyên dev** ngoài giờ; xóa EBS/snapshot/EIP không dùng.
- **S3 lifecycle** chuyển sang IA/Glacier; **Intelligent-Tiering** khi pattern không rõ.
- **Budgets + Cost Explorer + tag** để theo dõi chi phí theo dự án.
- **Tag chuẩn**: Project, Env, Owner — để phân bổ chi phí & tự động hóa.$cs$, 12)
) AS v(title, body, ord)
JOIN users u ON u.email = 'thehaohcm@tradesign.io.vn'
JOIN subjects s ON s.slug = 'aws'
WHERE NOT EXISTS (
    SELECT 1 FROM course_notes cn
    WHERE cn.user_id = u.id AND cn.subject_id = s.id AND cn.title = v.title
);
