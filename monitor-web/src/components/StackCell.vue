<template>
    <div class="content" @click="clicked">
        <div class="header">
            <div class="serial font-standard" :style="style(info.isHighlight)">
                {{`#${info.threadSerial}`}}
            </div>
            <div class="thread-name font-standard">
                {{info.threadName}}
            </div>
            <div class="top-frame font-standard ellipsis">
                {{info.topFrameSymbol}}
            </div>
        </div>
        <div class="frames" v-if="showFrames" v-for="(index, frame) in info.frames" :key="index">
            <div class="frame">
                <div class="serial font-standard" :style="style(frame.isHighlight)">
                    {{`#${index}`}}
                </div>
                <div class="image-name font-standard" :style="style(frame.isHighlight)">
                    {{frame.imageName}}
                </div>
                <div class="detail">
                    <div class="source font-standard" :style="style(frame.isHighlight)">
                        {{frame.source}}
                    </div>
                    <div class="symbol font-standard" :style="style(frame.isHighlight)">
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
    style (isHighlight) {
      return `{color: ${isHighlight ? '#008bf3' : '#5a5a5a'}}`
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
        cursor: pointer;
        .header {
            display: flex;
            flex-direction: row;
            align-items: center;
            padding: 15px;
            .thread-name {
                font-weight: 700;
            }
            .top-frame {
                margin-right: 10px;
                max-width: 300px;
                font-weight: 400;
            }
        }
        .frames {
            display: flex;
            flex: 1 1 auto;
            .frame {
                padding: 8px;
                display: flex;
                flex-direction: row;
                .image-name {
                    margin-left: 5px;
                    flex: 0 0 auto;
                }
                .detail {
                    display: flex;
                    margin-left: 30px;
                    flex: 1 0 auto;
                    .source {
                        font-weight: 700;
                    }
                    .symbol {
                        font-weight: 400;
                    }
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
            color: #008bf3;
        }
        .serial {
            margin-right: 15px;
            font-weight: 700;
        }
    }

</style>
