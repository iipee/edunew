import { defineStore } from 'pinia'
import { useRuntimeConfig } from 'nuxt/app'

export const useCourseStore = defineStore('course', {
  state: () => ({
    course: {},
    reviews: [],
    isPaid: false,
    canReview: false,
    transactionId: null,
    loading: true,
    error: null
  }),
  actions: {
    async loadCourse(id) {
      this.loading = true
      const config = useRuntimeConfig()
      const headers = process.client && localStorage.getItem('token') ? { Authorization: `Bearer ${localStorage.getItem('token')}` } : {}
      try {
        const data = await $fetch(`${config.public.apiBase}/api/courses/${id}`, { headers })
        this.course = data || {}
        await this.loadReviews(id)
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        this.course = {}
        throw error
      } finally {
        this.loading = false
      }
    },
    async loadReviews(id) {
      const config = useRuntimeConfig()
      const headers = process.client && localStorage.getItem('token') ? { Authorization: `Bearer ${localStorage.getItem('token')}` } : {}
      try {
        const data = await $fetch(`${config.public.apiBase}/api/reviews/course/${id}`, { headers })
        this.reviews = data || []
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        throw error
      }
    },
    async checkCanReview(userId) {
      if (!localStorage.getItem('token')) {
        this.isPaid = false
        this.canReview = false
        return
      }
      const config = useRuntimeConfig()
      const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
      try {
        const data = await $fetch(`${config.public.apiBase}/api/enrolled`, { headers })
        this.isPaid = data && Array.isArray(data) ? data.some(e => e.course_id === parseInt(this.course.id)) : false
        this.canReview = this.isPaid
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        this.isPaid = false
        this.canReview = false
      }
    },
    async submitPayment(courseId, userId) {
      const config = useRuntimeConfig()
      const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
      const body = { course_id: parseInt(courseId) }
      try {
        const data = await $fetch(`${config.public.apiBase}/api/payments/create`, { method: 'POST', headers, body })
        return data
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        throw error
      }
    },
    async submitReview(courseId, content, userId) {
      const config = useRuntimeConfig()
      const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
      const body = { course_id: parseInt(courseId), content, author_id: userId }
      try {
        const data = await $fetch(`${config.public.apiBase}/api/reviews`, { method: 'POST', headers, body })
        this.reviews.push(data)
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        throw error
      }
    },
    handleNewMessage(msg) {
      // Placeholder for WebSocket message handling if needed
    }
  }
})