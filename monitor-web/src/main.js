import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import iView from 'iview'
import 'iview/dist/styles/iview.css'
import { library } from '@fortawesome/fontawesome-svg-core'
import {faCaretRight, faCaretLeft, faCloudUploadAlt, faAngleRight, faBan} from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(faCaretRight, faCaretLeft, faBan, faAngleRight, faCloudUploadAlt)

Vue.component('font-awesome-icon', FontAwesomeIcon)
Vue.use(iView)

Vue.config.productionTip = false
Vue.use(iView)

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
