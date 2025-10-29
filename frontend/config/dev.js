module.exports = {
  env: {
    NODE_ENV: '"development"'
  },
  defineConstants: {
    API_BASE_URL: '"http://localhost:9000"',
    WS_BASE_URL: '"ws://localhost:9000"'
  },
  mini: {},
  h5: {
    devServer: {
      host: 'localhost',
      port: 10086,
      https: false,
      proxy: {
        '/api': {
          target: 'http://localhost:9000',
          changeOrigin: true,
          pathRewrite: {
            '^/api': '/api'
          }
        }
      }
    }
  }
}
