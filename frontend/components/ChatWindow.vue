<template>
  <v-card height="calc(100vh - 128px)" class="d-flex flex-column" elevation="2" style="border-radius: 12px; background: linear-gradient(180deg, #f9fafb 0%, #ffffff 100%);" aria-label="Окно чата">
    <v-card-title class="justify-space-between align-center py-3 px-4" style="background-color: #28A745; color: white; border-radius: 12px 12px 0 0;">
      <h2 class="text-h6" style="font-weight: 500;" aria-label="Чат с пользователем">{{ receiverFullName || 'Пользователь' }}</h2>
      <v-btn icon @click="closeChat" aria-label="Закрыть чат" color="white">
        <v-icon>mdi-close</v-icon>
      </v-btn>
    </v-card-title>
    <v-divider />
    <v-card-text class="flex-grow-1 overflow-y-auto messages-container" ref="messagesList" style="padding: 16px; background-color: #f5f7fa;">
      <div v-if="!messages.length" class="text-center py-8" aria-label="Нет сообщений">
        <v-icon size="48" color="grey-lighten-1">mdi-message-off</v-icon>
        <p class="mt-2" style="color: #6b7280;">Начните чат с {{ receiverFullName || 'пользователем' }}!</p>
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
    <v-card-actions class="pa-4" style="background-color: #ffffff; border-top: 1px solid #e5e7eb;">
      <v-text-field
        v-model="newMessage"
        label="Введите сообщение"
        prepend-inner-icon="mdi-message"
        clearable
        @keyup.enter="sendMessage"
        aria-label="Поле ввода сообщения"
        class="message-input flex-grow-1"
        variant="outlined"
        style="border-radius: 8px;"
        hide-details
      />
      <v-btn 
        color="#28A745" 
        @click="sendMessage" 
        v-tooltip="'Отправить'" 
        aria-label="Отправить сообщение" 
        class="ml-2" 
        elevation="0"
        style="border-radius: 8px; min-width: 48px;"
      >
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
import { useChatStore } from '~/stores/chat'

const props = defineProps({
  receiverId: { type: Number, required: true }
})
const emit = defineEmits(['update-dialogs'])
const router = useRouter()
const chatStore = useChatStore()
const messagesList = ref(null)
const newMessage = ref('')
const snackbar = ref(false)
const snackbarText = ref('')
const messages = ref([])
const userId = ref(null)
const receiverFullName = ref('')

onMounted(async () => {
  if (process.client) {
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
  }
  await loadMessages()
})

onUnmounted(() => {
  // WebSocket cleanup handled in store
})

watch(() => props.receiverId, async (newId) => {
  if (newId) {
    await loadMessages()
  }
})

const loadMessages = async () => {
  try {
    await chatStore.fetchMessages(props.receiverId)
    messages.value = chatStore.messages
    const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
    const userData = await $fetch(`${useRuntimeConfig().public.apiBase}/api/profile/${props.receiverId}`, { headers })
    receiverFullName.value = userData.profile?.full_name || 'Пользователь'
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
  background-color: #f5f7fa;
  scroll-behavior: smooth;
}
.messages-wrapper {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.message {
  display: flex;
  max-width: 100%;
  transition: all 0.2s ease-in-out;
}
.message-sent {
  justify-content: flex-end;
}
.message-received {
  justify-content: flex-start;
}
.message-bubble {
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 16px;
  position: relative;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.message-bubble:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0,0,0,0.15);
}
.message-bubble.sent {
  background-color: #28A745;
  color: white;
  border-bottom-right-radius: 4px;
}
.message-bubble.received {
  background-color: #ffffff;
  color: #1f2a44;
  border-bottom-left-radius: 4px;
  border: 1px solid #e5e7eb;
}
.message-content {
  word-break: break-word;
  font-size: 14px;
}
.message-time {
  font-size: 11px;
  opacity: 0.6;
  display: block;
  margin-top: 6px;
  text-align: right;
}
.message-input {
  background-color: #ffffff;
  min-height: 40px;
  width: 100%;
}
:deep(.v-text-field .v-field) {
  border-radius: 8px;
  background-color: #ffffff;
}
:deep(.v-text-field .v-field__input) {
  font-size: 14px;
}
:deep(.v-btn) {
  transition: background-color 0.2s ease;
}
:deep(.v-btn:hover) {
  background-color: #2f855a;
}
</style>