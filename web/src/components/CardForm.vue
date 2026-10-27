<template>
  <form class="row g-2" @submit.prevent="submit">
    <div class="col-12">
      <input v-model="front" class="form-control" required :placeholder="frontLabel" />
    </div>

    <template v-if="effKind === 'vocab' || effKind === 'phrasal'">
      <div class="col-md-6"><input v-model="meaningVi" class="form-control" placeholder="Nghĩa tiếng Việt" required /></div>
      <div class="col-md-6"><input v-model="meaningEn" class="form-control" placeholder="English meaning" /></div>
      <div class="col-md-6"><input v-model="exampleEn" class="form-control" placeholder="Example sentence (EN)" /></div>
      <div class="col-md-6"><input v-model="exampleVi" class="form-control" placeholder="Câu ví dụ (VI)" /></div>
      <div v-if="effKind === 'vocab'" class="col-md-6"><input v-model="ipa" class="form-control" placeholder="IPA, e.g. /ˈwɜːd/" /></div>
      <div v-if="effKind === 'vocab'" class="col-md-6"><input v-model="pos" class="form-control" placeholder="Part of speech" /></div>
      <div v-if="effKind === 'phrasal'" class="col-12 form-check ms-1">
        <input id="sep" v-model="separable" type="checkbox" class="form-check-input" />
        <label for="sep" class="form-check-label">Separable</label>
      </div>
    </template>

    <template v-else-if="effKind === 'concept'">
      <div class="col-12"><input v-model="back" class="form-control" placeholder="Definition / formula" required /></div>
      <div class="col-md-6"><input v-model="noteVi" class="form-control" placeholder="Giải thích tiếng Việt" /></div>
      <div class="col-md-6"><input v-model="example" class="form-control" placeholder="Example" /></div>
    </template>

    <template v-else-if="effKind === 'mcq'">
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
      <button class="btn btn-primary btn-sm" :disabled="busy">{{ editing ? 'Save changes' : 'Save card' }}</button>
      <button v-if="editing" type="button" class="btn btn-link btn-sm" @click="$emit('cancel')">Cancel</button>
      <small v-if="error" class="text-danger">{{ error }}</small>
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, errorMessage } from '../api'
import type { Card, CardKind } from '../types'

// Add a new card (no `card` prop) or edit an existing one (`card` given).
const props = defineProps<{ kind: CardKind; deckId: number; card?: Card | null }>()
const emit = defineEmits<{ created: []; saved: []; cancel: [] }>()

const editing = computed(() => !!props.card)
// when editing, the card's own kind wins over the deck default
const effKind = computed<CardKind>(() => props.card?.kind ?? props.kind)

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
}[effKind.value]))

/** Fill the fields from an existing card so the user edits in place. */
function prefill(c: Card) {
  front.value = c.front ?? ''
  const p = c.payload ?? {}
  const s = p.senses?.[0] ?? {}
  meaningVi.value = s.meaning_vi ?? ''
  meaningEn.value = s.meaning_en ?? ''
  exampleEn.value = s.example_en ?? ''
  exampleVi.value = s.example_vi ?? ''
  ipa.value = p.ipa ?? ''
  pos.value = p.pos ?? ''
  separable.value = !!p.separable
  back.value = p.back ?? ''
  noteVi.value = p.note_vi ?? ''
  example.value = p.example ?? ''
  const ch = p.choices ?? []
  choices.value = [ch[0] ?? '', ch[1] ?? '', ch[2] ?? '', ch[3] ?? '']
  answer.value = typeof p.answer === 'number' ? p.answer : 0
  explanation.value = p.explanation ?? ''
}

function reset() {
  front.value = meaningVi.value = meaningEn.value = exampleEn.value = exampleVi.value = ''
  ipa.value = pos.value = back.value = noteVi.value = example.value = explanation.value = ''
  choices.value = ['', '', '', '']
  answer.value = 0
  separable.value = false
}

function payload() {
  const sense = { level: 1, meaning_vi: meaningVi.value, meaning_en: meaningEn.value, example_en: exampleEn.value, example_vi: exampleVi.value }
  switch (effKind.value) {
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
    if (props.card) {
      await api.put(`/cards/${props.card.id}`, { front: front.value, kind: effKind.value, payload: payload() })
      emit('saved')
    } else {
      await api.post('/cards', { deck_id: props.deckId, kind: effKind.value, front: front.value, payload: payload() })
      reset()
      emit('created')
    }
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

onMounted(() => { if (props.card) prefill(props.card) })
</script>
