const path = require('path');
const webpack = require('webpack');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

module.exports = (env) => ({
    entry: {
      react:path.resolve(__dirname, './../../client/index.tsx'),
      prettierPage:path.resolve(__dirname, './../../client/pages/prettier/index.tsx'),
      home:path.resolve(__dirname, './../../client/pages/home/index.tsx')
    },
    output: {
        path: path.resolve(__dirname, './../../static/js/'),
        filename:"[name].js"
    },
    module: {
        rules: [
            {
              test: /\.css$/,
              use: [
                'style-loader',
                {
                  loader: MiniCssExtractPlugin.loader,
                  options: {
                    publicPath: '../',
                  },
                },
                'css-loader',
              ],
            }
        ],
    },
    optimization: {
      splitChunks: {
        chunks: 'async',
        minSize: 30000,
        maxSize: 0,
        minChunks: 1,
        maxAsyncRequests: 6,
        maxInitialRequests: 4,
        automaticNameDelimiter: '~',
        cacheGroups: {
          react: {
            test: /[\\/]node_modules[\\/](react|react-dom)[\\/]/,
            filename:"react-vendors.js",
            chunks: 'all',
          },
          prettier: {
            test: /[\\/]node_modules[\\/](prettier)[\\/]/,
            filename:"prettier.js",
            chunks: 'all',
          },
          babylon: {
            test: /[\\/]node_modules[\\/](babylon|babylonjs)[\\/]/,
            filename:"babylon.js",
            chunks: 'all',
          }
        }
      }
    }
})
