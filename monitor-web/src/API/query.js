import axios from 'axios'
import config from '../config'

let axiosInstance = axios.create({
  baseURL: config.baseURL
})

export default {
  getAllIssues: (finishHandler) => {
    let responseHandler = response => {
      finishHandler(response['data'])
    }
    axiosInstance.get('/query_issues').then(responseHandler)
  }
}