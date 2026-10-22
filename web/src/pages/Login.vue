<template>
  <div class="row justify-content-center">
    <div class="col-md-6 col-lg-5">
      <div class="card border-0 shadow-sm">
        <div class="card-body p-4">
          <h3 class="fw-bold mb-1">{{ mode === 'login' ? 'Sign in' : 'Create account' }}</h3>
          <p class="text-muted small mb-4">
            {{ mode === 'login'
              ? 'Welcome back. Your study queue is waiting.'
              : 'Accounts from the old app: register with the same email to set a password and keep your progress.' }}
          </p>

          <form @submit.prevent="submit">
            <div v-if="mode === 'register'" class="mb-3">
              <label class="form-label">Name</label>
              <input v-model="displayName" class="form-control" maxlength="100" autocomplete="name" />
            </div>
            <div class="mb-3">
              <label class="form-label">Email</label>
              <input v-model="email" type="email" class="form-control" required autocomplete="email" />
            </div>
            <div class="mb-3">
              <label class="form-label">Password</label>
              <input v-model="password" type="password" class="form-control" required
                :minlength="mode === 'register' ? 8 : undefined"
                :autocomplete="mode === 'login' ? 'current-password' : 'new-password'" />
              <div v-if="mode === 'register'" class="form-text">At least 8 characters.</div>
            </div>
            <div v-if="error" class="alert alert-danger py-2 small">{{ error }}</div>
            <button class="btn btn-primary w-100" :disabled="busy">
              {{ busy ? 'Please wait…' : mode === 'login' ? 'Sign in' : 'Create account' }}
            </button>
          </form>

          <div class="text-center mt-3 small">
            <a href="#" @click.prevent="toggle">
              {{ mode === 'login' ? 'No account yet? Create one' : 'Already have an account? Sign in' }}
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errorMessage } from '../api'
import { setSession } from '../auth'

const route = useRoute()
const router = useRouter()
const mode = ref<'login' | 'register'>('login')
const email = ref('')
const password = ref('')
const displayName = ref('')
const error = ref('')
const busy = ref(false)

function toggle() {
  mode.value = mode.value === 'login' ? 'register' : 'login'
  error.value = ''
}

async function submit() {
  busy.value = true
  error.value = ''
  try {
    const body = mode.value === 'login'
      ? { email: email.value, password: password.value }
      : { email: email.value, password: password.value, display_name: displayName.value }
    const { data } = await api.post(`/auth/${mode.value}`, body)
    setSession(data.token, data.user)
    const next = typeof route.query.next === 'string' && route.query.next.startsWith('/') ? route.query.next : '/'
    router.replace(next)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}
</script>
