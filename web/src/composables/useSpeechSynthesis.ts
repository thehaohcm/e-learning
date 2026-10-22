// Browser Web Speech API composable - FREE TTS (no Google Cloud cost)
import { ref } from 'vue'

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

  // Speak text in given language
  function speak(text: string, lang: 'en' | 'vi', rate: number = 1.0): Promise<void> {
    return new Promise((resolve, reject) => {
      speechSynthesis.cancel() // Stop any current speech

      const utterance = new SpeechSynthesisUtterance(text)
      utterance.rate = rate

      const langCode = lang === 'en' ? 'en-US' : 'vi-VN'
      const voice = findVoice(langCode)
      if (voice) {
        utterance.voice = voice
      }
      utterance.lang = langCode

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
    stop,
  }
}