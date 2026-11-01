export type CardKind = 'vocab' | 'phrasal' | 'concept' | 'mcq' | 'lab'

export interface Sense {
  level: number
  meaning_en?: string
  meaning_vi?: string
  example_en?: string
  example_vi?: string
  example_id?: number // set when the example is one of the stored AI examples (Study rotates them)
  usage_note?: string
  grammar_note?: string
}

export interface LabStep {
  output?: string // simulated tool/terminal output shown above the prompt (assignment-style steps)
  prompt: string
  hint?: string
  type?: 'command' | 'choice' // default 'command' (typed, checked against patterns)
  patterns?: string[] // regexes, matched case-insensitively anywhere in the trimmed input (type 'command')
  choices?: string[] // answer options (type 'choice')
  answer?: number // correct index into choices (type 'choice')
  explanation?: string
}

export interface Card {
  id: number
  deck_id: number
  kind: CardKind
  front: string
  source: string
  // kind-specific fields, see api/internal/content/payload.go
  payload: {
    ipa?: string
    pos?: string
    topic?: string
    senses?: Sense[]
    verb?: string
    particle?: string
    separable?: boolean
    back?: string
    note_vi?: string
    example?: string
    example_vi?: string // writing phrases (concept cards)
    group?: string // writing phrases: "Lý do", "Email: kết thư", …
    choices?: string[]
    answer?: number
    explanation?: string
    mode?: string // lab cards: 'lab' (default, command drill) or 'assignment' (full walkthrough)
    scenario?: string // lab cards: short framing for the whole lab
    steps?: LabStep[] // lab cards: ordered steps (typed commands and/or multiple-choice decisions)
    reference?: string // lab cards: optional Markdown study notes shown below the runner, for review
    difficulty?: 'co_ban' | 'trung_binh' | 'nang_cao' // lab cards: display-only difficulty badge
  }
}

export interface Progress {
  card_id: number
  state: number // 0 new, 1 learning, 2 review, 3 relearning
  stability: number
  due: string
  reps: number
  lapses: number
  suspended: boolean
}

export interface CardWithProgress extends Card {
  progress: Progress | null
}

export interface Deck {
  id: number
  subject_id: number
  slug: string
  name: string
  description: string
  default_kind: CardKind
  card_count: number
  learning_count: number
}

export interface Subject {
  id: number
  slug: string
  name: string
  description: string
  decks: Deck[]
}

export interface QueueItem {
  card: Card
  progress: Progress
  is_new: boolean
  preview: Record<string, string>
  /** 4 meanings for multiple choice; the correct one has id === card.id */
  choices?: { id: number; text: string }[]
}

export const RATINGS = [
  { value: 1, label: 'Again', vi: 'Quên', cls: 'btn-outline-danger' },
  { value: 2, label: 'Hard', vi: 'Khó', cls: 'btn-outline-warning' },
  { value: 3, label: 'Good', vi: 'Nhớ', cls: 'btn-outline-success' },
  { value: 4, label: 'Easy', vi: 'Dễ', cls: 'btn-outline-primary' },
]

export const LEVELS = [
  { level: 1, label: 'L1 Simple', color: '#28a745' },
  { level: 2, label: 'L2 Everyday', color: '#fd7e14' },
  { level: 3, label: 'L3 Formal', color: '#6f42c1' },
]

export function stateLabel(p: Progress | null): string {
  if (!p) return ''
  if (p.suspended) return 'Paused'
  return ['New', 'Learning', 'Review', 'Relearning'][p.state] ?? ''
}
