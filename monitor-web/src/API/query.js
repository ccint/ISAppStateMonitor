import axios from 'axios'
import config from '../config'

let axiosInstance = axios.create({
  baseURL: config.baseURL
})

let getAllIssues = (start, pageSize) => {
  return axiosInstance.get(`/query_issues?start=${start}&pageSize=${pageSize}`)
}

let getIssueDetails = (id) => {
  return axiosInstance.get(`/issue_detail?id=${id}`)
}

let getIssueSession = (id) => {
  return axiosInstance.get(`/issue_session?id=${id}`)
}

export {
  getAllIssues,
  getIssueDetails,
  getIssueSession
}
