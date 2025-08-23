<template>
  <v-container fluid class="pa-0">
    <v-row no-gutters>
      <v-col cols="12" md="4" class="dialogs-col">
        <v-card height="calc(100vh - 64px)" class="overflow-y-auto" aria-label="Список чатов">
          <v-card-title class="justify-center">
            <h2 aria-label="Чаты">Чаты</h2>
          </v-card-title>
          <v-card-text>
            <v-list aria-label="Список диалогов" v-if="chatStore.dialogs.length">
              <v-list-item 
                v-for="dialog in chatStore.dialogs" 
                :key="dialog.user_id" 
                @click="selectDialog(dialog.user_id)"
                :class="{ 'font-weight-bold': dialog.unread_count > 0, 'bg-primary': selectedId === dialog.user_id }"
                class="dialog-item"
                aria-label="Диалог"
              >
                <v-list-item-avatar size="16" class="avatar">
                  <v-img :src="dialog.avatar_url || '/images/nutri-placeholder.jpg'" class="avatar-img" aria-label="Аватар собеседника" />
                </v-list-item-avatar>
                <v-list-item-content>
                  <v-list-item-title class="text-body-2" aria-label="Имя собеседника">
                    {{ dialog.full_name || `User ${dialog.user_id}` }}
                  </v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
            <v-list-item v-else aria-label="Нет диалогов">
              <v-list-item-content>
                <v-list-item-title>Нет диалогов. Начните новый чат!</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="8" v-if="selectedId" class="chat-col">
        <ChatWindow :receiver-id="selectedId" @update-dialogs="loadDialogs" />
      </v-col>
      <v-col cols="12" md="8" v-else class="d-flex align-center justify-center">
        <p aria-label="Выберите диалог">Выберите диалог для просмотра</p>
      </v-col>
    </v-row>
    <v-snackbar v-model="snackbar" color="error" timeout="3000" aria-label="Уведомление об ошибке">
      {{ errorMessage }}
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useChatStore } from '~/stores/chat'
import { useNuxtApp } from 'nuxt/app'

const { $emitter } = useNuxtApp()
const chatStore = useChatStore()
const route = useRoute()
const router = useRouter()
const selectedId = ref(null)
const snackbar = ref(false)
const errorMessage = ref('')

onMounted(async () => {
  await loadDialogs()
  if (route.query.selected) {
    console.log('Auto-selecting dialog from query:', route.query.selected)
    selectedId.value = parseInt(route.query.selected)
    await selectDialog(selectedId.value)
  }
  $emitter.on('message', handleNewMessage)
  $emitter.on('chat:started', handleChatStarted)
})

watch(() => route.query.selected, async (newId) => {
  if (newId && parseInt(newId) !== selectedId.value) {
    console.log('Watch: selecting dialog from query:', newId)
    selectedId.value = parseInt(newId)
    await selectDialog(selectedId.value)
  } else if (!newId) {
    selectedId.value = null
  }
})

onUnmounted(() => {
  chatStore.messages = []
  $emitter.off('message')
  $emitter.off('chat:started')
})

const loadDialogs = async () => {
  try {
    await chatStore.fetchDialogs()
    console.log('Dialogs loaded:', chatStore.dialogs)
  } catch (error) {
    errorMessage.value = `Ошибка загрузки диалогов: ${error.message || 'Неизвестная ошибка'}`
    snackbar.value = true
  }
}

const selectDialog = async (id) => {
  if (!id) return
  if (id === selectedId.value) return
  console.log('Selecting dialog with id:', id)
  selectedId.value = id
  router.replace({ query: { selected: id } })
  try {
    await chatStore.fetchMessages(id)
    await nextTick()
  } catch (error) {
    errorMessage.value = `Ошибка загрузки сообщений: ${error.message || 'Неизвестная ошибка'}`
    snackbar.value = true
  }
}

const handleNewMessage = () => {
  loadDialogs()
}

const handleChatStarted = () => {
  loadDialogs()
  if (route.query.selected) {
    selectDialog(parseInt(route.query.selected))
  }
}
</script>

<style scoped>
.dialogs-col {
  padding: 0;
  border-right: 1px solid #e0e0e0;
}
.chat-col {
  padding: 0;
}
.dialog-item {
  border-radius: 6px;
  margin: 1px 0;
  padding: 1px 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  min-height: 16px;
}
.dialog-item:hover {
  background-color: #f5f5f5;
}
.bg-primary {
  background-color: #28A745 !important;
  color: #FFF !important;
}
.bg-primary .v-list-item-title {
  color: #FFF !important;
}
.avatar {
  margin-right: 4px;
  width: 16px;
  height: 16px;
}
.avatar-img {
  object-fit: cover;
  border-radius: 50%;
}
.v-list-item-content {
  display: flex;
  align-items: center;
  min-width: 0;
}
.v-list-item-title {
  font-size: 10px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}
@media (max-width: 960px) {
  .dialogs-col {
    border-right: none;
    border-bottom: 1px solid #e0e0e0;
  }
  .chat-col {
    padding-top: 8px;
  }
  .dialog-item {
    padding: 1px 3px;
  }
}
</style>