<template>
    <div class="content">
        <div class="issue text">
            <div class="title ellipsis">
                {{info.title}}
            </div>
            <div class="detail ellipsis">
                {{info.detail}}
            </div>
        </div>
        <div class="version ellipsis">
            {{info.version}}
        </div>
        <div class="count text">
            <div class="title ellipsis">
                {{info.issueCount}}
            </div>
            <div class="detail ellipsis">
                OCCURRED
            </div>
        </div>
    </div>
</template>

<script>
export default {
  name: 'IssueCell',
  props: {
    info: {}
  }
}
</script>

<style scoped lang="scss">
    .content {
        display: flex;
        position: relative;
        flex-direction: row;
        background: white;
        cursor: pointer;
        transition: background ease 0.5s;
        &:hover {
            background: #eef7fd;
        }
        &:after {
             content: ' ';
             position: absolute;
             height: 1px;
             background: rgb(224, 229, 232);
             right: 0;
             left: 0;
             top: 0;
         }
    }
    .text {
        display: flex;
        flex-direction: column;
        padding: 15px;
        .title {
            font-size: 14px;
            line-height: 1.4;
            font-family: Source Sans Pro,sans-serif;
            color: rgb(14, 30, 40);
            font-weight: 500;
        }
        .detail {
            font-size: 13px;
            padding-top: 3px;
            line-height: 1.5;
            color: rgb(90, 90, 90);
            font-family: Source Sans Pro,sans-serif;
        }
        .ellipsis {
            min-width: 0;
            text-overflow: ellipsis;
            overflow: hidden;
            white-space: nowrap;
        }
    }
    .version {
        display: flex;
        position: relative;
        justify-content: center;
        align-items: center;
        padding: 15px;
        color: rgb(102, 117, 127);
        font-size: 13px;
        font-weight: 500;
        width: 300px;
        &:after {
            content: ' ';
            position: absolute;
            width: 1px;
            background: rgb(224, 229, 232);
            right: 0;
            top: 0;
            bottom: 0;
        }
    }
    .count {
        width: 150px;
    }
    .issue {
        flex: 1 1 auto;
        position: relative;
        &:after {
            content: ' ';
            position: absolute;
            width: 1px;
            background: rgb(224, 229, 232);
            right: 0;
            top: 0;
            bottom: 0;
        }
    }

</style>
