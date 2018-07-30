<template>
    <div class="listcontaner">
        <div class="issueList">
            <IssueCell :info="issue" v-for="(issue, idx) in issues" :key="issue.id" @click="gotoIssueDetail(idx)">
            </IssueCell>
        </div>
    </div>
</template>

<script>
// @ is an alias to /src
import IssueCell from '../components/IssueCell'
import { mapState, mapActions } from 'vuex'

export default {
  name: 'anrIssueList',
  components: {
    IssueCell
  },
  computed: {
    ...mapState('anr', {
      issues: state => state.issueList,
      issueDetail: state => state.issueDetail
    })
  },
  methods: {
    ...mapActions('anr', ['getIssueList', 'getIssueDetail']),
    gotoIssueDetail (idx) {
      let issue = this.issues[idx]
      let id = issue.id
      this.getIssueDetail({id}).then(() => {
        let sid = this.issueDetail.sessions[0]
        this.$router.push(`/anr/issue_detail/${id}/session/${sid}`)
      })
    }
  },
  beforeMount () {
    this.getIssueList()
  }
}
</script>

<style lang="scss">
    .issueList {
        display: flex;
        flex-direction: column;
        flex: 1 1 auto;
    }
    .listcontaner {
        background: white;
        border-radius: 10px;
        overflow: hidden;
        max-width: 1500px;
        min-width: 850px;
    }
</style>
