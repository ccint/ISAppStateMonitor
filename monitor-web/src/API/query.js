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
    responseHandler(
      {
        data:
          [
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            },
            {
              threadSerial: '0',
              threadName: 'com.apple.main',
              topFrameSymbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]',
              isHighlight: true,
              frames: [
                {
                  imageName: 'libobjc.A.dylib', source: '', symbol: 'objc_msgSend + 28', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController _handleCapturedImagePickerPhotoWithCropOverlayOutput:]', isHighlight: false
                },
                {
                  imageName: 'CameraUI', source: '', symbol: '-[CAMImagePickerCameraViewController cropOverlayWasOKed:]', isHighlight: false
                },
                {
                  imageName: 'UIKit', source: '', symbol: 'UIApplicationMain + 236', isHighlight: false
                },
                {
                  imageName: 'CamCard_zh_lite', source: 'main.m line 23', symbol: 'main', isHighlight: true
                },
                {
                  imageName: 'libdyld.dylib', source: '', symbol: 'start + 4', isHighlight: false
                }
              ]
            }
          ]
      }
    )
    //axiosInstance.get('/query_issues').then(responseHandler)
  }
}
