module.exports = {
  chainWebpack: config => {
    config.module
      .rule('ttf')
      .test(/\.ttf$/)
      .use('url-loader')
      .loader('url-loader')
      .end()
  }
}