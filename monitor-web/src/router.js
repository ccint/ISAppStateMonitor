import Vue from 'vue'
import Router from 'vue-router'
import AnrIssueList from './views/AnrIssueList.vue'
import AnrSessionDetail from './views/AnrIssueDetail.vue'
import MissingDsymList from './views/MissingDsymList.vue'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/app/:aid/anr/',
      name: 'anrIssueList',
      component: AnrIssueList
    },
    {
      path: '/app/:aid/anr/issue_detail/:iid/session/:sid',
      name: 'anrIssueDetail',
      component: AnrSessionDetail
    },
    {
      path: '/app/:aid/missing_dsym',
      name: 'missingDsym',
      component: MissingDsymList
    }
  ]
})
