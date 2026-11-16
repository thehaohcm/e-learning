package ai

import (
	"context"
	"fmt"
	"strings"
)

// MistakeItem is one question the learner got wrong in a Daily set.
type MistakeItem struct {
	Stem        string   `json:"stem"`
	Correct     []string `json:"correct"`
	Picked      []string `json:"picked"`
	Explanation string   `json:"explanation"`
	Section     string   `json:"section"`
	Terms       []string `json:"terms"`
	Note        string   `json:"note"`
	Right       bool     `json:"right"` // true = câu đúng nhưng học viên có ghi chú muốn AI nhận xét
}

// ReviewDailyMistakes writes a short Vietnamese study review of the questions the
// learner got wrong, grouped by concept, grounded ONLY in the given items.
func (c *Client) ReviewDailyMistakes(ctx context.Context, model, courseName string, items []MistakeItem) (string, error) {
	if !c.Enabled() {
		return "", ErrNotConfigured
	}
	if len(items) == 0 {
		return "", nil
	}
	var b strings.Builder
	for i, it := range items {
		tag := "SAI"
		if it.Right {
			tag = "ĐÚNG, có ghi chú"
		}
		fmt.Fprintf(&b, "%d. [%s] Câu hỏi: %s\n", i+1, tag, strings.TrimSpace(it.Stem))
		if len(it.Correct) > 0 {
			fmt.Fprintf(&b, "   - Đáp án đúng: %s\n", strings.Join(it.Correct, " | "))
		}
		if !it.Right && len(it.Picked) > 0 {
			fmt.Fprintf(&b, "   - Học viên đã chọn (SAI): %s\n", strings.Join(it.Picked, " | "))
		}
		if strings.TrimSpace(it.Section) != "" {
			fmt.Fprintf(&b, "   - Chủ đề: %s\n", it.Section)
		}
		if strings.TrimSpace(it.Explanation) != "" {
			fmt.Fprintf(&b, "   - Giải thích đúng: %s\n", it.Explanation)
		}
		if strings.TrimSpace(it.Note) != "" {
			fmt.Fprintf(&b, "   - Ghi chú của học viên: %s\n", strings.TrimSpace(it.Note))
		}
		b.WriteString("\n")
	}
	lc := strings.ToLower(courseName)
	isPentest := strings.Contains(lc, "pentest") || strings.Contains(lc, "pen test")
	isAws := strings.Contains(lc, "aws")
	system := "Bạn là trợ giảng luyện thi PMP. Trả lời HOÀN TOÀN bằng tiếng Việt, chỉ dựa trên dữ liệu được cung cấp, không bịa kiến thức ngoài."
	rightPrinciple := `nguyên tắc/"tư duy PMI" đúng`
	groupHint := "theo CONCEPT / chủ đề PMP (vd: Scope trong Agile, EVM, vai trò Product Owner, Change Control...)"
	keepTerms := "Giữ thuật ngữ tiếng Anh (Product Owner, sprint backlog, baseline...) nhưng diễn giải bằng tiếng Việt."
	if isPentest {
		system = "Bạn là mentor pentest/bảo mật (THM PT1 / HTB CPTS). Trả lời HOÀN TOÀN bằng tiếng Việt, chỉ dựa trên dữ liệu được cung cấp, không bịa. Giữ nguyên lệnh, cờ và thuật ngữ tiếng Anh."
		rightPrinciple = "tư duy/methodology pentest đúng (enum trước khi khai thác, đọc kết quả, khi nào pivot, lệnh/công cụ phù hợp)"
		groupHint = "theo CONCEPT / giai đoạn pentest (vd: enum dịch vụ, web exploit, privilege escalation Linux/Windows, pivoting...)"
		keepTerms = "Giữ nguyên lệnh, cờ và thuật ngữ tiếng Anh (nmap, SeImpersonatePrivilege, LinPEAS...), diễn giải bằng tiếng Việt."
	}
	if isAws {
		system = "Bạn là mentor AWS DevOps/SRE giàu kinh nghiệm, chuyên vận hành hệ thống khách hàng (EC2, ECS, S3, CloudFront, KMS, Route53, CloudWatch, CloudTrail, WAF, Secrets Manager). Trả lời HOÀN TOÀN bằng tiếng Việt, chỉ dựa trên dữ liệu được cung cấp, không bịa. Giữ nguyên lệnh AWS CLI, tên dịch vụ và thuật ngữ tiếng Anh."
		rightPrinciple = "tư duy vận hành AWS đúng (điều tra bằng console/CLI trước khi sửa, least privilege, Multi-AZ/HA, mã hóa, tối ưu chi phí, remediate an toàn và có kiểm chứng)"
		groupHint = "theo dịch vụ / chủ đề AWS (vd: bảo mật S3, scale ECS/EC2, cache CloudFront, KMS rotation, alarm CloudWatch, audit CloudTrail, chặn WAF...)"
		keepTerms = "Giữ nguyên lệnh AWS CLI, tên dịch vụ (aws s3api, CloudFront, KMS...) và thuật ngữ tiếng Anh, diễn giải bằng tiếng Việt."
	}
	user := fmt.Sprintf(`Môn: %s. Học viên vừa làm bài Daily. Dưới đây là các câu SAI, và một số câu ĐÚNG nhưng học viên có GHI CHÚ muốn được nhận xét:

%s
Viết một bản đánh giá NGẮN GỌN bằng tiếng Việt giúp học viên nhớ lỗi và tránh lặp lại:
- Với câu [SAI]: chỉ ra học viên hiểu sai gì và %s.
- Với câu [ĐÚNG, có ghi chú]: KHÔNG coi là sai; chỉ phản hồi vào ghi chú của học viên (xác nhận cách hiểu đúng hoặc bổ sung/chỉnh nếu ghi chú chưa chuẩn).
- Nếu một câu có Ghi chú của học viên, luôn phản hồi trực tiếp vào ghi chú đó.
- Gom các câu sai %s.
- Mỗi nhóm: nêu học viên đang hiểu sai điều gì và %s cần nhớ (1-3 câu).
- Kết thúc bằng mục "## Cần ôn lại" gồm 3-6 gạch đầu dòng cụ thể, bám theo Chủ đề/Cheatsheet ở trên.
Dùng markdown (##, **, -). %s KHÔNG chép lại nguyên văn câu hỏi.`, courseName, b.String(), rightPrinciple, groupHint, rightPrinciple, keepTerms)
	return c.chatText(ctx, model, system, user, 0.4)
}
