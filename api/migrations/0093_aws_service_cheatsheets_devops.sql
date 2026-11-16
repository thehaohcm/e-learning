-- 0093_aws_service_cheatsheets_devops: detailed, DevOps-oriented cheatsheet
-- sections per common AWS service (EC2, ECS/Fargate, CloudFront, S3, KMS,
-- Secrets Manager, Route 53, CloudWatch, CloudTrail, WAF) plus a monthly
-- client-report playbook. Focus: config, features, monitoring and the real
-- CLI commands a DevOps runs to manage and report on a client's system.
-- These feed the AWS Daily Lab scenarios. Same mechanism as 0092; seeded once
-- for this account, skipping titles that already exist.

INSERT INTO decks (subject_id, slug, name, description, default_kind, sort_order)
SELECT s.id, 'cheatsheet', 'Cheatsheet', $cs$AWS cheatsheet by section — source for AI questions and Daily Lab.$cs$, 'concept', 0
FROM subjects s WHERE s.slug = 'aws'
ON CONFLICT (subject_id, slug) DO NOTHING;

INSERT INTO course_notes (user_id, subject_id, title, body, sort_order, links, created_at, updated_at)
SELECT u.id, s.id, v.title, v.body, v.ord, '[]'::jsonb, NOW(), NOW()
FROM (VALUES
  ($cs$13. EC2 — vận hành & giám sát (DevOps)$cs$, $cs$**Cấu hình & vận hành**
- **Launch Template** (ưu tiên hơn Launch Config): AMI, instance type, key pair, SG, IAM instance profile, user-data, tag. Version hoá để rollback.
- **Truy cập không SSH**: dùng **SSM Session Manager** (cần SSM Agent + instance profile có `AmazonSSMManagedInstanceCore`) — không cần mở port 22, có log. Hạn chế key pair.
- **EBS**: gp3 (chọn IOPS/throughput riêng, rẻ hơn gp2), io2 cho DB. Bật **encryption by default** (KMS). Snapshot định kỳ bằng **DLM (Data Lifecycle Manager)**.
- **Auto Scaling Group**: min/desired/max, health check type = ELB, target tracking (giữ CPU ~50%), scale-in protection cho node quan trọng.
- **Patch**: SSM **Patch Manager** + Maintenance Window thay vì vá tay.

**Giám sát (CloudWatch)**
- Metric sẵn: `CPUUtilization`, `StatusCheckFailed` (_System/_Instance), `NetworkIn/Out`, EBS `VolumeReadOps`. **RAM & disk KHÔNG có sẵn** — phải cài **CloudWatch Agent** để đẩy `mem_used_percent`, `disk_used_percent`.
- Alarm nên có: StatusCheckFailed_Instance >= 1 (auto-recover), CPU > 85% 10', disk > 85%.

**CLI hay dùng**
```
aws ec2 describe-instances --filters Name=instance-state-name,Values=running \
  --query 'Reservations[].Instances[].{id:InstanceId,type:InstanceType,ip:PrivateIpAddress}'
aws ec2 stop-instances --instance-ids i-0abc
aws ec2 create-snapshot --volume-id vol-0abc --description "pre-change"
aws ssm start-session --target i-0abc
aws cloudwatch get-metric-statistics --namespace AWS/EC2 --metric-name CPUUtilization \
  --dimensions Name=InstanceId,Value=i-0abc --start-time ... --end-time ... --period 3600 --statistics Average
```
**Chi phí**: right-size bằng **Compute Optimizer**; Savings Plans / Reserved cho tải ổn định; tắt non-prod ngoài giờ (Instance Scheduler); dùng Spot cho batch.$cs$, 13),

  ($cs$14. ECS / Fargate — triển khai & scale$cs$, $cs$**Khái niệm**
- **Task Definition**: bản thiết kế container (image, CPU/mem, port, env, log config, secrets). **Task Role** = quyền app chạy; **Execution Role** = quyền kéo image từ ECR + đọc Secrets/SSM.
- **Service**: giữ N task chạy, gắn **ALB target group**, rolling update (minHealthy/maxPercent) hoặc blue/green qua **CodeDeploy**.
- **Fargate** (serverless, không quản node) vs **EC2 launch type** (tự quản cluster, rẻ hơn ở quy mô lớn).
- **ECR**: registry image; scan on push; lifecycle policy xoá image cũ.

**Scale & HA**
- **Service Auto Scaling**: target tracking theo `ECSServiceAverageCPUUtilization` hoặc ALB `RequestCountPerTarget`. Trải task qua nhiều AZ.
- Log: driver `awslogs` đẩy sang CloudWatch Logs; hoặc **FireLens** (Fluent Bit) sang nơi khác.

**Giám sát**: metric `CPUUtilization`, `MemoryUtilization` (namespace AWS/ECS), ALB `HTTPCode_Target_5XX`, `TargetResponseTime`, `UnHealthyHostCount`. Bật **Container Insights** để có metric chi tiết theo task.

**CLI**
```
aws ecs update-service --cluster prod --service web --desired-count 4
aws ecs update-service --cluster prod --service web --force-new-deployment
aws ecs describe-services --cluster prod --services web --query 'services[0].deployments'
aws ecs list-tasks --cluster prod --service-name web
aws logs tail /ecs/web --follow
```
**Chi phí**: Fargate Spot cho worker; right-size CPU/mem task; Savings Plans áp cho Fargate.$cs$, 14),

  ($cs$15. CloudFront — CDN & phân phối$cs$, $cs$**Cấu hình**
- **Distribution** với một hay nhiều **origin** (S3, ALB, custom). Với S3 dùng **OAC (Origin Access Control)** để chỉ CloudFront đọc bucket (bucket vẫn Block Public Access).
- **Cache policy** + **Origin request policy**: quản TTL, header/cookie/query nào được forward & cache. Bật compression.
- **SSL**: chứng chỉ từ **ACM ở us-east-1** (bắt buộc cho CloudFront) + alternate domain (CNAME). TLS tối thiểu 1.2.
- **Invalidation** khi deploy nội dung mới (`/*` tốn tiền, nên invalidation theo path hoặc versioned asset).
- **Signed URL / Signed Cookie** cho nội dung riêng tư; **WAF** gắn trực tiếp vào distribution; **Geo restriction** nếu cần.

**Giám sát**: `Requests`, `BytesDownloaded`, `4xxErrorRate`, `5xxErrorRate`, `OriginLatency`. Bật **standard/real-time logs** (sang S3/Kinesis). Alarm 5xxErrorRate > ngưỡng.

**CLI**
```
aws cloudfront create-invalidation --distribution-id E123 --paths '/index.html' '/app/*'
aws cloudfront get-distribution-config --id E123
aws cloudfront list-distributions --query 'DistributionList.Items[].{id:Id,domain:DomainName,enabled:Enabled}'
```
**Chi phí**: chọn Price Class phù hợp vùng khách; tăng cache hit ratio để giảm tải origin; dùng Origin Shield nếu nhiều edge.$cs$, 15),

  ($cs$16. S3 — quản trị, bảo mật & chi phí$cs$, $cs$**Bảo mật (ưu tiên số 1)**
- **Block Public Access** bật ở cả account lẫn bucket. Phân quyền bằng **bucket policy** + IAM; dùng **OAC** cho CloudFront thay vì public.
- Mã hoá mặc định **SSE-KMS** (kiểm soát key, có audit) hoặc SSE-S3. Bắt buộc TLS bằng policy `aws:SecureTransport`.
- **Versioning** + **MFA Delete** cho bucket quan trọng; **Object Lock** (WORM) cho log/audit bất biến.
- **Access Analyzer for S3** phát hiện bucket chia sẻ ngoài.

**Quản trị & chi phí**
- **Lifecycle**: chuyển sang Standard-IA/Glacier sau N ngày, xoá version cũ, dọn multipart dở. **Intelligent-Tiering** khi pattern truy cập khó đoán.
- **Replication (CRR/SRR)** cho DR; **S3 Storage Lens** để thấy dung lượng/chi phí theo bucket.
- **Server access logging / CloudTrail data events** để audit truy cập object.

**Giám sát**: `BucketSizeBytes`, `NumberOfObjects` (daily), 4xx/5xx request metrics (bật request metrics).

**CLI**
```
aws s3api get-public-access-block --bucket my-bkt
aws s3api put-public-access-block --bucket my-bkt --public-access-block-configuration \
  BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
aws s3api get-bucket-encryption --bucket my-bkt
aws s3api list-buckets --query 'Buckets[].Name'
aws s3 presign s3://my-bkt/file.pdf --expires-in 600
```$cs$, 16),

  ($cs$17. KMS — quản lý khóa mã hoá$cs$, $cs$**Khái niệm**
- **CMK / KMS key**: symmetric (mã hoá S3/EBS/RDS/Secrets) hoặc asymmetric (ký). **AWS-managed key** (aws/s3...) vs **customer-managed key** (tự đặt policy, rotation, audit — nên dùng cho dữ liệu nhạy cảm).
- **Key policy** là nguồn quyền chính (kết hợp IAM + grants). Nguyên tắc least privilege: tách người _dùng_ key và người _quản_ key.
- **Envelope encryption**: data key mã hoá dữ liệu, KMS mã hoá data key — hiệu năng tốt.
- **Rotation**: bật **automatic annual rotation** cho customer-managed key; alias trỏ key để xoay không đổi code.
- **Multi-Region keys** cho ứng dụng đa vùng; **grants** cấp quyền tạm cho service.

**Giám sát & audit**: mọi thao tác (`Encrypt`, `Decrypt`, `GenerateDataKey`) ghi ở **CloudTrail** — alarm khi `DisableKey`/`ScheduleKeyDeletion`. Theo dõi key sắp bị xoá (có waiting period 7-30 ngày).

**CLI**
```
aws kms list-keys
aws kms describe-key --key-id alias/app-data
aws kms enable-key-rotation --key-id <id>
aws kms get-key-rotation-status --key-id <id>
aws kms create-alias --alias-name alias/app-data --target-key-id <id>
```
**Chi phí**: mỗi CMK ~1 USD/tháng + phí request; gộp key hợp lý, tránh tạo thừa.$cs$, 17),

  ($cs$18. Secrets Manager — bí mật & xoay vòng$cs$, $cs$**Khái niệm**
- Lưu DB credential, API key... mã hoá bằng **KMS**. Có **rotation tự động** bằng Lambda (vd xoay mật khẩu RDS định kỳ).
- Version theo stage: **AWSCURRENT** (đang dùng), **AWSPREVIOUS** (trước đó), **AWSPENDING** (đang xoay).
- **Resource policy** giới hạn ai đọc được secret; app lấy qua SDK/CLI, không hardcode.
- So với **SSM Parameter Store**: Parameter Store (SecureString) rẻ/đơn giản cho config; Secrets Manager mạnh hơn ở rotation & cross-account.

**Vận hành an toàn**
- Cấp quyền `secretsmanager:GetSecretValue` tối thiểu theo ARN secret cụ thể.
- ECS/EKS inject secret qua `secrets` trong task definition (execution role cần quyền).
- Audit truy cập qua CloudTrail; alarm khi `GetSecretValue` bất thường.

**CLI**
```
aws secretsmanager list-secrets --query 'SecretList[].Name'
aws secretsmanager get-secret-value --secret-id prod/db --query SecretString --output text
aws secretsmanager rotate-secret --secret-id prod/db
aws secretsmanager describe-secret --secret-id prod/db --query 'RotationEnabled'
```
**Chi phí**: ~0.40 USD/secret/tháng + phí API — gom secret theo JSON nếu nhiều trường nhỏ.$cs$, 18),

  ($cs$19. Route 53 — DNS & định tuyến$cs$, $cs$**Khái niệm**
- **Hosted zone** (public/private) chứa record: A/AAAA, CNAME, MX, TXT, NS. **Alias record** (A/AAAA đặc biệt) trỏ tới ALB/CloudFront/S3 — miễn phí query, cho phép dùng ở zone apex (@).
- **Routing policy**: Simple; **Weighted** (chia %, canary); **Latency** (tới vùng nhanh nhất); **Failover** (primary/secondary + health check); **Geolocation/Geoproximity**; **Multivalue**.
- **Health check**: HTTP/HTTPS/TCP tới endpoint, kết hợp failover để tự chuyển DR. CloudWatch alarm làm health check được.
- **TTL**: thấp (60s) khi sắp đổi/HA, cao khi ổn định (tiết kiệm query).
- **DNSSEC** ký zone chống giả mạo; **Resolver** cho hybrid (on-prem).

**Giám sát**: metric health check `HealthCheckStatus`, `HealthCheckPercentageHealthy`; query logging sang CloudWatch Logs.

**CLI**
```
aws route53 list-hosted-zones
aws route53 list-resource-record-sets --hosted-zone-id Z123
aws route53 change-resource-record-sets --hosted-zone-id Z123 --change-batch file://change.json
aws route53 get-health-check-status --health-check-id abc
```
**Chi phí**: 0.50 USD/zone/tháng + phí query (alias tới AWS resource không tính query).$cs$, 19),

  ($cs$20. CloudWatch — metrics, alarms, logs, dashboards$cs$, $cs$**Metrics & Alarms**
- Metric theo **namespace/dimension**. Alarm: static threshold hoặc **anomaly detection**; `TreatMissingData` hợp lý; action gửi **SNS** (email/Slack) hoặc auto-scaling/EC2 recover.
- **Composite alarm** gộp nhiều alarm (giảm nhiễu). Custom metric từ app bằng `put-metric-data` hoặc EMF (Embedded Metric Format).

**Logs**
- **Log group** (đặt **retention**, mặc định không hết hạn → tốn tiền). **Metric filter** biến pattern log thành metric (vd đếm "ERROR") rồi alarm.
- **Logs Insights** truy vấn nhanh:
```
fields @timestamp, @message | filter @message like /ERROR/ | sort @timestamp desc | limit 50
```
- **Synthetics canary** giám sát endpoint từ ngoài; **RUM** cho client.

**Dashboards**: 1 dashboard/hệ thống gom EC2/ECS/ALB/RDS/CloudFront để báo cáo nhanh.

**CLI**
```
aws cloudwatch put-metric-alarm --alarm-name cpu-high --namespace AWS/EC2 \
  --metric-name CPUUtilization --dimensions Name=InstanceId,Value=i-0abc \
  --statistic Average --period 300 --evaluation-periods 2 --threshold 85 \
  --comparison-operator GreaterThanThreshold --alarm-actions arn:aws:sns:...:ops
aws logs put-retention-policy --log-group-name /ecs/web --retention-in-days 30
aws logs start-query --log-group-name /ecs/web --start-time ... --end-time ... --query-string '...'
```
**Chi phí**: custom metric, log ingestion/storage, Logs Insights scan — đặt retention, lọc trước khi đẩy.$cs$, 20),

  ($cs$21. CloudTrail — audit & điều tra$cs$, $cs$**Khái niệm**
- **Trail** ghi API call: **management events** (mặc định) và **data events** (S3 object, Lambda — bật riêng, tốn tiền). Nên tạo **multi-region, organization trail**.
- Ghi sang **S3** (bật **log file validation** chống sửa) + có thể song song sang **CloudWatch Logs** để alarm realtime.
- **CloudTrail Lake** / **Athena** để truy vấn lịch sử bằng SQL.
- **Insights** phát hiện bất thường (spike API).

**Alarm an ninh nên có** (metric filter trên CloudTrail→CW Logs):
- Root account login, `ConsoleLogin` thất bại nhiều lần.
- Thay đổi IAM policy, tạo access key, `AuthorizeSecurityGroupIngress` mở 0.0.0.0/0.
- `StopLogging`/xoá trail, `DisableKey` của KMS.

**CLI**
```
aws cloudtrail describe-trails
aws cloudtrail get-trail-status --name org-trail
aws cloudtrail lookup-events --lookup-attributes AttributeKey=EventName,AttributeValue=ConsoleLogin --max-results 20
```
**Chi phí**: trail đầu (management) miễn phí; data events & Insights tính phí — bật có chọn lọc.$cs$, 21),

  ($cs$22. WAF — tường lửa ứng dụng web$cs$, $cs$**Khái niệm**
- **Web ACL** gắn vào **CloudFront** (global, scope=CLOUDFRONT) hoặc **ALB/API Gateway/AppSync** (regional). Đánh giá rule theo thứ tự priority; mỗi rule **Allow/Block/Count**.
- **Rule types**: **AWS Managed Rule Groups** (Common/OWASP, SQLi, Known Bad Inputs, IP Reputation, Bot Control), **rate-based** (chặn khi 1 IP vượt N req/5'), **IP set** (allow/deny), **geo match**, **regex/size/string match**.
- **Count mode** để thử rule (đo tác động) trước khi chuyển Block — tránh chặn nhầm khách.
- **Logging** sang S3/Kinesis Firehose/CloudWatch Logs; xem request bị chặn và lý do.

**Giám sát**: metric `BlockedRequests`, `AllowedRequests`, `CountedRequests` theo rule; **Sampled requests** trong console để điều tra. Alarm khi BlockedRequests tăng đột biến (có thể đang bị tấn công) hoặc khi rule mới chặn nhầm nhiều.

**CLI**
```
aws wafv2 list-web-acls --scope REGIONAL
aws wafv2 get-web-acl --name app-acl --scope REGIONAL --id <id>
aws wafv2 get-sampled-requests --web-acl-arn <arn> --rule-metric-name <rule> \
  --scope REGIONAL --time-window ... --max-items 100
```
**Chi phí**: phí/web ACL + phí/rule + phí/triệu request; Bot Control & một số managed group tính thêm.$cs$, 22),

  ($cs$23. Báo cáo hằng tháng cho client$cs$, $cs$Mục tiêu: mỗi tháng tổng hợp **chi phí, sức khoẻ hệ thống, bảo mật và việc đã làm** để gửi client. Gợi ý khung báo cáo & nguồn số liệu:

**1) Chi phí & tối ưu**
- **Cost Explorer**: tổng chi phí tháng, so tháng trước, top dịch vụ, chi phí theo **tag** (Project/Env/Owner). `ce get-cost-and-usage`.
- **Budgets**: ngân sách & cảnh báo vượt. **Compute Optimizer** / Trusted Advisor: khuyến nghị right-size, tài nguyên nhàn rỗi.

**2) Sức khoẻ & độ sẵn sàng**
- Uptime/latency từ CloudWatch (ALB `TargetResponseTime`, 5xx; EC2/ECS CPU/mem; RDS CPU/connections/free storage).
- Sự cố/alarm đã bắn trong tháng (từ SNS/alarm history), thời gian khắc phục.
- Backup/snapshot & kết quả test khôi phục (nếu có).

**3) Bảo mật & tuân thủ**
- **Security Hub** / **GuardDuty**: findings theo mức độ. **AWS Config**: tài nguyên không tuân thủ (S3 public, SG mở, chưa mã hoá). **IAM Access Analyzer**: chia sẻ ngoài. **Trusted Advisor** (security).
- CloudTrail: sự kiện nhạy cảm (root login, đổi IAM/SG).

**4) Thay đổi đã thực hiện**: deploy, vá lỗi, thay đổi hạ tầng (nên gắn với change ticket).

**Tự động hoá**: tag chuẩn hoá để bóc tách chi phí; script/CLI xuất số liệu; cân nhắc **AWS Cost & Usage Report (CUR) + Athena/QuickSight** cho dashboard; **Config conformance pack** cho PCI/SOC2.
```
aws ce get-cost-and-usage --time-period Start=2026-10-01,End=2026-11-01 \
  --granularity MONTHLY --metrics UnblendedCost --group-by Type=DIMENSION,Key=SERVICE
aws securityhub get-findings --filters '{"SeverityLabel":[{"Value":"HIGH","Comparison":"EQUALS"}]}' --max-results 20
aws configservice describe-compliance-by-config-rule
aws support describe-trusted-advisor-checks --language en
```
$cs$, 23)
) AS v(title, body, ord)
JOIN users u ON u.email = 'thehaohcm@tradesign.io.vn'
JOIN subjects s ON s.slug = 'aws'
WHERE NOT EXISTS (
    SELECT 1 FROM course_notes cn
    WHERE cn.user_id = u.id AND cn.subject_id = s.id AND cn.title = v.title
);
