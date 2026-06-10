package handlers

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"mvp-lang-app/internal/config"
	"mvp-lang-app/internal/http/dto"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB  *gorm.DB
	Cfg config.Config
}

// ---- Health ----

func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "time": time.Now()})
}

// ---- Auth ----

func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	var existing models.User
	if h.DB.Where("email = ?", req.Email).First(&existing).Error == nil {
		c.JSON(409, gin.H{"error": "email already registered"})
		return
	}
	user := models.User{Email: req.Email, DisplayName: req.DisplayName}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	// Create default preferences
	h.DB.Create(&models.UserPreference{UserID: user.ID})

	c.JSON(201, gin.H{"id": user.ID, "email": user.Email})
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	var user models.User
	if h.DB.Where("email = ?", req.Email).First(&user).Error != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(200, gin.H{"user_id": user.ID, "email": user.Email})
}

// ---- Vocabulary ----

func (h *Handler) ListVocabs(c *gin.Context) {
	topic := c.Query("topic")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	q := h.DB.Preload("Sentences")
	if topic != "" {
		q = q.Where("topic_id IN (SELECT id FROM topic_categories WHERE slug = ?)", topic)
	}
	if search != "" {
		q = q.Where("word ILIKE ?", "%"+search+"%")
	}

	var vocabs []models.Vocabulary
	var total int64
	q.Model(&models.Vocabulary{}).Count(&total)
	if err := q.Limit(limit).Offset(offset).Find(&vocabs).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	c.JSON(200, gin.H{"data": vocabs, "total": total, "page": page, "limit": limit})
}

func (h *Handler) GetVocab(c *gin.Context) {
	id := c.Param("id")
	var vocab models.Vocabulary
	if err := h.DB.Preload("Sentences").First(&vocab, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, vocab)
}

func (h *Handler) GetTopics(c *gin.Context) {
	var topics []models.TopicCategory
	h.DB.Order("sort_order").Find(&topics)
	c.JSON(200, topics)
}

// ---- SRS ----

func (h *Handler) StartLearning(c *gin.Context) {
	var req dto.StartLearningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	userID := getUserID(c)

	for _, vid := range req.VocabIDs {
		var existing models.UserVocabProgress
		if h.DB.Where("user_id = ? AND vocab_id = ?", userID, vid).First(&existing).Error == nil {
			continue
		}
		progress := models.UserVocabProgress{
			UserID:       userID,
			VocabID:      vid,
			State:        0,
			Stability:    2.5,
			Difficulty:   0.3,
			NextReviewAt: time.Now().Add(24 * time.Hour),
			LevelFocus:   1,
		}
		h.DB.Create(&progress)
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) GetDueReviews(c *gin.Context) {
	userID := getUserID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var items []dto.DueReviewItem
	h.DB.Raw(`
		SELECT uvp.vocab_id, v.word, v.phonetics, uvp.level_focus AS level,
			   s.example_en, s.example_vi, s.meaning_en, s.meaning_vi, s.usage_note
		FROM user_vocab_progress uvp
		JOIN vocabularies v ON v.id = uvp.vocab_id
		JOIN sentences s ON s.vocab_id = v.id AND s.level = uvp.level_focus
		WHERE uvp.user_id = ? AND uvp.is_active = true
		ORDER BY uvp.next_review_at ASC
		LIMIT ?
	`, userID, limit).Scan(&items)
	c.JSON(200, items)
}

func (h *Handler) SubmitReview(c *gin.Context) {
	var req dto.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	userID := getUserID(c)

	var progress models.UserVocabProgress
	if err := h.DB.Where("user_id = ? AND vocab_id = ?", userID, req.VocabID).First(&progress).Error; err != nil {
		c.JSON(404, gin.H{"error": "vocab not in learning list"})
		return
	}

	elapsed := int(time.Since(progress.NextReviewAt).Hours() / 24)
	if elapsed < 0 {
		elapsed = 0
	}

	result := srs.Schedule(srs.FSRSParams{
		Stability:   progress.Stability,
		Difficulty:  progress.Difficulty,
		ElapsedDays: elapsed,
		Reps:        progress.Reps,
		Lapses:      progress.Lapses,
		State:       progress.State,
	}, req.Rating)

	if req.Rating < 3 {
		progress.Lapses++
	}
	progress.Reps++
	progress.State = result.NextState
	progress.Stability = result.NextStability
	progress.Difficulty = result.NextDifficulty
	progress.ScheduledDays = result.NextInterval
	now := time.Now()
	progress.LastReviewAt = &now
	progress.NextReviewAt = result.NextReviewAt
	progress.MasteryScore = (1.0 - result.NextDifficulty) * 100

	h.DB.Save(&progress)

	// Update streak
	today := time.Now().Format("2006-01-02")
	h.DB.Where("user_id = ? AND streak_date = ?", userID, today).
		FirstOrCreate(&models.UserStreak{UserID: userID, StreakDate: today})
	h.DB.Model(&models.UserStreak{}).
		Where("user_id = ? AND streak_date = ?", userID, today).
		Update("vocab_reviewed", gorm.Expr("vocab_reviewed + 1"))

	c.JSON(200, dto.ReviewResponse{
		NextInterval: result.NextInterval,
		NextReviewAt: result.NextReviewAt.Format(time.RFC3339),
		Stability:    result.NextStability,
		Difficulty:   result.NextDifficulty,
		State:        result.NextState,
	})
}

// ---- Quiz ----

func (h *Handler) GetQuiz(c *gin.Context) {
	var questions []models.QuizQuestion
	h.DB.Limit(5).Order("random()").Find(&questions)

	var result []dto.QuizQuestionDTO
	for _, q := range questions {
		var choices []string
		json.Unmarshal([]byte(q.Choices), &choices)
		result = append(result, dto.QuizQuestionDTO{
			ID:      q.ID,
			Prompt:  q.Prompt,
			Choices: choices,
		})
	}
	c.JSON(200, result)
}

func (h *Handler) SubmitAnswer(c *gin.Context) {
	var req dto.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	userID := getUserID(c)

	var question models.QuizQuestion
	if err := h.DB.First(&question, req.QuestionID).Error; err != nil {
		c.JSON(404, gin.H{"error": "question not found"})
		return
	}

	var choices []map[string]interface{}
	json.Unmarshal([]byte(question.Choices), &choices)
	correctIdx := -1
	for i, ch := range choices {
		if isCorrect, _ := ch["is_correct"].(bool); isCorrect {
			correctIdx = i
			break
		}
	}
	correct := req.AnswerIdx == correctIdx

	// Log attempt
	var attempt models.QuizAttempt
	h.DB.Where("user_id = ? AND status = 'started'", userID).
		Order("started_at DESC").First(&attempt)
	h.DB.Create(&models.QuizAttempt{
		UserID:         userID,
		QuizID:         question.QuizID,
		TotalQuestions: 1,
	})

	c.JSON(200, dto.AnswerResponse{
		Correct:          correct,
		CorrectAnswerIdx: correctIdx,
		Explanation:      question.Explanation,
	})
}

// ---- Writing ----

func (h *Handler) SubmitWriting(c *gin.Context) {
	var req dto.WritingSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	userID := getUserID(c)

	reqVocabsJSON, _ := json.Marshal(req.RequiredVocabs)
	submission := models.WritingSubmission{
		UserID:         userID,
		Prompt:         req.Prompt,
		OriginalText:   req.Text,
		WordCount:      len(req.Text),
		RequiredVocabs: string(reqVocabsJSON),
		Status:         "pending",
	}
	if err := h.DB.Create(&submission).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	// AI evaluation runs async in background
	go h.evaluateWriting(submission.ID)

	c.JSON(201, dto.WritingFeedbackResponse{
		ID:     submission.ID,
		Status: "pending",
	})
}

func (h *Handler) GetWritingFeedback(c *gin.Context) {
	id := c.Param("id")
	var sub models.WritingSubmission
	if err := h.DB.First(&sub, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, gin.H{
		"id":              sub.ID,
		"prompt":          sub.Prompt,
		"original_text":   sub.OriginalText,
		"overall_score":   sub.OverallScore,
		"grammar_score":   sub.GrammarScore,
		"vocab_score":     sub.VocabScore,
		"coherence_score": sub.CoherenceScore,
		"ai_feedback":     sub.AIFeedbackJSON,
		"status":          sub.Status,
	})
}

func (h *Handler) GetWritingHistory(c *gin.Context) {
	userID := getUserID(c)
	var subs []models.WritingSubmission
	h.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(20).Find(&subs)
	c.JSON(200, subs)
}

func (h *Handler) evaluateWriting(submissionID int64) {
	// Mark as processing
	h.DB.Model(&models.WritingSubmission{}).Where("id = ?", submissionID).Update("status", "processing")

	// Simple scoring (placeholder - real version calls AI)
	score := 75.0
	grammar := 72.0
	vocab := 78.0
	coherence := 75.0

	now := time.Now()
	h.DB.Model(&models.WritingSubmission{}).Where("id = ?", submissionID).Updates(map[string]interface{}{
		"overall_score":   score,
		"grammar_score":   grammar,
		"vocab_score":     vocab,
		"coherence_score": coherence,
		"ai_model_used":   "placeholder",
		"status":          "completed",
		"completed_at":    now,
	})
}

// ---- Dashboard ----

func (h *Handler) GetDashboard(c *gin.Context) {
	userID := getUserID(c)

	var resp dto.DashboardResponse
	h.DB.Raw(`
		SELECT
			COALESCE(COUNT(*) FILTER (WHERE is_active = true), 0) AS active_vocabs,
			COALESCE(COUNT(*) FILTER (WHERE mastery_score >= 80), 0) AS mastered_vocabs,
			COALESCE(COUNT(*) FILTER (WHERE next_review_at <= NOW()), 0) AS due_reviews
		FROM user_vocab_progress
		WHERE user_id = ?
	`, userID).Scan(&resp)

	h.DB.Raw(`
		SELECT COALESCE(COUNT(*), 0) FROM user_streaks
		WHERE user_id = ? AND streak_date >= CURRENT_DATE - INTERVAL '7 days'
	`, userID).Scan(&resp.Streak7d)

	h.DB.Raw(`
		SELECT MAX(overall_score) FROM writing_submissions
		WHERE user_id = ? AND created_at >= NOW() - INTERVAL '7 days'
	`, userID).Scan(&resp.BestWriting)

	h.DB.Raw(`
		SELECT MAX(score) FROM quiz_attempts
		WHERE user_id = ? AND started_at >= NOW() - INTERVAL '7 days'
	`, userID).Scan(&resp.BestQuiz)

	c.JSON(200, resp)
}

// ---- Search ----

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(400, gin.H{"error": "query required"})
		return
	}
	var vocabs []models.Vocabulary
	h.DB.Preload("Sentences").
		Where("word ILIKE ?", "%"+query+"%").
		Limit(20).Find(&vocabs)
	c.JSON(200, gin.H{"results": vocabs, "query": query})
}

// ---- Preferences ----

func (h *Handler) GetPreferences(c *gin.Context) {
	userID := getUserID(c)
	var pref models.UserPreference
	if h.DB.Where("user_id = ?", userID).First(&pref).Error != nil {
		pref = models.UserPreference{UserID: userID}
		h.DB.Create(&pref)
	}
	c.JSON(200, pref)
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID := getUserID(c)
	var req dto.PreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	updates := map[string]interface{}{}
	if req.DailyVocabGoal != nil {
		updates["daily_vocab_goal"] = *req.DailyVocabGoal
	}
	if req.ReminderEnabled != nil {
		updates["reminder_enabled"] = *req.ReminderEnabled
	}
	if req.ReminderTimes != nil {
		updates["reminder_times"] = *req.ReminderTimes
	}
	if req.QuizDifficulty != nil {
		updates["quiz_difficulty"] = *req.QuizDifficulty
	}
	if req.TTSSpeed != nil {
		updates["tts_speed"] = *req.TTSSpeed
	}

	h.DB.Model(&models.UserPreference{}).Where("user_id = ?", userID).Updates(updates)
	c.JSON(200, gin.H{"ok": true})
}

// ---- Speech / Audio ----

func (h *Handler) UploadAudio(c *gin.Context) {
	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(400, gin.H{"error": "no audio file"})
		return
	}
	// In production: upload to S3/MinIO
	// For MVP: save to local disk
	path := "./uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(500, gin.H{"error": "save failed"})
		return
	}
	c.JSON(200, gin.H{"url": "/uploads/" + file.Filename, "ok": true})
}

// ---- AI Generate Random Vocabulary ----

func (h *Handler) GenerateRandomVocabs(c *gin.Context) {
	var req dto.GenerateRandomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Count = 10
	}
	if req.Count <= 0 {
		req.Count = 10
	}
	if req.Count > 20 {
		req.Count = 20
	}

	// Common words bank for MVP (when no AI API key configured)
	wordBank := []struct {
		Word      string
		Phonetics string
		Pos       string
		Topic     string
		Levels    []SentenceData
	}{
		{"leverage", "/ˈlevərɪdʒ/", "verb", "business", []SentenceData{
			{1, "to use something to get an advantage", "tận dụng để có lợi thế",
				"We can use our experience to get a better deal.", "Chúng ta có thể dùng kinh nghiệm để có deal tốt hơn.",
				"Level 1: Direct S+V+O, no phrasal verbs, Filipino business style"},
			{2, "to use something you have to achieve something new", "tận dụng thứ bạn sẵn có để đạt điều mới",
				"We should leverage our network to find new clients.", "Chúng ta nên tận dụng network để tìm khách hàng mới.",
				"Level 2: Natural business casual, uses 'network' collocation"},
			{3, "to strategically exploit an existing asset to maximize outcome", "khai thác chiến lược tài sản sẵn có để tối đa hóa kết quả",
				"By leveraging the organization's intellectual capital, the firm gained a sustainable competitive advantage.",
				"Bằng cách khai thác vốn trí tuệ, công ty đã đạt lợi thế cạnh tranh bền vững.",
				"Level 3: Complex gerund clause, formal vocabulary, IELTS 7.5+"},
		}},
		{"sustainable", "/səˈsteɪnəbl/", "adjective", "environment", []SentenceData{
			{1, "able to continue for a long time", "có thể duy trì lâu dài",
				"We need a plan that can last a long time.", "Chúng ta cần một kế hoạch có thể kéo dài lâu.",
				"Level 1: Simple definition, basic vocabulary"},
			{2, "able to be maintained at a certain rate or level", "có thể duy trì ở mức nhất định",
				"The company is working on more sustainable business practices.", "Công ty đang xây dựng các phương thức kinh doanh bền vững hơn.",
				"Level 2: Natural business usage"},
			{3, "meeting present needs without compromising future generations' ability to meet theirs", "đáp ứng nhu cầu hiện tại mà không ảnh hưởng thế hệ tương lai",
				"Sustainable development requires a paradigm shift in how we conceptualize economic growth.",
				"Phát triển bền vững đòi hỏi thay đổi mô hình trong cách ta khái niệm hóa tăng trưởng kinh tế.",
				"Level 3: Academic definition, complex structure"},
		}},
		{"implement", "/ˈɪmplɪment/", "verb", "business", []SentenceData{
			{1, "to put a plan into action", "đưa kế hoạch vào hành động",
				"We will start using the new system next week.", "Chúng tôi sẽ bắt đầu dùng hệ thống mới tuần tới.",
				"Level 1: Avoids 'implement', uses simpler verbs"},
			{2, "to put a decision or plan into effect", "thực thi một quyết định hoặc kế hoạch",
				"The team plans to implement the new software by Friday.", "Nhóm dự định triển khai phần mềm mới trước thứ Sáu.",
				"Level 2: Natural workplace English"},
			{3, "to execute or put into effect according to a definite plan or procedure", "thực thi theo một kế hoạch hoặc quy trình xác định",
				"The implementation of enterprise-wide digital transformation initiatives necessitates cross-functional collaboration.",
				"Việc triển khai các sáng kiến chuyển đổi số toàn doanh nghiệp đòi hỏi sự hợp tác liên chức năng.",
				"Level 3: Nominalization, complex sentence, formal register"},
		}},
		{"collaborate", "/kəˈlæbəreɪt/", "verb", "workplace", []SentenceData{
			{1, "to work together with others", "làm việc cùng với người khác",
				"We work together on the project.", "Chúng tôi làm việc cùng nhau trong dự án.",
				"Level 1: Simple phrasal alternative"},
			{2, "to work jointly with others on an activity", "cùng hợp tác với người khác trong một hoạt động",
				"I need to collaborate with the design team on this.", "Tôi cần hợp tác với nhóm thiết kế về việc này.",
				"Level 2: Natural workplace conversation"},
			{3, "to cooperate with others in a joint intellectual endeavor", "cộng tác với người khác trong nỗ lực trí tuệ chung",
				"Cross-departmental collaboration is imperative for fostering organizational innovation.",
				"Sự hợp tác liên phòng ban là cần thiết để thúc đẩy đổi mới tổ chức.",
				"Level 3: Academic, compound adjectives, formal vocabulary"},
		}},
		{"efficient", "/ɪˈfɪʃnt/", "adjective", "business", []SentenceData{
			{1, "doing things well without wasting time or energy", "làm việc tốt không lãng phí thời gian",
				"She works fast and does not waste time.", "Cô ấy làm nhanh và không lãng phí thời gian.",
				"Level 1: Descriptive, no abstract terms"},
			{2, "achieving maximum productivity with minimum wasted effort", "đạt năng suất tối đa với nỗ lực tối thiểu",
				"We need a more efficient way to handle customer requests.", "Chúng ta cần cách hiệu quả hơn để xử lý yêu cầu khách hàng.",
				"Level 2: Natural business context"},
			{3, "functioning in the best possible manner with the least waste of time and effort", "hoạt động tối ưu với ít lãng phí nhất",
				"Operational efficiency constitutes a cornerstone of competitive advantage in contemporary markets.",
				"Hiệu quả vận hành là nền tảng của lợi thế cạnh tranh trong thị trường hiện đại.",
				"Level 3: Academic essay style, advanced connectors"},
		}},
		{"negotiate", "/nɪˈɡəʊʃieɪt/", "verb", "business", []SentenceData{
			{1, "to talk to reach an agreement", "nói chuyện để đạt thỏa thuận",
				"We talked with the seller to get a lower price.", "Chúng tôi nói chuyện với người bán để có giá thấp hơn.",
				"Level 1: Simple vocabulary replacement"},
			{2, "to discuss something formally in order to reach an agreement", "thảo luận chính thức để đạt thỏa thuận",
				"They're negotiating a better salary package with the company.", "Họ đang thương lượng gói lương tốt hơn với công ty.",
				"Level 2: Natural workplace English"},
			{3, "to confer with others in order to reach a mutually acceptable settlement", "hội đàm với người khác để đạt giải pháp chấp nhận được",
				"Effective negotiation skills are indispensable for conflict resolution in complex organizational dynamics.",
				"Kỹ năng đàm phán hiệu quả không thể thiếu cho giải quyết xung đột trong động lực tổ chức phức tạp.",
				"Level 3: Academic vocabulary, complex noun phrases"},
		}},
		{"resolve", "/rɪˈzɒlv/", "verb", "workplace", []SentenceData{
			{1, "to find an answer to a problem", "tìm câu trả lời cho vấn đề",
				"We fixed the problem quickly.", "Chúng tôi sửa vấn đề nhanh chóng.",
				"Level 1: Simple everyday verb"},
			{2, "to settle or find a solution to a problem or dispute", "giải quyết một vấn đề hoặc tranh chấp",
				"We need to resolve this issue before the deadline.", "Chúng ta cần giải quyết vấn đề này trước deadline.",
				"Level 2: Natural workplace usage"},
			{3, "to determine or settle a complex matter through systematic analysis", "xác định hoặc giải quyết vấn đề phức tạp qua phân tích hệ thống",
				"The organization must resolve the inherent tension between short-term profitability and long-term sustainability.",
				"Tổ chức phải giải quyết mâu thuẫn cố hữu giữa lợi nhuận ngắn hạn và bền vững dài hạn.",
				"Level 3: Complex abstract concepts, parallel structure"},
		}},
		{"facilitate", "/fəˈsɪlɪteɪt/", "verb", "business", []SentenceData{
			{1, "to make something easier", "làm điều gì đó dễ dàng hơn",
				"The new tool makes our work easier.", "Công cụ mới giúp công việc dễ dàng hơn.",
				"Level 1: Simple meaning, common words"},
			{2, "to make an action or process easy or easier", "làm cho hành động hoặc quy trình dễ dàng hơn",
				"A good manager facilitates communication between teams.", "Người quản lý tốt tạo điều kiện giao tiếp giữa các nhóm.",
				"Level 2: Natural professional context"},
			{3, "to provide the necessary conditions for a process or activity to occur smoothly", "cung cấp điều kiện cần thiết để quá trình diễn ra suôn sẻ",
				"Effective leadership facilitates organizational transformation by cultivating a culture of psychological safety.",
				"Lãnh đạo hiệu quả tạo điều kiện chuyển đổi tổ chức bằng cách nuôi dưỡng văn hóa an toàn tâm lý.",
				"Level 3: Complex with gerund phrases, academic terminology"},
		}},
		{"mitigate", "/ˈmɪtɪɡeɪt/", "verb", "academic", []SentenceData{
			{1, "to make something less harmful or bad", "làm điều gì bớt có hại",
				"We need to reduce the risk of problems.", "Chúng ta cần giảm rủi ro vấn đề.",
				"Level 1: Simple synonym replacement"},
			{2, "to make something less severe, serious, or painful", "làm giảm mức độ nghiêm trọng",
				"We need to mitigate the impact of these budget cuts.", "Chúng ta cần giảm thiểu tác động của việc cắt giảm ngân sách.",
				"Level 2: Professional context, natural usage"},
			{3, "to reduce the severity, seriousness, or painfulness of an adverse condition", "giảm mức độ nghiêm trọng của điều kiện bất lợi",
				"Diversification of investment portfolios serves to mitigate systemic financial risk.",
				"Đa dạng hóa danh mục đầu tư giúp giảm thiểu rủi ro tài chính hệ thống.",
				"Level 3: Academic/financial, complex nominal group"},
		}},
		{"articulate", "/ɑːˈtɪkjuleɪt/", "verb", "communication", []SentenceData{
			{1, "to express ideas clearly in words", "diễn đạt ý tưởng rõ ràng bằng lời",
				"She can explain her ideas very clearly.", "Cô ấy có thể giải thích ý tưởng rất rõ ràng.",
				"Level 1: Common word replacement"},
			{2, "to express thoughts and feelings clearly and effectively", "bày tỏ suy nghĩ và cảm xúc rõ ràng, hiệu quả",
				"He articulated his vision for the company really well.", "Anh ấy trình bày tầm nhìn cho công ty rất tốt.",
				"Level 2: Natural business English"},
			{3, "to express an idea or feeling fluently and coherently in a structured manner", "diễn đạt ý tưởng trôi chảy và mạch lạc theo cách có cấu trúc",
				"The ability to articulate complex strategic objectives distinguishes exceptional leaders from merely competent ones.",
				"Khả năng diễn đạt mục tiêu chiến lược phức tạp phân biệt nhà lãnh đạo xuất sắc với người chỉ đủ năng lực.",
				"Level 3: Complex comparison structure, advanced vocabulary"},
		}},
	}

	// Shuffle and pick req.Count items
	count := req.Count
	if count > len(wordBank) {
		count = len(wordBank)
	}

	var generated []models.Vocabulary
	for i := 0; i < count; i++ {
		entry := wordBank[i]

		// Check if word already exists
		var existing models.Vocabulary
		if h.DB.Where("word = ?", entry.Word).First(&existing).Error == nil {
			// Load existing with sentences
			h.DB.Preload("Sentences").First(&existing, existing.ID)
			generated = append(generated, existing)
			continue
		}

		// Create vocab
		vocab := models.Vocabulary{
			Word:      entry.Word,
			Phonetics: entry.Phonetics,
			Pos:       entry.Pos,
			Source:    "ai-generated",
		}
		h.DB.Create(&vocab)

		// Create 3 level sentences
		for _, s := range entry.Levels {
			sentence := models.Sentence{
				VocabID:   vocab.ID,
				Level:     s.Level,
				MeaningEn: s.MeaningEn,
				MeaningVi: s.MeaningVi,
				ExampleEn: s.ExampleEn,
				ExampleVi: s.ExampleVi,
				UsageNote: s.UsageNote,
			}
			h.DB.Create(&sentence)
		}

		// Reload with sentences
		h.DB.Preload("Sentences").First(&vocab, vocab.ID)
		generated = append(generated, vocab)
	}

	c.JSON(200, gin.H{"generated": len(generated), "vocabs": generated})
}

type SentenceData struct {
	Level     int
	MeaningEn string
	MeaningVi string
	ExampleEn string
	ExampleVi string
	UsageNote string
}

// ---- Helpers ----

func getUserID(c *gin.Context) int64 {
	// For MVP, use header or default to 1
	if uid := c.GetHeader("X-User-ID"); uid != "" {
		id, _ := strconv.ParseInt(uid, 10, 64)
		return id
	}
	return 1
}

func init() {
	log.Println("[Handlers] Initialized")
}
