import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const role = ref('')
  const userId = ref(0)
  const isAuthenticated = ref(false)

  function setUser(newToken, newRole, newId) {
    token.value = newToken
    role.value = newRole
    userId.value = newId
    isAuthenticated.value = !!newToken
    if (process.client) {
      localStorage.setItem('token', newToken)
      localStorage.setItem('role', newRole)
      localStorage.setItem('userId', newId.toString())
    }
  }

  function logout() {
    token.value = ''
    role.value = ''
    userId.value = 0
    isAuthenticated.value = false
    if (process.client) {
      localStorage.removeItem('token')
      localStorage.removeItem('role')
      localStorage.removeItem('userId')
    }
  }

  function refresh() {
    if (process.client) {
      const storedToken = localStorage.getItem('token')
      const storedRole = localStorage.getItem('role')
      const storedUserId = localStorage.getItem('userId')
      if (storedToken && storedRole && storedUserId) {
        token.value = storedToken
        role.value = storedRole
        userId.value = parseInt(storedUserId)
        isAuthenticated.value = true
      } else {
        token.value = ''
        role.value = ''
        userId.value = 0
        isAuthenticated.value = false
      }
    }
  }

  return { token, role, userId, isAuthenticated, setUser, logout, refresh }
})