export type CardKind = 'vocab' | 'phrasal' | 'concept' | 'mcq'

export interface Sense {
  level: number
  meaning_en?: string
  meaning_vi?: string
  example_en?: string
  example_vi?: string
  usage_note?: string
  grammar_note?: string
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
    choices?: string[]
    answer?: number
    explanation?: string
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
