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
        @keyup.enter="sendMessage"
        aria-label="Поле ввода сообщения"
        class="message-input flex-grow-1"
        variant="outlined"
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
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useNuxtApp } from 'nuxt/app'
import { useChatStore } from '~/stores/chat'

const props = defineProps({
  receiverId: {
    type: Number,
    required: true
  }
})
const emit = defineEmits(['update-dialogs'])

const { $emitter } = useNuxtApp()
const chatStore = useChatStore()
const router = useRouter()
const messagesList = ref(null)
const newMessage = ref('')
const snackbar = ref(false)
const snackbarText = ref('')
const messages = ref([])
const userId = ref(parseInt(localStorage.getItem('userId')))
const receiverFullName = ref('')

onMounted(() => {
  console.log('ChatWindow mounted, receiverId:', props.receiverId, 'input field should be visible')
  loadMessages()
  $emitter.on('message', handleNewMessage)
  $emitter.on('chat:started', handleChatStarted)
  window.addEventListener('focus', loadMessages)
})

watch(() => chatStore.messages, () => {
  messages.value = chatStore.messages
  nextTick(() => scrollToBottom())
}, { deep: true })

onUnmounted(() => {
  $emitter.off('message', handleNewMessage)
  $emitter.off('chat:started', handleChatStarted)
  window.removeEventListener('focus', loadMessages)
})

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
  console.log('Send button clicked, message:', newMessage.value)
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
  min-height: 40px;
  width: 100%;
  min-width: 200px; /* Добавлено для видимости */
}
</style>