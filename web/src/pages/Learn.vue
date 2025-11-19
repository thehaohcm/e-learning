<template>
  <div>
    <!-- YouGlish Widget Container -->
    <div id="youglish-widget" class="youglish-embed" v-show="showYouglishWidget"></div>
    
    <!-- Header Section -->
    <div class="text-center mb-5">
      <h1 class="display-5 fw-bold text-primary mb-3">
        <i class="bi bi-book"></i> Learn Vocabulary
      </h1>
      <div class="badge bg-secondary fs-6 px-3 py-2">
        <i class="bi bi-tag me-1"></i> Topic: {{ topic }}
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="vocabs.length === 0" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
      <p class="mt-3 text-muted">Loading vocabulary...</p>
    </div>

    <!-- Vocabulary Cards -->
    <div v-else class="row g-4">
      <div v-for="v in vocabs" :key="v.ID" class="col-md-6 col-lg-4">
        <div class="card border-0 shadow-sm h-100 vocabulary-card">
          <div class="card-body">
            <!-- Word Header -->
            <div class="d-flex justify-content-between align-items-start mb-3">
              <h3 class="card-title text-primary fw-bold mb-0">
                {{ v.Word }}
              </h3>
              <div class="d-flex gap-2">
                <!-- YouGlish Video Button -->
                <button
                  @click="loadYouglishWidget(v.Word)"
                  class="btn btn-outline-info btn-sm"
                  :class="{ 'active': currentWord === v.Word }"
                  title="Load YouGlish video"
                >
                  <i class="bi bi-play-circle"></i>
                </button>
                <button
                  @click="select(v.ID)"
                  class="btn btn-outline-success btn-sm"
                  :disabled="selectedWords.includes(v.ID)"
                >
                  <i class="bi" :class="selectedWords.includes(v.ID) ? 'bi-check-circle-fill' : 'bi-plus-circle'"></i>
                  {{ selectedWords.includes(v.ID) ? 'Selected' : 'Select' }}
                </button>
              </div>
            </div>

            <!-- YouGlish Embedded Widget -->
            <div v-if="currentWord === v.Word && showYouglishWidget" class="mb-3 youglish-widget-container">
              <div class="d-flex justify-content-between align-items-center mb-2">
                <small class="text-muted">
                  <i class="bi bi-play-circle me-1"></i>
                  Pronunciation videos for "{{ v.Word }}"
                </small>
                <button @click="closeYouglishWidget" class="btn btn-sm btn-outline-secondary">
                  <i class="bi bi-x"></i>
                </button>
              </div>
              <div id="youglish-widget"></div>
            </div>

            <!-- Meanings -->
            <div class="mb-3">
              <div class="mb-2">
                <span class="badge bg-primary me-2">EN</span>
                <span class="text-muted">{{ v.MeaningEn }}</span>
              </div>
              <div>
                <span class="badge bg-danger me-2">VI</span>
                <span class="text-muted">{{ v.MeaningVi }}</span>
              </div>
            </div>

            <!-- Examples -->
            <div class="bg-light p-3 rounded-3">
              <div class="mb-2">
                <small class="text-muted">
                  <i class="bi bi-quote"></i> {{ v.ExampleEn }}
                </small>
              </div>
              <div>
                <small class="text-muted">
                  <i class="bi bi-quote"></i> {{ v.ExampleVi }}
                </small>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Action Button -->
    <div class="text-center mt-5">
      <button @click="goQuiz" class="btn btn-primary btn-lg px-5 py-3 shadow-sm">
        <i class="bi bi-arrow-right-circle me-2"></i>
        <strong>Go to Quiz</strong>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

interface Vocab {
  ID: number
  Word: string
  MeaningEn: string
  MeaningVi: string
  ExampleEn: string
  ExampleVi: string
}

const router = useRouter()
const topic = ref('Environment')
const vocabs = ref<Vocab[]>([])
const selectedWords = ref<number[]>([])
const activeVideoId = ref<number | null>(null)
const currentWord = ref<string>('')
const showYouglishWidget = ref<boolean>(false)
const API = (import.meta as any).env.VITE_API_BASE_URL || 'http://localhost:5173'

onMounted(async () => {
  // Load YouGlish script
  loadYouglishScript()
  
  try {
    const { data } = await axios.get(`${API}/api/vocabs?topic=${encodeURIComponent(topic.value)}`)
    vocabs.value = data
  } catch (error) {
    console.error('Error loading vocabulary:', error)
    vocabs.value = []
  }
})

function loadYouglishScript(): void {
  // Check if YouGlish script is already loaded
  if (document.getElementById('youglish-script')) {
    return
  }
  
  const script = document.createElement('script')
  script.id = 'youglish-script'
  script.src = 'https://youglish.com/public/emb/widget.js'
  script.charset = 'utf-8'
  script.async = true
  script.onload = () => {
    console.log('YouGlish script loaded successfully')
  }
  script.onerror = () => {
    console.error('Failed to load YouGlish script')
  }
  document.head.appendChild(script)
}

async function select(id: number) {
  try {
    await axios.post(`${API}/api/vocabs/select`, {
      user_id: 1,
      vocab_ids: [id]
    })
    selectedWords.value.push(id)
    
    // Show success message
    const toast = document.createElement('div')
    toast.className = 'position-fixed top-0 end-0 p-3'
    toast.style.zIndex = '9999'
    toast.innerHTML = `
      <div class="toast show" role="alert">
        <div class="toast-header bg-success text-white">
          <i class="bi bi-check-circle me-2"></i>
          <strong class="me-auto">Success!</strong>
          <button type="button" class="btn-close btn-close-white" data-bs-dismiss="toast"></button>
        </div>
        <div class="toast-body">
          Word selected for study!
        </div>
      </div>
    `
    document.body.appendChild(toast)
    setTimeout(() => toast.remove(), 3000)
  } catch (error) {
    console.error('Error selecting word:', error)
    alert('Failed to select word. Please try again.')
  }
}

function loadYouglishWidget(word: string): void {
  // Set current word and show widget
  currentWord.value = word
  showYouglishWidget.value = true
  
  // Wait for DOM update and YouGlish API to be ready
  setTimeout(() => {
    if ((window as any).YG && (window as any).YG.Widget) {
      try {
        // Clear previous widget if exists
        const widgetContainer = document.getElementById('youglish-widget')
        if (widgetContainer) {
          widgetContainer.innerHTML = ''
          
          // Create new widget instance with configuration
          const widget = new (window as any).YG.Widget("youglish-widget", {
            width: 640,
            components: 9, // search box, caption, control, etc.
            autoStart: 1,  // auto play video
            events: {
              onFetchDone: (e: any) => console.log("YouGlish fetch done", e),
              onVideoChange: (e: any) => console.log("YouGlish video changed", e),
            }
          })
          
          // Fetch videos for the word
          widget.fetch(word, "english")
          console.log('YouGlish widget loaded for word:', word)
        }
      } catch (error) {
        console.error('Failed to load YouGlish widget:', error)
        // Fallback to opening in new tab if widget fails
        window.open(`https://youglish.com/search/${word}/us`, '_blank', 'noopener,noreferrer')
        showYouglishWidget.value = false
      }
    } else {
      console.error('YouGlish API not available')
      // Fallback if YouGlish library not loaded
      window.open(`https://youglish.com/search/${word}/us`, '_blank', 'noopener,noreferrer')
      showYouglishWidget.value = false
    }
  }, 500) // Small delay to ensure DOM is ready
}

function closeYouglishWidget(): void {
  showYouglishWidget.value = false
  currentWord.value = ''
  
  // Clear the widget container
  const widgetContainer = document.getElementById('youglish-widget')
  if (widgetContainer) {
    widgetContainer.innerHTML = ''
  }
}


function toggleYouglishVideo(vocabId: number) {
  if (activeVideoId.value === vocabId) {
    activeVideoId.value = null
  } else {
    activeVideoId.value = vocabId
  }
}

function goQuiz() {
  router.push('/quiz')
}
</script>

<style scoped>
.vocabulary-card {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.vocabulary-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15) !important;
}

.card-title {
  font-size: 1.5rem;
}

.btn-lg {
  transition: all 0.3s ease;
}

.btn-lg:hover {
  transform: translateY(-2px);
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15) !important;
}
/* YouGlish video button styling */
.btn-outline-info {
  transition: all 0.3s ease;
}

.btn-outline-info:hover {
  background-color: #17a2b8;
  border-color: #17a2b8;
  transform: scale(1.05);
}

.btn-outline-info.active {
  background-color: #17a2b8;
  border-color: #17a2b8;
  color: white;
}

/* YouGlish Widget Styling */
.youglish-embed {
  margin: 0 auto;
}

.youglish-widget-container {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* Video embed styling */
.ratio {
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>
