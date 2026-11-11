package ai

import (
	"context"
	"fmt"
	"strings"
)

// Roleplay: the learner plays a project manager talking (in English) with a
// difficult client. The AI plays the client turn by turn, then reviews the
// conversation: was the PM's handling right by PMI standards, and how good was
// the English. Scenarios are fixed (below) so they tie to PMP cheatsheet topics.

// RoleplayTurn is one line of the conversation. Role is "client" or "pm".
type RoleplayTurn struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

// RoleplayScenario is shown to the learner (everything except Client).
type RoleplayScenario struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Setup   string `json:"setup"`    // English: the situation the PM is in
	SetupVI string `json:"setup_vi"` // Vietnamese help for the same
	Goal    string `json:"goal"`     // what a good PM should achieve (English)
	Topics  string `json:"topics"`   // PMP topics practised
	Client  string `json:"-"`        // hidden brief for the AI playing the client
}

// RoleplayPersona is how difficult the client is.
type RoleplayPersona struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Desc     string `json:"desc"`
	Behavior string `json:"-"`
}

var RoleplayPersonas = []RoleplayPersona{
	{ID: "firm", Label: "Calm but firm", Desc: "Polite, direct, keeps pressing until the answer is clear.",
		Behavior: "You are polite and professional but firm. You keep asking until you get a clear, specific answer."},
	{ID: "impatient", Label: "Impatient", Desc: "Short sentences, wants answers now, cuts the PM off.",
		Behavior: "You are impatient. You use short sentences, interrupt long explanations and say things like 'Get to the point.' You want the bottom line first."},
	{ID: "angry", Label: "Frustrated / angry", Desc: "Blames the team, raises the tension. Calms down only with respect and a concrete plan.",
		Behavior: "You are frustrated and a bit angry. You complain, blame the project team and show disappointment. You calm down only when the PM stays respectful, takes ownership and gives a concrete plan."},
	{ID: "skeptic", Label: "Skeptical / detail-oriented", Desc: "Challenges numbers and asks 'why' and 'how do you know?'.",
		Behavior: "You are skeptical and detail-oriented. You challenge numbers, ask 'why?' and 'how do you know?', and do not accept claims without evidence."},
}

var RoleplayScenarios = []RoleplayScenario{
	{
		ID:      "change-request",
		Title:   "Client wants extra scope, same deadline",
		Setup:   "You are the project manager of a customer-portal project, six weeks before go-live. Daniel Brooks, the client's VP of Operations, has asked for a meeting.",
		SetupVI: "Bạn là PM dự án cổng khách hàng, còn 6 tuần nữa go-live. Daniel Brooks (VP Operations của client) muốn họp. Anh ấy sẽ đòi thêm tính năng mà không đổi deadline.",
		Goal:    "Handle the change request properly: acknowledge the need, clarify it, explain the impact on scope, schedule, cost and risk, offer options, and explain the change-control process. Do not simply say yes or no.",
		Topics:  "Integrated change control, scope creep, stakeholder management",
		Client:  "You are Daniel Brooks, VP of Operations at the client company. Your team now wants a new real-time reporting dashboard added to the portal. You expect it before the existing go-live date in six weeks, at no extra cost, because you think it should be simple. You do not see why it could affect the schedule. You accept a formal impact analysis only if the PM explains it clearly and offers options (for example: a later phase, trading off other features, or extra budget and time). You are under pressure from your own CEO.",
	},
	{
		ID:      "schedule-slip",
		Title:   "Project is three weeks behind",
		Setup:   "Your software implementation project is behind schedule: the Schedule Performance Index (SPI) is 0.8, about three weeks late. The client sponsor, Maria Lopez, has called an urgent status meeting.",
		SetupVI: "Dự án chậm khoảng 3 tuần (SPI = 0.8). Sponsor Maria Lopez gọi họp khẩn. Bạn phải báo cáo trung thực và đưa ra phương án.",
		Goal:    "Report honestly with facts (planned versus actual progress, the cause), take ownership, present recovery options with their costs and risks (for example adding resources, fast tracking, reducing scope), and ask for a decision.",
		Topics:  "Schedule control, EVM (SPI), schedule compression (crashing, fast tracking), communication",
		Client:  "You are Maria Lopez, the client's project sponsor. You heard from your own team that the project is late and nobody told you. You want to know why, who is responsible, what the new date is and what it will cost. You dislike excuses and vague numbers. You calm down if the PM takes ownership, gives data and presents options with trade-offs.",
	},
	{
		ID:      "budget-overrun",
		Title:   "CFO questions the budget overrun",
		Setup:   "The Cost Performance Index (CPI) is 0.85. At this rate the project will finish at about 590,000 USD against a 500,000 USD budget (BAC). The client's CFO, Tom Reilly, wants an explanation.",
		SetupVI: "CPI = 0.85, dự báo hoàn thành khoảng 590k USD so với ngân sách 500k USD. CFO Tom Reilly muốn bạn giải thích bằng ngôn ngữ dễ hiểu.",
		Goal:    "Explain CPI, the forecast (EAC) and the variance in plain English, give the root cause without blaming, describe corrective actions, and ask for a decision about reserves or scope.",
		Topics:  "EVM (CPI, EAC, VAC), cost control, reserves, reporting to executives",
		Client:  "You are Tom Reilly, the client's CFO. You are not a project person; you think in money. You are alarmed that the project may cost about 90,000 USD more than the 500,000 USD budget. You challenge the numbers and ask how the forecast was calculated, whether it will get worse, and who pays. You accept explanations that are simple, honest and come with a concrete plan.",
	},
	{
		ID:      "risk-occurred",
		Title:   "A risk happened and the client was not warned",
		Setup:   "Hardware from a key supplier will arrive four weeks late. This risk was in your risk register, but the client's COO, Priya Nair, says nobody told her about it.",
		SetupVI: "Nhà cung cấp giao phần cứng trễ 4 tuần. Rủi ro này đã có trong risk register nhưng COO Priya Nair nói chưa ai báo cho chị ấy.",
		Goal:    "Acknowledge the communication gap, explain that the risk was identified and what the planned response is, describe the impact and the recovery plan, and agree how you will keep her informed from now on.",
		Topics:  "Risk register, risk response (contingency plan, workaround), stakeholder communication",
		Client:  "You are Priya Nair, the client's COO. You are upset that you learned about the supplier delay from someone else. You want to know what happens to the go-live date, why you were not warned, and what guarantee you get. You calm down when the PM is open about the gap, shows a clear plan and promises regular updates.",
	},
	{
		ID:      "acceptance",
		Title:   "Client refuses to accept the deliverable",
		Setup:   "Your team delivered the payment module. In user acceptance testing, the client's product owner, Kevin Zhang, says it does not meet his expectations and refuses to sign it off.",
		SetupVI: "Team đã giao module thanh toán. Trong UAT, Kevin Zhang (product owner của client) nói chưa đạt kỳ vọng và từ chối ký nghiệm thu.",
		Goal:    "Ask what exactly is wrong, compare it with the documented requirements and acceptance criteria, separate real defects from new requests, agree a path (defect repair or change request) and a new acceptance date.",
		Topics:  "Validate Scope vs Control Quality, acceptance criteria, requirements traceability, defect repair vs change request",
		Client:  "You are Kevin Zhang, the client's product owner. You tested the payment module and feel it is not what you imagined: some screens look different from what you had in mind and one feature you expected (saving cards for later) is missing. That feature was never written in the requirements, but you believe it was obvious. You refuse to sign off until it is fixed. You respect a PM who uses the agreed acceptance criteria calmly and offers a clear way forward.",
	},
	{
		ID:      "agile-fixed-scope",
		Title:   "Fixed scope, date and price on an agile project",
		Setup:   "Your team works in Scrum with two-week sprints. A new client director, Elena Rossi, expects a fixed scope list, a fixed date, a fixed price and a detailed Gantt chart.",
		SetupVI: "Team dùng Scrum, sprint 2 tuần. Giám đốc mới của client, Elena Rossi, muốn scope, ngày và giá cố định kèm Gantt chi tiết. Bạn phải giải thích cách làm agile.",
		Goal:    "Explain agile delivery in business terms (timeboxes, a prioritised backlog, frequent demos, early value), show how she still gets predictability (roadmap, release plan, velocity forecast) and explain her role in prioritising the backlog.",
		Topics:  "Agile vs predictive, product backlog, velocity, release planning, managing expectations",
		Client:  "You are Elena Rossi, a client director used to traditional projects. You want a fixed list of features, a fixed delivery date, a fixed price and a detailed Gantt chart so you can report to your board. You are worried that 'agile' means no commitment. You accept an agile approach if the PM explains it in terms of risk, value and control and gives you a way to see progress and make decisions.",
	},
}

// FindRoleplayScenario looks a scenario up by id.
func FindRoleplayScenario(id string) (RoleplayScenario, bool) {
	for _, s := range RoleplayScenarios {
		if s.ID == id {
			return s, true
		}
	}
	return RoleplayScenario{}, false
}

func findPersona(id string) RoleplayPersona {
	for _, p := range RoleplayPersonas {
		if p.ID == id {
			return p
		}
	}
	return RoleplayPersonas[0]
}

func roleplayTranscript(history []RoleplayTurn) string {
	if len(history) > 24 {
		history = history[len(history)-24:]
	}
	var b strings.Builder
	for _, t := range history {
		who := "PM"
		if t.Role == "client" {
			who = "Client"
		}
		fmt.Fprintf(&b, "%s: %s\n", who, truncate(strings.TrimSpace(t.Text), 700))
	}
	return b.String()
}

// RoleplayReply writes the client's next spoken line. With an empty history it
// is the client's opening line.
func (c *Client) RoleplayReply(ctx context.Context, model string, sc RoleplayScenario, personaID string, history []RoleplayTurn) (string, error) {
	if !c.Enabled() {
		return "", ErrNotConfigured
	}
	p := findPersona(personaID)
	system := "You play a client in a business role-play to help a project manager practise spoken English. " +
		"Stay in character. Reply with the client's spoken line only."
	var b strings.Builder
	fmt.Fprintf(&b, "Situation: %s\n\nYour character (hidden from the PM): %s\nYour attitude: %s\n\n", sc.Setup, sc.Client, p.Behavior)
	if len(history) == 0 {
		b.WriteString("The meeting has just started. Say your opening line and state your concern right away.\n")
	} else {
		fmt.Fprintf(&b, "Conversation so far:\n%s\nNow write the Client's next line.\n", roleplayTranscript(history))
	}
	b.WriteString(`
Rules:
- Speak natural business English at about B1-B2 level: clear words, short sentences, no heavy slang or idioms. The learner has weak listening.
- 1 to 3 sentences, at most 45 words. Never write stage directions, quotes, names before the line, or Vietnamese.
- Do not solve the problem for the PM and do not accept vague answers: ask a follow-up or push back.
- React to what the PM actually said. If the PM acknowledges the issue, gives facts and offers clear options, soften gradually. After a strong answer, you may agree on a concrete next step and end the meeting politely.
- Never break character and never mention that this is a role-play or an exercise.`)

	out, err := c.chatText(ctx, model, system, b.String(), 0.8)
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "Client:")
	out = strings.Trim(strings.TrimSpace(out), "\"“”")
	if out == "" {
		return "", fmt.Errorf("empty reply")
	}
	return out, nil
}

// RoleplayCorrection is one thing the PM said, a better way to say it, and why.
type RoleplayCorrection struct {
	You    string `json:"you"`
	Better string `json:"better"`
	WhyVI  string `json:"why_vi"`
}

// RoleplayPhrase is a reusable phrase from the conversation.
type RoleplayPhrase struct {
	EN string `json:"en"`
	VI string `json:"vi"`
}

// RoleplayFeedback is the review shown at the end of a conversation.
type RoleplayFeedback struct {
	SummaryVI    string               `json:"summary_vi"`
	PMIScore     int                  `json:"pmi_score"` // 1-5
	PMIGoodVI    []string             `json:"pmi_good_vi"`
	PMIImproveVI []string             `json:"pmi_improve_vi"`
	EnglishScore int                  `json:"english_score"` // 1-5
	Corrections  []RoleplayCorrection `json:"corrections"`
	Phrases      []RoleplayPhrase     `json:"phrases"`
	ModelAnswer  string               `json:"model_answer"`
	NextStepVI   string               `json:"next_step_vi"`
}

// RoleplayReview reviews the PM's side of the conversation: how well the
// situation was handled by PMI standards, and the quality of the English.
func (c *Client) RoleplayReview(ctx context.Context, model string, sc RoleplayScenario, history []RoleplayTurn) (RoleplayFeedback, error) {
	var fb RoleplayFeedback
	if !c.Enabled() {
		return fb, ErrNotConfigured
	}
	system := "You are a PMP instructor and an English speaking coach. Reply with a single JSON object only."
	user := fmt.Sprintf(`A learner (Vietnamese, intermediate English, weak speaking) played the PM in this role-play.

Situation: %s
What a good PM should achieve: %s
PMP topics: %s

Conversation:
%s
Review ONLY the PM's lines. Be honest and specific, not flattering.

Notes:
- The PM's lines may come from speech recognition: ignore punctuation, capitalisation and words that are probably mis-transcriptions. Judge grammar, word choice, clarity, politeness and tone.
- If the PM said very little (fewer than 2 lines), say there is too little to judge and keep the lists short.

Return JSON:
{
 "summary_vi": "2-3 sentences in Vietnamese: overall impression",
 "pmi_score": 1-5 integer (how well the situation was handled by PMI standards),
 "pmi_good_vi": ["what the PM did right, Vietnamese, max 4 items"],
 "pmi_improve_vi": ["what to do differently by PMI practice (e.g. assess impact first, use change control, give options, ownership), Vietnamese, max 4 items"],
 "english_score": 1-5 integer (clarity, grammar, tone for a business meeting),
 "corrections": [ {"you":"exact PM sentence","better":"natural business English version","why_vi":"short Vietnamese explanation"} ] (max 6, the most useful ones; skip lines that were fine),
 "phrases": [ {"en":"reusable phrase for similar meetings","vi":"Vietnamese meaning"} ] (4-6, mostly taken from your 'better' versions),
 "model_answer": "one strong example reply (3-4 sentences, English) for the hardest moment of the conversation",
 "next_step_vi": "one concrete thing to practise next, Vietnamese"
}`, sc.Setup, sc.Goal, sc.Topics, roleplayTranscript(history))

	if err := c.chatJSONModel(ctx, model, system, user, 0.3, &fb); err != nil {
		return fb, err
	}
	fb.PMIScore = clampScore(fb.PMIScore)
	fb.EnglishScore = clampScore(fb.EnglishScore)
	if len(fb.Corrections) > 6 {
		fb.Corrections = fb.Corrections[:6]
	}
	if len(fb.Phrases) > 6 {
		fb.Phrases = fb.Phrases[:6]
	}
	return fb, nil
}

func clampScore(n int) int {
	if n < 1 {
		return 1
	}
	if n > 5 {
		return 5
	}
	return n
}

// ---------------------------------------------------------------- teach-back mode
//
// Mock-PM: the learner explains a PMP concept out loud (in Vietnamese) to an
// AI playing a junior colleague who is trying to understand it. This is the
// reverse of the client role-play above: here the learner is tested on
// whether they can teach the concept correctly and clearly (the "Feynman
// technique"), which is one of the most effective ways to cement it. Topics
// are fixed, tied to PMP cheat-sheet sections; a learner can also bring
// their own topic (the caller builds an ad-hoc TeachTopic for "custom").

// TeachTopic is a PMP concept the learner should be able to explain.
type TeachTopic struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Brief     string `json:"brief"` // what the junior is confused about / asks to have explained (Vietnamese)
	Checklist string `json:"-"`     // hidden: key points a correct explanation should cover, used only for grading
}

var TeachTopics = []TeachTopic{
	{
		ID:        "evm",
		Title:     "EVM & các công thức (CV, SV, CPI, SPI, EAC, ETC, TCPI)",
		Brief:     "Em thấy sếp hay nhắc CPI với SPI trong báo cáo dự án mà em không hiểu lắm, sếp giải thích cho em với?",
		Checklist: "3 biến gốc BAC/AC/EV; CV=EV-AC, SV=EV-PV (variance = trừ); CPI=EV/AC, SPI=EV/PV (index = chia); > 1 hoặc > 0 là tốt; EAC=AC+ETC và EAC=BAC/CPI; TCPI=(BAC-EV)/(BAC-AC); ý nghĩa thực tế (đang chi vượt/chậm tiến độ).",
	},
	{
		ID:        "change-control",
		Title:     "Integrated Change Control — quy trình xử lý yêu cầu thay đổi",
		Brief:     "Khách hàng vừa nhắn bảo thêm một tính năng mới, mình cứ làm luôn được không anh, sao phải qua quy trình gì đó?",
		Checklist: "Không tự ý làm luôn; ghi nhận change request; đánh giá tác động (scope/schedule/cost/risk/quality); trình Change Control Board (CCB) hoặc người có thẩm quyền phê duyệt; cập nhật baseline + tài liệu dự án nếu được duyệt; từ chối hoặc hoãn cũng là một outcome hợp lệ.",
	},
	{
		ID:        "agile-vs-predictive",
		Title:     "Agile vs Predictive (Waterfall) — khi nào chọn cái nào",
		Brief:     "Dự án mới sắp làm, sếp định chọn Agile hay theo kiểu cũ (Waterfall) vậy anh, dựa vào đâu để chọn?",
		Checklist: "Predictive phù hợp khi yêu cầu rõ ràng, ổn định, rủi ro thấp, cần kế hoạch chi tiết trước; Agile phù hợp khi yêu cầu còn mơ hồ/hay đổi, cần phản hồi sớm, giá trị giao theo từng đợt (incremental); hybrid khi một phần dự án ổn định một phần biến động; quyết định dựa trên mức độ chắc chắn của yêu cầu và công nghệ, không phải sở thích.",
	},
	{
		ID:        "risk-response",
		Title:     "4 chiến lược ứng phó rủi ro (Avoid / Transfer / Mitigate / Accept)",
		Brief:     "Em thấy risk register ghi 'mitigate', 'transfer' các kiểu, mấy cái đó khác nhau sao vậy anh?",
		Checklist: "Avoid = loại bỏ nguyên nhân/thay đổi kế hoạch để rủi ro không còn xảy ra; Transfer = chuyển hậu quả cho bên thứ ba (bảo hiểm, outsourcing, hợp đồng); Mitigate = giảm xác suất hoặc tác động chứ không loại bỏ hẳn; Accept = chủ động chấp nhận (passive) hoặc chuẩn bị contingency (active) khi chi phí xử lý cao hơn tác động; chọn chiến lược dựa trên mức độ ưu tiên (probability x impact) và chi phí xử lý.",
	},
	{
		ID:        "stakeholder-raci",
		Title:     "Stakeholder management & RACI — ai chịu trách nhiệm gì",
		Brief:     "Dự án nhiều người liên quan quá, ai cũng có ý kiến, làm sao biết ai quyết định cuối cùng vậy anh?",
		Checklist: "Xác định & phân loại stakeholder (quyền lực/mức quan tâm — power/interest grid); RACI: Responsible (người làm), Accountable (người chịu trách nhiệm cuối, chỉ 1 người/việc), Consulted (được hỏi ý kiến trước), Informed (được thông báo sau); mỗi task nên có đúng 1 Accountable để tránh đùn đẩy; cập nhật kế hoạch quản lý stakeholder khi dự án tiến triển.",
	},
}

// FindTeachTopic looks a teach-mode topic up by id.
func FindTeachTopic(id string) (TeachTopic, bool) {
	for _, t := range TeachTopics {
		if t.ID == id {
			return t, true
		}
	}
	return TeachTopic{}, false
}

// TeachPersonas: how the junior colleague listens and reacts.
var TeachPersonas = []RoleplayPersona{
	{ID: "curious", Label: "Tò mò, dễ tính", Desc: "Hỏi thêm để hiểu rõ hơn, phản hồi tích cực khi hiểu.",
		Behavior: "Bạn là một đồng nghiệp mới (junior), tò mò và dễ tính. Khi nghe giải thích rõ ràng, bạn phản hồi kiểu 'à em hiểu rồi' và có thể hỏi thêm 1 câu liên quan để đào sâu. Khi còn mơ hồ, bạn hỏi lại lịch sự."},
	{ID: "confused", Label: "Hay bối rối, cần ví dụ", Desc: "Dễ bị rối với thuật ngữ, liên tục xin ví dụ cụ thể hơn.",
		Behavior: "Bạn là một đồng nghiệp mới (junior) dễ bị rối với thuật ngữ chuyên môn. Nếu câu trả lời dùng từ chuyên ngành mà không giải thích, bạn nói 'anh/chị giải thích lại bằng từ đơn giản hơn được không' hoặc xin một ví dụ thực tế cụ thể. Bạn chỉ thật sự hiểu khi có ví dụ rõ ràng."},
	{ID: "challenging", Label: "Hay vặn hỏi 'tại sao'", Desc: "Nghi ngờ, muốn biết tại sao điều này quan trọng trong thực tế.",
		Behavior: "Bạn là một đồng nghiệp mới (junior) hay thắc mắc và hơi hoài nghi. Bạn hỏi 'tại sao phải làm vậy', 'không làm thì sao', và muốn một ví dụ thực tế chứng minh điều đó quan trọng. Bạn chỉ bị thuyết phục khi nghe lý do + hậu quả cụ thể nếu bỏ qua."},
}

func findTeachPersona(id string) RoleplayPersona {
	for _, p := range TeachPersonas {
		if p.ID == id {
			return p
		}
	}
	return TeachPersonas[0]
}

// TeachTurn writes the junior colleague's next line (the opening question
// when history is empty). The learner's lines are their attempt to explain
// topic.Title out loud; reply in Vietnamese, spoken style.
func (c *Client) TeachTurn(ctx context.Context, model string, topic TeachTopic, personaID string, history []RoleplayTurn) (string, error) {
	if !c.Enabled() {
		return "", ErrNotConfigured
	}
	p := findTeachPersona(personaID)
	system := "You play a junior project team member asking a senior PM to explain a PMP concept, to help the PM practise explaining it clearly. " +
		"Stay in character. Reply in Vietnamese, spoken style, with the junior's line only."
	var b strings.Builder
	fmt.Fprintf(&b, "Chủ đề cần được giải thích: %s\nĐiều bạn (junior) đang thắc mắc: %s\nTính cách của bạn: %s\n\n", topic.Title, topic.Brief, p.Behavior)
	if len(history) == 0 {
		b.WriteString("Hãy mở đầu bằng đúng câu thắc mắc trên (có thể diễn đạt lại tự nhiên hơn một chút).\n")
	} else {
		fmt.Fprintf(&b, "Cuộc trò chuyện cho đến giờ:\n%s\nBây giờ viết câu tiếp theo của bạn (junior).\n", roleplayTranscript(history))
	}
	b.WriteString(`
Quy tắc:
- Nói tiếng Việt tự nhiên, đời thường, 1-2 câu ngắn.
- Không tự giải thích hộ PM; chỉ phản ứng với điều PM vừa nói (hỏi thêm, xin ví dụ, hoặc xác nhận đã hiểu).
- Nếu PM giải thích rõ và đúng, bạn có thể nói hiểu rồi và dừng lại (kết thúc hội thoại một cách tự nhiên) sau khi đã hỏi đủ sâu theo tính cách của bạn.
- Không bao giờ thoát vai và không nhắc đây là bài luyện tập.`)

	out, err := c.chatText(ctx, model, system, b.String(), 0.8)
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	out = strings.Trim(strings.TrimSpace(out), "\"“”")
	if out == "" {
		return "", fmt.Errorf("empty reply")
	}
	return out, nil
}

// TeachReview grades how well the learner (playing the PM) explained the
// topic to the junior: PMI correctness/completeness and clarity. Reuses the
// RoleplayFeedback shape; EnglishScore here means "clarity" (the frontend
// relabels it for teach mode), since this mode is about concept mastery, not
// English.
func (c *Client) TeachReview(ctx context.Context, model string, topic TeachTopic, history []RoleplayTurn) (RoleplayFeedback, error) {
	var fb RoleplayFeedback
	if !c.Enabled() {
		return fb, ErrNotConfigured
	}
	system := "You are a PMP instructor grading how well someone explained a concept to a junior colleague. Reply with a single JSON object only, all text in Vietnamese."
	user := fmt.Sprintf(`Một học viên (đang ôn thi PMP) đóng vai PM, giải thích khái niệm sau cho một đồng nghiệp mới:

Chủ đề: %s
Những điểm một câu trả lời đúng/đủ nên có (ẩn với học viên, chỉ dùng để chấm): %s

Cuộc hội thoại:
%s
Chỉ chấm các lượt nói của PM (học viên). Thẳng thắn, cụ thể, không tâng bốc.

Ghi chú:
- Lời của học viên có thể đến từ nhận dạng giọng nói: bỏ qua lỗi chính tả/dấu câu, chỉ chấm nội dung, độ chính xác và độ rõ ràng.
- Nếu học viên nói quá ít (dưới 2 lượt), ghi nhận là chưa đủ để đánh giá và giữ các danh sách ngắn gọn.

Trả về JSON:
{
 "summary_vi": "2-3 câu tổng quan",
 "pmi_score": 1-5 (độ chính xác & đầy đủ theo PMBOK so với các điểm cần có ở trên),
 "pmi_good_vi": ["những gì đã giải thích đúng/tốt, tối đa 4 ý"],
 "pmi_improve_vi": ["những điểm còn thiếu hoặc chưa chính xác theo PMBOK, tối đa 4 ý"],
 "english_score": 1-5 (độ rõ ràng, dễ hiểu của cách diễn đạt — không liên quan tiếng Anh),
 "corrections": [],
 "phrases": [ {"en":"một câu diễn đạt mẫu rõ ràng cho điểm khó nhất", "vi":"giải thích ngắn tại sao câu này rõ"} ] (2-4 mục),
 "model_answer": "một ví dụ giải thích mẫu đầy đủ, rõ ràng, tiếng Việt (4-6 câu) cho chủ đề này",
 "next_step_vi": "một việc cụ thể nên luyện tiếp"
}`, topic.Title, topic.Checklist, roleplayTranscript(history))

	if err := c.chatJSONModel(ctx, model, system, user, 0.3, &fb); err != nil {
		return fb, err
	}
	fb.PMIScore = clampScore(fb.PMIScore)
	fb.EnglishScore = clampScore(fb.EnglishScore)
	if len(fb.Corrections) > 6 {
		fb.Corrections = fb.Corrections[:6]
	}
	if len(fb.Phrases) > 6 {
		fb.Phrases = fb.Phrases[:6]
	}
	return fb, nil
}
