import { defineNuxtPlugin } from 'nuxt/app'
import { useChatStore } from '~/stores/chat'

export default defineNuxtPlugin((nuxtApp) => {
  let ws = null
  let reconnectAttempts = 0
  const maxReconnectAttempts = 5
  const reconnectDelay = 5000
  const chatStore = useChatStore()

  const connectWebSocket = () => {
    if (!process.client) return
    const token = localStorage.getItem('token')
    if (!token) {
      console.log('WebSocket: Токен отсутствует, подключение отменено')
      return
    }
    const config = nuxtApp.$config.public
    try {
      console.log('Attempting WebSocket connection with token:', token) // Дебаг-лог
      ws = new WebSocket(`${config.wsBase}/ws?token=${token}`)
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('WebSocket message received:', data) // Дебаг-лог
          if (data.type === 'message') {
            chatStore.handleNewMessage(data.data)
            nuxtApp.$emitter.emit('message', data.data)
          } else if (data.type === 'notification') {
            nuxtApp.$emitter.emit('notification', data.data)
          } else if (data.type === 'chat:started') {
            nuxtApp.$emitter.emit('chat:started', data.data)
          }
        } catch (error) {
          console.error('WebSocket message parse error:', error)
        }
      }
      ws.onopen = () => {
        reconnectAttempts = 0
        console.log('WebSocket connected')
      }
      ws.onclose = () => {
        console.log('WebSocket closed, attempting reconnect...')
        if (reconnectAttempts < maxReconnectAttempts) {
          setTimeout(() => {
            reconnectAttempts++
            console.log(`Reconnect attempt ${reconnectAttempts}/${maxReconnectAttempts}`)
            connectWebSocket()
          }, reconnectDelay * Math.pow(2, reconnectAttempts))
        }
      }
      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
      }
    } catch (error) {
      console.error('WebSocket connection error:', error)
    }
  }

  const disconnectWebSocket = () => {
    if (ws) {
      console.log('Disconnecting WebSocket') // Дебаг-лог
      ws.close()
      ws = null
      reconnectAttempts = 0
    }
  }

  if (process.client) {
    const token = localStorage.getItem('token')
    if (token) connectWebSocket()
    nuxtApp.$emitter.on('login', connectWebSocket)
    nuxtApp.$emitter.on('logout', disconnectWebSocket)
  }

  return {
    provide: {
      websocket: {
        connect: connectWebSocket,
        disconnect: disconnectWebSocket
      }
    }
  }
})