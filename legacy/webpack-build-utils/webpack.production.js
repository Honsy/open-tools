// for more on devtool see https://webpack.js.org/configuration/devtool/

const webpack = require('webpack');
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;

module.exports = () => ({
    devtool: 'hidden-source-map',
    plugins:[
        // new BundleAnalyzerPlugin()
    ]
});
