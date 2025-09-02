import { defineNuxtRouteMiddleware, navigateTo } from 'nuxt/app'

export default defineNuxtRouteMiddleware((to) => {
  console.log('Middleware admin.js: проверка роли, route:', to.path)
  const token = localStorage.getItem('token')
  const role = localStorage.getItem('role')
  console.log('Middleware admin.js: role:', role, 'token:', !!token)
  if (!token || role !== 'admin') {
    console.log('Middleware admin.js: редирект на /login')
    return navigateTo('/login')
  }
})