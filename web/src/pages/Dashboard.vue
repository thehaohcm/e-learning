<template>
  <div>
    <h2>Dashboard</h2>
    <div>Learned: {{ summary.learned }}</div>
    <div>Weak: {{ summary.weak }}</div>
    <div>Due today: {{ summary.due }}</div>
    <div style="margin-top:12px;">
      <router-link to="/learn">Start learning</router-link> |
      <router-link to="/quiz">Take quiz</router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted } from 'vue'
const summary = ref<any>({ learned: 0, weak: 0, due: 0 })
const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088'

onMounted(async () => {
  const { data } = await axios.get(`${API}/api/dashboard`)
  summary.value = data
})
</script>

