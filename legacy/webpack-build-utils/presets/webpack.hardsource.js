const HardSourceWebpackPlugin = require('hard-source-webpack-plugin');

// for more on this plugin see https://github.com/mzgoddard/hard-source-webpack-plugin
module.exports = () => ({
    plugins: [new HardSourceWebpackPlugin(
        {
            configHash: function(webpackConfig) {
                // node-object-hash on npm can be used to build this.
                return require('node-object-hash')({sort: false}).hash(webpackConfig);
            },
        }
    )]
});
