<template>
  <!-- Added w-full and swapped background to match your Hermes theme -->
  <div class="w-full min-h-screen bg-hermes-bg flex items-center justify-center p-4">
    
    <!-- Background Glow -->
    <div class="absolute w-[500px] h-[500px] bg-hermes-accent/10 rounded-full mix-blend-screen filter blur-[100px] opacity-50"></div>

    <div class="bg-black/40 border border-white/10 rounded-2xl p-8 shadow-2xl relative z-10 w-full max-w-md backdrop-blur-md">
      <div class="text-center mb-8">
        <div class="w-16 h-16 bg-hermes-accent/10 text-hermes-accent rounded-2xl flex items-center justify-center mx-auto mb-4 border border-hermes-accent/20 shadow-[0_0_15px_rgba(52,211,153,0.2)]">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
        </div>
        <h2 class="text-3xl font-bold text-white tracking-tight uppercase">Hermes</h2>
        <p class="text-gray-400 text-sm mt-2">Enter your credentials to access the network.</p>
      </div>

      <form @submit.prevent="doLogin" class="space-y-5">
        <div v-if="errorMsg" class="bg-red-500/10 border border-red-500/30 text-red-400 px-4 py-3 rounded-lg text-sm text-center font-medium">
          {{ errorMsg }}
        </div>

        <div>
          <label class="block text-xs font-bold uppercase text-gray-400 mb-1.5 tracking-wider">Username</label>
          <input type="text" v-model="username" required 
                 class="w-full bg-black/60 border border-white/10 rounded-lg p-3 text-white focus:border-hermes-accent outline-none transition placeholder-gray-600">
        </div>

        <div>
          <label class="block text-xs font-bold uppercase text-gray-400 mb-1.5 tracking-wider">Password</label>
          <input type="password" v-model="password" required 
                 class="w-full bg-black/60 border border-white/10 rounded-lg p-3 text-white focus:border-hermes-accent outline-none transition placeholder-gray-600">
        </div>

        <button type="submit" :disabled="isLoading" 
                class="w-full py-3 bg-hermes-accent hover:bg-emerald-400 text-black rounded-lg font-bold transition shadow-lg shadow-hermes-accent/20 mt-4 disabled:opacity-50 uppercase tracking-wider text-sm">
          {{ isLoading ? 'Authenticating...' : 'Access Dashboard' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['login-success'])

const username = ref('')
const password = ref('')
const errorMsg = ref('')
const isLoading = ref(false)

const doLogin = async () => {
  isLoading.value = true
  errorMsg.value = ''
  
  try {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value })
    })
    
    const data = await res.json()
    
    if (res.ok) {
      // Save the token!
      localStorage.setItem('hermes_token', data.token)
      localStorage.setItem('hermes_username', data.username)
      // Tell App.vue to switch screens
      emit('login-success')
    } else {
      errorMsg.value = data.error || "Authentication failed."
    }
  } catch (err) {
    errorMsg.value = "Network error. Is the backend running?"
  } finally {
    isLoading.value = false
  }
}
</script>