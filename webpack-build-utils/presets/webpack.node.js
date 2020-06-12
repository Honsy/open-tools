const path = require('path');

module.exports = () => ({
    target: 'node',
    entry: path.resolve(__dirname, './../../client/ssr.tsx'),
    output: {
        filename: 'ssr.js',
        path: path.resolve(__dirname, './../../client/build/'),
        libraryExport: 'default',
        libraryTarget: 'commonjs2'
    }
});
