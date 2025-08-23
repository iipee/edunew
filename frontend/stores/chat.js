import { defineStore } from 'pinia'
import { useRuntimeConfig } from 'nuxt/app'

export const useChatStore = defineStore('chat', {
  state: () => ({
    dialogs: [],
    messages: [],
    loading: false,
    error: null
  }),
  getters: {
    unreadCount: (state) => state.dialogs.reduce((sum, dialog) => sum + (dialog.unread_count || 0), 0)
  },
  actions: {
    async fetchDialogs() {
      this.loading = true
      this.error = null
      const config = useRuntimeConfig()
      const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
      try {
        const data = await $fetch(`${config.public.apiBase}/api/chats`, { headers })
        this.dialogs = data ? data.sort((a, b) => new Date(b.last_message_at) - new Date(a.last_message_at)) : []
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        this.dialogs = []
        throw error
      } finally {
        this.loading = false
      }
    },
    async fetchMessages(receiverId) {
      this.loading = true
      this.error = null
      const config = useRuntimeConfig()
      const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
      try {
        const data = await $fetch(`${config.public.apiBase}/api/messages?receiver_id=${receiverId}`, { headers })
        this.messages = data || []
        await this.markRead(receiverId)
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        this.messages = []
        throw error
      } finally {
        this.loading = false
      }
    },
    async sendMessage(receiverId, content) {
      console.log('Sending message to:', receiverId, 'Content:', content) // Дебаг-лог
      const config = useRuntimeConfig()
      const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
      const body = { receiver_id: receiverId, content }
      try {
        const optimisticMsg = {
          id: Date.now(),
          sender_id: parseInt(localStorage.getItem('userId')),
          receiver_id: receiverId,
          content,
          created_at: new Date().toISOString()
        }
        this.messages.push(optimisticMsg)
        const data = await $fetch(`${config.public.apiBase}/api/messages`, { method: 'POST', headers, body })
        const index = this.messages.findIndex(m => m.id === optimisticMsg.id)
        if (index !== -1) this.messages[index] = data
        await this.fetchDialogs()
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
        this.messages.pop()
        throw error
      }
    },
    async markRead(receiverId) {
      const config = useRuntimeConfig()
      const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
      try {
        await $fetch(`${config.public.apiBase}/api/messages/read`, { 
          method: 'PUT', 
          headers, 
          body: { receiver_id: receiverId }
        })
        const dialog = this.dialogs.find(d => d.user_id === receiverId)
        if (dialog) dialog.unread_count = 0
      } catch (error) {
        this.error = error.message || 'Неизвестная ошибка'
      }
    },
    handleNewMessage(msg) {
      console.log('New message received:', msg) // Дебаг-лог
      const currentUserId = parseInt(localStorage.getItem('userId'))
      if (msg.sender_id === currentUserId || msg.receiver_id === currentUserId) {
        if (!this.messages.some(m => m.id === msg.id)) {
          this.messages.push(msg)
        }
        const dialog = this.dialogs.find(d => d.user_id === (msg.sender_id === currentUserId ? msg.receiver_id : msg.sender_id))
        if (dialog) {
          dialog.last_message = msg.content
          dialog.unread_count = msg.sender_id === currentUserId ? 0 : (dialog.unread_count || 0) + 1
        } else {
          this.fetchDialogs()
        }
      }
    }
  }
})