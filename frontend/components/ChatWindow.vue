<template>
    <v-card height="calc(100vh - 128px)" class="d-flex flex-column" elevation="0" style="border-radius: 8px;" aria-label="Окно чата">
      <v-card-title class="justify-space-between align-center py-2">
        <h2 class="text-h6" aria-label="Чат с пользователем">{{ receiverFullName || 'Пользователь' }}</h2>
        <v-btn icon @click="closeChat" aria-label="Закрыть чат">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>
      <v-divider />
      <v-card-text class="flex-grow-1 overflow-y-auto messages-container" ref="messagesList">
        <div v-if="!messages.length" class="text-center py-4" aria-label="Нет сообщений">
          <v-icon size="48" color="grey-lighten-1">mdi-message-off</v-icon>
          <p class="mt-2">Начните чат с {{ receiverFullName || 'пользователем' }}!</p>
        </div>
        <div v-else class="messages-wrapper">
          <div 
            v-for="(message, index) in messages" 
            :key="index" 
            :class="['message', message.sender_id === userId ? 'message-sent' : 'message-received']"
            aria-label="Сообщение"
          >
            <div :class="['message-bubble', message.sender_id === userId ? 'sent' : 'received']">
              <span class="message-content">{{ message.content }}</span>
              <span class="message-time">{{ formatDate(message.created_at) }}</span>
            </div>
          </div>
        </div>
      </v-card-text>
      <v-card-actions class="pa-4">
        <v-text-field
          v-model="newMessage"
          label="Введите сообщение"
          prepend-inner-icon="mdi-message"
          clearable
          @input="newMessage = $event.target.value || ''"
          @keyup.enter="sendMessage"
          aria-label="Поле ввода сообщения"
          class="message-input"
          outlined
          dense
        />
        <v-btn color="primary" @click="sendMessage" v-tooltip="'Отправить'" aria-label="Отправить сообщение" class="ml-2">
          <v-icon>mdi-send</v-icon>
        </v-btn>
      </v-card-actions>
      <v-snackbar v-model="snackbar" color="error" timeout="3000" aria-label="Уведомление об ошибке">
        {{ snackbarText }}
      </v-snackbar>
    </v-card>
  </template>
  
  <script setup>
  import { ref, onMounted, onUnmounted, nextTick } from 'vue'
  import { useRuntimeConfig } from 'nuxt/app'
  import { useRouter } from 'vue-router'
  import { useChatStore } from '~/stores/chat'
  import { useAuthStore } from '~/stores/auth'
  import { useNuxtApp } from 'nuxt/app'
  
  const { $emitter } = useNuxtApp()
  const props = defineProps({
    receiverId: {
      type: Number,
      required: true
    }
  })
  const emit = defineEmits(['update-dialogs'])
  const config = useRuntimeConfig()
  const router = useRouter()
  const chatStore = useChatStore()
  const authStore = useAuthStore()
  const messages = ref([])
  const newMessage = ref('')
  const receiverFullName = ref('')
  const messagesList = ref(null)
  const userId = ref(authStore.userId)
  const snackbar = ref(false)
  const snackbarText = ref('')
  
  onMounted(async () => {
    if (!authStore.isLoggedIn) {
      router.push('/login')
      return
    }
    await loadReceiverName()
    await loadMessages()
    $emitter.on('message', handleNewMessage)
    $emitter.on('chat:started', handleChatStarted)
  })
  
  onUnmounted(() => {
    chatStore.messages = []
    $emitter.off('message', handleNewMessage)
    $emitter.off('chat:started', handleChatStarted)
  })
  
  const loadReceiverName = async () => {
    try {
      const headers = { Authorization: `Bearer ${authStore.token}` }
      const { profile } = await $fetch(`${config.public.apiBase}/api/profile/${props.receiverId}`, { headers })
      receiverFullName.value = profile.full_name || `User ${props.receiverId}`
    } catch (error) {
      snackbarText.value = 'Ошибка загрузки имени получателя'
      snackbar.value = true
    }
  }
  
  const loadMessages = async () => {
    try {
      await chatStore.fetchMessages(props.receiverId)
      messages.value = chatStore.messages
      await nextTick()
      scrollToBottom()
    } catch (error) {
      snackbarText.value = 'Ошибка загрузки сообщений'
      snackbar.value = true
    }
  }
  
  const sendMessage = async () => {
    if (!newMessage.value.trim()) return
    try {
      await chatStore.sendMessage(props.receiverId, newMessage.value)
      newMessage.value = ''
      messages.value = chatStore.messages
      await nextTick()
      scrollToBottom()
      emit('update-dialogs')
    } catch (error) {
      snackbarText.value = 'Ошибка отправки сообщения'
      snackbar.value = true
    }
  }
  
  const handleNewMessage = (msg) => {
    if (msg.sender_id === props.receiverId || msg.receiver_id === props.receiverId) {
      messages.value = chatStore.messages
      nextTick(() => scrollToBottom())
    }
  }
  
  const handleChatStarted = () => {
    if (props.receiverId) {
      loadMessages()
      emit('update-dialogs')
    }
  }
  
  const scrollToBottom = () => {
    if (messagesList.value) {
      messagesList.value.scrollTop = messagesList.value.scrollHeight
    }
  }
  
  const closeChat = () => {
    router.replace('/chats')
  }
  
  const formatDate = (timestamp) => {
    return new Date(timestamp).toLocaleString('ru-RU', { timeStyle: 'short', dateStyle: 'short' })
  }
  </script>
  
  <style scoped>
  .messages-container {
    padding: 16px;
    background-color: #f5f5f5;
  }
  .messages-wrapper {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .message {
    display: flex;
    max-width: 100%;
  }
  .message-sent {
    justify-content: flex-end;
  }
  .message-received {
    justify-content: flex-start;
  }
  .message-bubble {
    max-width: 70%;
    padding: 8px 12px;
    border-radius: 12px;
    position: relative;
  }
  .message-bubble.sent {
    background-color: #28A745;
    color: white;
    border-bottom-right-radius: 4px;
  }
  .message-bubble.received {
    background-color: #ffffff;
    color: #333;
    border: 1px solid #e0e0e0;
    border-bottom-left-radius: 4px;
  }
  .message-content {
    word-break: break-word;
  }
  .message-time {
    font-size: 12px;
    opacity: 0.7;
    display: block;
    margin-top: 4px;
  }
  .message-input {
    background-color: #ffffff;
  }
  </style>