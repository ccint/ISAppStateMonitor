import query from '../API/query'

const state = () => {
  return {
    issueList: []
  }
}

const actions = {
  async getIssueList ({ commit }) {
    query.getAllIssues((result) => {
      commit('setIssueList', result)
    })
  }
}

const mutations = {
  setIssueList (state, data) {
    state.issueList = data
  }
}

export default {
  namespaced: true,
  state,
  actions,
  mutations
}
