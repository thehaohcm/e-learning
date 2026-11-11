package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/ai"

	"github.com/gin-gonic/gin"
)

// Roleplay (PMP): two modes, nothing stored (the browser keeps the
// conversation and sends it with every request).
//   - "client" (default): the learner plays a project manager in English
//     with an AI-played difficult client. Scenarios are fixed, or "custom"
//     for a situation the learner types themselves.
//   - "teach": the learner explains a PMP concept (in Vietnamese) to an
//     AI-played junior colleague, to practise teaching it back. Topics are
//     fixed, or "custom" for a concept the learner types themselves.

type roleplayReq struct {
	Mode       string            `json:"mode"` // "client" (default) or "teach"
	ScenarioID string            `json:"scenario_id"`
	TopicID    string            `json:"topic_id"`
	Persona    string            `json:"persona"`
	Model      string            `json:"model"`
	History    []ai.RoleplayTurn `json:"history"`

	// client mode, scenario_id == "custom": a situation the learner typed themselves.
	CustomTitle  string `json:"custom_title"`
	CustomSetup  string `json:"custom_setup"`
	CustomGoal   string `json:"custom_goal"`
	CustomClient string `json:"custom_client"`

	// teach mode, topic_id == "custom": a concept the learner typed themselves.
	CustomTopicTitle string `json:"custom_topic_title"`
	CustomTopicBrief string `json:"custom_topic_brief"`
}

const (
	maxRoleplayTurns = 30
	maxRoleplayText  = 800
	maxCustomField   = 1500
)

// RoleplayScenarios: GET /api/roleplay/scenarios
func (h *Handler) RoleplayScenarios(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"scenarios":      ai.RoleplayScenarios,
		"personas":       ai.RoleplayPersonas,
		"teach_topics":   ai.TeachTopics,
		"teach_personas": ai.TeachPersonas,
	})
}

// roleplayBind reads and validates the request body shared by turn +
// feedback. Returns the mode, plus either the client-mode scenario or the
// teach-mode topic (whichever the mode uses), built from a fixed id or from
// the learner's own "custom" text.
func (h *Handler) roleplayBind(c *gin.Context) (roleplayReq, string, ai.RoleplayScenario, ai.TeachTopic, bool) {
	var req roleplayReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return req, "", ai.RoleplayScenario{}, ai.TeachTopic{}, false
	}
	mode := "client"
	if req.Mode == "teach" {
		mode = "teach"
	}

	var sc ai.RoleplayScenario
	var tp ai.TeachTopic

	if mode == "teach" {
		if req.TopicID == "custom" {
			title := clip(req.CustomTopicTitle, maxCustomField)
			if title == "" {
				abort(c, http.StatusBadRequest, "nhập chủ đề bạn muốn tự luyện giải thích")
				return req, mode, sc, tp, false
			}
			brief := clip(req.CustomTopicBrief, maxCustomField)
			if brief == "" {
				brief = fmt.Sprintf("Em chưa rõ lắm về %s, anh/chị giải thích cho em với?", title)
			}
			tp = ai.TeachTopic{ID: "custom", Title: title, Brief: brief,
				Checklist: "Đánh giá dựa trên tính chính xác, logic và sự rõ ràng chung của lời giải thích so với kiến thức PMP/PMBOK chuẩn về chủ đề này."}
		} else {
			t, ok := ai.FindTeachTopic(req.TopicID)
			if !ok {
				abort(c, http.StatusBadRequest, "unknown topic")
				return req, mode, sc, tp, false
			}
			tp = t
		}
	} else {
		if req.ScenarioID == "custom" {
			setup := clip(req.CustomSetup, maxCustomField)
			if setup == "" {
				abort(c, http.StatusBadRequest, "nhập tình huống bạn muốn luyện")
				return req, mode, sc, tp, false
			}
			title := clip(req.CustomTitle, 200)
			if title == "" {
				title = "Tình huống của bạn"
			}
			goal := clip(req.CustomGoal, maxCustomField)
			if goal == "" {
				goal = "Xử lý tình huống hợp lý theo PMBOK: làm rõ vấn đề, đánh giá tác động, đưa ra phương án và đi đến một quyết định hoặc bước tiếp theo cụ thể."
			}
			clientBrief := clip(req.CustomClient, maxCustomField)
			if clientBrief == "" {
				clientBrief = "Suy ra một lập trường hợp lý, nhất quán cho client/đối phương dựa trên tình huống dưới đây, và giữ thái độ đó xuyên suốt cuộc trò chuyện."
			}
			sc = ai.RoleplayScenario{
				ID: "custom", Title: title, Setup: setup, SetupVI: setup,
				Goal: goal, Topics: "Tình huống tự nhập", Client: clientBrief,
			}
		} else {
			s, ok := ai.FindRoleplayScenario(req.ScenarioID)
			if !ok {
				abort(c, http.StatusBadRequest, "unknown scenario")
				return req, mode, sc, tp, false
			}
			sc = s
		}
	}

	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return req, mode, sc, tp, false
	}
	clean := make([]ai.RoleplayTurn, 0, len(req.History))
	for _, t := range req.History {
		text := strings.TrimSpace(t.Text)
		if text == "" || (t.Role != "client" && t.Role != "pm") {
			continue
		}
		if len(text) > maxRoleplayText {
			text = text[:maxRoleplayText]
		}
		clean = append(clean, ai.RoleplayTurn{Role: t.Role, Text: text})
	}
	if len(clean) > maxRoleplayTurns {
		clean = clean[len(clean)-maxRoleplayTurns:]
	}
	req.History = clean
	return req, mode, sc, tp, true
}

// RoleplayTurn: POST /api/roleplay/turn
// {scenario_id, persona, model?, history:[{role,text}]} -> {"reply": "..."}
// An empty history asks for the client's opening line; otherwise the last
// history entry must be the PM's.
func (h *Handler) RoleplayTurn(c *gin.Context) {
	req, mode, sc, tp, ok := h.roleplayBind(c)
	if !ok {
		return
	}
	if n := len(req.History); n > 0 && req.History[n-1].Role != "pm" {
		abort(c, http.StatusBadRequest, "the last line must be from the PM")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	model := strings.TrimSpace(req.Model)
	if model != "" && !h.chatModelAllowed(ctx, model) {
		abort(c, http.StatusBadRequest, "that model is not available")
		return
	}
	var reply string
	var err error
	if mode == "teach" {
		reply, err = h.AI.TeachTurn(ctx, model, tp, req.Persona, req.History)
	} else {
		reply, err = h.AI.RoleplayReply(ctx, model, sc, req.Persona, req.History)
	}
	if err != nil {
		if errors.Is(err, ai.ErrNotConfigured) {
			abort(c, http.StatusServiceUnavailable, err.Error())
			return
		}
		abort(c, http.StatusBadGateway, "AI could not reply, please try again")
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

// RoleplayFeedback: POST /api/roleplay/feedback
// {scenario_id, model?, history} -> {"feedback": {...}}
func (h *Handler) RoleplayFeedback(c *gin.Context) {
	req, mode, sc, tp, ok := h.roleplayBind(c)
	if !ok {
		return
	}
	pm := 0
	for _, t := range req.History {
		if t.Role == "pm" {
			pm++
		}
	}
	if pm == 0 {
		abort(c, http.StatusBadRequest, "say something as the PM first")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 150*time.Second)
	defer cancel()
	model := strings.TrimSpace(req.Model)
	if model != "" && !h.chatModelAllowed(ctx, model) {
		abort(c, http.StatusBadRequest, "that model is not available")
		return
	}
	var fb ai.RoleplayFeedback
	var err error
	if mode == "teach" {
		fb, err = h.AI.TeachReview(ctx, model, tp, req.History)
	} else {
		fb, err = h.AI.RoleplayReview(ctx, model, sc, req.History)
	}
	if err != nil {
		if errors.Is(err, ai.ErrNotConfigured) {
			abort(c, http.StatusServiceUnavailable, err.Error())
			return
		}
		abort(c, http.StatusBadGateway, "AI could not review the conversation, please try again")
		return
	}
	c.JSON(http.StatusOK, gin.H{"feedback": fb})
}
