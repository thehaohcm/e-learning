/** One sentence of a reading or one line of a conversation, for SpeakingPractice. */
export interface PracticeItem {
  index: number // sentence/line index, used to highlight it in the page
  text: string
  vi?: string
  speaker?: string
}
