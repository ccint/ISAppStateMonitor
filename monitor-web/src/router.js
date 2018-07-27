import Vue from 'vue'
import Router from 'vue-router'
import AnrIssueList from './views/AnrIssueList.vue'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/anr/',
      name: 'anrIssueList',
      component: AnrIssueList
    }
  ]
})
