<template>
    <v-container>
      <v-row justify="center">
        <v-col cols="12">
          <v-card class="pa-6" aria-label="Админ-панель выплат">
            <v-card-title class="justify-center">
              <h2 aria-label="Выплаты нутрициологам">Выплаты нутрициологам</h2>
            </v-card-title>
            <v-card-text>
              <v-alert v-if="errorMessage" type="error" dismissible class="mb-4" aria-label="Сообщение об ошибке">
                {{ errorMessage }}
              </v-alert>
              <v-data-table
                :headers="headers"
                :items="nutris"
                :sort-by="[{ key: 'balance', order: 'desc' }]"
                :loading="loading"
                aria-label="Таблица нутрициологов"
                no-data-text="Нет доступных данных"
              >
                <template v-slot:item.full_name="{ item }">
                  {{ item.full_name || 'Не указано' }}
                </template>
                <template v-slot:item.balance="{ item }">
                  {{ item.balance }} руб.
                </template>
                <template v-slot:item.payout_amount="{ item }">
                  {{ item.payout_amount || 0 }} руб.
                </template>
                <template v-slot:item.encrypted_card="{ item }">
                  {{ item.encrypted_card || 'Не указана' }}
                </template>
                <template v-slot:item.actions="{ item }">
                  <v-text-field
                    v-model.number="payoutAmounts[item.id]"
                    label="Сумма выплаты"
                    type="number"
                    :rules="[v => v >= 0 || 'Сумма не может быть отрицательной']"
                    dense
                    outlined
                    aria-label="Сумма выплаты"
                  />
                  <v-btn 
                    color="primary" 
                    small 
                    :disabled="payoutAmounts[item.id] <= 0"
                    @click="processPayout(item.id)" 
                    v-tooltip="'Выплатить'"
                    aria-label="Выплатить"
                  >
                    Выплатить
                  </v-btn>
                  <v-btn 
                    color="secondary" 
                    small 
                    @click="decryptCard(item.id)" 
                    v-tooltip="'Расшифровать карту'"
                    aria-label="Расшифровать карту"
                    :disabled="!item.encrypted_card"
                  >
                    Расшифровать
                  </v-btn>
                </template>
              </v-data-table>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
      <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000" aria-label="Уведомление">
        {{ snackbarText }}
      </v-snackbar>
    </v-container>
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue'
  import { useRuntimeConfig } from 'nuxt/app'
  
  definePageMeta({ middleware: 'admin' })
  
  const config = useRuntimeConfig()
  const loading = ref(false)
  const errorMessage = ref('')
  const snackbar = ref(false)
  const snackbarText = ref('')
  const snackbarColor = ref('success')
  const nutris = ref([])
  const payoutAmounts = ref({})
  const headers = [
    { title: 'Имя', key: 'full_name', align: 'start' },
    { title: 'Баланс', key: 'balance', align: 'center' },
    { title: 'Выплачено', key: 'payout_amount', align: 'center' },
    { title: 'Карта', key: 'encrypted_card', align: 'center' },
    { title: 'Действия', key: 'actions', align: 'end', sortable: false }
  ]
  
  onMounted(async () => {
    await loadNutris()
  })
  
  async function loadNutris() {
    loading.value = true
    const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
    try {
      const data = await $fetch(`${config.public.apiBase}/api/admin/nutris`, { headers })
      console.log('admin.vue: Загружено нутрициологов:', data)
      nutris.value = data || []
      nutris.value.forEach(nutri => {
        payoutAmounts.value[nutri.id] = Number(nutri.payout_amount) || 0
      })
    } catch (error) {
      console.error('admin.vue: Ошибка загрузки нутрициологов:', error)
      errorMessage.value = 'Ошибка загрузки нутрициологов: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
      nutris.value = []
    } finally {
      loading.value = false
    }
  }
  
  async function processPayout(userId) {
    const amount = payoutAmounts.value[userId]
    if (!amount || amount <= 0) {
      snackbarText.value = 'Укажите корректную сумму выплаты'
      snackbarColor.value = 'error'
      snackbar.value = true
      return
    }
    const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
    try {
      await $fetch(`${config.public.apiBase}/api/admin/payout`, {
        method: 'POST',
        headers,
        body: { user_id: userId, amount }
      })
      // Локально обновляем баланс нутрициолога
      nutris.value = nutris.value.map(nutri => {
        if (nutri.id === userId) {
          return { ...nutri, balance: nutri.balance - amount, payout_amount: (Number(nutri.payout_amount) || 0) + amount }
        }
        return nutri
      })
      // Очищаем поле ввода суммы
      payoutAmounts.value[userId] = 0
      snackbarText.value = `Выплата ${amount} руб. успешно обработана`
      snackbarColor.value = 'success'
      snackbar.value = true
      // Синхронизируем данные с сервером
      await loadNutris()
    } catch (error) {
      console.error('admin.vue: Ошибка выплаты:', error)
      snackbarText.value = 'Ошибка выплаты: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
      snackbarColor.value = 'error'
      snackbar.value = true
    }
  }
  
  async function updatePayoutAmount(userId) {
    const amount = payoutAmounts.value[userId]
    if (amount < 0) {
      snackbarText.value = 'Выплаченная сумма не может быть отрицательной'
      snackbarColor.value = 'error'
      snackbar.value = true
      return
    }
    const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
    const body = { user_id: userId, payout_amount: amount }
    try {
      await $fetch(`${config.public.apiBase}/api/admin/update-payout-amount`, { 
        method: 'POST', 
        headers, 
        body 
      })
      snackbarText.value = `Выплаченная сумма обновлена: ${amount} руб.`
      snackbarColor.value = 'success'
      snackbar.value = true
      await loadNutris()
    } catch (error) {
      console.error('admin.vue: Ошибка обновления суммы:', error)
      snackbarText.value = 'Ошибка обновления выплаченной суммы: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
      snackbarColor.value = 'error'
      snackbar.value = true
    }
  }
  
  async function decryptCard(userId) {
    const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
    try {
      const data = await $fetch(`${config.public.apiBase}/api/admin/decrypt-card`, {
        method: 'POST',
        headers,
        body: { user_id: userId }
      })
      snackbarText.value = `Номер карты: ${data.card_number}`
      snackbarColor.value = 'success'
      snackbar.value = true
    } catch (error) {
      console.error('admin.vue: Ошибка расшифровки карты:', error)
      snackbarText.value = 'Ошибка расшифровки карты: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
      snackbarColor.value = 'error'
      snackbar.value = true
    }
  }
  </script>