<template>
    <div class="listcontaner">
        <div class="issueList" v-for="issue in issues" :key="issue.id">
            <IssueCell :info="issue">
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
      issues: state => state.issueList
    })
  },
  methods: {
    ...mapActions('anr', ['getIssueList'])
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
        width: 1400px;
    }
    .listcontaner {
        border-radius: 10px;
        overflow: hidden;
        width: 1400px;
    }
</style>
