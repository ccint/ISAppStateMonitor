# Debug模式编译是否上传，1＝上传 0＝不上传，默认不上传
UPLOAD_DEBUG_SYMBOLS=0

# 模拟器编译是否上传，1＝上传，0＝不上传，默认不上传
UPLOAD_SIMULATOR_SYMBOLS=0

# 退出执行并打印提示信息
exitWithMessage() {
echo "--------------------------------"
echo -e "${1}"
echo "--------------------------------"
echo "No upload and exit."
echo "----------------------------------------------------------------"
exit ${2}
}

echo "Uploading dSYM..."

DSYM_UPLOAD_DOMAIN="127.0.0.1:4001"

##检查模拟器是否允许上传符号
if [ "$EFFECTIVE_PLATFORM_NAME" == "-iphonesimulator" ]; then
if [[ $UPLOAD_SIMULATOR_SYMBOLS -eq 0 ]]; then
exitWithMessage "Warning: Build for simulator and skipping to upload. \nYou can modify 'UPLOAD_SIMULATOR_SYMBOLS' to 1 in the script." 0
fi
fi

# 检查DEBUG模式是否允许上传符号
if [ "${CONFIGURATION=}" == "Debug" ]; then
if [[ $UPLOAD_DEBUG_SYMBOLS -eq 0 ]]; then
exitWithMessage "Warning: Build for debug mode and skipping to upload. \nYou can modify 'UPLOAD_DEBUG_SYMBOLS' to 1 in the script." 0
fi
fi

function uploadDSYM {
DSYM_SRC="$1"
if [ ! -d "$DSYM_SRC" ]; then
exitWithMessage "dSYM source not found: ${DSYM_SRC}" 1
fi

# 清理
$(find ${BUILT_PRODUCTS_DIR} -name "*.zip" -mindepth 1 -delete)
FILENAME=`basename ${DSYM_SRC}`
DSYM_SYMBOL_OUT_ZIP_NAME="${FILENAME}.zip"
DSYM_ZIP_FPATH="${BUILT_PRODUCTS_DIR}/${DSYM_SYMBOL_OUT_ZIP_NAME}"
cd "${BUILT_PRODUCTS_DIR}";
PAD=`zip -r ${DSYM_SYMBOL_OUT_ZIP_NAME} ${FILENAME}`

if [ ! -e "${DSYM_ZIP_FPATH}" ] ; then
exitWithMessage "no dSYM zip archive generated: ${DSYM_ZIP_FPATH}" 1
fi

echo "dSYM upload domain: ${DSYM_UPLOAD_DOMAIN}"
DSYM_UPLOAD_URL="https://${DSYM_UPLOAD_DOMAIN}/upload_dsym?ignore_ret=1&appName=${PRODUCT_NAME}"

if [ "${CONFIGURATION=}" == "Debug" ]; then
DSYM_UPLOAD_URL="${DSYM_UPLOAD_URL}&isDebug=1"
fi

echo "dSYM upload url: ${DSYM_UPLOAD_URL}"

echo "curl -F \"file=@${DSYM_ZIP_FPATH};type=application/zip\" \"${DSYM_UPLOAD_URL}\" --verbose"

echo "--------------------------------"
STATUS=$(/usr/bin/curl -F "file=@${DSYM_ZIP_FPATH};type=application/zip" "${DSYM_UPLOAD_URL}" --verbose -k)
echo "--------------------------------"

UPLOAD_RESULT="FAILTURE"
echo "server response: ${STATUS}"
if [ ! "${STATUS}" ]; then
echo "Error: Failed to upload the zip archive file."
elif [[ "${STATUS}" == *"\"ret\":\"0\""* ]]; then
UPLOAD_RESULT="SUCCESS"
else
echo "Error: Failed to upload the zip archive file."
fi

echo "--------------------------------"
echo "${UPLOAD_RESULT} - dSYM upload complete."

if [[ "${UPLOAD_RESULT}" == "FAILTURE" ]]; then
echo "--------------------------------"
echo "Failed to upload the dSYM"
echo "Please try it again or upload mannually."
echo "symbol file location: ${DSYM_ZIP_FPATH}"
else
echo "${DSYM_SYMBOL_OUT_ZIP_NAME}" > $CHECKFILEPATH
echo "Remove temporary zip archive: ${DSYM_ZIP_FPATH}"
`rm -f ${DSYM_ZIP_FPATH}`

if [ "$?" -ne 0 ]; then
exitWithMessage "Error: Failed to remove temporary zip archive." 1
fi
fi

echo "-----------------------------------------------------------------"
}


# .dSYM文件信息
echo "DSYM FOLDER ${DWARF_DSYM_FOLDER_PATH}"

DSYM_FOLDER="${DWARF_DSYM_FOLDER_PATH}"

IFS=$'\n'

for dsymFile in `find "$DSYM_FOLDER" -name "${PRODUCT_NAME}.*.dSYM"`; do
echo "Found dSYM file: $dsymFile"
uploadDSYM $dsymFile
done


