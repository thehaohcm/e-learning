<template>
  <div>
    <h2>Quiz</h2>
    <div v-if="questions.length === 0">Loading...</div>
    <div v-for="q in questions" :key="q.id" style="margin-bottom:18px;">
      <div><strong>{{ q.prompt }}</strong></div>
      <ul style="list-style: none; padding: 0;">
        <li v-for="(c, idx) in JSON.parse(q.choices)" :key="idx">
          <button @click="answer(q.id, idx)">{{ c }}</button>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted } from 'vue'

const questions = ref<any[]>([])
const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088'

onMounted(async () => {
  const { data } = await axios.get(`${API}/api/quiz`)
  questions.value = data
})

async function answer(questionId: number, answerIdx: number) {
  const { data } = await axios.post(`${API}/api/quiz/attempt`, {
    user_id: 1,
    question_id: questionId,
    answer_idx: answerIdx
  })
  alert(data.correct ? 'Correct!' : 'Incorrect')
}
</script>

