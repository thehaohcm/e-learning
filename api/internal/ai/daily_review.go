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
		fmt.Fprintf(&b, "%d. Câu hỏi: %s\n", i+1, strings.TrimSpace(it.Stem))
		if len(it.Correct) > 0 {
			fmt.Fprintf(&b, "   - Đáp án đúng: %s\n", strings.Join(it.Correct, " | "))
		}
		if len(it.Picked) > 0 {
			fmt.Fprintf(&b, "   - Học viên đã chọn (SAI): %s\n", strings.Join(it.Picked, " | "))
		}
		if strings.TrimSpace(it.Section) != "" {
			fmt.Fprintf(&b, "   - Chủ đề: %s\n", it.Section)
		}
		if strings.TrimSpace(it.Explanation) != "" {
			fmt.Fprintf(&b, "   - Giải thích đúng: %s\n", it.Explanation)
		}
		b.WriteString("\n")
	}
	system := "Bạn là trợ giảng luyện thi PMP. Trả lời HOÀN TOÀN bằng tiếng Việt, chỉ dựa trên dữ liệu được cung cấp, không bịa kiến thức ngoài."
	user := fmt.Sprintf(`Môn: %s. Học viên vừa làm bài Daily và SAI những câu dưới đây:

%s
Viết một bản đánh giá NGẮN GỌN bằng tiếng Việt giúp học viên nhớ lỗi và tránh lặp lại:
- Gom các câu sai theo CONCEPT / chủ đề PMP (vd: Scope trong Agile, EVM, vai trò Product Owner, Change Control...).
- Mỗi nhóm: nêu học viên đang hiểu sai điều gì và "tư duy PMI" / nguyên tắc đúng cần nhớ (1-3 câu).
- Kết thúc bằng mục "## Cần ôn lại" gồm 3-6 gạch đầu dòng cụ thể, bám theo Chủ đề/Cheatsheet ở trên.
Dùng markdown (##, **, -). Giữ thuật ngữ tiếng Anh (Product Owner, sprint backlog, baseline...) nhưng diễn giải bằng tiếng Việt. KHÔNG chép lại nguyên văn câu hỏi.`, courseName, b.String())
	return c.chatText(ctx, model, system, user, 0.4)
}
