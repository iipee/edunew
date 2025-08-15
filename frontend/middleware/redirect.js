import { defineNuxtRouteMiddleware, navigateTo } from 'nuxt/app'

export default defineNuxtRouteMiddleware((to) => {
  if (to.path === '/courses') {
    return navigateTo('/')
  }
})