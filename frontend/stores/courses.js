import { defineStore } from 'pinia'
import { useRuntimeConfig } from 'nuxt/app'

export const useCoursesStore = defineStore('courses', {
  state: () => ({
    courses: [],
    loading: false,
    error: null
  }),
  getters: {
    getCourses: (state) => state.courses.sort((a, b) => a.title.localeCompare(b.title) || a.id - b.id)
  },
  actions: {
    async fetchCourses(params = {}) {
      this.loading = true
      const config = useRuntimeConfig()
      try {
        const query = new URLSearchParams(params).toString()
        const headers = process.client && localStorage.getItem('token') ? { Authorization: `Bearer ${localStorage.getItem('token')}` } : {}
        const data = await $fetch(`${config.public.apiBase}/api/search?${query}`, { headers })
        this.courses = data || []
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        throw error
      } finally {
        this.loading = false
      }
    }
  }
})