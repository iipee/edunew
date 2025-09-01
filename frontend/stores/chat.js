import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useRuntimeConfig } from 'nuxt/app'

export const useChatStore = defineStore('chat', {
  state: () => ({
    dialogs: [],
    messages: {}, // Храним сообщения как { [receiverId]: Message[] }
    loading: false,
    error: null,
    websocket: null,
    subscriptions: new Set() // Для отслеживания активных receiverId
  }),
  getters: {
    getUnreadCount: (state) => state.dialogs.reduce((sum, dialog) => sum + (dialog.unread_count || 0), 0)
  },
  actions: {
    connectWebSocket() {
      if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
        return
      }
      const config = useRuntimeConfig()
      const token = localStorage.getItem('token')
      if (!token) {
        this.error = 'Токен не найден'
        return
      }
      try {
        this.websocket = new WebSocket(`${config.public.wsBase}/ws?token=${token}`)
        this.websocket.onopen = () => {
          console.log('WebSocket подключен')
        }
        this.websocket.onmessage = (event) => {
          try {
            const msg = JSON.parse(event.data)
            if (msg.type === 'message') {
              this.handleNewMessage(msg.data)
            } else if (msg.type === 'avatar_updated') {
              this.handleAvatarUpdate(msg.data)
            } else if (msg.type === 'notification') {
              console.log('Уведомление:', msg.data)
            }
          } catch (error) {
            console.error('Ошибка парсинга WebSocket:', error)
            this.error = 'Ошибка обработки сообщения WebSocket'
          }
        }
        this.websocket.onerror = (error) => {
          console.error('Ошибка WebSocket:', error)
          this.error = 'Ошибка соединения с WebSocket'
        }
        this.websocket.onclose = () => {
          console.log('WebSocket закрыт, переподключение...')
          this.websocket = null
          setTimeout(() => this.connectWebSocket(), 1000)
        }
      } catch (error) {
        console.error('Ошибка инициализации WebSocket:', error)
        this.error = 'Ошибка инициализации WebSocket'
      }
    },
    closeWebSocket() {
      if (this.websocket) {
        this.websocket.close()
        this.websocket = null
      }
    },
    async fetchDialogs() {
      this.loading = true
      this.error = null
      const config = useRuntimeConfig()
      const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
      try {
        const data = await $fetch(`${config.public.apiBase}/api/chats`, { headers })
        this.dialogs = data ? data.sort((a, b) => {
          const aTime = a.last_message_at ? new Date(a.last_message_at) : new Date(a.created_at || 0)
          const bTime = b.last_message_at ? new Date(b.last_message_at) : new Date(b.created_at || 0)
          return bTime - aTime
        }) : []
        this.unreadCount = this.getUnreadCount
      } catch (error) {
        console.error('Ошибка fetchDialogs:', error)
        this.error = error.message || 'Ошибка загрузки диалогов'
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
        this.messages[receiverId] = data ? data.filter((msg, index, self) => 
          index === self.findIndex((t) => t.id === msg.id)
        ) : []
      } catch (error) {
        console.error('Ошибка fetchMessages:', error)
        this.error = error.message || 'Ошибка загрузки сообщений'
        throw error
      } finally {
        this.loading = false
      }
    },
    async sendMessage(receiverId, content) {
      const config = useRuntimeConfig()
      const headers = { 
        Authorization: `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
      try {
        const response = await $fetch(`${config.public.apiBase}/api/messages`, {
          method: 'POST',
          headers,
          body: JSON.stringify({ receiver_id: receiverId, content })
        })
        this.addMessage(receiverId, response)
        await this.fetchDialogs()
      } catch (error) {
        console.error('Ошибка sendMessage:', error)
        this.error = error.message || 'Ошибка отправки сообщения'
        throw error
      }
    },
    async markRead(receiverId) {
      const config = useRuntimeConfig()
      const headers = { 
        Authorization: `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
      try {
        await $fetch(`${config.public.apiBase}/api/messages/read`, { 
          method: 'PUT', 
          headers, 
          body: JSON.stringify({ receiver_id: receiverId })
        })
        const dialog = this.dialogs.find(d => d.user_id === receiverId)
        if (dialog) dialog.unread_count = 0
        this.unreadCount = this.getUnreadCount
      } catch (error) {
        console.error('Ошибка markRead:', error)
        this.error = error.message || 'Ошибка отметки прочитанных'
      }
    },
    addMessage(receiverId, message) {
      if (!this.messages[receiverId]) this.messages[receiverId] = []
      if (!this.messages[receiverId].some(m => m.id === message.id)) {
        this.messages[receiverId].push(message)
      }
      const dialog = this.dialogs.find(d => d.user_id === (message.sender_id === parseInt(localStorage.getItem('userId')) ? message.receiver_id : message.sender_id))
      if (dialog) {
        dialog.last_message = message.content
        dialog.last_message_at = message.created_at
      }
    },
    handleNewMessage(msg) {
      console.log('Новое сообщение:', msg)
      const currentUserId = parseInt(localStorage.getItem('userId'))
      const dialogId = msg.sender_id === currentUserId ? msg.receiver_id : msg.sender_id
      if (this.subscriptions.has(dialogId)) {
        if (!this.messages[dialogId]) this.messages[dialogId] = []
        if (!this.messages[dialogId].some(m => m.id === msg.id)) {
          this.messages[dialogId].push(msg)
        }
      }
      const dialog = this.dialogs.find(d => d.user_id === dialogId)
      if (dialog) {
        dialog.last_message = msg.content
        dialog.last_message_at = msg.created_at
        if (msg.sender_id !== currentUserId && !this.subscriptions.has(dialogId)) {
          dialog.unread_count = (dialog.unread_count || 0) + 1
          this.unreadCount = this.getUnreadCount
        }
      } else {
        this.fetchDialogs()
      }
    },
    handleAvatarUpdate(data) {
      console.log('Аватар обновлён:', data)
      const dialog = this.dialogs.find(d => d.user_id === data.user_id)
      if (dialog) {
        dialog.avatar_url = data.avatar_url
      }
      if (localStorage.getItem('profile_user_id') === String(data.user_id)) {
        localStorage.setItem('profile_avatar_url', data.avatar_url)
      }
    },
    subscribeToMessages(receiverId) {
      this.subscriptions.add(receiverId)
      if (!this.messages[receiverId]) this.messages[receiverId] = []
      this.fetchMessages(receiverId)
    },
    unsubscribeFromMessages(receiverId) {
      this.subscriptions.delete(receiverId)
      delete this.messages[receiverId]
    }
  }
})