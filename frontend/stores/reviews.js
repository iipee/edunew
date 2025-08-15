import { defineStore } from 'pinia'
import { useRuntimeConfig } from 'nuxt/app'

export const useReviewsStore = defineStore('reviews', {
  state: () => ({
    randomReviews: [],
    loading: false,
    error: null
  }),
  getters: {
    getRandomReviews: (state) => state.randomReviews
  },
  actions: {
    async fetchRandomReviews() {
      this.loading = true
      const config = useRuntimeConfig()
      try {
        const { data } = await $fetch(`${config.public.apiBase}/api/reviews/random`)
        this.randomReviews = data || []
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        throw error
      } finally {
        this.loading = false
      }
    }
  }
})