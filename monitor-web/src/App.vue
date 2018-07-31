<template>
    <div class="layout">
        <Sider :style="{position: 'fixed', height: '100vh', left: 0, backgroundColor: 'rgb(8, 31, 40)'}">
            <div class="app-board">
                <img src="./assets/icon-cc.png" class="app-icon"/>
                <div class="app-info">
                    <div class="app-name ellipsis">
                        名片全能王
                    </div>
                    <div class="app-id ellipsis">
                        com.intsig.camcard.lite
                    </div>
                </div>
            </div>
            <div class="menu">
                <div class="menu-item">
                    <font-awesome-icon icon="spinner" class="muen-item-icon"/>
                    <span class="menu-item-text">Anr Issues</span>
                </div>
            </div>
        </Sider>
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
            <router-view/>
        </div>
    </div>
</template>

<script>
export default {
  name: 'app',
  data () {
    return {
      navis: []
    }
  },
  watch: {
    routePath (newValue) {
      this.updateNavis(newValue)
    }
  },
  computed: {
    routePath () {
      return this.$route.path
    }
  },
  methods: {
    naviClicked (idx) {
      let navi = this.navis[idx]
      let path = navi.to
      if (typeof path !== 'undefined') {
        this.$router.push(path)
      }
    },
    updateNavis (path) {
      if (path.includes('/anr/issue_detail/') && path.includes('session')) {
        this.navis = [{to: '/anr', name: 'Anr Issues'}, {name: 'Issue Details'}]
      } else {
        this.navis = [{to: '/anr', name: 'Anr Issues'}]
      }
    }
  },
  beforeMount () {
    let path = this.$route.path
    this.updateNavis(path)
    if (path === '/' || path === '') {
      this.naviClicked(0) // 默认项 之后会改造
    }
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
        position: relative;
        border-radius: 4px;
        overflow: hidden;
        .app-board {
            font-size: 16px;
            color: white;
            height: 100px;
            display: flex;
            align-items: center;
            transition: background ease 0.5s;
            border-bottom: 1px solid rgb(32, 53, 61);
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
            padding: 0 30px 30px 30px;
            margin-left: 200px;
            overflow: scroll;
            .app-navi {
                display: flex;
                align-items: center;
                margin: 30px 0;
                font-size: 20px;
                font-weight: 400;
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
        .menu-item {
            cursor: pointer;
            border-left: 6px solid rgb(0, 139, 243);
            background: rgb(20, 52, 79);
            font-size: 14px;
            font-weight: 500;
            display: flex;
            align-items: center;
            height: 35px;
            color: rgb(234, 245, 252);
            .muen-item-icon {
                margin: 0 15px 0 30px;
            }
        }
    }
</style>
