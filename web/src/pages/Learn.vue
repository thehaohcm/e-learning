<template>
  <div>
    <h2>Learn (Topic: {{ topic }})</h2>
    <div v-if="vocabs.length === 0">Loading...</div>
    <ul>
      <li v-for="v in vocabs" :key="v.id">
        <strong>{{ v.word }}</strong> — {{ v.meaning_en }} / {{ v.meaning_vi }}<br />
        <em>{{ v.example_en }}</em> / {{ v.example_vi }}
        <div style="margin-top:6px;">
          <button @click="select(v.id)">Select</button>
        </div>
      </li>
    </ul>
    <button @click="goQuiz" style="margin-top:12px;">Go to quiz</button>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const topic = ref('Environment')
const vocabs = ref<any[]>([])
const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088'

onMounted(async () => {
  const { data } = await axios.get(`${API}/api/vocabs?topic=${encodeURIComponent(topic.value)}`)
  vocabs.value = data
})

async function select(id: number) {
  await axios.post(`${API}/api/vocabs/select`, {
    user_id: 1,
    vocab_ids: [id]
  })
  alert('Selected for study!')
}

function goQuiz() {
  router.push('/quiz')
}
</script>

