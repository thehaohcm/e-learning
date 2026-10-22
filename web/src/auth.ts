import { reactive } from 'vue'

export interface User {
  id: number
  email: string
  display_name: string
  role: 'admin' | 'user'
  timezone: string
}

const TOKEN_KEY = 'elearn.token'
const USER_KEY = 'elearn.user'

function read<T>(key: string): T | null {
  try {
    const v = localStorage.getItem(key)
    return v ? (JSON.parse(v) as T) : null
  } catch {
    return null
  }
}

export const auth = reactive({
  token: read<string>(TOKEN_KEY),
  user: read<User>(USER_KEY),
  get isAdmin() {
    return this.user?.role === 'admin'
  },
})

export function setSession(token: string, user: User) {
  auth.token = token
  auth.user = user
  try {
    localStorage.setItem(TOKEN_KEY, JSON.stringify(token))
    localStorage.setItem(USER_KEY, JSON.stringify(user))
  } catch {
    /* private mode: keep the session in memory only */
  }
}

export function logout() {
  auth.token = null
  auth.user = null
  try {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  } catch {
    /* ignore */
  }
}
