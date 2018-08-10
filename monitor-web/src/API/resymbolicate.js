import axios from 'axios'
import config from '../config'

let axiosInstance = axios.create({
  baseURL: config.baseURL
})

let reSymbolicate = (reportId) => {
  return axiosInstance.get(`/resymbolicate?report_id=${reportId}`)
}

export {
  reSymbolicate
}
