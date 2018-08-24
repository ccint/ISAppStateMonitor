<template>
    <div class="listcontaner">
        <div class="issueList">
            <div class="header">
                <div class="header-title">
                    TOTAL ISSUES
                </div>
                <div class="issue-num">
                    {{totalIssues || 0}}
                </div>
                <div class="unsymbole-issue" @click="reClassfiedReports" v-if="unclassfiedCount > 0">
                    {{`${unclassfiedCount} report are unclassified, press to resymbolicate!`}}
                </div>
            </div>
            <IssueCell :info="issue"
                       v-for="(issue, idx) in issues"
                       :key="issue.id"
                       @click="gotoIssueDetail(idx)">
            </IssueCell>
        </div>
        <Page class="page-control"
              :current="currentPage"
              :total="totalIssues"
              :pageSize="pageSize"
              @on-change="onPageChange"
              show-elevator />
    </div>
</template>

<script>
// @ is an alias to /src
import IssueCell from '../components/IssueCell'
import { mapState, mapActions, mapMutations } from 'vuex'
import {reClassfiedReports} from '../API/resymbolicate'

export default {
  name: 'anrIssueList',
  components: {
    IssueCell
  },
  data () {
    return {
      pageSize: 20
    }
  },
  computed: {
    ...mapState('anr', {
      issues: state => state.issueList.issues,
      issueDetail: state => state.issueDetail,
      currentPage: state => state.currentIssuePage,
      totalIssues: state => state.issueList.total,
      selectedApp: state => state.apps[state.selectedAppIdx] || {},
      unclassfiedCount: state => state.unclassfiedCount
    })
  },
  methods: {
    ...mapActions('anr', ['getIssueList', 'getIssueDetail']),
    ...mapMutations('anr', ['setCurrentIssuePage']),
    gotoIssueDetail (idx) {
      let issue = this.issues[idx]
      let id = issue.id
      this.getIssueDetail({id}).then(() => {
        let sid = this.issueDetail.sessions[0]
        this.$router.push(`/app/${this.selectedApp.appIdentifier}/anr/issue_detail/${id}/session/${sid}`)
      })
    },
    onPageChange (page) {
      this.setCurrentIssuePage(page)
      this.loadIssues(this.selectedApp.appIdentifier)
    },
    loadIssues (appId) {
      let start = (this.currentPage - 1) * this.pageSize
      let pageSize = this.pageSize
      this.getIssueList({start, pageSize, appId})
    },
    reClassfiedReports () {
      reClassfiedReports(this.selectedApp.appIdentifier).then(() => {
        this.loadIssues(this.selectedApp.appIdentifier)
      })
    }
  },
  beforeMount () {
    this.loadIssues(this.$route.params.aid)
  },
  beforeRouteUpdate (to, from, next) {
    if (to.name === from.name) { // 手动刷新数据
      this.loadIssues(to.params.aid)
    }
    next()
  }
}
</script>

<style lang="scss">
    .issueList {
        display: flex;
        flex-direction: column;
    }
    .listcontaner {
        background: white;
        border-radius: 10px;
        overflow: hidden;
        flex: 1 1 auto;
        .page-control {
            margin: 15px;
        }
        .header {
            display: flex;
            flex: 1 1 auto;
            flex-direction: row;
            align-items: center;
            padding: 15px;
            height: 65px;
            background: rgb(215, 226, 233);
            .header-title {
                color: rgb(102, 117, 127);
                font-weight: 500;
            }
            .issue-num {
                margin: -6px 15px 0 10px;
                color: rgb(12, 46, 69);
                font-weight: 400;
                font-size: 23px;
                letter-spacing: 0.5px
            }
            .unsymbole-issue {
                margin-left: auto;
                color: rgb(0, 139, 243);
                font-weight: 500;
                font-size: 14px;
                cursor: pointer;
                text-decoration: underline;
                &:hover {
                    color: rgba(0, 139, 243, 0.8);
                }
            }
        }
    }
</style>
