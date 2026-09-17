<template>
  <div class="min-h-screen flex flex-col font-sans">
    <main class="flex-1 p-6">
      <div class="w-full max-w-[1800px] mx-auto px-4">
        
        <!-- Header Controls -->
        <div class="mb-6 flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4">
          <div>
            <h2 class="text-3xl font-semibold text-white mb-2">Dashboard</h2>
          </div>
          
          <div class="flex flex-wrap items-center gap-4 bg-white/5 border border-white/10 rounded-xl p-3 backdrop-blur-sm shadow-lg">
            
            <div class="flex items-center gap-2 border-r border-white/10 pr-4">
              <div class="w-2.5 h-2.5 rounded-full bg-[#34d399] animate-pulse shadow-[0_0_8px_#34d399]"></div>
              <span class="text-sm font-medium text-gray-300">System Online</span>
            </div>

            <div class="flex items-center gap-2 border-r border-white/10 pr-4 hidden md:flex">
              <span class="text-gray-500 text-[10px] uppercase font-bold tracking-wide">Discovered:</span>
              <span class="text-white font-semibold text-sm">{{ discoveredCount }}</span>
            </div>

            <div class="flex items-center gap-2 border-r border-white/10 pr-4 hidden md:flex">
              <span class="text-gray-500 text-[10px] uppercase font-bold tracking-wide">Monitored:</span>
              <span class="text-[#FF5FCF] font-semibold text-sm">{{ monitoredCount }}</span>
            </div>

            <!-- COLUMN DENSITY SWITCHER -->
            <div class="flex items-center gap-1 border-r border-white/10 pr-4">
              <span class="text-gray-500 text-[10px] uppercase font-bold tracking-wide mr-1 hidden sm:inline">Layout:</span>
              <div class="flex bg-black/40 p-1 rounded-lg border border-white/5 gap-1">
                <button 
                  v-for="col in [1, 2, 3, 4]" 
                  :key="col"
                  @click="selectedCols = col"
                  :class="[
                    'px-2 py-0.5 text-xs font-bold rounded transition-all cursor-pointer',
                    selectedCols === col ? 'bg-[#FF5FCF] text-white shadow' : 'text-gray-400 hover:text-white hover:bg-white/10'
                  ]"
                  :title="`${col} graph${col > 1 ? 's' : ''} per row`"
                >
                  {{ col }}x
                </button>
              </div>
            </div>

            <button @click="addWidget" type="button" 
                    :disabled="monitoredInterfaces.length === 0"
                    :class="['flex items-center gap-1 text-sm font-bold uppercase transition-colors', monitoredInterfaces.length === 0 ? 'text-gray-600 cursor-not-allowed' : 'text-emerald-400 hover:text-emerald-300 cursor-pointer']">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
              Add Graph
            </button>

          </div>
        </div>

        <!-- DYNAMIC GRID -->
        <div v-if="monitoredInterfaces.length > 0" :class="['grid gap-6 dashboard-grid', gridClass]">
          <div v-for="(widget, index) in widgets" :key="widget.id" class="relative group transition-all duration-300 ease-in-out h-full">
            <GraphWidget 
              :monitoredInterfaces="monitoredInterfaces" 
              :widgetId="widget.id"
              :widgetConfig="widget" 
              :canDelete="true"
              @remove="removeWidget(index)" 
              @update-widget-config="updateWidgetConfig(index, $event)"
            />
          </div>
        </div>

        <!-- EMPTY STATE -->
        <div v-else class="text-center text-gray-500 mt-20 flex flex-col items-center justify-center min-h-[300px] border border-dashed border-white/10 rounded-2xl bg-white/5 p-8">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mb-4 text-gray-700 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" /></svg>
          <p class="text-lg font-medium text-gray-400">No active monitored interfaces found.</p>
          <p class="text-sm text-gray-600 mt-1">Make sure you have configured hosts and enabled monitoring for their interfaces in the Settings tab.</p>
        </div>

      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import GraphWidget from './GraphWidget.vue'

// --- GRID LAYOUT DENSITY PREFERENCE ---
const selectedCols = ref(2) // Default to 2 per row

watch(selectedCols, () => {
  // Trigger window resize event so ApexCharts redraws instantly when switching columns
  setTimeout(() => window.dispatchEvent(new Event('resize')), 100)
})

const gridClass = computed(() => {
  switch (selectedCols.value) {
    case 1: return 'grid-cols-1'
    case 2: return 'grid-cols-1 lg:grid-cols-2'
    case 3: return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3'
    case 4: return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'
    default: return 'grid-cols-1 lg:grid-cols-2'
  }
})

// --- MULTI-GRAPH WIDGET STATE ---
const widgets = ref([])

const addWidget = () => {
  widgets.value.push({ 
    id: Date.now(), 
    portId: null, 
    duration: '1h', 
    isExpanded: false,
    startDateTime: '',
    endDateTime: ''
  })
}

const removeWidget = (index) => {
  widgets.value.splice(index, 1)
}

const updateWidgetConfig = (index, newConfig) => {
  widgets.value[index] = { ...widgets.value[index], ...newConfig }
}

// --- BACKEND SYNC LOGIC ---
const fetchDashboardConfig = async () => {
  const token = localStorage.getItem('hermes_token')
  try {
    const res = await fetch('/api/dashboard', { headers: { 'Authorization': `Bearer ${token}` } })
    const data = await res.json()
    
    if (data.config) {
      const parsed = JSON.parse(data.config)
      selectedCols.value = parsed.grid_cols || 2
      widgets.value = parsed.widgets || []
    }
    
    // If no widgets exist (new user or empty config), create a default one
    if (widgets.value.length === 0) {
      addWidget()
    }
  } catch (e) {
    console.error("Failed to load dashboard config", e)
    if (widgets.value.length === 0) addWidget()
  }
}

let saveTimeout = null
const saveDashboardToBackend = () => {
  clearTimeout(saveTimeout)
  // Debounce the save to prevent spamming the DB on rapid changes
  saveTimeout = setTimeout(async () => {
    const token = localStorage.getItem('hermes_token')
    const payload = {
      grid_cols: selectedCols.value,
      widgets: widgets.value
    }
    
    try {
      await fetch('/api/dashboard', {
        method: 'PUT',
        headers: { 
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}` 
        },
        body: JSON.stringify({ config: JSON.stringify(payload) })
      })
    } catch (e) {
      console.error("Failed to save dashboard config", e)
    }
  }, 1000) // Save 1 second after the last change
}

// Watch for any changes to the layout or widgets and save to DB
watch([selectedCols, widgets], () => {
  saveDashboardToBackend()
}, { deep: true })

// --- DASHBOARD DATA STATE ---
const hosts = ref([])
const interfaces = ref([])

const discoveredCount = computed(() => Array.isArray(interfaces.value) ? interfaces.value.length : 0)
const monitoredCount = computed(() => Array.isArray(interfaces.value) ? interfaces.value.filter(i => i.monitoring).length : 0)
const monitoredInterfaces = computed(() => Array.isArray(interfaces.value) ? interfaces.value.filter(i => i.monitoring) : [])

let interfacePollInterval = null

const fetchHosts = async () => {
  try {
    const response = await fetch('/api/hosts')
    const data = await response.json()
    if (Array.isArray(data) && data.length > 0) {
      hosts.value = data
      fetchInterfaces()
    }
  } catch (error) {
    console.error("Failed to fetch hosts:", error)
  }
}

const fetchInterfaces = async () => {
  if (!hosts.value || hosts.value.length === 0) return
  
  try {
    const promises = hosts.value.map(async (host) => {
      const response = await fetch(`/api/interfaces?host_id=${host.id}`)
      const data = await response.json()
      if (Array.isArray(data)) {
        return data.map(iface => ({ ...iface, hostName: host.name }))
      }
      return []
    })
    
    const results = await Promise.all(promises)
    interfaces.value = results.flat()
  } catch (error) {
    console.error("Failed to fetch interfaces:", error)
    interfaces.value = [] 
  }
}

onMounted(() => {
  fetchDashboardConfig()
  fetchHosts()
  interfacePollInterval = setInterval(fetchInterfaces, 60000) 
})

onUnmounted(() => {
  if (interfacePollInterval) clearInterval(interfacePollInterval)
})
</script>

<style>
.dashboard-grid > div:has(.lg\:col-span-full) {
  grid-column: 1 / -1 !important;
}
</style>