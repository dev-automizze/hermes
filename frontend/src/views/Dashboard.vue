<template>
  <div class="min-h-screen flex flex-col font-sans">
    <main class="flex-1 p-6">
      <div class="w-full max-w-[1800px] mx-auto px-4">
        
        <div class="mb-6 flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4">
          <div>
            <h2 class="text-3xl font-semibold text-white mb-2">Dashboard</h2>
          </div>
          
          <div class="flex flex-wrap items-center gap-6 bg-white/5 border border-white/10 rounded-xl p-3 backdrop-blur-sm shadow-lg">
            
            <div class="flex items-center gap-3 border-r border-white/10 pr-6">
              <label class="text-[10px] font-bold uppercase text-gray-500 tracking-wider">Device</label>
              <div class="relative flex items-center">
                <select v-model="selectedHostId" @change="loadHostData" 
                        class="bg-transparent text-emerald-400 font-semibold focus:outline-none cursor-pointer appearance-none pr-6">
                  <option v-if="hosts.length === 0" value="null" disabled>No hosts configured</option>
                  <option v-for="h in hosts" :key="h.id" :value="h.id" class="bg-gray-800 text-white">
                    {{ h.name }} ({{ h.ip }})
                  </option>
                </select>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-emerald-500 absolute right-0 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </div>
            </div>

            <div class="flex items-center gap-2 border-r border-white/10 pr-6">
              <div class="w-2.5 h-2.5 rounded-full bg-[#34d399] animate-pulse shadow-[0_0_8px_#34d399]"></div>
              <span class="text-sm font-medium text-gray-300">System Online</span>
            </div>

            <div class="flex items-center gap-2 border-r border-white/10 pr-6 hidden md:flex">
              <span class="text-gray-500 text-sm uppercase tracking-wide text-[10px] font-bold">Discovered:</span>
              <span class="text-white font-semibold">{{ discoveredCount }}</span>
            </div>

            <div class="flex items-center gap-2 border-r border-white/10 pr-6 hidden md:flex">
              <span class="text-gray-500 text-sm uppercase tracking-wide text-[10px] font-bold">Monitored:</span>
              <span class="text-[#FF5FCF] font-semibold">{{ monitoredCount }}</span>
            </div>

            <button @click="addWidget" type="button" 
                    :disabled="monitoredInterfaces.length === 0"
                    :class="['flex items-center gap-1 text-sm font-bold uppercase transition-colors', monitoredInterfaces.length === 0 ? 'text-gray-600 cursor-not-allowed' : 'text-emerald-400 hover:text-emerald-300 cursor-pointer']">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
              Add Graph
            </button>

          </div>
        </div>

        <div v-if="selectedHostId && monitoredInterfaces.length > 0" class="grid grid-cols-1 xl:grid-cols-2 gap-6 dashboard-grid">
          
          <div v-for="(widget, index) in widgets" :key="widget.id" class="relative group transition-all duration-300 ease-in-out h-full">
            
            <GraphWidget 
              :hostId="selectedHostId" 
              :monitoredInterfaces="monitoredInterfaces" 
              :widgetId="widget.id"
              :canDelete="true"
              @remove="removeWidget(index)" 
            />
            
          </div>
        </div>

        <div v-else class="text-center text-gray-500 mt-20 flex flex-col items-center justify-center min-h-[300px] border border-dashed border-white/10 rounded-2xl bg-white/5 p-8">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mb-4 text-gray-700 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" /></svg>
          <p class="text-lg font-medium text-gray-400">No active monitored interfaces found for this device.</p>
          <p class="text-sm text-gray-600 mt-1">Make sure you have enabled monitoring for interfaces in the Settings tab.</p>
        </div>

      </div>

      <div v-if="deleteModal.isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm px-4">
        <div class="bg-[#121212] border border-white/10 rounded-2xl p-6 w-full max-w-md shadow-2xl transform transition-all">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 bg-red-500/10 rounded-lg border border-red-500/20 text-red-400">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
            </div>
            <h3 class="text-xl font-bold text-white">Remove Graph</h3>
          </div>
          <p class="text-gray-400 text-sm mb-6 ml-11">
            Are you sure you want to remove this widget from your dashboard? You can always add it back later.
          </p>
          <div class="flex justify-end gap-3">
            <button @click="cancelDelete" class="px-5 py-2 rounded-lg text-sm font-bold text-gray-400 hover:text-white hover:bg-white/5 transition-colors cursor-pointer">
              Cancel
            </button>
            <button @click="executeDelete" class="px-5 py-2 rounded-lg text-sm font-bold bg-red-500/20 text-red-400 hover:bg-red-500/40 border border-red-500/30 transition-colors cursor-pointer shadow-lg shadow-red-500/10">
              Remove Graph
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style>
/* THE MAGIC TRICK: 
  This tells the outer grid cage to stretch to full width (grid-column: 1 / -1) 
  whenever the GraphWidget inside of it tries to apply its expansion class.
*/
.dashboard-grid > div:has(.lg\:col-span-full) {
  grid-column: 1 / -1 !important;
}
</style>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import GraphWidget from './GraphWidget.vue'

// --- MULTI-GRAPH WIDGET STATE ---
// Try to load saved widgets from localStorage first
const savedWidgets = localStorage.getItem('hermes_widgets')
const widgets = ref(savedWidgets ? JSON.parse(savedWidgets) : [{ id: Date.now() }])

// Deep watch the array so it saves automatically when you add/remove a graph
watch(widgets, (newVal) => {
  localStorage.setItem('hermes_widgets', JSON.stringify(newVal))
}, { deep: true, immediate: true }) // 👈 FIX 1: Added "immediate: true" so the first Date.now() ID saves instantly!

const addWidget = () => {
  widgets.value.push({ id: Date.now() })
}

// --- MODAL STATE LOGIC ---
const deleteModal = ref({
  isOpen: false,
  index: null
})

const confirmRemoveWidget = (index) => {
  deleteModal.value = { isOpen: true, index }
}

const executeDelete = () => {
  if (deleteModal.value.index !== null) {
    removeWidget(deleteModal.value.index)
  }
  deleteModal.value.isOpen = false
}

const cancelDelete = () => {
  deleteModal.value.isOpen = false
}

const removeWidget = (index) => {
  const widgetToRemove = widgets.value[index]
  localStorage.removeItem(`hermes_port_${widgetToRemove.id}`)
  localStorage.removeItem(`hermes_duration_${widgetToRemove.id}`)
  localStorage.removeItem(`hermes_expanded_${widgetToRemove.id}`)
  widgets.value.splice(index, 1)
}

// --- DASHBOARD DATA STATE ---
const hosts = ref([])
const selectedHostId = ref(null)
const interfaces = ref([])

// SAFE COMPUTED PROPERTIES
const discoveredCount = computed(() => Array.isArray(interfaces.value) ? interfaces.value.length : 0)
const monitoredCount = computed(() => Array.isArray(interfaces.value) ? interfaces.value.filter(i => i.monitoring).length : 0)
const monitoredInterfaces = computed(() => Array.isArray(interfaces.value) ? interfaces.value.filter(i => i.monitoring) : [])

let interfacePollInterval = null

const fetchHosts = async () => {
  try {
    const apiUrl = '/api/hosts'
    const response = await fetch(apiUrl)
    const data = await response.json()
    if (Array.isArray(data) && data.length > 0) {
      hosts.value = data
      const savedHost = localStorage.getItem('hermes_selected_host')
      if (savedHost && hosts.value.find(h => h.id == savedHost)) {
        selectedHostId.value = parseInt(savedHost)
      } else {
        selectedHostId.value = data[0].id
      }
      loadHostData()
    }
  } catch (error) {
    console.error("Failed to fetch hosts:", error)
  }
}

const loadHostData = () => {
  if (selectedHostId.value) {
    localStorage.setItem('hermes_selected_host', selectedHostId.value)
    fetchInterfaces()
  }
}

const fetchInterfaces = async () => {
  if (!selectedHostId.value) return
  try {
    const apiUrl = `/api/interfaces?host_id=${selectedHostId.value}`
    const response = await fetch(apiUrl)
    const data = await response.json()
    interfaces.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error("Failed to fetch interfaces:", error)
    interfaces.value = [] 
  }
}

onMounted(() => {
  fetchHosts()
  interfacePollInterval = setInterval(fetchInterfaces, 60000) 
})

onUnmounted(() => {
  if (interfacePollInterval) clearInterval(interfacePollInterval)
})
</script>