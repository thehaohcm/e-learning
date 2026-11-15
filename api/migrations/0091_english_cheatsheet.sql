-- 0091_english_cheatsheet: a starter Cheatsheet for the English subject —
-- study plan, core grammar, TOEIC & IELTS strategy, writing/speaking phrases,
-- common Vietnamese-learner mistakes, casual/forum & email register. Same
-- mechanism as PMP/Pentest cheatsheets. Seeded once for this account, skipping
-- titles that already exist. Expand or AI-generate questions from these.

INSERT INTO decks (subject_id, slug, name, description, default_kind, sort_order)
SELECT s.id, 'cheatsheet', 'Cheatsheet', $cs$English cheatsheet by section (grammar, TOEIC/IELTS strategy, writing/speaking phrases). Editable, studyable, source for AI questions.$cs$, 'concept', 0
FROM subjects s WHERE s.slug = 'english'
ON CONFLICT (subject_id, slug) DO NOTHING;

INSERT INTO course_notes (user_id, subject_id, title, body, sort_order, links, created_at, updated_at)
SELECT u.id, s.id, v.title, v.body, v.ord, '[]'::jsonb, NOW(), NOW()
FROM (VALUES
  ($cs$1. Lộ trình & chiến lược học$cs$, $cs$**Mục tiêu**: học hội thoại + TOEIC song song → lấy TOEIC trước → chuyển sang IELTS. Viết nhắm register casual/forum.

- **Nền từ vựng**: TOEIC 700 cần ~3.000–4.000 word family; IELTS 6.0 cần ~4.000–5.000. Học theo deck + Daily 20.
- **4 kỹ năng**: mỗi ngày chạm cả 4 — nghe (podcast/dialog), nói (shadowing), đọc (bài ngắn), viết (1 đoạn).
- **Điểm yếu nghe/nói**: ưu tiên *shadowing* (nghe 1 câu → nhại lại ngay) và *dictation* (chép chính tả) 15'/ngày.
- **Ôn**: dùng spaced repetition (Study) cho từ vựng; Daily 20 cho câu trắc nghiệm.$cs$, 1),
  ($cs$2. Ngữ pháp cốt lõi — Thì (tenses)$cs$, $cs$12 thì nhưng 80% đề dùng ~6 thì:

- **Present simple**: sự thật, thói quen. *She works here.*
- **Present continuous**: đang xảy ra / tạm thời. *I am studying now.*
- **Present perfect**: quá khứ còn liên quan hiện tại / kinh nghiệm. *I have finished.* (dấu hiệu: just, already, yet, ever, since, for)
- **Past simple**: xong trong quá khứ (có mốc thời gian). *I went yesterday.*
- **Past continuous**: đang xảy ra khi có việc khác chen vào. *I was cooking when he called.*
- **Future (will / be going to)**: will = quyết định tức thì/dự đoán; be going to = kế hoạch/định trước.$cs$, 2),
  ($cs$3. Ngữ pháp — Mạo từ, giới từ, đại từ$cs$, $cs$- **a/an**: danh từ đếm được số ít, lần đầu nhắc. **the**: xác định / đã nhắc / duy nhất. Không the với danh từ chung chung số nhiều/không đếm được (*I like music*).
- **Giới từ thời gian**: **at** (giờ: at 5pm), **on** (ngày/thứ: on Monday), **in** (tháng/năm/buổi: in May, in the morning).
- **Giới từ nơi chốn**: at (điểm), on (bề mặt), in (không gian bao quanh).
- **Đại từ quan hệ**: who (người), which (vật), that (cả hai), whose (sở hữu), where (nơi chốn).$cs$, 3),
  ($cs$4. Ngữ pháp — Điều kiện, bị động, mệnh đề$cs$, $cs$**Câu điều kiện**:
- Type 1 (có thật tương lai): If + present, will + V. *If it rains, I will stay.*
- Type 2 (không có thật hiện tại): If + past, would + V. *If I were rich, I would travel.*
- Type 3 (không có thật quá khứ): If + past perfect, would have + V3. *If I had studied, I would have passed.*

**Bị động**: be + V3. *The report was written by Tom.* Dùng khi không rõ/không cần nói ai làm.

**Mệnh đề quan hệ**: rút gọn bằng V-ing (chủ động) / V3 (bị động): *The man **standing** there…*, *The book **written** by her…*$cs$, 4),
  ($cs$5. TOEIC — Listening (Part 1-4)$cs$, $cs$- **Part 1 (ảnh)**: nghe động từ + chủ ngữ; cẩn thận bẫy từ đồng âm.
- **Part 2 (hỏi-đáp)**: nhận diện từ để hỏi (Wh-) → loại đáp án Yes/No cho câu Wh-. Bẫy: lặp lại từ trong câu hỏi.
- **Part 3-4 (hội thoại/bài nói)**: đọc câu hỏi TRƯỚC khi nghe; nghe ý chính + chi tiết (số, thời gian, lý do). Từ khóa hỏi: who/where/why/what next.
- **Chiến lược**: không dừng lại ở câu đã lỡ — đi tiếp. Shadowing để quen tốc độ.$cs$, 5),
  ($cs$6. TOEIC — Reading (Part 5-7)$cs$, $cs$- **Part 5 (điền 1 câu)**: nhìn loại từ cần điền (noun/verb/adj/adv), chia thì, giới từ, liên từ. ~30s/câu.
- **Part 6 (điền đoạn)**: như Part 5 + 1 câu điền nguyên câu → đọc ngữ cảnh trước/sau.
- **Part 7 (đọc hiểu)**: skim câu hỏi → quét (scan) từ khóa trong bài. Câu NOT/EXCEPT đọc kỹ. Double/triple passage: liên kết thông tin giữa các văn bản.
- **Thời gian**: ~75 phút cho 100 câu — không sa đà 1 câu.$cs$, 6),
  ($cs$7. IELTS Writing Task 1$cs$, $cs$Mô tả biểu đồ/bảng/quy trình (≥150 từ, 20 phút).

- **Cấu trúc**: Introduction (paraphrase đề) → Overview (2-3 xu hướng chính, KHÔNG số liệu) → Details (2 đoạn, có số liệu).
- **Cụm xu hướng**: tăng (increase, rise, climb, surge), giảm (decrease, decline, drop, fall), ổn định (remain stable/steady), dao động (fluctuate), đạt đỉnh (peak at), chạm đáy (bottom out).
- **So sánh**: twice as high as, the highest/lowest, compared to, whereas, while.$cs$, 7),
  ($cs$8. IELTS Writing Task 2$cs$, $cs$Essay quan điểm (≥250 từ, 40 phút) — chiếm 2/3 điểm Writing.

- **Cấu trúc 4 đoạn**: Intro (paraphrase + thesis) → Body 1 (ý 1 + ví dụ) → Body 2 (ý 2 + ví dụ) → Conclusion (tóm tắt).
- **Dạng đề**: opinion (agree/disagree), discussion (both views), problem-solution, advantage-disadvantage.
- **Linking**: Firstly, Moreover, However, On the other hand, For instance, As a result, In conclusion.
- **Tiêu chí**: Task Response, Coherence, Lexical Resource, Grammar. Dùng từ vựng đa dạng + câu phức.$cs$, 8),
  ($cs$9. IELTS Speaking (Part 1-3)$cs$, $cs$- **Part 1 (4-5 phút)**: câu hỏi quen thuộc (home, work, hobby). Trả lời 2-3 câu, thêm lý do/ví dụ.
- **Part 2 (cue card, nói 1-2 phút)**: 1 phút chuẩn bị — note nhanh who/what/when/where/why/feeling. Nói liền mạch.
- **Part 3 (thảo luận)**: trả lời sâu, có lập luận 2 chiều.
- **Cụm hữu ích**: *To be honest…, It depends on…, As far as I'm concerned…, The thing is…, That's a good question.* Dùng khi câu giờ thay vì im lặng.
- **Tiêu chí**: Fluency, Vocabulary, Grammar, Pronunciation. Trôi chảy > hoàn hảo.$cs$, 9),
  ($cs$10. Collocations & Phrasal verbs$cs$, $cs$**Collocations** (cụm từ đi với nhau tự nhiên):
- make a decision, take a risk, do homework, pay attention, keep in touch, meet a deadline, make progress.

**Phrasal verbs** hay gặp:
- look into (điều tra), carry out (thực hiện), put off (hoãn), come up with (nghĩ ra), figure out (hiểu ra), deal with (xử lý), set up (thiết lập), turn down (từ chối).$cs$, 10),
  ($cs$11. Linking words (từ nối)$cs$, $cs$- **Thêm ý**: moreover, furthermore, in addition, besides.
- **Tương phản**: however, nevertheless, on the other hand, whereas, although.
- **Nguyên nhân-kết quả**: because, since, therefore, as a result, consequently, thus.
- **Ví dụ**: for example, for instance, such as.
- **Kết luận**: in conclusion, to sum up, overall, all in all.
- **Trình tự**: firstly, then, after that, finally.$cs$, 11),
  ($cs$12. Lỗi thường gặp của người Việt$cs$, $cs$- Quên **-s** ngôi thứ 3 số ít: *He go* → *He goes*.
- Quên **-ed/V2** ở quá khứ.
- Nhầm **a/an/the** (bỏ mạo từ): *I am student* → *I am **a** student*.
- Dịch word-by-word từ tiếng Việt → sai collocation.
- Phát âm thiếu âm cuối (final sound): *work, like, used* — người nghe mất thông tin.
- Nhầm **much/many**, **fewer/less**, **since/for**.$cs$, 12),
  ($cs$13. Văn phong casual / forum & email$cs$, $cs$**Casual (Reddit/comment)**:
- Viết ngắn, tự nhiên: *Honestly, …*, *IMO* (in my opinion), *TBH* (to be honest), *btw*, *imo*, *lol*.
- Dùng contraction: *I'm, don't, it's, gonna, wanna* (chỉ trong casual).
- Không cần câu hoàn chỉnh mọi lúc; giọng thân thiện.

**Email (formal)**:
- Mở: *Dear…, I hope this email finds you well.*
- Thân: *I am writing to…, Could you please…, I would appreciate it if…*
- Kết: *Best regards, Looking forward to your reply.*$cs$, 13)
) AS v(title, body, ord)
JOIN users u ON u.email = 'thehaohcm@tradesign.io.vn'
JOIN subjects s ON s.slug = 'english'
WHERE NOT EXISTS (
    SELECT 1 FROM course_notes cn
    WHERE cn.user_id = u.id AND cn.subject_id = s.id AND cn.title = v.title
);
