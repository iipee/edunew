<template>
  <v-card height="100%" class="d-flex flex-column" elevation="0" style="border-radius: 0; background: #ffffff;" aria-label="Окно чата">
    <v-card-title class="justify-space-between align-center py-2 px-4" style="background-color: #ffffff; color: #1f2a44; position: sticky; top: 0; z-index: 1; min-height: 48px; border-bottom: 1px solid #e5e7eb;">
      <h2 class="text-subtitle-1" style="font-weight: 500; font-size: 16px;" aria-label="Чат с пользователем">{{ receiverFullName }}</h2>
      <v-btn icon @click="closeChat" aria-label="Закрыть чат" color="#1f2a44" size="small">
        <v-icon size="20">mdi-close</v-icon>
      </v-btn>
    </v-card-title>
    <v-card-text class="flex-grow-1 overflow-y-auto messages-container" ref="messagesList" style="padding: 12px; background-color: #f5f7fa;">
      <div v-if="loading" class="text-center py-8" aria-label="Загрузка сообщений">
        <v-progress-circular indeterminate color="#28A745" size="24" aria-label="Индикатор загрузки" />
      </div>
      <div v-else-if="!sortedMessages.length" class="text-center py-8" aria-label="Нет сообщений">
        <v-icon size="48" color="grey-lighten-1">mdi-message-off</v-icon>
        <p class="mt-2" style="color: #6b7280; font-size: 14px;">Начните чат с {{ receiverFullName }}!</p>
      </div>
      <div v-else class="messages-wrapper">
        <div 
          v-for="(message, index) in sortedMessages" 
          :key="message.id || index" 
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
    <v-card-actions class="pa-2" style="background-color: #ffffff; border-top: 1px solid #e5e7eb; position: sticky; bottom: 0; z-index: 1;">
      <v-text-field
        v-model="message"
        placeholder="Введите сообщение..."
        hide-details
        density="compact"
        variant="outlined"
        class="message-input flex-grow-1"
        aria-label="Поле ввода сообщения"
        @keyup.enter="sendMessage"
      />
      <v-btn icon color="#28A745" @click="sendMessage" aria-label="Отправить сообщение" class="ml-2">
        <v-icon>mdi-send</v-icon>
      </v-btn>
    </v-card-actions>
    <v-snackbar v-model="showError" color="error" timeout="3000" aria-label="Сообщение об ошибке">
      {{ errorMessage }}
      <template v-slot:actions>
        <v-btn color="white" @click="showError = false" aria-label="Закрыть уведомление">
          Закрыть
        </v-btn>
      </template>
    </v-snackbar>
  </v-card>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRuntimeConfig } from 'nuxt/app'
import { useChatStore } from '~/stores/chat'

const props = defineProps({
  receiverId: {
    type: Number,
    required: true
  },
  receiverFullName: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['close'])

const config = useRuntimeConfig()
const chatStore = useChatStore()
const message = ref('')
const messagesList = ref(null)
const userId = ref(null)
const loading = ref(false)
const errorMessage = ref('')
const showError = ref(false)

const sortedMessages = computed(() => {
  return (chatStore.messages[props.receiverId] || []).sort((a, b) => new Date(a.created_at) - new Date(b.created_at))
})

onMounted(async () => {
  if (process.client) {
    userId.value = parseInt(localStorage.getItem('userId'))
    chatStore.connectWebSocket()
  }
  loading.value = true
  try {
    await chatStore.fetchMessages(props.receiverId)
    await chatStore.markRead(props.receiverId) // Отметить сообщения как прочитанные
    scrollToBottom()
  } catch (error) {
    errorMessage.value = 'Ошибка загрузки сообщений: ' + (error.message || 'Неизвестная ошибка')
    showError.value = true
    console.error('onMounted error:', error)
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  chatStore.closeWebSocket()
})

watch(() => chatStore.messages[props.receiverId], () => {
  nextTick(() => scrollToBottom())
}, { deep: true })

watch(() => props.receiverId, async (newId) => {
  loading.value = true
  try {
    await chatStore.fetchMessages(newId)
    await chatStore.markRead(newId) // Отметить сообщения как прочитанные при смене получателя
    scrollToBottom()
  } catch (error) {
    errorMessage.value = 'Ошибка загрузки сообщений: ' + (error.message || 'Неизвестная ошибка')
    showError.value = true
    console.error('watch receiverId error:', error)
  } finally {
    loading.value = false
  }
})

async function sendMessage() {
  if (!message.value.trim()) return
  try {
    const response = await $fetch(`${config.public.apiBase}/api/messages`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        receiver_id: props.receiverId,
        content: message.value.trim()
      })
    })
    chatStore.addMessage(props.receiverId, response)
    await chatStore.fetchDialogs() // Обновить диалоги для last_message и unread_count
    message.value = ''
    scrollToBottom()
  } catch (error) {
    errorMessage.value = 'Ошибка отправки сообщения: ' + (error.message || 'Неизвестная ошибка')
    showError.value = true
    console.error('sendMessage error:', error)
  }
}

function scrollToBottom() {
  if (messagesList.value) {
    messagesList.value.scrollTop = messagesList.value.scrollHeight
  }
}

function closeChat() {
  emit('close')
}

function formatDate(timestamp) {
  return new Date(timestamp).toLocaleString('ru-RU', { timeStyle: 'short', dateStyle: 'short' })
}
</script>

<style scoped>
.messages-container {
  padding: 12px;
  background-color: #f5f7fa;
  scroll-behavior: smooth;
}
.messages-wrapper {
  display: flex;
  flex-direction: column;
  gap: 8px;
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
  padding: 10px 14px;
  border-radius: 18px;
  position: relative;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  font-size: 14px;
  line-height: 1.4;
}
.message-bubble:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
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
}
.message-time {
  font-size: 10px;
  opacity: 0.6;
  display: block;
  margin-top: 4px;
  text-align: right;
}
.message-input :deep(.v-field) {
  border-radius: 24px;
  background-color: #f1f3f5;
  min-height: 40px;
}
.message-input :deep(.v-field__input) {
  font-size: 14px;
  padding: 0 12px;
}
:deep(.v-btn) {
  transition: background-color 0.2s ease;
}
:deep(.v-btn:hover) {
  background-color: #2f855a;
}
</style>