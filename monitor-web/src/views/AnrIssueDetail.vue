<template>
    <div class="listcontaner">
        <div class="navi">
            <div class="navi-bar">
                <font-awesome-icon size="2x"
                                   icon="caret-left"
                                   :style="{ color: hasNext ? 'rgb(85, 172, 238)' : 'rgb(222, 228, 232)'}"
                                   @click="goNext(false)"
                />
                <div style="margin: 0 15px 0 15px;">
                    {{sessionDate}}
                </div>
                <font-awesome-icon size="2x"
                                   icon="caret-right"
                                   :style="{ color: hasPrev ? 'rgb(85, 172, 238)' : 'rgb(222, 228, 232)'}"
                                   @click="goNext(true)"
                />
            </div>
            <div style="font-style: italic">
                {{`App Version ${currentSession.appVersion}`}}
            </div>
            <div style="font-style: italic">
                {{`Runloop duration ${Math.round(currentSession.duration * 100) / 100} ms `}}
            </div>
        </div>
        <div class="stack-header">
            Stacktrace
            <label class="resymbol-label"
                   @click="tryReSymbolicate"
            >
                Re-Symbolicate
            </label>
        </div>
        <stack-cell class="stack-frames"
                    v-for="(stack, idx) in stacks"
                    :key="idx"
                    :expandDefault="stack.expandDefault"
                    :info="stack" >
        </stack-cell>
    </div>
</template>

<script>
import stackCell from '../components/StackCell'
import { mapState, mapActions } from 'vuex'

export default {
  name: 'AnrIssueDetail',
  components: {
    stackCell
  },
  computed: {
    ...mapState('anr', {
      issueDetail: state => state.issueDetail,
      currentSession: state => state.currentSession,
      sessionCount: state => state.issueDetail.sessions.length,
      selectedApp: state => state.apps[state.selectedAppIdx] || {}
    }),
    hasNext () {
      return this.currentSession.idx < this.sessionCount - 1 && this.currentSession.idx >= 0
    },
    hasPrev () {
      return this.currentSession.idx > 0
    },
    sessionDate () {
      let date = new Date(this.currentSession.date)
      return date.toLocaleString()
    },
    stacks () {
      let stacks = this.currentSession.stacks
      if (typeof stacks === 'undefined') {
        return []
      }
      stacks[0].isHighlight = true
      stacks[0].expandDefault = true
      for (let index = 0; index < stacks.length; ++index) {
        let stack = stacks[index]
        stack.threadSerial = index
        stack.topFrameSymbol = stack && stack.frames && stack.frames[0] && stack.frames[0].symbol
        let frames = stack.frames
        if (frames) {
          for (let frame of frames) {
            if (frame.imageName === this.currentSession.appImage) {
              frame.isHighlight = true
            }
          }
        }
      }
      return stacks
    }
  },
  methods: {
    ...mapActions('anr', ['getSessionDetail', 'reSymbolicate']),
    goNext (isPrev) {
      if ((isPrev && this.hasPrev) || (!isPrev && this.hasNext)) {
        let nextIdx = isPrev ? this.currentSession.idx - 1 : this.currentSession.idx + 1
        let nextId = this.issueDetail.sessions[nextIdx]
        this.$router.push(`/app/${this.selectedApp.appIdentifier}/anr/issue_detail/${this.$route.params.iid}/session/${nextId}`)
      }
    },
    tryReSymbolicate () {
      this.reSymbolicate({sid: this.$route.params.sid})
    }
  },
  beforeMount () {
    this.getSessionDetail({iid: this.$route.params.iid, sid: this.$route.params.sid})
  },
  beforeRouteUpdate (to, from, next) {
    if (to.name === from.name) { // 手动刷新数据
      this.getSessionDetail({iid: to.params.iid, sid: to.params.sid})
    }
    next()
  }
}
</script>

<style scoped lang="scss">
    .listcontaner {
        display: flex;
        flex-direction: column;
        flex: 1 1 auto;
        padding: 0 25px;

        background: white;
        border-radius: 10px;
        overflow: hidden;
        .navi {
            display: flex;
            flex-direction: column;
            justify-content: center;
            height: 95px;
            color: rgb(5, 5, 5);
            background: rgb(245, 247, 249);
            margin: 0 -25px 0 -25px;
            padding: 0 25px 0 25px;
            border-bottom: 1px solid rgb(225, 229, 232);
            .navi-bar {
                display: flex;
                flex-direction: row;
                padding-bottom: 3px;
                align-items: center;
                font-size: 1.4em;
            }
        }
        .stack-header {
            font-size: 16px;
            margin: 15px 0 15px 0;
            color: rgb(5, 5, 5);
            display: flex;
            align-items: center;
            justify-content: space-between;
            .resymbol-label {
                cursor: pointer;
                color: white;
                background: rgb(0, 139, 243);
                border-radius: 5px;
                padding: 2px 8px 2px 8px;
                font-size: 14px;
                font-style: italic;
                transition: background ease 0.3s;
                &:hover {
                    background: rgba(0, 139, 243, 0.8);
                }
            }
        }
        .stack-frames {
            margin-bottom: 20px;
            &:last-child {
                margin-bottom: 40px;
            }
        }
    }
</style>
