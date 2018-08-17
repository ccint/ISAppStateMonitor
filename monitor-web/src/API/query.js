import axios from 'axios'
import config from '../config'

let axiosInstance = axios.create({
  baseURL: config.baseURL
})

let getApps = () => {
  return axiosInstance.get('/allApp')
}

let getAllIssues = (start, pageSize, appId) => {
  return axiosInstance.get(`/query_issues?start=${start}&pageSize=${pageSize}&appId=${appId}`)
}

let getIssueDetails = (id) => {
  return axiosInstance.get(`/issue_detail?id=${id}`)
}

let getIssueSession = (id) => {
  return axiosInstance.get(`/issue_session?id=${id}`)
}

let getMissingDsyms = (appId) => {
  return axiosInstance.get(`/missing_dsym?appId=${appId}`)
}

export {
  getAllIssues,
  getIssueDetails,
  getIssueSession,
  getMissingDsyms,
  getApps
}
