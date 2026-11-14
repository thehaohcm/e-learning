-- 0084_pmp_practice_mcq: add a PMP practice-question deck ("Practice") plus a
-- seed set of original, PMI-mindset multiple-choice questions for exam prep.
-- The PMP subject so far only had the "cheatsheet" deck (0064) and the
-- AI-generated pmp_daily_sets; there was no static, reviewable MCQ bank.
--
-- Questions are written in ENGLISH ONLY (the real PMP exam language), original
-- (no real exam items — same policy noted in 0066), scenario/mindset style,
-- spanning the ECO domains People / Process / Business Environment plus
-- Agile-Hybrid and EVM calculations. Card kind 'mcq', payload
-- {choices, answer(index), explanation} — identical shape to the AWS practice
-- deck (0045). Idempotent: deck via ON CONFLICT (subject_id, slug); cards via
-- ON CONFLICT on ux_cards_deck_kind_front (deck_id, kind, lower(front)).

-- Practice deck for the PMP subject.
INSERT INTO decks (subject_id, slug, name, description, default_kind, sort_order)
SELECT s.id, 'practice', 'Practice (MCQ)',
       'Original PMP-style practice questions (English), by ECO domain + Agile + EVM. Studyable and scored.',
       'mcq', 1
FROM subjects s WHERE s.slug = 'pmp'
ON CONFLICT (subject_id, slug) DO NOTHING;

-- ===== People domain =====
INSERT INTO cards (deck_id, kind, front, payload, source)
SELECT d.id, 'mcq', q.front, q.payload, 'seed'
FROM (VALUES
  ('A new member of an Agile team keeps getting interrupted and talked over during the daily standup. What should the project manager / Scrum Master do first?',
   '{"choices": ["Report the behavior to the functional manager", "Remove the disruptive members from the team", "Facilitate a safe environment and coach the team on respectful turn-taking", "Ignore it because standups are short"], "answer": 2, "explanation": "A servant leader protects the team environment. The first move is to facilitate psychological safety and coach the team, not to escalate or remove people."}'::jsonb),
  ('Two senior engineers strongly disagree on a technical approach and tension is rising. Which conflict-resolution approach gives the most durable resolution?',
   '{"choices": ["Forcing / directing a decision", "Collaborating / problem solving", "Avoiding / withdrawing", "Smoothing / accommodating"], "answer": 1, "explanation": "PMI favors collaborate / problem solve (confront the problem) because it addresses root causes and produces a win-win that lasts, unlike forcing or smoothing."}'::jsonb),
  ('A newly formed team is arguing about roles and showing low output. Which Tuckman stage is this?',
   '{"choices": ["Forming", "Storming", "Norming", "Performing"], "answer": 1, "explanation": "Storming is the conflict stage where members push back on roles and approach. The leader coaches the team through it toward Norming and Performing."}'::jsonb),
  ('What is the primary role of a servant leader on a project?',
   '{"choices": ["Assign tasks and track individual hours", "Remove impediments and support the team", "Make all the technical decisions", "Maximize each person utilization"], "answer": 1, "explanation": "Servant leadership centers on serving the team: removing blockers, shielding from distractions, and enabling self-organization, rather than command-and-control."}'::jsonb),
  ('A stakeholder has high power and high interest in the project. What engagement strategy fits the power/interest grid?',
   '{"choices": ["Monitor with minimum effort", "Keep informed", "Keep satisfied", "Manage closely"], "answer": 3, "explanation": "High power + high interest = manage closely: engage fully and frequently. Keep satisfied is for high power/low interest; keep informed is low power/high interest."}'::jsonb),
  ('Which tool compares the CURRENT and the DESIRED engagement level of each stakeholder (for example from Resistant to Supportive)?',
   '{"choices": ["RACI chart", "Stakeholder Engagement Assessment Matrix", "Power/Interest grid", "Work Breakdown Structure"], "answer": 1, "explanation": "The Stakeholder Engagement Assessment Matrix marks current (C) and desired (D) engagement (Unaware, Resistant, Neutral, Supportive, Leading). The power/interest grid only classifies; RACI maps responsibility."}'::jsonb),
  ('A team member lacks a skill the project needs. What is the best response?',
   '{"choices": ["Replace the member right away", "Permanently assign only easy work", "Provide training and mentoring and update the plan", "Escalate to the sponsor"], "answer": 2, "explanation": "Developing the team is a core PM responsibility. Provide training/mentoring and update the resource plan; replacement or escalation is not the first step."}'::jsonb),
  ('In Herzberg two-factor theory, which is a HYGIENE factor (its absence causes dissatisfaction but its presence does not motivate)?',
   '{"choices": ["Recognition", "The work itself", "Salary and working conditions", "Opportunity for growth"], "answer": 2, "explanation": "Salary, company policy, and working conditions are hygiene factors. Motivators are recognition, achievement, the work itself, responsibility and growth."}'::jsonb),
  ('A member of a distributed virtual team repeatedly misunderstands written messages. What is the best first step?',
   '{"choices": ["Switch the whole project to email only", "Establish communication ground rules and use richer technology such as video", "Ask the member to work on-site", "Reduce the number of status updates"], "answer": 1, "explanation": "For virtual teams, invest in communication technology and set ground rules. Richer channels (video) reduce misunderstanding; forcing relocation or cutting updates does not."}'::jsonb),
  ('A team member consistently arrives late to meetings, which is affecting the team. What should the project manager do first?',
   '{"choices": ["Raise it publicly in the next standup", "Escalate to HR immediately", "Talk privately and directly with the person", "Remove the person from the team"], "answer": 2, "explanation": "Address performance issues privately and directly first. Public call-outs and immediate escalation damage trust and skip the respectful direct conversation PMI expects."}'::jsonb),
  ('In a self-organizing Agile team, the project manager or Scrum Master mainly acts as a:',
   '{"choices": ["Facilitator and coach", "Task assigner", "Final technical authority", "Gatekeeper for all decisions"], "answer": 0, "explanation": "Agile leadership is facilitation and coaching. The team self-organizes to decide how to do the work; the leader removes impediments and enables them."}'::jsonb),
  ('A manager who assumes people are self-motivated, seek responsibility, and can be trusted to self-direct is applying which theory?',
   '{"choices": ["McGregor Theory X", "McGregor Theory Y", "Maslow safety needs", "Expectancy theory"], "answer": 1, "explanation": "Theory Y assumes employees are intrinsically motivated and seek responsibility. Theory X assumes they must be closely controlled and coerced."}'::jsonb)
) AS q(front, payload)
JOIN decks d ON d.slug = 'practice' AND d.subject_id = (SELECT id FROM subjects WHERE slug = 'pmp')
ON CONFLICT DO NOTHING;

-- ===== Process domain =====
INSERT INTO cards (deck_id, kind, front, payload, source)
SELECT d.id, 'mcq', q.front, q.payload, 'seed'
FROM (VALUES
  ('Midway through execution the customer requests a change to the scope. What should the project manager do first?',
   '{"choices": ["Implement it immediately to keep the customer happy", "Reject it to protect the baseline", "Submit a change request and evaluate it through integrated change control", "Defer it to a future project"], "answer": 2, "explanation": "All changes go through Perform Integrated Change Control: log a change request, assess impact across scope/schedule/cost, and get a decision (often from the CCB) before acting."}'::jsonb),
  ('The team adds small extra features the customer did not request because they are easy. What is this called and why avoid it?',
   '{"choices": ["Gold plating, which adds risk and unapproved scope", "Fast tracking, which saves time", "Progressive elaboration, which is required", "Rolling wave planning, a good practice"], "answer": 0, "explanation": "Gold plating is adding unrequested extras. It consumes budget, adds risk, and bypasses change control. Deliver to the agreed scope."}'::jsonb),
  ('Who formally issues the project charter and authorizes the project manager to use resources?',
   '{"choices": ["The project manager", "The sponsor or initiator", "The project team", "The PMO administrator"], "answer": 1, "explanation": "The sponsor (or initiator) signs the charter, which authorizes the project and empowers the project manager."}'::jsonb),
  ('Which baseline is the combination of the scope, schedule, and cost baselines used to measure performance?',
   '{"choices": ["The project management plan", "The Performance Measurement Baseline", "The cost baseline only", "The requirements traceability matrix"], "answer": 1, "explanation": "The Performance Measurement Baseline (PMB) integrates scope, schedule, and cost baselines and is the reference for earned value analysis."}'::jsonb),
  ('A project has BAC = 100k, and at the status date PV = 40k, EV = 36k, AC = 45k. What is the CPI?',
   '{"choices": ["0.80", "0.90", "1.25", "1.10"], "answer": 0, "explanation": "CPI = EV / AC = 36 / 45 = 0.80. Below 1.0 means the project is getting less value than the money spent (over budget)."}'::jsonb),
  ('Using the same data (PV = 40k, EV = 36k, AC = 45k), what is the SPI?',
   '{"choices": ["0.80", "0.90", "1.11", "1.25"], "answer": 1, "explanation": "SPI = EV / PV = 36 / 40 = 0.90. Below 1.0 means the project is behind schedule."}'::jsonb),
  ('A CPI of 0.80 tells you the project is:',
   '{"choices": ["Under budget", "Over budget", "Ahead of schedule", "Exactly on budget"], "answer": 1, "explanation": "CPI below 1.0 means you are spending more than the value earned, so the project is over budget."}'::jsonb),
  ('With BAC = 100k and CPI = 0.80, if current cost performance is expected to continue, what is the EAC?',
   '{"choices": ["80k", "100k", "125k", "145k"], "answer": 2, "explanation": "When variances are typical, EAC = BAC / CPI = 100 / 0.80 = 125k."}'::jsonb),
  ('With BAC = 100k and EAC = 125k, what is the Variance At Completion (VAC)?',
   '{"choices": ["+25k", "-25k", "0", "-45k"], "answer": 1, "explanation": "VAC = BAC - EAC = 100 - 125 = -25k. A negative VAC forecasts finishing over budget."}'::jsonb),
  ('The critical path of a schedule network is:',
   '{"choices": ["The cheapest path", "The longest path, with zero total float", "The path with the most resources", "Any path the sponsor chooses"], "answer": 1, "explanation": "The critical path is the longest duration path through the network and normally has zero total float; it determines the shortest possible project duration."}'::jsonb),
  ('To shorten the schedule by adding resources to critical activities, accepting higher cost, the technique is:',
   '{"choices": ["Fast tracking", "Crashing", "Resource leveling", "Using leads and lags"], "answer": 1, "explanation": "Crashing adds resources to compress duration at added cost. Fast tracking instead runs activities in parallel and adds risk."}'::jsonb),
  ('Fast tracking a schedule primarily increases:',
   '{"choices": ["Cost", "Risk and rework", "Scope", "Quality"], "answer": 1, "explanation": "Fast tracking overlaps activities that were planned in sequence, which raises the risk of rework. Crashing is the one that mainly adds cost."}'::jsonb),
  ('Buying insurance to deal with a threat is which risk response?',
   '{"choices": ["Avoid", "Transfer", "Mitigate", "Accept"], "answer": 1, "explanation": "Transfer shifts the impact (and ownership of the response) to a third party, for example via insurance, warranties, or contracts."}'::jsonb),
  ('To make a positive risk (opportunity) definitely occur, the appropriate response is:',
   '{"choices": ["Enhance", "Share", "Exploit", "Accept"], "answer": 2, "explanation": "Exploit aims to ensure the opportunity happens. Enhance increases probability/impact, share assigns it to a partner better able to capture it, accept takes no action."}'::jsonb),
  ('A new risk that appears as a direct result of implementing a risk response is a:',
   '{"choices": ["Residual risk", "Secondary risk", "Trigger", "Risk owner"], "answer": 1, "explanation": "A secondary risk arises from implementing a response. Residual risk is what remains after the response is applied."}'::jsonb),
  ('According to cost of quality, PMI prefers to spend on:',
   '{"choices": ["Inspection to catch defects late", "Prevention to avoid defects", "Rework after delivery", "Warranty claims"], "answer": 1, "explanation": "Prevention (training, good process) is cheaper than appraisal/inspection and far cheaper than internal/external failure costs."}'::jsonb),
  ('Formal acceptance of completed deliverables by the customer happens in which process?',
   '{"choices": ["Control Quality", "Validate Scope", "Close Project", "Manage Quality"], "answer": 1, "explanation": "Validate Scope is the customer formally accepting deliverables. Control Quality is the internal check for correctness that normally precedes it."}'::jsonb),
  ('In a Firm Fixed Price (FFP) contract, who carries the cost risk?',
   '{"choices": ["The buyer", "The seller", "Shared equally", "The PMO"], "answer": 1, "explanation": "In FFP the price is fixed, so the seller absorbs cost overruns and carries the cost risk. In cost-reimbursable contracts the buyer carries more of it."}'::jsonb),
  ('When the scope cannot be well defined at contract signing, the most appropriate contract type is:',
   '{"choices": ["Firm Fixed Price", "Cost-reimbursable or Time and Materials", "Fixed Price Incentive Fee", "Purchase order"], "answer": 1, "explanation": "When scope is uncertain, cost-reimbursable or T&M contracts fit because a fixed price cannot be set fairly without a defined scope."}'::jsonb),
  ('How many communication channels exist among 8 stakeholders?',
   '{"choices": ["16", "28", "36", "64"], "answer": 1, "explanation": "Channels = n(n-1)/2 = 8 * 7 / 2 = 28."}'::jsonb),
  ('The risk register is first created during which process?',
   '{"choices": ["Plan Risk Management", "Identify Risks", "Perform Qualitative Risk Analysis", "Monitor Risks"], "answer": 1, "explanation": "Identify Risks creates the risk register. Later processes add analysis, responses, and status to it."}'::jsonb)
) AS q(front, payload)
JOIN decks d ON d.slug = 'practice' AND d.subject_id = (SELECT id FROM subjects WHERE slug = 'pmp')
ON CONFLICT DO NOTHING;

-- ===== Agile / Hybrid =====
INSERT INTO cards (deck_id, kind, front, payload, source)
SELECT d.id, 'mcq', q.front, q.payload, 'seed'
FROM (VALUES
  ('Who is responsible for ordering and prioritizing the product backlog?',
   '{"choices": ["The Scrum Master", "The Product Owner", "The development team", "The project sponsor"], "answer": 1, "explanation": "The Product Owner owns and orders the product backlog to maximize value. The team forecasts how much it can deliver."}'::jsonb),
  ('What is the main purpose of the sprint retrospective?',
   '{"choices": ["Demonstrate the product increment to stakeholders", "Inspect and improve the team process and working agreements", "Plan the scope of the next sprint", "Report status to the sponsor"], "answer": 1, "explanation": "The retrospective is the team improving its own process. Demonstrating the increment to stakeholders is the sprint review."}'::jsonb),
  ('The sprint review is mainly for:',
   '{"choices": ["Improving the team internal process", "Demonstrating the increment to stakeholders and getting feedback", "Assigning tasks for the next sprint", "Updating the project budget"], "answer": 1, "explanation": "The sprint review inspects the product increment with stakeholders and adapts the backlog based on feedback."}'::jsonb),
  ('How should team velocity be used?',
   '{"choices": ["To forecast the team own future capacity", "To compare one team against another", "To rank individual performance", "To set a permanent fixed commitment"], "answer": 0, "explanation": "Velocity is a planning aid for the same team to forecast its capacity. It must not be used to compare teams or judge individuals."}'::jsonb),
  ('In Kanban, limiting Work In Progress (WIP) primarily helps to:',
   '{"choices": ["Expose bottlenecks and improve flow", "Increase the number of tasks started", "Remove the need for a backlog", "Guarantee fixed-length iterations"], "answer": 0, "explanation": "WIP limits surface bottlenecks and pull the team toward finishing work, improving flow and cycle time."}'::jsonb),
  ('During a sprint the team realizes it cannot complete all committed scope. What is the best action?',
   '{"choices": ["Extend the sprint length", "Mandate overtime to finish", "Talk with the Product Owner and reduce scope while keeping the timebox fixed", "Cancel the sprint"], "answer": 2, "explanation": "The sprint timebox is fixed. The team collaborates with the Product Owner to reduce or re-prioritize scope rather than extend the sprint."}'::jsonb),
  ('The Definition of Done is:',
   '{"choices": ["A list of all project requirements", "A shared set of criteria an increment must meet to be releasable", "The product roadmap", "The sprint goal"], "answer": 1, "explanation": "The Definition of Done is a shared quality checklist that every increment must satisfy to be considered complete and potentially releasable."}'::jsonb)
) AS q(front, payload)
JOIN decks d ON d.slug = 'practice' AND d.subject_id = (SELECT id FROM subjects WHERE slug = 'pmp')
ON CONFLICT DO NOTHING;

-- ===== Business Environment =====
INSERT INTO cards (deck_id, kind, front, payload, source)
SELECT d.id, 'mcq', q.front, q.payload, 'seed'
FROM (VALUES
  ('Company internal templates, historical information, and the lessons-learned repository are examples of:',
   '{"choices": ["Enterprise Environmental Factors", "Organizational Process Assets", "The project charter", "The business case"], "answer": 1, "explanation": "Organizational Process Assets (OPA) are the organization internal processes, templates, and knowledge bases. EEFs are external or organizational conditions outside the team control."}'::jsonb),
  ('Government regulations, market conditions, and the organization culture are examples of:',
   '{"choices": ["Organizational Process Assets", "Enterprise Environmental Factors", "Change requests", "Work performance data"], "answer": 1, "explanation": "Enterprise Environmental Factors (EEF) are conditions not under the team control, such as regulations, market, and culture, that influence the project."}'::jsonb),
  ('A mandatory legal compliance requirement conflicts with the project schedule. What should the project manager do?',
   '{"choices": ["Skip compliance to protect the deadline", "Treat compliance as a priority and adjust the plan", "Ask the team to decide", "Hide the conflict from the sponsor"], "answer": 1, "explanation": "Compliance is not optional. The PM prioritizes meeting the legal requirement and re-plans schedule/cost accordingly, communicating transparently."}'::jsonb),
  ('How are the intended benefits of a project tracked over time, often beyond project closure?',
   '{"choices": ["The issue log", "The benefits management plan and business case", "The risk register", "The stakeholder register"], "answer": 1, "explanation": "The benefits management plan (aligned with the business case) defines target benefits, metrics, and when they are realized, often after the product is in use."}'::jsonb),
  ('In a strong matrix organization, the project manager authority is:',
   '{"choices": ["Low, the functional manager controls resources", "Moderate to high", "None, the PM is only a coordinator", "Equal to the sponsor"], "answer": 1, "explanation": "In a strong matrix the PM has moderate-to-high authority and controls much of the budget. In a weak matrix the functional manager dominates."}'::jsonb)
) AS q(front, payload)
JOIN decks d ON d.slug = 'practice' AND d.subject_id = (SELECT id FROM subjects WHERE slug = 'pmp')
ON CONFLICT DO NOTHING;
