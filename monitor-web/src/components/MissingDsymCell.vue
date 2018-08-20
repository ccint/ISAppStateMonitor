<template>
    <div class="content">
        <div class="issue text">
            <div class="title ellipsis">
                {{info.Name}}
            </div>
            <div class="detail ellipsis">
                {{info.UUID || ''}}
            </div>
        </div>
        <div class="version ellipsis">
            {{info.Arch}}
        </div>
        <div class="version ellipsis">
            {{`${info.AppName || ''} ${info.AppVersion || ''}`}}
        </div>
        <div class="version ellipsis">
            {{'iOS ' + info.SystemVersion}}
        </div>
    </div>
</template>

<script>
export default {
  name: 'missingDsymCell',
  props: {
    info: {}
  }
}
</script>

<style scoped lang="scss">
    .content {
        display: flex;
        flex-direction: row;
        background: white;
        transition: background ease 0.5s;
        border-bottom: 1px solid rgb(224, 229, 232);
        &:hover {
            background: #eef7fd;
        }
    }
    .text {
        display: flex;
        flex-direction: column;
        padding: 15px;
        .title {
            font-size: 14px;
            line-height: 1.4;
            color: rgb(14, 30, 40);
            font-weight: 500;
        }
        .detail {
            padding-top: 3px;
            line-height: 1.5;
            color: rgb(90, 90, 90);
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
        justify-content: center;
        align-items: center;
        padding: 15px;
        color: rgb(102, 117, 127);
        font-weight: 500;
        width: 150px;
        flex: 0 0 auto;
        border-right: 1px solid rgb(224, 229, 232);
    }
    .issue {
        border-right: 1px solid rgb(224, 229, 232);
        width: calc(100% - 450px);
    }

</style>
