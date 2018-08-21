<template>
    <div class="layout">
        <Modal  v-model="displayAppSelectModal"  width="400" title="选择App">
            <RadioGroup :value="selectedAppIdx" @input="selectedAppChanged" vertical>
                <Radio v-for="(app, idx) in apps" :label="idx" :key="app.appIdentifier">
                    <span>{{app.appIdentifier}}</span>
                </Radio>
            </RadioGroup>
            <div slot="footer">
            </div>
        </Modal>
        <div class="navi-board">
            <div class="app-board" @click="appBoardClicked">
                <img src="./assets/icon-cc.png" class="app-icon"/>
                <div class="app-info">
                    <div class="app-name ellipsis">
                        {{selectedApp.appName}}
                    </div>
                    <div class="app-id ellipsis">
                        {{selectedApp.appIdentifier}}
                    </div>
                </div>
            </div>
            <div class="menu">
                <div v-for="(item, idx) in menuItems"
                     :class="item.class"
                     :key="idx"
                     @click="menuItemClicked(idx)"
                >
                    <font-awesome-icon :icon="item.icon" class="muen-item-icon"/>
                    <span class="menu-item-text">{{item.name}}</span>
                </div>
            </div>
        </div>
        <div class="right-content">
            <div class="app-navi">
                <div class="app-navi-item"
                     v-for="(navi, idx) in navis"
                     :key="idx"
                     @click="naviClicked(idx)"
                >
                    {{navi.name}}
                    <font-awesome-icon icon="angle-right"
                                       style="margin: 2px 10px 0 10px; color: rgb(168, 181, 191); font-size: 15px"
                                       v-if="idx < navis.length - 1"
                    />
                </div>
            </div>
            <router-view class="app-router-view"/>
        </div>
    </div>
</template>

<script>
import { mapMutations, mapState, mapActions } from 'vuex'

export default {
  name: 'app',
  data () {
    return {
      navis: [],
      menuItems: [],
      selectedItemIdx: -1,
      displayAppSelectModal: false
    }
  },
  watch: {
    routePath (newValue) {
      this.updateNavis(newValue)
    },
    selectedItemIdx (newValue) {
      this.updateMenuItems()
    }
  },
  computed: {
    ...mapState('anr', {
      selectedAppIdx: state => state.selectedAppIdx,
      selectedApp: state => state.apps[state.selectedAppIdx] || {},
      apps: state => state.apps
    }),
    routePath () {
      return this.$route.path
    }
  },
  methods: {
    ...mapMutations('anr', ['setSelectedAppIdx']),
    ...mapActions('anr', ['getApps']),
    naviClicked (idx) {
      let navi = this.navis[idx]
      let path = navi.to
      if (typeof path !== 'undefined') {
        this.$router.push(path)
      }
    },
    menuItemClicked (idx, forceRefresh) {
      if (idx === this.selectedItemIdx && !forceRefresh) {
        return
      }
      let item = this.menuItems[idx]
      let path = item.to
      if (typeof path !== 'undefined') {
        this.$router.push(path)
      }
      this.selectedItemIdx = idx
    },
    appBoardClicked () {
      this.displayAppSelectModal = true
    },
    updateMenuItems () {
      let items = [{name: 'Anr Issues', icon: 'ban', to: `/app/${this.selectedApp.appIdentifier}/anr`},
        {name: 'Missing dSYMs', icon: 'cloud-upload-alt', to: `/app/${this.selectedApp.appIdentifier}/missing_dsym`}]
      for (let idx = 0; idx < items.length; ++idx) {
        let item = items[idx]
        if (this.selectedItemIdx === idx) {
          item['class'] = ['menu-item-selected']
        } else {
          item['class'] = ['menu-item-normal']
        }
      }
      this.menuItems = items
    },
    updateNavis (path) {
      if (typeof path === 'undefined') {
        path = this.$route.path
      }
      if (path.includes('/issue_detail') && path.includes('session')) {
        this.navis = [{to: `/app/${this.selectedApp.appIdentifier}/anr`, name: 'Anr Issues'}, {name: 'Issue Details'}]
      } else if (this.selectedItemIdx === 0 || this.selectedItemIdx === -1 || path === '/' || path === '') {
        this.navis = [{to: `/app/${this.selectedApp.appIdentifier}/anr`, name: 'Anr Issues'}]
      } else {
        this.navis = [{name: 'Missing dSYMs'}]
      }
    },
    selectedAppChanged (newValue) {
      this.displayAppSelectModal = false
      this.setSelectedAppIdx(newValue)
      this.$router.push(`/app/${this.selectedApp.appIdentifier}/anr`)
      this.updateMenuItems()
      this.menuItemClicked(0, true)
    }
  },
  beforeMount () {
    this.getApps().then(() => {
      let aid = this.$route.params.aid
      if (aid) {
        let matched = false
        for (let idx = 0; idx < this.apps.length; ++idx) {
          let app = this.apps[idx]
          if (app.appIdentifier === aid) {
            this.setSelectedAppIdx(idx)
            matched = true
          }
        }
        if (matched === false && this.apps.length > 0) {
          this.selectedAppChanged(0)
        }
      }
      let path = this.$route.path
      if (path.startsWith('/app') && path.includes('anr')) {
        this.selectedItemIdx = 0
      } else if (path.startsWith('/app') && path.includes('missing_dsym')) {
        this.selectedItemIdx = 1
      } else {
        this.selectedItemIdx = 0
        this.updateMenuItems()
        this.menuItemClicked(this.selectedItemIdx, true)
      }
      this.updateNavis(path)
    })
  }
}
</script>

<style lang="scss">
    @import "style/font.css";
    .layout {
        font-family: Source Sans Pro, sans-serif;
        font-size: 13px;
        font-weight: 400;
        background: rgb(16, 36, 49);
        border-radius: 4px;
        overflow: hidden;
        position: relative;
        .navi-board {
            float: left;
            height: 100vh;
            width: 250px;
            background: rgb(8, 31, 40);
            position: fixed;
        }
        .app-board {
            cursor: pointer;
            font-size: 16px;
            color: white;
            height: 100px;
            display: flex;
            align-items: center;
            transition: background ease 0.3s;
            border-bottom: 1px solid rgb(32, 53, 61);
            &:hover {
                background: rgba(20, 52, 79, 0.3);
            }
            .app-icon {
                height: 30px;
                width: 30px;
                margin-left: 20px;
                border-radius: 6px;
            }
            .app-info {
                display: flex;
                flex-direction: column;
                margin-left: 10px;
                margin-right: 20px;
                max-width: 185px;
                .app-name {
                    font-size: 16px;
                    font-weight: 500;
                    color: rgb(234, 245, 252);
                    line-height: 20px;
                    flex-shrink: 1;
                }
                .app-id {
                    font-size: 12px;
                    font-weight: 500;
                    color: rgb(161, 171, 187);
                    line-height: 15px;
                }
                .ellipsis {
                    min-width: 0;
                    text-overflow: ellipsis;
                    overflow: hidden;
                    white-space: nowrap;
                }
            }
        }
        .right-content {
            padding: 0 30px 0 30px;
            margin-left: 250px;
            width: calc(100% - 250px);
            overflow: scroll;
            &::-webkit-scrollbar {
                width: 0 !important;
            }
            .app-router-view {
                max-width: 1500px;
                width: 100%;
                min-width: 850px;
                margin: 0 auto 10px auto;
            }
            .app-navi {
                display: flex;
                align-items: center;
                margin: 30px auto 30px auto;
                font-size: 20px;
                font-weight: 400;
                max-width: 1500px;
                width: 100%;
                min-width: 850px;
                .app-navi-item {
                    color: rgb(168, 181, 191);
                    transition: color .2s ease-in-out;
                    display: flex;
                    align-items: center;
                    cursor: pointer;
                    &:hover {
                        color: #55acee;
                    }
                    &:last-child {
                        color: rgb(245, 248, 250);
                        cursor: default;
                    }
                    &:last-child:hover {
                        color: rgb(245, 248, 250);
                    }
                }
            }
        }
    }
    .menu {
        display: flex;
        flex-direction: column;
        .menu-item-normal {
            cursor: pointer;
            border-left: 6px solid transparent;
            font-size: 14px;
            font-weight: 500;
            display: flex;
            align-items: center;
            height: 40px;
            color: rgb(136, 153, 166);
            transition: background ease 0.3s;
            &:hover {
                background: rgba(20, 52, 79, 0.3);
            }
        }
        .menu-item-selected {
            cursor: pointer;
            border-left: 6px solid rgb(0, 139, 243);
            font-size: 14px;
            font-weight: 500;
            display: flex;
            align-items: center;
            height: 40px;
            color: rgb(234, 245, 252);
            background: rgb(20, 52, 79);
        }
        .muen-item-icon {
            margin: 0 15px 0 30px;
            width: 20px;
        }
    }
</style>
