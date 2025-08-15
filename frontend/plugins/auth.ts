import { defineNuxtPlugin } from 'nuxt/app'

export default defineNuxtPlugin((nuxtApp: any) => {
  nuxtApp.provide('getToken', () => {
    if (process.client) {
      return localStorage.getItem('token')
    }
    return null
  })
  nuxtApp.provide('setToken', (token: string, role: string) => {
    if (process.client) {
      localStorage.setItem('token', token)
      localStorage.setItem('role', role)
    }
  })
  nuxtApp.provide('clearToken', () => {
    if (process.client) {
      localStorage.removeItem('token')
      localStorage.removeItem('role')
    }
  })
})