// Single HTTP client for the whole app. Adds the JWT and sends the user to the
// login page when the session expires.
import axios, { AxiosError } from 'axios'
import { auth, logout } from './auth'
import router from './router'

// Empty base URL = same origin. In Docker, nginx proxies /api to the API
// container; in `npm run dev`, vite.config.ts proxies /api to localhost:8000.
export const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

export const api = axios.create({ baseURL: `${API_BASE}/api`, timeout: 200_000 })

api.interceptors.request.use((config) => {
  if (auth.token) config.headers.Authorization = `Bearer ${auth.token}`
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err: AxiosError) => {
    if (err.response?.status === 401 && auth.token) {
      logout()
      router.push({ path: '/login', query: { next: router.currentRoute.value.fullPath } })
    }
    return Promise.reject(err)
  },
)

/** Human readable message from an API error. */
export function errorMessage(e: unknown): string {
  const err = e as AxiosError<{ error?: string }>
  return err.response?.data?.error || err.message || 'Something went wrong'
}
