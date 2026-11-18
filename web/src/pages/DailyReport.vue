<template>
  <div class="row justify-content-center">
    <div class="col-12 col-lg-10">
      <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
        <div>
          <h2 class="fw-bold mb-0"><i class="bi bi-graph-up-arrow"></i> {{ title }} · Báo cáo</h2>
          <small class="text-muted">Mạnh / yếu theo thành phần bài thi, ngày qua ngày</small>
        </div>
        <router-link class="btn btn-sm btn-outline-secondary" :to="backLink">
          <i class="bi bi-arrow-left"></i> {{ isLab ? 'Daily Lab' : 'Daily 20' }}
        </router-link>
      </div>

      <div v-if="err" class="alert alert-danger py-2 small">{{ err }}</div>
      <div v-if="loading" class="text-center py-5"><div class="spinner-border text-primary"></div></div>

      <template v-else-if="rep && rep.sets_counted > 0">
        <!-- readiness -->
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <div class="d-flex flex-wrap align-items-center gap-3">
              <span class="badge fs-6" :class="readyBadge">{{ rep.readiness.label }}</span>
              <div class="text-muted small">
                {{ rep.sets_counted }} bộ đã nộp · {{ rep.questions_counted }} câu · tổng đúng
                <b :class="pctText(rep.overall.pct)">{{ rep.overall.pct }}%</b>
              </div>
            </div>
            <ul v-if="rep.readiness.reasons.length" class="small text-muted mb-0 mt-2 ps-3">
              <li v-for="(r, i) in rep.readiness.reasons" :key="i">{{ r }}</li>
            </ul>
          </div>
        </div>

        <!-- by component -->
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <h6 class="fw-bold mb-3"><i class="bi bi-bar-chart-line"></i> Theo {{ rep.group_label }}</h6>
            <div v-for="g in rep.by_group" :key="g.key" class="mb-2">
              <div class="small d-flex justify-content-between">
                <span class="text-capitalize">{{ g.key }}</span>
                <span :class="pctText(g.pct)">{{ g.pct }}% <span class="text-muted">({{ g.right }}/{{ g.total }})</span></span>
              </div>
              <div class="progress" style="height: 8px">
                <div class="progress-bar" :class="barClass(g.pct)" :style="{ width: g.pct + '%' }"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- weak topics (MCQ subjects only) -->
        <div v-if="rep.weak_topics && rep.weak_topics.length" class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <h6 class="fw-bold mb-3"><i class="bi bi-exclamation-triangle"></i> Chủ đề cần ôn thêm</h6>
            <div v-for="t in rep.weak_topics" :key="t.key" class="mb-2">
              <div class="small d-flex justify-content-between">
                <span class="text-truncate pe-2" style="max-width: 80%">{{ t.key }}</span>
                <span :class="pctText(t.pct)">{{ t.pct }}% <span class="text-muted">({{ t.right }}/{{ t.total }})</span></span>
              </div>
              <div class="progress" style="height: 6px">
                <div class="progress-bar" :class="barClass(t.pct)" :style="{ width: t.pct + '%' }"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- trend by day -->
        <div v-if="rep.trend.length > 1" class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <h6 class="fw-bold mb-3"><i class="bi bi-activity"></i> Xu hướng điểm theo ngày</h6>
            <div class="d-flex align-items-end gap-1" style="height: 90px">
              <div v-for="(p, i) in rep.trend" :key="i" class="flex-fill d-flex flex-column justify-content-end"
                   :title="`${p.date}: ${p.pct}%`" style="min-width: 6px">
                <div class="rounded-top" :class="barClass(p.pct)" :style="{ height: Math.max(3, p.pct) + '%' }"></div>
              </div>
            </div>
            <div class="d-flex justify-content-between small text-muted mt-1">
              <span>{{ rep.trend[0].date.slice(5) }}</span>
              <span>{{ rep.trend[rep.trend.length - 1].date.slice(5) }}</span>
            </div>
          </div>
        </div>

        <!-- coverage -->
        <div v-if="rep.coverage && rep.coverage.total > 0" class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <div class="small d-flex justify-content-between mb-1">
              <span class="fw-bold"><i class="bi bi-grid-3x3"></i> Độ phủ cheatsheet</span>
              <span>{{ rep.coverage.seen }}/{{ rep.coverage.total }} mục · {{ rep.coverage.pct }}%</span>
            </div>
            <div class="progress" style="height: 8px">
              <div class="progress-bar bg-info" :style="{ width: rep.coverage.pct + '%' }"></div>
            </div>
            <p class="small text-muted mb-0 mt-2">Số mục cheatsheet bạn đã được hỏi ít nhất một lần. Mở rộng cheatsheet để tăng độ phủ và sát đề thật hơn.</p>
          </div>
        </div>

        <p class="small text-muted"><i class="bi bi-info-circle"></i> {{ rep.disclaimer }}</p>
      </template>

      <template v-else-if="rep">
        <div class="card border-0 shadow-sm">
          <div class="card-body text-center py-5">
            <i class="bi bi-inbox fs-1 text-muted"></i>
            <p class="mt-2 mb-1">Chưa có bộ nào được nộp cho {{ title }}.</p>
            <p class="small text-muted mb-3">Làm và nộp vài bộ {{ isLab ? 'Daily Lab' : 'Daily 20' }}, báo cáo sẽ tổng hợp tại đây.</p>
            <router-link class="btn btn-primary" :to="backLink">Bắt đầu</router-link>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import { subjectSlug, shortName } from '../composables/useSubject'
import { currentPrefix } from '../router'

interface Group { key: string; right: number; total: number; pct: number }
interface TrendPoint { date: string; pct: number; sets?: number }
interface Coverage { seen: number; total: number; pct: number }
interface Readiness { level: 'ready' | 'almost' | 'not_ready'; label: string; reasons: string[] }
interface Report {
  subject: string
  group_label: string
  sets_counted: number
  questions_counted: number
  overall: Group
  by_group: Group[]
  weak_topics: Group[] | null
  trend: TrendPoint[]
  coverage: Coverage | null
  readiness: Readiness
  disclaimer: string
}

const rep = ref<Report | null>(null)
const loading = ref(true)
const err = ref('')

const isLab = computed(() => subjectSlug.value === 'pentest') // only pentest is a lab now; AWS is MCQ (Daily 20)
const title = computed(() => shortName(subjectSlug.value))
const backLink = computed(() => `/${currentPrefix()}/${isLab.value ? 'daily-lab' : 'daily'}`)

const readyBadge = computed(() => {
  switch (rep.value?.readiness.level) {
    case 'ready': return 'text-bg-success'
    case 'almost': return 'text-bg-warning'
    default: return 'text-bg-secondary'
  }
})

function barClass(pct: number): string {
  if (pct >= 75) return 'bg-success'
  if (pct >= 50) return 'bg-warning'
  return 'bg-danger'
}
function pctText(pct: number): string {
  if (pct >= 75) return 'text-success fw-semibold'
  if (pct >= 50) return 'text-warning-emphasis fw-semibold'
  return 'text-danger fw-semibold'
}

async function load() {
  loading.value = true
  err.value = ''
  try {
    const { data } = await api.get('/pmp-daily/competency', { params: { subject: subjectSlug.value } })
    rep.value = data
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(subjectSlug, load)
</script>
