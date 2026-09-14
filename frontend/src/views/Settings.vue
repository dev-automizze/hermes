<template>
  <main class="flex-1 p-6">
    <div class="max-w-7xl mx-auto">
      
      <div class="mb-8 flex justify-between items-center border-b border-white/10 pb-6">
        <div>
          <h2 class="text-3xl font-semibold">Settings</h2>
          <p class="text-gray-400 mt-1">Configure your core network monitoring engines.</p>
        </div>
        
        <button v-if="activeTab === 'hosts' && currentUserRole !== 'guest'" @click="resetFormForNew" class="px-4 py-2 bg-emerald-600/20 text-emerald-400 border border-emerald-500/30 rounded-lg font-medium hover:bg-emerald-600/30 transition text-sm shadow-lg shadow-emerald-500/10 cursor-pointer">
          + Add New Host
        </button>
        <button v-if="activeTab === 'alerts' && isEditingAlerts && currentUserRole !== 'guest'" @click="addAlertRule" class="px-4 py-2 bg-emerald-600/20 text-emerald-400 border border-emerald-500/30 rounded-lg font-medium hover:bg-emerald-600/30 transition text-sm shadow-lg shadow-emerald-500/10 cursor-pointer">
          + Add New Rule
        </button>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
        
        <div class="lg:col-span-3 space-y-2">
          <h3 class="text-xs font-bold uppercase text-gray-500 tracking-wider px-3 mb-4">Configuration Menu</h3>
          
          <button @click="activeTab = 'hosts'" 
                  class="w-full text-left px-4 py-3 rounded-xl font-medium transition flex items-center gap-3 cursor-pointer"
                  :class="activeTab === 'hosts' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'text-gray-400 hover:bg-white/5 hover:text-gray-200 border border-transparent'">
            <span class="text-lg">🖥️</span> Host Management
          </button>

          <button @click="activeTab = 'alerts'" 
                  class="w-full text-left px-4 py-3 rounded-xl font-medium transition flex items-center gap-3 cursor-pointer"
                  :class="activeTab === 'alerts' ? 'bg-red-500/10 text-red-400 border border-red-500/20' : 'text-gray-400 hover:bg-white/5 hover:text-gray-200 border border-transparent'">
            <span class="text-lg">🚨</span> Alert Rules
          </button>

          <button @click="activeTab = 'telegram'" 
                  class="w-full text-left px-4 py-3 rounded-xl font-medium transition flex items-center gap-3 cursor-pointer"
                  :class="activeTab === 'telegram' ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20' : 'text-gray-400 hover:bg-white/5 hover:text-gray-200 border border-transparent'">
            <span class="text-lg">✈️</span> Telegram Integration
          </button>

          <button @click="activeTab = 'database'" 
                  class="w-full text-left px-4 py-3 rounded-xl font-medium transition flex items-center gap-3 mt-4 cursor-pointer"
                  :class="activeTab === 'database' ? 'bg-purple-500/10 text-purple-400 border border-purple-500/20' : 'text-gray-400 hover:bg-white/5 hover:text-gray-200 border border-transparent'">
            <span class="text-lg">💾</span> Database & History
          </button>

          <button @click="activeTab = 'users'" 
                  class="w-full text-left px-4 py-3 rounded-xl font-medium transition flex items-center gap-3 mt-4 border-t border-white/5 pt-4 cursor-pointer"
                  :class="activeTab === 'users' ? 'bg-orange-500/10 text-orange-400 border border-orange-500/20' : 'text-gray-400 hover:bg-white/5 hover:text-gray-200 border border-transparent'">
            <span class="text-lg">👥</span> User Access
          </button>
        </div>

        <div class="lg:col-span-9">
          
          <div v-if="activeTab === 'hosts'" class="grid grid-cols-1 xl:grid-cols-12 gap-6 animate-fade-in">
              <div class="xl:col-span-4 space-y-3">
              <h3 class="text-xs font-bold uppercase text-gray-400 tracking-wider px-1">Monitored Devices</h3>
              <div v-if="hostsList.length === 0" class="text-sm text-gray-500 p-4 border border-dashed border-white/10 rounded-xl text-center">
                No hosts added yet.
              </div>
              <div v-for="h in hostsList" :key="h.id" @click="selectHost(h)"
                   class="p-4 border rounded-xl cursor-pointer transition text-left backdrop-blur-sm"
                   :class="selectedHostId === h.id ? 'bg-emerald-500/10 border-emerald-500 text-white shadow-[0_0_15px_rgba(16,185,129,0.1)]' : 'bg-white/5 border-white/10 hover:border-white/20 text-gray-300'">
                <div class="flex justify-between items-center">
                  <span class="font-semibold block truncate max-w-[140px]">{{ h.name }}</span>
                  <span class="text-xs font-bold uppercase px-2 py-0.5 rounded-full" 
                        :class="h.enabled ? 'bg-emerald-500/10 text-emerald-400' : 'bg-gray-500/20 text-gray-400'">
                    {{ h.enabled ? 'Active' : 'Disabled' }}
                  </span>
                </div>
                <span class="text-xs text-gray-400 block mt-1 font-mono">{{ h.ip }}</span>
              </div>
            </div>

            <div class="xl:col-span-8 space-y-6">
              <div class="bg-[#121826] border border-white/5 rounded-xl p-6 shadow-xl">
                <div class="flex justify-between items-center mb-6 border-b border-white/5 pb-4">
                  <div class="flex items-center gap-4">
                    <h3 class="text-xl font-semibold text-white">
                      {{ selectedHostId ? 'Host Details' : 'Register New Host' }}
                    </h3>
                    <button v-if="!isEditingHost && selectedHostId && currentUserRole !== 'guest'" @click="isEditingHost = true" 
                            class="text-xs bg-white/10 hover:bg-white/20 text-gray-300 px-3 py-1 rounded-full transition cursor-pointer">
                      ✏️ Edit
                    </button>
                  </div>
                  
                  <div class="flex items-center gap-2 px-3 py-1 rounded-full border bg-black/40"
                       :class="status === 'online' ? 'border-emerald-500/30 text-emerald-400' : 
                               status === 'offline' ? 'border-red-500/30 text-red-400' : 
                               'border-gray-500/30 text-gray-400'">
                    <span class="relative flex h-2.5 w-2.5">
                      <span v-if="status === 'online'" class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                      <span class="relative inline-flex rounded-full h-2.5 w-2.5" :class="status === 'online' ? 'bg-emerald-500' : status === 'offline' ? 'bg-red-500' : 'bg-gray-500'"></span>
                    </span>
                    <span class="text-xs font-bold uppercase tracking-wider">
                      {{ status === 'online' ? 'Connected' : status === 'offline' ? 'Failed' : 'Not Tested' }}
                    </span>
                  </div>
                </div>

                <form @submit.prevent="saveHost" class="space-y-5">
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                    <div>
                      <label class="block text-xs font-bold uppercase text-gray-400 mb-1">Device Name</label>
                      <input type="text" v-model="host.name" required placeholder="e.g. Mikrotik Home" 
                             :disabled="(selectedHostId !== null && !isEditingHost) || currentUserRole === 'guest'"
                             class="w-full bg-black/40 border border-gray-700 rounded-lg p-2.5 text-white focus:border-emerald-500 outline-none transition disabled:opacity-60 disabled:cursor-not-allowed">
                    </div>
                    <div>
                      <label class="block text-xs font-bold uppercase text-gray-400 mb-1">IP Address</label>
                      <input type="text" v-model="host.ip" required placeholder="172.10.10.1" 
                             :disabled="(selectedHostId !== null && !isEditingHost) || currentUserRole === 'guest'"
                             class="w-full bg-black/40 border border-gray-700 rounded-lg p-2.5 text-white focus:border-emerald-500 outline-none transition disabled:opacity-60 disabled:cursor-not-allowed">
                    </div>
                  </div>

                  <div>
                    <label class="block text-xs font-bold uppercase text-gray-400 mb-1">SNMP Community String</label>
                    <input type="text" v-model="host.community" required placeholder="public" 
                           :disabled="(selectedHostId !== null && !isEditingHost) || currentUserRole === 'guest'"
                           class="w-full bg-black/40 border border-gray-700 rounded-lg p-2.5 text-white focus:border-emerald-500 outline-none transition disabled:opacity-60 disabled:cursor-not-allowed">
                    
                    <p class="text-[10px] text-gray-500 mt-1.5">
                      <span class="text-emerald-400 font-bold">Note:</span> Hermes uses <strong>SNMP v2c</strong>. Ensure v2/v2c is enabled on your target device.
                    </p>
                  </div>

                  <div class="flex items-center gap-3 pt-2" :class="{'opacity-60 pointer-events-none': !isEditingHost || currentUserRole === 'guest'}">
                    <input type="checkbox" id="enabled" v-model="host.enabled" class="w-4 h-4 accent-emerald-500 bg-gray-800 border-gray-600 rounded">
                    <label for="enabled" class="text-sm font-medium text-gray-300">Enable Polling Engine for this Host</label>
                  </div>

                  <div class="pt-4 flex justify-end gap-3 border-t border-white/5 mt-4">
                    <button type="button" @click="testConnection" class="px-4 py-2 bg-gray-700/50 hover:bg-gray-600 text-white rounded-lg font-medium transition cursor-pointer">
                      Test Connection
                    </button>
                    <button v-if="isEditingHost && selectedHostId" type="button" @click="cancelHostEdit" class="px-4 py-2 bg-red-900/20 hover:bg-red-900/40 text-red-400 border border-red-500/30 rounded-lg font-medium transition cursor-pointer">
                      Cancel
                    </button>
                    <button v-if="isEditingHost && selectedHostId && currentUserRole !== 'guest'" type="button" @click="deleteHost" class="px-4 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg font-medium transition shadow-lg shadow-red-500/20 cursor-pointer">
                      Delete Host
                    </button>
                    <button v-if="isEditingHost && currentUserRole !== 'guest'" type="submit" class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-medium transition shadow-lg shadow-emerald-500/20 cursor-pointer">
                      {{ selectedHostId ? 'Update Device' : 'Save Host' }}
                    </button>
                  </div>
                </form>

                <div v-if="selectedHostId" class="mt-10 pt-8 border-t border-white/10">
                  <div class="flex justify-between items-center mb-6">
                    <div>
                      <h3 class="text-xl font-semibold text-white">Network Interfaces</h3>
                      <p class="text-xs text-gray-400 mt-1">Toggle which ports should be actively graphed on the Dashboard.</p>
                    </div>
                    <button @click="fetchInterfaces(selectedHostId)" type="button" 
                            class="px-3 py-2 bg-white/5 hover:bg-white/10 text-gray-300 text-xs font-bold uppercase tracking-wider rounded-lg border border-white/10 transition-colors shadow-sm cursor-pointer">
                      Refresh Ports
                    </button>
                  </div>

                  <div v-if="isLoadingInterfaces" class="flex justify-center items-center py-8">
                    <div class="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin"></div>
                  </div>
                  
                  <div v-else-if="hostInterfaces.length === 0" class="bg-black/20 border border-white/5 p-8 rounded-xl text-center flex flex-col items-center">
                    <span class="text-gray-400 text-sm font-medium">No interfaces found for this device.</span>
                    <span class="text-gray-600 text-xs mt-1">Make sure SNMP is configured correctly and the device is actively reachable.</span>
                  </div>

                  <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar">
                    <div v-for="iface in hostInterfaces" :key="iface.id" 
                        class="flex items-center justify-between p-3.5 rounded-xl border transition-all duration-300"
                        :class="iface.monitoring ? 'bg-emerald-500/10 border-emerald-500/30 shadow-[0_0_15px_rgba(16,185,129,0.1)]' : 'bg-black/40 border-white/5 hover:border-white/10'">
                      
                      <div class="flex items-center gap-3 overflow-hidden">
                        <div :class="iface.monitoring ? 'text-emerald-400' : 'text-gray-600'">
                          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
                        </div>
                        <div class="truncate">
                          <div class="text-sm font-bold text-white truncate" :title="iface.name">{{ iface.name }}</div>
                        </div>
                      </div>

                      <label v-if="currentUserRole !== 'guest'" class="relative inline-flex items-center cursor-pointer ml-4 shrink-0">
                        <input type="checkbox" v-model="iface.monitoring" @change="toggleInterface(iface)" class="sr-only peer">
                        <div class="w-9 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-emerald-500"></div>
                      </label>
                      <span v-else class="text-[10px] text-gray-500 uppercase">{{ iface.monitoring ? 'Monitored' : 'Ignored' }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="activeTab === 'alerts'" class="animate-fade-in space-y-6">
              <div class="bg-[#121826] border border-white/5 rounded-xl p-6 shadow-xl">
              <div class="mb-6 border-b border-white/5 pb-4 flex justify-between items-center">
                <div class="flex items-center gap-4">
                  <h3 class="text-2xl font-semibold text-white">Global Alert Engine</h3>
                  <button v-if="!isEditingAlerts && currentUserRole !== 'guest'" @click="isEditingAlerts = true" 
                          class="text-xs bg-white/10 hover:bg-white/20 text-gray-300 px-3 py-1 rounded-full transition cursor-pointer">
                    ✏️ Edit Rules
                  </button>
                </div>
              </div>

              <div v-if="globalAlerts.length === 0" class="text-center py-12 text-gray-500 border border-dashed border-white/10 rounded-xl">
                No alert rules configured yet. Click "Edit Rules" to add one.
              </div>

              <div class="space-y-8">
                <div v-for="(rule, index) in globalAlerts" :key="index" class="bg-black/20 border border-gray-800 rounded-xl p-5 relative shadow-inner">
                  
                  <div class="grid grid-cols-1 lg:grid-cols-12 gap-4 mb-5 border-b border-gray-800/50 pb-5 items-end">
                    
                    <div class="lg:col-span-1 flex items-center h-full pb-3 pl-1">
                      <input type="checkbox" v-model="rule.enabled" :disabled="!isEditingAlerts" class="w-5 h-5 accent-red-500 bg-gray-800 border-gray-600 rounded cursor-pointer disabled:cursor-not-allowed">
                    </div>
                    
                    <div class="lg:col-span-4">
                      <label class="block text-xs font-bold uppercase text-gray-400 mb-1.5">Target Host</label>
                      <select v-model="rule.host_id" @change="updatePortsForRule(rule)" :disabled="!isEditingAlerts" class="w-full bg-black/40 border border-gray-700 rounded-lg p-2.5 text-white focus:border-red-500 outline-none disabled:opacity-60">
                        <option value="" disabled>Select Host...</option>
                        <option v-for="h in hostsList" :key="h.id" :value="h.id">{{ h.name }}</option>
                      </select>
                    </div>

                <div class="lg:col-span-4">
                  <label class="block text-[10px] font-bold uppercase text-gray-500 tracking-wider mb-1">Target Interface(s)</label>
                  
                  <div class="flex flex-col gap-2">
                    <div v-for="(p, pIndex) in rule.ports" :key="pIndex" class="flex items-center gap-2">
                      
                      <select v-model="rule.ports[pIndex]" class="bg-transparent text-white text-sm border border-white/10 rounded-lg px-3 py-2 focus:outline-none focus:border-[#FF5FCF] w-full">
                        <option disabled value="" v-if="!portCache[rule.host_id] || portCache[rule.host_id].length === 0" class="text-gray-500 italic">
                          No interfaces found
                        </option>
                        <option v-for="iface in portCache[rule.host_id]" :key="iface.id" :value="iface.name" class="bg-gray-800 text-white">
                          {{ iface.name }}
                        </option>
                      </select>
                      
                      <button v-if="rule.ports.length > 1" @click="rule.ports.splice(pIndex, 1)" type="button" class="p-2 bg-red-500/10 text-red-400 hover:bg-red-500/30 rounded-lg border border-red-500/20 transition cursor-pointer" title="Remove Interface">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" /></svg>
                      </button>

                      <button v-if="pIndex === rule.ports.length - 1" @click="rule.ports.push('')" type="button" class="p-2 bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/30 rounded-lg border border-emerald-500/20 transition cursor-pointer" title="Add Interface to Sum">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
                      </button>

                    </div>
                  </div>
                </div>

                    <div class="lg:col-span-2">
                      <label class="block text-xs font-bold uppercase text-gray-400 mb-1.5">Direction</label>
                      <select v-model="rule.direction" :disabled="!isEditingAlerts" class="w-full bg-black/40 border border-gray-700 rounded-lg p-2.5 text-white focus:border-red-500 outline-none disabled:opacity-60">
                        <option value="download">Download</option>
                        <option value="upload">Upload</option>
                      </select>
                    </div>

                    <div class="lg:col-span-1 flex justify-end pb-1">
                      <button v-if="!rule.is_default && isEditingAlerts" @click="removeAlertRule(index)" class="text-gray-500 hover:text-red-400 p-2 rounded-lg hover:bg-red-900/10 transition cursor-pointer" title="Delete Rule">
                        🗑️
                      </button>
                    </div>
                  </div>

                  <div class="grid grid-cols-1 xl:grid-cols-2 gap-6">
                    <div class="space-y-4 bg-red-950/10 p-5 rounded-lg border border-red-900/20">
                      <div class="flex items-center flex-wrap gap-3">
                        <h4 class="text-sm font-bold text-red-400 flex items-center gap-2 mr-2">
                          <span class="w-2 h-2 rounded-full bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.8)]"></span> Alert Trigger
                        </h4>
                        <span class="text-gray-400 text-sm whitespace-nowrap">If speed ></span>
                        <input type="number" v-model="rule.alert_threshold_mbps" :disabled="!isEditingAlerts" class="w-20 bg-black/40 border border-gray-700 rounded-lg p-1.5 text-white text-center outline-none focus:border-red-500 disabled:opacity-60">
                        <span class="text-gray-400 text-sm whitespace-nowrap">Mbps for</span>
                        <input type="number" v-model="rule.alert_duration_mins" :disabled="!isEditingAlerts" class="w-16 bg-black/40 border border-gray-700 rounded-lg p-1.5 text-white text-center outline-none focus:border-red-500 disabled:opacity-60">
                        <span class="text-gray-400 text-sm whitespace-nowrap">mins</span>
                      </div>
                      <textarea v-model="rule.alert_template" :disabled="!isEditingAlerts" rows="2" class="w-full bg-black/40 border border-gray-700 rounded-lg p-3 text-gray-300 text-sm focus:border-red-500 outline-none transition disabled:opacity-60 resize-y"></textarea>
                    </div>

                    <div class="space-y-4 bg-emerald-950/10 p-5 rounded-lg border border-emerald-900/20">
                      <div class="flex items-center flex-wrap gap-3">
                        <h4 class="text-sm font-bold text-emerald-400 flex items-center gap-2 mr-2">
                          <span class="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)]"></span> Recovery Trigger
                        </h4>
                        <span class="text-gray-400 text-sm whitespace-nowrap">Drops &lt;</span>
                        <input type="number" v-model="rule.recovery_threshold_mbps" :disabled="!isEditingAlerts" class="w-20 bg-black/40 border border-gray-700 rounded-lg p-1.5 text-white text-center outline-none focus:border-emerald-500 disabled:opacity-60">
                        <span class="text-gray-400 text-sm whitespace-nowrap">Mbps for</span>
                        <input type="number" v-model="rule.recovery_duration_mins" :disabled="!isEditingAlerts" class="w-16 bg-black/40 border border-gray-700 rounded-lg p-1.5 text-white text-center outline-none focus:border-emerald-500 disabled:opacity-60">
                        <span class="text-gray-400 text-sm whitespace-nowrap">mins</span>
                      </div>
                      <textarea v-model="rule.recovery_template" :disabled="!isEditingAlerts" rows="2" class="w-full bg-black/40 border border-gray-700 rounded-lg p-3 text-gray-300 text-sm focus:border-emerald-500 outline-none transition disabled:opacity-60 resize-y"></textarea>
                    </div>
                  </div>
                </div>

                <div v-if="isEditingAlerts" class="pt-6 border-t border-white/5 flex justify-end gap-3 mt-6">
                  <button @click="cancelAlertsEdit" class="px-6 py-2.5 bg-red-900/20 hover:bg-red-900/40 text-red-400 border border-red-500/30 rounded-lg font-medium transition text-sm cursor-pointer">
                    Cancel Changes
                  </button>
                  <button @click="saveAlerts" class="px-6 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-medium transition shadow-lg shadow-emerald-500/20 text-sm cursor-pointer">
                    Save All Rules
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div v-if="activeTab === 'telegram'" class="animate-fade-in space-y-6">
             <div class="bg-[#121826] border border-white/5 rounded-xl p-8 shadow-xl">
              <div class="mb-8 border-b border-white/5 pb-4 flex items-center gap-4">
                <h3 class="text-2xl font-semibold text-white">Telegram Global Setup</h3>
                <button v-if="!isEditingTelegram && currentUserRole !== 'guest'" @click="isEditingTelegram = true" 
                        class="text-xs bg-white/10 hover:bg-white/20 text-gray-300 px-3 py-1 rounded-full transition cursor-pointer">
                  ✏️ Edit
                </button>
              </div>
              
              <div v-if="currentUserRole === 'guest'" class="flex flex-col items-center justify-center p-12 border border-white/5 bg-black/20 rounded-xl">
                <span class="text-4xl mb-4">🔒</span>
                <h4 class="text-lg font-bold text-white mb-1">Access Restricted</h4>
                <p class="text-sm text-gray-400">Your role does not have permission to view or manage Telegram Bot secrets.</p>
              </div>

              <div v-else class="space-y-6 max-w-3xl">
                <div class="flex items-center gap-3 p-4 bg-blue-950/20 border border-blue-900/30 rounded-xl" :class="{'opacity-60 pointer-events-none': !isEditingTelegram}">
                  <input type="checkbox" id="telegramEnabled" v-model="telegram.enabled" class="w-5 h-5 accent-blue-500 bg-gray-800 border-gray-600 rounded">
                  <label for="telegramEnabled" class="text-base font-medium text-gray-200">Enable Global Telegram Alerts</label>
                </div>

                <div class="space-y-5">
                  <div>
                    <label class="block text-xs font-bold uppercase text-gray-400 mb-2">Bot Token</label>
                    <input type="text" v-model="telegram.bot_token" 
                           :disabled="!isEditingTelegram || !telegram.enabled"
                           placeholder="e.g. 123456789:ABCdefGhIJKlmNoPQRstuVwxyZ" 
                           class="w-full bg-black/40 border border-gray-700 rounded-lg p-3 text-white focus:border-blue-500 outline-none transition disabled:opacity-60 disabled:cursor-not-allowed">
                  </div>
                  <div>
                    <label class="block text-xs font-bold uppercase text-gray-400 mb-2">Chat ID</label>
                    <input type="text" v-model="telegram.chat_id" 
                           :disabled="!isEditingTelegram || !telegram.enabled"
                           placeholder="e.g. 987654321" 
                           class="w-full bg-black/40 border border-gray-700 rounded-lg p-3 text-white focus:border-blue-500 outline-none transition disabled:opacity-60 disabled:cursor-not-allowed">
                  </div>
                </div>

                <div class="pt-8 flex justify-end gap-3 border-t border-white/5 mt-8">
                  <button @click="testTelegramConnection" 
                          :disabled="!telegram.enabled || !telegram.bot_token || !telegram.chat_id"
                          class="px-6 py-2.5 bg-gray-700/50 hover:bg-gray-600 text-white rounded-lg font-medium transition disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer">
                    Test Connection
                  </button>
                  <button v-if="isEditingTelegram" @click="cancelTelegramEdit" class="px-6 py-2.5 bg-red-900/20 hover:bg-red-900/40 text-red-400 border border-red-500/30 rounded-lg font-medium transition cursor-pointer">
                    Cancel
                  </button>
                  <button v-if="isEditingTelegram" @click="saveTelegramSettings" 
                          class="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-lg font-medium transition shadow-lg shadow-blue-500/20 cursor-pointer">
                    Save Settings
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div v-if="activeTab === 'database'" class="animate-fade-in space-y-6">
            <div class="bg-[#121826] border border-white/5 rounded-xl p-8 shadow-xl">
              <div class="mb-8 border-b border-white/5 pb-4 flex items-center gap-4">
                <h3 class="text-2xl font-semibold text-white">Database & History</h3>
                <button v-if="!isEditingDb && currentUserRole !== 'guest'" @click="isEditingDb = true" 
                        class="text-xs bg-white/10 hover:bg-white/20 text-gray-300 px-3 py-1 rounded-full transition cursor-pointer">
                  ✏️ Edit
                </button>
              </div>
              
              <div class="space-y-8 max-w-3xl">
                <div class="bg-black/20 border border-gray-800 rounded-xl p-6">
                  <h4 class="text-lg font-medium text-white mb-2">Data Retention Policy</h4>
                  <p class="text-sm text-gray-400 mb-6">Select how long graph data should be kept. The background housekeeper runs once daily to sweep old metrics.</p>
                  
                  <div class="flex items-center gap-4">
                    <select v-model.number="dbSettings.retention_days" :disabled="!isEditingDb || currentUserRole === 'guest'" class="w-48 bg-black/40 border border-gray-700 rounded-lg p-3 text-white focus:border-purple-500 outline-none disabled:opacity-60">
                      <option value="1">1 Day</option>
                      <option value="3">3 Days</option>
                      <option value="7">7 Days</option>
                      <option value="14">14 Days</option>
                      <option value="30">30 Days</option>
                      <option value="90">90 Days</option>
                      <option value="180">180 Days</option>
                      <option value="365">1 Year</option>
                    </select>
                    
                    <div v-if="isEditingDb" class="flex gap-2">
                      <button @click="saveDbSettings" class="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-lg font-medium transition shadow-lg shadow-purple-500/20 cursor-pointer">Save</button>
                      <button @click="cancelDbEdit" class="px-4 py-2 bg-red-900/20 hover:bg-red-900/40 text-red-400 border border-red-500/30 rounded-lg font-medium transition cursor-pointer">Cancel</button>
                    </div>
                  </div>
                </div>

                <div class="bg-red-950/10 border border-red-900/20 rounded-xl p-6">
                  <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div>
                      <h4 class="text-lg font-medium text-red-400 mb-1">Manual Purge</h4>
                      <p class="text-sm text-gray-400">Instantly delete graph data older than your current retention policy.</p>
                    </div>
                    <button @click="triggerManualPurge" :disabled="isPurging || currentUserRole === 'guest'" class="px-6 py-2.5 bg-red-600/20 hover:bg-red-600/40 text-red-400 border border-red-500/50 rounded-lg font-medium transition disabled:opacity-50 min-w-[140px] cursor-pointer">
                      {{ isPurging ? 'Purging...' : 'Purge Now 🧹' }}
                    </button>
                  </div>
                </div>

              </div>
            </div>
          </div>

          <div v-if="activeTab === 'users'" class="animate-fade-in space-y-6">
            <div class="bg-[#121826] border border-white/5 rounded-xl p-6 shadow-xl">
              <div class="flex justify-between items-center mb-6 border-b border-white/5 pb-4">
                <div>
                  <h3 class="text-xl font-semibold text-white">Access Management</h3>
                  <p class="text-sm text-gray-400 mt-1">Control who can access Hermes.</p>
                </div>
                <button v-if="currentUserRole !== 'guest'" @click="openUserModal()" class="px-4 py-2 bg-emerald-600/20 text-emerald-400 border border-emerald-500/30 rounded-lg text-sm font-medium hover:bg-emerald-600/30 transition cursor-pointer">
                  + Create User
                </button>
              </div>

              <div class="overflow-x-auto rounded-lg border border-white/5">
                <table class="w-full text-left text-sm text-gray-400">
                  <thead class="bg-black/40 text-xs uppercase text-gray-500 border-b border-white/5">
                    <tr>
                      <th class="px-4 py-3">ID</th>
                      <th class="px-4 py-3">Username</th>
                      <th class="px-4 py-3">Role</th>
                      <th class="px-4 py-3 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="user in usersList" :key="user.id" class="border-b border-white/5 hover:bg-white/5 transition-colors">
                      <td class="px-4 py-3 font-mono text-gray-500">#{{ user.id }}</td>
                      <td class="px-4 py-3 font-medium text-gray-200">
                        {{ user.username }}
                        <span v-if="user.id === 1" class="ml-2 text-[10px] bg-red-500/20 text-red-400 px-2 py-0.5 rounded uppercase font-bold tracking-wider">Sudo</span>
                        <span v-if="currentUsername === user.username" class="ml-2 text-[10px] bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded uppercase font-bold tracking-wider">You</span>
                      </td>
                      <td class="px-4 py-3 uppercase text-xs tracking-wider font-semibold" :class="user.role === 'admin' ? 'text-emerald-400' : 'text-gray-400'">{{ user.role }}</td>
                      <td class="px-4 py-3 text-right">
                        <button v-if="canEditUser(user)" @click="openUserModal(user)" class="text-gray-400 hover:text-emerald-400 mx-3 transition cursor-pointer">✏️ Edit</button>
                        <button v-if="canDeleteUser(user)" @click="deleteUser(user.id)" class="text-gray-400 hover:text-red-400 transition cursor-pointer">🗑️</button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div v-if="showUserModal" class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
              <div class="bg-[#121826] border border-white/10 rounded-xl p-6 w-full max-w-md shadow-2xl">
                <h3 class="text-xl font-semibold text-white mb-4">{{ isEditingUser ? 'Edit User' : 'Create New User' }}</h3>
                
                <form @submit.prevent="saveUser" class="space-y-4">
                  <div>
                    <label class="block text-xs font-bold uppercase text-gray-400 mb-1">Username</label>
                    <input type="text" v-model="userForm.username" required :disabled="isEditingUser" class="w-full bg-black/40 border border-gray-700 rounded-lg p-2.5 text-white focus:border-emerald-500 outline-none transition disabled:opacity-50">
                  </div>
                  
                  <div>
                    <label class="block text-xs font-bold uppercase text-gray-400 mb-1">
                      Password <span v-if="isEditingUser" class="text-gray-500 font-normal lowercase">(Leave blank to keep current)</span>
                    </label>
                    <input type="password" v-model="userForm.password" :required="!isEditingUser" class="w-full bg-black/40 border border-gray-700 rounded-lg p-2.5 text-white focus:border-emerald-500 outline-none transition">
                  </div>

                  <div>
                    <label class="block text-xs font-bold uppercase text-gray-400 mb-1">Role</label>
                    <select v-model="userForm.role" :disabled="isEditingUser && userForm.id === 1" class="w-full bg-black/40 border border-gray-700 rounded-lg p-2.5 text-white focus:border-emerald-500 outline-none transition disabled:opacity-50">
                      <option value="admin">Administrator</option>
                      <option value="guest">Guest (Read-Only)</option>
                    </select>
                  </div>

                  <div class="pt-4 flex justify-end gap-3 border-t border-white/5 mt-4">
                    <button type="button" @click="showUserModal = false" class="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-lg font-medium transition cursor-pointer">Cancel</button>
                    <button type="submit" class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-medium transition shadow-lg shadow-emerald-500/20 cursor-pointer">Save User</button>
                  </div>
                </form>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>
  </main>
</template>

<script setup>
import { ref, onMounted } from 'vue'

// --- NAVIGATION STATE ---
const activeTab = ref('hosts') 

// --- USER CONTEXT & MANAGEMENT ---
const currentUserRole = ref('')
const currentUsername = ref('')
const usersList = ref([])
const showUserModal = ref(false)
const isEditingUser = ref(false)
const userForm = ref({ id: null, username: '', password: '', role: 'guest' })

// --- EDIT STATE LOCKS ---
const isEditingHost = ref(false)
const isEditingAlerts = ref(false)
const isEditingTelegram = ref(false)
const isEditingDb = ref(false)

const hostsList = ref([])
const selectedHostId = ref(null)
const status = ref('unknown')

const host = ref({
  id: null, 
  name: '',
  ip: '',
  community: '',
  enabled: true
})

// --- INTERFACE MONITORING STATE ---
const hostInterfaces = ref([])
const isLoadingInterfaces = ref(false)

// --- ALERT STATE ---
const globalAlerts = ref([])
const portCache = ref({}) 

// --- TELEGRAM & DATABASE STATE ---
const telegram = ref({ bot_token: '', chat_id: '', enabled: false })
const dbSettings = ref({ retention_days: 7 }) 
const isPurging = ref(false)

// ==========================================
// 1. AUTHORIZATION HEADERS 
// ==========================================
const getAuthHeaders = () => {
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${localStorage.getItem('hermes_token')}`
  }
}

const getAuthHeadersGet = () => {
  return {
    'Authorization': `Bearer ${localStorage.getItem('hermes_token')}`
  }
}

// ==========================================
// 2. USER SECURITY FUNCTIONS
// ==========================================
const loadUserContext = async () => {
  try {
    const res = await fetch('/api/auth/check', { headers: getAuthHeadersGet() })
    const data = await res.json()
    if (data.authenticated) {
      currentUserRole.value = data.role
      currentUsername.value = data.username
    }
  } catch (error) {
    console.error("Failed fetching context", error)
  }
}

const fetchUsers = async () => {
  if (currentUserRole.value === 'guest') return
  try {
    const res = await fetch('/api/users', { headers: getAuthHeadersGet() })
    if (res.ok) usersList.value = await res.json()
  } catch (e) { console.error("Failed fetching users", e) }
}

const openUserModal = (existingUser = null) => {
  if (existingUser) {
    isEditingUser.value = true
    userForm.value = { id: existingUser.id, username: existingUser.username, password: '', role: existingUser.role }
  } else {
    isEditingUser.value = false
    userForm.value = { id: null, username: '', password: '', role: 'guest' }
  }
  showUserModal.value = true
}

const saveUser = async () => {
  const method = isEditingUser.value ? 'PUT' : 'POST'
  const url = isEditingUser.value
    ? `/api/users/${userForm.value.id}`
    : '/api/users'

  try {
    const res = await fetch(url, { method, headers: getAuthHeaders(), body: JSON.stringify(userForm.value) })
    if (res.ok) {
      showUserModal.value = false
      fetchUsers()
    } else {
      const data = await res.json()
      alert(`Error: ${data.error}`)
    }
  } catch (e) { console.error("User save failed:", e) }
}

const deleteUser = async (id) => {
  if (!confirm("Are you sure you want to delete this user?")) return
  try {
    const res = await fetch(`/api/users/${id}`, { method: 'DELETE', headers: getAuthHeadersGet() })
    if (res.ok) fetchUsers()
  } catch (e) { console.error("Delete failed:", e) }
}

const canEditUser = (user) => {
  if (currentUserRole.value === 'guest') return false
  if (user.id === 1 && currentUsername.value !== 'admin') return false 
  return true
}

const canDeleteUser = (user) => {
  if (currentUserRole.value === 'guest') return false
  if (user.id === 1) return false 
  if (user.username === currentUsername.value) return false 
  return true
}

// ==========================================
// 3. HOSTS AND INTERFACES
// ==========================================
const fetchHosts = async () => {
  try {
    const apiUrl = '/api/hosts'
    const res = await fetch(apiUrl, { headers: getAuthHeadersGet() }) 
    if (res.ok) {
      hostsList.value = await res.json()
      if (hostsList.value.length > 0 && !selectedHostId.value) {
        selectHost(hostsList.value[0])
      }
    }
  } catch (err) { console.error("Failed fetching hosts:", err) }
}

const fetchInterfaces = async (hostId) => {
  if (!hostId) return
  isLoadingInterfaces.value = true
  try {
    const res = await fetch(`/api/interfaces?host_id=${hostId}`, { headers: getAuthHeadersGet() })
    const data = await res.json()
    hostInterfaces.value = Array.isArray(data) ? data : []
  } catch (err) {
    console.error("Failed fetching interfaces:", err)
    hostInterfaces.value = []
  } finally {
    isLoadingInterfaces.value = false
  }
}

const toggleInterface = async (iface) => {
  try {
    const res = await fetch(`/api/interfaces/${iface.id}`, {
      method: 'PUT', 
      headers: getAuthHeaders(),
      body: JSON.stringify({ ...iface, monitoring: iface.monitoring })
    })
    if (!res.ok) throw new Error("Failed to update database")
  } catch (err) {
    console.error("Failed toggling interface:", err)
    iface.monitoring = !iface.monitoring
    alert("❌ Failed to update monitoring status. Check connection.")
  }
}

const selectHost = (target) => {
  selectedHostId.value = target.id
  status.value = 'unknown'
  host.value = { ...target }
  isEditingHost.value = false
  fetchInterfaces(target.id)
}

const deleteHost = async () => {
  if (!confirm(`Are you sure you want to completely delete ${host.value.name}? This will remove all associated ports and data!`)) {
    return
  }

  try {
    const res = await fetch(`/api/hosts/${selectedHostId.value}`, { 
      method: 'DELETE',
      headers: getAuthHeaders() 
    })
    
    if (res.ok) {
      alert("✅ Host deleted successfully!")
      fetchHosts() // Refresh the sidebar list
      resetFormForNew() // Clear the form
    } else {
      const data = await res.json()
      alert(`❌ Failed to delete host: ${data.error || 'Unknown error'}`)
    }
  } catch (err) {
    console.error("Failed to delete host:", err)
    alert("❌ Network error connecting to backend.")
  }
}

const resetFormForNew = () => {
  selectedHostId.value = null
  status.value = 'unknown'
  host.value = { id: null, name: '', ip: '', community: '', enabled: true }
  isEditingHost.value = true
}

const cancelHostEdit = () => {
  const original = hostsList.value.find(h => h.id === selectedHostId.value)
  if (original) host.value = { ...original }
  isEditingHost.value = false
}

const testConnection = async () => { status.value = 'online' }

const saveHost = async () => {
  try {
    const apiUrl = '/api/hosts'
    const res = await fetch(apiUrl, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(host.value)
    })
    if (res.ok) {
      alert("Device settings updated successfully!")
      isEditingHost.value = false
      await fetchHosts()
    }
  } catch (err) { console.error(err) }
}

// ==========================================
// 4. ALERTS 
// ==========================================
const createDefaultRule = () => ({
  host_id: '', 
  ports: [''], // 👈 CHANGED: This is now an array!
  direction: 'download', 
  alert_threshold_mbps: 15, 
  alert_duration_mins: 5,
  alert_template: '🚨 ALERT: [DEVICE] port [PORT] download maxed out! Hit [SPEED] Mbps.',
  recovery_threshold_mbps: 5, 
  recovery_duration_mins: 5,
  recovery_template: '✅ RECOVERY: [DEVICE] port [PORT] has calmed down to [SPEED] Mbps.',
  enabled: true, 
  is_default: false
})

const fetchAllAlerts = async () => {
  try {
    const apiUrl = '/api/alerts'
    const res = await fetch(apiUrl, { headers: getAuthHeadersGet() }) 
    if (res.ok) {
      const data = await res.json()
      globalAlerts.value = data || []
      for (const rule of globalAlerts.value) {
        if (rule.host_id) await fetchPortsForHost(rule.host_id)
      }
    }
  } catch (err) { console.error("Failed fetching alerts:", err) }
}

const fetchPortsForHost = async (hostId) => {
  if (!hostId || portCache.value[hostId]) return; 
  try {
    const apiUrl = `/api/interfaces?host_id=${hostId}`
    const res = await fetch(apiUrl, { headers: getAuthHeadersGet() })
    if (res.ok) {
      const interfaces = await res.json()
      portCache.value[hostId] = interfaces || []
    }
  } catch (err) { console.error(`Failed fetching ports for host ${hostId}:`, err) }
}

const updatePortsForRule = async (rule) => {
  rule.ports = ['']
  if (rule.host_id) await fetchPortsForHost(rule.host_id)
}

const addAlertRule = () => { globalAlerts.value.push(createDefaultRule()) }
const removeAlertRule = (index) => { globalAlerts.value.splice(index, 1) }

const cancelAlertsEdit = () => {
  fetchAllAlerts() 
  isEditingAlerts.value = false
}

const saveAlerts = async () => {
  try {
    const apiUrl = '/api/alerts'
    const res = await fetch(apiUrl, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(globalAlerts.value)
    })
    if (res.ok) {
      alert("✅ Global Alert Rules saved to database successfully!")
      isEditingAlerts.value = false
      await fetchAllAlerts() 
    } else { alert("❌ Failed to save alert rules.") }
  } catch (err) { console.error("Failed saving alerts:", err) }
}

// ==========================================
// 5. TELEGRAM
// ==========================================
const fetchTelegramSettings = async () => {
  try {
    const apiUrl = '/api/settings/telegram'
    const res = await fetch(apiUrl, { headers: getAuthHeadersGet() })
    if (res.ok) telegram.value = await res.json()
  } catch (err) { console.error("Failed fetching telegram settings:", err) }
}

const cancelTelegramEdit = () => {
  fetchTelegramSettings() 
  isEditingTelegram.value = false
}

const saveTelegramSettings = async () => {
  try {
    const apiUrl = '/api/settings/telegram'
    const res = await fetch(apiUrl, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(telegram.value)
    })
    if (res.ok) {
      alert("✅ Telegram settings saved successfully!")
      isEditingTelegram.value = false
      await fetchTelegramSettings() 
    } else { alert("❌ Failed to save Telegram settings.") }
  } catch (err) { console.error("Failed saving telegram settings:", err) }
}

const testTelegramConnection = async () => {
  try {
    const apiUrl = '/api/settings/telegram/test'
    const res = await fetch(apiUrl, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(telegram.value)
    })
    if (res.ok) {
      alert("🚀 Test message sent! Check your phone.")
    } else {
      const errData = await res.json()
      alert(`❌ Test failed: ${errData.error || 'Check backend console logs'}`)
    }
  } catch (err) { console.error("Failed running telegram test:", err) }
}

// ==========================================
// 6. DATABASE SETTINGS
// ==========================================
const fetchDbSettings = async () => {
  try {
    const apiUrl = '/api/settings/database'
    const res = await fetch(apiUrl, { headers: getAuthHeadersGet() })
    if (res.ok) {
      const data = await res.json()
      if (data && data.retention_days) {
        dbSettings.value.retention_days = parseInt(data.retention_days, 10)
      }
    }
  } catch (err) { console.error("Failed fetching database settings:", err) }
}

const cancelDbEdit = () => {
  fetchDbSettings()
  isEditingDb.value = false
}

const saveDbSettings = async () => {
  try {
    const apiUrl = '/api/settings/database'
    const res = await fetch(apiUrl, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(dbSettings.value) 
    })
    if (res.ok) {
      alert("✅ Database retention policy saved!")
      isEditingDb.value = false
    } else { alert("❌ Failed to save database settings.") }
  } catch (err) { console.error("Failed saving database settings:", err) }
}

const triggerManualPurge = async () => {
  isPurging.value = true
  try {
    const apiUrl = '/api/settings/database/purge'
    const res = await fetch(apiUrl, { 
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ retention_days: dbSettings.value.retention_days }) 
    })
    if (res.ok) {
      const data = await res.json()
      alert(`🧹 ${data.message}`)
    } else { alert("❌ Failed to purge database.") }
  } catch (err) {
    console.error("Failed triggering purge:", err)
    alert("❌ Network error connecting to backend.")
  } finally { isPurging.value = false }
}

// ==========================================
// MOUNTING
// ==========================================
onMounted(async () => {
  await loadUserContext() 
  fetchHosts()
  fetchAllAlerts()
  fetchTelegramSettings()
  fetchDbSettings()
  fetchUsers() 
})
</script>

<style>
.animate-fade-in {
  animation: fadeIn 0.3s ease-in-out;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.02);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>