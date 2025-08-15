import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: null,
    role: '',
    userId: null,
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
  },
  actions: {
    initialize() {
      if (process.client) {
        this.token = localStorage.getItem('token') || null
        this.role = localStorage.getItem('role') || ''
        this.userId = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
      }
    },
    setUser(token, role, userId) {
      this.token = token
      this.role = role
      this.userId = userId
      if (process.client) {
        localStorage.setItem('token', token)
        localStorage.setItem('role', role)
        localStorage.setItem('userId', userId)
      }
    },
    clearUser() {
      this.token = null
      this.role = ''
      this.userId = null
      if (process.client) {
        localStorage.removeItem('token')
        localStorage.removeItem('role')
        localStorage.removeItem('userId')
      }
    },
  },
})