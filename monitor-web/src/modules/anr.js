import {getAllIssues, getIssueSession, getIssueDetails} from '../API/query'

const state = () => {
  return {
    issueList: {total: 0, issues: []},
    currentIssuePage: 1,
    issueDetail: {id: '', sessions: []},
    currentSession: {idx: -1, id: ''}
  }
}

const actions = {
  async getIssueList ({ commit }, {start, pageSize}) {
    let result = await getAllIssues(start, pageSize)
    commit('setIssueList', {total: result.data.total || 0, issues: result.data.issues || []})
  },
  async getIssueDetail ({ commit, state }, {id}) {
    if (id !== state.issueDetail.id) {
      commit('setIssueDetail', {id: '', sessions: []})
      let result = await getIssueDetails(id)
      let sessions = result.data.sessions
      commit('setIssueDetail', {id, sessions: sessions || []})
    }
  },
  async getSessionDetail ({ dispatch, commit, state }, {iid, sid}) {
    await dispatch('getIssueDetail', {id: iid})
    if (state.currentSession.id !== sid) {
      commit('setCurrentSession', {idx: -1})
      let result = await getIssueSession(sid)
      commit('setCurrentSession', {sid, ...result.data})
    }
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
  }
}

export default {
  namespaced: true,
  state,
  actions,
  mutations
}
