// Browser Web Speech API composable - FREE TTS (no Google Cloud cost)
import { ref } from 'vue'

/** Speech rates: automatic reading is normal speed; when the user asks to hear it
 *  again (play / speaker buttons, "Try again") it is slower so it can be followed
 *  and repeated; the 🐢 button is slower still. */
export const RATE_NORMAL = 1
export const RATE_REPLAY = 0.8
export const RATE_SLOW = 0.65

export function useSpeechSynthesis() {
  const isSpeaking = ref(false)
  const voicesLoaded = ref(false)
  const availableVoices = ref<SpeechSynthesisVoice[]>([])

  // Load available voices (lazy - browser loads them async)
  function loadVoices(): Promise<SpeechSynthesisVoice[]> {
    return new Promise((resolve) => {
      if (typeof speechSynthesis === 'undefined') return resolve([])
      const voices = speechSynthesis.getVoices()
      if (voices.length > 0) {
        availableVoices.value = voices
        voicesLoaded.value = true
        resolve(voices)
        return
      }
      // Some browsers never fire "voiceschanged": don't wait forever,
      // speak with the default voice after 1s instead.
      const timer = setTimeout(() => resolve([]), 1000)
      speechSynthesis.addEventListener('voiceschanged', () => {
        clearTimeout(timer)
        const v = speechSynthesis.getVoices()
        availableVoices.value = v
        voicesLoaded.value = true
        resolve(v)
      }, { once: true })
    })
  }

  // Find best voice for a language
  function findVoice(langPrefix: string): SpeechSynthesisVoice | null {
    const voices = availableVoices.value
    // Prefer Google voices, then Microsoft, then any
    const google = voices.find(v => v.lang.startsWith(langPrefix) && v.name.includes('Google'))
    if (google) return google
    const ms = voices.find(v => v.lang.startsWith(langPrefix) && v.name.includes('Microsoft'))
    if (ms) return ms
    return voices.find(v => v.lang.startsWith(langPrefix)) || null
  }

  // English voices, best first (Google, Microsoft, others), for multi-speaker dialogues
  function englishVoices(): SpeechSynthesisVoice[] {
    const en = availableVoices.value.filter(v => v.lang.toLowerCase().startsWith('en'))
    const rank = (v: SpeechSynthesisVoice) => (v.name.includes('Google') ? 0 : v.name.includes('Microsoft') ? 1 : 2)
    return [...en].sort((a, b) => rank(a) - rank(b))
  }

  // Speak text in given language
  function speak(
    text: string,
    lang: 'en' | 'vi',
    rate: number = 1.0,
    opts: { voice?: SpeechSynthesisVoice | null; pitch?: number } = {},
  ): Promise<void> {
    return new Promise((resolve, reject) => {
      speechSynthesis.cancel() // Stop any current speech

      const utterance = new SpeechSynthesisUtterance(text)
      utterance.rate = rate
      if (opts.pitch) utterance.pitch = opts.pitch

      const langCode = lang === 'en' ? 'en-US' : 'vi-VN'
      const voice = opts.voice ?? findVoice(langCode)
      if (voice) {
        utterance.voice = voice
      }
      utterance.lang = voice?.lang || langCode

      utterance.onstart = () => { isSpeaking.value = true }
      utterance.onend = () => { isSpeaking.value = false; resolve() }
      utterance.onerror = (e) => { isSpeaking.value = false; reject(e) }

      speechSynthesis.speak(utterance)
    })
  }

  // Speak bilingual: English first, then Vietnamese
  async function speakBilingual(enText: string, viText: string, rate: number = 1.0) {
    await loadVoices()
    await speak(enText, 'en', rate)
    // Small pause between languages
    await new Promise(r => setTimeout(r, 500))
    await speak(viText, 'vi', rate)
  }

  // Speak English only
  async function speakEnglish(text: string, rate: number = 1.0) {
    await loadVoices()
    await speak(text, 'en', rate)
  }

  // Speak Vietnamese (only when the learner asks for it, e.g. Reading → Vietnamese → Listen)
  async function speakVietnamese(text: string, rate: number = 1.0) {
    await loadVoices()
    await speak(text, 'vi', rate)
  }

  /** The device has a Vietnamese voice (otherwise the browser reads with a foreign accent or not at all). */
  async function hasVietnameseVoice(): Promise<boolean> {
    await loadVoices()
    return !!findVoice('vi')
  }

  // Speak English with a different voice per speaker (dialogues). If the
  // browser has only one English voice, speakers are told apart by pitch.
  async function speakAs(text: string, speaker: number, rate: number = 1.0) {
    await loadVoices()
    const voices = englishVoices()
    if (voices.length >= 2) {
      await speak(text, 'en', rate, { voice: voices[speaker % voices.length] })
    } else {
      await speak(text, 'en', rate, { pitch: speaker % 2 === 0 ? 1.0 : 0.75 })
    }
  }

  function stop() {
    speechSynthesis.cancel()
    isSpeaking.value = false
  }

  return {
    isSpeaking,
    voicesLoaded,
    availableVoices,
    loadVoices,
    speak,
    speakBilingual,
    speakEnglish,
    speakAs,
    speakVietnamese,
    hasVietnameseVoice,
    stop,
  }
}