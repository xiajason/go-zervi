const config = {
  projectName: 'zervigo-mvp-frontend',
  date: '2025-10-29',
  designWidth: 750,
  deviceRatio: {
    640: 2.34 / 2,
    750: 1,
    828: 1.81 / 2
  },
  sourceRoot: 'src',
  outputRoot: 'dist',
  plugins: [
    '@tarojs/plugin-platform-weapp',
    '@tarojs/plugin-platform-alipay',
    '@tarojs/plugin-platform-tt',
    '@tarojs/plugin-platform-swan',
    '@tarojs/plugin-platform-jd',
    '@tarojs/plugin-platform-qq',
    '@tarojs/plugin-platform-h5'
  ],
  defineConstants: {
  },
  copy: {
    patterns: [
    ],
    options: {
    }
  },
  framework: 'react',
  compiler: 'webpack5',
  cache: {
    enable: false // Webpack 持久化缓存配置，建议开启。默认配置请参考：https://docs.taro.zone/docs/config-detail#cache
  },
  mini: {
    postcss: {
      pxtransform: {
        enable: true,
        config: {
          selectorBlackList: ['nut-']
        }
      },
      url: {
        enable: true,
        config: {
          limit: 1024 // 设定转换尺寸上限
        }
      },
      cssModules: {
        enable: false, // 默认为 false，如需使用 css modules 功能，则设为 true
        namingPattern: 'module', // 转换模式，取值为 global/module
        generateScopedName: '[name]__[local]___[hash:base64:5]'
      }
    },
    webpackChain(chain) {
      chain.resolve.alias
        .set('@', path.resolve(__dirname, '..', 'src'))
        .set('@components', path.resolve(__dirname, '..', 'src/components'))
        .set('@utils', path.resolve(__dirname, '..', 'src/utils'))
        .set('@services', path.resolve(__dirname, '..', 'src/services'))
        .set('@store', path.resolve(__dirname, '..', 'src/store'))
        .set('@types', path.resolve(__dirname, '..', 'src/types'))
        .set('@assets', path.resolve(__dirname, '..', 'src/assets'))
    }
  },
  h5: {
    publicPath: '/',
    staticDirectory: 'static',
    postcss: {
      autoprefixer: {
        enable: true,
        config: {
        }
      },
      cssModules: {
        enable: false, // 默认为 false，如需使用 css modules 功能，则设为 true
        namingPattern: 'module', // 转换模式，取值为 global/module
        generateScopedName: '[name]__[local]___[hash:base64:5]'
      }
    },
    webpackChain(chain) {
      chain.resolve.alias
        .set('@', path.resolve(__dirname, '..', 'src'))
        .set('@components', path.resolve(__dirname, '..', 'src/components'))
        .set('@utils', path.resolve(__dirname, '..', 'src/utils'))
        .set('@services', path.resolve(__dirname, '..', 'src/services'))
        .set('@store', path.resolve(__dirname, '..', 'src/store'))
        .set('@types', path.resolve(__dirname, '..', 'src/types'))
        .set('@assets', path.resolve(__dirname, '..', 'src/assets'))
    },
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
  },
  rn: {
    appName: 'zervigoMvpFrontend',
    postcss: {
      cssModules: {
        enable: false, // 默认为 false，如需使用 css modules 功能，则设为 true
      }
    }
  }
}

module.exports = function (merge) {
  if (process.env.NODE_ENV === 'development') {
    return merge({}, config, require('./dev'))
  }
  return merge({}, config, require('./prod'))
}
