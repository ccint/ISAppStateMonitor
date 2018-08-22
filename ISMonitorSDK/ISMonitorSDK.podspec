#
#  Be sure to run `pod spec lint ISMonitorSDK.podspec' to ensure this is a
#  valid spec and to remove all comments including this before submitting the spec.
#
#  To learn more about Podspec attributes see http://docs.cocoapods.org/specification.html
#  To see working Podspecs in the CocoaPods repo see https://github.com/CocoaPods/Specs/
#

Pod::Spec.new do |s|
  s.name                   = "ISMonitorSDK"
  s.version                = "0.0.7"
  s.summary                = "ISMonitor platform App-SDK"
  s.homepage               = "http://gitlab.intsig.net/CCiOS/ISMonitorSDK"
  s.license                = "MIT"
  s.author                 = { "Brent Shu" => "brent_shu@intsig.net" }
  s.source                 = { :git => "http://gitlab.intsig.net/CCiOS/ISMonitorSDK.git", :tag => s.version.to_s }
  s.source_files           = "ISMonitorSDK/**/*", "ISMonitorSDK/**/**/*"
  s.public_header_files    = "ISMonitorSDK/arc/ISANRWatcher.h"
  s.requires_arc           = false
  s.requires_arc           = "ISMonitorSDK/arc/*"
  s.platform               = :ios, '8.0'
  s.dependency             "leveldb-library", "~> 1.20"
end
