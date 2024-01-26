#/bin/bash!

APP_ID="com.plur.swarmmobile"
OS="darwin"
# OS="android"
# OS="ios"
# OS="iossimulator"
echo "Building for ${OS} with app id: ${APP_ID}"

BUILD_COMMAND='fyne package -os ${OS} -appID ${APP_ID}'
eval "$BUILD_COMMAND";