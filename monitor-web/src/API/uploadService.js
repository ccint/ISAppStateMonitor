import axios from 'axios'
import config from '../config'

let axiosInstance = axios.create({
  baseURL: config.baseURL
})

let uploadDsym = (formData, processHandler) => {
  let config = {
    onUploadProgress: (progressEvent) => {
      let percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
      processHandler(percentCompleted)
    },
    headers: {'Content-Type': 'application/zip'}
  }
  return axiosInstance.post(`/upload_dsym`, formData, config)
}

export {
  uploadDsym
}
