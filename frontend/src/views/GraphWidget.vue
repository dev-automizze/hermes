<template>
  <div :class="[
  'bg-[#121212] border border-white/10 rounded-2xl p-6 shadow-2xl relative overflow-hidden group h-full transition-all duration-300 ease-in-out',
  isExpanded ? 'col-span-1 lg:col-span-full' : 'col-span-1'
]">
    
    <div class="absolute -top-24 -right-24 w-64 h-64 bg-[#FF5FCF] rounded-full mix-blend-multiply filter blur-[100px] opacity-10"></div>
    <div class="absolute -bottom-24 -left-24 w-64 h-64 bg-[#FFA239] rounded-full mix-blend-multiply filter blur-[100px] opacity-10"></div>

    <div class="flex flex-col 2xl:flex-row justify-between items-start 2xl:items-center mb-6 gap-4 border-b border-white/10 pb-4 relative z-10">
      
      <div class="flex items-center gap-3 w-full 2xl:w-auto">
        <div class="p-2 bg-emerald-500/10 rounded-lg border border-emerald-500/20 text-emerald-400">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
        <div class="flex flex-col w-full">
          <label class="text-[10px] font-bold uppercase text-gray-500 tracking-wider mb-1">Select Interface Graph</label>
          <div class="relative flex items-center w-full max-w-[320px]">
            <select v-model="selectedGraphPort" @change="fetchGraphData" 
                    class="bg-transparent text-white font-semibold text-lg focus:outline-none cursor-pointer appearance-none w-full pr-6 truncate">
              <option v-if="!monitoredInterfaces || monitoredInterfaces.length === 0" value="null" disabled>No active interfaces</option>
              <option v-for="iface in monitoredInterfaces" :key="iface.id" :value="iface.id" class="bg-gray-800 text-white">
                ({{ iface.hostName }}) {{ iface.name }}
              </option>
            </select>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400 absolute right-0 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-3 w-full 2xl:w-auto justify-end">
        
        <div class="flex gap-1 bg-black/40 p-1 rounded-lg border border-white/5 overflow-x-auto">
          <button @click="selectedDuration = '5m'" :class="['px-2 py-1 rounded-md text-xs font-bold transition-all', selectedDuration === '5m' ? 'bg-[#FF5FCF]/20 text-[#FF5FCF]' : 'text-gray-400 hover:text-white hover:bg-white/5']">5M</button>
          <button @click="selectedDuration = '15m'" :class="['px-2 py-1 rounded-md text-xs font-bold transition-all', selectedDuration === '15m' ? 'bg-[#FF5FCF]/20 text-[#FF5FCF]' : 'text-gray-400 hover:text-white hover:bg-white/5']">15M</button>
          <button @click="selectedDuration = '30m'" :class="['px-2 py-1 rounded-md text-xs font-bold transition-all', selectedDuration === '30m' ? 'bg-[#FF5FCF]/20 text-[#FF5FCF]' : 'text-gray-400 hover:text-white hover:bg-white/5']">30M</button>
          <button @click="selectedDuration = '1h'" :class="['px-2 py-1 rounded-md text-xs font-bold transition-all', selectedDuration === '1h' ? 'bg-[#FF5FCF]/20 text-[#FF5FCF]' : 'text-gray-400 hover:text-white hover:bg-white/5']">1H</button>
          <button @click="selectedDuration = '24h'" :class="['px-2 py-1 rounded-md text-xs font-bold transition-all', selectedDuration === '24h' ? 'bg-[#FF5FCF]/20 text-[#FF5FCF]' : 'text-gray-400 hover:text-white hover:bg-white/5']">24H</button>
          <button @click="selectedDuration = '7d'" :class="['px-2 py-1 rounded-md text-xs font-bold transition-all', selectedDuration === '7d' ? 'bg-[#FF5FCF]/20 text-[#FF5FCF]' : 'text-gray-400 hover:text-white hover:bg-white/5']">7D</button>
          <button @click="selectedDuration = 'custom'" :class="['px-2 py-1 rounded-md text-xs font-bold transition-all flex items-center gap-1', selectedDuration === 'custom' ? 'bg-[#FF5FCF]/20 text-[#FF5FCF]' : 'text-gray-400 hover:text-white hover:bg-white/5']">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
            CSTM
          </button>
        </div>

        <button @click="toggleSize" 
                class="p-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg text-gray-400 hover:text-white transition cursor-pointer flex items-center justify-center shrink-0" 
                :title="isExpanded ? 'Shrink Widget' : 'Expand Full Width'">
          <svg v-if="!isExpanded" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 14h6m0 0v6m0-6l-7 7m17-11h-6m0 0V4m0 6l7-7M4 10h6m0 0V4m0 6l-7-7m17 11h-6m0 0v6m0-6l7 7" />
          </svg>
        </button>

        <div v-if="canDelete" class="relative">
          <button @click="showDeleteDropdown = !showDeleteDropdown" 
                  class="p-2 bg-red-500/10 hover:bg-red-500/20 border border-red-500/20 rounded-lg text-red-400 hover:text-red-300 transition cursor-pointer flex items-center justify-center shrink-0" 
                  title="Remove Graph">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          
          <div v-if="showDeleteDropdown" class="absolute right-0 mt-2 w-40 bg-[#1a1a1a] border border-red-500/30 rounded-lg shadow-2xl z-50 overflow-hidden">
            <button @click="confirmDelete" class="w-full text-left px-4 py-2.5 text-sm font-bold text-red-400 hover:bg-red-500/20 transition-colors flex items-center gap-2 cursor-pointer">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
              Delete Graph
            </button>
          </div>
        </div>

      </div>
    </div>

    <div v-if="selectedDuration === 'custom'" class="flex flex-wrap items-end gap-4 mb-6 bg-black/30 p-4 rounded-xl border border-white/5 relative z-10">
      <div>
        <label class="block text-[10px] font-bold uppercase text-gray-500 tracking-wider mb-1">Start Date/Time</label>
        <input type="datetime-local" v-model="startDateTime" class="bg-gray-800 text-white text-sm border border-gray-600 rounded px-3 py-2 focus:outline-none focus:border-[#FF5FCF]">
      </div>
      <div>
        <label class="block text-[10px] font-bold uppercase text-gray-500 tracking-wider mb-1">End Date/Time</label>
        <input type="datetime-local" v-model="endDateTime" class="bg-gray-800 text-white text-sm border border-gray-600 rounded px-3 py-2 focus:outline-none focus:border-[#FF5FCF]">
      </div>
    </div>

    <div class="flex flex-col lg:flex-row justify-between items-stretch mb-6 relative z-10 gap-4">
      <div class="flex-1 flex justify-between items-center bg-black/20 p-4 rounded-xl border border-white/5">
        <div>
          <div class="text-[#FF5FCF] text-[10px] font-bold uppercase mb-1 tracking-wider">Download (Current)</div>
          <div class="text-3xl font-bold text-white tracking-tight font-mono">
            {{ currentDownload.toFixed(2) }} <span class="text-sm font-medium text-gray-500 uppercase">Mbps</span>
          </div>
        </div>
        <div class="flex gap-4 text-right">
          <div>
            <div class="text-gray-500 text-[10px] font-bold uppercase mb-1">Max</div>
            <div class="text-sm font-mono text-white font-semibold">{{ maxDownload.toFixed(2) }}</div>
          </div>
          <div>
            <div class="text-gray-500 text-[10px] font-bold uppercase mb-1">Avg</div>
            <div class="text-sm font-mono text-white font-semibold">{{ avgDownload.toFixed(2) }}</div>
          </div>
          <div>
            <div class="text-gray-500 text-[10px] font-bold uppercase mb-1">Min</div>
            <div class="text-sm font-mono text-white font-semibold">{{ minDownload.toFixed(2) }}</div>
          </div>
        </div>
      </div>

      <div class="flex-1 flex justify-between items-center bg-black/20 p-4 rounded-xl border border-white/5">
        <div>
          <div class="text-[#FFA239] text-[10px] font-bold uppercase mb-1 tracking-wider">Upload (Current)</div>
          <div class="text-3xl font-bold text-white tracking-tight font-mono">
            {{ currentUpload.toFixed(2) }} <span class="text-sm font-medium text-gray-500 uppercase">Mbps</span>
          </div>
        </div>
        <div class="flex gap-4 text-right">
          <div>
            <div class="text-gray-500 text-[10px] font-bold uppercase mb-1">Max</div>
            <div class="text-sm font-mono text-white font-semibold">{{ maxUpload.toFixed(2) }}</div>
          </div>
          <div>
            <div class="text-gray-500 text-[10px] font-bold uppercase mb-1">Avg</div>
            <div class="text-sm font-mono text-white font-semibold">{{ avgUpload.toFixed(2) }}</div>
          </div>
          <div>
            <div class="text-gray-500 text-[10px] font-bold uppercase mb-1">Min</div>
            <div class="text-sm font-mono text-white font-semibold">{{ minUpload.toFixed(2) }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="min-h-[300px] relative z-10">
      <apexchart 
        type="area" 
        height="300" 
        :options="chartOptions" 
        :series="chartSeries"
      />
    </div>

  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  monitoredInterfaces: { type: Array, required: true },
  widgetId: { type: [Number, String], required: true },
  widgetConfig: { type: Object, default: () => ({}) }, // NEW: Receive config from parent
  canDelete: { type: Boolean, default: true } 
})

const emit = defineEmits(['remove', 'update-widget-config'])

// --- DROPDOWN LOGIC ---
const showDeleteDropdown = ref(false)

const confirmDelete = () => {
  showDeleteDropdown.value = false
  emit('remove')
}

// ==========================================
// TRACKING AND UI STATE (Initialized from Props)
// ==========================================
const selectedGraphPort = ref(props.widgetConfig.portId || null)
const selectedDuration = ref(props.widgetConfig.duration || '1h')
const isExpanded = ref(props.widgetConfig.isExpanded || false)
const startDateTime = ref(props.widgetConfig.startDateTime || '')
const endDateTime = ref(props.widgetConfig.endDateTime || '')

// Emit changes to the parent whenever local state changes
watch([selectedGraphPort, selectedDuration, isExpanded, startDateTime, endDateTime], () => {
  emit('update-widget-config', {
    portId: selectedGraphPort.value,
    duration: selectedDuration.value,
    isExpanded: isExpanded.value,
    startDateTime: startDateTime.value,
    endDateTime: endDateTime.value
  })
})

const toggleSize = () => {
  isExpanded.value = !isExpanded.value
  setTimeout(() => window.dispatchEvent(new Event('resize')), 100)
}

const timestamps = ref([])
const downloadData = ref([])
const uploadData = ref([])

let pollingInterval = null

// ==========================================
// CALCULATIONS
// ==========================================
const currentDownload = computed(() => {
  if (downloadData.value.length === 0) return 0
  return downloadData.value[downloadData.value.length - 1] || 0
})

const maxDownload = computed(() => {
  const valid = downloadData.value.filter(n => n !== null)
  return valid.length ? Math.max(...valid) : 0
})
const avgDownload = computed(() => {
  const valid = downloadData.value.filter(n => n !== null)
  return valid.length ? (valid.reduce((a, b) => a + b, 0) / valid.length) : 0
})
const minDownload = computed(() => {
  const valid = downloadData.value.filter(n => n !== null && n > 0)
  return valid.length ? Math.min(...valid) : 0
})

const currentUpload = computed(() => {
  if (uploadData.value.length === 0) return 0
  return uploadData.value[uploadData.value.length - 1] || 0
})

const maxUpload = computed(() => {
  const valid = uploadData.value.filter(n => n !== null)
  return valid.length ? Math.max(...valid) : 0
})
const avgUpload = computed(() => {
  const valid = uploadData.value.filter(n => n !== null)
  return valid.length ? (valid.reduce((a, b) => a + b, 0) / valid.length) : 0
})
const minUpload = computed(() => {
  const valid = uploadData.value.filter(n => n !== null && n > 0)
  return valid.length ? Math.min(...valid) : 0
})

const staticChartId = `widget-${Math.random().toString(36).substring(2, 9)}`

const chartOptions = computed(() => ({
  chart: {
    id: staticChartId,
    background: 'transparent',
    toolbar: { show: false },
    animations: { enabled: true, easing: 'linear', dynamicAnimation: { speed: 1000 } },
    zoom: { enabled: false }
  },
  colors: ['#FF5FCF', '#FFA239'],
  theme: { mode: 'dark' },
  stroke: { curve: 'smooth', width: 2 },
  fill: {
    type: 'gradient',
    gradient: { shadeIntensity: 1, opacityFrom: 0.45, opacityTo: 0.05, stops: [0, 100] }
  },
  grid: { 
    borderColor: '#ffffff10',
    strokeDashArray: 4,
    xaxis: { lines: { show: true } },
    yaxis: { lines: { show: true } }
  },
  dataLabels: { enabled: false },
  xaxis: { 
    categories: timestamps.value, 
    tickAmount: 6,
    labels: { style: { colors: '#EEEEEE60' } } 
  },
  yaxis: { 
    labels: { style: { colors: '#EEEEEE60' } }, 
    title: { text: 'Speed (Mbps)', style: { color: '#EEEEEE60' } },
    min: 0
  },
  legend: { labels: { colors: '#EEEEEE' } }
}))

const chartSeries = computed(() => [
  { name: 'Download', data: downloadData.value },
  { name: 'Upload', data: uploadData.value }
])

// ==========================================
// DATA FETCHING
// ==========================================
const clearGraph = () => {
  timestamps.value = []
  downloadData.value = []
  uploadData.value = []
}

const fetchGraphData = async () => {
  if (!selectedGraphPort.value) return
  if (selectedDuration.value === 'custom' && (!startDateTime.value || !endDateTime.value)) return

  try {
    let targetUrl = `/api/metrics/${selectedGraphPort.value}?duration=${selectedDuration.value}`
    let gapThresholdMs = 125000 

    if (selectedDuration.value === 'custom') {
      targetUrl = `/api/metrics/${selectedGraphPort.value}?duration=custom&start=${startDateTime.value}&end=${endDateTime.value}`
      const startT = new Date(startDateTime.value).getTime()
      const endT = new Date(endDateTime.value).getTime()
      if (endT - startT > 24 * 60 * 60 * 1000) gapThresholdMs = 3600000 + 60000
    }

    const token = localStorage.getItem('hermes_token')
    const response = await fetch(targetUrl, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    
    const data = await response.json() 

    if (!Array.isArray(data)) return

    const newTimestamps = []
    const newDownload = []
    const newUpload = []

    data.forEach((item, index) => {
      const currentT = new Date(item.timestamp).getTime()
      if (index > 0) {
        const prevT = new Date(data[index - 1].timestamp).getTime()
        if (currentT - prevT > gapThresholdMs) {
          const gapTime = new Date(prevT + 30000)
          newTimestamps.push(gapTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }))
          newDownload.push(null)
          newUpload.push(null)
        }
      }

      const dateObj = new Date(item.timestamp)
      let timeStr = ''
      if (selectedDuration.value === 'custom') {
        timeStr = dateObj.toLocaleDateString([], { month: 'short', day: 'numeric' }) + ', ' + 
               dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      } else {
        timeStr = dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
      }

      newTimestamps.push(timeStr)
      newDownload.push(Number((item.download_mbps || 0).toFixed(2)))
      newUpload.push(Number((item.upload_mbps || 0).toFixed(2)))
    })

    timestamps.value = newTimestamps
    downloadData.value = newDownload
    uploadData.value = newUpload

  } catch (error) {
    console.error("Failed to fetch graph metrics:", error)
  }
}

const pollData = () => {
  if (selectedDuration.value !== 'custom') fetchGraphData()
}

// ==========================================
// WATCHERS
// ==========================================
watch(selectedDuration, (newVal) => {
  if (newVal !== 'custom') {
    clearGraph()
    fetchGraphData()
  }
})

watch(() => props.monitoredInterfaces, (newInterfaces) => {
  const safeInterfaces = newInterfaces || []
  if (safeInterfaces.length > 0) {
    // If the saved port doesn't exist in the new list, default to the first one
    if (!selectedGraphPort.value || !safeInterfaces.find(i => i.id === selectedGraphPort.value)) {
      selectedGraphPort.value = safeInterfaces[0].id
    }
    fetchGraphData()
  } else {
    selectedGraphPort.value = null
    clearGraph()
  }
}, { immediate: true })

watch([startDateTime, endDateTime], () => {
  if (selectedDuration.value === 'custom') {
    clearGraph()
    fetchGraphData()
  }
})

// ==========================================
// LIFECYCLE
// ==========================================
onMounted(() => {
  pollingInterval = setInterval(pollData, 60000) 
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
})
</script>