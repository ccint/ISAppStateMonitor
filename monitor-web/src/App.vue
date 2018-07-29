<template>
    <div class="layout">
        <Sider :style="{position: 'fixed', height: '100vh', left: 0, overflow: 'auto', backgroundColor: 'rgb(8, 31, 40)'}">
            <Menu active-name="1-2" style="background: transparent" theme="dark" width="auto" :open-names="['1']">
                <div class="app-board" >名片全能王com.intsig.camcard.lite</div>
                <Submenu style="background: transparent" name="1">
                    <template slot="title">
                        <Icon type="ios-navigate"></Icon>
                        ANR
                    </template>
                    <MenuItem name="1-1">ANR</MenuItem>
                </Submenu>
            </Menu>
        </Sider>
        <div class="right-content">
            <Breadcrumb :style="{margin: '30px 0', color: 'color: white'}">
                <BreadcrumbItem style="font-size: 17px; color: white" to="/anr">ANR</BreadcrumbItem>
                <BreadcrumbItem style="font-size: 17px; color: white" to="/anr/issue_detail/">Issue Details</BreadcrumbItem>
            </Breadcrumb>
            <router-view/>
        </div>
    </div>
</template>

<style lang="scss">
    .layout {
        background: rgb(16, 36, 49);
        position: relative;
        border-radius: 4px;
        overflow: hidden;
    }
    .app-board {
        font-family: "PingFang SC", serif;
        font-size: 16px;
        color: white;
        height: 100px;
        display: flex;
        align-items: center;
        justify-content: center;
        text-align: center;
        cursor: pointer;
        transition: background ease 0.5s;
        &:hover {
            background: rgb(20, 52 ,79);
        }
        &:after {
            position: absolute;
            right: 0;
            left: 0;
            top: 100px;
            height: 1px;
            content: ' ';
            background: rgb(32, 53, 61);
        }
    }
    .ivu-breadcrumb a {
        color: white !important;
    }
    .right-content {
        padding: 0 30px 30px 30px;
        margin-left: 200px;
        overflow: scroll;
    }
</style>
