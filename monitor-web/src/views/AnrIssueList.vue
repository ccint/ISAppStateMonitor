<template>
    <div class="listcontaner">
        <div class="issueList" v-for="issue in issues" :key="issue.id">
            <IssueCell class="cell" :info="issue">
            </IssueCell>
        </div>
    </div>
</template>

<script>
// @ is an alias to /src
import IssueCell from '../components/StackCell'
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
        flex: 1 1 auto;
    }
    .listcontaner {
        background: white;
        padding: 15px;
        border-radius: 10px;
        overflow: hidden;
        max-width: 1500px;
        min-width: 850px;
    }
    .cell {
        margin-bottom: 15px;
    }
</style>
