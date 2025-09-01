<template>
  <v-container fluid class="pa-0" style="height: calc(100vh - 64px);">
    <v-row no-gutters style="height: 100%;">
      <v-col cols="12" sm="4" md="3" class="dialogs-panel" style="border-right: 1px solid #e5e7eb; overflow-y: auto; background-color: #f9fafb;">
        <v-card flat style="background: transparent;">
          <v-card-title class="py-2 px-4" style="font-size: 16px; color: #2E7D32; position: sticky; top: 0; z-index: 1; background-color: #ffffff; border-bottom: 1px solid #e5e7eb;">
            Диалоги
          </v-card-title>
          <v-list class="pa-0" aria-label="Список диалогов">
            <v-list-item
              v-for="dialog in sortedDialogs"
              :key="dialog.user_id"
              :class="{ 'selected-dialog': selectedDialog && selectedDialog.user_id === dialog.user_id }"
              @click="selectDialog(dialog)"
              class="py-1 px-3"
              style="cursor: pointer; min-height: 48px; border-bottom: 1px solid #e5e7eb;"
              aria-label="Диалог"
            >
              <v-list-item-avatar size="36" class="mr-2">
                <v-img
                  :src="dialog.avatar_url || '/images/nutri-placeholder.jpg'"
                  alt="Аватар пользователя"
                  @error="dialog.avatar_url = '/images/nutri-placeholder.jpg'"
                />
              </v-list-item-avatar>
              <v-list-item-content class="py-0">
                <v-list-item-title class="text-subtitle-2" style="font-size: 15px; font-weight: 500; color: #1f2a44; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" aria-label="Имя собеседника">
                  {{ dialog.full_name || dialog.username || 'Без имени' }}
                </v-list-item-title>
                <v-list-item-subtitle class="text-caption" style="font-size: 12px; color: #6b7280; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" aria-label="Последнее сообщение">
                  {{ dialog.last_message ? dialog.last_message.slice(0, 25) + (dialog.last_message.length > 25 ? '...' : '') : 'Нет сообщений' }}
                </v-list-item-subtitle>
              </v-list-item-content>
              <v-list-item-action class="ml-1">
                <v-badge
                  v-if="dialog.unread_count > 0"
                  :content="dialog.unread_count"
                  color="#EF4444"
                  overlap
                  small
                  aria-label="Количество непрочитанных сообщений"
                />
              </v-list-item-action>
            </v-list-item>
          </v-list>
        </v-card>
      </v-col>
      <v-col cols="12" sm="8" md="9" class="chat-panel" style="height: 100%; display: flex; flex-direction: column; background-color: #ffffff;">
        <ChatWindow 
          v-if="selectedDialog" 
          :receiver-id="selectedDialog.user_id" 
          :receiver-full-name="selectedDialog.full_name || selectedDialog.username || 'Без имени'" 
          @close="selectedDialog = null"
          :key="selectedDialog.user_id"
        />
        <div v-else class="d-flex align-center justify-center" style="height: 100%; color: #6b7280; font-size: 15px;">
          <p aria-label="Выберите диалог">Выберите диалог для начала чата</p>
        </div>
      </v-col>
    </v-row>
    <v-snackbar v-model="showError" color="error" timeout="3000" aria-label="Сообщение об ошибке">
      {{ errorMessage }}
      <template v-slot:actions>
        <v-btn color="white" @click="showError = false" aria-label="Закрыть уведомление">
          Закрыть
        </v-btn>
      </template>
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'nuxt/app'
import { useChatStore } from '~/stores/chat'
import ChatWindow from '~/components/ChatWindow.vue'

const router = useRouter()
const route = useRoute()
const chatStore = useChatStore()
const dialogs = ref([])
const selectedDialog = ref(null)
const loading = ref(false)
const errorMessage = ref('')
const showError = ref(false)
const token = ref(null)

const sortedDialogs = computed(() => {
  return [...dialogs.value].sort((a, b) => {
    const aTime = a.last_message_at ? new Date(a.last_message_at) : new Date(a.created_at || 0)
    const bTime = b.last_message_at ? new Date(b.last_message_at) : new Date(b.created_at || 0)
    return bTime - aTime
  })
})

onMounted(async () => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    if (!token.value) {
      router.push('/login')
      return
    }
    chatStore.connectWebSocket()
  }
  await loadDialogs()
  const selectedId = route.query.selected ? parseInt(route.query.selected) : null
  if (selectedId) {
    const dialog = dialogs.value.find(d => d.user_id === selectedId)
    if (dialog) selectDialog(dialog)
  }
})

watch(() => route.query.selected, (newSelected) => {
  const selectedId = newSelected ? parseInt(newSelected) : null
  if (selectedId) {
    const dialog = dialogs.value.find(d => d.user_id === selectedId)
    if (dialog) selectDialog(dialog)
  } else {
    selectedDialog.value = null
  }
})

async function loadDialogs() {
  loading.value = true
  try {
    await chatStore.fetchDialogs()
    dialogs.value = chatStore.dialogs
  } catch (error) {
    errorMessage.value = 'Ошибка загрузки диалогов: ' + (error.message || 'Неизвестная ошибка')
    showError.value = true
    console.error('loadDialogs error:', error)
  } finally {
    loading.value = false
  }
}

function selectDialog(dialog) {
  selectedDialog.value = dialog
  router.replace({ query: { selected: dialog.user_id } })
  chatStore.fetchMessages(dialog.user_id)
  chatStore.markRead(dialog.user_id)
}
</script>

<style scoped>
.dialogs-panel {
  height: 100%;
  overflow-y: auto;
}
.chat-panel {
  height: 100%;
  overflow: hidden;
}
.selected-dialog {
  background-color: #E8F5E9;
}
.v-list-item {
  min-height: 48px !important;
  padding: 8px 12px !important;
  display: flex;
  align-items: center;
}
.v-list-item-avatar {
  margin-right: 12px !important;
}
.v-list-item-content {
  flex: 1 1 auto;
}
.v-list-item-title {
  line-height: 1.2;
  font-size: 15px;
  font-weight: 500;
}
.v-list-item-subtitle {
  line-height: 1.2;
  margin-top: 4px;
  font-size: 12px;
}
.v-list-item-action {
  margin-left: 8px;
  min-width: 24px;
}
</style>