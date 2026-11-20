<template>
  <div>
    <!-- trigger (default hidden panel; this button opens it) -->
    <button v-if="!open" class="btn btn-dark rounded-circle alp-fab shadow" :style="{ bottom: fabBottom }" title="Nghe audio" @click="openPanel">
      <i class="bi bi-headphones"></i>
    </button>

    <!-- backdrop -->
    <div v-if="open" class="alp-backdrop" @click="open = false"></div>

    <!-- right panel -->
    <div class="alp-panel shadow-lg" :class="{ 'alp-open': open }">
      <div class="alp-head d-flex align-items-center gap-2 px-3 py-2 border-bottom">
        <h6 class="fw-bold mb-0 flex-grow-1"><i class="bi bi-headphones"></i> Nghe audio</h6>
        <button class="btn btn-sm btn-outline-secondary" @click="load" :disabled="loading" title="Tải lại"><i class="bi bi-arrow-clockwise"></i></button>
        <button class="btn btn-sm btn-outline-secondary" @click="open = false" title="Đóng"><i class="bi bi-x-lg"></i></button>
      </div>

      <div class="alp-controls px-3 py-2 border-bottom d-flex flex-wrap align-items-center gap-2">
        <button class="btn btn-sm btn-primary" :disabled="!selectedCount || trackLoading" @click="playSelected">
          <span v-if="trackLoading" class="spinner-border spinner-border-sm me-1"></span>
          <i v-else class="bi bi-play-fill"></i> {{ trackLoading ? 'Đang tải…' : 'Phát đã chọn (' + selectedCount + ')' }}
        </button>
        <button class="btn btn-sm btn-outline-secondary" @click="selectAll(!allSelected)">
          {{ allSelected ? 'Bỏ chọn hết' : 'Chọn hết' }}
        </button>
      </div>

      <div v-if="subjects.length > 1" class="alp-controls px-3 py-2 border-bottom d-flex flex-wrap align-items-center gap-1">
        <button class="btn btn-sm" :class="subjectFilter === '' ? 'btn-primary' : 'btn-outline-secondary'" @click="setFilter('')">Tất cả</button>
        <button v-for="s in subjects" :key="s.slug" class="btn btn-sm" :class="subjectFilter === s.slug ? 'btn-primary' : 'btn-outline-secondary'" @click="setFilter(s.slug)">{{ s.name }}</button>
      </div>

      <div class="alp-body">
        <div v-if="loading" class="text-center py-4"><div class="spinner-border text-primary"></div></div>
        <div v-else-if="err" class="alert alert-danger m-3 py-2 small">{{ err }}</div>
        <div v-else-if="!groups.length" class="text-muted small m-3">
          Chưa có audio nào. Tải audio lên trong mục Cheatsheet hoặc sau khi làm Daily 20 / Daily Lab, rồi mở lại đây.
        </div>

        <template v-else>
          <template v-for="cat in categories" :key="cat.kind">
            <div v-if="cat.groups.length" class="alp-cat px-3 pt-3 pb-1 text-uppercase small fw-bold text-muted">{{ cat.title }}</div>
            <div v-for="(g, gi) in cat.groups" :key="g.key" class="alp-group">
              <div class="d-flex align-items-center gap-2 px-3 py-2">
                <input type="checkbox" class="form-check-input mt-0" :checked="groupAllSel(g)" :indeterminate.prop="groupSomeSel(g) && !groupAllSel(g)" @change="toggleGroup(g)" />
                <button class="btn btn-link btn-sm p-0 text-start flex-grow-1 text-decoration-none alp-glabel" @click="toggleExpand(g.key)">
                  <i class="bi" :class="expanded.has(g.key) ? 'bi-chevron-down' : 'bi-chevron-right'"></i>
                  <span class="fw-semibold ms-1">{{ g.label }}</span>
                  <span class="text-muted ms-1">· {{ g.subject_name }} · {{ g.audios.length }}</span>
                </button>
                <div class="btn-group btn-group-sm alp-move">
                  <button class="btn btn-outline-secondary" :disabled="gi === 0" title="Chuyển mục lên" @click="moveGroup(cat.kind, gi, -1)"><i class="bi bi-chevron-up"></i></button>
                  <button class="btn btn-outline-secondary" :disabled="gi === cat.groups.length - 1" title="Chuyển mục xuống" @click="moveGroup(cat.kind, gi, 1)"><i class="bi bi-chevron-down"></i></button>
                </div>
              </div>
              <div v-if="expanded.has(g.key)" class="alp-tracks">
                <div v-for="(a, ai) in g.audios" :key="a.kind + '-' + a.id" class="d-flex align-items-center gap-2 px-3 py-1 alp-track"
                  :class="{ 'alp-playing': isPlaying(a) }">
                  <input type="checkbox" class="form-check-input mt-0" :checked="selected.has(keyOf(a))" @change="toggleSel(keyOf(a))" />
                  <button class="btn btn-link btn-sm p-0 text-start flex-grow-1 text-decoration-none alp-tlabel" :title="a.name" @click="playFrom(a)">
                    <span v-if="isPlaying(a) && trackLoading" class="spinner-border spinner-border-sm text-primary"></span>
                    <i v-else class="bi" :class="isPlaying(a) && playing ? 'bi-volume-up-fill text-primary' : 'bi-music-note-beamed'"></i>
                    <span class="ms-1">{{ a.name }}</span>
                  </button>
                  <div class="btn-group btn-group-sm alp-move">
                    <button class="btn btn-outline-secondary" :disabled="ai === 0" title="Lên" @click="moveAudio(g, ai, -1)"><i class="bi bi-chevron-up"></i></button>
                    <button class="btn btn-outline-secondary" :disabled="ai === g.audios.length - 1" title="Xuống" @click="moveAudio(g, ai, 1)"><i class="bi bi-chevron-down"></i></button>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { api, errorMessage } from '../api'
import { useAudioPlayer, type QueueItem } from '../composables/useAudioPlayer'
import { isEnglish } from '../composables/useSubject'

interface ApiAudio { id: number; name: string; kind: 'note' | 'daily' }
interface ApiGroup { key: string; kind: 'cheatsheet' | 'daily' | 'lab'; subject: string; subject_name: string; label: string; audios: ApiAudio[] }

const player = useAudioPlayer()
const { playing, curKey, loading: trackLoading } = player
const fabBottom = computed(() => (isEnglish.value ? '148px' : '84px')) // stack above the notes buttons (bottom-right)

const open = ref(false)
const loading = ref(false)
const err = ref('')
const groups = ref<ApiGroup[]>([])
const subjectFilter = ref('') // '' = all courses, else a subject slug
// distinct courses present, for the filter chips
const subjects = computed(() => {
  const seen = new Map<string, string>()
  for (const g of groups.value) if (!seen.has(g.subject)) seen.set(g.subject, g.subject_name)
  return Array.from(seen, ([slug, name]) => ({ slug, name }))
})
const visibleGroups = computed(() => (subjectFilter.value ? groups.value.filter((g) => g.subject === subjectFilter.value) : groups.value))
function setFilter(slug: string) { subjectFilter.value = slug; selectAll(true) }
const selected = reactive(new Set<string>())
const expanded = reactive(new Set<string>())
let loadedOnce = false

const keyOf = (a: ApiAudio) => `${a.kind}-${a.id}`
const endpointOf = (a: ApiAudio) => (a.kind === 'note' ? `/note-audio/${a.id}` : `/pmp-daily/audio/${a.id}`)

const CAT_TITLES: Record<string, string> = { cheatsheet: 'Cheatsheet', daily: 'Daily 20', lab: 'Daily Lab' }
const categories = computed(() =>
  (['cheatsheet', 'daily', 'lab'] as const).map((kind) => ({
    kind, title: CAT_TITLES[kind], groups: visibleGroups.value.filter((g) => g.kind === kind),
  })),
)

const allAudios = computed(() => visibleGroups.value.flatMap((g) => g.audios))
const selectedCount = computed(() => allAudios.value.filter((a) => selected.has(keyOf(a))).length)
const allSelected = computed(() => allAudios.value.length > 0 && selectedCount.value === allAudios.value.length)

function isPlaying(a: ApiAudio) { return curKey.value === keyOf(a) }
function toggleSel(k: string) { if (selected.has(k)) selected.delete(k); else selected.add(k) }
function toggleExpand(k: string) { if (expanded.has(k)) expanded.delete(k); else expanded.add(k) }
function groupAllSel(g: ApiGroup) { return g.audios.length > 0 && g.audios.every((a) => selected.has(keyOf(a))) }
function groupSomeSel(g: ApiGroup) { return g.audios.some((a) => selected.has(keyOf(a))) }
function toggleGroup(g: ApiGroup) {
  const on = !groupAllSel(g)
  for (const a of g.audios) { if (on) selected.add(keyOf(a)); else selected.delete(keyOf(a)) }
}
function selectAll(on: boolean) {
  selected.clear()
  if (on) for (const a of allAudios.value) selected.add(keyOf(a))
}

// ordered QueueItems for the current selection, across all groups (display order)
function selectedQueue(): QueueItem[] {
  const out: QueueItem[] = []
  for (const g of visibleGroups.value) for (const a of g.audios) {
    if (selected.has(keyOf(a))) out.push({ key: keyOf(a), id: a.id, name: a.name, endpoint: endpointOf(a) })
  }
  return out
}
function playSelected() {
  const q = selectedQueue()
  if (q.length) player.playQueue(q, 0)
}
// click a track: if it's in the selection, start the selected queue at it;
// otherwise play just that track.
function playFrom(a: ApiAudio) {
  const k = keyOf(a)
  if (selected.has(k)) {
    const q = selectedQueue()
    const idx = q.findIndex((t) => t.key === k)
    player.playQueue(q, Math.max(0, idx))
  } else {
    player.playQueue([{ key: k, id: a.id, name: a.name, endpoint: endpointOf(a) }], 0)
  }
}

function moveGroup(kind: string, gi: number, dir: number) {
  const list = visibleGroups.value.filter((x) => x.kind === kind)
  const target = gi + dir
  if (target < 0 || target >= list.length) return
  const flat = groups.value.slice()
  const i = flat.indexOf(list[gi]); const j = flat.indexOf(list[target])
  ;[flat[i], flat[j]] = [flat[j], flat[i]]
  groups.value = flat
}
function moveAudio(g: ApiGroup, ai: number, dir: number) {
  const target = ai + dir
  if (target < 0 || target >= g.audios.length) return
  const arr = g.audios.slice()
  ;[arr[ai], arr[target]] = [arr[target], arr[ai]]
  g.audios = arr
  groups.value = groups.value.slice()
}

async function load() {
  loading.value = true; err.value = ''
  try {
    const { data } = await api.get('/audio-library')
    groups.value = (data.groups ?? []) as ApiGroup[]
    selectAll(true)
    expanded.clear()
    for (const g of groups.value) expanded.add(g.key) // default expanded
  } catch (e) { err.value = errorMessage(e) } finally { loading.value = false; loadedOnce = true }
}
function openPanel() { open.value = true; if (!loadedOnce) load() }
</script>

<style scoped>
.alp-fab {
  position: fixed; right: 20px; bottom: 84px; z-index: 1050; /* stacked above the notes buttons (bottom-right) */
  width: 3rem; height: 3rem; display: flex; align-items: center; justify-content: center; font-size: 1.1rem;
  transition: bottom .2s ease;
}
@media (max-width: 576px) { .alp-fab { right: 14px; } }
.alp-move .btn { padding: 0.05rem 0.3rem; line-height: 1; }
.alp-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,.35); z-index: 1058; }
.alp-panel {
  position: fixed; top: 0; right: 0; height: 100%; z-index: 1059;
  width: min(380px, 100vw); background: #fff; display: flex; flex-direction: column;
  transform: translateX(105%); transition: transform .22s ease; will-change: transform;
}
.alp-panel.alp-open { transform: translateX(0); }
.alp-body { overflow-y: auto; flex: 1; padding-bottom: 6rem; }
.alp-cat { letter-spacing: .04em; }
.alp-group { border-bottom: 1px solid #f1f5f9; }
.alp-glabel, .alp-tlabel { color: #1e293b; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.alp-track.alp-playing { background: #eff6ff; }
.alp-tracks { background: #f8fafc; }
</style>
