import { defineNuxtPlugin } from '#app'
import { useChatStore } from '~/stores/chat'

export default defineNuxtPlugin((nuxtApp) => {
  let ws = null
  let reconnectAttempts = 0
  const maxReconnectAttempts = 5
  const reconnectDelay = 5000
  const listeners = new Set()

  const connectWebSocket = () => {
    if (!process.client) return
    const token = localStorage.getItem('token')
    if (!token) {
      console.log('WebSocket: Токен отсутствует, подключение отменено')
      return
    }
    const config = nuxtApp.$config.public
    try {
      console.log('Attempting WebSocket connection with token:', token)
      ws = new WebSocket(`${config.wsBase}/ws?token=${token}`)
      ws.onopen = () => {
        reconnectAttempts = 0
        console.log('WebSocket connected')
      }
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('WebSocket message received:', data)
          if (data.type === 'message') {
            useChatStore().handleNewMessage(data.data)
            listeners.forEach(callback => callback(data.data))
          }
        } catch (error) {
          console.error('WebSocket message parse error:', error)
        }
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

  const send = (data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(data))
      console.log('WebSocket message sent:', data)
    } else {
      console.error('WebSocket is not connected')
    }
  }

  const closeWebSocket = () => {
    if (ws) {
      console.log('Closing WebSocket')
      ws.close()
      ws = null
      reconnectAttempts = 0
      listeners.clear()
    }
  }

  if (process.client) {
    const token = localStorage.getItem('token')
    if (token) connectWebSocket()
  }

  return {
    provide: {
      websocket: {
        connect: connectWebSocket,
        onMessage: (callback) => {
          listeners.add(callback)
          return () => listeners.delete(callback)
        },
        send,
        close: closeWebSocket
      }
    }
  }
})