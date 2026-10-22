// Browser speech-to-text (Web Speech API). Free, no server work.
// Works in Chrome / Edge (desktop + Android) and Safari 14.1+. Firefox has no support.
// Needs HTTPS (or localhost) for microphone access. In Chrome the audio is sent
// to Google's speech service to be transcribed.
import { onUnmounted, ref } from 'vue'

// Minimal typings: the API is still prefixed in most browsers and not in lib.dom.
interface RecAlternative { transcript: string; confidence: number }
interface RecResult { isFinal: boolean; length: number; [i: number]: RecAlternative }
interface RecEvent { resultIndex: number; results: { length: number; [i: number]: RecResult } }
interface RecErrorEvent { error: string }
interface Recognition {
  lang: string
  interimResults: boolean
  continuous: boolean
  maxAlternatives: number
  onresult: ((e: RecEvent) => void) | null
  onerror: ((e: RecErrorEvent) => void) | null
  onend: (() => void) | null
  start(): void
  stop(): void
  abort(): void
}
type RecognitionCtor = new () => Recognition

function ctor(): RecognitionCtor | null {
  const w = window as unknown as { SpeechRecognition?: RecognitionCtor; webkitSpeechRecognition?: RecognitionCtor }
  return w.SpeechRecognition ?? w.webkitSpeechRecognition ?? null
}

export const speechRecognitionSupported = typeof window !== 'undefined' && ctor() !== null

const ERRORS: Record<string, string> = {
  'not-allowed': 'Microphone is blocked. Allow it in the browser (lock icon in the address bar). It only works on HTTPS.',
  'service-not-allowed': 'Speech recognition is not allowed here. It needs HTTPS.',
  'no-speech': "Didn't hear anything. Try again a bit closer to the mic.",
  'audio-capture': 'No microphone found.',
  network: 'Speech service unreachable (network). Try again.',
}

export function useSpeechRecognition() {
  const listening = ref(false)
  const transcript = ref('') // final + interim text of the current attempt
  const alternatives = ref<string[]>([]) // final alternatives (best first)
  const error = ref('')
  let rec: Recognition | null = null

  /** Listen for one sentence. Resolves with the alternatives when the user stops talking. */
  function listen(lang = 'en-US'): Promise<string[]> {
    const C = ctor()
    if (!C) {
      error.value = 'This browser has no speech recognition. Use Chrome or Edge.'
      return Promise.resolve([])
    }
    stop()
    // the app's own TTS would be heard by the mic
    if (typeof speechSynthesis !== 'undefined') speechSynthesis.cancel()

    transcript.value = ''
    alternatives.value = []
    error.value = ''
    return new Promise((resolve) => {
      const r = new C()
      rec = r
      r.lang = lang
      r.interimResults = true
      r.continuous = false
      r.maxAlternatives = 3
      let finals: string[] = []
      r.onresult = (e) => {
        let fin = ''
        let interim = ''
        for (let i = 0; i < e.results.length; i++) {
          const res = e.results[i]
          if (res.isFinal) {
            fin += res[0].transcript
            if (i === e.results.length - 1) {
              finals = []
              for (let j = 0; j < res.length; j++) finals.push(res[j].transcript.trim())
            }
          } else {
            interim += res[0].transcript
          }
        }
        transcript.value = (fin + interim).trim()
      }
      r.onerror = (e) => {
        if (e.error !== 'aborted') error.value = ERRORS[e.error] ?? `Speech recognition error: ${e.error}`
      }
      r.onend = () => {
        if (rec === r) rec = null
        listening.value = false
        if (!finals.length && transcript.value) finals = [transcript.value]
        alternatives.value = finals
        resolve(finals)
      }
      try {
        r.start()
        listening.value = true
      } catch (err) {
        error.value = String(err)
        resolve([])
      }
    })
  }

  /** Stop listening; whatever was heard so far is still evaluated. */
  function stop() {
    rec?.stop()
  }

  onUnmounted(() => rec?.abort())

  return { supported: speechRecognitionSupported, listening, transcript, alternatives, error, listen, stop }
}

// ---------------------------------------------------------------- scoring

export interface WordMark { word: string; ok: boolean }
export interface SpeechScore { words: WordMark[]; matched: number; total: number; pct: number; heard: string }

const NUMBERS: Record<string, string> = {
  '0': 'zero', '1': 'one', '2': 'two', '3': 'three', '4': 'four', '5': 'five',
  '6': 'six', '7': 'seven', '8': 'eight', '9': 'nine', '10': 'ten', '11': 'eleven', '12': 'twelve',
}

/** Lower-case word list: no punctuation, curly quotes normalised, small digits spelled out. */
export function normWords(s: string): string[] {
  return s
    .toLowerCase()
    .replace(/[’‘`]/g, "'")
    .replace(/[^a-z0-9'\s-]/g, ' ')
    .replace(/-/g, ' ')
    .split(/\s+/)
    .map((w) => w.replace(/^'+|'+$/g, ''))
    .filter(Boolean)
    .map((w) => NUMBERS[w] ?? w)
}

/**
 * Word-level comparison of the target sentence with what was heard (longest
 * common subsequence). Each target word is marked ok if it was said in order.
 * Several recognition alternatives are tried and the best one wins.
 */
export function scoreSpeech(target: string, heard: string[]): SpeechScore {
  const display = target.split(/\s+/).filter(Boolean) // keep original punctuation/case for display
  const tWords = display.map((w) => normWords(w).join(' '))
  let best: SpeechScore = { words: display.map((w) => ({ word: w, ok: false })), matched: 0, total: display.length, pct: 0, heard: heard[0] ?? '' }
  for (const h of heard) {
    const hw = normWords(h)
    const t = tWords
    // DP table for LCS; target tokens that are empty (pure punctuation) count as matched
    const n = t.length
    const m = hw.length
    const dp: number[][] = Array.from({ length: n + 1 }, () => new Array<number>(m + 1).fill(0))
    for (let i = n - 1; i >= 0; i--) {
      for (let j = m - 1; j >= 0; j--) {
        dp[i][j] = t[i] && t[i] === hw[j] ? dp[i + 1][j + 1] + 1 : Math.max(dp[i + 1][j], dp[i][j + 1])
      }
    }
    const ok = new Array<boolean>(n).fill(false)
    let i = 0
    let j = 0
    while (i < n && j < m) {
      if (t[i] && t[i] === hw[j]) { ok[i] = true; i++; j++ }
      else if (dp[i + 1][j] >= dp[i][j + 1]) i++
      else j++
    }
    // a multi-word token ("2-day" -> "two day") or merged words: accept if all parts occur
    for (let k = 0; k < n; k++) {
      if (!ok[k] && t[k] && t[k].includes(' ') && t[k].split(' ').every((p) => hw.includes(p))) ok[k] = true
    }
    const scored = t.filter(Boolean).length || 1
    const matched = ok.filter((x, k) => x && t[k]).length
    const pct = Math.round((matched / scored) * 100)
    if (pct > best.pct || (pct === best.pct && !best.heard)) {
      best = { words: display.map((w, k) => ({ word: w, ok: ok[k] || !t[k] })), matched, total: scored, pct, heard: h }
    }
  }
  return best
}
