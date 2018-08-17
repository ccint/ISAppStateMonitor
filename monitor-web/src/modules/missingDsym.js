import {getMissingDsyms} from '../API/query'

const state = () => {
  return {
    missingDsym: {count: 0, data: []}
  }
}

const actions = {
  async getMissingDsyms ({ commit }, {appId}) {
    let result = await getMissingDsyms(appId)
    commit('setMissingDsyms', result.data)
  }
}

const mutations = {
  setMissingDsyms (state, data) {
    state.missingDsym = data
  }
}

export default {
  namespaced: true,
  state,
  actions,
  mutations
}
