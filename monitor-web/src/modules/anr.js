import {getAllIssues, getIssueSession, getIssueDetails, getApps} from '../API/query'
import {reSymbolicate} from '../API/resymbolicate'

const state = () => {
  return {
    issueList: {total: 0, issues: []},
    unclassfiedCount: 0,
    currentIssuePage: 1,
    issueDetail: {id: '', sessions: []},
    currentSession: {idx: -1, id: ''},
    selectedAppIdx: 0,
    apps: [{}]
  }
}

const actions = {
  async getApps ({ commit }) {
    let result = await getApps()
    commit('setApps', result.data.data || [])
  },
  async getIssueList ({ commit }, {start, pageSize, appId}) {
    let result = await getAllIssues(start, pageSize, appId)
    commit('setIssueList', {total: result.data.total || 0, issues: result.data.issues || []})
    commit('setUnclassfiedCount', result.data.unclassfiedCount)
  },
  async getIssueDetail ({ commit, state }, {id}) {
    if (id !== state.issueDetail.id) {
      commit('setIssueDetail', {id: '', sessions: []})
      let result = await getIssueDetails(id)
      let sessions = result.data.sessions
      commit('setIssueDetail', {id, sessions: sessions || []})
    }
  },
  async getSessionDetail ({ dispatch, commit, state }, {iid, sid, foreUpdate}) {
    if (typeof iid !== 'undefined') {
      await dispatch('getIssueDetail', {id: iid})
    }
    if (state.currentSession.id !== sid || foreUpdate === true) {
      commit('setCurrentSession', {idx: -1})
      let result = await getIssueSession(sid)
      commit('setCurrentSession', {sid, ...result.data})
    }
  },
  async reSymbolicate ({ dispatch, commit, state }, {sid}) {
    await reSymbolicate(sid)
    await dispatch('getSessionDetail', {sid, foreUpdate: true})
  }
}

const mutations = {
  setIssueList (state, data) {
    state.issueList = data
  },
  setIssueDetail (state, data) {
    state.issueDetail = data
  },
  setCurrentSession (state, data) {
    let idx = state.issueDetail.sessions.indexOf(data.sid)
    state.currentSession = {idx, ...data}
  },
  setCurrentIssuePage (state, data) {
    state.currentIssuePage = data
  },
  setSelectedAppIdx (state, data) {
    state.selectedAppIdx = data
    state.issueList = {total: 0, issues: []}
    state.issueList = {id: '', sessions: []}
    state.issueList = {idx: -1, id: ''}
  },
  setApps (state, data) {
    state.apps = data
  },
  setUnclassfiedCount (state, data) {
    state.unclassfiedCount = data
  }
}

export default {
  namespaced: true,
  state,
  actions,
  mutations
}
