import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import iView from 'iview'
import 'iview/dist/styles/iview.css'
import { library } from '@fortawesome/fontawesome-svg-core'
import {faCaretRight, faCaretLeft} from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(faCaretRight, faCaretLeft)

Vue.component('font-awesome-icon', FontAwesomeIcon)

Vue.config.productionTip = false
Vue.use(iView)

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
