<!-- pages/admin.vue -->
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
                    :rules="[v => v >= 0 && v <= item.balance || 'Неверная сумма']"
                    dense
                    outlined
                    aria-label="Сумма выплаты"
                  />
                  <v-btn 
                    color="primary" 
   Rosanne B. 4:08 PM
                    small 
                    @click="processPayout(item.id)" 
                    v-tooltip="'Выплатить'" 
                    aria-label="Выплатить"
                    :disabled="payoutAmounts[item.id] > item.balance || payoutAmounts[item.id] <= 0"
                  >
                    Выплатить
                  </v-btn>
                  <v-text-field
                    v-model.number="payoutPaidAmounts[item.id]"
                    label="Выплачено"
                    type="number"
                    :rules="[v => v >= 0 || 'Сумма не может быть отрицательной']"
                    dense
                    outlined
                    aria-label="Сумма выплачено"
                  />
                  <v-btn 
                    color="success" 
                    small 
                    @click="updatePayoutAmount(item.id)" 
                    v-tooltip="'Обновить выплаченную сумму'" 
                    aria-label=" Updating the paid amount"
                    :disabled="payoutPaidAmounts[item.id] < 0"
                  >
                    Обновить
                  </v-btn>
                  <v-btn 
                    color="secondary" 
                    small 
                    @click="decryptCard(item.id)" 
                    v-tooltip="'Показать карту'" 
                    aria-label="Показать карту"
                    :disabled="!item.encrypted_card || item.encrypted_card === 'Не указана' || item.encrypted_card === 'Ошибка расшифровки'"
                  >
                    Показать карту
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
  import { useRouter } from 'vue-router'
  import { useRuntimeConfig } from 'nuxt/app'
  
  definePageMeta({ middleware: 'admin' })
  
  const config = useRuntimeConfig()
  const router = useRouter()
  const nutris = ref([])
  const payoutAmounts = ref({})
  const payoutPaidAmounts = ref({})
  const loading = ref(true)
  const errorMessage = ref('')
  const snackbar = ref(false)
  const snackbarText = ref('')
  const snackbarColor = ref('success')
  const headers = [
    { title: 'ФИО', key: 'full_name' },
    { title: 'Баланс', key: 'balance' },
    { title: 'Выплачено', key: 'payout_amount' },
    { title: 'Карта', key: 'encrypted_card' },
    { title: 'Действия', key: 'actions', sortable: false }
  ]
  
  onMounted(async () => {
    const token = localStorage.getItem('token')
    if (!token || localStorage.getItem('role') !== 'admin') {
      router.push('/login')
      return
    }
    await loadNutris()
  })
  
  async function loadNutris() {
    loading.value = true
    const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
    try {
      const data = await $fetch(`${config.public.apiBase}/api/admin/nutris`, { headers })
      nutris.value = data || []
      nutris.value.forEach(nutri => {
        payoutAmounts.value[nutri.id] = 0
        payoutPaidAmounts.value[nutri.id] = nutri.payout_amount || 0
      })
    } catch (error) {
      errorMessage.value = 'Ошибка загрузки: ' + (error.message || 'Неизвестная ошибка')
      snackbarText.value = errorMessage.value
      snackbarColor.value = 'error'
      snackbar.value = true
    } finally {
      loading.value = false
    }
  }
  
  async function processPayout(userId) {
    const amount = payoutAmounts.value[userId]
    if (amount <= 0) {
      snackbarText.value = 'Сумма должна быть больше 0'
      snackbarColor.value = 'error'
      snackbar.value = true
      return
    }
    const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
    const body = { user_id: userId, amount }
    try {
      const response = await $fetch(`${config.public.apiBase}/api/admin/payout`, { 
        method: 'POST', 
        headers, 
        body 
      })
      snackbarText.value = `Выплата на ${amount} руб. инициирована. Выполните перевод вручную.`
      snackbarColor.value = 'success'
      snackbar.value = true
      // Обновляем payout_amount
      await updatePayoutAmount(userId, amount)
      await loadNutris()
    } catch (error) {
      snackbarText.value = 'Ошибка выплаты: ' + (error.message || 'Неизвестная ошибка')
      snackbarColor.value = 'error'
      snackbar.value = true
    }
  }
  
  async function updatePayoutAmount(userId, amount = null) {
    const paidAmount = amount !== null ? amount : payoutPaidAmounts.value[userId]
    if (paidAmount < 0) {
      snackbarText.value = 'Выплаченная сумма не может быть отрицательной'
      snackbarColor.value = 'error'
      snackbar.value = true
      return
    }
    const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
    const body = { user_id: userId, payout_amount: paidAmount }
    try {
      await $fetch(`${config.public.apiBase}/api/admin/update-payout-amount`, { 
        method: 'POST', 
        headers, 
        body 
      })
      snackbarText.value = `Выплаченная сумма обновлена: ${paidAmount} руб.`
      snackbarColor.value = 'success'
      snackbar.value = true
      await loadNutris()
    } catch (error) {
      snackbarText.value = 'Ошибка обновления выплаченной суммы: ' + (error.message || 'Неизвестная ошибка')
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
      snackbarText.value = 'Ошибка расшифровки карты: ' + (error.message || 'Неизвестная ошибка')
      snackbarColor.value = 'error'
      snackbar.value = true
    }
  }
  </script>