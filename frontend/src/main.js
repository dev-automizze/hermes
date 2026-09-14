import { createApp } from 'vue'
import App from './App.vue'
import './style.css' 
import router from './router'
import VueApexCharts from 'vue3-apexcharts'

// --- NEW: GLOBAL FETCH INTERCEPTOR ---
// This automatically adds your security token to EVERY fetch call in your app!
const originalFetch = window.fetch;
window.fetch = async (...args) => {
  let [resource, config] = args;
  if (!config) config = {};
  
  // Ensure headers object exists
  if (!config.headers) config.headers = {};
  
  // Grab the token from browser memory
  const token = localStorage.getItem('hermes_token');
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  
  return originalFetch(resource, config);
};
// -------------------------------------

const app = createApp(App)
app.use(router) // NEW
app.use(VueApexCharts)
app.mount('#app')