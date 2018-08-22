<template>
    <div class="listcontaner">
        <Modal v-model="displayModal" width="500" :title="`${succeedItems.length}份符号表导入成功`">
            <div style="display: flex; flex-direction: column; max-height: 400px; overflow-y: scroll">
                <div style="display: flex; flex: 1 0 auto;
                justify-content: space-between;
                height: 35px; align-items: center;
                border-bottom: 1px solid #e7e7e7;
                color: #008bf3;"
                     v-for="item in succeedItems"
                     :key="item.uuid"
                >
                    <label>{{(item.uuid && item.uuid.toUpperCase()) || ''}}</label>
                    <label>{{item.name || ''}}</label>
                </div>
            </div>
            <div slot="footer">
                <Button type="primary" size="default"  @click="confirm">确定</Button>
            </div>
        </Modal>
        <div class="issueList">
            <div class="header">
                <div class="header-title">
                    TOTAL MISSING DSYMS
                </div>
                <div class="issue-num" style="margin-right: auto">
                    {{count || 0}}
                </div>
                <i-circle style="margin-right: 10px" v-show="progress !== 100" :size="45" :percent="this.progress" dashboard>
                    <span class="demo-circle-inner" style="font-size:10px">{{`${progress}%`}}</span>
                </i-circle>
                <label class="upload-label"
                >
                    <font-awesome-icon icon="cloud-upload-alt" style="margin-right: 3px"/>
                    Upload dSYM Files
                    <input
                            id="input"
                            ref="inputButton"
                            type="file"
                            name="uploadFile"
                            style="display: none"
                            accept=".zip"
                            @change="uploadDsyms"
                            :disabled="progressDisabled"
                    />
                </label>
            </div>
            <MissingDsymCell :info="dsym"
                       v-for="(dsym, idx) in dsyms"
                       :key="idx"
            >
            </MissingDsymCell>
        </div>
    </div>
</template>

<script>
// @ is an alias to /src
import MissingDsymCell from '../components/MissingDsymCell'
import {mapState, mapActions} from 'vuex'
import {uploadDsym} from '../API/uploadService'

export default {
  name: 'MissingDsymList',
  components: {
    MissingDsymCell
  },
  data () {
    return {
      progress: 100,
      progressDisabled: false,
      displayModal: false,
      succeedItems: []
    }
  },
  computed: {
    ...mapState('missingDsym', {
      dsyms: state => state.missingDsym.data,
      count: state => state.missingDsym.count
    })
  },
  methods: {
    ...mapActions('missingDsym', ['getMissingDsyms']),
    uploadDsyms (event) {
      let file = event.target.files[0]
      if (!file) {
        return
      }
      this.progress = 0
      this.progressDisabled = true
      let dsymData = new FormData()
      dsymData.append('file', file)
      uploadDsym(dsymData, process => {
        this.progress = Math.min(process, 50)
      })
        .then(response => {
          this.progress = 100
          this.progressDisabled = false
          if (response.data.ret === '0') {
            this.succeedItems = response.data.data
            this.displayModal = true
          } else {
            alert('ret error: ' + JSON.stringify(response.data))
          }
          event.target.value = null
          this.getMissingDsyms({appId: this.$route.params.aid})
        })
        .catch(error => {
          this.progressDisabled = false
          alert(`Error: ${error}`)
          event.target.value = null
          this.progress = 100
        })
    },
    confirm () {
      this.displayModal = false
      this.succeedItems = []
    }
  },
  beforeMount () {
    this.getMissingDsyms({appId: this.$route.params.aid})
  },
  beforeRouteUpdate (to, from, next) {
    if (to.name === from.name) { // 手动刷新数据
      this.getMissingDsyms({appId: to.params.aid})
    }
    next()
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
            .upload-label {
                cursor: pointer;
                color: white;
                background: rgb(0, 139, 243);
                border-radius: 5px;
                padding: 6px 15px 6px 15px;
                font-size: 14px;
                transition: background ease 0.3s;
                &:hover {
                    background: rgba(0, 139, 243, 0.8);
                }
            }
        }
    }
</style>
