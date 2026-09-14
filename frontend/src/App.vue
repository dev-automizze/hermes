<template>
  <div v-if="!isAuthenticated" class="min-h-screen bg-hermes-bg text-hermes-text flex items-center justify-center">
    <Login @login-success="verifyToken" />
  </div>

  <div v-else class="min-h-screen flex flex-col font-sans bg-hermes-bg text-hermes-text relative">
    
    <header class="border-b border-hermes-accent/20 bg-black/10 p-4 sticky top-0 z-50 backdrop-blur-md">
      <div class="max-w-7xl mx-auto flex justify-between items-center">
        <div class="flex items-center gap-6">
          <h1 class="text-2xl font-bold text-hermes-accent tracking-widest uppercase">
            Hermes
          </h1>
          <nav class="hidden md:flex gap-4">
            <router-link to="/" class="px-3 py-1.5 rounded-md transition font-medium" :class="$route.name === 'dashboard' ? 'bg-hermes-accent/20 text-hermes-accent' : 'text-gray-400 hover:text-white'">
              Dashboard
            </router-link>
            <router-link to="/settings" class="px-3 py-1.5 rounded-md transition font-medium" :class="$route.name === 'settings' ? 'bg-hermes-accent/20 text-hermes-accent' : 'text-gray-400 hover:text-white'">
              Settings
            </router-link>
          </nav>
        </div>

        <div class="flex items-center gap-4">
          
          <button @click="isFeedbackOpen = true" class="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-white/5 border border-white/10 rounded-md hover:bg-white/10 text-xs font-semibold transition cursor-pointer text-gray-300 hover:text-white">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-[#FF5FCF]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
            Feedback
          </button>

          <div class="hidden sm:flex items-center gap-2 px-3 py-1 bg-white/5 border border-white/10 rounded-full shadow-inner">
            <span class="w-2 h-2 rounded-full bg-hermes-accent shadow-[0_0_8px_rgba(52,211,153,0.6)] animate-pulse"></span>
            <span class="text-xs font-bold tracking-wider text-hermes-text/80 uppercase">{{ currentUsername }}</span>
          </div>
          
          <button @click="logout" class="px-3 py-1.5 border border-red-500/30 text-red-400 bg-red-500/10 hover:bg-red-500/20 text-xs font-semibold rounded-md transition flex items-center gap-1 cursor-pointer">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Logout
          </button>
        </div>
      </div>
    </header>

    <router-view></router-view>

    <footer class="w-full py-6 mt-auto text-center border-t border-hermes-accent/10 bg-black/10 relative z-10">
      <p class="text-xs tracking-wider text-hermes-text/50 font-medium">
        Hermes Network Monitor &bull; Built by
        <span class="font-bold text-hermes-text/80 transition-colors hover:text-hermes-accent cursor-default">
          Chamroeun_Soknara
        </span>
      </p>
    </footer>

    <div v-if="isFeedbackOpen" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm px-4">
      <div class="bg-[#121212] border border-white/10 rounded-2xl p-6 w-full max-w-md shadow-2xl relative">
        
        <button @click="isFeedbackOpen = false" class="absolute top-4 right-4 text-gray-500 hover:text-white transition-colors cursor-pointer">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>

        <h3 class="text-xl font-bold text-white mb-1 flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-[#FF5FCF]" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" /></svg>
          Send Feedback
        </h3>
        <p class="text-gray-400 text-sm mb-6">Found a bug or have a suggestion? Let me know!</p>

        <form @submit.prevent="submitFeedback" class="flex flex-col gap-4">
          <div>
            <label class="block text-[10px] font-bold uppercase text-gray-500 tracking-wider mb-1">Name</label>
            <input v-model="feedbackData.name" type="text" required class="w-full bg-black/30 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-[#FF5FCF] transition-colors" placeholder="Your Name">
          </div>
          
          <div>
            <label class="block text-[10px] font-bold uppercase text-gray-500 tracking-wider mb-1">Email</label>
            <input v-model="feedbackData.email" type="email" required class="w-full bg-black/30 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-[#FF5FCF] transition-colors" placeholder="you@example.com">
          </div>

          <div>
            <label class="block text-[10px] font-bold uppercase text-gray-500 tracking-wider mb-1">Message</label>
            <textarea v-model="feedbackData.message" required rows="4" class="w-full bg-black/30 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-[#FF5FCF] transition-colors resize-none" placeholder="What's on your mind?"></textarea>
          </div>

          <button type="submit" :disabled="isSubmitting" class="mt-2 w-full py-2.5 bg-[#FF5FCF]/20 text-[#FF5FCF] hover:bg-[#FF5FCF]/30 border border-[#FF5FCF]/30 rounded-lg font-bold transition-colors flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed">
            <span v-if="isSubmitting" class="w-4 h-4 border-2 border-[#FF5FCF] border-t-transparent rounded-full animate-spin"></span>
            {{ isSubmitting ? 'Sending...' : 'Send Message' }}
          </button>
        </form>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Login from './components/Login.vue' 

const router = useRouter()
const isAuthenticated = ref(false)
const currentUsername = ref('') 

// --- FEEDBACK FORM LOGIC ---
const isFeedbackOpen = ref(false)
const isSubmitting = ref(false)
const feedbackData = ref({
  name: '',
  email: '',
  message: ''
})

const submitFeedback = async () => {
  isSubmitting.value = true
  try {
    const response = await fetch('https://formspree.io/f/xpqgdeqz', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(feedbackData.value)
    })

    if (response.ok) {
      alert("Feedback sent successfully! Thank you.")
      feedbackData.value = { name: '', email: '', message: '' } // Clear form
      isFeedbackOpen.value = false // Close modal
    } else {
      alert("Oops! There was a problem submitting your feedback.")
    }
  } catch (error) {
    console.error(error)
    alert("Network error. Please try again.")
  } finally {
    isSubmitting.value = false
  }
}
// ---------------------------

const verifyToken = async () => {
  const token = localStorage.getItem('hermes_token')
  if (!token) {
    isAuthenticated.value = false
    return
  }

  try {
    const apiUrl = '/api/auth/check'
    const res = await fetch(apiUrl, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    const data = await res.json()
    
    if (data.authenticated) {
      isAuthenticated.value = true
      currentUsername.value = data.username 
    } else {
      logout() 
    }
  } catch (err) {
    console.error("Failed connecting to the authentication engine:", err)
    isAuthenticated.value = false
  }
}

const logout = () => {
  localStorage.removeItem('hermes_token')
  localStorage.removeItem('hermes_username')
  isAuthenticated.value = false
  currentUsername.value = ''
  
  if (router) {
    router.push('/')
  }
}

onMounted(() => {
  verifyToken()
})
</script>

<style>
/* Keep any custom global styles you had defined down here intact */
</style>