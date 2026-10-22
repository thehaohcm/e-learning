<template>
  <form class="row g-2" @submit.prevent="submit">
    <div class="col-12">
      <input v-model="front" class="form-control" required :placeholder="frontLabel" />
    </div>

    <template v-if="kind === 'vocab' || kind === 'phrasal'">
      <div class="col-md-6"><input v-model="meaningVi" class="form-control" placeholder="Nghĩa tiếng Việt" required /></div>
      <div class="col-md-6"><input v-model="meaningEn" class="form-control" placeholder="English meaning" /></div>
      <div class="col-md-6"><input v-model="exampleEn" class="form-control" placeholder="Example sentence (EN)" /></div>
      <div class="col-md-6"><input v-model="exampleVi" class="form-control" placeholder="Câu ví dụ (VI)" /></div>
      <div v-if="kind === 'vocab'" class="col-md-6"><input v-model="ipa" class="form-control" placeholder="IPA, e.g. /ˈwɜːd/" /></div>
      <div v-if="kind === 'vocab'" class="col-md-6"><input v-model="pos" class="form-control" placeholder="Part of speech" /></div>
      <div v-if="kind === 'phrasal'" class="col-12 form-check ms-1">
        <input id="sep" v-model="separable" type="checkbox" class="form-check-input" />
        <label for="sep" class="form-check-label">Separable</label>
      </div>
    </template>

    <template v-else-if="kind === 'concept'">
      <div class="col-12"><input v-model="back" class="form-control" placeholder="Definition / formula" required /></div>
      <div class="col-md-6"><input v-model="noteVi" class="form-control" placeholder="Giải thích tiếng Việt" /></div>
      <div class="col-md-6"><input v-model="example" class="form-control" placeholder="Example" /></div>
    </template>

    <template v-else-if="kind === 'mcq'">
      <div class="col-md-6" v-for="(_, i) in choices" :key="i">
        <div class="input-group">
          <span class="input-group-text">
            <input v-model.number="answer" type="radio" :value="i" class="form-check-input mt-0" :title="'Correct answer'" />
          </span>
          <input v-model="choices[i]" class="form-control" :placeholder="`Choice ${'ABCD'[i]}`" required />
        </div>
      </div>
      <div class="col-12"><input v-model="explanation" class="form-control" placeholder="Explanation" /></div>
      <small class="text-muted">Select the radio button of the correct answer.</small>
    </template>

    <div class="col-12 d-flex gap-2 align-items-center">
      <button class="btn btn-primary btn-sm" :disabled="busy">Save card</button>
      <small v-if="error" class="text-danger">{{ error }}</small>
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { api, errorMessage } from '../api'
import type { CardKind } from '../types'

const props = defineProps<{ kind: CardKind; deckId: number }>()
const emit = defineEmits<{ created: [] }>()

const front = ref('')
const meaningVi = ref(''), meaningEn = ref(''), exampleEn = ref(''), exampleVi = ref('')
const ipa = ref(''), pos = ref(''), separable = ref(false)
const back = ref(''), noteVi = ref(''), example = ref('')
const choices = ref(['', '', '', ''])
const answer = ref(0)
const explanation = ref('')
const busy = ref(false)
const error = ref('')

const frontLabel = computed(() => ({
  vocab: 'Word', phrasal: 'Phrasal verb, e.g. “give up”', concept: 'Term', mcq: 'Question',
}[props.kind]))

function payload() {
  const sense = { level: 1, meaning_vi: meaningVi.value, meaning_en: meaningEn.value, example_en: exampleEn.value, example_vi: exampleVi.value }
  switch (props.kind) {
    case 'vocab': return { ipa: ipa.value, pos: pos.value, senses: [sense] }
    case 'phrasal': return { separable: separable.value, senses: [sense] }
    case 'concept': return { back: back.value, note_vi: noteVi.value, example: example.value }
    case 'mcq': return { choices: choices.value, answer: answer.value, explanation: explanation.value }
  }
}

async function submit() {
  busy.value = true
  error.value = ''
  try {
    await api.post('/cards', { deck_id: props.deckId, kind: props.kind, front: front.value, payload: payload() })
    front.value = meaningVi.value = meaningEn.value = exampleEn.value = exampleVi.value = ''
    ipa.value = pos.value = back.value = noteVi.value = example.value = explanation.value = ''
    choices.value = ['', '', '', '']
    answer.value = 0
    emit('created')
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}
</script>
