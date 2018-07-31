<template>
    <div class="content">
        <div class="header" @click="clicked">
            <div class="serial font-standard" :style="style(info.isHighlight, true, true)">
                {{`#${info.threadSerial}`}}
            </div>
            <div class="thread-name font-standard" :style="style(info.isHighlight, true, true)">
                {{`${info.threadName}`}}
            </div>
            <div class="top-frame font-standard ellipsis" v-if="!showFrames">
                {{info.topFrameSymbol || "missing"}}
            </div>
        </div>
        <div class="frames" v-if="showFrames">
            <div class="frame" v-for="(frame, index) in info.frames" :key="index">
                <div class="serial font-standard" :style="style(frame.isHighlight, true)">
                    {{`${index}`}}
                </div>
                <div class="image-name font-standard" :style="style(frame.isHighlight, true)">
                    {{frame.imageName}}
                </div>
                <div class="detail">
                    <div class="source font-standard ellipsis" :style="style(frame.isHighlight, true)">
                        {{frame.source}}
                    </div>
                    <div class="symbol font-standard ellipsis" :style="style(frame.isHighlight, false)">
                        {{frame.symbol || "missing"}}
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
export default {
  name: 'StackCell',
  props: {
    info: {},
    expandDefault: false
  },
  data () {
    return {
      expand: false
    }
  },
  computed: {
    showFrames () {
      return this.expand === true
    }
  },
  methods: {
    clicked () {
      this.expand = !this.expand
    },
    style (isHighlight, strongHighlight, strong) {
      return {
        color: `${isHighlight ? '#008bf3' : '#5a5a5a'}`,
        fontWeight: (strongHighlight && isHighlight) || strong ? 700 : 400
      }
    }
  },
  created () {
    this.expand = this.expandDefault
  }
}
</script>

<style scoped lang="scss">
    .content {
        display: flex;
        flex-direction: column;
        border: 1px solid #e1e5e8;
        border-radius: 5px;
        overflow: hidden;
        .header {
            display: flex;
            flex: 1 1 auto;
            flex-direction: row;
            align-items: center;
            padding: 15px;
            height: 65px;
            background: #f5f7f9;
            cursor: pointer;
            .thread-name {
                margin-right: 30px;
                flex: 1 0 auto;
            }
            .top-frame {
                margin-left: auto;
                margin-right: 10px;
            }
        }
        .frames {
            display: flex;
            flex-direction: column;
            flex: 1 0 auto;
            background: white;
            .frame {
                padding: 8px 15px 8px 15px;
                display: flex;
                flex-direction: row;
                align-items: center;
                position: relative;
                border-top: 1px solid #e1e5e8;
                .image-name {
                    margin-left: 5px;
                    width: 150px;
                    flex: 0 0 auto;
                }
                .detail {
                    display: flex;
                    flex-direction: column;
                    margin-left: 30px;
                }
            }
        }
        .ellipsis {
            min-width: 0;
            text-overflow: ellipsis;
            overflow: hidden;
            white-space: nowrap;
        }
        .font-standard {
            color: #5a5a5a;
            letter-spacing: 0;
            -webkit-font-smoothing: antialiased;
        }
        .serial {
            margin-right: 15px;
            width: 20px;
            flex: 0 0 auto;
        }
    }

</style>
