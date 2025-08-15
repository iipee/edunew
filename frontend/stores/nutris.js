import { defineStore } from 'pinia'
import { useRuntimeConfig } from 'nuxt/app'

export const useNutrisStore = defineStore('nutris', {
  state: () => ({
    recommended: [],
    searchResults: [],
    loading: false,
    error: null
  }),
  getters: {
    getRecommended: (state) => state.recommended,
    getSearchResults: (state) => state.searchResults
  },
  actions: {
    async fetchRecommended() {
      this.loading = true
      const config = useRuntimeConfig()
      try {
        const { data } = await $fetch(`${config.public.apiBase}/api/nutris?limit=4&random=true`)
        console.log('fetchRecommended response:', data)
        this.recommended = data || []
        if (this.recommended.length === 0) {
          console.warn('No nutris returned from /api/nutris')
        }
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        console.error('fetchRecommended error:', error)
        setTimeout(() => this.fetchRecommended(), 3000) // Retry on error
        throw error
      } finally {
        this.loading = false
      }
    },
    async fetchNutris(params = {}) {
      this.loading = true
      const config = useRuntimeConfig()
      try {
        if (params.q) {
          const query = new URLSearchParams(params).toString()
          const headers = process.client && localStorage.getItem('token') ? { Authorization: `Bearer ${localStorage.getItem('token')}` } : {}
          const { data } = await $fetch(`${config.public.apiBase}/api/search?${query}`, { headers })
          console.log('fetchNutris search response:', data)
          const nutris = []
          const seenIds = new Set()
          for (const course of data || []) {
            if (course.teacher && !seenIds.has(course.teacher.id)) {
              nutris.push(course.teacher)
              seenIds.add(course.teacher.id)
            }
          }
          this.searchResults = nutris
        } else {
          const { data } = await $fetch(`${config.public.apiBase}/api/nutris?limit=100`)
          console.log('fetchNutris all nutris response:', data)
          this.searchResults = data || []
        }
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        console.error('fetchNutris error:', error)
        setTimeout(() => this.fetchNutris(params), 3000) // Retry on error
        throw error
      } finally {
        this.loading = false
      }
    }
  }
})