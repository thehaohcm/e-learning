// Package http wires routes and middleware.
package http

import (
	"strings"
	"time"

	"mvp-lang-app/internal/config"
	"mvp-lang-app/internal/http/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	// Requests arrive via Cloudflare -> edge nginx -> web nginx (all on the
	// docker network), which pass the visitor's IP in X-Real-IP.
	_ = r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.1/32"})
	r.RemoteIPHeaders = []string{"X-Real-IP", "X-Forwarded-For"}
	r.MaxMultipartMemory = 1 << 20

	r.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		// the browser extension calls /api/ext from its own origin (chrome-extension://<id>);
		// it authenticates with its own key, there are no cookies to protect
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "chrome-extension://") || strings.HasPrefix(origin, "moz-extension://")
		},
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	h := handlers.New(db, cfg)

	r.GET("/health", h.Health)
	api := r.Group("/api")

	// ---- public
	authLimit := handlers.RateLimit(20)
	api.POST("/auth/register", authLimit, h.Register)
	api.POST("/auth/login", authLimit, h.Login)

	// ---- study device (ESP32). Token: "Authorization: Bearer dev_..."
	api.POST("/device/register", handlers.RateLimit(10), h.DeviceRegister)
	// words read aloud for the device (key = hash of the text, see internal/tts)
	api.GET("/audio/:file", handlers.RateLimit(300), h.Audio)
	dev := api.Group("/device", h.RequireDevice)
	{
		dev.GET("/status", h.DeviceStatus)
		dev.GET("/sync", h.DeviceSync)
		dev.POST("/reviews", h.DeviceReviews)
		dev.GET("/firmware/:file", h.DeviceFirmware)
		dev.POST("/extra-new", h.DeviceExtraNew)
		dev.POST("/practice", handlers.RateLimit(10), h.DevicePractice)          // may call AI
		dev.POST("/prepare-forms", handlers.RateLimit(10), h.DevicePrepareForms) // AI, in the background
	}

	// ---- browser extension. Key: "Authorization: Bearer ext_..." (Settings → Browser extension)
	ext := api.Group("/ext", handlers.RateLimit(60), h.RequireExtKey)
	{
		ext.GET("/me", h.ExtMe)
		ext.GET("/notes", h.ExtRecentNotes)
		ext.POST("/notes", h.ExtAddNote)
	}

	// ---- logged-in user (web)
	u := api.Group("", h.RequireUser)
	{
		practiceLimit := handlers.RateLimit(40) // endpoints that call AI (Write it checks each sentence)
		u.GET("/auth/me", h.Me)

		u.GET("/subjects", h.ListSubjects)
		u.GET("/cards", h.ListCards)
		u.GET("/cards/:id", h.GetCard)
		u.POST("/cards/pronunciations", h.Pronunciations)                          // stored only, no AI
		u.POST("/cards/:id/pronunciation", practiceLimit, h.GeneratePronunciation) // AI, on request
		u.GET("/cards/:id/examples", h.CardExamples)
		u.POST("/cards/:id/examples", practiceLimit, h.GenerateCardExamples)

		u.POST("/study/enroll", h.Enroll)
		u.GET("/study/queue", h.Queue)
		u.POST("/study/review", h.Review)
		u.POST("/study/suspend", h.Suspend)
		u.POST("/study/extra-new", h.ExtraNew)

		u.POST("/practice/word-questions", practiceLimit, h.WordQuestions) // may call AI (shared with the device)
		u.POST("/practice/word-questions/answer", h.WordQuestionAnswer)
		u.GET("/quiz", h.Quiz)
		u.POST("/quiz/answer", h.QuizAnswer)

		u.GET("/stats", h.Stats)
		u.GET("/stats/progress", h.Progress)
		u.POST("/quiz/stress", h.QuizStress)
		u.POST("/quiz/sound", h.QuizSound)
		u.POST("/pron/forgot", h.PronForgot)
		u.GET("/quiz/pron-status", h.PronStatus)
		u.POST("/quiz/pron-generate", practiceLimit, h.PronGenerate)
		u.GET("/extension/keys", h.ListExtKeys)
		u.POST("/extension/keys", handlers.RateLimit(10), h.CreateExtKey)
		u.DELETE("/extension/keys/:id", h.DeleteExtKey)
		u.GET("/push/key", h.PushKey)
		u.POST("/push/subscribe", h.PushSubscribe)
		u.POST("/push/unsubscribe", h.PushUnsubscribe)
		u.POST("/push/test", handlers.RateLimit(10), h.PushTest)
		u.GET("/settings", h.GetSettings)
		u.PUT("/settings", h.UpdateSettings)

		u.GET("/devices", h.ListDevices)
		u.POST("/devices/pair", h.PairDevice)
		u.PATCH("/devices/:id", h.RenameDevice)
		u.PUT("/devices/:id/settings", h.UpdateDeviceSettings)
		u.DELETE("/devices/:id", h.UnpairDevice)

		u.POST("/readings", h.CreateReading)
		u.GET("/readings", h.ListReadings)
		u.GET("/readings/:id", h.GetReading)
		u.POST("/readings/:id/answers", h.AnswerReading)
		u.POST("/readings/:id/writing-task", practiceLimit, h.ReadingWritingTask)
		u.DELETE("/readings/:id", h.DeleteReading)

		// practise forgotten words (AI: one request per batch)
		u.GET("/practice/forgotten", h.PracticeForgotten)
		u.POST("/practice/examples", practiceLimit, h.PracticeExamples)
		u.POST("/practice/grade", practiceLimit, h.PracticeGrade)
		u.POST("/practice/write/prompts", practiceLimit, h.WritePrompts)
		u.POST("/practice/cloze", practiceLimit, h.ClozePrompts)
		u.POST("/practice/ask", practiceLimit, h.PracticeAsk)
		u.POST("/practice/cloze/forms", practiceLimit, h.ClozeForms)
		u.POST("/practice/write/grade", practiceLimit, h.WriteGrade)

		// word notes (floating panel) and the personal "My words" deck
		u.GET("/dict", h.DictLookup) // quick Vietnamese meaning (dict.minhqnd.com)
		u.GET("/notes", h.ListNotes)
		u.POST("/notes", h.CreateNote)
		u.PATCH("/notes/:id", h.UpdateNote)
		u.DELETE("/notes/:id", h.DeleteNote)
		u.POST("/notes/fill", practiceLimit, h.FillNotes)
		u.POST("/notes/add", practiceLimit, h.AddNotes)

		// per-course free-form notes / cheat sheets (Markdown) + AI summary
		u.GET("/course-notes", h.ListCourseNotes)
		u.POST("/course-notes", h.CreateCourseNote)
		u.PATCH("/course-notes/:id", h.UpdateCourseNote)
		u.DELETE("/course-notes/:id", h.DeleteCourseNote)
		u.POST("/course-notes/summarize", practiceLimit, h.SummarizeCourseNotes)
		u.POST("/course-notes/adopt-summary", h.AdoptSummary)
		u.POST("/course-notes/ask", practiceLimit, h.AskCourse)
		u.GET("/chat-models", h.ChatModels)

		u.GET("/writing/phrases", h.WritingPhrases)              // phrase bank (Writing page panel)
		u.POST("/writing/phrases/study", h.StudyWritingPhrases) // add the whole bank to Study
		u.POST("/writing", h.SubmitWriting)
		u.GET("/writing", h.ListWriting)
		u.GET("/writing/:id", h.GetWriting)
	}

	// ---- admin: manage content
	a := api.Group("", h.RequireUser, h.RequireAdmin)
	{
		a.POST("/subjects", h.CreateSubject)
		a.PUT("/subjects/:id", h.UpdateSubject)
		a.POST("/decks", h.CreateDeck)
		a.PUT("/decks/:id", h.UpdateDeck)
		a.POST("/decks/:id/generate", h.GenerateCards)
		a.POST("/decks/:id/outline", h.OutlineDeck)
		a.POST("/cards", h.CreateCard)
		a.PUT("/cards/:id", h.UpdateCard)
		a.DELETE("/cards/:id", h.DeleteCard)
		a.GET("/admin/audio", h.AudioStats)
		a.GET("/admin/audio/test", h.AudioTest)
		a.DELETE("/admin/audio", h.AudioPurge)
	}

	return r
}
