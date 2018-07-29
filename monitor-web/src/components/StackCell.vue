<template>
    <div class="content" @click="clicked">
        <div class="header">
            <div class="serial font-standard" :style="style(info.isHighlight, true)">
                {{`#${info.threadSerial}`}}
            </div>
            <div class="thread-name font-standard" :style="style(info.isHighlight, true)">
                {{`Thread: ${info.threadName}`}}
            </div>
            <div class="top-frame font-standard ellipsis" v-if="!showFrames">
                {{info.topFrameSymbol}}
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
                        {{frame.symbol}}
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
    style (isHighlight, strongHighlight) {
      return {color: `${isHighlight ? '#008bf3' : '#5a5a5a'}`, fontWeight: strongHighlight && isHighlight ? 700 : 400}
    }
  },
  created () {
    this.expand = this.expandDefault
  }
}
</script>

<style lang="scss">
    .content {
        display: flex;
        flex-direction: column;
        border: 1px solid rgb(218, 223, 226);
        border-radius: 5px;
        overflow: hidden;
        .header {
            display: flex;
            flex: 1 1 auto;
            flex-direction: row;
            align-items: center;
            padding: 15px;
            background: #eef7fd;
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
                border-top: 1px solid rgb(218, 223, 226);
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
            font-family: Source Sans Pro,sans-serif;
            font-size: 1em;
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
